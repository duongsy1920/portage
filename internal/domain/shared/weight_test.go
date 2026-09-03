package shared_test

import (
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
)

const (
	airExpressDivisor int64 = 5000 // cm³ per kg
)

var billingStep = shared.Grams(100)

// A Nike shoe box: heavier than air, but still bills above its mass.
func TestChargeableWeight_shoeBox(t *testing.T) {
	actual := shared.Grams(1200)
	box := shared.NewDimensionsCM(33, 22, 13)

	if got, want := box.VolumeCM3(), int64(9438); got != want {
		t.Fatalf("VolumeCM3 = %d, want %d", got, want)
	}
	if got, want := box.VolumetricWeight(airExpressDivisor).String(), "1.887 kg"; got != want {
		t.Fatalf("volumetric = %s, want %s", got, want)
	}

	chargeable := shared.ChargeableWeight(actual, box, airExpressDivisor, billingStep)
	if got, want := chargeable.String(), "1.900 kg"; got != want {
		t.Fatalf("chargeable = %s, want %s", got, want)
	}
}

// A down jacket weighs almost nothing and bills at five times its mass.
// Quoting this on actual weight loses more than the margin on the order.
func TestChargeableWeight_downJacketBillsOnVolume(t *testing.T) {
	actual := shared.Grams(900)
	carton := shared.NewDimensionsCM(40, 32, 18)

	chargeable := shared.ChargeableWeight(actual, carton, airExpressDivisor, billingStep)

	if got, want := chargeable.String(), "4.700 kg"; got != want {
		t.Fatalf("chargeable = %s, want %s", got, want)
	}
	if chargeable.Grams() <= actual.Grams()*5 {
		t.Errorf("expected volumetric to dominate by >5x, got %s vs actual %s",
			chargeable, actual)
	}
}

// A dense, small parcel bills on its real mass instead.
func TestChargeableWeight_denseParcelBillsOnMass(t *testing.T) {
	actual := shared.Kilos(3)
	small := shared.NewDimensionsCM(20, 15, 10)

	chargeable := shared.ChargeableWeight(actual, small, airExpressDivisor, billingStep)
	if got, want := chargeable.String(), "3.000 kg"; got != want {
		t.Fatalf("chargeable = %s, want %s", got, want)
	}
}

// Carriers round the billing step UP, never to nearest.
func TestWeight_roundUpTo(t *testing.T) {
	cases := []struct{ in, step, want int64 }{
		{1887, 100, 1900},
		{1900, 100, 1900},
		{1901, 100, 2000},
		{4608, 500, 5000},
	}
	for _, c := range cases {
		got := shared.Grams(c.in).RoundUpTo(shared.Grams(c.step)).Grams()
		if got != c.want {
			t.Errorf("Grams(%d).RoundUpTo(%d) = %d, want %d", c.in, c.step, got, c.want)
		}
	}
}

// The divisor is carrier data. A 6000 lane bills the same box less.
func TestVolumetricWeight_divisorIsCarrierData(t *testing.T) {
	box := shared.NewDimensionsCM(33, 22, 13)

	at5000 := box.VolumetricWeight(5000)
	at6000 := box.VolumetricWeight(6000)

	if !(at6000.Grams() < at5000.Grams()) {
		t.Fatalf("divisor 6000 (%s) should bill less than 5000 (%s)", at6000, at5000)
	}
}
