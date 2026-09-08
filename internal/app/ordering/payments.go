package orderingapp

import (
	"context"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Payment is "we received this much for this order". Today an operator
// confirms it by hand; a payment-gateway adapter will send the same command.
type Payment struct {
	Order  ordering.OrderID
	Amount shared.Money
}

// PayDepositHandler — the first half. The aggregate demands the exact amount.
type PayDepositHandler struct {
	deps Deps
}

func NewPayDepositHandler(d Deps) *PayDepositHandler {
	d.requireMutation("PayDepositHandler")
	return &PayDepositHandler{deps: d}
}

func (h *PayDepositHandler) Handle(ctx context.Context, cmd Payment) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, cmd.Order, func(o *ordering.CustomerOrder) error {
		return o.PayDeposit(cmd.Amount, now)
	})
}

// PayBalanceHandler — the second half, once the parcel is in transit.
type PayBalanceHandler struct {
	deps Deps
}

func NewPayBalanceHandler(d Deps) *PayBalanceHandler {
	d.requireMutation("PayBalanceHandler")
	return &PayBalanceHandler{deps: d}
}

func (h *PayBalanceHandler) Handle(ctx context.Context, cmd Payment) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, cmd.Order, func(o *ordering.CustomerOrder) error {
		return o.PayBalance(cmd.Amount, now)
	})
}
