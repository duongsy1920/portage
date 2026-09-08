package shared_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
)

func TestParsePercent_keepsExactValue(t *testing.T) {
	cases := []struct {
		in   string
		ppm  int64
		text string
	}{
		{"8.81", 88_100, "8.8100%"},
		{"30", 300_000, "30.0000%"},
		{"0.0001", 1, "0.0001%"},
		{"-2.5", -25_000, "-2.5000%"},
		{"8.81%", 88_100, "8.8100%"},
	}
	for _, c := range cases {
		got, err := shared.ParsePercent(c.in)
		if err != nil {
			t.Fatalf("ParsePercent(%q): %v", c.in, err)
		}
		if got.PPM() != c.ppm {
			t.Errorf("ParsePercent(%q).PPM() = %d, want %d", c.in, got.PPM(), c.ppm)
		}
		if got.String() != c.text {
			t.Errorf("ParsePercent(%q).String() = %q, want %q", c.in, got.String(), c.text)
		}
	}
}

// An empty rate is not "0%": a missing tax rate silently becoming zero is
// exactly how an order ships untaxed. Same strictness as ParseMoney.
func TestParsePercent_rejectsEmptyOrMalformed(t *testing.T) {
	for _, in := range []string{"", "%", "  %  ", "+5", "abc", "1.23456", "5,5", "5."} {
		if _, err := shared.ParsePercent(in); !errors.Is(err, shared.ErrMalformedAmount) {
			t.Errorf("ParsePercent(%q): got %v, want ErrMalformedAmount", in, err)
		}
	}
}

// -0.5% is a small discount; printing it as "0.5000%" flips a discount into a
// surcharge in every log line and invoice.
func TestRate_stringKeepsSignBelowOnePercent(t *testing.T) {
	if got, want := shared.RatePPM(-5_000).String(), "-0.5000%"; got != want {
		t.Fatalf("RatePPM(-5000).String() = %q, want %q", got, want)
	}
}

func TestRate_plus1TurnsTaxIntoMultiplier(t *testing.T) {
	denver := shared.MustParsePercent("8.81")
	if got, want := denver.Plus1().String(), "108.8100%"; got != want {
		t.Fatalf("8.81%%.Plus1() = %s, want %s", got, want)
	}
}

// A zero rate converts every amount to nothing. It must be impossible to
// construct, not merely unlikely.
func TestExchangeRate_rejectsNonPositive(t *testing.T) {
	for _, in := range []string{"0", "0.0000", "-1", "-26000"} {
		if _, err := shared.NewExchangeRate(shared.USD, shared.VND, in); !errors.Is(err, shared.ErrInvalidRate) {
			t.Errorf("NewExchangeRate(%q): got %v, want ErrInvalidRate", in, err)
		}
	}
	for _, in := range []string{"", "abc", "26,000", "+26000"} {
		if _, err := shared.NewExchangeRate(shared.USD, shared.VND, in); !errors.Is(err, shared.ErrMalformedAmount) {
			t.Errorf("NewExchangeRate(%q): got %v, want ErrMalformedAmount", in, err)
		}
	}
	if _, err := shared.NewExchangeRate(shared.USD, shared.USD, "1"); !errors.Is(err, shared.ErrInvalidRate) {
		t.Errorf("USD->USD: got %v, want ErrInvalidRate", err)
	}
}

// Rate() is what the repository persists and what NewExchangeRate reads
// back. It is canonical, so two quotes written "26000.50" and "26000.5"
// compare equal with == and store as the same string.
func TestExchangeRate_rateIsCanonicalAndComparable(t *testing.T) {
	a := shared.MustExchangeRate(shared.USD, shared.VND, "26000.50")
	b := shared.MustExchangeRate(shared.USD, shared.VND, "26000.5")
	if a != b {
		t.Fatal("26000.50 and 26000.5 must be the same value object")
	}
	cases := map[string]string{
		"26000.50":       "26000.5",
		"26000":          "26000",
		"26000.00":       "26000",
		"0.0000384615":   "0.0000384615",
		"0.000000000001": "0.000000000001", // 12 decimals is the ceiling
	}
	for in, want := range cases {
		got := shared.MustExchangeRate(shared.USD, shared.VND, in).Rate()
		if got != want {
			t.Errorf("Rate(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := shared.NewExchangeRate(shared.USD, shared.VND, "0.0000000000001"); err == nil {
		t.Error("13 decimals must be rejected")
	}
}

func TestExchangeRate_stringReadsLikeAQuote(t *testing.T) {
	r := shared.MustExchangeRate(shared.USD, shared.VND, "26000.5")
	if got, want := r.String(), "1 USD = 26000.5 VND"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if r.From() != shared.USD || r.To() != shared.VND {
		t.Fatal("From/To getters")
	}
}
