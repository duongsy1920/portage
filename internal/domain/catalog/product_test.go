package catalog_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// What a customer pasting a product page gives us: everything unverified.
func customerSaid() catalog.Provenance {
	return catalog.MustProvenance(catalog.SourcedByCustomer, provAt, shared.OperatorID{})
}

// What an operator confirms after looking (or weighing) for themselves.
func operatorChecked() catalog.Provenance {
	return catalog.MustProvenance(catalog.SourcedByOperator, provAt, someone)
}

func exampleProduct() catalog.ProductDetails {
	return catalog.ProductDetails{
		Name:              "Air Trainer 90",
		Merchant:          catalog.NewMerchantID(),
		Category:          catalog.MustParseCategoryCode("footwear"),
		Source:            catalog.MustParseSourceURL("https://www.example.com/t/air-trainer-90/abc"),
		Price:             shared.MustParseMoney("150.00", shared.USD),
		ListingProvenance: customerSaid(),
		PriceProvenance:   customerSaid(),
	}
}

func aProduct(t *testing.T) *catalog.Product {
	t.Helper()
	p, err := catalog.AddProduct(exampleProduct(), testNow)
	if err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	return p
}

func addVariant(t *testing.T, p *catalog.Product, size, color string) catalog.VariantID {
	t.Helper()
	id, err := p.AddVariant(catalog.VariantDetails{Size: size, Color: color, MerchantRef: "ref-" + size}, testNow)
	if err != nil {
		t.Fatalf("AddVariant(%q, %q): %v", size, color, err)
	}
	return id
}

// The catalogue grows from what customers ask for (CATALOG.md §0): a product
// starts as a DRAFT built from unverified data, with an identity from line one.
func TestAddProduct_startsAsDraft(t *testing.T) {
	p := aProduct(t)

	if p.ID().IsZero() {
		t.Fatal("product must have an id before it is saved")
	}
	if p.Status() != catalog.ProductStatusDraft || p.IsPublished() {
		t.Fatalf("Status() = %v, want draft", p.Status())
	}
	if p.Name() != "Air Trainer 90" || p.Category().String() != "footwear" || p.Price().String() != "150.00 USD" {
		t.Errorf("getters: %q %q %s", p.Name(), p.Category(), p.Price())
	}
	if p.Source().Host().String() != "www.example.com" {
		t.Errorf("Source().Host() = %q", p.Source().Host())
	}
	if _, measured := p.ParcelSpec(); measured {
		t.Error("a new product has not been measured")
	}
	if len(p.Variants()) != 0 {
		t.Error("a new product has no variants")
	}
	if evs := p.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.product_added" {
		t.Fatalf("got %v, want one product_added", evs)
	}
}

// Every required field is checked (convention 10). Two provenances are
// required even for a draft: data with no origin cannot be trusted OR
// distrusted later.
func TestAddProduct_rejectsBadInput(t *testing.T) {
	cases := map[string]struct {
		mutate func(*catalog.ProductDetails)
		want   error
	}{
		"empty name":        {func(d *catalog.ProductDetails) { d.Name = "  " }, catalog.ErrEmptyName},
		"zero merchant":     {func(d *catalog.ProductDetails) { d.Merchant = catalog.MerchantID{} }, catalog.ErrMerchantRequired},
		"zero category":     {func(d *catalog.ProductDetails) { d.Category = catalog.CategoryCode{} }, catalog.ErrInvalidCategoryCode},
		"zero source":       {func(d *catalog.ProductDetails) { d.Source = catalog.SourceURL{} }, catalog.ErrInvalidSourceURL},
		"zero price":        {func(d *catalog.ProductDetails) { d.Price = shared.Money{} }, catalog.ErrPriceRequired},
		"negative price":    {func(d *catalog.ProductDetails) { d.Price = shared.NewMoney(-1, shared.USD) }, catalog.ErrNegativePrice},
		"no listing origin": {func(d *catalog.ProductDetails) { d.ListingProvenance = catalog.Provenance{} }, catalog.ErrProvenanceRequired},
		"no price origin":   {func(d *catalog.ProductDetails) { d.PriceProvenance = catalog.Provenance{} }, catalog.ErrProvenanceRequired},
	}
	for name, c := range cases {
		d := exampleProduct()
		c.mutate(&d)
		if _, err := catalog.AddProduct(d, testNow); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
}

// The price of a details struct (convention 10): a forgotten field compiles.
// So every field is zeroed ONE AT A TIME on otherwise-valid details, and
// AddProduct must refuse each. A new field either gets validated or is listed
// here with the reason its zero value is a legitimate answer — a conscious
// choice either way. (A single all-zero struct would prove nothing: it dies on
// the first check and never reaches the others.)
func TestAddProduct_everyFieldIsValidated(t *testing.T) {
	zeroIsMeaningful := map[string]string{
		// An operator may add a product on spec, before any customer asks for
		// it. Demanding a requester would make that impossible, and a made-up
		// id would be worse: the worklist would promise to tell somebody.
		"RequestedBy": "an operator may add a product nobody has asked for yet",
		// A one-size product has nothing to ask for, and an operator pasting
		// on spec is not asking for anything either. Demanding it would make
		// both impossible.
		"RequestedVariant": "a one-size product has no size to ask for",
	}

	typ := reflect.TypeOf(exampleProduct())
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		if why, ok := zeroIsMeaningful[name]; ok {
			t.Logf("skip %s: %s", name, why)
			continue
		}
		d := exampleProduct()
		reflect.ValueOf(&d).Elem().Field(i).SetZero()
		if _, err := catalog.AddProduct(d, testNow); err == nil {
			t.Errorf("AddProduct accepted ProductDetails with zero %s", name)
		}
	}
}

