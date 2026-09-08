package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

func usd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.USD)
}

func vnd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.VND)
}

// aLane has every optional set — a surcharge AND an itemised duty — so the
// round trip proves the nullable and the flag columns both survive.
func aLane(t *testing.T) pricing.ShippingLane {
	t.Helper()
	rates, err := pricing.NewRateCard(usd("9.00"), usd("10.00"), usd("12.00"), usd("14.00"))
	if err != nil {
		t.Fatal(err)
	}
	duty, err := pricing.DutyItemised(shared.MustParsePercent("5"))
	if err != nil {
		t.Fatal(err)
	}
	lane, err := pricing.NewShippingLane(pricing.LaneDetails{
		Code: pricing.MustParseLaneCode("us_forwarder"), Name: "US forwarder", Divisor: 5000, Step: shared.Grams(500),
		Rates: rates, BatterySurcharge: usd("3.00"), Duty: duty,
	})
	if err != nil {
		t.Fatal(err)
	}
	return lane
}

func aPolicy(t *testing.T) pricing.QuotePolicy {
	t.Helper()
	margin, err := pricing.NewMarginPolicy(shared.MustParsePercent("10"), vnd("500000"))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := pricing.NewQuotePolicy(pricing.QuotePolicyDetails{
		SalesTax: shared.MustParsePercent("8.81"), Margin: margin, Deposit: shared.MustParsePercent("50"), TTL: 48 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func TestLaneRepo_roundTripAndRedefine(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewLaneRepo(p)
	lane := aLane(t)

	if _, err := repo.ByCode(ctx, lane.Code()); !errors.Is(err, pricing.ErrLaneNotFound) {
		t.Fatalf("missing lane: %v, want ErrLaneNotFound", err)
	}
	if err := repo.Save(ctx, lane); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.ByCode(ctx, lane.Code())
	if err != nil {
		t.Fatalf("ByCode: %v", err)
	}
	if !reflect.DeepEqual(got, lane) {
		t.Fatalf("round trip changed the lane:\n got %+v\nwant %+v", got, lane)
	}

	// A bundled lane with no surcharge: the NULL and the false paths.
	plain, _ := pricing.NewShippingLane(pricing.LaneDetails{
		Code: pricing.MustParseLaneCode("us_forwarder"), Name: "US forwarder, new list", Divisor: 6000, Step: shared.Grams(100), Rates: lane.Rates(),
	})
	if err := repo.Save(ctx, plain); err != nil {
		t.Fatalf("redefine: %v", err)
	}
	all, err := repo.All(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("All = %d, %v — redefining must replace, not add", len(all), err)
	}
	if !reflect.DeepEqual(all[0], plain) {
		t.Fatalf("redefined lane = %+v", all[0])
	}
}

func TestQuoteRepo_roundTripAndStatusChange(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewQuoteRepo(p)
	lane := aLane(t)

	q, err := pricing.IssueQuote(pricing.QuoteInputs{
		Listing: pricing.Listing{Product: shared.NewID(), Name: "Air Trainer 90", Category: "footwear", Price: usd("150.00"),
			Parcel: shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), Measured: true, Active: true},
		Profile: pricing.CategoryProfile{Code: "footwear", Class: pricing.ClassBranded, Restrictions: []string{"battery"}},
		Lane:    lane, FX: shared.MustExchangeRate(shared.USD, shared.VND, "26000.5"), Policy: aPolicy(t),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	q.PullEvents()

	if _, err := repo.ByID(ctx, q.ID()); !errors.Is(err, pricing.ErrQuoteNotFound) {
		t.Fatalf("missing quote: %v, want ErrQuoteNotFound", err)
	}
	if err := repo.Save(ctx, q); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.ByID(ctx, q.ID())
	if err != nil {
		t.Fatalf("ByID: %v", err)
	}
	if !reflect.DeepEqual(got.Snapshot(), q.Snapshot()) {
		t.Fatalf("round trip changed the quote:\n got %+v\nwant %+v", got.Snapshot(), q.Snapshot())
	}
	if b := got.Breakdown(); b.FX.Rate() != "26000.5" || b.Surcharge != usd("3.00") || b.Duty != usd("7.50") {
		t.Errorf("breakdown lines: fx %s surcharge %s duty %s", b.FX.Rate(), b.Surcharge, b.Duty)
	}

	if err := got.Accept(now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, got); err != nil {
		t.Fatalf("Save after accept: %v", err)
	}
	again, _ := repo.ByID(ctx, q.ID())
	if again.Status() != pricing.QuoteAccepted {
		t.Fatalf("status after accept = %s", again.Status())
	}
}

func TestListingRepo_measuredAndUnmeasured(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewListingRepo(p)
	id := shared.NewID()

	if _, err := repo.ByProduct(ctx, id); !errors.Is(err, pricing.ErrListingNotFound) {
		t.Fatalf("missing listing: %v", err)
	}
	unmeasured := pricing.Listing{Product: id, Name: "Air Trainer 90", Category: "footwear", Price: usd("150.00"), Active: true}
	if err := repo.Save(ctx, unmeasured); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ByProduct(ctx, id)
	if err != nil || !reflect.DeepEqual(got, unmeasured) {
		t.Fatalf("unmeasured round trip = %+v, %v", got, err)
	}

	measured := unmeasured
	measured.Parcel, measured.Measured = shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), true
	measured.Price = usd("160.00")
	if err := repo.Save(ctx, measured); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.ByProduct(ctx, id)
	if !reflect.DeepEqual(got, measured) {
		t.Fatalf("upsert = %+v, want %+v", got, measured)
	}
}

func TestProfileRepo_roundTrip(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewProfileRepo(p)

	if _, err := repo.ByCode(ctx, "footwear"); !errors.Is(err, pricing.ErrProfileNotFound) {
		t.Fatalf("missing profile: %v", err)
	}
	prof := pricing.CategoryProfile{Code: "electronics", Class: pricing.ClassElectronics,
		Estimate: shared.MustParcelSpec(shared.Grams(400), shared.NewDimensionsCM(20, 18, 8)), Restrictions: []string{"battery", "magnet"}}
	if err := repo.Save(ctx, prof); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ByCode(ctx, "electronics")
	if err != nil || !reflect.DeepEqual(got, prof) {
		t.Fatalf("round trip = %+v, %v", got, err)
	}
	plain := pricing.CategoryProfile{Code: "footwear", Class: pricing.ClassBranded, Estimate: prof.Estimate}
	if err := repo.Save(ctx, plain); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.ByCode(ctx, "footwear"); !reflect.DeepEqual(got, plain) {
		t.Fatalf("no restrictions must read back as nil: %+v", got)
	}
}

func TestExchangeRates_setReplacesCurrent(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	rates := postgres.NewExchangeRates(p)

	if _, err := rates.Current(ctx, shared.USD, shared.VND); !errors.Is(err, pricing.ErrNoExchangeRate) {
		t.Fatalf("no rate yet: %v", err)
	}
	for _, r := range []string{"26000", "26150.25"} {
		if err := rates.Set(ctx, shared.MustExchangeRate(shared.USD, shared.VND, r)); err != nil {
			t.Fatal(err)
		}
	}
	got, err := rates.Current(ctx, shared.USD, shared.VND)
	if err != nil || got.Rate() != "26150.25" {
		t.Fatalf("Current = %s, %v", got, err)
	}
	if _, err := rates.Current(ctx, shared.VND, shared.USD); !errors.Is(err, pricing.ErrNoExchangeRate) {
		t.Fatalf("reverse pair must not exist: %v", err)
	}
}

// IssuedBefore is the sweep's one query, and the only place the partial index
// quotes_open_idx is used. Three quotes in three states: it must return the
// stale issued one and neither of the others — a sweep that picked up an
// accepted quote would cancel an order somebody already paid a deposit on.
func TestQuoteRepo_issuedBeforeSeesOnlyOpenAndStale(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewQuoteRepo(p)

	issue := func(at time.Time) *pricing.Quote {
		t.Helper()
		q, err := pricing.IssueQuote(pricing.QuoteInputs{
			Listing: pricing.Listing{Product: shared.NewID(), Name: "Air Trainer 90", Category: "footwear", Price: usd("150.00"),
				Parcel: shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), Measured: true, Active: true},
			Profile: pricing.CategoryProfile{Code: "footwear", Class: pricing.ClassBranded},
			Lane:    aLane(t), FX: shared.MustExchangeRate(shared.USD, shared.VND, "26000"), Policy: aPolicy(t),
		}, at)
		if err != nil {
			t.Fatal(err)
		}
		q.PullEvents()
		if err := repo.Save(ctx, q); err != nil {
			t.Fatal(err)
		}
		return q
	}

	stale := issue(now.Add(-72 * time.Hour)) // expired 24h ago (TTL 48h)
	fresh := issue(now)                      // valid for another 48h
	accepted := issue(now.Add(-72 * time.Hour))
	if err := accepted.Accept(now.Add(-71 * time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, accepted); err != nil {
		t.Fatal(err)
	}

	due, err := repo.IssuedBefore(ctx, now, 100)
	if err != nil {
		t.Fatalf("IssuedBefore: %v", err)
	}
	if len(due) != 1 || due[0].ID() != stale.ID() {
		t.Fatalf("IssuedBefore returned %d quotes, want only %s", len(due), stale.ID())
	}
	if !reflect.DeepEqual(due[0].Snapshot(), stale.Snapshot()) {
		t.Errorf("the row read by the sweep is not the quote that was saved:\n got %+v\nwant %+v", due[0].Snapshot(), stale.Snapshot())
	}
	_ = fresh

	// limit caps the pass; ordering is oldest deadline first, so the quotes
	// that have waited longest are the ones a capped pass takes.
	older := issue(now.Add(-96 * time.Hour))
	due, err = repo.IssuedBefore(ctx, now, 1)
	if err != nil {
		t.Fatalf("IssuedBefore(limit 1): %v", err)
	}
	if len(due) != 1 || due[0].ID() != older.ID() {
		t.Fatalf("limit 1 took %v, want the oldest deadline %s", due, older.ID())
	}

	// Once expired, a quote leaves the sweep's sight for good.
	if err := stale.Expire(now); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, stale); err != nil {
		t.Fatal(err)
	}
	due, _ = repo.IssuedBefore(ctx, now, 100)
	for _, q := range due {
		if q.ID() == stale.ID() {
			t.Fatal("an expired quote must not come back")
		}
	}
}
