package procurementapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Projector is procurement's ear on catalog: which shop a product belongs to
// and what that shop bills in. Two upserts; idempotent; order-tolerant in the
// only way that matters here (a product whose shop is unknown is still saved —
// the shop row arrives on its own event).
type Projector struct {
	deps Deps
}

func NewProjector(d Deps) *Projector {
	mustHave("Projector", map[string]any{"UoW": d.UoW, "Shops": d.Shops, "Items": d.Items, "Variants": d.Variants})
	return &Projector{deps: d}
}

func (p *Projector) OnMerchantRegistered(ctx context.Context, m contracts.MerchantRegisteredV1) error {
	id, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("merchant_registered: %w", err)
	}
	cur, err := shared.CurrencyFromCode(m.Currency)
	if err != nil {
		return fmt.Errorf("merchant_registered %s: %w", m.ID, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Shops.Save(ctx, procurement.Shop{Merchant: id, Name: m.Name, Site: m.Site, Currency: cur})
	})
}

func (p *Projector) OnProductPublished(ctx context.Context, m contracts.ProductPublishedV1) error {
	product, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("product_published: %w", err)
	}
	merchant, err := shared.ParseID(m.Merchant)
	if err != nil {
		return fmt.Errorf("product_published %s: %w", m.ID, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Items.Save(ctx, procurement.Item{
			Product: product, Merchant: merchant, Name: m.Name, Source: m.Source,
		})
	})
}

// OnVariantAdded records which size a variant id means.
//
// It does NOT require the product to be known first. A variant is announced
// when the operator types it, which is BEFORE publish, so this row nearly
// always arrives before the item does — the projection is a dictionary, not a
// child of anything, and order does not matter to it.
func (p *Projector) OnVariantAdded(ctx context.Context, m contracts.VariantAddedV1) error {
	variant, err := shared.ParseID(m.Variant)
	if err != nil {
		return fmt.Errorf("variant_added: %w", err)
	}
	product, err := shared.ParseID(m.Product)
	if err != nil {
		return fmt.Errorf("variant_added %s: %w", m.Variant, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Variants.Save(ctx, procurement.Variant{
			Variant: variant, Product: product,
			Size: m.Size, Color: m.Color, MerchantRef: m.MerchantRef,
		})
	})
}
