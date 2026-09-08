package orderingapp

import (
	"context"

	"github.com/duongsy/portage/internal/domain/ordering"
)

// The state changes other contexts REPORT rather than the customer requests:
// procurement bought (or could not), logistics shipped, the courier delivered.
// Each is a thin handler; procurement and logistics call them through event
// subscriptions (wire.Subscribe), an operator can call Deliver through HTTP.

// ConfirmPurchaseHandler — procurement bought the goods: the point of no return.
type ConfirmPurchaseHandler struct {
	deps Deps
}

func NewConfirmPurchaseHandler(d Deps) *ConfirmPurchaseHandler {
	d.requireMutation("ConfirmPurchaseHandler")
	return &ConfirmPurchaseHandler{deps: d}
}

func (h *ConfirmPurchaseHandler) Handle(ctx context.Context, id ordering.OrderID) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, id, func(o *ordering.CustomerOrder) error {
		return o.ConfirmPurchase(now)
	})
}

// FailPurchase — procurement could not buy; the order waits for a decision.
type FailPurchase struct {
	Order  ordering.OrderID
	Reason string
}

type FailPurchaseHandler struct {
	deps Deps
}

func NewFailPurchaseHandler(d Deps) *FailPurchaseHandler {
	d.requireMutation("FailPurchaseHandler")
	return &FailPurchaseHandler{deps: d}
}

func (h *FailPurchaseHandler) Handle(ctx context.Context, cmd FailPurchase) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, cmd.Order, func(o *ordering.CustomerOrder) error {
		return o.FailPurchase(cmd.Reason, now)
	})
}

// ShipOrderHandler — logistics: the parcel left the US warehouse.
type ShipOrderHandler struct {
	deps Deps
}

func NewShipOrderHandler(d Deps) *ShipOrderHandler {
	d.requireMutation("ShipOrderHandler")
	return &ShipOrderHandler{deps: d}
}

func (h *ShipOrderHandler) Handle(ctx context.Context, id ordering.OrderID) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, id, func(o *ordering.CustomerOrder) error {
		return o.Ship(now)
	})
}

// DeliverOrderHandler — the parcel reached the customer.
type DeliverOrderHandler struct {
	deps Deps
}

func NewDeliverOrderHandler(d Deps) *DeliverOrderHandler {
	d.requireMutation("DeliverOrderHandler")
	return &DeliverOrderHandler{deps: d}
}

func (h *DeliverOrderHandler) Handle(ctx context.Context, id ordering.OrderID) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, id, func(o *ordering.CustomerOrder) error {
		return o.Deliver(now)
	})
}
