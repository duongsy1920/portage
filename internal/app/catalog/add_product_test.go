package catalogapp_test

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

var operator = shared.NewOperatorID()

// seedMerchant puts an active USD merchant in the store, the way a previous
// request would have.
func (w *world) seedMerchant(t *testing.T) *catalog.Merchant {
	t.Helper()
	m, err := catalog.RegisterMerchant(exampleMerchant(), now)
	if err != nil {
		t.Fatal(err)
	}
	m.PullEvents() // already published by the request that created it
	if err := w.merchants.Save(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	return m
}

func (w *world) seedCategory(t *testing.T, code string) catalog.CategoryPolicy {
	t.Helper()
	c, err := catalog.NewCategoryPolicy(
		catalog.MustParseCategoryCode(code),
		shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13)),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.categories.Save(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	return c
}

func exampleAddProduct(m *catalog.Merchant) catalogapp.AddProduct {
	return catalogapp.AddProduct{
		Name:      "Air Trainer 90",
		Merchant:  m.ID(),
		Category:  catalog.MustParseCategoryCode("footwear"),
		Source:    catalog.MustParseSourceURL("https://www.example.com/t/air-trainer-90/abc"),
		Price:     shared.MustParseMoney("150.00", shared.USD),
		SourcedBy: catalog.SourcedByCustomer,
	}
}

// The use case ASSEMBLES what the domain needs: the clock's now and the
// caller's identity become a Provenance; two repositories are consulted for
// rules that span aggregates; then the domain decides.
func TestAddProduct_assemblesProvenanceAndSaves(t *testing.T) {
	w := newWorld()
	m := w.seedMerchant(t)
	w.seedCategory(t, "footwear")
	h := catalogapp.NewAddProductHandler(w.deps)

	id, err := h.Handle(context.Background(), exampleAddProduct(m))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	p, err := w.products.ByID(context.Background(), id)
	if err != nil {
		t.Fatalf("ByID: %v", err)
	}
	if p.Status() != catalog.ProductStatusDraft || p.Merchant() != m.ID() {
		t.Errorf("saved product: status %v, merchant %v", p.Status(), p.Merchant())
	}
	prov := p.ListingProvenance()
	if prov.Source() != catalog.SourcedByCustomer || !prov.At().Equal(now) || prov.Verified() {
		t.Errorf("listing provenance = %v; want customer, at the app clock, unverified", prov)
	}
	if evs := w.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "catalog.product_added" {
		t.Fatalf("outbox = %v, want one product_added", evs)
	}
}

// Rules that need TWO aggregates live in the use case, and each has its own
// error so the adapter can answer precisely.
func TestAddProduct_crossAggregateRules(t *testing.T) {
	w := newWorld()
	m := w.seedMerchant(t)
	w.seedCategory(t, "footwear")
	h := catalogapp.NewAddProductHandler(w.deps)

	cases := map[string]struct {
		mutate func(*catalogapp.AddProduct)
		want   error
	}{
		"unknown merchant": {func(c *catalogapp.AddProduct) { c.Merchant = catalog.NewMerchantID() }, catalog.ErrMerchantNotFound},
		"unknown category": {func(c *catalogapp.AddProduct) { c.Category = catalog.MustParseCategoryCode("luggage") }, catalog.ErrCategoryNotFound},
		"price in another currency": {func(c *catalogapp.AddProduct) {
			c.Price = shared.MustParseMoney("3900000", shared.VND)
		}, catalogapp.ErrPriceCurrency},
		"operator claim without operator": {func(c *catalogapp.AddProduct) {
			c.SourcedBy = catalog.SourcedByOperator
		}, catalog.ErrOperatorRequired},
		"domain refusal passes through": {func(c *catalogapp.AddProduct) { c.Name = "  " }, catalog.ErrEmptyName},
	}
	for name, c := range cases {
		cmd := exampleAddProduct(m)
		c.mutate(&cmd)
		if _, err := h.Handle(context.Background(), cmd); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
	if n := w.products.Len(); n != 0 {
		t.Errorf("%d products saved after refused commands", n)
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Errorf("%d events published after refused commands", len(evs))
	}
}

// Nothing new is listed for a shop we cannot buy from.
func TestAddProduct_refusesSuspendedMerchant(t *testing.T) {
	w := newWorld()
	m := w.seedMerchant(t)
	w.seedCategory(t, "footwear")
	if err := m.Suspend("account banned", now); err != nil {
		t.Fatal(err)
	}
	h := catalogapp.NewAddProductHandler(w.deps)

	if _, err := h.Handle(context.Background(), exampleAddProduct(m)); !errors.Is(err, catalogapp.ErrMerchantInactive) {
		t.Fatalf("got %v, want ErrMerchantInactive", err)
	}
}

// Option C (CATALOG.md §7) running for real: the same page recorded twice
// creates a second product AND flags it, with the detector's reason — the
// operator queue gets both the product and the event.
func TestAddProduct_flagsSuspectedDuplicate(t *testing.T) {
	w := newWorld()
	m := w.seedMerchant(t)
	w.seedCategory(t, "footwear")
	h := catalogapp.NewAddProductHandler(w.deps)

	first, err := h.Handle(context.Background(), exampleAddProduct(m))
	if err != nil {
		t.Fatal(err)
	}
	w.outbox.Drain()

	second, err := h.Handle(context.Background(), exampleAddProduct(m))
	if err != nil {
		t.Fatalf("second Handle: %v", err)
	}
	if second == first {
		t.Fatal("a suspected duplicate is still a new product")
	}
	p, _ := w.products.ByID(context.Background(), second)
	if of, flagged := p.SuspectedDuplicateOf(); !flagged || of != first {
		t.Fatalf("SuspectedDuplicateOf() = %v, %v; want the first product", of, flagged)
	}
	evs := w.outbox.Drain()
	if len(evs) != 2 || evs[0].EventName() != "catalog.product_added" || evs[1].EventName() != "catalog.product_flagged_duplicate" {
		t.Fatalf("outbox = %v", evs)
	}
	if flag, ok := evs[1].(catalog.ProductFlaggedDuplicate); !ok || flag.Reason != "same page url" || flag.Of != first {
		t.Errorf("flag event = %+v", evs[1])
	}
}
