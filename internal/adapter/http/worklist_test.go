package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/adapter/eventcodec"
	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/contracts"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/shared"
)

type queueRow struct {
	ProductID        string `json:"product_id"`
	Name             string `json:"name"`
	Source           string `json:"source"`
	NextStep         string `json:"next_step"`
	RequestedVariant string `json:"requested_variant"`
	VariantLabel     string `json:"variant_label"`
	Requester        string `json:"requested_by"`
	Published        bool   `json:"published"`
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

// The whole point of the field: what the customer typed reaches the person who
// has to create the real variant. Without it that person guesses which size a
// paid order was for.
func TestProductQueue_carriesTheSizeTheCustomerAskedFor(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)

	body := `{"name":"Air Trainer 90","merchant_id":"` + m.ID().String() + `",
		"category":"footwear","source_url":"https://www.example.com/t/air-trainer-90/abc",
		"price":"150.00","currency":"USD","requested_variant":"  M 8 / W 9.5  "}`
	id := idOf(t, a.call(t, "POST", "/products", body, a.asCustomer()))

	// The projector is what fills the queue, so drive it with the event the
	// paste produced rather than pretending the read model wrote itself.
	relayCatalog(t, a)

	var rows []queueRow
	rec := a.call(t, "GET", "/product-queue", "", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("GET /product-queue = %d %s (%v)", rec.Code, rec.Body, err)
	}
	if len(rows) != 1 || rows[0].ProductID != id {
		t.Fatalf("queue = %+v", rows)
	}
	// Trimmed, because a trailing space is not a different size.
	if rows[0].RequestedVariant != "M 8 / W 9.5" {
		t.Fatalf("requested_variant = %q, want it trimmed and carried through", rows[0].RequestedVariant)
	}
	// And the customer sees their own wish echoed back.
	rec = a.call(t, "GET", "/me/products", "", a.asCustomer())
	rows = nil
	_ = json.Unmarshal(rec.Body.Bytes(), &rows)
	if len(rows) != 1 || rows[0].RequestedVariant != "M 8 / W 9.5" {
		t.Fatalf("my products = %+v", rows)
	}
}

// relayCatalog does for the catalog events in the outbox what worker.Relay
// does in production: encode, decode, hand to the reporting projector.
//
// The read model is written ONLY by events, so a test that wants a queue row
// has to deliver one. Going through the real codec rather than building the
// contract by hand is the point: if a field stops being encoded, this breaks.
func relayCatalog(t *testing.T, a *api) {
	t.Helper()
	ctx := context.Background()
	proj := reportingapp.NewProjector(reportingapp.Deps{
		UoW: memory.UnitOfWork{}, Summaries: a.summaries, Names: a.names, Worklist: a.worklist,
	})
	for _, ev := range a.outbox.Drain() {
		env, err := eventcodec.Encode(ev)
		if err != nil {
			t.Fatalf("encode %s: %v", ev.EventName(), err)
		}
		msg, err := eventcodec.Decode(env.Name, env.Payload)
		if errors.Is(err, eventcodec.ErrNoDecoder) {
			continue // nobody consumes it; not this test's business
		}
		if err != nil {
			t.Fatalf("decode %s: %v", env.Name, err)
		}
		switch m := msg.(type) {
		case contracts.ProductAddedV1:
			err = proj.OnProductAdded(ctx, m)
		case contracts.VariantAddedV1:
			err = proj.OnVariantAdded(ctx, m)
		case contracts.ListingConfirmedV1:
			err = proj.OnListingConfirmed(ctx, m)
		case contracts.ProductMeasuredV1:
			err = proj.OnProductMeasured(ctx, m)
		case contracts.ProductPublishedV1:
			err = proj.OnProductPublished(ctx, m)
		}
		if err != nil {
			t.Fatalf("project %s: %v", env.Name, err)
		}
	}
}
