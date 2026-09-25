// Package config is the ADAPTER that reads the business numbers a quote is
// built with — sales tax, margin, deposit, the forwarder's lanes and price
// list — from a YAML file, and hands them to the composition root as the
// domain's own value objects.
//
// It validates NOTHING itself (CLAUDE.md convention 8): every number goes
// through shared.ParsePercent / shared.ParseMoney and every object through
// pricing.New*. What this package adds is the NAME of the field that was
// wrong, so an operator editing the file reads "quote.deposit" and not a
// stack trace. A wrong file is the operator's input, hence errors, never
// panics (convention 1).
//
// The YAML library lives here and nowhere else: internal/domain/ keeps its
// stdlib-plus-uuid allowlist (decisions_test.go).
//
// [PHP] Đây là Configuration::getConfigTreeBuilder() + đoạn Extension::load()
// [PHP] gộp lại: đọc yaml, đặt tên trường vào lỗi, dựng value object. Khác một
// [PHP] chỗ: không có "default value" cho số nghiệp vụ — thiếu là lỗi.
package config

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Pricing is what the file becomes: the three things wire plugs into
// pricingapp.Deps and into the seed.
type Pricing struct {
	Policy         pricing.QuotePolicy
	Classification pricing.Classification
	Lanes          []pricing.LaneDetails
}

// file is the YAML shape. Every number is a string on purpose: yaml.v3 would
// otherwise hand us a float64, and money never goes through a float
// (convention 2). The tags are the operator's vocabulary; the Go names the
// domain's.
//
// [PHP] Tag `yaml:"sales_tax"` ~ #[SerializedName('sales_tax')]: tên trong
// [PHP] file khác tên field, và chỉ có ở đây, không lan vào domain.
type file struct {
	Quote   quoteSection      `yaml:"quote"`
	Classes map[string]string `yaml:"classes"`
	Lanes   []laneSection     `yaml:"lanes"`
}

type quoteSection struct {
	SalesTax       string `yaml:"sales_tax"`
	MarginPercent  string `yaml:"margin_percent"`
	MarginFloorVND string `yaml:"margin_floor_vnd"`
	Deposit        string `yaml:"deposit"`
	TTLHours       int    `yaml:"ttl_hours"`
}

type laneSection struct {
	Code             string      `yaml:"code"`
	Name             string      `yaml:"name"`
	Currency         string      `yaml:"currency"`
	Divisor          int64       `yaml:"divisor"`
	StepG            int64       `yaml:"step_g"`
	Rates            rateSection `yaml:"rates"`
	BatterySurcharge string      `yaml:"battery_surcharge"`
	Duty             string      `yaml:"duty"`
	DutyRate         string      `yaml:"duty_rate"`
}

type rateSection struct {
	Standard    string `yaml:"standard"`
	Branded     string `yaml:"branded"`
	Electronics string `yaml:"electronics"`
	Sensitive   string `yaml:"sensitive"`
}

// Load reads and parses the rate card at path. The error names the file, so
// "no such file" points at the flag that chose it.
func Load(path string) (Pricing, error) {
	f, err := os.Open(path)
	if err != nil {
		return Pricing{}, fmt.Errorf("ratecard %s: %w", path, err)
	}
	defer f.Close()
	p, err := Parse(f)
	if err != nil {
		return Pricing{}, fmt.Errorf("ratecard %s: %w", path, err)
	}
	return p, nil
}

// Parse builds the domain values from YAML. Unknown keys are an error: a
// misspelt "deposit" must not become "no deposit".
func Parse(r io.Reader) (Pricing, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	var in file
	if err := dec.Decode(&in); err != nil {
		return Pricing{}, err // the library already says "yaml:" and the line
	}

	policy, err := quotePolicy(in.Quote)
	if err != nil {
		return Pricing{}, err
	}
	classes, err := classification(in.Classes)
	if err != nil {
		return Pricing{}, err
	}
	if len(in.Lanes) == 0 {
		return Pricing{}, fmt.Errorf("lanes: want at least one lane: %w", pricing.ErrInvalidLane)
	}
	lanes := make([]pricing.LaneDetails, 0, len(in.Lanes))
	for i, l := range in.Lanes {
		d, err := laneDetails(i, l)
		if err != nil {
			return Pricing{}, err
		}
		lanes = append(lanes, d)
	}
	return Pricing{Policy: policy, Classification: classes, Lanes: lanes}, nil
}

// field prefixes an error with the YAML path that caused it — the one thing
// this package adds to the domain's own message.
func field(path string, err error) error {
	return fmt.Errorf("%s: %w", path, err)
}

