package pricing_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	now = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)
	vnd = func(s string) shared.Money { return shared.MustParseMoney(s, shared.VND) }
	fx  = shared.MustExchangeRate(shared.USD, shared.VND, "26000")
)

// The business numbers as chốt in SETUP.md §7: Denver sales tax 8.81 %, margin
// max(10 %, 500 000 ₫ floor), deposit 50 %, quote valid 48 h.
func policy(t *testing.T) pricing.QuotePolicy {
	t.Helper()
	margin, err := pricing.NewMarginPolicy(shared.MustParsePercent("10"), vnd("500000"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := pricing.NewQuotePolicy(pricing.QuotePolicyDetails{
		SalesTax: shared.MustParsePercent("8.81"), Margin: margin,
		Deposit: shared.MustParsePercent("50"), TTL: 48 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func footwearProfile() pricing.CategoryProfile {
	return pricing.CategoryProfile{
		Code: "footwear", Class: pricing.ClassStandard,
		Estimate: shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13)),
	}
}

func shoeListing(measured bool) pricing.Listing {
	l := pricing.Listing{
		Product: shared.NewID(), Name: "Air Trainer 90", Category: "footwear",
		Price: usd("150.00"), Active: true,
	}
	if measured {
		l.Parcel = shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))
		l.Measured = true
	}
	return l
}

func inputs(t *testing.T, measured bool) pricing.QuoteInputs {
	t.Helper()
	return pricing.QuoteInputs{
		Listing: shoeListing(measured), Profile: footwearProfile(), Lane: aLane(t), FX: fx, Policy: policy(t),
	}
}

// THE calculation, checked line by line with the numbers a spreadsheet gives:
//
//	item          150.00 USD
//	sales tax      13.22 USD   (8.81 %, half-up)
//	freight        22.50 USD   (2034 g volumetric → 2.5 kg step × $9.00)
//	subtotal      185.72 USD  → × 26 000 = 4 828 720 ₫
//	service fee  500 000 ₫    (10 % = 482 872 < floor)
//	total      5 328 720 ₫    deposit 50 % = 2 664 360 ₫
func TestCalculate_measuredProduct(t *testing.T) {
	b, err := pricing.Calculate(inputs(t, true))
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	want := map[string]string{
		"ItemPrice": "150.00 USD", "SalesTax": "13.22 USD", "Freight": "22.50 USD", "Surcharge": "0.00 USD", "Duty": "0.00 USD",
		"SubtotalUSD": "185.72 USD", "SubtotalVND": "4828720 VND", "ServiceFee": "500000 VND", "TotalVND": "5328720 VND", "Deposit": "2664360 VND",
	}
	got := map[string]string{
		"ItemPrice": b.ItemPrice.String(), "SalesTax": b.SalesTax.String(), "Freight": b.Freight.String(), "Surcharge": b.Surcharge.String(), "Duty": b.Duty.String(),
		"SubtotalUSD": b.SubtotalUSD.String(), "SubtotalVND": b.SubtotalVND.String(), "ServiceFee": b.ServiceFee.String(), "TotalVND": b.TotalVND.String(), "Deposit": b.Deposit.String(),
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s = %s, want %s", k, got[k], w)
		}
	}
	if b.Chargeable.String() != "2.500 kg" || b.Estimated || b.Class != pricing.ClassStandard {
		t.Errorf("chargeable %s estimated %v class %s", b.Chargeable, b.Estimated, b.Class)
	}
}

// Nobody has weighed this product: the quote is built on the CATEGORY default
// and says so (Estimated). Quote vs Actual starts with knowing you guessed.
func TestCalculate_unmeasuredProductUsesCategoryEstimate(t *testing.T) {
	b, err := pricing.Calculate(inputs(t, false))
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if !b.Estimated || b.Chargeable.String() != "2.000 kg" || b.Freight.String() != "18.00 USD" {
		t.Errorf("estimated %v, chargeable %s, freight %s", b.Estimated, b.Chargeable, b.Freight)
	}
	if b.TotalVND.String() != "5211720 VND" || b.Deposit.String() != "2605860 VND" {
		t.Errorf("total %s, deposit %s", b.TotalVND, b.Deposit)
	}
}

// When 10 % beats the floor, the percentage wins — the floor is a minimum.
func TestCalculate_marginFloorIsAMinimum(t *testing.T) {
	in := inputs(t, true)
	in.Listing.Price = usd("1000.00") // subtotal ≈ 1 111 USD → 10 % ≈ 2.9 M ₫ > 500 000 ₫
	b, err := pricing.Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	if b.ServiceFee.Minor() <= 500000 {
		t.Errorf("fee = %s, want the 10 %% share above the floor", b.ServiceFee)
	}
	if b.ServiceFee != b.SubtotalVND.Mul(shared.MustParsePercent("10")) {
		t.Errorf("fee = %s, want exactly 10 %% of %s", b.ServiceFee, b.SubtotalVND)
	}
}

func TestCalculate_refusals(t *testing.T) {
	in := inputs(t, true)
	in.Listing.Active = false
	if _, err := pricing.Calculate(in); !errors.Is(err, pricing.ErrListingInactive) {
		t.Errorf("retired listing: got %v", err)
	}
	in = inputs(t, true)
	in.Listing.Price = vnd("3900000")
	if _, err := pricing.Calculate(in); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("price not in the lane's currency: got %v", err)
	}
	in = inputs(t, true)
	in.FX = shared.MustExchangeRate(shared.VND, shared.USD, "0.0000384615")
	if _, err := pricing.Calculate(in); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("wrong FX direction: got %v", err)
	}
	in = inputs(t, false)
	in.Profile = pricing.CategoryProfile{} // unmeasured AND no category default: nothing to quote from
	if _, err := pricing.Calculate(in); !errors.Is(err, pricing.ErrNothingToWeigh) {
		t.Errorf("no parcel anywhere: got %v", err)
	}
}

