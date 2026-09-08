package ordering

import (
	"context"
	"errors"

	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrAcceptedQuoteNotFound = errors.New("accepted quote not found")
	ErrVariantNotFound       = errors.New("variant not found")
)

// The PORTS of ordering (DDD.md §16).

type OrderRepository interface {
	ByID(ctx context.Context, id OrderID) (*CustomerOrder, error) // ErrOrderNotFound
	// ByQuote finds the order placed on a quote, if any — one quote, one order.
	ByQuote(ctx context.Context, quote shared.ID) (*CustomerOrder, error) // ErrOrderNotFound
	Save(ctx context.Context, o *CustomerOrder) error
}

// AcceptedQuoteRepository stores ordering's projection of accepted quotes.
// Written only by the projector that consumes pricing's events.
// VariantRepository stores ordering's projection of catalog's variants.
// Written only by the projector that consumes catalog.variant_added.
type VariantRepository interface {
	ByID(ctx context.Context, variant shared.ID) (Variant, error) // ErrVariantNotFound
	Save(ctx context.Context, v Variant) error
}

type AcceptedQuoteRepository interface {
	ByID(ctx context.Context, quote shared.ID) (AcceptedQuote, error) // ErrAcceptedQuoteNotFound
	Save(ctx context.Context, q AcceptedQuote) error
}
