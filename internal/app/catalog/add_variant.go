package catalogapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
)

// AddVariant: "this product also comes in US 9 / black". Sizes and colours
// are the shop's words, kept verbatim — the domain only refuses duplicates.
type AddVariant struct {
	Product     catalog.ProductID
	Size        string
	Color       string
	MerchantRef string
}

// AddVariantHandler is load–act–save with no event: nothing downstream
// reacts to a variant on its own (yet). The outbox call stays anyway, so
// every handler reads alike and a future event needs no new plumbing.
//
// [PHP] Cùng khuôn với PublishProductHandler: find → gọi method → flush.
type AddVariantHandler struct {
	deps Deps
}

func NewAddVariantHandler(d Deps) *AddVariantHandler {
	mustHave("AddVariantHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Products": d.Products, "Outbox": d.Outbox,
	})
	return &AddVariantHandler{deps: d}
}

func (h *AddVariantHandler) Handle(ctx context.Context, cmd AddVariant) (catalog.VariantID, error) {
	now := h.deps.Clock.Now()
	var id catalog.VariantID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		p, err := h.deps.Products.ByID(ctx, cmd.Product)
		if err != nil {
			return err
		}
		id, err = p.AddVariant(catalog.VariantDetails{Size: cmd.Size, Color: cmd.Color, MerchantRef: cmd.MerchantRef}, now)
		if err != nil {
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
	return id, err
}
