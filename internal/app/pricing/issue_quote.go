package pricingapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// IssueQuote: "what would this product cost me, door to door, on this lane?"
type IssueQuote struct {
	Product shared.ID
	Lane    string // lane code as typed
}

// IssueQuoteHandler gathers the inputs the domain needs — what pricing knows
// about the product and its category, the lane, TODAY's rate — and lets
// pricing.IssueQuote freeze them into a quote.
type IssueQuoteHandler struct {
	deps Deps
}

func NewIssueQuoteHandler(d Deps) *IssueQuoteHandler {
	mustHave("IssueQuoteHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Lanes": d.Lanes, "Quotes": d.Quotes,
		"Listings": d.Listings, "Profiles": d.Profiles, "Rates": d.Rates,
	})
	if d.Policy.IsZero() {
		panic("pricingapp: IssueQuoteHandler wired without a QuotePolicy")
	}
	return &IssueQuoteHandler{deps: d}
}

func (h *IssueQuoteHandler) Handle(ctx context.Context, cmd IssueQuote) (pricing.QuoteID, error) {
	now := h.deps.Clock.Now()
	laneCode, err := pricing.ParseLaneCode(cmd.Lane)
	if err != nil {
		return pricing.QuoteID{}, err
	}

	var id pricing.QuoteID
	err = h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		listing, err := h.deps.Listings.ByProduct(ctx, cmd.Product)
		if err != nil {
			return err
		}
		lane, err := h.deps.Lanes.ByCode(ctx, laneCode)
		if err != nil {
			return err
		}
		// A category profile is optional for a measured product: Calculate only
		// needs it when it has to guess the parcel.
		profile, err := h.deps.Profiles.ByCode(ctx, listing.Category)
		if err != nil && !isNotFound(err) {
			return err
		}
		home := h.deps.Policy.Margin().Floor().Currency()
		fx, err := h.deps.Rates.Current(ctx, lane.Currency(), home)
		if err != nil {
			return err
		}

		q, err := pricing.IssueQuote(pricing.QuoteInputs{
			Listing: listing, Profile: profile, Lane: lane, FX: fx, Policy: h.deps.Policy,
		}, now)
		if err != nil {
			return err
		}
		if err := h.deps.Quotes.Save(ctx, q); err != nil {
			return fmt.Errorf("save quote: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, q.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		id = q.ID()
		return nil
	})
	return id, err
}
