package orderingapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// PlaceOrder: the customer commits to a quote they accepted, for one variant.
type PlaceOrder struct {
	Quote    shared.ID
	Variant  shared.ID
	Customer shared.ID
}

// PlaceOrderHandler enforces the two rules that span aggregates: the quote
// must have been accepted (ordering's projection knows), and one quote makes
// ONE order (a double click is not two commitments).
type PlaceOrderHandler struct {
	deps Deps
}

func NewPlaceOrderHandler(d Deps) *PlaceOrderHandler {
	mustHave("PlaceOrderHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox,
		"Orders": d.Orders, "Quotes": d.Quotes, "Variants": d.Variants})
	return &PlaceOrderHandler{deps: d}
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrder) (ordering.OrderID, error) {
	now := h.deps.Clock.Now()
	var id ordering.OrderID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		q, err := h.deps.Quotes.ByID(ctx, cmd.Quote)
		if errors.Is(err, ordering.ErrAcceptedQuoteNotFound) {
			return fmt.Errorf("quote %s: %w", cmd.Quote, ordering.ErrQuoteNotAccepted) // never issued, not accepted yet, or not relayed yet
		}
		if err != nil {
			return err
		}
		if existing, err := h.deps.Orders.ByQuote(ctx, cmd.Quote); err == nil {
			return fmt.Errorf("quote %s already ordered as %s: %w", cmd.Quote, existing.ID(), ordering.ErrQuoteAlreadyUsed)
		} else if !errors.Is(err, ordering.ErrOrderNotFound) {
			return err
		}
		// The variant is the ONE thing in this request the customer chose
		// freely, so it is the one thing to check. It must exist, and it must
		// be a form of the product the quote priced — otherwise the buyer is
		// sent to a shop to ask for a size that belongs to another shoe.
		v, err := h.deps.Variants.ByID(ctx, cmd.Variant)
		if errors.Is(err, ordering.ErrVariantNotFound) {
			return fmt.Errorf("variant %s: %w", cmd.Variant, ordering.ErrVariantUnknown)
		}
		if err != nil {
			return err
		}
		if v.Product != q.Product {
			return fmt.Errorf("variant %s is a form of product %s, quote %s prices %s: %w",
				cmd.Variant, v.Product, cmd.Quote, q.Product, ordering.ErrVariantNotForProduct)
		}
		o, err := ordering.PlaceOrder(ordering.OrderDetails{
			Quote: q.Quote, Product: q.Product, Variant: cmd.Variant, Customer: cmd.Customer, Total: q.Total, Deposit: q.Deposit,
		}, now)
		if err != nil {
			return err
		}
		if err := h.deps.Orders.Save(ctx, o); err != nil {
			return fmt.Errorf("save order: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, o.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		id = o.ID()
		return nil
	})
	return id, err
}
