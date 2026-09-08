package orderingapp

import (
	"context"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// CancelOrder: the customer (or we) call it off. The refund is the domain's
// decision (DDD.md §26) and is returned so the caller can pay it out.
type CancelOrder struct {
	Order  ordering.OrderID
	Reason string

	// OnBehalfOf is the customer this cancellation must belong to. The HTTP
	// adapter fills it from the BEARER TOKEN when a customer calls, and
	// leaves it zero when an operator does — staff cancel on anyone's behalf.
	//
	// Zero meaning "no ownership check" is a deliberate exception to
	// convention 9, and it is safe for one reason: the adapter can only ever
	// set this from a principal, never from a request body. A caller cannot
	// send OnBehalfOf at all, so a caller cannot send the zero that skips the
	// check. TestCancelOrder_customerCannotCancelSomebodyElses pins it.
	OnBehalfOf shared.ID
}

type CancelOrderHandler struct {
	deps Deps
}

func NewCancelOrderHandler(d Deps) *CancelOrderHandler {
	d.requireMutation("CancelOrderHandler")
	return &CancelOrderHandler{deps: d}
}

func (h *CancelOrderHandler) Handle(ctx context.Context, cmd CancelOrder) (ordering.Refund, error) {
	now := h.deps.Clock.Now()
	var refund ordering.Refund
	err := h.deps.mutate(ctx, cmd.Order, func(o *ordering.CustomerOrder) error {
		// Ownership is checked HERE, inside the transaction, on the order we
		// actually loaded — not in the adapter against a copy read earlier.
		if !cmd.OnBehalfOf.IsZero() && o.Customer() != cmd.OnBehalfOf {
			return ordering.ErrNotOwner
		}
		r, err := o.Cancel(cmd.Reason, now)
		refund = r
		return err
	})
	return refund, err
}
