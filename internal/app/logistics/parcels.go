package logisticsapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/shared"
)

// ExpectParcelHandler reacts to procurement.purchase_confirmed: a box with
// this reference is now on its way to the warehouse. Idempotent by order.
type ExpectParcelHandler struct {
	deps Deps
}

func NewExpectParcelHandler(d Deps) *ExpectParcelHandler {
	mustHave("ExpectParcelHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Parcels": d.Parcels})
	return &ExpectParcelHandler{deps: d}
}

func (h *ExpectParcelHandler) Handle(ctx context.Context, d logistics.ParcelDetails) (logistics.ParcelID, error) {
	now := h.deps.Clock.Now()
	var id logistics.ParcelID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		if existing, err := h.deps.Parcels.ByOrder(ctx, d.Order); err == nil {
			id = existing.ID()
			return nil
		} else if !errors.Is(err, logistics.ErrParcelNotFound) {
			return err
		}
		p, err := logistics.ExpectParcel(d, now)
		if err != nil {
			return err
		}
		if err := h.deps.Parcels.Save(ctx, p); err != nil {
			return fmt.Errorf("save parcel: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, p.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		id = p.ID()
		return nil
	})
	return id, err
}

func (h *ExpectParcelHandler) OnPurchaseConfirmed(ctx context.Context, m contracts.PurchaseConfirmedV1) error {
	order, err := shared.ParseID(m.Order)
	if err != nil {
		return fmt.Errorf("purchase_confirmed %s: %w", m.ID, err)
	}
	_, err = h.Handle(ctx, logistics.ParcelDetails{Order: order, Reference: m.Reference})
	return err
}

// ReceiveParcel: the box is on our scale.
type ReceiveParcel struct {
	Parcel   logistics.ParcelID
	Actual   shared.ParcelSpec
	Operator shared.OperatorID
}

type ReceiveParcelHandler struct {
	deps Deps
}

func NewReceiveParcelHandler(d Deps) *ReceiveParcelHandler {
	mustHave("ReceiveParcelHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Parcels": d.Parcels})
	return &ReceiveParcelHandler{deps: d}
}

func (h *ReceiveParcelHandler) Handle(ctx context.Context, cmd ReceiveParcel) error {
	now := h.deps.Clock.Now()
	return h.deps.mutateParcel(ctx, cmd.Parcel, func(p *logistics.Parcel) error {
		return p.Receive(cmd.Actual, cmd.Operator, now)
	})
}

// Projector keeps logistics' copy of each lane's counting rule.
type Projector struct {
	deps Deps
}

func NewProjector(d Deps) *Projector {
	mustHave("Projector", map[string]any{"UoW": d.UoW, "Lanes": d.Lanes})
	return &Projector{deps: d}
}

func (p *Projector) OnLaneDefined(ctx context.Context, m contracts.LaneDefinedV1) error {
	step, err := shared.NewWeight(m.StepG)
	if err != nil {
		return fmt.Errorf("lane_defined %s: %w", m.Code, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Lanes.Save(ctx, logistics.LaneRule{Code: m.Code, Divisor: m.Divisor, Step: step})
	})
}
