package shared_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
)

func TestParseMoney_keepsExactValue(t *testing.T) {
	cases := []struct {
		in    string
		cur   shared.Currency
		minor int64
		text  string
	}{
		{"150.00", shared.USD, 15000, "150.00 USD"},
		{"150.5", shared.USD, 15050, "150.50 USD"},
		{"0.01", shared.USD, 1, "0.01 USD"},
		{"3900000", shared.VND, 3900000, "3900000 VND"},
		{"-25.30", shared.USD, -2530, "-25.30 USD"},
	}
	for _, c := range cases {
		got, err := shared.ParseMoney(c.in, c.cur)
		if err != nil {
			t.Fatalf("ParseMoney(%q): %v", c.in, err)
		}
		if got.Minor() != c.minor {
			t.Errorf("ParseMoney(%q).Minor() = %d, want %d", c.in, got.Minor(), c.minor)
		}
		if got.String() != c.text {
			t.Errorf("ParseMoney(%q).String() = %q, want %q", c.in, got.String(), c.text)
		}
	}
}

// VND has no subunit. A Money type that assumes 2 decimals everywhere
// would silently accept "3900000.50 VND", which does not exist.
func TestParseMoney_rejectsDecimalsVND(t *testing.T) {
	if _, err := shared.ParseMoney("3900000.50", shared.VND); err == nil {
		t.Fatal("want error for fractional VND, got nil")
	}
}

func TestMoney_addRefusesMixedCurrency(t *testing.T) {
	usd := shared.MustParseMoney("150.00", shared.USD)
	vnd := shared.MustParseMoney("3900000", shared.VND)

	_, err := usd.Add(vnd)
	if !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Fatalf("adding VND to USD: got %v, want ErrCurrencyMismatch", err)
	}
}

// Denver sales tax on a $150 pair of shoes.
func TestMoney_mulAppliesTaxRate(t *testing.T) {
	price := shared.MustParseMoney("150.00", shared.USD)
	denver := shared.MustParsePercent("8.81")

	tax := price.Mul(denver)
	if got, want := tax.String(), "13.22 USD"; got != want {
		t.Fatalf("150.00 USD * 8.81%% = %s, want %s", got, want)
	}
}

func TestMoney_convertUSDtoVND(t *testing.T) {
	price := shared.MustParseMoney("150.00", shared.USD)
	rate := shared.MustExchangeRate(shared.USD, shared.VND, "26000", "bank", "2026-09-03")

	vnd, err := price.Convert(rate)
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if got, want := vnd.String(), "3900000 VND"; got != want {
		t.Fatalf("150.00 USD at 26000 = %s, want %s", got, want)
	}
}

func TestMoney_convertRejectsWrongSourceCurrency(t *testing.T) {
	vnd := shared.MustParseMoney("3900000", shared.VND)
	usdRate := shared.MustExchangeRate(shared.USD, shared.VND, "26000", "bank", "2026-09-03")

	if _, err := vnd.Convert(usdRate); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Fatalf("got %v, want ErrCurrencyMismatch", err)
	}
}

// Money never uses float64. This is the value that proves why.
func TestMoney_noFloatingPointDrift(t *testing.T) {
	ten := shared.MustParseMoney("0.10", shared.USD)
	twenty := shared.MustParseMoney("0.20", shared.USD)

	sum, err := ten.Add(twenty)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Minor() != 30 {
		t.Fatalf("0.10 + 0.20 = %s (%d minor), want 0.30 USD (30 minor)", sum, sum.Minor())
	}
	if got := 0.1 + 0.2; got == 0.3 {
		t.Log("note: this build's float64 got lucky; the domain still must not rely on it")
	}
}
