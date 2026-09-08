package catalog_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The domain never reads the clock (convention 7), so tests pick the time.
var testNow = time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)

// exampleDetails is a valid registration; tests copy it and break one field.
func exampleDetails() catalog.MerchantDetails {
	return catalog.MerchantDetails{
		Name:         "Example Sports",
		Site:         catalog.MustParseHostname("example.com"),
		Currency:     shared.USD,
		FreeShipping: catalog.MustFreeShippingOver(shared.MustParseMoney("50.00", shared.USD)),
		Sourcing:     []catalog.SourcingMode{catalog.SourcedByOperator},
	}
}

func aMerchant(t *testing.T) *catalog.Merchant {
	t.Helper()
	m, err := catalog.RegisterMerchant(exampleDetails(), testNow)
	if err != nil {
		t.Fatalf("RegisterMerchant: %v", err)
	}
	return m
}

func TestRegisterMerchant_hasIdentityImmediately(t *testing.T) {
	m := aMerchant(t)

	// No round trip to a database: the aggregate has an id from its first
	// line, unlike a Doctrine entity waiting for flush() (convention 5).
	if m.ID().IsZero() {
		t.Fatal("merchant must have an id before it is ever saved")
	}
	if got := m.Name(); got != "Example Sports" {
		t.Errorf("Name() = %q", got)
	}
	if got := m.Site().String(); got != "example.com" {
		t.Errorf("Site() = %q", got)
	}
}

func TestRegisterMerchant_recordsEventButDoesNotPublish(t *testing.T) {
	m := aMerchant(t)

	// The aggregate only RECORDS; the app layer pulls after the repository
	// has saved (convention 6).
	evs := m.PullEvents()
	if len(evs) != 1 {
		t.Fatalf("got %d events, want 1", len(evs))
	}
	if got, want := evs[0].EventName(), "catalog.merchant_registered"; got != want {
		t.Errorf("EventName() = %q, want %q", got, want)
	}
	if !evs[0].OccurredAt().Equal(testNow) {
		t.Errorf("OccurredAt() = %v, want %v", evs[0].OccurredAt(), testNow)
	}
	if rest := m.PullEvents(); len(rest) != 0 {
		t.Errorf("a second pull must be empty, got %d", len(rest))
	}
}

// Every required field is checked, so a caller who forgets one gets an
// error, not a merchant with a hole in it (convention 10).
func TestRegisterMerchant_rejectsBadInput(t *testing.T) {
	cases := map[string]struct {
		mutate func(*catalog.MerchantDetails)
		want   error
	}{
		"empty name":    {func(d *catalog.MerchantDetails) { d.Name = "   " }, catalog.ErrEmptyName},
		"zero hostname": {func(d *catalog.MerchantDetails) { d.Site = catalog.Hostname{} }, catalog.ErrInvalidHostname},
		"zero currency": {func(d *catalog.MerchantDetails) {
			d.Currency = shared.Currency{}
			d.FreeShipping = catalog.NoFreeShipping()
		}, catalog.ErrCurrencyRequired},
		"unknown sourcing": {func(d *catalog.MerchantDetails) {
			d.Sourcing = []catalog.SourcingMode{"telepathy"}
		}, catalog.ErrUnknownSourcing},
		// A USD threshold on a merchant selling in VND is a data-entry
		// mistake that would otherwise surface deep inside a quote.
		"threshold in wrong currency": {func(d *catalog.MerchantDetails) { d.Currency = shared.VND }, catalog.ErrFreeShipCurrency},
	}
	for name, c := range cases {
		d := exampleDetails()
		c.mutate(&d)
		_, err := catalog.RegisterMerchant(d, testNow)
		if !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
}

// The price of a details struct (convention 10): a forgotten field compiles.
// So every field is zeroed ONE AT A TIME on otherwise-valid details, and
// RegisterMerchant must refuse each. A new field either gets validated or is
// listed here with the reason its zero value is a legitimate answer — a
// conscious choice either way.
//
// Why not one all-zero struct: it dies on the first check (empty Name) and
// never reaches the others, so an unvalidated new field slips through green.
// Verified by adding an unchecked `Country string` — every test stayed green.
//
// [PHP] reflect duyệt field của struct lúc chạy — như ReflectionClass::
// [PHP] getProperties(). Chỉ dùng trong test; domain không import reflect.
func TestRegisterMerchant_everyFieldIsValidated(t *testing.T) {
	// Fields whose zero value IS a business answer, not missing data.
	zeroIsMeaningful := map[string]string{
		"FreeShipping": "FreeShipping{} means the shop always charges — the safe default",
		"Sourcing":     "nil means no way to get data yet",
	}

	typ := reflect.TypeOf(exampleDetails())
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		if why, ok := zeroIsMeaningful[name]; ok {
			t.Logf("skip %s: %s", name, why)
			continue
		}
		d := exampleDetails()
		reflect.ValueOf(&d).Elem().Field(i).SetZero()
		if _, err := catalog.RegisterMerchant(d, testNow); err == nil {
			t.Errorf("RegisterMerchant accepted MerchantDetails with zero %s", name)
		}
	}
}

