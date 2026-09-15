package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// seedAcceptedQuote puts in place both projections PlaceOrder reads: pricing's
// accepted quote and catalog's variant dictionary. It returns the quote and the
// ONE variant id the API will take for it.
func (a *api) seedAcceptedQuote(t *testing.T) (shared.ID, string) {
	t.Helper()
	ctx := context.Background()
	q := ordering.AcceptedQuote{Quote: shared.NewID(), Product: shared.NewID(),
		Total: shared.MustParseMoney("5393720", shared.VND), Deposit: shared.MustParseMoney("2696860", shared.VND)}
	if err := a.acceptedQuotes.Save(ctx, q); err != nil {
		t.Fatal(err)
	}
	v := ordering.Variant{Variant: shared.NewID(), Product: q.Product}
	if err := a.variants.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	return q.Quote, v.Variant.String()
}

// seedForeignVariant is a real variant of a DIFFERENT product: the shape of a
// copy-pasted id, which must not become an order.
func (a *api) seedForeignVariant(t *testing.T) string {
	t.Helper()
	v := ordering.Variant{Variant: shared.NewID(), Product: shared.NewID()}
	if err := a.variants.Save(context.Background(), v); err != nil {
		t.Fatal(err)
	}
	return v.Variant.String()
}

func TestOrders_placePayCancel(t *testing.T) {
	a := newAPI()
	quote, variant := a.seedAcceptedQuote(t)
	body := `{"quote_id":"` + quote.String() + `","variant_id":"` + variant + `"}`

	expectError(t, a.call(t, "POST", "/orders", `{"quote_id":"x","variant_id":"`+variant+`"}`, a.asCustomer()), 400, "invalid_id")
	expectError(t, a.call(t, "POST", "/orders", `{"quote_id":"`+shared.NewID().String()+`","variant_id":"`+variant+`"}`, a.asCustomer()), 409, "quote_not_accepted")
	// The size is the customer's own choice, so it is the one field the API
	// checks against catalog instead of believing. Both refusals are 409:
	// nothing about the request is malformed, the state says no.
	expectError(t, a.call(t, "POST", "/orders", `{"quote_id":"`+quote.String()+`","variant_id":"`+shared.NewID().String()+`"}`, a.asCustomer()), 409, "variant_unknown")
	expectError(t, a.call(t, "POST", "/orders", `{"quote_id":"`+quote.String()+`","variant_id":"`+a.seedForeignVariant(t)+`"}`, a.asCustomer()), 409, "variant_not_for_product")
	a.outbox.Drain()
	id := idOf(t, a.call(t, "POST", "/orders", body, a.asCustomer()))
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "ordering.order_placed" {
		t.Fatalf("outbox = %v", evs)
	}
	expectError(t, a.call(t, "POST", "/orders", body, a.asCustomer()), 409, "quote_already_used")

	// read model
	expectError(t, a.call(t, "GET", "/orders/"+ordering.NewOrderID().String(), "", nil), 404, "order_not_found")
	var v struct {
		Status  string `json:"status"`
		Total   money  `json:"total"`
		Balance money  `json:"balance"`
		Refund  *money `json:"refund"`
	}
	read := func() {
		rec := a.call(t, "GET", "/orders/"+id, "", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET = %d %s", rec.Code, rec.Body)
		}
		v.Refund = nil
		if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
	}
	read()
	if v.Status != "awaiting_deposit" || v.Total.Amount != "5393720" || v.Balance.Amount != "2696860" || v.Refund != nil {
		t.Fatalf("view = %+v", v)
	}

	// deposit: exact amount, locale-aware like every amount at the edge
	expectError(t, a.call(t, "POST", "/orders/"+id+"/deposit", `{"amount":"1.000.000","currency":"VND"}`, map[string]string{"Accept-Language": "vi"}), 409, "wrong_amount")
	expectError(t, a.call(t, "POST", "/orders/"+id+"/deposit", `{"amount":"2696860","currency":"XXX"}`, nil), 400, "unknown_currency")
	if rec := a.call(t, "POST", "/orders/"+id+"/deposit", `{"amount":"2.696.860","currency":"VND"}`, map[string]string{"Accept-Language": "vi"}); rec.Code != http.StatusNoContent {
		t.Fatalf("deposit = %d %s", rec.Code, rec.Body)
	}
	expectError(t, a.call(t, "POST", "/orders/"+id+"/deposit", `{"amount":"2696860","currency":"VND"}`, nil), 409, "not_awaiting_deposit")
	expectError(t, a.call(t, "POST", "/orders/"+id+"/balance", `{"amount":"2696860","currency":"VND"}`, nil), 409, "not_in_transit")
	expectError(t, a.call(t, "POST", "/orders/"+id+"/deliver", "", nil), 409, "not_in_transit")
	read()
	if v.Status != "deposited" {
		t.Fatalf("status = %s", v.Status)
	}

	// cancel before purchase: full refund, visible in the read model
	expectError(t, a.call(t, "POST", "/orders/"+id+"/cancel", `{"reason":""}`, nil), 400, "empty_reason")
	a.outbox.Drain()
	if rec := a.call(t, "POST", "/orders/"+id+"/cancel", `{"reason":"found it cheaper"}`, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("cancel = %d %s", rec.Code, rec.Body)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "ordering.order_cancelled" {
		t.Fatalf("outbox = %v", evs)
	}
	read()
	if v.Status != "cancelled" || v.Refund == nil || v.Refund.Amount != "2696860" {
		t.Fatalf("view after cancel = %+v", v)
	}
	expectError(t, a.call(t, "POST", "/orders/"+id+"/cancel", `{"reason":"again"}`, nil), 409, "order_cancelled")
	expectError(t, a.call(t, "POST", "/orders/"+id+"/deposit", `{"amount":"2696860","currency":"VND"}`, nil), 409, "order_cancelled")
}

// The one test that matters most in P10-PLAN.md: a customer's own token names
// them already, so a customer_id in the body can only be an attempt to order
// in somebody else's name — refused outright, and no order left behind.
func TestPlaceOrder_customerCannotOrderInSomebodyElsesName(t *testing.T) {
	a := newAPI()
	quote, variant := a.seedAcceptedQuote(t)
	body := `{"quote_id":"` + quote.String() + `","variant_id":"` + variant + `","customer_id":"` + shared.NewID().String() + `"}`

	expectError(t, a.call(t, "POST", "/orders", body, a.asCustomer()), http.StatusForbidden, "forbidden")

	if _, err := a.orders.ByQuote(context.Background(), quote); err == nil {
		t.Fatal("a refused impersonation attempt must not create an order")
	}
}

// An operator ordering for a customer who is not holding a token must say
// who the customer is — there is no token to fall back on.
func TestPlaceOrder_operatorMustNameTheCustomer(t *testing.T) {
	a := newAPI()
	quote, variant := a.seedAcceptedQuote(t)
	body := `{"quote_id":"` + quote.String() + `","variant_id":"` + variant + `"}`

	expectError(t, a.call(t, "POST", "/orders", body, a.asOperator()), 400, "customer_required")

	if _, err := a.orders.ByQuote(context.Background(), quote); err == nil {
		t.Fatal("a rejected order must not have been created")
	}
}

// The other side of P10: an operator NAMING a customer places the order for
// them, and the order remembers who did it.
func TestPlaceOrder_operatorPlacesOnBehalfOfANamedCustomer(t *testing.T) {
	a := newAPI()
	quote, variant := a.seedAcceptedQuote(t)
	customer := shared.NewID()
	body := `{"quote_id":"` + quote.String() + `","variant_id":"` + variant + `","customer_id":"` + customer.String() + `"}`

	id := idOf(t, a.call(t, "POST", "/orders", body, a.asOperator()))
	oid, err := ordering.ParseOrderID(id)
	if err != nil {
		t.Fatal(err)
	}
	o, err := a.orders.ByID(context.Background(), oid)
	if err != nil {
		t.Fatal(err)
	}
	if o.Customer() != customer || o.PlacedBy() != a.operatorID {
		t.Fatalf("order = %+v, want customer %s placed by %s", o.Snapshot(), customer, a.operatorID)
	}
}
