package catalogapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// MeasureProduct is the warehouse moment: the real item on a real scale
// (CATALOG.md §3). Measurements are operator work, so the operator is
// required — this is the "verified parcel" Publish asks for, and the
// number pricing will quote on from then on (catalog.product_measured).
type MeasureProduct struct {
	Product  catalog.ProductID
	Spec     shared.ParcelSpec
	Operator shared.OperatorID
}

type MeasureProductHandler struct {
	deps Deps
}

func NewMeasureProductHandler(d Deps) *MeasureProductHandler {
	mustHave("MeasureProductHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Products": d.Products, "Outbox": d.Outbox,
	})
	return &MeasureProductHandler{deps: d}
}

func (h *MeasureProductHandler) Handle(ctx context.Context, cmd MeasureProduct) error {
	now := h.deps.Clock.Now()
	prov, err := catalog.NewProvenance(catalog.SourcedByOperator, now, cmd.Operator)
	if err != nil {
		return err
	}
	return h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		p, err := h.deps.Products.ByID(ctx, cmd.Product)
		if err != nil {
			return err
		}
		if err := p.Measure(cmd.Spec, prov, now); err != nil {
			return err
		}
		if err := h.deps.Products.Save(ctx, p); err != nil {
			return fmt.Errorf("save product: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, p.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}
