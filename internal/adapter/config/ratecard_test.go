package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/config"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// good is the smallest complete rate card, in two halves so a case can drop
// the lanes altogether. The table below mutates one line at a time, so every
// rejection is tied to the one field that caused it.
const head = `
quote:
  sales_tax: "8.81"
  margin_percent: "10"
  margin_floor_vnd: "500000"
  deposit: "50"
  ttl_hours: 48
classes:
  footwear: branded
  electronics: electronics
`

const lanesBlock = `lanes:
  - code: us_forwarder
    name: "US forwarder"
    currency: USD
    divisor: 5000
    step_g: 500
    rates: {standard: "9.00", branded: "10.00", electronics: "12.00", sensitive: "14.00"}
    battery_surcharge: "3.00"
    duty: bundled
`

const good = head + lanesBlock

func TestParse_readsEveryValueExactly(t *testing.T) {
	p, err := config.Parse(strings.NewReader(good))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Compare in ppm and minor units — the domain's own representation — so
	// a float sneaking in anywhere would show as a wrong integer, not as a
	// rounding difference that happens to pass.
	if got, want := p.Policy.SalesTax().PPM(), shared.MustParsePercent("8.81").PPM(); got != want {
		t.Errorf("sales tax %d ppm, want %d", got, want)
	}
	if got, want := p.Policy.Margin().Percent().PPM(), shared.MustParsePercent("10").PPM(); got != want {
		t.Errorf("margin %d ppm, want %d", got, want)
	}
	if got, want := p.Policy.Margin().Floor(), shared.MustParseMoney("500000", shared.VND); got != want {
		t.Errorf("margin floor %s, want %s", got, want)
	}
	if got, want := p.Policy.Deposit().PPM(), shared.MustParsePercent("50").PPM(); got != want {
		t.Errorf("deposit %d ppm, want %d", got, want)
	}
	if got, want := p.Policy.TTL(), 48*time.Hour; got != want {
		t.Errorf("ttl %s, want %s", got, want)
	}
	for category, want := range map[string]pricing.GoodsClass{
		"footwear": pricing.ClassBranded, "electronics": pricing.ClassElectronics, "unlisted": pricing.ClassStandard,
	} {
		if got := p.Classification.ClassOf(category); got != want {
			t.Errorf("class of %q = %q, want %q", category, got, want)
		}
	}
	if len(p.Lanes) != 1 {
		t.Fatalf("%d lanes, want 1", len(p.Lanes))
	}
	lane := p.Lanes[0]
	if lane.Code.String() != "us_forwarder" || lane.Divisor != 5000 || lane.Step.Grams() != 500 {
		t.Errorf("lane %+v: code/divisor/step wrong", lane)
	}
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	for class, want := range map[pricing.GoodsClass]shared.Money{
		pricing.ClassStandard: usd("9.00"), pricing.ClassBranded: usd("10.00"),
		pricing.ClassElectronics: usd("12.00"), pricing.ClassSensitive: usd("14.00"),
	} {
		if got := lane.Rates.PerKg(class); got != want {
			t.Errorf("rate %s = %s, want %s", class, got, want)
		}
	}
	if lane.BatterySurcharge != usd("3.00") || lane.Duty.Itemised() {
		t.Errorf("surcharge %s / itemised %v: want 3.00 USD, bundled", lane.BatterySurcharge, lane.Duty.Itemised())
	}
}

// Each row breaks exactly one thing. The error must name the field an
// operator would open the file to fix — that is the whole contract of this
// adapter over the domain's own validation.
func TestParse_rejectsOneWrongFieldByName(t *testing.T) {
	cases := []struct {
		name string
		from string // line in `good` to replace…
		to   string
		file string // …or a whole file instead, when one replacement cannot express the case
		want string // substring of the error
	}{
		{name: "deposit over 100 %", from: `deposit: "50"`, to: `deposit: "150"`, want: "quote"},
		{name: "deposit not a number", from: `deposit: "50"`, to: `deposit: "half"`, want: "quote.deposit"},
		{name: "negative rate", from: `standard: "9.00"`, to: `standard: "-9.00"`, want: "lanes[0].rates"},
		{name: "missing rate", from: `standard: "9.00", `, to: ``, want: "lanes[0].rates.standard"},
		{name: "unknown goods class", from: `footwear: branded`, to: `footwear: luxury`, want: "classes"},
		{name: "unknown key", from: `ttl_hours: 48`, to: "ttl_hours: 48\n  deposit_pct: \"50\"", want: "deposit_pct"},
		{name: "no lanes", file: head + "lanes: []\n", want: "lanes"},
		{name: "bad lane code", from: `code: us_forwarder`, to: `code: "US Forwarder"`, want: "lanes[0].code"},
		{name: "unknown currency", from: `currency: USD`, to: `currency: EUR`, want: "lanes[0].currency"},
		{name: "unknown duty", from: `duty: bundled`, to: `duty: prepaid`, want: "lanes[0].duty"},
		{name: "zero divisor", from: `divisor: 5000`, to: `divisor: 0`, want: "lanes[0]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.file
			if in == "" {
				if !strings.Contains(good, tc.from) {
					t.Fatalf("test bug: %q is not in the good file", tc.from)
				}
				in = strings.Replace(good, tc.from, tc.to, 1)
			}
			_, err := config.Parse(strings.NewReader(in))
			if err == nil {
				t.Fatal("parsed a rate card with a wrong field")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not name %q", err, tc.want)
			}
		})
	}
}

// A number written without quotes must still be read exactly, not as a
// float64 that yaml would otherwise produce — the file says "write strings"
// but an operator will forget, and forgetting must not change a price.
func TestParse_unquotedNumbersStayExact(t *testing.T) {
	in := strings.NewReplacer(`sales_tax: "8.81"`, `sales_tax: 8.81`, `standard: "9.00"`, `standard: 9.00`).Replace(good)
	p, err := config.Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := p.Policy.SalesTax().PPM(), shared.MustParsePercent("8.81").PPM(); got != want {
		t.Errorf("unquoted sales tax %d ppm, want %d", got, want)
	}
	if got, want := p.Lanes[0].Rates.PerKg(pricing.ClassStandard), shared.MustParseMoney("9.00", shared.USD); got != want {
		t.Errorf("unquoted rate %s, want %s", got, want)
	}
}

func TestLoad_namesTheFileOnEveryError(t *testing.T) {
	_, err := config.Load("testdata/does-not-exist.yaml")
	if err == nil || !strings.Contains(err.Error(), "does-not-exist.yaml") {
		t.Fatalf("error %v does not name the missing file", err)
	}
}
