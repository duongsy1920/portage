package shared_test

import (
	"errors"
	"math"
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
	if got, want := box.VolumetricWeight(airExpressDivisor).String(), "1.888 kg"; got != want {
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

// 32x27x11 cm = 9504 cm3 -> 1900.8 g volumetric. Truncating to 1900 g lands
// exactly on a 100 g step and bills 1.9 kg; the carrier bills 2.0 kg. Every
// rounding step in a weight calculation goes UP, never down.
func TestVolumetricWeight_roundsUpNotDown(t *testing.T) {
	box := shared.NewDimensionsCM(32, 27, 11)

	if got, want := box.VolumetricWeight(airExpressDivisor).String(), "1.901 kg"; got != want {
		t.Fatalf("volumetric = %s, want %s", got, want)
	}
	chargeable := shared.ChargeableWeight(shared.Grams(500), box, airExpressDivisor, billingStep)
	if got, want := chargeable.String(), "2.000 kg"; got != want {
		t.Fatalf("chargeable = %s, want %s", got, want)
	}
}

// A parcel cannot weigh -3 kg. Literal constructors trust the programmer and
// panic; NewWeight validates untrusted input and returns an error.
func TestWeight_refusesNegative(t *testing.T) {
	mustPanic(t, func() { shared.Grams(-1) })
	mustPanic(t, func() { shared.Kilos(-1) })
	mustPanic(t, func() { shared.NewDimensionsCM(-1, 1, 1) })

	if _, err := shared.NewWeight(-1); !errors.Is(err, shared.ErrNegativeWeight) {
		t.Errorf("NewWeight(-1): got %v, want ErrNegativeWeight", err)
	}
	if w, err := shared.NewWeight(0); err != nil || !w.IsZero() {
		t.Errorf("NewWeight(0) = %s, %v", w, err)
	}
}

// The repository has to write a Dimensions row back to three columns.
func TestDimensions_exposeSidesForPersistence(t *testing.T) {
	d := shared.NewDimensionsCM(33, 22, 13)
	if d.LengthMM() != 330 || d.WidthMM() != 220 || d.HeightMM() != 130 {
		t.Fatalf("getters = %d %d %d", d.LengthMM(), d.WidthMM(), d.HeightMM())
	}
}

// Overflow in Weight is worse than in Money: a wrapped sum is NEGATIVE, the
// one state NewWeight refuses to build, and String() then prints garbage
// like "-9223372036854775.-808 kg". Every path that can add panics instead.
func TestWeight_panicsOnOverflow(t *testing.T) {
	huge := shared.Grams(math.MaxInt64)

	mustPanic(t, func() { huge.Add(shared.Grams(1)) })
	mustPanic(t, func() { shared.Grams(math.MaxInt64 - 5).RoundUpTo(shared.Grams(100)) })
	mustPanic(t, func() { shared.Kilos(math.MaxInt64/1000 + 1) })

	// The ordinary cases still work: a batch of parcels, a billing step.
	if got := shared.Kilos(3).Add(shared.Grams(500)); got.String() != "3.500 kg" {
		t.Fatalf("3kg + 500g = %s", got)
	}
}

// Millimetres in, millimetres out: no rounding to centimetres in a String()
// that ends up in logs and error messages.
func TestDimensions_string(t *testing.T) {
	if got, want := shared.NewDimensionsCM(33, 22, 13).String(), "330x220x130 mm"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
