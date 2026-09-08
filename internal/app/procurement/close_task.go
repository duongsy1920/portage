package procurementapp

import (
	"context"

	"github.com/duongsy/portage/internal/domain/procurement"
)

// ConfirmTask: the buyer bought it — here is the shop's reference and what we paid.
type ConfirmTask struct {
	Task    procurement.TaskID
	Receipt procurement.PurchaseReceipt
}

type ConfirmTaskHandler struct {
	deps Deps
}

func NewConfirmTaskHandler(d Deps) *ConfirmTaskHandler {
	mustHave("ConfirmTaskHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Tasks": d.Tasks})
	return &ConfirmTaskHandler{deps: d}
}

func (h *ConfirmTaskHandler) Handle(ctx context.Context, cmd ConfirmTask) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, cmd.Task, func(t *procurement.PurchaseTask) error {
		return t.Confirm(cmd.Receipt, now)
	})
}

// FailTask: the buyer could not buy — sold out, price jumped, shop refused.
type FailTask struct {
	Task   procurement.TaskID
	Reason string
}

type FailTaskHandler struct {
	deps Deps
}

func NewFailTaskHandler(d Deps) *FailTaskHandler {
	mustHave("FailTaskHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Tasks": d.Tasks})
	return &FailTaskHandler{deps: d}
}

func (h *FailTaskHandler) Handle(ctx context.Context, cmd FailTask) error {
	now := h.deps.Clock.Now()
	return h.deps.mutate(ctx, cmd.Task, func(t *procurement.PurchaseTask) error {
		return t.Fail(cmd.Reason, now)
	})
}
