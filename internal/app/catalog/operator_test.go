package catalogapp_test

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The three operator steps between "pasted by a customer" and "published":
// a variant, a confirmed listing, a measured parcel. Together they are the
// ONLY way a product becomes publishable through the application.
func TestOperatorSteps_makeAProductPublishable(t *testing.T) {
	w := newWorld()
	ctx := context.Background()
	id := w.seedDraft(t)
	operator := shared.NewOperatorID()
	w.outbox.Drain()

	publish := catalogapp.NewPublishProductHandler(w.deps)
	if err := publish.Handle(ctx, id); !errors.Is(err, catalog.ErrNoVariants) {
		t.Fatalf("publish before any operator step = %v, want ErrNoVariants", err)
	}

	vid, err := catalogapp.NewAddVariantHandler(w.deps).Handle(ctx, catalogapp.AddVariant{Product: id, Size: "US 9", Color: "black"})
	if err != nil || vid.IsZero() {
		t.Fatalf("AddVariant = %v, %v", vid, err)
	}
	_, err = catalogapp.NewAddVariantHandler(w.deps).Handle(ctx, catalogapp.AddVariant{Product: id, Size: "us 9", Color: "Black"})
	if !errors.Is(err, catalog.ErrDuplicateVariant) {
		t.Fatalf("same variant twice = %v, want ErrDuplicateVariant", err)
	}
	if err := publish.Handle(ctx, id); !errors.Is(err, catalog.ErrUnverified) {
		t.Fatalf("publish with a variant but nothing verified = %v, want ErrUnverified", err)
	}

	confirm := catalogapp.NewConfirmListingHandler(w.deps)
	if err := confirm.Handle(ctx, catalogapp.ConfirmListing{Product: id}); !errors.Is(err, catalog.ErrOperatorRequired) {
		t.Fatalf("confirm without an operator = %v, want ErrOperatorRequired", err)
	}
	if err := confirm.Handle(ctx, catalogapp.ConfirmListing{Product: id, Operator: operator}); err != nil {
		t.Fatal(err)
	}

	spec := shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))
	measure := catalogapp.NewMeasureProductHandler(w.deps)
	if err := measure.Handle(ctx, catalogapp.MeasureProduct{Product: id, Spec: spec}); !errors.Is(err, catalog.ErrOperatorRequired) {
		t.Fatalf("measure without an operator = %v, want ErrOperatorRequired", err)
	}
	if err := measure.Handle(ctx, catalogapp.MeasureProduct{Product: id, Spec: spec, Operator: operator}); err != nil {
		t.Fatal(err)
	}
	// Two events now: adding the variant announced WHICH SIZE exists, because
	// procurement has to buy that size and may not read catalog to find it.
	evs := w.outbox.Drain()
	if len(evs) != 2 || evs[0].EventName() != "catalog.variant_added" || evs[1].EventName() != "catalog.product_measured" {
		t.Fatalf("outbox = %v, want variant_added then product_measured", evs)
	}
	if va, ok := evs[0].(catalog.VariantAdded); !ok || va.Size != "US 9" || va.Color != "black" {
		t.Fatalf("variant_added must carry the shop's own words: %+v", evs[0])
	}

	if err := publish.Handle(ctx, id); err != nil {
		t.Fatalf("publish after the three steps: %v", err)
	}
	p, _ := w.products.ByID(ctx, id)
	if !p.IsPublished() || len(p.Variants()) != 1 {
		t.Fatalf("product = %s with %d variants", p.Status(), len(p.Variants()))
	}
	if evs := w.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "catalog.product_published" {
		t.Fatalf("outbox after publishing = %v", evs)
	}

	if _, err := catalogapp.NewAddVariantHandler(w.deps).Handle(ctx, catalogapp.AddVariant{Product: catalog.NewProductID(), Size: "M"}); !errors.Is(err, catalog.ErrProductNotFound) {
		t.Fatalf("unknown product = %v, want ErrProductNotFound", err)
	}
}
