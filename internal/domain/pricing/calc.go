package pricing

import (
	"fmt"

	"github.com/duongsy/portage/internal/domain/shared"
)

// QuoteInputs is everything a quote is computed from. Every field is a
// snapshot the Quote will keep: change any of them tomorrow, this quote's
// numbers stay.
type QuoteInputs struct {
	Listing Listing
	Profile CategoryProfile
	Lane    ShippingLane
	FX      shared.ExchangeRate // lane currency → home currency
	Policy  QuotePolicy
}

// Breakdown is the quote, line by line — the value object a customer sees and
// the record reconciliation compares against actual costs. Exported fields:
// it is a result, computed once by Calculate, never edited.
type Breakdown struct {
	Class      GoodsClass
	Chargeable shared.Weight
	Estimated  bool // parcel came from the category default, not a scale

	ItemPrice   shared.Money // lane currency
	SalesTax    shared.Money
	Freight     shared.Money
	Surcharge   shared.Money
	Duty        shared.Money
	SubtotalUSD shared.Money // "landed" cost in the lane's currency

	FX          shared.ExchangeRate
	SubtotalVND shared.Money // home currency
	ServiceFee  shared.Money
	TotalVND    shared.Money
	Deposit     shared.Money
}

// Calculate is the DOMAIN SERVICE at the heart of pricing (DDD.md §17): a pure
// function from inputs to a Breakdown. No clock, no repository, no state —
// which is why it can be checked line by line against a spreadsheet.
//
//	chargeable = lane.ChargeableWeight(parcel)          parcel = measured, else category default
//	subtotal   = item + item×tax + freight(class, chargeable) + surcharge + duty
//	subtotalₕ  = subtotal × fx
//	total      = subtotalₕ + max(subtotalₕ × margin%, floor)
//	deposit    = total × deposit%
func Calculate(in QuoteInputs) (Breakdown, error) {
	if !in.Listing.Active {
		return Breakdown{}, fmt.Errorf("product %s: %w", in.Listing.Product, ErrListingInactive)
	}
	if in.Lane.IsZero() || in.Policy.IsZero() {
		return Breakdown{}, fmt.Errorf("quote: lane or policy missing: %w", ErrInvalidLane)
	}
	item := in.Listing.Price
	if !item.IsValid() || item.Currency() != in.Lane.Currency() {
		return Breakdown{}, fmt.Errorf("price %s on a %s lane: %w", item, in.Lane.Currency(), shared.ErrCurrencyMismatch)
	}
	home := in.Policy.Margin().Floor().Currency()
	if in.FX.From() != in.Lane.Currency() || in.FX.To() != home {
		return Breakdown{}, fmt.Errorf("fx %s for a %s→%s quote: %w", in.FX, in.Lane.Currency(), home, shared.ErrCurrencyMismatch)
	}

	// What do we weigh? The real thing if it has been on our scale, else the
	// category's default — and the breakdown says which.
	parcel, estimated := in.Listing.Parcel, false
	if !in.Listing.Measured || parcel.IsZero() {
		if in.Profile.Estimate.IsZero() {
			return Breakdown{}, fmt.Errorf("product %s: %w", in.Listing.Product, ErrNothingToWeigh)
		}
		parcel, estimated = in.Profile.Estimate, true
	}
	class := in.Profile.Class
	if !class.IsValid() {
		class = ClassStandard
	}

	chargeable := in.Lane.ChargeableWeight(parcel)
	salesTax := item.Mul(in.Policy.SalesTax())
	freight := in.Lane.Freight(class, chargeable)
	surcharge := in.Lane.Surcharge(in.Profile.Restrictions)
	duty := in.Lane.Duty(item)
	subtotal, err := shared.Sum(item, salesTax, freight, surcharge, duty)
	if err != nil {
		return Breakdown{}, err
	}
	subtotalHome, err := subtotal.Convert(in.FX)
	if err != nil {
		return Breakdown{}, err
	}
	fee, err := in.Policy.Margin().Fee(subtotalHome)
	if err != nil {
		return Breakdown{}, err
	}
	total, err := subtotalHome.Add(fee)
	if err != nil {
		return Breakdown{}, err
	}
	return Breakdown{
		Class: class, Chargeable: chargeable, Estimated: estimated,
		ItemPrice: item, SalesTax: salesTax, Freight: freight, Surcharge: surcharge, Duty: duty, SubtotalUSD: subtotal,
		FX: in.FX, SubtotalVND: subtotalHome, ServiceFee: fee, TotalVND: total, Deposit: total.Mul(in.Policy.Deposit()),
	}, nil
}
