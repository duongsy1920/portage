package shared_test

import (
	"errors"
	"math"
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
	rate := shared.MustExchangeRate(shared.USD, shared.VND, "26000")

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
	usdRate := shared.MustExchangeRate(shared.USD, shared.VND, "26000")

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
	// And the counter-example, computed at runtime (constants would be folded
	// exactly by the compiler and prove nothing): float64 really does drift.
	a, b := 0.1, 0.2
	if a+b == 0.3 {
		t.Fatal("expected float64 0.1+0.2 != 0.3; this test's premise is wrong")
	}
}

// A decimal comma is a locale convention, not a number. Stripping commas
// would turn "150,50" (a Vietnamese user meaning 150.50) into 15050.00 USD —
// a 100x error that no test downstream would catch. The domain accepts one
// canonical format and refuses everything else; locale handling belongs in
// the adapter that talks to humans.
func TestParseMoney_rejectsAmbiguousInput(t *testing.T) {
	for _, in := range []string{"150,50", "1,000", "+150", "", "   ", "150.", ".5", "1.2.3", "abc", "1e3", "- 5"} {
		if _, err := shared.ParseMoney(in, shared.USD); !errors.Is(err, shared.ErrMalformedAmount) {
			t.Errorf("ParseMoney(%q): got %v, want ErrMalformedAmount", in, err)
		}
	}
}

// Persistence stores "USD", not a Currency struct. The adapter needs a way
// back from the code to the value object — and an unknown code must fail
// loudly, not become a currency with exponent 0.
func TestCurrencyFromCode(t *testing.T) {
	if c, err := shared.CurrencyFromCode("USD"); err != nil || c != shared.USD {
		t.Fatalf("CurrencyFromCode(USD) = %v, %v", c, err)
	}
	if c, err := shared.CurrencyFromCode("VND"); err != nil || c != shared.VND {
		t.Fatalf("CurrencyFromCode(VND) = %v, %v", c, err)
	}
	for _, code := range []string{"usd", "XXX", "", "US"} {
		if _, err := shared.CurrencyFromCode(code); !errors.Is(err, shared.ErrUnknownCurrency) {
			t.Errorf("CurrencyFromCode(%q): got %v, want ErrUnknownCurrency", code, err)
		}
	}
}

// Currency{} is a legal Go literal but not a legal currency. Money built on
// it would happily Add to other Currency{} money and print "500 ".
func TestNewMoney_refusesZeroCurrency(t *testing.T) {
	mustPanic(t, func() { shared.NewMoney(500, shared.Currency{}) })
	mustPanic(t, func() { shared.Zero(shared.Currency{}) })
}

func TestMoney_zeroValueIsNotValid(t *testing.T) {
	if (shared.Money{}).IsValid() {
		t.Fatal("Money{} must not be valid")
	}
	if !shared.NewMoney(0, shared.USD).IsValid() {
		t.Fatal("0 USD must be valid")
	}
}

func TestMoney_subAndSum(t *testing.T) {
	a := shared.MustParseMoney("150.00", shared.USD)
	b := shared.MustParseMoney("25.30", shared.USD)

	diff, err := a.Sub(b)
	if err != nil || diff.String() != "124.70 USD" {
		t.Fatalf("150.00 - 25.30 = %s, %v", diff, err)
	}
	total, err := shared.Sum(a, b, b)
	if err != nil || total.String() != "200.60 USD" {
		t.Fatalf("Sum = %s, %v", total, err)
	}
	vnd := shared.MustParseMoney("1000", shared.VND)
	if _, err := a.Sub(vnd); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("Sub across currencies: got %v", err)
	}
	if _, err := shared.Sum(a, vnd); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("Sum across currencies: got %v", err)
	}
}

// Half-up rounding mirrors around zero: 13.215 -> 13.22 and -13.215 -> -13.22.
// (A refund of a taxed amount must equal the tax charged, sign flipped.)
func TestMoney_mulRoundsHalfUpAwayFromZero(t *testing.T) {
	rate := shared.MustParsePercent("8.81")
	if got := shared.MustParseMoney("150.00", shared.USD).Mul(rate).String(); got != "13.22 USD" {
		t.Errorf("150.00 * 8.81%% = %s", got)
	}
	if got := shared.MustParseMoney("-150.00", shared.USD).Mul(rate).String(); got != "-13.22 USD" {
		t.Errorf("-150.00 * 8.81%% = %s", got)
	}
}

