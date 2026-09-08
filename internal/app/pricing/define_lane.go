package pricingapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// DefineLaneHandler saves a lane and ANNOUNCES it, the way DefineCategory does
// in catalog: logistics keeps a copy of the divisor and the step to split
// freight, so saving straight to the repository would tell nobody. The dev
// seed runs this on every start-up; consumers are idempotent.
type DefineLaneHandler struct {
	deps Deps
}

func NewDefineLaneHandler(d Deps) *DefineLaneHandler {
	mustHave("DefineLaneHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Lanes": d.Lanes})
	return &DefineLaneHandler{deps: d}
}

func (h *DefineLaneHandler) Handle(ctx context.Context, d pricing.LaneDetails) error {
	now := h.deps.Clock.Now()
	lane, err := pricing.NewShippingLane(d)
	if err != nil {
		return err
	}
	return h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		if err := h.deps.Lanes.Save(ctx, lane); err != nil {
			return fmt.Errorf("save lane: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, []shared.Event{lane.Defined(now)}); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}
