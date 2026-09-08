package shared

import (
	"errors"
	"fmt"
)

var ErrNegativeWeight = errors.New("weight cannot be negative")

// Weight is an immutable value object held in grams.
//
// A bare number is not a weight: "1.2" is meaningless until you know the
// unit. Shipping a value typed as float64 across a system is how a package
// gets billed in pounds and quoted in kilograms.
//
// A Weight is never negative: the constructors refuse it, and no operation
// in this package can produce one.
//
// Symfony: an #[ORM\Embeddable] with a single int column — except the
// invariant lives in the constructor, not in a #[Assert\PositiveOrZero]
// that only runs when someone remembers to call the validator.
type Weight struct {
	grams int64
}

// NewWeight validates untrusted input — a form field, a carrier's CSV.
//
// [PHP] So sánh hai constructor ngay dưới đây để thấy quy ước error/panic:
// [PHP]     NewWeight(g) (Weight, error)  → input NGOÀI  → trả lỗi  (~ DomainException)
// [PHP]     Grams(g)      Weight          → literal code → panic    (~ LogicException)
func NewWeight(grams int64) (Weight, error) {
	if grams < 0 {
		// [PHP] `%d` = in số nguyên (PHP: %d trong sprintf, giống hệt).
		return Weight{}, fmt.Errorf("%d g: %w", grams, ErrNegativeWeight)
	}
	return Weight{grams: grams}, nil
}

// Grams and Kilos are literal constructors for code and tests. A negative
// literal is a programming error and panics; see the package comment.
func Grams(g int64) Weight {
	w, err := NewWeight(g)
	if err != nil {
		panic("shared: " + err.Error())
	}
	return w
}

func Kilos(kg int64) Weight {
	return Grams(mulExact(kg, 1000))
}

func (w Weight) Grams() int64 {
	return w.grams
}
func (w Weight) IsZero() bool {
	return w.grams == 0
}
func (w Weight) String() string {
	return fmt.Sprintf("%d.%03d kg", w.grams/1000, w.grams%1000)
}

// Add totals two weights — the parcels in a consolidation batch, say.
//
// It cannot return an error, because two non-negative weights always make a
// third: the only way out of range is int64 overflow, and addExact panics on
// that. Letting it wrap would produce a NEGATIVE weight, the one state
// NewWeight exists to make impossible, and String() would then print
// nonsense like "-9223372036854775.-808 kg".
func (w Weight) Add(o Weight) Weight {
	return Weight{grams: addExact(w.grams, o.grams)}
}

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
		return Weight{grams: addExact(w.grams, step.grams-rem)}
	}
	return w
}

// Dimensions of a package, held in millimetres. Never negative.
//
// The getters exist for one caller: the repository, which has to write the
// three sides back to three columns. Nothing in the domain reads them.
// [PHP] `lengthMM, widthMM, heightMM int64` — ba field CÙNG KIỂU gộp một dòng.
// [PHP]     PHP: private int $lengthMM; private int $widthMM; private int $heightMM;
// [PHP] Cùng cú pháp đó dùng được cho tham số hàm: `func f(a, b int)`.
type Dimensions struct {
	lengthMM, widthMM, heightMM int64
}

func NewDimensionsMM(l, w, h int64) Dimensions {
	if l < 0 || w < 0 || h < 0 {
		panic(fmt.Sprintf("shared: negative dimension %dx%dx%d mm", l, w, h))
	}
	return Dimensions{lengthMM: l, widthMM: w, heightMM: h}
}

// NewDimensionsCM is the unit couriers actually quote in.
func NewDimensionsCM(l, w, h int64) Dimensions {
	return NewDimensionsMM(l*10, w*10, h*10)
}

func (d Dimensions) LengthMM() int64 {
	return d.lengthMM
}
func (d Dimensions) WidthMM() int64 {
	return d.widthMM
}
func (d Dimensions) HeightMM() int64 {
	return d.heightMM
}

// String prints millimetres, never centimetres: a value that reaches a log
// line or an error message must be exactly the value stored.
//
// [PHP] Struct có method `String() string` là tự thoả interface fmt.Stringer:
// [PHP] fmt.Sprintf("%s", d) và "%v" đều gọi nó. Tương đương __toString().
func (d Dimensions) String() string {
	return fmt.Sprintf("%dx%dx%d mm", d.lengthMM, d.widthMM, d.heightMM)
}

func (d Dimensions) VolumeCM3() int64 {
	return d.lengthMM * d.widthMM * d.heightMM / 1000
}

// VolumetricWeight is what an airline charges for: space, not mass.
// The divisor is the carrier's cm³-per-kg constant — 5000 for most air
// express, 6000 for some lanes. It is carrier data, never a constant here.
//
//	grams = cm³ × 1000 ÷ divisor = mm³ ÷ divisor
//
// Working from mm³ avoids a first truncation in VolumeCM3, and the division
// rounds UP: 1900.8 g is 1901 g to a carrier, never 1900. Rounding down
// here once landed exactly on a 100 g billing step and under-billed a
// whole step.
func (d Dimensions) VolumetricWeight(divisor int64) Weight {
	if divisor <= 0 {
		return Weight{}
	}
	mm3 := d.lengthMM * d.widthMM * d.heightMM
	return Weight{grams: (mm3 + divisor - 1) / divisor}
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
