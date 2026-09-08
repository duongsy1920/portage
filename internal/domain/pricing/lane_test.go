package pricing_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

func usd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.USD)
}

// The forwarder's price list, as far as we know it today (SETUP.md §7: numbers
// still to be confirmed with the shipper — these are placeholders that make
// the arithmetic checkable).
func rates(t *testing.T) pricing.RateCard {
	t.Helper()
	r, err := pricing.NewRateCard(usd("9.00"), usd("10.00"), usd("12.00"), usd("14.00"))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func laneDetails(t *testing.T) pricing.LaneDetails {
	t.Helper()
	return pricing.LaneDetails{
		Code:             pricing.MustParseLaneCode("us_forwarder"),
		Name:             "US forwarder, door to door",
		Divisor:          5000,
		Step:             shared.Grams(500),
		Rates:            rates(t),
		BatterySurcharge: usd("3.00"),
		Duty:             pricing.DutyBundled(),
	}
}

func aLane(t *testing.T) pricing.ShippingLane {
	t.Helper()
	l, err := pricing.NewShippingLane(laneDetails(t))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestParseLaneCode(t *testing.T) {
	for in, want := range map[string]string{"us_forwarder": "us_forwarder", "  US_Forwarder ": "us_forwarder"} {
		c, err := pricing.ParseLaneCode(in)
		if err != nil || c.String() != want {
			t.Errorf("ParseLaneCode(%q) = %q, %v", in, c, err)
		}
	}
	for _, in := range []string{"", "us forwarder", "1abc", "us-forwarder"} {
		if _, err := pricing.ParseLaneCode(in); !errors.Is(err, pricing.ErrInvalidLaneCode) {
			t.Errorf("ParseLaneCode(%q): got %v", in, err)
		}
	}
}

// The three numbers catalog handed over (CATALOG.md §6): the volumetric
// divisor, the billing step and the duty rule are properties of the LANE.
// A shoe on this lane and the same shoe on a 6000-divisor lane bill differently.
func TestShippingLane_chargeableWeightUsesItsOwnDivisorAndStep(t *testing.T) {
	lane := aLane(t)
	shoeBox := shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)) // 10166 cm3 → 2034 g

	if got := lane.ChargeableWeight(shoeBox).String(); got != "2.500 kg" {
		t.Fatalf("chargeable = %s, want 2.500 kg (2034 g rounded UP to the 500 g step)", got)
	}
	d := laneDetails(t)
	d.Divisor, d.Step = 6000, shared.Grams(100)
	other, _ := pricing.NewShippingLane(d)
	if got := other.ChargeableWeight(shoeBox).String(); got != "1.700 kg" {
		t.Fatalf("on a 6000/100 g lane = %s, want 1.700 kg", got)
	}
}

// Freight = rate per kg × chargeable kg, exact (no float): 2.5 kg × $9.00 = $22.50.
func TestShippingLane_freightAndSurcharge(t *testing.T) {
	lane := aLane(t)
	if got := lane.Freight(pricing.ClassStandard, shared.Grams(2500)).String(); got != "22.50 USD" {
		t.Errorf("standard 2.5 kg = %s", got)
	}
	if got := lane.Freight(pricing.ClassElectronics, shared.Grams(1000)).String(); got != "12.00 USD" {
		t.Errorf("electronics 1 kg = %s", got)
	}
	if got := lane.Surcharge([]string{"battery", "magnet"}).String(); got != "3.00 USD" {
		t.Errorf("battery surcharge = %s", got)
	}
	if got := lane.Surcharge([]string{"magnet"}); !got.IsZero() {
		t.Errorf("no battery → %s, want zero", got)
	}
	if got := lane.Duty(usd("150.00")); !got.IsZero() {
		t.Errorf("bundled lane charges duty %s, want zero", got)
	}
	itemised, _ := pricing.DutyItemised(shared.MustParsePercent("10"))
	d := laneDetails(t)
	d.Duty = itemised
	commercial, _ := pricing.NewShippingLane(d)
	if got := commercial.Duty(usd("150.00")).String(); got != "15.00 USD" {
		t.Errorf("itemised 10%% on 150 = %s", got)
	}
}

func TestNewRateCard_rejectsMixedCurrencyAndNonPositive(t *testing.T) {
	if _, err := pricing.NewRateCard(usd("9.00"), shared.MustParseMoney("200000", shared.VND), usd("12.00"), usd("14.00")); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("mixed currency: got %v", err)
	}
	if _, err := pricing.NewRateCard(usd("0.00"), usd("10.00"), usd("12.00"), usd("14.00")); !errors.Is(err, pricing.ErrInvalidRate) {
		t.Errorf("zero rate: got %v", err)
	}
	r := rates(t)
	if r.PerKg(pricing.ClassSensitive).String() != "14.00 USD" || r.Currency() != shared.USD {
		t.Errorf("PerKg/Currency = %s / %s", r.PerKg(pricing.ClassSensitive), r.Currency())
	}
}

// Convention 10 applied to the lane: every required field refused when zero,
// one at a time.
func TestNewShippingLane_everyFieldIsValidated(t *testing.T) {
	zeroIsMeaningful := map[string]string{
		"BatterySurcharge": "Money{} means the lane does not surcharge batteries",
		"Duty":             "DutyPolicy{} is DutyBundled(), the lane in use",
	}
	typ := reflect.TypeOf(laneDetails(t))
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		if why, ok := zeroIsMeaningful[name]; ok {
			t.Logf("skip %s: %s", name, why)
			continue
		}
		d := laneDetails(t)
		reflect.ValueOf(&d).Elem().Field(i).SetZero()
		if _, err := pricing.NewShippingLane(d); err == nil {
			t.Errorf("NewShippingLane accepted LaneDetails with zero %s", name)
		}
	}
	d := laneDetails(t)
	d.BatterySurcharge = shared.MustParseMoney("50000", shared.VND)
	if _, err := pricing.NewShippingLane(d); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("surcharge in another currency: got %v", err)
	}
}

func TestClassification_defaultsToStandard(t *testing.T) {
	c, err := pricing.NewClassification(map[string]pricing.GoodsClass{
		"electronics": pricing.ClassElectronics, "watches": pricing.ClassBranded,
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.ClassOf("electronics") != pricing.ClassElectronics || c.ClassOf("footwear") != pricing.ClassStandard {
		t.Errorf("ClassOf: %s / %s", c.ClassOf("electronics"), c.ClassOf("footwear"))
	}
	if _, err := pricing.NewClassification(map[string]pricing.GoodsClass{"x": "luxury"}); !errors.Is(err, pricing.ErrUnknownGoodsClass) {
		t.Errorf("unknown class: got %v", err)
	}
}
