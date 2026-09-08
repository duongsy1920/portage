package memory_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The in-memory repositories of the four later contexts. They are the adapter
// every unit test in the system runs on, so the properties tested here are the
// ones every other test quietly depends on:
//
//	a miss returns the DOMAIN's sentinel, not nil and not a bespoke error
//	a second Save REPLACES (upsert), it does not add a row
//	the list queries answer the question their name asks, in a stable order
//
// A Postgres adapter that disagreed with any of these would make the same test
// pass here and fail there — which is the failure mode a shared port exists to
// prevent (DDD.md §20).

func vnd(s string) shared.Money { return shared.MustParseMoney(s, shared.VND) }

func anOrder(t *testing.T, at time.Time) *ordering.CustomerOrder {
	t.Helper()
	o, err := ordering.PlaceOrder(ordering.OrderDetails{
		Quote: shared.NewID(), Product: shared.NewID(), Variant: shared.NewID(), Customer: shared.NewID(),
		Total: vnd("5393720"), Deposit: vnd("2696860"),
	}, at)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestOrderRepo_byIDByQuoteAndUpsert(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewOrderRepo()
	o := anOrder(t, now)

	if _, err := repo.ByID(ctx, o.ID()); !errors.Is(err, ordering.ErrOrderNotFound) {
		t.Fatalf("miss = %v, want ErrOrderNotFound", err)
	}
	if _, err := repo.ByQuote(ctx, o.Quote()); !errors.Is(err, ordering.ErrOrderNotFound) {
		t.Fatalf("ByQuote miss = %v, want ErrOrderNotFound", err)
	}
	if err := repo.Save(ctx, o); err != nil {
		t.Fatal(err)
	}
	// ByQuote is the second half of "one quote, one order": the use case asks
	// it before placing, so a repo that could not answer would make the rule
	// unenforceable.
	got, err := repo.ByQuote(ctx, o.Quote())
	if err != nil || got.ID() != o.ID() {
		t.Fatalf("ByQuote = %v, %v", got, err)
	}

	if err := o.PayDeposit(vnd("2696860"), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, o); err != nil {
		t.Fatal(err)
	}
	if repo.Len() != 1 {
		t.Fatalf("%d orders after saving the same one twice", repo.Len())
	}
	if again, _ := repo.ByID(ctx, o.ID()); again.Status() != ordering.StatusDeposited {
		t.Errorf("status after the second save = %s", again.Status())
	}
}

func TestAcceptedQuoteRepo_upsert(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAcceptedQuoteRepo()
	q := ordering.AcceptedQuote{Quote: shared.NewID(), Product: shared.NewID(), Total: vnd("100"), Deposit: vnd("50")}

	if _, err := repo.ByID(ctx, q.Quote); !errors.Is(err, ordering.ErrAcceptedQuoteNotFound) {
		t.Fatalf("miss = %v", err)
	}
	for range 2 { // the relay is at-least-once: the projector saves twice
		if err := repo.Save(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.ByID(ctx, q.Quote)
	if err != nil || got.Total != q.Total {
		t.Fatalf("ByID = %+v, %v", got, err)
	}
}

func aTask(t *testing.T, at time.Time) *procurement.PurchaseTask {
	t.Helper()
	task, err := procurement.OpenTask(procurement.TaskDetails{
		Order: shared.NewID(), Product: shared.NewID(), Variant: shared.NewID(), Currency: shared.USD,
	}, at)
	if err != nil {
		t.Fatal(err)
	}
	return task
}

// Open is the buyer's to-do list: only tasks nobody has closed, OLDEST FIRST.
// The order matters — it is what makes the list a queue rather than a pile.
func TestTaskRepo_openIsTheBuyersQueue(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewTaskRepo()
	older, newer, done := aTask(t, now), aTask(t, now.Add(time.Hour)), aTask(t, now.Add(2*time.Hour))

	if _, err := repo.ByID(ctx, older.ID()); !errors.Is(err, procurement.ErrTaskNotFound) {
		t.Fatalf("miss = %v", err)
	}
	if _, err := repo.ByOrder(ctx, older.Order()); !errors.Is(err, procurement.ErrTaskNotFound) {
		t.Fatalf("ByOrder miss = %v", err)
	}
	for _, task := range []*procurement.PurchaseTask{newer, older, done} {
		if err := repo.Save(ctx, task); err != nil {
			t.Fatal(err)
		}
	}
	receipt := procurement.PurchaseReceipt{Reference: "NK-1", Paid: shared.MustParseMoney("163.22", shared.USD), PaidBy: shared.NewOperatorID()}
	if err := done.Confirm(receipt, now.Add(3*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, done); err != nil {
		t.Fatal(err)
	}

	open, err := repo.Open(ctx)
	if err != nil || len(open) != 2 {
		t.Fatalf("Open = %d tasks, %v — a closed task must leave the list", len(open), err)
	}
	if open[0].ID() != older.ID() || open[1].ID() != newer.ID() {
		t.Error("the buyer's list is not oldest-first")
	}
	if got, _ := repo.ByOrder(ctx, older.Order()); got.ID() != older.ID() {
		t.Error("ByOrder found the wrong task")
	}
}

func TestShopAndItemRepos_missesUseDomainSentinels(t *testing.T) {
	ctx := context.Background()
	shops, items := memory.NewShopRepo(), memory.NewItemRepo()
	merchant, product := shared.NewID(), shared.NewID()

	if _, err := shops.ByID(ctx, merchant); !errors.Is(err, procurement.ErrShopNotFound) {
		t.Fatalf("shop miss = %v", err)
	}
	if _, err := items.ByProduct(ctx, product); !errors.Is(err, procurement.ErrItemNotFound) {
		t.Fatalf("item miss = %v", err)
	}
	if err := shops.Save(ctx, procurement.Shop{Merchant: merchant, Name: "Example", Site: "www.example.com", Currency: shared.USD}); err != nil {
		t.Fatal(err)
	}
	if err := items.Save(ctx, procurement.Item{Product: product, Merchant: merchant, Name: "Air Trainer 90"}); err != nil {
		t.Fatal(err)
	}
	if s, _ := shops.ByID(ctx, merchant); s.Currency != shared.USD {
		t.Errorf("shop = %+v", s)
	}
	if i, _ := items.ByProduct(ctx, product); i.Merchant != merchant {
		t.Errorf("item = %+v", i)
	}
}

func aParcel(t *testing.T, at time.Time) *logistics.Parcel {
	t.Helper()
	p, err := logistics.ExpectParcel(logistics.ParcelDetails{
		Order: shared.NewID(), Reference: "NK-1",
	}, at)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// Pending is the warehouse screen: boxes not yet in a batch, oldest first.
func TestParcelRepo_pendingIsTheWarehouseList(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewParcelRepo()
	first, second := aParcel(t, now), aParcel(t, now.Add(time.Hour))

	if _, err := repo.ByID(ctx, first.ID()); !errors.Is(err, logistics.ErrParcelNotFound) {
		t.Fatalf("miss = %v", err)
	}
	if _, err := repo.ByOrder(ctx, first.Order()); !errors.Is(err, logistics.ErrParcelNotFound) {
		t.Fatalf("ByOrder miss = %v", err)
	}
	for _, p := range []*logistics.Parcel{second, first} {
		if err := repo.Save(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	pending, err := repo.Pending(ctx)
	if err != nil || len(pending) != 2 {
		t.Fatalf("Pending = %d, %v", len(pending), err)
	}
	if pending[0].ID() != first.ID() {
		t.Error("Pending is not oldest-first")
	}
	if got, _ := repo.ByOrder(ctx, first.Order()); got.ID() != first.ID() {
		t.Error("ByOrder found the wrong parcel")
	}
}

func TestBatchAndLaneRuleRepos(t *testing.T) {
	ctx := context.Background()
	batches, rules := memory.NewBatchRepo(), memory.NewLaneRuleRepo()

	rule := logistics.LaneRule{Code: "us_forwarder", Divisor: 5000, Step: shared.Grams(500)}
	if _, err := rules.ByCode(ctx, rule.Code); !errors.Is(err, logistics.ErrLaneRuleNotFound) {
		t.Fatalf("rule miss = %v", err)
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatal(err)
	}
	if got, _ := rules.ByCode(ctx, rule.Code); got.Divisor != 5000 {
		t.Errorf("rule = %+v", got)
	}

	b, err := logistics.OpenBatch("us_forwarder", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := batches.ByID(ctx, b.ID()); !errors.Is(err, logistics.ErrBatchNotFound) {
		t.Fatalf("batch miss = %v", err)
	}
	if err := batches.Save(ctx, b); err != nil {
		t.Fatal(err)
	}
	if got, _ := batches.ByID(ctx, b.ID()); got.ID() != b.ID() {
		t.Error("ByID found the wrong batch")
	}
}

func TestReconciliationRepo_upsertGrowsTheRow(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewReconciliationRepo()
	order := shared.NewID()

	if _, err := repo.ByOrder(ctx, order); !errors.Is(err, pricing.ErrReconciliationNotFound) {
		t.Fatalf("miss = %v", err)
	}
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	rec := pricing.Reconciliation{Order: order, Quote: shared.NewID(),
		QuotedGoods: usd("163.22"), QuotedFreight: usd("25.00"), QuotedChargeable: shared.Grams(2500)}
	if err := repo.Save(ctx, rec); err != nil {
		t.Fatal(err)
	}
	// The row grows as the three feeds arrive: a second Save replaces it.
	rec.ActualGoods, rec.ActualFreight, rec.ActualChargeable = usd("163.22"), usd("27.50"), shared.Grams(2500)
	if err := repo.Save(ctx, rec); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ByOrder(ctx, order)
	if err != nil || !got.Complete() {
		t.Fatalf("reconciliation = %+v, %v", got, err)
	}
	v, err := got.Variance()
	if err != nil || v.String() != "-2.50 USD" {
		t.Errorf("variance = %s, %v", v, err)
	}
}

// The pricing repos, plus the outbox's read side. Both are exercised heavily
// by other packages' tests, which is why they were the last to get one of
// their own — and exactly why they needed it: "covered by somebody else's
// test" is coverage that disappears the day that test changes.
func TestPricingRepos_lanesQuotesListingsProfilesAndRates(t *testing.T) {
	ctx := context.Background()
	lanes, quotes := memory.NewLaneRepo(), memory.NewQuoteRepo()
	listings, profiles, rates := memory.NewListingRepo(), memory.NewProfileRepo(), memory.NewExchangeRates()

	code := pricing.MustParseLaneCode("us_forwarder")
	if _, err := lanes.ByCode(ctx, code); !errors.Is(err, pricing.ErrLaneNotFound) {
		t.Fatalf("lane miss = %v", err)
	}
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	card, err := pricing.NewRateCard(usd("9.00"), usd("10.00"), usd("12.00"), usd("14.00"))
	if err != nil {
		t.Fatal(err)
	}
	lane, err := pricing.NewShippingLane(pricing.LaneDetails{
		Code: code, Name: "US forwarder", Divisor: 5000, Step: shared.Grams(500), Rates: card,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := lanes.Save(ctx, lane); err != nil {
		t.Fatal(err)
	}
	if all, err := lanes.All(ctx); err != nil || len(all) != 1 {
		t.Fatalf("All = %d lanes, %v", len(all), err)
	}
	if err := lanes.Save(ctx, lane); err != nil { // redefining replaces
		t.Fatal(err)
	}
	if all, _ := lanes.All(ctx); len(all) != 1 {
		t.Errorf("redefining a lane added a second one: %d", len(all))
	}

	product := shared.NewID()
	if _, err := listings.ByProduct(ctx, product); !errors.Is(err, pricing.ErrListingNotFound) {
		t.Fatalf("listing miss = %v", err)
	}
	if err := listings.Save(ctx, pricing.Listing{Product: product, Name: "Air Trainer 90", Price: usd("150.00"), Active: true}); err != nil {
		t.Fatal(err)
	}
	if l, _ := listings.ByProduct(ctx, product); !l.Active || listings.Len() != 1 {
		t.Errorf("listing = %+v, len %d", l, listings.Len())
	}

	if _, err := profiles.ByCode(ctx, "footwear"); !errors.Is(err, pricing.ErrProfileNotFound) {
		t.Fatalf("profile miss = %v", err)
	}
	if err := profiles.Save(ctx, pricing.CategoryProfile{Code: "footwear", Class: pricing.ClassBranded}); err != nil {
		t.Fatal(err)
	}
	if p, _ := profiles.ByCode(ctx, "footwear"); p.Class != pricing.ClassBranded {
		t.Errorf("profile = %+v", p)
	}

	// Rates: "today's rate" has one answer at a time, and Set replaces it.
	if _, err := rates.Current(ctx, shared.USD, shared.VND); !errors.Is(err, pricing.ErrNoExchangeRate) {
		t.Fatalf("rate miss = %v", err)
	}
	if err := rates.Set(ctx, shared.MustExchangeRate(shared.USD, shared.VND, "26000")); err != nil {
		t.Fatal(err)
	}
	if err := rates.Set(ctx, shared.MustExchangeRate(shared.USD, shared.VND, "27000")); err != nil {
		t.Fatal(err)
	}
	if r, _ := rates.Current(ctx, shared.USD, shared.VND); r.Rate() != "27000" {
		t.Errorf("rate = %s, want the latest", r.Rate())
	}

	// Quotes, including the sweep's query (IssuedBefore).
	if _, err := quotes.ByID(ctx, pricing.NewQuoteID()); !errors.Is(err, pricing.ErrQuoteNotFound) {
		t.Fatalf("quote miss = %v", err)
	}
	margin, _ := pricing.NewMarginPolicy(shared.MustParsePercent("10"), vnd("500000"))
	policy, err := pricing.NewQuotePolicy(pricing.QuotePolicyDetails{
		SalesTax: shared.MustParsePercent("8.81"), Margin: margin, Deposit: shared.MustParsePercent("50"), TTL: 48 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	issue := func(at time.Time) *pricing.Quote {
		t.Helper()
		q, err := pricing.IssueQuote(pricing.QuoteInputs{
			Listing: pricing.Listing{Product: shared.NewID(), Name: "x", Category: "footwear", Price: usd("150.00"),
				Parcel: shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), Measured: true, Active: true},
			Profile: pricing.CategoryProfile{Code: "footwear", Class: pricing.ClassBranded},
			Lane:    lane, FX: shared.MustExchangeRate(shared.USD, shared.VND, "26000"), Policy: policy,
		}, at)
		if err != nil {
			t.Fatal(err)
		}
		q.PullEvents()
		if err := quotes.Save(ctx, q); err != nil {
			t.Fatal(err)
		}
		return q
	}
	stale := issue(now.Add(-72 * time.Hour))
	issue(now) // still valid
	due, err := quotes.IssuedBefore(ctx, now, 100)
	if err != nil || len(due) != 1 || due[0].ID() != stale.ID() {
		t.Fatalf("IssuedBefore = %d quotes, %v — only the one past its deadline", len(due), err)
	}
	if quotes.Len() != 2 {
		t.Errorf("%d quotes", quotes.Len())
	}
}

// The outbox seen from the RELAY's side: Pending, then MarkSent, and the rows
// marked do not come back. Drain (the test-only side) is covered elsewhere.
func TestOutbox_pendingAndMarkSent(t *testing.T) {
	ctx := context.Background()
	outbox := memory.NewOutbox()
	o := anOrder(t, now)
	if err := outbox.Append(ctx, o.PullEvents()); err != nil {
		t.Fatal(err)
	}

	pending, err := outbox.Pending(ctx, 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("Pending = %d, %v", len(pending), err)
	}
	if pending[0].Name != "ordering.order_placed" {
		t.Fatalf("entry = %+v", pending[0])
	}
	if err := outbox.MarkSent(ctx, []int64{pending[0].ID}, now); err != nil {
		t.Fatal(err)
	}
	if again, _ := outbox.Pending(ctx, 10); len(again) != 0 {
		t.Errorf("a sent row came back: %+v", again)
	}
}

// The two reference lists a form turns into select boxes, and the property
// that makes them usable: a STABLE order, matching what Postgres returns.
//
// Go randomises map iteration on purpose, so a repository that just ranges
// over its map hands back a different order every call. On a screen that is a
// select box whose default answer changes on every page load, which reads as a
// broken form long before anybody suspects the adapter.
func TestReferenceLists_comeBackInAStableOrder(t *testing.T) {
	ctx := context.Background()
	cats := memory.NewCategoryRepo()
	for _, code := range []string{"luggage", "footwear", "electronics", "apparel"} {
		p, err := catalog.NewCategoryPolicy(catalog.MustParseCategoryCode(code),
			shared.MustParcelSpec(shared.Grams(1000), shared.NewDimensionsCM(30, 20, 10)), nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := cats.Save(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"apparel", "electronics", "footwear", "luggage"} // by code, like ORDER BY code
	for range 8 {                                                     // enough calls that a random order would show
		all, err := cats.All(ctx)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, 0, len(all))
		for _, c := range all {
			got = append(got, c.Code().String())
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("categories = %v, want %v", got, want)
		}
	}

	// Merchants sort by id, which is UUIDv7 and therefore creation order: the
	// list a person reads grows at the bottom instead of reshuffling.
	shops := memory.NewMerchantRepo()
	first := aShop(t, "www.first.com")
	second := aShop(t, "www.second.com")
	for _, m := range []*catalog.Merchant{first, second} {
		if err := shops.Save(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	for range 8 {
		all, err := shops.All(ctx)
		if err != nil || len(all) != 2 {
			t.Fatalf("All = %d, %v", len(all), err)
		}
		if all[0].ID() != first.ID() || all[1].ID() != second.ID() {
			t.Fatalf("merchants came back %s, %s; want %s, %s", all[0].ID(), all[1].ID(), first.ID(), second.ID())
		}
	}
}

func aShop(t *testing.T, site string) *catalog.Merchant {
	t.Helper()
	m, err := catalog.RegisterMerchant(catalog.MerchantDetails{
		Name: site, Site: catalog.MustParseHostname(site), Currency: shared.USD,
		Sourcing: []catalog.SourcingMode{catalog.SourcedByOperator},
	}, time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	m.PullEvents()
	return m
}
