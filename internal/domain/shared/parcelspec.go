package shared

import (
	"errors"
	"fmt"
)

var ErrIncompleteParcelSpec = errors.New("parcel spec needs a weight and a box with three sides")

// ParcelSpec is what something weighs and how big its box is — the two numbers
// every freight quote is built on, and the data no merchant feed carries
// (docs/CATALOG.md §3).
//
// One value object, two roles, and the difference is not the shape but the
// TRUST — which is why the trust travels separately, as a Provenance:
//
//   - on a CategoryPolicy it is the DEFAULT for a kind of goods — a shoe box is
//     about 1.2 kg and 33×22×13 cm — used to quote a product nobody has put on
//     a scale yet;
//   - on a Product it is what THIS item measured, and Product.ParcelProvenance
//     says who measured it and when.
//
// Quote vs Actual (DDD.md §28) starts here: pricing quotes from the category
// default, logistics records the real one, the gap is the margin.
//
// It lives in the shared kernel because three contexts hold one: catalog (a
// category's default, a product's measurement), pricing (the basis of a
// quote) and logistics (what the scale in Denver said). The carrier's divisor
// and billing step are NOT here — they belong to pricing.ShippingLane, which
// puts the two together:
//
//	lane.ChargeableWeight(spec.Weight(), spec.Dimensions())
//
// [PHP] Value object thuần: field private, một constructor validate, chỉ có
// [PHP] getter. Symfony: một #[ORM\Embeddable] với 4 cột (weight_g, length_mm,
// [PHP] width_mm, height_mm) — nhúng vào bảng category VÀ bảng product.
type ParcelSpec struct {
	weight Weight
	dims   Dimensions
}

// NewParcelSpec validates operator input (convention 1). Both halves are
// required: a weight without a box, or a box with a zero side, cannot produce
// a chargeable weight — and a half-spec that "mostly works" is worse than
// none, because it quotes confidently and wrongly.
func NewParcelSpec(weight Weight, dims Dimensions) (ParcelSpec, error) {
	if weight.IsZero() || dims.LengthMM() == 0 || dims.WidthMM() == 0 || dims.HeightMM() == 0 {
		return ParcelSpec{}, fmt.Errorf("%s in %s: %w", weight, dims, ErrIncompleteParcelSpec)
	}
	return ParcelSpec{weight: weight, dims: dims}, nil
}

// MustParcelSpec is for tests and seed data; it panics on bad input.
func MustParcelSpec(weight Weight, dims Dimensions) ParcelSpec {
	p, err := NewParcelSpec(weight, dims)
	if err != nil {
		panic(err)
	}
	return p
}

func (p ParcelSpec) Weight() Weight {
	return p.weight
}

func (p ParcelSpec) Dimensions() Dimensions {
	return p.dims
}

// IsZero reports the zero value ParcelSpec{} — "nobody has measured or
// estimated this yet" (convention 9).
//
// [PHP] `p == ParcelSpec{}` so sánh struct bằng `==`: Go cho phép khi mọi field
// [PHP] đều so sánh được (int64, struct của int64…). Slice/map thì không.
func (p ParcelSpec) IsZero() bool {
	return p == ParcelSpec{}
}

func (p ParcelSpec) String() string {
	return fmt.Sprintf("~%s, %s", p.weight, p.dims)
}
