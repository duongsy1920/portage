package shared

import (
	"fmt"
	"strconv"
	"strings"
)

// ppmScale is the denominator for Rate: 1_000_000 ppm == 100%.
// Basis points (1/10_000) cannot express a rate like 8.815%, so the
// domain works in parts-per-million instead.
const ppmScale int64 = 1_000_000

// Rate is a proportion — a tax rate, a service fee, a margin.
// Like Money it avoids floating point, because a rate multiplied by an
// amount becomes money, and the error would ride along.
type Rate struct{ ppm int64 }

// RatePPM: 300_000 ppm == 30%.
func RatePPM(ppm int64) Rate { return Rate{ppm: ppm} }

// ParsePercent reads a written percentage exactly: "8.81" -> 8.81%.
func ParsePercent(s string) (Rate, error) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "%"))
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	whole, frac, _ := strings.Cut(s, ".")
	const fracDigits = 4 // percent * 10^4 == ppm
	if len(frac) > fracDigits {
		return Rate{}, fmt.Errorf("percent %q: at most %d decimals: %w",
			s, fracDigits, ErrMalformedAmount)
	}
	frac += strings.Repeat("0", fracDigits-len(frac))

	ppm, err := strconv.ParseInt(whole+frac, 10, 64)
	if err != nil {
		return Rate{}, fmt.Errorf("percent %q: %w", s, ErrMalformedAmount)
	}
	if neg {
		ppm = -ppm
	}
	return Rate{ppm: ppm}, nil
}

func MustParsePercent(s string) Rate {
	r, err := ParsePercent(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (r Rate) PPM() int64 { return r.ppm }

// Plus returns 1 + r, so a 8.81% tax becomes a 108.81% multiplier.
func (r Rate) Plus1() Rate { return Rate{ppm: ppmScale + r.ppm} }

func (r Rate) String() string {
	return fmt.Sprintf("%d.%04d%%", r.ppm/10_000, abs64(r.ppm%10_000))
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// ExchangeRate converts between two currencies at a point in time.
//
// It is a value object on purpose: a Quote must SNAPSHOT the rate it was
// priced with. If a quote merely pointed at "the current rate", yesterday's
// quote would silently change price today.
type ExchangeRate struct {
	from, to         Currency
	minorPerMinorPPM int64 // minor units of `to` per minor unit of `from`, in ppm
	Source           string
	AsOf             string
}

// NewExchangeRate takes the human-facing rate: how many MAJOR units of
// `to` one MAJOR unit of `from` buys — e.g. 26000 VND per 1 USD.
func NewExchangeRate(from, to Currency, majorPerMajor string, source, asOf string) (ExchangeRate, error) {
	r, err := ParsePercent(majorPerMajor) // reuse: 4-decimal fixed point
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("exchange rate %q: %w", majorPerMajor, err)
	}
	// r.ppm holds majorPerMajor * 10_000.
	// minorPerMinor = majorPerMajor * 10^to.exp / 10^from.exp
	// expressed in ppm: * 1_000_000 / 10_000 == * 100
	v := r.ppm * 100 * pow10(int(to.exponent)) / pow10(int(from.exponent))
	return ExchangeRate{from: from, to: to, minorPerMinorPPM: v, Source: source, AsOf: asOf}, nil
}

func MustExchangeRate(from, to Currency, majorPerMajor, source, asOf string) ExchangeRate {
	r, err := NewExchangeRate(from, to, majorPerMajor, source, asOf)
	if err != nil {
		panic(err)
	}
	return r
}

func (e ExchangeRate) From() Currency { return e.from }
func (e ExchangeRate) To() Currency   { return e.to }
