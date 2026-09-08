package catalogapp_test

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// seedDraft records a draft product through the real use case and drains the
// events it produced, so the test under way starts from a clean outbox.
func (w *world) seedDraft(t *testing.T) catalog.ProductID {
	t.Helper()
	m := w.seedMerchant(t)
	w.seedCategory(t, "footwear")
	id, err := catalogapp.NewAddProductHandler(w.deps).Handle(context.Background(), exampleAddProduct(m))
	if err != nil {
		t.Fatal(err)
	}
	w.outbox.Drain()
	return id
}

// The load → decide → save path. The use case adds nothing of its own here:
// every "no" comes from Product.Publish and is passed through untouched.
func TestPublishProduct_passesDomainRefusalThrough(t *testing.T) {
	w := newWorld()
	id := w.seedDraft(t)
	h := catalogapp.NewPublishProductHandler(w.deps)

	if err := h.Handle(context.Background(), catalog.NewProductID()); !errors.Is(err, catalog.ErrProductNotFound) {
		t.Errorf("unknown product: got %v", err)
	}
	if err := h.Handle(context.Background(), id); !errors.Is(err, catalog.ErrNoVariants) {
		t.Errorf("draft without variants: got %v, want ErrNoVariants", err)
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Errorf("%d events published after a refused publish", len(evs))
	}
}

func TestPublishProduct_savesThenPublishes(t *testing.T) {
	w := newWorld()
	id := w.seedDraft(t)
	p, _ := w.products.ByID(context.Background(), id)
	verified := catalog.MustProvenance(catalog.SourcedByOperator, now, operator)
	if _, err := p.AddVariant(catalog.VariantDetails{Size: "US 9"}, now); err != nil {
		t.Fatal(err)
	}
	if err := p.ConfirmListing(verified); err != nil {
		t.Fatal(err)
	}
	if err := p.Measure(shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), verified, now); err != nil {
		t.Fatal(err)
	}
	p.PullEvents() // those steps will be their own use cases; not under test here
	h := catalogapp.NewPublishProductHandler(w.deps)

	if err := h.Handle(context.Background(), id); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	saved, _ := w.products.ByID(context.Background(), id)
	if !saved.IsPublished() {
		t.Fatal("product must be published")
	}
	evs := w.outbox.Drain()
	if len(evs) != 1 || evs[0].EventName() != "catalog.product_published" {
		t.Fatalf("outbox = %v, want one product_published", evs)
	}
	if !evs[0].OccurredAt().Equal(now) {
		t.Errorf("event time = %v, want the app clock's %v", evs[0].OccurredAt(), now)
	}
}
