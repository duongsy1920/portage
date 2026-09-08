package catalogapp_test

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Categories are reference data other contexts depend on (pricing turns a
// category into a goods class and a default parcel). Saving them straight to
// the repository tells nobody; the use case saves AND announces.
func TestDefineCategory_savesAndAnnounces(t *testing.T) {
	w := newWorld()
	h := catalogapp.NewDefineCategoryHandler(w.deps)

	cmd := catalogapp.DefineCategory{
		Code:         "footwear",
		Estimate:     shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13)),
		Restrictions: []catalog.Restriction{catalog.RestrictionMagnet},
	}
	if err := h.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	saved, err := w.categories.ByCode(context.Background(), catalog.MustParseCategoryCode("footwear"))
	if err != nil || !saved.Restricted(catalog.RestrictionMagnet) {
		t.Fatalf("ByCode = %+v, %v", saved, err)
	}
	evs := w.outbox.Drain()
	if len(evs) != 1 || evs[0].EventName() != "catalog.category_defined" {
		t.Fatalf("outbox = %v, want one category_defined", evs)
	}
	if !evs[0].OccurredAt().Equal(now) {
		t.Errorf("event time = %v, want the app clock's %v", evs[0].OccurredAt(), now)
	}

	// Defining it again REPLACES the value object and announces again — a
	// consumer sees the latest definition; at-least-once means it must cope
	// with seeing the same one twice anyway.
	cmd.Restrictions = nil
	if err := h.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("redefine: %v", err)
	}
	saved, _ = w.categories.ByCode(context.Background(), catalog.MustParseCategoryCode("footwear"))
	if saved.Restricted(catalog.RestrictionMagnet) {
		t.Error("redefinition must replace the policy")
	}
	if evs := w.outbox.Drain(); len(evs) != 1 {
		t.Errorf("redefinition published %d events, want 1", len(evs))
	}
}

func TestDefineCategory_domainRefusalsPassThrough(t *testing.T) {
	w := newWorld()
	h := catalogapp.NewDefineCategoryHandler(w.deps)
	good := shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13))

	if err := h.Handle(context.Background(), catalogapp.DefineCategory{Code: "Giày Dép!", Estimate: good}); !errors.Is(err, catalog.ErrInvalidCategoryCode) {
		t.Errorf("bad code: got %v", err)
	}
	if err := h.Handle(context.Background(), catalogapp.DefineCategory{Code: "footwear"}); !errors.Is(err, shared.ErrIncompleteParcelSpec) {
		t.Errorf("no estimate: got %v", err)
	}
	if err := h.Handle(context.Background(), catalogapp.DefineCategory{Code: "footwear", Estimate: good, Restrictions: []catalog.Restriction{"cursed"}}); !errors.Is(err, catalog.ErrUnknownRestriction) {
		t.Errorf("bad restriction: got %v", err)
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Errorf("%d events after refused commands", len(evs))
	}
}