// ── Quote aggregate ──────────────────────────────────────────────────────────

func TestIssueQuote_snapshotsEverythingAndExpires(t *testing.T) {
	q, err := pricing.IssueQuote(inputs(t, true), now)
	if err != nil {
		t.Fatalf("IssueQuote: %v", err)
	}
	if q.ID().IsZero() || q.Status() != pricing.QuoteIssued || !q.ExpiresAt().Equal(now.Add(48*time.Hour)) {
		t.Fatalf("quote = %+v", q.Snapshot())
	}
	if q.Breakdown().TotalVND.String() != "5328720 VND" || q.Breakdown().FX != fx {
		t.Errorf("breakdown not snapshotted: %+v", q.Breakdown())
	}
	evs := q.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "pricing.quote_issued" {
		t.Fatalf("got %v, want one quote_issued", evs)
	}
	issued := evs[0].(pricing.QuoteIssuedEvent)
	if issued.TotalVND.String() != "5328720 VND" || issued.Deposit.String() != "2664360 VND" || !issued.ExpiresAt.Equal(q.ExpiresAt()) {
		t.Errorf("event = %+v", issued)
	}

	if q.IsExpired(now.Add(47*time.Hour)) || !q.IsExpired(now.Add(49*time.Hour)) {
		t.Error("IsExpired must compare against ExpiresAt")
	}
}

// The quote is a promise with a deadline. Accepting after it is not a quote
// any more — the rate and the parcel may both have moved (DDD.md §28).
func TestQuote_acceptAndExpire(t *testing.T) {
	q, _ := pricing.IssueQuote(inputs(t, true), now)
	q.PullEvents()

	if err := q.Accept(now.Add(49 * time.Hour)); !errors.Is(err, pricing.ErrQuoteExpired) {
		t.Fatalf("accept after expiry: got %v", err)
	}
	if q.Status() != pricing.QuoteExpired {
		t.Error("a late Accept must leave the quote expired, not issued")
	}
	if evs := q.PullEvents(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_expired" {
		t.Errorf("got %v, want quote_expired", evs)
	}

	q2, _ := pricing.IssueQuote(inputs(t, true), now)
	q2.PullEvents()
	if err := q2.Accept(now.Add(time.Hour)); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	if q2.Status() != pricing.QuoteAccepted {
		t.Error("must be accepted")
	}
	if err := q2.Accept(now.Add(time.Hour)); !errors.Is(err, pricing.ErrQuoteNotIssued) {
		t.Errorf("accept twice: got %v", err)
	}
	if err := q2.Expire(now.Add(72 * time.Hour)); !errors.Is(err, pricing.ErrQuoteNotIssued) {
		t.Errorf("expire an accepted quote: got %v", err)
	}
	if evs := q2.PullEvents(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_accepted" {
		t.Errorf("got %v, want one quote_accepted", evs)
	}
}

func TestQuote_snapshotRoundTrip(t *testing.T) {
	q, _ := pricing.IssueQuote(inputs(t, false), now)
	if err := q.Accept(now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	q.PullEvents()
	snap := q.Snapshot()
	back, err := pricing.QuoteFromSnapshot(snap)
	if err != nil {
		t.Fatalf("QuoteFromSnapshot: %v", err)
	}
	if !reflect.DeepEqual(back.Snapshot(), snap) {
		t.Fatalf("round trip changed the state:\n got  %+v\n want %+v", back.Snapshot(), snap)
	}
	snap.Status = "bananas"
	if _, err := pricing.QuoteFromSnapshot(snap); !errors.Is(err, pricing.ErrInvalidSnapshot) {
		t.Errorf("bad status: got %v", err)
	}
}

func TestNewQuotePolicy_everyFieldIsValidated(t *testing.T) {
	good := policy(t)
	_ = good
	margin, _ := pricing.NewMarginPolicy(shared.MustParsePercent("10"), vnd("500000"))
	base := pricing.QuotePolicyDetails{SalesTax: shared.MustParsePercent("8.81"), Margin: margin, Deposit: shared.MustParsePercent("50"), TTL: 48 * time.Hour}
	zeroIsMeaningful := map[string]string{"SalesTax": "0 % is a legal sales tax (Oregon has none)"}
	typ := reflect.TypeOf(base)
	for i := range typ.NumField() {
		if why, ok := zeroIsMeaningful[typ.Field(i).Name]; ok {
			t.Logf("skip %s: %s", typ.Field(i).Name, why)
			continue
		}
		d := base
		reflect.ValueOf(&d).Elem().Field(i).SetZero()
		if _, err := pricing.NewQuotePolicy(d); err == nil {
			t.Errorf("NewQuotePolicy accepted zero %s", typ.Field(i).Name)
		}
	}
	d := base
	d.Deposit = shared.MustParsePercent("150")
	if _, err := pricing.NewQuotePolicy(d); !errors.Is(err, pricing.ErrInvalidPolicy) {
		t.Errorf("deposit over 100%%: got %v", err)
	}
	if _, err := pricing.NewMarginPolicy(shared.MustParsePercent("-1"), vnd("500000")); !errors.Is(err, pricing.ErrInvalidPolicy) {
		t.Errorf("negative margin: got %v", err)
	}
}
