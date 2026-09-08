package catalog_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// A snapshot is the aggregate's state with the door open: everything a store
// needs to write it down and build it back, nothing more. Round trip is the
// whole contract — after Snapshot → FromSnapshot → Snapshot, nothing may have
// changed, and no event may have been re-raised.
func TestMerchant_snapshotRoundTrip(t *testing.T) {
	m := aMerchant(t)
	if err := m.Rename("Example Athletics", testNow); err != nil {
		t.Fatal(err)
	}
	if err := m.EnableSourcing(catalog.SourcedByCustomer, testNow); err != nil {
		t.Fatal(err)
	}
	if err := m.ChangeFreeShipping(catalog.AlwaysFreeShipping(), testNow); err != nil {
		t.Fatal(err)
	}
	if err := m.Suspend("account banned", testNow); err != nil {
		t.Fatal(err)
	}
	m.PullEvents()

	snap := m.Snapshot()
	back, err := catalog.MerchantFromSnapshot(snap)
	if err != nil {
		t.Fatalf("MerchantFromSnapshot: %v", err)
	}
	if !reflect.DeepEqual(back.Snapshot(), snap) {
		t.Fatalf("round trip changed the state:\n got %+v\nwant %+v", back.Snapshot(), snap)
	}
	if back.ID() != m.ID() || back.IsActive() || back.Name() != "Example Athletics" || !back.Supports(catalog.SourcedByCustomer) {
		t.Errorf("rehydrated merchant lost behaviour: %+v", back.Snapshot())
	}
	if evs := back.PullEvents(); len(evs) != 0 {
		t.Errorf("rehydration raised %d events; loading is not a business fact", len(evs))
	}
	// The snapshot is a copy: mutating it must not reach the aggregate.
	snap.Sourcing[0] = "telepathy"
	if !m.Supports(catalog.SourcedByOperator) {
		t.Error("snapshot shares memory with the aggregate")
	}
}

func TestProduct_snapshotRoundTrip(t *testing.T) {
	p := aProduct(t)
	verified := operatorChecked()
	addVariant(t, p, "US 9", "black")
	addVariant(t, p, "US 10", "black")
	if err := p.ConfirmListing(verified); err != nil {
		t.Fatal(err)
	}
	if err := p.Measure(shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), verified, testNow); err != nil {
		t.Fatal(err)
	}
	if err := p.Reprice(shared.MustParseMoney("160.00", shared.USD), verified, testNow); err != nil {
		t.Fatal(err)
	}
	other := catalog.NewProductID()
	if err := p.FlagDuplicateOf(other, "same page url", testNow); err != nil {
		t.Fatal(err)
	}
	if err := p.ClearDuplicateFlag("checked", testNow); err != nil {
		t.Fatal(err)
	}
	if err := p.Publish(testNow); err != nil {
		t.Fatal(err)
	}
	p.PullEvents()

	snap := p.Snapshot()
	back, err := catalog.ProductFromSnapshot(snap)
	if err != nil {
		t.Fatalf("ProductFromSnapshot: %v", err)
	}
	if !reflect.DeepEqual(back.Snapshot(), snap) {
		t.Fatalf("round trip changed the state:\n got %+v\nwant %+v", back.Snapshot(), snap)
	}
	if !back.IsPublished() || len(back.Variants()) != 2 || back.Price().String() != "160.00 USD" {
		t.Errorf("rehydrated product lost behaviour: %+v", back.Snapshot())
	}
	// The operator's past decision survives a reload: the dismissed pair is
	// still not flaggable.
	if err := back.Retire("x", testNow); err != nil {
		t.Fatal(err)
	}
	back.PullEvents()
	if err := back.FlagDuplicateOf(other, "same page url", testNow); err != nil {
		t.Fatal(err)
	}
	if _, flagged := back.SuspectedDuplicateOf(); flagged {
		t.Error("a dismissed duplicate was flagged again after rehydration")
	}
	if evs := back.PullEvents(); len(evs) != 0 {
		t.Errorf("got %d events, want none", len(evs))
	}
}

// A draft with nothing measured must round-trip too: zero parcel, zero
// provenance, no variants, no flags.
func TestProduct_snapshotRoundTripOfABareDraft(t *testing.T) {
	p := aProduct(t)
	p.PullEvents()
	snap := p.Snapshot()
	back, err := catalog.ProductFromSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back.Snapshot(), snap) {
		t.Fatalf("round trip changed the state:\n got %+v\nwant %+v", back.Snapshot(), snap)
	}
	if _, measured := back.ParcelSpec(); measured {
		t.Error("a bare draft has no parcel")
	}
}

// FromSnapshot trusts the store on business values — they were validated on
// the way in — but refuses a shape that cannot have come from Snapshot():
// that is corruption, not data.
func TestFromSnapshot_rejectsCorruptShapes(t *testing.T) {
	m := aMerchant(t)
	ms := m.Snapshot()
	ms.ID = catalog.MerchantID{}
	if _, err := catalog.MerchantFromSnapshot(ms); !errors.Is(err, catalog.ErrInvalidSnapshot) {
		t.Errorf("merchant without id: got %v", err)
	}
	ms = m.Snapshot()
	ms.Status = "bananas"
	if _, err := catalog.MerchantFromSnapshot(ms); !errors.Is(err, catalog.ErrInvalidSnapshot) {
		t.Errorf("merchant with unknown status: got %v", err)
	}

	p := aProduct(t)
	ps := p.Snapshot()
	ps.Parcel = shared.MustParcelSpec(shared.Grams(1), shared.NewDimensionsCM(1, 1, 1)) // parcel without provenance
	if _, err := catalog.ProductFromSnapshot(ps); !errors.Is(err, catalog.ErrInvalidSnapshot) {
		t.Errorf("parcel without provenance: got %v", err)
	}
	ps = p.Snapshot()
	ps.Variants = []catalog.VariantSnapshot{
		{ID: catalog.NewVariantID(), Size: "US 9", Color: "black", AddedAt: testNow},
		{ID: catalog.NewVariantID(), Size: " us 9 ", Color: "BLACK", AddedAt: testNow},
	}
	if _, err := catalog.ProductFromSnapshot(ps); !errors.Is(err, catalog.ErrInvalidSnapshot) {
		t.Errorf("duplicate variants: got %v", err)
	}
	ps = p.Snapshot()
	ps.Status = "published"
	ps.Variants = nil // published without variants cannot have happened
	if _, err := catalog.ProductFromSnapshot(ps); !errors.Is(err, catalog.ErrInvalidSnapshot) {
		t.Errorf("published without variants: got %v", err)
	}
}

// The store decomposes FreeShipping into (kind, threshold); Kind() is the
// half that was missing.
func TestFreeShipping_kind(t *testing.T) {
	if catalog.NoFreeShipping().Kind() != catalog.FreeShipNever ||
		catalog.AlwaysFreeShipping().Kind() != catalog.FreeShipAlways ||
		catalog.MustFreeShippingOver(shared.MustParseMoney("50.00", shared.USD)).Kind() != catalog.FreeShipOver {
		t.Fatal("Kind() does not match the constructor")
	}
	if got := string(catalog.FreeShipOver); got != "over" {
		t.Errorf("kinds are persisted as text; got %q", got)
	}
}
