package pricing

import (
	"context"
	"errors"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	ErrLaneNotFound    = errors.New("lane not found")
	ErrQuoteNotFound   = errors.New("quote not found")
	ErrListingNotFound = errors.New("listing not found")
	ErrProfileNotFound = errors.New("category profile not found")
	ErrNoExchangeRate  = errors.New("no exchange rate")
)

// The PORTS of pricing (DDD.md §16): declared here, implemented in
// internal/adapter/{memory,postgres}. Save on the value-object stores (lanes,
// listings, profiles, rates) replaces whole.

type LaneRepository interface {
	ByCode(ctx context.Context, code LaneCode) (ShippingLane, error) // ErrLaneNotFound
	All(ctx context.Context) ([]ShippingLane, error)
	Save(ctx context.Context, lane ShippingLane) error
}

type QuoteRepository interface {
	ByID(ctx context.Context, id QuoteID) (*Quote, error) // ErrQuoteNotFound
	Save(ctx context.Context, q *Quote) error

	// IssuedBefore is the expiry sweep's one query: quotes still waiting for
	// an answer whose deadline has passed. limit keeps a pass bounded — a
	// backlog is worked through over several passes rather than in one
	// transaction the size of the table.
	IssuedBefore(ctx context.Context, t time.Time, limit int) ([]*Quote, error)
}

// ListingRepository stores pricing's projection of products. Written only by
// the projector that consumes catalog's events.
type ListingRepository interface {
	ByProduct(ctx context.Context, product shared.ID) (Listing, error) // ErrListingNotFound
	Save(ctx context.Context, l Listing) error
}

type CategoryProfileRepository interface {
	ByCode(ctx context.Context, code string) (CategoryProfile, error) // ErrProfileNotFound
	Save(ctx context.Context, p CategoryProfile) error
}

// ExchangeRates answers "what is the rate right now" — the one input a quote
// takes from the world and then freezes. Set is how an operator (or, later, a
// feed adapter) publishes today's rate.
type ExchangeRates interface {
	Current(ctx context.Context, from, to shared.Currency) (shared.ExchangeRate, error) // ErrNoExchangeRate
	Set(ctx context.Context, rate shared.ExchangeRate) error
}