// Variant is a CHILD ENTITY: it has an id, but it is created, found and
// changed only through the Product — nobody holds a *Variant.
func TestProduct_addVariant(t *testing.T) {
	p := aProduct(t)

	id := addVariant(t, p, "US 9", "black")
	if id.IsZero() {
		t.Fatal("variant must have an id")
	}
	vs := p.Variants()
	if len(vs) != 1 || vs[0].ID() != id || vs[0].Size() != "US 9" || vs[0].Color() != "black" || vs[0].MerchantRef() != "ref-US 9" {
		t.Fatalf("Variants() = %+v", vs)
	}

	// Same size and colour, different spelling: the same variant, refused.
	if _, err := p.AddVariant(catalog.VariantDetails{Size: " us 9 ", Color: "BLACK"}, testNow); !errors.Is(err, catalog.ErrDuplicateVariant) {
		t.Errorf("duplicate: got %v, want ErrDuplicateVariant", err)
	}
	addVariant(t, p, "US 10", "black")
	if n := len(p.Variants()); n != 2 {
		t.Fatalf("got %d variants, want 2", n)
	}

	// Returned slice is a copy (the same trap as Merchant.Sourcing).
	vs = p.Variants()
	vs[0] = catalog.Variant{}
	if p.Variants()[0].ID().IsZero() {
		t.Fatal("mutating the returned slice changed the product")
	}

	// Spacing is not part of a size: shops write the same shoe both ways, and
	// two rows for one physical size is one the buyer can pick by mistake.
	p2 := aProduct(t)
	addVariant(t, p2, "M 8 / W 9.5", "black")
	if _, err := p2.AddVariant(catalog.VariantDetails{Size: "M8/W9.5", Color: "BLACK"}, testNow); !errors.Is(err, catalog.ErrDuplicateVariant) {
		t.Errorf("same size written without spaces: got %v, want ErrDuplicateVariant", err)
	}
	// Reordering stays a different variant on purpose (see variantKey).
	if _, err := p2.AddVariant(catalog.VariantDetails{Size: "8 M / 9.5 W", Color: "black"}, testNow); err != nil {
		t.Errorf("reordered size must NOT be merged: %v", err)
	}

	// A nameless variant beside named ones is an operator who forgot the size,
	// and the buyer would be sent to a shop with no size to ask for.
	if _, err := p2.AddVariant(catalog.VariantDetails{}, testNow); !errors.Is(err, catalog.ErrUnnamedVariant) {
		t.Errorf("nameless variant beside named ones: got %v, want ErrUnnamedVariant", err)
	}

	// A one-size product has exactly one variant with nothing to say.
	single := aProduct(t)
	if _, err := single.AddVariant(catalog.VariantDetails{}, testNow); err != nil {
		t.Fatalf("attribute-less variant: %v", err)
	}
	if _, err := single.AddVariant(catalog.VariantDetails{}, testNow); !errors.Is(err, catalog.ErrDuplicateVariant) {
		t.Errorf("second attribute-less variant: got %v", err)
	}
	// …and nothing named may join it either: same rule, other direction.
	if _, err := single.AddVariant(catalog.VariantDetails{Size: "US 9"}, testNow); !errors.Is(err, catalog.ErrUnnamedVariant) {
		t.Errorf("named variant beside a nameless one: got %v, want ErrUnnamedVariant", err)
	}

	// The event carries the shop's words untouched, slashes and spaces and all.
	p3 := aProduct(t)
	p3.PullEvents()
	vid, err := p3.AddVariant(catalog.VariantDetails{Size: " M 8 / W 9.5 ", Color: "Sail/Gum", MerchantRef: "EX-AT90-8-BLK"}, testNow)
	if err != nil {
		t.Fatal(err)
	}
	evs := p3.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "catalog.variant_added" {
		t.Fatalf("got %v, want one variant_added", evs)
	}
	va, ok := evs[0].(catalog.VariantAdded)
	if !ok || va.Variant != vid || va.Size != "M 8 / W 9.5" || va.Color != "Sail/Gum" || va.MerchantRef != "EX-AT90-8-BLK" {
		t.Fatalf("VariantAdded = %+v", evs[0])
	}
}