// Two fields have a zero value that IS a legal business answer, on purpose:
// no free-shipping rule means "we pay", no sourcing means "cannot buy yet".
func TestRegisterMerchant_safeZeroFieldsAreLegal(t *testing.T) {
	d := exampleDetails()
	d.FreeShipping = catalog.FreeShipping{}
	d.Sourcing = nil

	m, err := catalog.RegisterMerchant(d, testNow)
	if err != nil {
		t.Fatalf("RegisterMerchant: %v", err)
	}
	if free, _ := m.FreeShipping().AppliesTo(shared.MustParseMoney("999.00", shared.USD)); free {
		t.Error("an unset rule must mean never free")
	}
	if len(m.Sourcing()) != 0 {
		t.Error("nil sourcing must mean no modes")
	}
}

func TestParseHostname(t *testing.T) {
	ok := map[string]string{
		"  Example.COM  ":  "example.com",
		"shop.example.com": "shop.example.com",
	}
	for in, want := range ok {
		h, err := catalog.ParseHostname(in)
		if err != nil || h.String() != want {
			t.Errorf("ParseHostname(%q) = %q, %v; want %q", in, h, err, want)
		}
	}

	// Refused for the reason ParseMoney refuses "150,50": the domain does not
	// guess. Stripping a scheme or a path is the adapter's job.
	for _, in := range []string{
		"", "   ", "localhost", "https://example.com", "example.com/shoes",
		"example.com:443", ".example.com", "example.com.", "exam ple.com",
	} {
		if _, err := catalog.ParseHostname(in); !errors.Is(err, catalog.ErrInvalidHostname) {
			t.Errorf("ParseHostname(%q): got %v, want ErrInvalidHostname", in, err)
		}
	}
}

func TestMerchant_sourcingIsACopy(t *testing.T) {
	m := aMerchant(t)

	// A slice is a window onto an array, not a copy. Handing the caller the
	// real one would let outside code rewrite the aggregate's insides.
	got := m.Sourcing()
	got[0] = "telepathy"

	if !m.Supports(catalog.SourcedByOperator) {
		t.Fatal("mutating the returned slice changed the merchant")
	}
}

func TestMerchant_enableSourcingIsIdempotent(t *testing.T) {
	m := aMerchant(t)
	m.PullEvents() // drop the registration event

	if err := m.EnableSourcing(catalog.SourcedByCustomer, testNow); err != nil {
		t.Fatalf("EnableSourcing: %v", err)
	}
	if err := m.EnableSourcing(catalog.SourcedByCustomer, testNow); err != nil {
		t.Fatalf("EnableSourcing again: %v", err)
	}

	if !m.Supports(catalog.SourcedByCustomer) {
		t.Fatal("customer sourcing should be enabled")
	}
	if n := len(m.Sourcing()); n != 2 {
		t.Errorf("got %d modes, want 2 (no duplicate)", n)
	}
	// Nothing happened the second time, so nothing is announced.
	if evs := m.PullEvents(); len(evs) != 1 {
		t.Errorf("got %d events, want 1", len(evs))
	}

	if err := m.EnableSourcing("telepathy", testNow); !errors.Is(err, catalog.ErrUnknownSourcing) {
		t.Errorf("unknown mode: got %v", err)
	}
}

func TestMerchant_renameRaisesEventOnlyWhenItChanges(t *testing.T) {
	m := aMerchant(t)
	m.PullEvents()

	if err := m.Rename("Example Sports", testNow); err != nil {
		t.Fatalf("Rename to the same name: %v", err)
	}
	if evs := m.PullEvents(); len(evs) != 0 {
		t.Fatalf("renaming to the same name raised %d events", len(evs))
	}

	if err := m.Rename("  Example Athletics  ", testNow); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if got := m.Name(); got != "Example Athletics" {
		t.Errorf("Name() = %q, want trimmed", got)
	}
	evs := m.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "catalog.merchant_renamed" {
		t.Fatalf("got %v", evs)
	}

	if err := m.Rename("  ", testNow); !errors.Is(err, catalog.ErrEmptyName) {
		t.Errorf("empty rename: got %v", err)
	}
}

