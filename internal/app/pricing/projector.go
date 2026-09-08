package pricingapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Projector is pricing's ear on catalog: it consumes the Published Language
// (contracts.*V1) and maintains pricing's own records — Listing per product,
// CategoryProfile per category. It is the first real subscriber in the system.
//
// Three properties, all tested:
//
//   - IDEMPOTENT: every write is an upsert, so a message delivered twice
//     (at-least-once, DDD.md §27) changes nothing the second time.
//   - ORDER-TOLERANT: a measurement or a price for a product pricing has not
//     heard of is ignored, not an error — product_published carries the full
//     state and will overwrite anyway.
//   - OWN VOCABULARY: the goods class comes from pricing's Classification, not
//     from the wire; catalog does not know what a "sensitive" class is.
//
// [PHP] Một MessageHandler nhận event Catalog và ghi bảng của Pricing —
// [PHP] "projection" trong CQRS. Idempotent = chạy lại không đổi kết quả.
type Projector struct {
	deps Deps
}

func NewProjector(d Deps) *Projector {
	mustHave("Projector", map[string]any{"UoW": d.UoW, "Listings": d.Listings, "Profiles": d.Profiles})
	return &Projector{deps: d}
}

func (p *Projector) OnProductPublished(ctx context.Context, m contracts.ProductPublishedV1) error {
	product, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("product_published: %w", err)
	}
	price, err := moneyFrom(m.Price)
	if err != nil {
		return fmt.Errorf("product_published %s: %w", m.ID, err)
	}
	parcel, err := parcelFrom(m.Parcel)
	if err != nil {
		return fmt.Errorf("product_published %s: %w", m.ID, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Listings.Save(ctx, pricing.Listing{
			Product: product, Name: m.Name, Category: m.Category, Price: price,
			Parcel: parcel, Measured: true, Active: true, // publishing requires a verified parcel
		})
	})
}

func (p *Projector) OnProductMeasured(ctx context.Context, m contracts.ProductMeasuredV1) error {
	parcel, err := parcelFrom(m.Parcel)
	if err != nil {
		return fmt.Errorf("product_measured %s: %w", m.ID, err)
	}
	return p.update(ctx, m.ID, func(l *pricing.Listing) {
		l.Parcel, l.Measured = parcel, true
	})
}

func (p *Projector) OnProductRepriced(ctx context.Context, m contracts.ProductRepricedV1) error {
	price, err := moneyFrom(m.To)
	if err != nil {
		return fmt.Errorf("product_repriced %s: %w", m.ID, err)
	}
	return p.update(ctx, m.ID, func(l *pricing.Listing) {
		l.Price = price
	})
}

func (p *Projector) OnProductRetired(ctx context.Context, m contracts.ProductRetiredV1) error {
	return p.update(ctx, m.ID, func(l *pricing.Listing) {
		l.Active = false
	})
}

func (p *Projector) OnCategoryDefined(ctx context.Context, m contracts.CategoryDefinedV1) error {
	estimate, err := parcelFrom(m.Estimate)
	if err != nil {
		return fmt.Errorf("category_defined %s: %w", m.Code, err)
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Profiles.Save(ctx, pricing.CategoryProfile{
			Code: m.Code, Class: p.deps.Classification.ClassOf(m.Code), Estimate: estimate, Restrictions: m.Restrictions,
		})
	})
}

// update applies fn to an existing listing; an unknown product is not an
// error (see the package comment on order tolerance).
func (p *Projector) update(ctx context.Context, id string, fn func(*pricing.Listing)) error {
	product, err := shared.ParseID(id)
	if err != nil {
		return err
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		l, err := p.deps.Listings.ByProduct(ctx, product)
		if errors.Is(err, pricing.ErrListingNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		fn(&l)
		return p.deps.Listings.Save(ctx, l)
	})
}

// moneyFrom / parcelFrom turn wire primitives back into value objects through
// the domain's validating constructors — corrupted messages fail here.
func moneyFrom(m contracts.MoneyV1) (shared.Money, error) {
	cur, err := shared.CurrencyFromCode(m.Currency)
	if err != nil {
		return shared.Money{}, err
	}
	return shared.NewMoney(m.Minor, cur), nil
}

func parcelFrom(p contracts.ParcelV1) (shared.ParcelSpec, error) {
	w, err := shared.NewWeight(p.WeightG)
	if err != nil {
		return shared.ParcelSpec{}, err
	}
	if p.LengthMM < 0 || p.WidthMM < 0 || p.HeightMM < 0 {
		return shared.ParcelSpec{}, errors.New("negative dimension")
	}
	return shared.NewParcelSpec(w, shared.NewDimensionsMM(p.LengthMM, p.WidthMM, p.HeightMM))
}

func isNotFound(err error) bool {
	return errors.Is(err, pricing.ErrProfileNotFound)
}