// Measuring is the moment the catalogue earns its data (CATALOG.md §3):
// the real box, on our scale, with a provenance that says so.
func TestProduct_measure(t *testing.T) {
	p := aProduct(t)
	p.PullEvents()
	box := shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))

	if err := p.Measure(box, operatorChecked(), testNow); err != nil {
		t.Fatalf("Measure: %v", err)
	}
	got, measured := p.ParcelSpec()
	if !measured || got != box {
		t.Fatalf("Parcel() = %v, %v", got, measured)
	}
	if prov, ok := p.ParcelProvenance(); !ok || !prov.Verified() {
		t.Errorf("ParcelProvenance() = %v, %v", prov, ok)
	}
	if evs := p.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.product_measured" {
		t.Fatalf("got %v, want one product_measured", evs)
	}

	if err := p.Measure(shared.ParcelSpec{}, operatorChecked(), testNow); !errors.Is(err, shared.ErrIncompleteParcelSpec) {
		t.Errorf("zero spec: got %v", err)
	}
	if err := p.Measure(box, catalog.Provenance{}, testNow); !errors.Is(err, catalog.ErrProvenanceRequired) {
		t.Errorf("no provenance: got %v", err)
	}
}

// A shop's price moves; the product follows without losing its identity, and
// only a real change is announced. Quotes already issued keep their own
// snapshot (DDD.md §28) — this event is for the NEXT quote.
func TestProduct_reprice(t *testing.T) {
	p := aProduct(t)
	p.PullEvents()

	same := shared.MustParseMoney("150.00", shared.USD)
	if err := p.Reprice(same, customerSaid(), testNow); err != nil {
		t.Fatalf("Reprice(same): %v", err)
	}
	if evs := p.PullEvents(); len(evs) != 0 {
		t.Fatalf("an unchanged price raised %d events", len(evs))
	}

	if err := p.Reprice(shared.MustParseMoney("160.00", shared.USD), operatorChecked(), testNow); err != nil {
		t.Fatalf("Reprice: %v", err)
	}
	if p.Price().String() != "160.00 USD" || !p.PriceProvenance().Verified() {
		t.Errorf("Price() = %s, prov = %v", p.Price(), p.PriceProvenance())
	}
	if evs := p.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.product_repriced" {
		t.Fatalf("got %v, want one product_repriced", evs)
	}

	if err := p.Reprice(shared.MustParseMoney("3900000", shared.VND), operatorChecked(), testNow); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("currency change: got %v", err)
	}
	if err := p.Reprice(shared.NewMoney(-1, shared.USD), operatorChecked(), testNow); !errors.Is(err, catalog.ErrNegativePrice) {
		t.Errorf("negative: got %v", err)
	}
	if err := p.Reprice(same, catalog.Provenance{}, testNow); !errors.Is(err, catalog.ErrProvenanceRequired) {
		t.Errorf("no provenance: got %v", err)
	}
	if p.Price().String() != "160.00 USD" {
		t.Error("a refused Reprice must leave the price alone")
	}
}

