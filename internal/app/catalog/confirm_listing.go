package catalogapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// ConfirmListing is an operator vouching for what a product IS — the name,
// the shop, the category — after a customer pasted it. Publish demands this.
type ConfirmListing struct {
	Product  catalog.ProductID
	Operator shared.OperatorID // who is vouching; required
}

// ConfirmListingHandler builds the verified provenance (operator + now) and
// hands it to the aggregate. A missing operator is refused by the domain's
// own NewProvenance (ErrOperatorRequired) — the handler adds no rule.
type ConfirmListingHandler struct {
	deps Deps
}

func NewConfirmListingHandler(d Deps) *ConfirmListingHandler {
	mustHave("ConfirmListingHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Products": d.Products, "Outbox": d.Outbox,
	})
	return &ConfirmListingHandler{deps: d}
}

func (h *ConfirmListingHandler) Handle(ctx context.Context, cmd ConfirmListing) error {
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
		if err := p.ConfirmListing(prov, now); err != nil {
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
