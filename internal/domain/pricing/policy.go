package pricing

import (
	"fmt"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// MarginPolicy is what we add for ourselves: a percentage of the landed cost,
// but never less than a floor — the "lãi 500k/đơn" the business was designed
// around (SETUP.md §7). The floor is in the home currency (VND).
type MarginPolicy struct {
	percent shared.Rate
	floor   shared.Money
}

func NewMarginPolicy(percent shared.Rate, floor shared.Money) (MarginPolicy, error) {
	if percent.PPM() < 0 {
		return MarginPolicy{}, fmt.Errorf("margin %s: %w", percent, ErrInvalidPolicy)
	}
	if !floor.IsValid() || floor.IsNegative() {
		return MarginPolicy{}, fmt.Errorf("margin floor %s: %w", floor, ErrInvalidPolicy)
	}
	return MarginPolicy{percent: percent, floor: floor}, nil
}

// Fee is max(percent × subtotal, floor), in the subtotal's currency — which
// must be the floor's.
func (m MarginPolicy) Fee(subtotal shared.Money) (shared.Money, error) {
	if subtotal.Currency() != m.floor.Currency() {
		return shared.Money{}, fmt.Errorf("margin on %s with floor in %s: %w", subtotal.Currency(), m.floor.Currency(), shared.ErrCurrencyMismatch)
	}
	share := subtotal.Mul(m.percent)
	if share.Minor() < m.floor.Minor() {
		return m.floor, nil
	}
	return share, nil
}

func (m MarginPolicy) Percent() shared.Rate {
	return m.percent
}

func (m MarginPolicy) Floor() shared.Money {
	return m.floor
}

func (m MarginPolicy) IsZero() bool {
	return m == MarginPolicy{}
}

// QuotePolicyDetails is the operator's input for the business constants a
// quote uses (convention 10).
type QuotePolicyDetails struct {
	SalesTax shared.Rate   // the warehouse state's sales tax: Denver 8.81 %
	Margin   MarginPolicy  // our fee
	Deposit  shared.Rate   // what the customer pays up front: 50 %
	TTL      time.Duration // how long a quote is a promise: 48 h
}

// QuotePolicy is the set of business constants every quote is built with.
// Like the lane, it is snapshotted INTO each quote: change the policy tomorrow
// and yesterday's quotes keep their numbers.
type QuotePolicy struct {
	salesTax shared.Rate
	margin   MarginPolicy
	deposit  shared.Rate
	ttl      time.Duration
}

func NewQuotePolicy(d QuotePolicyDetails) (QuotePolicy, error) {
	switch {
	case d.SalesTax.PPM() < 0:
		return QuotePolicy{}, fmt.Errorf("sales tax %s: %w", d.SalesTax, ErrInvalidPolicy)
	case d.Margin.IsZero():
		return QuotePolicy{}, fmt.Errorf("no margin policy: %w", ErrInvalidPolicy)
	case d.Deposit.PPM() <= 0 || d.Deposit.PPM() > 1_000_000:
		return QuotePolicy{}, fmt.Errorf("deposit %s: want 0 < deposit ≤ 100%%: %w", d.Deposit, ErrInvalidPolicy)
	case d.TTL <= 0:
		return QuotePolicy{}, fmt.Errorf("ttl %s: %w", d.TTL, ErrInvalidPolicy)
	}
	return QuotePolicy{salesTax: d.SalesTax, margin: d.Margin, deposit: d.Deposit, ttl: d.TTL}, nil
}

func (p QuotePolicy) SalesTax() shared.Rate {
	return p.salesTax
}

func (p QuotePolicy) Margin() MarginPolicy {
	return p.margin
}

func (p QuotePolicy) Deposit() shared.Rate {
	return p.deposit
}

func (p QuotePolicy) TTL() time.Duration {
	return p.ttl
}

func (p QuotePolicy) IsZero() bool {
	return p == QuotePolicy{}
}
