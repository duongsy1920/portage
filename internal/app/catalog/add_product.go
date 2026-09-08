package catalogapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// AddProduct is the command behind "record what a customer (or operator)
// pointed at". Compare catalog.ProductDetails: the command carries WHO and
// HOW the data arrived; the use case turns that plus the clock into the
// Provenance the domain wants. That translation is the reason this layer
// exists.
//
// [PHP] Command của Messenger — DTO thuần, không logic. Khác ProductDetails
// [PHP] của domain ở đúng hai field: SourcedBy và Operator, là dữ liệu của
// [PHP] REQUEST (auth, nguồn), không phải của sản phẩm.
type AddProduct struct {
	Name      string
	Merchant  catalog.MerchantID
	Category  catalog.CategoryCode
	Source    catalog.SourceURL
	Price     shared.Money
	SourcedBy catalog.SourcingMode // how the data reached us: customer or operator
	Operator  shared.OperatorID    // who, when SourcedBy is operator (from auth)
}

// AddProductHandler records a draft product, enforcing the two rules that need
// more than one aggregate, and runs duplicate detection (option C).
type AddProductHandler struct {
	deps Deps
}

func NewAddProductHandler(d Deps) *AddProductHandler {
	mustHave("AddProductHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Merchants": d.Merchants,
		"Categories": d.Categories, "Products": d.Products, "Outbox": d.Outbox,
	})
	return &AddProductHandler{deps: d}
}

// Handle, in order:
//
//  1. the merchant must exist and be active — nothing new is listed for a
//     shop we cannot buy from (ErrMerchantInactive, a two-aggregate rule);
//  2. the price must be in the merchant's currency (ErrPriceCurrency — the
//     Product cannot know the merchant's currency, the use case can);
//  3. the category must exist;
//  4. clock + caller → Provenance; the domain builds the draft;
//  5. duplicate detection: any product already recorded for this page flags
//     the new one, with the detector's reason (CATALOG.md §7, option C);
//  6. save, then pull events into the outbox.
func (h *AddProductHandler) Handle(ctx context.Context, cmd AddProduct) (catalog.ProductID, error) {
	now := h.deps.Clock.Now()

	var id catalog.ProductID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		m, err := h.deps.Merchants.ByID(ctx, cmd.Merchant)
		if err != nil {
			return err
		}
		if !m.IsActive() {
			return fmt.Errorf("merchant %q: %w", m.Name(), ErrMerchantInactive)
		}
		if cmd.Price.IsValid() && cmd.Price.Currency() != m.Currency() {
			return fmt.Errorf("price %s for merchant %q selling in %s: %w",
				cmd.Price, m.Name(), m.Currency(), ErrPriceCurrency)
		}
		if _, err := h.deps.Categories.ByCode(ctx, cmd.Category); err != nil {
			return err
		}

		prov, err := catalog.NewProvenance(cmd.SourcedBy, now, cmd.Operator)
		if err != nil {
			return err
		}
		p, err := catalog.AddProduct(catalog.ProductDetails{
			Name:              cmd.Name,
			Merchant:          cmd.Merchant,
			Category:          cmd.Category,
			Source:            cmd.Source,
			Price:             cmd.Price,
			ListingProvenance: prov,
			PriceProvenance:   prov,
		}, now)
		if err != nil {
			return err
		}

		dups, err := h.deps.Products.BySource(ctx, cmd.Source)
		if err != nil {
			return fmt.Errorf("look up duplicates: %w", err)
		}
		if len(dups) > 0 {
			if err := p.FlagDuplicateOf(dups[0].ID(), "same page url", now); err != nil {
				return err
			}
		}

		if err := h.deps.Products.Save(ctx, p); err != nil {
			return fmt.Errorf("save product: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, p.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		id = p.ID()
		return nil
	})
	return id, err
}
