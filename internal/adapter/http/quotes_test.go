package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// money is one amount as the API renders it.
type money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// seedPricing gives pricing what the relay would have delivered: a lane, a
// rate, a category profile and a published listing.
func (a *api) seedPricing(t *testing.T) shared.ID {
	t.Helper()
	ctx := context.Background()
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	rates, _ := pricing.NewRateCard(usd("9.00"), usd("10.00"), usd("12.00"), usd("14.00"))
	lane, err := pricing.NewShippingLane(pricing.LaneDetails{
		Code: pricing.MustParseLaneCode("us_forwarder"), Name: "US forwarder", Divisor: 5000, Step: shared.Grams(500), Rates: rates,
	})
	if err != nil {
		t.Fatal(err)
	}
	product := shared.NewID()
	for _, err := range []error{
		a.lanes.Save(ctx, lane),
		a.profiles.Save(ctx, pricing.CategoryProfile{Code: "footwear", Class: pricing.ClassBranded,
			Estimate: shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13))}),
		a.listings.Save(ctx, pricing.Listing{Product: product, Name: "Air Trainer 90", Category: "footwear", Price: usd("150.00"),
			Parcel: shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), Measured: true, Active: true}),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	return product
}

func TestQuotes_issueReadAccept(t *testing.T) {
	a := newAPI()
	product := a.seedPricing(t)
	body := `{"product_id":"` + product.String() + `","lane":"us_forwarder"}`

	// 400 / 404 / 409 before anything is right
	expectError(t, a.call(t, "POST", "/quotes", `{"product_id":"nope","lane":"us_forwarder"}`, nil), 400, "invalid_id")
	expectError(t, a.call(t, "POST", "/quotes", `{"product_id":"`+product.String()+`","lane":"US Forwarder!"}`, nil), 400, "invalid_lane_code")
	expectError(t, a.call(t, "POST", "/quotes", `{"product_id":"`+shared.NewID().String()+`","lane":"us_forwarder"}`, nil), 404, "listing_not_found")
	expectError(t, a.call(t, "POST", "/quotes", `{"product_id":"`+product.String()+`","lane":"sea_freight"}`, nil), 404, "lane_not_found")
	expectError(t, a.call(t, "POST", "/quotes", body, nil), 409, "no_exchange_rate") // nobody set today's rate yet

	if err := a.rates.Set(context.Background(), shared.MustExchangeRate(shared.USD, shared.VND, "26000")); err != nil {
		t.Fatal(err)
	}
	a.outbox.Drain()
	id := idOf(t, a.call(t, "POST", "/quotes", body, nil))
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_issued" {
		t.Fatalf("outbox = %v", evs)
	}

	// read model
	expectError(t, a.call(t, "GET", "/quotes/"+pricing.NewQuoteID().String(), "", nil), 404, "quote_not_found")
	rec := a.call(t, "GET", "/quotes/"+id, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET = %d %s", rec.Code, rec.Body)
	}
	var v struct {
		Status      string `json:"status"`
		Class       string `json:"class"`
		ChargeableG int64  `json:"chargeable_g"`
		Estimated   bool   `json:"estimated"`
		ExpiresAt   string `json:"expires_at"`
		FX          string `json:"fx"`
		Lines       struct {
			Item     money `json:"item"`
			SalesTax money `json:"sales_tax"`
			Freight  money `json:"freight"`
			Subtotal money `json:"subtotal"`
		} `json:"lines"`
		Home struct {
			Total   money `json:"total"`
			Deposit money `json:"deposit"`
		} `json:"home"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	// 1250 g vs 34×23×13 cm / 5000 = 2034 g → 2500 g at 10.00/kg = 25.00; tax 8.81 % of 150 = 13.22;
	// 188.22 USD × 26000 = 4 893 720; fee max(10 %, 500 000) = 500 000; total 5 393 720; deposit half.
	want := map[string]string{
		"status": "issued", "class": "branded", "fx": "26000", "expires": now.Add(48 * time.Hour).Format(time.RFC3339),
		"item": "150.00", "tax": "13.22", "freight": "25.00", "subtotal": "188.22", "total": "5393720", "deposit": "2696860",
	}
	got := map[string]string{
		"status": v.Status, "class": v.Class, "fx": v.FX, "expires": v.ExpiresAt,
		"item": v.Lines.Item.Amount, "tax": v.Lines.SalesTax.Amount, "freight": v.Lines.Freight.Amount, "subtotal": v.Lines.Subtotal.Amount,
		"total": v.Home.Total.Amount, "deposit": v.Home.Deposit.Amount,
	}
	for k := range want {
		if got[k] != want[k] {
			t.Errorf("%s = %q, want %q", k, got[k], want[k])
		}
	}
	if v.ChargeableG != 2500 || v.Estimated || v.Home.Total.Currency != "VND" || v.Lines.Item.Currency != "USD" {
		t.Errorf("view = %+v", v)
	}

	// accept
	expectError(t, a.call(t, "POST", "/quotes/nope/accept", "", a.asCustomer()), 400, "invalid_id")
	expectError(t, a.call(t, "POST", "/quotes/"+pricing.NewQuoteID().String()+"/accept", "", a.asCustomer()), 404, "quote_not_found")
	if rec := a.call(t, "POST", "/quotes/"+id+"/accept", "", a.asCustomer()); rec.Code != http.StatusNoContent {
		t.Fatalf("accept = %d %s", rec.Code, rec.Body)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_accepted" {
		t.Fatalf("outbox = %v", evs)
	}
	expectError(t, a.call(t, "POST", "/quotes/"+id+"/accept", "", a.asCustomer()), 409, "quote_not_issued")
}

// Time passes; the quote is a promise with a deadline. Accepting after it is
// refused AND flips the quote to expired — the read model shows it.
func TestQuotes_lateAcceptExpires(t *testing.T) {
	a := newAPI()
	product := a.seedPricing(t)
	if err := a.rates.Set(context.Background(), shared.MustExchangeRate(shared.USD, shared.VND, "26000")); err != nil {
		t.Fatal(err)
	}
	id := idOf(t, a.call(t, "POST", "/quotes", `{"product_id":"`+product.String()+`","lane":"us_forwarder"}`, nil))
	a.outbox.Drain()

	a.clock.Advance(49 * time.Hour)
	expectError(t, a.call(t, "POST", "/quotes/"+id+"/accept", "", a.asCustomer()), 409, "quote_expired")
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.quote_expired" {
		t.Fatalf("outbox = %v", evs)
	}
	rec := a.call(t, "GET", "/quotes/"+id, "", nil)
	var v struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &v)
	if v.Status != "expired" {
		t.Fatalf("status after a late accept = %q, want expired", v.Status)
	}
}