// THE invariant of this aggregate (CATALOG.md §4): a product is published —
// available for confident quotes — only when a person stands behind what it
// is and what it weighs, it has something to order, and it is not a suspected
// copy of something we already have. Each missing piece names itself.
func TestProduct_publishRequiresEverything(t *testing.T) {
	p := aProduct(t)
	p.PullEvents()

	if err := p.Publish(testNow); !errors.Is(err, catalog.ErrNoVariants) {
		t.Fatalf("no variants: got %v", err)
	}
	addVariant(t, p, "US 9", "black")

	if err := p.Publish(testNow); !errors.Is(err, catalog.ErrUnverified) {
		t.Fatalf("customer-supplied listing: got %v", err)
	}
	if err := p.ConfirmListing(customerSaid(), testNow); !errors.Is(err, catalog.ErrUnverified) {
		t.Fatalf("confirming with an unverified provenance: got %v", err)
	}
	if err := p.ConfirmListing(operatorChecked(), testNow); err != nil {
		t.Fatalf("ConfirmListing: %v", err)
	}

	if err := p.Publish(testNow); !errors.Is(err, catalog.ErrUnverified) {
		t.Fatalf("never measured: got %v", err)
	}
	box := shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))
	if err := p.Measure(box, customerSaid(), testNow); err != nil {
		t.Fatal(err)
	}
	if err := p.Publish(testNow); !errors.Is(err, catalog.ErrUnverified) {
		t.Fatalf("customer-guessed box: got %v", err)
	}
	if err := p.Measure(box, operatorChecked(), testNow); err != nil {
		t.Fatal(err)
	}

	other := catalog.NewProductID()
	if err := p.FlagDuplicateOf(other, "same page url", testNow); err != nil {
		t.Fatal(err)
	}
	if err := p.Publish(testNow); !errors.Is(err, catalog.ErrSuspectedDuplicate) {
		t.Fatalf("suspected duplicate: got %v", err)
	}
	if err := p.ClearDuplicateFlag("checked: different model, same page", testNow); err != nil {
		t.Fatal(err)
	}

	p.PullEvents()
	if err := p.Publish(testNow); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if !p.IsPublished() || p.Status() != catalog.ProductStatusPublished {
		t.Fatal("must be published")
	}
	evs := p.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "catalog.product_published" {
		t.Fatalf("got %v, want one product_published", evs)
	}
	// A listener in another context cannot load this product (guard 7): the
	// event must carry everything pricing needs to quote it.
	pub, ok := evs[0].(catalog.ProductPublished)
	if !ok || pub.Name != "Air Trainer 90" || pub.Price != p.Price() || pub.Parcel != box {
		t.Fatalf("ProductPublished must carry name, price and parcel: %+v", evs[0])
	}
	// And the page itself: whoever has to buy this cannot ask catalog for it.
	if pub.Source != p.Source() {
		t.Errorf("ProductPublished.Source = %q, want %q", pub.Source, p.Source())
	}
	if err := p.Publish(testNow); !errors.Is(err, catalog.ErrNotDraft) {
		t.Errorf("publish twice: got %v", err)
	}
}