// The intermediate product must not wrap: 1<<62 cents at 100% used to come
// back as 0.00 USD. And when the RESULT itself exceeds int64 (92 quadrillion
// dollars) that is a bug, which must panic rather than return garbage.
func TestMoney_mulDoesNotOverflowSilently(t *testing.T) {
	huge := shared.NewMoney(1<<62, shared.USD)
	if got := huge.Mul(shared.RatePPM(1_000_000)); got != huge {
		t.Fatalf("100%% of %s = %s", huge, got)
	}
	mustPanic(t, func() { huge.Mul(shared.RatePPM(3_000_000)) })
}

// The reverse direction: refunds are paid in VND, booked in USD. 1/26000 has
// no 4-decimal form, so the rate must keep more precision than a percent.
func TestMoney_convertVNDtoUSD(t *testing.T) {
	rate := shared.MustExchangeRate(shared.VND, shared.USD, "0.0000384615")
	got, err := shared.MustParseMoney("3900000", shared.VND).Convert(rate)
	if err != nil || got.String() != "150.00 USD" {
		t.Fatalf("3900000 VND at 0.0000384615 = %s, %v", got, err)
	}
}

func TestMoney_convertRoundsHalfUp(t *testing.T) {
	rate := shared.MustExchangeRate(shared.USD, shared.VND, "26000.05")
	got, _ := shared.MustParseMoney("10.00", shared.USD).Convert(rate) // 260000.5
	if got.String() != "260001 VND" {
		t.Fatalf("10.00 USD at 26000.05 = %s, want 260001 VND", got)
	}
}

// int64 wraps silently: MaxInt64 + 1 is a large NEGATIVE number. A balance
// that goes negative because of a wrap is worse than a crash, so Add and Sub
// panic the way Mul and Convert already do.
func TestMoney_addAndSubPanicOnOverflow(t *testing.T) {
	max := shared.NewMoney(math.MaxInt64, shared.VND)
	min := shared.NewMoney(math.MinInt64, shared.VND)
	one := shared.NewMoney(1, shared.VND)

	mustPanic(t, func() { _, _ = max.Add(one) })
	mustPanic(t, func() { _, _ = min.Sub(one) })
	mustPanic(t, func() { _, _ = shared.Sum(max, one) })

	// Nothing near the boundary should trip: these are ordinary sums.
	if got, err := max.Sub(one); err != nil || got.Minor() != math.MaxInt64-1 {
		t.Fatalf("MaxInt64 - 1 = %v, %v", got.Minor(), err)
	}
}

// Money{} has no currency, so two of them must not "agree" with each other
// and produce a currencyless amount that prints as "0 ".
func TestMoney_arithmeticRejectsZeroValue(t *testing.T) {
	var unset shared.Money
	real := shared.MustParseMoney("150.00", shared.USD)

	for name, err := range map[string]error{
		"unset + unset": second(unset.Add(unset)),
		"unset + real":  second(unset.Add(real)),
		"real + unset":  second(real.Add(unset)),
		"real - unset":  second(real.Sub(unset)),
		"Sum(unset...)": second(shared.Sum(unset, unset)),
	} {
		if !errors.Is(err, shared.ErrCurrencyMismatch) {
			t.Errorf("%s: got %v, want ErrCurrencyMismatch", name, err)
		}
	}
}

func second(_ shared.Money, err error) error { return err }

func TestMoney_isPositive(t *testing.T) {
	for _, c := range []struct {
		in   shared.Money
		want bool
	}{
		{shared.MustParseMoney("0.01", shared.USD), true},
		{shared.Zero(shared.USD), false},
		{shared.NewMoney(-1, shared.USD), false},
		{shared.Money{}, false},
	} {
		if got := c.in.IsPositive(); got != c.want {
			t.Errorf("%v.IsPositive() = %v, want %v", c.in, got, c.want)
		}
	}
}
