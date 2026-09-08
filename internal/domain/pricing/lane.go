package pricing

import (
	"fmt"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// LaneCode is a shipping lane's natural key: "us_forwarder". Same shape and
// same reasons as catalog.CategoryCode — a struct, so only ParseLaneCode can
// make one.
type LaneCode struct {
	code string
}

func ParseLaneCode(s string) (LaneCode, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if !isSnakeCase(s) {
		return LaneCode{}, fmt.Errorf("lane code %q: want ^[a-z][a-z0-9_]*$: %w", s, ErrInvalidLaneCode)
	}
	return LaneCode{code: s}, nil
}

func MustParseLaneCode(s string) LaneCode {
	c, err := ParseLaneCode(s)
	if err != nil {
		panic(err)
	}
	return c
}

func (c LaneCode) String() string {
	return c.code
}

func (c LaneCode) IsZero() bool {
	return c.code == ""
}

func isSnakeCase(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case i > 0 && (r == '_' || (r >= '0' && r <= '9')):
		default:
			return false
		}
	}
	return true
}

// GoodsClass is how the FORWARDER groups goods on its price list — not how the
// catalogue groups them. "thường / hàng hiệu / điện tử / nhạy cảm" is their
// vocabulary (SETUP.md §7); a Classification maps our categories onto it.
type GoodsClass string

const (
	ClassStandard    GoodsClass = "standard"
	ClassBranded     GoodsClass = "branded"
	ClassElectronics GoodsClass = "electronics"
	ClassSensitive   GoodsClass = "sensitive"
)

func (c GoodsClass) IsValid() bool {
	switch c {
	case ClassStandard, ClassBranded, ClassElectronics, ClassSensitive:
		return true
	}
	return false
}

// RateCard is the price per kilogram for each goods class — one value object,
// four fields, one currency. A field per class instead of a map so the card is
// comparable (==) and cannot be missing a class.
type RateCard struct {
	standard, branded, electronics, sensitive shared.Money
}

func NewRateCard(standard, branded, electronics, sensitive shared.Money) (RateCard, error) {
	all := []shared.Money{standard, branded, electronics, sensitive}
	for _, m := range all {
		if !m.IsValid() || m.IsNegative() || m.IsZero() {
			return RateCard{}, fmt.Errorf("rate %s: %w", m, ErrInvalidRate)
		}
		if m.Currency() != standard.Currency() {
			return RateCard{}, fmt.Errorf("rate %s vs %s: %w", m, standard, shared.ErrCurrencyMismatch)
		}
	}
	return RateCard{standard: standard, branded: branded, electronics: electronics, sensitive: sensitive}, nil
}

func (r RateCard) PerKg(c GoodsClass) shared.Money {
	switch c {
	case ClassBranded:
		return r.branded
	case ClassElectronics:
		return r.electronics
	case ClassSensitive:
		return r.sensitive
	default:
		return r.standard
	}
}

func (r RateCard) Currency() shared.Currency {
	return r.standard.Currency()
}

func (r RateCard) IsZero() bool {
	return r == RateCard{}
}

// DutyPolicy says how import duty appears on this lane. The forwarder lane in
// use BUNDLES clearance into the per-kg price (no duty line at all); a
// self-managed commercial import would itemise it from an HS code. This is
// the field catalog rightly refused to own (CATALOG.md §6): same shoe, two
// lanes, two answers.
type DutyPolicy struct {
	itemised bool
	rate     shared.Rate
}

// DutyBundled is the zero value: the safe default and the lane in use.
func DutyBundled() DutyPolicy {
	return DutyPolicy{}
}

func DutyItemised(rate shared.Rate) (DutyPolicy, error) {
	if rate.PPM() < 0 {
		return DutyPolicy{}, fmt.Errorf("duty %s: %w", rate, ErrInvalidRate)
	}
	return DutyPolicy{itemised: true, rate: rate}, nil
}

func (d DutyPolicy) Itemised() bool {
	return d.itemised
}

func (d DutyPolicy) Rate() shared.Rate {
	return d.rate
}

// LaneDetails is what an operator enters for a lane (convention 10).
type LaneDetails struct {
	Code             LaneCode
	Name             string
	Divisor          int64         // cm³ per kg the carrier bills on: 5000, 6000
	Step             shared.Weight // billing step: 100 g, 500 g
	Rates            RateCard
	BatterySurcharge shared.Money // per parcel with a lithium cell; Money{} = none
	Duty             DutyPolicy   // zero = bundled
}

// ShippingLane is one way goods travel from Denver to a door in Vietnam, and
// everything the carrier's price list says about it: the volumetric divisor,
// the billing step, the rate per kg per goods class, surcharges, duty rule.
//
// It is an immutable value object with a natural key, like CategoryPolicy:
// the shipper sends a new price list, the operator defines the lane again.
// Quotes SNAPSHOT the numbers they used, so redefining a lane never changes a
// quote already issued (DDD.md §28).
type ShippingLane struct {
	code      LaneCode
	name      string
	divisor   int64
	step      shared.Weight
	rates     RateCard
	surcharge shared.Money
	duty      DutyPolicy
}

