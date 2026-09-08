package orderingapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Projector is ordering's ear on pricing: when a customer accepts a quote,
// ordering remembers the numbers so PlaceOrder can use them without asking
// pricing. Idempotent (upsert), like pricing's projector on catalog.
type Projector struct {
	deps Deps
}

func NewProjector(d Deps) *Projector {
	mustHave("Projector", map[string]any{"UoW": d.UoW, "Quotes": d.Quotes, "Variants": d.Variants})
	return &Projector{deps: d}
}

func (p *Projector) OnQuoteAccepted(ctx context.Context, m contracts.QuoteAcceptedV1) error {
	quote, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("quote_accepted: %w", err)
	}
	product, err := shared.ParseID(m.Product)
	if err != nil {
		return fmt.Errorf("quote_accepted %s: %w", m.ID, err)
	}
	total, err := moneyFrom(m.Total)
	if err != nil {
		return fmt.Errorf("quote_accepted %s: %w", m.ID, err)
	}
	deposit, err := moneyFrom(m.Deposit)
	if err != nil {
		return fmt.Errorf("quote_accepted %s: %w", m.ID, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Quotes.Save(ctx, ordering.AcceptedQuote{Quote: quote, Product: product, Total: total, Deposit: deposit})
	})
}

func moneyFrom(m contracts.MoneyV1) (shared.Money, error) {
	cur, err := shared.CurrencyFromCode(m.Currency)
	if err != nil {
		return shared.Money{}, err
	}
	return shared.NewMoney(m.Minor, cur), nil
}

// OnVariantAdded remembers that a variant exists and whose product it is, so
// PlaceOrder can refuse an id the customer did not get from us.
//
// Idempotent and order-free: it is a dictionary, and a variant is announced
// long before anybody orders it.
func (p *Projector) OnVariantAdded(ctx context.Context, m contracts.VariantAddedV1) error {
	variant, err := shared.ParseID(m.Variant)
	if err != nil {
		return fmt.Errorf("variant_added: %w", err)
	}
	product, err := shared.ParseID(m.Product)
	if err != nil {
		return fmt.Errorf("variant_added %s: %w", m.Variant, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Variants.Save(ctx, ordering.Variant{Variant: variant, Product: product})
	})
}
