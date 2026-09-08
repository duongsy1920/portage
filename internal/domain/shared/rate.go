package shared

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRate = errors.New("invalid exchange rate")

// ppmScale is the denominator for Rate: 1_000_000 ppm == 100%.
// Basis points (1/10_000) cannot express a rate like 8.815%, so the
// domain works in parts-per-million instead.
const ppmScale int64 = 1_000_000

// Rate is a proportion — a tax rate, a service fee, a margin.
// Like Money it avoids floating point, because a rate multiplied by an
// amount becomes money, and the error would ride along.
//
// A Rate MAY be negative (a discount). An ExchangeRate may not (see below).
//
// Symfony: there is no standard VO for this; most PHP code passes a float
// and hopes. Brick\Math\BigDecimal with scale 6 is the honest equivalent.
type Rate struct {
	ppm int64
}

// RatePPM: 300_000 ppm == 30%.
func RatePPM(ppm int64) Rate {
	return Rate{ppm: ppm}
}

// ParsePercent reads a written percentage exactly: "8.81" -> 8.81%.
// An optional trailing "%" is allowed; everything else follows ParseMoney's
// strict grammar, so "" is an error rather than a silent 0% — a missing
// tax rate becoming zero is exactly how an order ships untaxed.
func ParsePercent(s string) (Rate, error) {
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	const fracDigits = 4 // percent * 10^4 == ppm
	ppm, err := parseDecimal(s, fracDigits)
	if err != nil {
		return Rate{}, fmt.Errorf("percent %q: %w", s, err)
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

func (r Rate) PPM() int64 {
	return r.ppm
}

// Plus1 returns 1 + r, so an 8.81% tax becomes a 108.81% multiplier.
func (r Rate) Plus1() Rate {
	return Rate{ppm: ppmScale + r.ppm}
}

func (r Rate) String() string {
	sign, v := "", r.ppm
	if v < 0 {
		// Format the magnitude, then restore the sign: -5000/10_000 is 0 in
		// integer math and would print -0.5% as "0.5000%".
		sign, v = "-", -v
	}
	return fmt.Sprintf("%s%d.%04d%%", sign, v/10_000, v%10_000)
}

// ExchangeRate converts between two currencies.
//
// It is a value object on purpose: a Quote must SNAPSHOT the rate it was
// priced with. If a quote merely pointed at "the current rate", yesterday's
// quote would silently change price today.
//
// Provenance — which bank, fetched when — is NOT here. That belongs to the
// pricing context (a RateSnapshot wrapping this value with source and
// time); the shared kernel only knows how to convert.
//
// The rate is stored as exact decimal digits over 10^scale: "26000.5" is
// digits=260005, scale=1. It is normalised (no trailing zeros) so two rates
// written differently but equal in value are == equal. Up to 12 decimals —
// enough for the reverse direction, 1 VND = 0.0000384615 USD, with room.
//
// Symfony: moneyphp's CurrencyPair + FixedExchange. Doctrine would map it
// as an #[ORM\Embeddable] with three columns: from, to, and Rate() as text.
type ExchangeRate struct {
	from, to Currency
	digits   int64 // the quoted rate with the decimal point removed
	scale    int32 // how many of those digits were fractional
}

const maxRateScale = 12

// NewExchangeRate takes the human-facing rate: how many MAJOR units of `to`
// one MAJOR unit of `from` buys — "26000" for USD→VND, "0.0000384615" back.
// The rate must be positive, and the currencies must differ.
func NewExchangeRate(from, to Currency, rate string) (ExchangeRate, error) {
	if from.IsZero() || to.IsZero() || from == to {
		return ExchangeRate{}, fmt.Errorf("exchange rate %s->%s: %w", from, to, ErrInvalidRate)
	}
	digits, err := parseDecimal(rate, maxRateScale)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("exchange rate %q: %w", rate, err)
	}
	if digits <= 0 {
		return ExchangeRate{}, fmt.Errorf("exchange rate %q: must be positive: %w", rate, ErrInvalidRate)
	}
	scale := int32(maxRateScale)
	for scale > 0 && digits%10 == 0 { // canonical form: drop trailing zeros
		digits /= 10
		scale--
	}
	return ExchangeRate{from: from, to: to, digits: digits, scale: scale}, nil
}

func MustExchangeRate(from, to Currency, rate string) ExchangeRate {
	r, err := NewExchangeRate(from, to, rate)
	if err != nil {
		panic(err)
	}
	return r
}

func (e ExchangeRate) From() Currency {
	return e.from
}
func (e ExchangeRate) To() Currency {
	return e.to
}

// Rate returns the canonical decimal text ("26000.5") — what a repository
// stores and what NewExchangeRate reads back.
func (e ExchangeRate) Rate() string {
	s := strconv.FormatInt(e.digits, 10)
	if e.scale == 0 {
		return s
	}
	if pad := int(e.scale) + 1 - len(s); pad > 0 {
		s = strings.Repeat("0", pad) + s // "384615" -> "00000384615"
	}
	cut := len(s) - int(e.scale)
	return s[:cut] + "." + s[cut:]
}

// String reads like a quote board: "1 USD = 26000.5 VND".
func (e ExchangeRate) String() string {
	return fmt.Sprintf("1 %s = %s %s", e.from, e.Rate(), e.to)
}
