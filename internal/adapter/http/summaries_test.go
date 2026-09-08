package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

type summaryJSON struct {
	Order         string `json:"order_id"`
	ProductName   string `json:"product_name"`
	Status        string `json:"status"`
	Tracking      string `json:"tracking"`
	Total         money  `json:"total"`
	ShopReference string `json:"shop_reference"`
	DepositPaid   bool   `json:"deposit_paid"`
}

func summariesOf(t *testing.T, rec interface{ Bytes() []byte }) []summaryJSON {
	t.Helper()
	var out []summaryJSON
	if err := json.Unmarshal(rec.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v — body %s", err, rec.Bytes())
	}
	return out
}

// seedSummary writes a row the way the projector would, so the HTTP test is
// about the ROUTES: who may read what, and what the view hides.
func (a *api) seedSummary(t *testing.T, customer shared.ID, status ordering.OrderStatus, at time.Time) ordering.OrderID {
	t.Helper()
	id := ordering.NewOrderID()
	err := a.summaries.Save(context.Background(), reportingapp.OrderSummary{
		Order: id, Customer: customer, Product: shared.NewID(), Variant: shared.NewID(),
		ProductName: "Air Trainer 90", Status: status, Tracking: reportingapp.TrackingExpected,
		Total: shared.MustParseMoney("5393720", shared.VND), Deposit: shared.MustParseMoney("2696860", shared.VND),
		DepositPaid: status != ordering.StatusAwaitingDeposit, ShopReference: "NK-20260905-001",
		PlacedAt: at, UpdatedAt: at,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// GET /me/orders is "me" — the token — not a path parameter. There is no way
// to ASK for another customer's orders, which is a stronger guarantee than
// remembering to compare two ids in every handler.
func TestGETMeOrders_showsOnlyTheCallersOwn(t *testing.T) {
	a := newAPI()
	mine := a.seedSummary(t, a.customerID, ordering.StatusDeposited, now)
	newer := a.seedSummary(t, a.customerID, ordering.StatusAwaitingDeposit, now.Add(time.Hour))
	a.seedSummary(t, shared.NewID(), ordering.StatusDelivered, now.Add(2*time.Hour)) // somebody else's

	rec := a.call(t, "GET", "/me/orders", "", a.asCustomer())
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	rows := summariesOf(t, rec.Body)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want only the caller's two", len(rows))
	}
	if rows[0].Order != newer.String() || rows[1].Order != mine.String() {
		t.Errorf("not newest-first: %v", rows)
	}
	if rows[0].ProductName != "Air Trainer 90" || rows[0].Total.Amount != "5393720" {
		t.Errorf("row = %+v", rows[0])
	}
	// The shop's own order number is an internal reference.
	if rows[0].ShopReference != "" {
		t.Errorf("a customer must not see the shop reference, got %q", rows[0].ShopReference)
	}

	expectError(t, a.call(t, "GET", "/me/orders", "", a.asOperator()), http.StatusForbidden, "forbidden")
	expectError(t, a.callAnon(t, "GET", "/me/orders", "", nil), http.StatusUnauthorized, "unauthenticated")
}

// GET /orders?status= is the operator's work queue.
func TestGETOrders_isTheOperatorsWorkQueue(t *testing.T) {
	a := newAPI()
	waiting := a.seedSummary(t, shared.NewID(), ordering.StatusAwaitingDeposit, now)
	a.seedSummary(t, shared.NewID(), ordering.StatusDeposited, now.Add(time.Hour))
	a.seedSummary(t, shared.NewID(), ordering.StatusDelivered, now.Add(2*time.Hour))

	if rows := summariesOf(t, a.call(t, "GET", "/orders", "", nil).Body); len(rows) != 3 {
		t.Fatalf("no filter = %d rows, want all three", len(rows))
	}
	rows := summariesOf(t, a.call(t, "GET", "/orders?status=awaiting_deposit", "", nil).Body)
	if len(rows) != 1 || rows[0].Order != waiting.String() {
		t.Fatalf("status filter = %v", rows)
	}
	// The operator DOES see the shop reference: it is their own record.
	if rows[0].ShopReference != "NK-20260905-001" {
		t.Errorf("shop reference = %q", rows[0].ShopReference)
	}

	// A typo must not look like "nothing to do".
	expectError(t, a.call(t, "GET", "/orders?status=deposted", "", nil), http.StatusBadRequest, "invalid_order")

	expectError(t, a.call(t, "GET", "/orders", "", a.asCustomer()), http.StatusForbidden, "forbidden")
}

// An empty result is [], never null: a client should not have to handle two
// shapes of "nothing".
func TestGETMeOrders_emptyIsAnEmptyArray(t *testing.T) {
	a := newAPI()
	rec := a.call(t, "GET", "/me/orders", "", a.asCustomer())
	if body := rec.Body.String(); body != "[]\n" && body != "[]" {
		t.Errorf("empty list serialised as %q", body)
	}
}
