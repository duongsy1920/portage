package catalog_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// A shoe box as measured on the warehouse scale: the number a quote is built
// on until this exact product has been weighed for real.
func shoeParcel() shared.ParcelSpec {
	return shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13))
}

func footwear(t *testing.T) catalog.CategoryPolicy {
	t.Helper()
	c, err := catalog.NewCategoryPolicy(
		catalog.MustParseCategoryCode("  FOOTWEAR  "),
		shoeParcel(),
		[]catalog.Restriction{catalog.RestrictionMagnet, catalog.RestrictionMagnet},
	)
	if err != nil {
		t.Fatalf("NewCategoryPolicy: %v", err)
	}
	return c
}

// The code IS the identity: it ends up in URLs, config files and a primary
// key column, so it is validated like Hostname, not merely lowercased.
func TestParseCategoryCode(t *testing.T) {
	ok := map[string]string{
		"footwear":        "footwear",
		"  FOOTWEAR  ":    "footwear",
		"home_appliances": "home_appliances",
		"tv4k":            "tv4k",
	}
	for in, want := range ok {
		c, err := catalog.ParseCategoryCode(in)
		if err != nil || c.String() != want {
			t.Errorf("ParseCategoryCode(%q) = %q, %v; want %q", in, c, err, want)
		}
	}
	for _, in := range []string{"", "   ", "foot wear", "giày", "footwear!", "1abc", "home-appliances", "_x"} {
		if _, err := catalog.ParseCategoryCode(in); !errors.Is(err, catalog.ErrInvalidCategoryCode) {
			t.Errorf("ParseCategoryCode(%q): got %v, want ErrInvalidCategoryCode", in, err)
		}
	}
}

// Two spellings of one code must be ONE key — in a map here, in a table later.
func TestCategoryCode_isComparable(t *testing.T) {
	seen := map[catalog.CategoryCode]int{}
	seen[catalog.MustParseCategoryCode("FOOTWEAR")]++
	seen[catalog.MustParseCategoryCode("footwear")]++
	if len(seen) != 1 || seen[catalog.MustParseCategoryCode("footwear")] != 2 {
		t.Fatalf("expected one key counted twice, got %v", seen)
	}
}

// An estimate is only useful whole: a weight without a box, or a box with a
// zero side, cannot produce a chargeable weight.
func TestNewParcelSpec_requiresWeightAndBox(t *testing.T) {
	box := shared.NewDimensionsCM(33, 22, 13)

	est, err := shared.NewParcelSpec(shared.Grams(1200), box)
	if err != nil || est.Weight() != shared.Grams(1200) || est.Dimensions() != box {
		t.Fatalf("NewParcelSpec = %v, %v", est, err)
	}
	if _, err := shared.NewParcelSpec(shared.Grams(0), box); !errors.Is(err, shared.ErrIncompleteParcelSpec) {
		t.Errorf("zero weight: got %v", err)
	}
	if _, err := shared.NewParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 0, 13)); !errors.Is(err, shared.ErrIncompleteParcelSpec) {
		t.Errorf("zero side: got %v", err)
	}
	if !(shared.ParcelSpec{}).IsZero() {
		t.Error("ParcelSpec{} must be zero")
	}
}

func TestNewCategoryPolicy_normalisesRestrictions(t *testing.T) {
	c := footwear(t)

	if got := c.Code().String(); got != "footwear" {
		t.Errorf("Code() = %q", got)
	}
	if n := len(c.Restrictions()); n != 1 {
		t.Errorf("got %d restrictions, want 1 (duplicate dropped)", n)
	}
	if !c.Restricted(catalog.RestrictionMagnet) {
		t.Error("magnet restriction missing")
	}
	if c.Restricted(catalog.RestrictionBattery) {
		t.Error("battery must not be restricted here")
	}
}

func TestNewCategoryPolicy_rejectsBadInput(t *testing.T) {
	code := catalog.MustParseCategoryCode("apparel")
	cases := map[string]struct {
		code  catalog.CategoryCode
		est   shared.ParcelSpec
		restr []catalog.Restriction
		want  error
	}{
		"zero code":           {catalog.CategoryCode{}, shoeParcel(), nil, catalog.ErrInvalidCategoryCode},
		"zero parcel spec":    {code, shared.ParcelSpec{}, nil, shared.ErrIncompleteParcelSpec},
		"unknown restriction": {code, shoeParcel(), []catalog.Restriction{"cursed"}, catalog.ErrUnknownRestriction},
	}
	for name, c := range cases {
		_, err := catalog.NewCategoryPolicy(c.code, c.est, c.restr)
		if !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
}

// The category supplies what the GOODS are like; the lane supplies how the
// carrier bills. Pricing puts the two together — the maths stays in shared.
func TestCategoryPolicy_estimateFeedsAQuote(t *testing.T) {
	est := footwear(t).DefaultParcelSpec()
	if got := est.Weight().String(); got != "1.200 kg" {
		t.Fatalf("estimated weight = %s", got)
	}

	// Lane data, owned by pricing.ShippingLane: divisor 5000, 100 g steps.
	quoted := shared.ChargeableWeight(est.Weight(), est.Dimensions(), 5000, shared.Grams(100))
	if got := quoted.String(); got != "1.900 kg" {
		t.Fatalf("quoted weight = %s, want 1.900 kg", got)
	}
}

// Restrictions travel with the goods, not with the lane: a lithium cell is a
// lithium cell whoever carries it.
func TestCategoryPolicy_restrictionsAreAboutTheGoods(t *testing.T) {
	electronics, err := catalog.NewCategoryPolicy(
		catalog.MustParseCategoryCode("electronics"),
		shared.MustParcelSpec(shared.Grams(400), shared.NewDimensionsCM(20, 18, 8)),
		[]catalog.Restriction{catalog.RestrictionBattery, catalog.RestrictionMagnet},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !electronics.Restricted(catalog.RestrictionBattery) {
		t.Error("headphones carry a lithium cell")
	}
	if electronics.Restricted(catalog.RestrictionLiquid) {
		t.Error("electronics are not a liquid")
	}
}

func TestCategoryPolicy_restrictionsAreACopy(t *testing.T) {
	c := footwear(t)

	got := c.Restrictions()
	got[0] = "cursed"

	if !c.Restricted(catalog.RestrictionMagnet) {
		t.Fatal("mutating the returned slice changed the policy")
	}
}

func TestCategoryPolicy_zeroValueAndString(t *testing.T) {
	var unset catalog.CategoryPolicy
	if !unset.IsZero() {
		t.Fatal("the zero value must report itself as unset")
	}
	c := footwear(t)
	if c.IsZero() {
		t.Fatal("a real policy must not report itself as zero")
	}
	if got, want := c.String(), "footwear (~1.200 kg, 330x220x130 mm)"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
