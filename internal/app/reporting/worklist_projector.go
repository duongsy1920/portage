package reportingapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The worklist half of the projector. Same two properties as the summary half,
// bought the same way: every handler upserts, and any event may create the row.
//
// Order tolerance matters more here than it looks. The four events come from
// four separate operator actions, and the relay makes no promise about which
// row it delivers first. A handler that ignored a measurement for a product it
// had not heard of would lose that fact forever, and the screen would ask a
// person to weigh a box they already weighed.

// OnProductAdded is where a row normally starts: a link was pasted, nothing
// has been done to it yet.
func (p *Projector) OnProductAdded(ctx context.Context, m contracts.ProductAddedV1) error {
	product, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("product_added: %w", err)
	}
	merchant, err := shared.ParseID(m.Merchant)
	if err != nil {
		return fmt.Errorf("product_added %s: %w", m.ID, err)
	}
	price, err := moneyFrom(m.Price)
	if err != nil {
		return fmt.Errorf("product_added %s: %w", m.ID, err)
	}
	// RequestedBy is optional: an operator adds products on spec. An empty
	// string stays the zero id rather than becoming a parse error.
	var requester shared.ID
	if m.RequestedBy != "" {
		requester, err = shared.ParseID(m.RequestedBy)
		if err != nil {
			return fmt.Errorf("product_added %s: requested_by: %w", m.ID, err)
		}
	}
	return p.worklist(ctx, product, m.At, func(w *WorklistItem) {
		w.Merchant, w.Category, w.Name = merchant, m.Category, m.Name
		w.Source, w.Price = m.Source, price
		w.SourcedBy, w.RequestedBy = m.SourcedBy, requester
		if w.AddedAt.IsZero() {
			w.AddedAt = m.At
		}
	})
}

// OnVariantAdded appends a purchasable size, with the id the customer will
// have to send back when ordering and the label a person reads.
//
// Appending is idempotent by ID, not by position: the relay may deliver the
// same variant twice, and a list that grew a duplicate would show the customer
// the same size in a dropdown twice.
func (p *Projector) OnVariantAdded(ctx context.Context, m contracts.VariantAddedV1) error {
	product, err := shared.ParseID(m.Product)
	if err != nil {
		return fmt.Errorf("variant_added: %w", err)
	}
	variant, err := shared.ParseID(m.Variant)
	if err != nil {
		return fmt.Errorf("variant_added %s: %w", m.Variant, err)
	}
	label := m.Size
	if m.Color != "" {
		if label != "" {
			label += " · "
		}
		label += m.Color
	}
	return p.worklist(ctx, product, m.At, func(w *WorklistItem) {
		for i, have := range w.Variants {
			if have.ID == variant {
				w.Variants[i].Label = label // a re-delivery, not a new size
				return
			}
		}
		w.Variants = append(w.Variants, WorklistVariant{ID: variant, Label: label})
	})
}

// OnListingConfirmed: an operator put their name to what this product is.
func (p *Projector) OnListingConfirmed(ctx context.Context, m contracts.ListingConfirmedV1) error {
	product, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("listing_confirmed: %w", err)
	}
	return p.worklist(ctx, product, m.At, func(w *WorklistItem) {
		w.ListingConfirmed = true
	})
}

// OnProductMeasured: somebody put the box on a scale.
func (p *Projector) OnProductMeasured(ctx context.Context, m contracts.ProductMeasuredV1) error {
	product, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("product_measured: %w", err)
	}
	return p.worklist(ctx, product, m.At, func(w *WorklistItem) {
		w.Measured = true
	})
}

// worklist is the upsert every handler above shares: read the row or start a
// skeleton, apply the change, write it back. UpdatedAt never goes backwards,
// so a re-delivered old event cannot make a row look fresher than it is.
func (p *Projector) worklist(ctx context.Context, product shared.ID, at time.Time, mutate func(*WorklistItem)) error {
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		item, err := p.deps.Worklist.ByProduct(ctx, product)
		if errors.Is(err, ErrWorklistItemNotFound) {
			item = WorklistItem{Product: product, AddedAt: at}
		} else if err != nil {
			return err
		}
		mutate(&item)
		if at.After(item.UpdatedAt) {
			item.UpdatedAt = at
		}
		return p.deps.Worklist.Save(ctx, item)
	})
}
