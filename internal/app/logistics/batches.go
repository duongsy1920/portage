package logisticsapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/shared"
)

// OpenBatchHandler: a new carton for a lane.
type OpenBatchHandler struct {
	deps Deps
}

func NewOpenBatchHandler(d Deps) *OpenBatchHandler {
	mustHave("OpenBatchHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Batches": d.Batches, "Lanes": d.Lanes})
	return &OpenBatchHandler{deps: d}
}

func (h *OpenBatchHandler) Handle(ctx context.Context, lane string) (logistics.BatchID, error) {
	now := h.deps.Clock.Now()
	var id logistics.BatchID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		if _, err := h.deps.Lanes.ByCode(ctx, lane); err != nil {
			return err // an unknown lane cannot be shipped later; refuse now
		}
		b, err := logistics.OpenBatch(lane, now)
		if err != nil {
			return err
		}
		if err := h.deps.Batches.Save(ctx, b); err != nil {
			return fmt.Errorf("save batch: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, b.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		id = b.ID()
		return nil
	})
	return id, err
}

// AddParcelToBatch touches TWO aggregates — the batch gains an item, the
// parcel learns its batch — in one transaction. That relaxes "one aggregate
// per transaction" on purpose: the parcel's status is bookkeeping derived
// from the batch, not an invariant another party depends on, and an event
// round-trip to keep them in sync would buy nothing but latency.
type AddParcelToBatch struct {
	Batch  logistics.BatchID
	Parcel logistics.ParcelID
}

type AddParcelHandler struct {
	deps Deps
}

func NewAddParcelHandler(d Deps) *AddParcelHandler {
	mustHave("AddParcelHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Batches": d.Batches, "Parcels": d.Parcels})
	return &AddParcelHandler{deps: d}
}

func (h *AddParcelHandler) Handle(ctx context.Context, cmd AddParcelToBatch) error {
	return h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		b, err := h.deps.Batches.ByID(ctx, cmd.Batch)
		if err != nil {
			return err
		}
		p, err := h.deps.Parcels.ByID(ctx, cmd.Parcel)
		if err != nil {
			return err
		}
		if err := b.AddParcel(p.Item()); err != nil {
			return err
		}
		if err := p.AssignToBatch(b.ID()); err != nil {
			return err
		}
		if err := h.deps.Batches.Save(ctx, b); err != nil {
			return fmt.Errorf("save batch: %w", err)
		}
		if err := h.deps.Parcels.Save(ctx, p); err != nil {
			return fmt.Errorf("save parcel: %w", err)
		}
		return nil // no events: the batch speaks when it ships
	})
}

type CloseBatchHandler struct {
	deps Deps
}

func NewCloseBatchHandler(d Deps) *CloseBatchHandler {
	mustHave("CloseBatchHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Batches": d.Batches})
	return &CloseBatchHandler{deps: d}
}

func (h *CloseBatchHandler) Handle(ctx context.Context, id logistics.BatchID) error {
	now := h.deps.Clock.Now()
	return h.deps.mutateBatch(ctx, id, func(b *logistics.ConsolidationBatch) error {
		return b.Close(now)
	})
}

// ShipBatch: the carrier's invoice for the whole carton.
type ShipBatch struct {
	Batch   logistics.BatchID
	Freight shared.Money
}

// ShipBatchHandler picks the allocator from the lane's rule, lets the batch
// split the invoice, then marks every parcel shipped — same transaction, same
// reasoning as AddParcelHandler.
type ShipBatchHandler struct {
	deps Deps
}

func NewShipBatchHandler(d Deps) *ShipBatchHandler {
	mustHave("ShipBatchHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Batches": d.Batches, "Parcels": d.Parcels, "Lanes": d.Lanes})
	return &ShipBatchHandler{deps: d}
}

func (h *ShipBatchHandler) Handle(ctx context.Context, cmd ShipBatch) ([]logistics.Allocation, error) {
	now := h.deps.Clock.Now()
	var out []logistics.Allocation
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		b, err := h.deps.Batches.ByID(ctx, cmd.Batch)
		if err != nil {
			return err
		}
		rule, err := h.deps.Lanes.ByCode(ctx, b.Lane())
		if err != nil {
			return err
		}
		allocations, err := b.Ship(cmd.Freight, logistics.ByChargeableWeight{Divisor: rule.Divisor, Step: rule.Step}, now)
		if err != nil {
			return err
		}
		for _, a := range allocations {
			p, err := h.deps.Parcels.ByID(ctx, logistics.ParcelID{ID: a.Parcel})
			if err != nil {
				return err
			}
			if err := p.MarkShipped(); err != nil {
				return err
			}
			if err := h.deps.Parcels.Save(ctx, p); err != nil {
				return fmt.Errorf("save parcel: %w", err)
			}
		}
		if err := h.deps.Batches.Save(ctx, b); err != nil {
			return fmt.Errorf("save batch: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, b.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		out = allocations
		return nil
	})
	return out, err
}