func NewShippingLane(d LaneDetails) (ShippingLane, error) {
	bad := func(why string) error { return fmt.Errorf("lane %s: %s: %w", d.Code, why, ErrInvalidLane) }
	switch {
	case d.Code.IsZero():
		return ShippingLane{}, fmt.Errorf("lane: no code: %w", ErrInvalidLane)
	case strings.TrimSpace(d.Name) == "":
		return ShippingLane{}, bad("no name")
	case d.Divisor <= 0:
		return ShippingLane{}, bad("divisor must be positive")
	case d.Step.IsZero():
		return ShippingLane{}, bad("billing step must be positive")
	case d.Rates.IsZero():
		return ShippingLane{}, bad("no rate card")
	}
	if d.BatterySurcharge.IsValid() {
		if d.BatterySurcharge.IsNegative() {
			return ShippingLane{}, bad("negative surcharge")
		}
		if d.BatterySurcharge.Currency() != d.Rates.Currency() {
			return ShippingLane{}, fmt.Errorf("lane %s: surcharge %s vs rates in %s: %w",
				d.Code, d.BatterySurcharge, d.Rates.Currency(), shared.ErrCurrencyMismatch)
		}
	}
	return ShippingLane{
		code: d.Code, name: strings.TrimSpace(d.Name), divisor: d.Divisor, step: d.Step,
		rates: d.Rates, surcharge: d.BatterySurcharge, duty: d.Duty,
	}, nil
}

func (l ShippingLane) Code() LaneCode {
	return l.code
}

func (l ShippingLane) Name() string {
	return l.name
}

func (l ShippingLane) Currency() shared.Currency {
	return l.rates.Currency()
}

func (l ShippingLane) Divisor() int64 {
	return l.divisor
}

func (l ShippingLane) Step() shared.Weight {
	return l.step
}

func (l ShippingLane) Rates() RateCard {
	return l.rates
}

func (l ShippingLane) BatterySurcharge() shared.Money {
	return l.surcharge
}

// DutyPolicy is the rule; Duty (below) is the amount it yields for an item.
func (l ShippingLane) DutyPolicy() DutyPolicy {
	return l.duty
}

func (l ShippingLane) IsZero() bool {
	return l.code.IsZero()
}

// Defined is the announcement of this lane (see catalog.CategoryPolicy.Defined):
// a value object has no event log of its own, so the use case that saves it
// asks for the event to append. Rates stay home.
func (l ShippingLane) Defined(at time.Time) LaneDefinedEvent {
	return LaneDefinedEvent{Code: l.code, Name: l.name, Divisor: l.divisor, Step: l.step, Currency: l.Currency(), At: at}
}

// ChargeableWeight is the number the carrier bills on: the greater of actual
// and volumetric weight, rounded UP to this lane's step. The maths is the
// shared kernel's; the two carrier constants are this lane's.
func (l ShippingLane) ChargeableWeight(spec shared.ParcelSpec) shared.Weight {
	return shared.ChargeableWeight(spec.Weight(), spec.Dimensions(), l.divisor, l.step)
}

// Freight is rate per kg × kilograms, exact: 2.5 kg is 2 500 000 ppm of a kg.
//
// [PHP] Không nhân float: 2500 g → RatePPM(2 500 000) → Money.Mul làm tròn
// [PHP] half-up trên int64. Cùng kỹ thuật với thuế.
func (l ShippingLane) Freight(class GoodsClass, w shared.Weight) shared.Money {
	return l.rates.PerKg(class).Mul(shared.RatePPM(w.Grams() * 1000))
}

// Surcharge applies the lane's extras for what the goods are (restrictions
// travel with the category, DDD.md §23). Today: lithium batteries.
func (l ShippingLane) Surcharge(restrictions []string) shared.Money {
	for _, r := range restrictions {
		if r == "battery" && l.surcharge.IsValid() {
			return l.surcharge
		}
	}
	return shared.Zero(l.Currency())
}

// Duty is the duty line for an item on this lane: nothing when bundled.
func (l ShippingLane) Duty(item shared.Money) shared.Money {
	if !l.duty.itemised {
		return shared.Zero(item.Currency())
	}
	return item.Mul(l.duty.rate)
}

// Classification maps OUR category codes onto the forwarder's goods classes.
// Unknown categories are standard: the cheapest class, so a missing row shows
// up as a too-low quote in reconciliation rather than a customer overcharged.
type Classification struct {
	byCategory map[string]GoodsClass
}

func NewClassification(byCategory map[string]GoodsClass) (Classification, error) {
	out := make(map[string]GoodsClass, len(byCategory))
	for code, class := range byCategory {
		if !class.IsValid() {
			return Classification{}, fmt.Errorf("category %q → %q: %w", code, class, ErrUnknownGoodsClass)
		}
		out[code] = class
	}
	return Classification{byCategory: out}, nil
}

func (c Classification) ClassOf(category string) GoodsClass {
	if class, ok := c.byCategory[category]; ok {
		return class
	}
	return ClassStandard
}