func quotePolicy(q quoteSection) (pricing.QuotePolicy, error) {
	salesTax, err := shared.ParsePercent(q.SalesTax)
	if err != nil {
		return pricing.QuotePolicy{}, field("quote.sales_tax", err)
	}
	percent, err := shared.ParsePercent(q.MarginPercent)
	if err != nil {
		return pricing.QuotePolicy{}, field("quote.margin_percent", err)
	}
	floor, err := shared.ParseMoney(q.MarginFloorVND, shared.VND)
	if err != nil {
		return pricing.QuotePolicy{}, field("quote.margin_floor_vnd", err)
	}
	margin, err := pricing.NewMarginPolicy(percent, floor)
	if err != nil {
		return pricing.QuotePolicy{}, field("quote.margin_percent/margin_floor_vnd", err)
	}
	deposit, err := shared.ParsePercent(q.Deposit)
	if err != nil {
		return pricing.QuotePolicy{}, field("quote.deposit", err)
	}
	policy, err := pricing.NewQuotePolicy(pricing.QuotePolicyDetails{
		SalesTax: salesTax,
		Margin:   margin,
		Deposit:  deposit,
		TTL:      time.Duration(q.TTLHours) * time.Hour,
	})
	if err != nil {
		// NewQuotePolicy names the field it rejected ("deposit 150%: …");
		// the prefix says which section of the file to open.
		return pricing.QuotePolicy{}, field("quote", err)
	}
	return policy, nil
}

func classification(in map[string]string) (pricing.Classification, error) {
	byCategory := make(map[string]pricing.GoodsClass, len(in))
	for category, class := range in {
		byCategory[category] = pricing.GoodsClass(strings.ToLower(strings.TrimSpace(class)))
	}
	c, err := pricing.NewClassification(byCategory)
	if err != nil {
		return pricing.Classification{}, field("classes", err)
	}
	return c, nil
}

func laneDetails(i int, l laneSection) (pricing.LaneDetails, error) {
	at := func(name string) string { return fmt.Sprintf("lanes[%d].%s", i, name) }
	var zero pricing.LaneDetails

	code, err := pricing.ParseLaneCode(l.Code)
	if err != nil {
		return zero, field(at("code"), err)
	}
	// CurrencyFromCode is case-sensitive by design; normalising here is the
	// adapter's job, not the domain's.
	currency, err := shared.CurrencyFromCode(strings.ToUpper(strings.TrimSpace(l.Currency)))
	if err != nil {
		return zero, field(at("currency"), err)
	}
	step, err := shared.NewWeight(l.StepG)
	if err != nil {
		return zero, field(at("step_g"), err)
	}
	money := func(name, s string) (shared.Money, error) {
		m, err := shared.ParseMoney(s, currency)
		if err != nil {
			return shared.Money{}, field(at(name), err)
		}
		return m, nil
	}
	standard, err := money("rates.standard", l.Rates.Standard)
	if err != nil {
		return zero, err
	}
	branded, err := money("rates.branded", l.Rates.Branded)
	if err != nil {
		return zero, err
	}
	electronics, err := money("rates.electronics", l.Rates.Electronics)
	if err != nil {
		return zero, err
	}
	sensitive, err := money("rates.sensitive", l.Rates.Sensitive)
	if err != nil {
		return zero, err
	}
	rates, err := pricing.NewRateCard(standard, branded, electronics, sensitive)
	if err != nil {
		return zero, field(at("rates"), err)
	}
	var surcharge shared.Money // Money{} = no surcharge on this lane
	if strings.TrimSpace(l.BatterySurcharge) != "" {
		if surcharge, err = money("battery_surcharge", l.BatterySurcharge); err != nil {
			return zero, err
		}
	}
	duty, err := dutyPolicy(l.Duty, l.DutyRate)
	if err != nil {
		return zero, field(at("duty"), err)
	}
	d := pricing.LaneDetails{
		Code: code, Name: l.Name, Divisor: l.Divisor, Step: step, Rates: rates,
		BatterySurcharge: surcharge, Duty: duty,
	}
	// Run the aggregate's own constructor now, so a lane the seed would
	// reject is rejected here, at load, with the file's name on the error.
	if _, err := pricing.NewShippingLane(d); err != nil {
		return zero, field(at(""), err)
	}
	return d, nil
}

func dutyPolicy(kind, rate string) (pricing.DutyPolicy, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "bundled":
		return pricing.DutyBundled(), nil
	case "itemised", "itemized":
		r, err := shared.ParsePercent(rate)
		if err != nil {
			return pricing.DutyPolicy{}, fmt.Errorf("duty_rate: %w", err)
		}
		return pricing.DutyItemised(r)
	default:
		return pricing.DutyPolicy{}, fmt.Errorf("%q: want bundled or itemised: %w", kind, pricing.ErrInvalidLane)
	}
}
