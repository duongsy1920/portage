package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/shared"
)

type queueRow struct {
	ProductID    string `json:"product_id"`
	Name         string `json:"name"`
	Source       string `json:"source"`
	NextStep     string `json:"next_step"`
	VariantLabel string `json:"variant_label"`
	Requester    string `json:"requested_by"`
	Published    bool   `json:"published"`
}

func (a *api) seedWorklist(t *testing.T, item reportingapp.WorklistItem) {
	t.Helper()
	if err := a.worklist.Save(context.Background(), item); err != nil {
		t.Fatal(err)
	}
}

// The staff queue: unfinished work, and the next step spelled out so a screen
// never has to guess which of four buttons to offer.
func TestProductQueue_showsWhatIsLeftToDo(t *testing.T) {
	a := newAPI()
	waiting := a.customerID
	a.seedWorklist(t, reportingapp.WorklistItem{
		Product: shared.NewID(), Name: "Air Trainer 90", Source: "https://www.example.com/t/x",
		Price: shared.MustParseMoney("150.00", shared.USD), SourcedBy: "customer", RequestedBy: waiting,
		Variants: []reportingapp.WorklistVariant{{ID: shared.NewID(), Label: "US 9 · black"}},
		AddedAt:  now, UpdatedAt: now,
	})
	// Published work is done and must not clutter the queue.
	a.seedWorklist(t, reportingapp.WorklistItem{
		Product: shared.NewID(), Name: "Old Thing", Published: true, AddedAt: now, UpdatedAt: now,
		Price: shared.MustParseMoney("10.00", shared.USD),
	})

	var rows []queueRow
	rec := a.call(t, "GET", "/product-queue", "", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("GET /product-queue = %d %s (%v)", rec.Code, rec.Body, err)
	}
	if len(rows) != 1 {
		t.Fatalf("queue = %+v, want only the unfinished row", rows)
	}
	if rows[0].NextStep != "listing" || rows[0].VariantLabel != "US 9 · black" {
		t.Fatalf("row = %+v", rows[0])
	}
	if rows[0].Source == "" {
		t.Fatal("a queue row without the page to open is not something a person can act on")
	}

	// It is the STAFF queue: a customer must not read the whole workload.
	expectError(t, a.call(t, "GET", "/product-queue", "", a.asCustomer()), 403, "forbidden")
}

// GET /me/products cannot name anybody else, exactly like GET /me/orders: the
// id comes from the token, so there is no path that leaks another list.
func TestMyProducts_isScopedToTheToken(t *testing.T) {
	a := newAPI()
	theirs := shared.NewID()
	a.seedWorklist(t, reportingapp.WorklistItem{
		Product: shared.NewID(), Name: "Mine", RequestedBy: a.customerID,
		Price: shared.MustParseMoney("150.00", shared.USD), AddedAt: now, UpdatedAt: now,
	})
	a.seedWorklist(t, reportingapp.WorklistItem{
		Product: shared.NewID(), Name: "Somebody else's", RequestedBy: theirs,
		Price: shared.MustParseMoney("150.00", shared.USD), AddedAt: now, UpdatedAt: now,
	})
	// And one an operator added on spec: nobody is waiting for it.
	a.seedWorklist(t, reportingapp.WorklistItem{
		Product: shared.NewID(), Name: "On spec", SourcedBy: "operator",
		Price: shared.MustParseMoney("150.00", shared.USD), AddedAt: now, UpdatedAt: now,
	})

	var rows []queueRow
	rec := a.call(t, "GET", "/me/products", "", a.asCustomer())
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("GET /me/products = %d %s (%v)", rec.Code, rec.Body, err)
	}
	if len(rows) != 1 || rows[0].Name != "Mine" {
		t.Fatalf("my products = %+v", rows)
	}
	// An operator has no "my products": the route is about who is waiting.
	expectError(t, a.call(t, "GET", "/me/products", "", nil), 403, "forbidden")
}
