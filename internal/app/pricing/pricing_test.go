package pricingapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/clock"
)

var now = time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)

type world struct {
	clock    *clock.Fixed
	lanes    *memory.LaneRepo
	quotes   *memory.QuoteRepo
	listings *memory.ListingRepo
	profiles *memory.ProfileRepo
	rates    *memory.ExchangeRates
	outbox   *memory.Outbox
	deps     pricingapp.Deps
}

func newWorld(t *testing.T) *world {
	t.Helper()
	margin, _ := pricing.NewMarginPolicy(shared.MustParsePercent("10"), shared.MustParseMoney("500000", shared.VND))
	policy, err := pricing.NewQuotePolicy(pricing.QuotePolicyDetails{
		SalesTax: shared.MustParsePercent("8.81"), Margin: margin, Deposit: shared.MustParsePercent("50"), TTL: 48 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	classes, _ := pricing.NewClassification(map[string]pricing.GoodsClass{"electronics": pricing.ClassElectronics})
	w := &world{
		clock: clock.FixedAt(now), lanes: memory.NewLaneRepo(), quotes: memory.NewQuoteRepo(),
		listings: memory.NewListingRepo(), profiles: memory.NewProfileRepo(), rates: memory.NewExchangeRates(),
		outbox: memory.NewOutbox(),
	}
	w.deps = pricingapp.Deps{
		Clock: w.clock, UoW: memory.UnitOfWork{}, Outbox: w.outbox,
		Lanes: w.lanes, Quotes: w.quotes, Listings: w.listings, Profiles: w.profiles, Rates: w.rates,
		Policy: policy, Classification: classes,
	}
	return w
}

func (w *world) seedLaneAndRate(t *testing.T) {
	t.Helper()
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	rates, _ := pricing.NewRateCard(usd("9.00"), usd("10.00"), usd("12.00"), usd("14.00"))
	lane, err := pricing.NewShippingLane(pricing.LaneDetails{
		Code: pricing.MustParseLaneCode("us_forwarder"), Name: "US forwarder", Divisor: 5000, Step: shared.Grams(500),
		Rates: rates, BatterySurcharge: usd("3.00"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.lanes.Save(context.Background(), lane); err != nil {
		t.Fatal(err)
	}
	if err := w.rates.Set(context.Background(), shared.MustExchangeRate(shared.USD, shared.VND, "26000")); err != nil {
		t.Fatal(err)
	}
}

var shoe = shared.NewID()

func published() contracts.ProductPublishedV1 {
	return contracts.ProductPublishedV1{
		ID: shoe.String(), Merchant: shared.NewID().String(), Category: "footwear", Name: "Air Trainer 90",
		Price:  contracts.MoneyV1{Minor: 15000, Currency: "USD"},
		Parcel: contracts.ParcelV1{WeightG: 1250, LengthMM: 340, WidthMM: 230, HeightMM: 130}, At: now,
	}
}

func footwearDefined() contracts.CategoryDefinedV1 {
	return contracts.CategoryDefinedV1{Code: "footwear",
		Estimate: contracts.ParcelV1{WeightG: 1200, LengthMM: 330, WidthMM: 220, HeightMM: 130}, At: now}
}

// The projector is the first REAL subscriber: it turns catalog's wire messages
// into pricing's own records. It must be idempotent — the relay may deliver a
// message twice (at-least-once) — and tolerate order (a measurement for a
// product it has not heard of yet is not an error).
func TestProjector_buildsListingsAndProfilesIdempotently(t *testing.T) {
	w := newWorld(t)
	p := pricingapp.NewProjector(w.deps)
	ctx := context.Background()

	if err := p.OnProductMeasured(ctx, contracts.ProductMeasuredV1{ID: shoe.String(), Parcel: contracts.ParcelV1{WeightG: 1, LengthMM: 1, WidthMM: 1, HeightMM: 1}, At: now}); err != nil {
		t.Fatalf("measurement before publish must be ignored, got %v", err)
	}
	if _, err := w.listings.ByProduct(ctx, shoe); !errors.Is(err, pricing.ErrListingNotFound) {
		t.Fatal("no listing may exist before the product is published")
	}

	for range 2 { // twice: at-least-once
		if err := p.OnProductPublished(ctx, published()); err != nil {
			t.Fatal(err)
		}
	}
	l, err := w.listings.ByProduct(ctx, shoe)
	if err != nil || !l.Active || !l.Measured || l.Price.String() != "150.00 USD" || l.Parcel.Weight() != shared.Grams(1250) || l.Category != "footwear" {
		t.Fatalf("listing = %+v, %v", l, err)
	}
	if w.listings.Len() != 1 {
		t.Errorf("%d listings after publishing the same product twice", w.listings.Len())
	}

	if err := p.OnProductRepriced(ctx, contracts.ProductRepricedV1{ID: shoe.String(), From: contracts.MoneyV1{Minor: 15000, Currency: "USD"}, To: contracts.MoneyV1{Minor: 16000, Currency: "USD"}, At: now}); err != nil {
		t.Fatal(err)
	}
	if err := p.OnProductMeasured(ctx, contracts.ProductMeasuredV1{ID: shoe.String(), Parcel: contracts.ParcelV1{WeightG: 1300, LengthMM: 340, WidthMM: 230, HeightMM: 130}, Verified: true, At: now}); err != nil {
		t.Fatal(err)
	}
	l, _ = w.listings.ByProduct(ctx, shoe)
	if l.Price.String() != "160.00 USD" || l.Parcel.Weight() != shared.Grams(1300) {
		t.Errorf("updates not applied: %+v", l)
	}
	if err := p.OnProductRetired(ctx, contracts.ProductRetiredV1{ID: shoe.String(), Reason: "gone", At: now}); err != nil {
		t.Fatal(err)
	}
	if l, _ = w.listings.ByProduct(ctx, shoe); l.Active {
		t.Error("retired product must be inactive")
	}

	// Categories: class comes from pricing's OWN classification, not the wire.
	for range 2 {
		if err := p.OnCategoryDefined(ctx, footwearDefined()); err != nil {
			t.Fatal(err)
		}
	}
	prof, err := w.profiles.ByCode(ctx, "footwear")
	if err != nil || prof.Class != pricing.ClassStandard || prof.Estimate.Weight() != shared.Grams(1200) {
		t.Fatalf("profile = %+v, %v", prof, err)
	}
	if err := p.OnCategoryDefined(ctx, contracts.CategoryDefinedV1{Code: "electronics", Estimate: contracts.ParcelV1{WeightG: 400, LengthMM: 200, WidthMM: 180, HeightMM: 80}, Restrictions: []string{"battery"}, At: now}); err != nil {
		t.Fatal(err)
	}
	if prof, _ = w.profiles.ByCode(ctx, "electronics"); prof.Class != pricing.ClassElectronics || len(prof.Restrictions) != 1 {
		t.Errorf("electronics profile = %+v", prof)
	}
}

// The use case: load what pricing knows, ask the world for today's rate, let
// the domain compute, freeze. Same numbers as the domain test — through the
// whole stack.
func TestIssueQuote_freezesTheNumbers(t *testing.T) {
	w := newWorld(t)
	w.seedLaneAndRate(t)
	p := pricingapp.NewProjector(w.deps)
	if err := p.OnCategoryDefined(context.Background(), footwearDefined()); err != nil {
		t.Fatal(err)
	}
	if err := p.OnProductPublished(context.Background(), published()); err != nil {
		t.Fatal(err)
	}
	h := pricingapp.NewIssueQuoteHandler(w.deps)

	id, err := h.Handle(context.Background(), pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	q, err := w.quotes.ByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if q.Breakdown().TotalVND.String() != "5328720 VND" || q.Breakdown().Deposit.String() != "2664360 VND" || q.Breakdown().Estimated {
		t.Errorf("breakdown = %+v", q.Breakdown())
	}
	if !q.ExpiresAt().Equal(now.Add(48 * time.Hour)) {
		t.Errorf("expires %v", q.ExpiresAt())
	}
	if evs := w.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_issued" {
		t.Fatalf("outbox = %v", evs)
	}
}

func TestIssueQuote_refusals(t *testing.T) {
	w := newWorld(t)
	w.seedLaneAndRate(t)
	p := pricingapp.NewProjector(w.deps)
	h := pricingapp.NewIssueQuoteHandler(w.deps)
	ctx := context.Background()

	if _, err := h.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"}); !errors.Is(err, pricing.ErrListingNotFound) {
		t.Errorf("unknown product: got %v", err)
	}
	if err := p.OnProductPublished(ctx, published()); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "sea_freight"}); !errors.Is(err, pricing.ErrLaneNotFound) {
		t.Errorf("unknown lane: got %v", err)
	}
	if _, err := h.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "Sea Freight!"}); !errors.Is(err, pricing.ErrInvalidLaneCode) {
		t.Errorf("bad lane code: got %v", err)
	}
	// A measured product needs no category profile; an unmeasured one with no
	// profile has nothing to weigh — the domain's refusal passes through.
	if _, err := h.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"}); err != nil {
		t.Errorf("measured product without profile must still quote: %v", err)
	}
	if err := p.OnProductRetired(ctx, contracts.ProductRetiredV1{ID: shoe.String(), Reason: "gone", At: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"}); !errors.Is(err, pricing.ErrListingInactive) {
		t.Errorf("retired: got %v", err)
	}

	w2 := newWorld(t) // lane but no rate
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	rates, _ := pricing.NewRateCard(usd("9.00"), usd("10.00"), usd("12.00"), usd("14.00"))
	lane, _ := pricing.NewShippingLane(pricing.LaneDetails{Code: pricing.MustParseLaneCode("us_forwarder"), Name: "x", Divisor: 5000, Step: shared.Grams(500), Rates: rates})
	_ = w2.lanes.Save(ctx, lane)
	_ = pricingapp.NewProjector(w2.deps).OnProductPublished(ctx, published())
	if _, err := pricingapp.NewIssueQuoteHandler(w2.deps).Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"}); !errors.Is(err, pricing.ErrNoExchangeRate) {
		t.Errorf("no rate: got %v", err)
	}
}

