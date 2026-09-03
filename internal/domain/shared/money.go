// Package shared holds the Shared Kernel: value objects every bounded
// context depends on.
//
// RULES FOR THIS PACKAGE (and every package under internal/domain):
//   - It may NOT import a database driver, an HTTP framework, or any SDK.
//   - It may NOT import another bounded context.
//   - Everything here is immutable: methods return new values, never mutate.
package shared

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrMalformedAmount  = errors.New("malformed amount")
)

// Currency is a value object. Exponent is how many decimal digits the
// currency uses: USD has 2 (cents), VND has 0 (no subunit at all).
//
// A Money type that hardcodes 2 decimals silently breaks on VND, which is
// why the exponent travels with the currency instead.
type Currency struct {
	code     string
	exponent int32
}

func (c Currency) Code() string    { return c.code }
func (c Currency) Exponent() int32 { return c.exponent }
func (c Currency) String() string  { return c.code }

var (
	USD = Currency{code: "USD", exponent: 2}
	VND = Currency{code: "VND", exponent: 0}
)

// Money is an immutable amount in a single currency.
//
// The amount is held in MINOR UNITS as an integer (USD cents, VND đồng).
// Floating point is never used: 0.1 + 0.2 != 0.3 in binary floating point,
// and money that is off by a cent per row is money lost at scale.
type Money struct {
	minor    int64
	currency Currency
}

// NewMoney builds an amount directly from minor units.
func NewMoney(minor int64, c Currency) Money { return Money{minor: minor, currency: c} }

func Zero(c Currency) Money { return Money{currency: c} }

// ParseMoney reads a human-written decimal amount ("150.00", "3900000")
// exactly, without ever going through a float.
func ParseMoney(s string, c Currency) (Money, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	if s == "" {
		return Money{}, fmt.Errorf("parse %q: %w", s, ErrMalformedAmount)
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	whole, frac, _ := strings.Cut(s, ".")
	exp := int(c.exponent)
	if len(frac) > exp {
		return Money{}, fmt.Errorf("parse %q: %s allows %d decimals: %w",
			s, c.code, exp, ErrMalformedAmount)
	}
	frac += strings.Repeat("0", exp-len(frac)) // pad "5" -> "50" for USD

	minor, err := strconv.ParseInt(whole+frac, 10, 64)
	if err != nil {
		return Money{}, fmt.Errorf("parse %q: %w", s, ErrMalformedAmount)
	}
	if neg {
		minor = -minor
	}
	return Money{minor: minor, currency: c}, nil
}

// MustParseMoney is for tests and constants only. It panics on bad input.
func MustParseMoney(s string, c Currency) Money {
	m, err := ParseMoney(s, c)
	if err != nil {
		panic(err)
	}
	return m
}

func (m Money) Minor() int64       { return m.minor }
func (m Money) Currency() Currency { return m.currency }
func (m Money) IsZero() bool       { return m.minor == 0 }
func (m Money) IsNegative() bool   { return m.minor < 0 }

// Add refuses to add different currencies. USD and VND are not
// interchangeable, so the type system and the domain both say no.
func (m Money) Add(o Money) (Money, error) {
	if m.currency != o.currency {
		return Money{}, fmt.Errorf("add %s to %s: %w",
			o.currency, m.currency, ErrCurrencyMismatch)
	}
	return Money{minor: m.minor + o.minor, currency: m.currency}, nil
}

func (m Money) Sub(o Money) (Money, error) {
	if m.currency != o.currency {
		return Money{}, fmt.Errorf("subtract %s from %s: %w",
			o.currency, m.currency, ErrCurrencyMismatch)
	}
	return Money{minor: m.minor - o.minor, currency: m.currency}, nil
}

// Sum adds many amounts, failing on the first currency mismatch.
func Sum(first Money, rest ...Money) (Money, error) {
	total := first
	for _, m := range rest {
		var err error
		if total, err = total.Add(m); err != nil {
			return Money{}, err
		}
	}
	return total, nil
}

// Mul applies a rate (a tax rate, a service fee, a margin) and rounds
// half-up to the currency's smallest unit.
func (m Money) Mul(r Rate) Money {
	return Money{minor: mulDivRoundHalfUp(m.minor, r.ppm, ppmScale), currency: m.currency}
}

// Convert changes currency. The result is a DIFFERENT value object; the
// exchange rate used must be recorded by the caller, because a quote has
// to remember the rate it was priced at.
func (m Money) Convert(rate ExchangeRate) (Money, error) {
	if m.currency != rate.from {
		return Money{}, fmt.Errorf("convert %s with %s rate: %w",
			m.currency, rate.from, ErrCurrencyMismatch)
	}
	// big.Int keeps the intermediate product exact for any real-world amount.
	n := new(big.Int).Mul(big.NewInt(m.minor), big.NewInt(rate.minorPerMinorPPM))
	q, rem := new(big.Int).QuoRem(n, big.NewInt(ppmScale), new(big.Int))
	if new(big.Int).Abs(new(big.Int).Mul(rem, big.NewInt(2))).Cmp(big.NewInt(ppmScale)) >= 0 {
		if m.minor < 0 {
			q.Sub(q, big.NewInt(1))
		} else {
			q.Add(q, big.NewInt(1))
		}
	}
	return Money{minor: q.Int64(), currency: rate.to}, nil
}

func (m Money) String() string {
	exp := int(m.currency.exponent)
	if exp == 0 {
		return fmt.Sprintf("%d %s", m.minor, m.currency.code)
	}
	sign, v := "", m.minor
	if v < 0 {
		sign, v = "-", -v
	}
	unit := pow10(exp)
	return fmt.Sprintf("%s%d.%0*d %s", sign, v/unit, exp, v%unit, m.currency.code)
}

func pow10(n int) int64 {
	p := int64(1)
	for range n {
		p *= 10
	}
	return p
}

func mulDivRoundHalfUp(a, b, div int64) int64 {
	n := a * b
	q, rem := n/div, n%div
	if rem < 0 {
		rem = -rem
	}
	if rem*2 >= div {
		if n < 0 {
			q--
		} else {
			q++
		}
	}
	return q
}
