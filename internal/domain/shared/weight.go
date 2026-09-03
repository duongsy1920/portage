package shared

import "fmt"

// Weight is an immutable value object held in grams.
//
// A bare number is not a weight: "1.2" is meaningless until you know the
// unit. Shipping a value typed as float64 across a system is how a package
// gets billed in pounds and quoted in kilograms.
type Weight struct{ grams int64 }

func Grams(g int64) Weight    { return Weight{grams: g} }
func Kilos(kg int64) Weight   { return Weight{grams: kg * 1000} }
func (w Weight) Grams() int64 { return w.grams }
func (w Weight) IsZero() bool { return w.grams == 0 }
func (w Weight) String() string {
	return fmt.Sprintf("%d.%03d kg", w.grams/1000, abs64(w.grams%1000))
}

func (w Weight) Add(o Weight) Weight { return Weight{grams: w.grams + o.grams} }

func MaxWeight(a, b Weight) Weight {
	if b.grams > a.grams {
		return b
	}
	return a
}

// RoundUpTo rounds up to the carrier's billing step (100 g, 500 g, 1 kg).
// Carriers always round UP, never to nearest — so the domain does too.
func (w Weight) RoundUpTo(step Weight) Weight {
	if step.grams <= 0 {
		return w
	}
	if rem := w.grams % step.grams; rem != 0 {
		return Weight{grams: w.grams + step.grams - rem}
	}
	return w
}

// Dimensions of a package, held in millimetres.
type Dimensions struct{ lengthMM, widthMM, heightMM int64 }

func NewDimensionsMM(l, w, h int64) Dimensions {
	return Dimensions{lengthMM: l, widthMM: w, heightMM: h}
}

// NewDimensionsCM is the unit couriers actually quote in.
func NewDimensionsCM(l, w, h int64) Dimensions {
	return Dimensions{lengthMM: l * 10, widthMM: w * 10, heightMM: h * 10}
}

func (d Dimensions) VolumeCM3() int64 {
	return d.lengthMM * d.widthMM * d.heightMM / 1000
}

// VolumetricWeight is what an airline charges for: space, not mass.
// The divisor is the carrier's cm³-per-kg constant — 5000 for most air
// express, 6000 for some lanes. It is carrier data, never a constant here.
func (d Dimensions) VolumetricWeight(divisor int64) Weight {
	if divisor <= 0 {
		return Weight{}
	}
	return Weight{grams: d.VolumeCM3() * 1000 / divisor}
}

// ChargeableWeight is the number the invoice is built on.
//
//	chargeable = roundUp( max(actual, volumetric) )
//
// Quoting on actual weight alone is the single most expensive mistake in
// parcel forwarding: a down jacket weighs 0.9 kg and bills at 4.7 kg.
func ChargeableWeight(actual Weight, d Dimensions, divisor int64, step Weight) Weight {
	return MaxWeight(actual, d.VolumetricWeight(divisor)).RoundUpTo(step)
}