// Accept flips the quote and announces; a late Accept expires it — and that
// change is SAVED, so the quote does not come back as "issued" on reload.
func TestAcceptQuote(t *testing.T) {
	w := newWorld(t)
	w.seedLaneAndRate(t)
	ctx := context.Background()
	p := pricingapp.NewProjector(w.deps)
	_ = p.OnCategoryDefined(ctx, footwearDefined())
	_ = p.OnProductPublished(ctx, published())
	id, err := pricingapp.NewIssueQuoteHandler(w.deps).Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"})
	if err != nil {
		t.Fatal(err)
	}
	w.outbox.Drain()
	accept := pricingapp.NewAcceptQuoteHandler(w.deps)

	if err := accept.Handle(ctx, pricing.NewQuoteID()); !errors.Is(err, pricing.ErrQuoteNotFound) {
		t.Errorf("unknown quote: got %v", err)
	}
	w.clock.Advance(time.Hour)
	if err := accept.Handle(ctx, id); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	q, _ := w.quotes.ByID(ctx, id)
	if q.Status() != pricing.QuoteAccepted {
		t.Error("must be accepted")
	}
	if evs := w.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_accepted" {
		t.Errorf("outbox = %v", evs)
	}

	late, _ := pricingapp.NewIssueQuoteHandler(w.deps).Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"})
	w.outbox.Drain()
	w.clock.Advance(72 * time.Hour)
	if err := accept.Handle(ctx, late); !errors.Is(err, pricing.ErrQuoteExpired) {
		t.Fatalf("late accept: got %v", err)
	}
	if q, _ := w.quotes.ByID(ctx, late); q.Status() != pricing.QuoteExpired {
		t.Error("the expiry decided during Accept must be persisted")
	}
	if evs := w.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_expired" {
		t.Errorf("outbox = %v", evs)
	}
}