// Option C from CATALOG.md §7: a possible duplicate is created anyway,
// flagged, and held back from publishing until an operator decides. Both the
// flag and the decision are recorded WITH a reason: the queue reads the flag,
// the audit trail reads the decision.
func TestProduct_duplicateFlag(t *testing.T) {
	p := aProduct(t)
	p.PullEvents()
	other := catalog.NewProductID()

	if err := p.FlagDuplicateOf(p.ID(), "same page url", testNow); !errors.Is(err, catalog.ErrInvalidDuplicate) {
		t.Errorf("self: got %v", err)
	}
	if err := p.FlagDuplicateOf(catalog.ProductID{}, "same page url", testNow); !errors.Is(err, catalog.ErrInvalidDuplicate) {
		t.Errorf("zero id: got %v", err)
	}
	if err := p.FlagDuplicateOf(other, "  ", testNow); !errors.Is(err, catalog.ErrEmptyReason) {
		t.Errorf("no reason: got %v", err)
	}
	if _, flagged := p.SuspectedDuplicateOf(); flagged {
		t.Fatal("a refused flag must change nothing")
	}

	if err := p.FlagDuplicateOf(other, "same page url", testNow); err != nil {
		t.Fatal(err)
	}
	if of, flagged := p.SuspectedDuplicateOf(); !flagged || of != other {
		t.Fatalf("SuspectedDuplicateOf() = %v, %v", of, flagged)
	}
	if err := p.FlagDuplicateOf(other, "same page url", testNow); err != nil {
		t.Fatalf("flagging the same twice must be a no-op, got %v", err)
	}
	if evs := p.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.product_flagged_duplicate" {
		t.Fatalf("got %v, want one product_flagged_duplicate", evs)
	}

	if err := p.ClearDuplicateFlag("   ", testNow); !errors.Is(err, catalog.ErrEmptyReason) {
		t.Errorf("clearing without a reason: got %v", err)
	}
	if err := p.ClearDuplicateFlag("checked: different model, same page", testNow); err != nil {
		t.Fatal(err)
	}
	if _, flagged := p.SuspectedDuplicateOf(); flagged {
		t.Fatal("flag must be cleared")
	}
	if err := p.ClearDuplicateFlag("again", testNow); err != nil {
		t.Fatalf("clearing an unflagged product must be a no-op, got %v", err)
	}
	if evs := p.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.product_duplicate_cleared" {
		t.Fatalf("got %v, want one product_duplicate_cleared", evs)
	}
}

// An operator's decision is not overridden by the next automated pass: once a
// pair has been cleared, flagging the SAME pair again is a no-op. Otherwise the
// dedup run flags, the operator clears, the run flags again — forever.
func TestProduct_clearedDuplicateIsNotFlaggedAgain(t *testing.T) {
	p := aProduct(t)
	other, third := catalog.NewProductID(), catalog.NewProductID()
	if err := p.FlagDuplicateOf(other, "same page url", testNow); err != nil {
		t.Fatal(err)
	}
	if err := p.ClearDuplicateFlag("checked: different model", testNow); err != nil {
		t.Fatal(err)
	}
	p.PullEvents()

	if err := p.FlagDuplicateOf(other, "same page url", testNow); err != nil {
		t.Fatalf("re-flagging a cleared pair must be a quiet no-op, got %v", err)
	}
	if _, flagged := p.SuspectedDuplicateOf(); flagged {
		t.Fatal("a cleared pair must not be flagged again")
	}
	if evs := p.PullEvents(); len(evs) != 0 {
		t.Fatalf("re-flagging a cleared pair raised %d events", len(evs))
	}

	// A DIFFERENT candidate is a new question.
	if err := p.FlagDuplicateOf(third, "same name", testNow); err != nil {
		t.Fatal(err)
	}
	if of, flagged := p.SuspectedDuplicateOf(); !flagged || of != third {
		t.Fatalf("SuspectedDuplicateOf() = %v, %v", of, flagged)
	}
}

// Shops discontinue things. A retired product stays for the orders that
// reference it, but can never be published (again).
func TestProduct_retire(t *testing.T) {
	p := aProduct(t)
	p.PullEvents()

	if err := p.Retire("   ", testNow); !errors.Is(err, catalog.ErrEmptyReason) {
		t.Errorf("empty reason: got %v", err)
	}
	if err := p.Retire("discontinued by the shop", testNow); err != nil {
		t.Fatalf("Retire: %v", err)
	}
	if p.Status() != catalog.ProductStatusRetired {
		t.Fatal("must be retired")
	}
	if err := p.Retire("again", testNow); err != nil {
		t.Fatalf("second Retire must be a no-op, got %v", err)
	}
	if evs := p.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.product_retired" {
		t.Fatalf("got %v, want one product_retired", evs)
	}
	if err := p.Publish(testNow); !errors.Is(err, catalog.ErrNotDraft) {
		t.Errorf("publish after retire: got %v", err)
	}
}

func TestParseProductID_roundTrips(t *testing.T) {
	p := aProduct(t)
	back, err := catalog.ParseProductID(p.ID().String())
	if err != nil || back != p.ID() {
		t.Fatalf("ParseProductID = %v, %v", back, err)
	}
	if _, err := catalog.ParseProductID("nope"); !errors.Is(err, shared.ErrInvalidID) {
		t.Errorf("garbage: got %v", err)
	}
}
