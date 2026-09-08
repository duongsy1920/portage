// Package orderingapp holds the use cases of the ordering context: place an
// order on an accepted quote, take the two payments, cancel with the right
// refund, and react to what procurement and logistics report. The rules are
// CustomerOrder's; the handlers open the transaction and hand over events.
//
// [PHP] MessageHandlers của bundle Ordering. `mutate` bên dưới là cái
// [PHP] AbstractHandler/trait "find → gọi method → flush" bạn hay viết.
package orderingapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/app"
	"github.com/duongsy/portage/internal/domain/ordering"
)

// Deps is everything the ordering use cases are wired to (convention 10).
type Deps struct {
	Clock  app.Clock
	UoW    app.UnitOfWork
	Outbox app.Outbox
	Orders ordering.OrderRepository
	Quotes ordering.AcceptedQuoteRepository
	// Variants: does this variant id exist, and is it this product's? The
	// customer sends the id, so somebody has to know (PlaceOrderHandler).
	Variants ordering.VariantRepository
}

func mustHave(handler string, deps map[string]any) {
	app.MustHave("orderingapp: "+handler, deps)
}

// mutate is the load → act → save → outbox shape every state change shares
// (WALKTHROUGH.md §8), written once. fn is the ONE line that differs per use
// case: which method of the aggregate to call. A domain refusal rolls back
// and is returned untouched.
func (d Deps) mutate(ctx context.Context, id ordering.OrderID, fn func(o *ordering.CustomerOrder) error) error {
	return d.UoW.InTx(ctx, func(ctx context.Context) error {
		o, err := d.Orders.ByID(ctx, id)
		if err != nil {
			return err
		}
		if err := fn(o); err != nil {
			return err
		}
		if err := d.Orders.Save(ctx, o); err != nil {
			return fmt.Errorf("save order: %w", err)
		}
		if err := d.Outbox.Append(ctx, o.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}

func (d Deps) requireMutation(handler string) {
	mustHave(handler, map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Orders": d.Orders})
}