// The sweep is the first use case nothing calls: no customer, no operator, no
// event — only the clock. Three quotes in three states prove it touches
// exactly the one the business means, and a second pass proves it is
// idempotent (the relay is at-least-once; so is a cron that overlaps itself).
func TestExpireQuotes_sweepsOnlyTheOnesPastTheirDeadline(t *testing.T) {
	w := newWorld(t)
	w.seedLaneAndRate(t)
	ctx := context.Background()
	p := pricingapp.NewProjector(w.deps)
	_ = p.OnCategoryDefined(ctx, footwearDefined())
	_ = p.OnProductPublished(ctx, published())
	issue := pricingapp.NewIssueQuoteHandler(w.deps)

	stale, err := issue.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"}) // issued at T
	if err != nil {
		t.Fatal(err)
	}
	answered, err := issue.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"})
	if err != nil {
		t.Fatal(err)
	}
	if err := pricingapp.NewAcceptQuoteHandler(w.deps).Handle(ctx, answered); err != nil {
		t.Fatal(err)
	}
	w.clock.Advance(47 * time.Hour) // both above are now 47h old; the TTL is 48h
	fresh, err := issue.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"})
	if err != nil {
		t.Fatal(err)
	}
	w.clock.Advance(2 * time.Hour) // stale is 49h old, fresh is 2h old
	w.outbox.Drain()

	sweep := pricingapp.NewExpireQuotesHandler(w.deps, 100)
	n, err := sweep.Handle(ctx)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if n != 1 {
		t.Fatalf("expired %d quotes, want exactly the one past its deadline", n)
	}
	if q, _ := w.quotes.ByID(ctx, stale); q.Status() != pricing.QuoteExpired {
		t.Errorf("stale quote = %s, want expired", q.Status())
	}
	if q, _ := w.quotes.ByID(ctx, fresh); q.Status() != pricing.QuoteIssued {
		t.Errorf("a quote still inside its TTL must be left alone, got %s", q.Status())
	}
	if q, _ := w.quotes.ByID(ctx, answered); q.Status() != pricing.QuoteAccepted {
		t.Errorf("an accepted quote must never be swept, got %s", q.Status())
	}
	if evs := w.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_expired" {
		t.Fatalf("outbox = %v, want one pricing.quote_expired", evs)
	}

	// Second pass, same clock: nothing left to do, and nothing announced
	// twice. A sweep that re-published quote_expired every minute would give
	// every subscriber the same message forever.
	n, err = sweep.Handle(ctx)
	if err != nil || n != 0 {
		t.Fatalf("second pass = %d, %v; want 0, nil", n, err)
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Errorf("second pass announced %v", evs)
	}
}

// limit bounds ONE pass, it does not drop work: the backlog is finished by the
// passes that follow. Getting this wrong the other way — a limit that silently
// skips quotes — would leave the oldest ones expiring never.
func TestExpireQuotes_limitSpreadsTheBacklogOverPasses(t *testing.T) {
	w := newWorld(t)
	w.seedLaneAndRate(t)
	ctx := context.Background()
	p := pricingapp.NewProjector(w.deps)
	_ = p.OnCategoryDefined(ctx, footwearDefined())
	_ = p.OnProductPublished(ctx, published())
	issue := pricingapp.NewIssueQuoteHandler(w.deps)
	for range 5 {
		if _, err := issue.Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"}); err != nil {
			t.Fatal(err)
		}
	}
	w.clock.Advance(49 * time.Hour)

	sweep := pricingapp.NewExpireQuotesHandler(w.deps, 2)
	total := 0
	for range 3 {
		n, err := sweep.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
		total += n
	}
	if total != 5 {
		t.Fatalf("three passes of 2 expired %d of 5 quotes", total)
	}
	if n, _ := sweep.Handle(ctx); n != 0 {
		t.Errorf("a fourth pass found %d more", n)
	}
}