func TestParseMerchantID_roundTrips(t *testing.T) {
	m := aMerchant(t)

	back, err := catalog.ParseMerchantID(m.ID().String())
	if err != nil || back != m.ID() {
		t.Fatalf("ParseMerchantID = %v, %v", back, err)
	}
	if _, err := catalog.ParseMerchantID("not-an-id"); !errors.Is(err, shared.ErrInvalidID) {
		t.Errorf("garbage id: got %v", err)
	}
}

// A shop that banned our account, or closed, must stop being bought from.
// Suspension is the cross-context fact: procurement listens for it.
func TestMerchant_suspendAndReinstate(t *testing.T) {
	m := aMerchant(t)
	m.PullEvents()

	if !m.IsActive() || m.Status() != catalog.StatusActive {
		t.Fatal("a new merchant must be active")
	}
	if err := m.Suspend("account banned after 3 cancelled orders", testNow); err != nil {
		t.Fatalf("Suspend: %v", err)
	}
	if m.IsActive() || m.Status() != catalog.StatusSuspended {
		t.Fatal("suspended merchant must not be active")
	}
	if err := m.Suspend("again", testNow); err != nil {
		t.Fatalf("second Suspend must be a no-op, got %v", err)
	}
	evs := m.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "catalog.merchant_suspended" {
		t.Fatalf("got %v, want one merchant_suspended", evs)
	}

	if err := m.Reinstate(testNow); err != nil {
		t.Fatalf("Reinstate: %v", err)
	}
	if !m.IsActive() {
		t.Fatal("reinstated merchant must be active")
	}
	if err := m.Reinstate(testNow); err != nil {
		t.Fatalf("second Reinstate must be a no-op, got %v", err)
	}
	evs = m.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "catalog.merchant_reinstated" {
		t.Fatalf("got %v, want one merchant_reinstated", evs)
	}
}

// The reason is what an operator reads in procurement six weeks later.
func TestMerchant_suspendNeedsAReason(t *testing.T) {
	m := aMerchant(t)
	if err := m.Suspend("   ", testNow); !errors.Is(err, catalog.ErrEmptyReason) {
		t.Fatalf("got %v, want ErrEmptyReason", err)
	}
	if !m.IsActive() {
		t.Fatal("a refused Suspend must change nothing")
	}
}

func TestMerchant_disableSourcing(t *testing.T) {
	m := aMerchant(t)
	m.PullEvents()

	if err := m.DisableSourcing(catalog.SourcedByOperator, testNow); err != nil {
		t.Fatalf("DisableSourcing: %v", err)
	}
	if m.Supports(catalog.SourcedByOperator) || len(m.Sourcing()) != 0 {
		t.Fatal("operator sourcing should be gone")
	}
	if err := m.DisableSourcing(catalog.SourcedByOperator, testNow); err != nil {
		t.Fatalf("disabling an absent mode must be a no-op, got %v", err)
	}
	if evs := m.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.merchant_sourcing_disabled" {
		t.Fatalf("got %v, want one merchant_sourcing_disabled", evs)
	}
	if err := m.DisableSourcing("telepathy", testNow); !errors.Is(err, catalog.ErrUnknownSourcing) {
		t.Errorf("unknown mode: got %v", err)
	}
}

// Shops change their delivery rules all the time; the merchant must follow
// without being re-registered — and never into a rule in the wrong currency.
func TestMerchant_changeFreeShipping(t *testing.T) {
	m := aMerchant(t)
	m.PullEvents()

	same := catalog.MustFreeShippingOver(shared.MustParseMoney("50.00", shared.USD))
	if err := m.ChangeFreeShipping(same, testNow); err != nil {
		t.Fatalf("ChangeFreeShipping(same): %v", err)
	}
	if evs := m.PullEvents(); len(evs) != 0 {
		t.Fatalf("an unchanged rule raised %d events", len(evs))
	}

	if err := m.ChangeFreeShipping(catalog.AlwaysFreeShipping(), testNow); err != nil {
		t.Fatalf("ChangeFreeShipping: %v", err)
	}
	if got := m.FreeShipping().String(); got != "always free" {
		t.Errorf("FreeShipping() = %q", got)
	}
	if evs := m.PullEvents(); len(evs) != 1 || evs[0].EventName() != "catalog.merchant_free_shipping_changed" {
		t.Fatalf("got %v, want one merchant_free_shipping_changed", evs)
	}

	vnd := catalog.MustFreeShippingOver(shared.MustParseMoney("1000000", shared.VND))
	if err := m.ChangeFreeShipping(vnd, testNow); !errors.Is(err, catalog.ErrFreeShipCurrency) {
		t.Fatalf("wrong currency: got %v", err)
	}
	if got := m.FreeShipping().String(); got != "always free" {
		t.Error("a refused change must leave the rule alone")
	}
}
