package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The ordering ports, in memory.

type OrderRepo struct {
	mu      sync.Mutex
	byID    map[ordering.OrderID]*ordering.CustomerOrder
	byQuote map[shared.ID]ordering.OrderID
}

var _ ordering.OrderRepository = (*OrderRepo)(nil)

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{byID: map[ordering.OrderID]*ordering.CustomerOrder{}, byQuote: map[shared.ID]ordering.OrderID{}}
}

func (r *OrderRepo) ByID(ctx context.Context, id ordering.OrderID) (*ordering.CustomerOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("order %s: %w", id, ordering.ErrOrderNotFound)
	}
	return o, nil
}

func (r *OrderRepo) ByQuote(ctx context.Context, quote shared.ID) (*ordering.CustomerOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byQuote[quote]
	if !ok {
		return nil, fmt.Errorf("order for quote %s: %w", quote, ordering.ErrOrderNotFound)
	}
	return r.byID[id], nil
}

func (r *OrderRepo) Save(ctx context.Context, o *ordering.CustomerOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[o.ID()] = o
	r.byQuote[o.Quote()] = o.ID()
	return nil
}

func (r *OrderRepo) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byID)
}

// VariantRepo is ordering's dictionary of catalog's variants: does this id
// exist, and whose product is it.
type OrderingVariantRepo struct {
	mu   sync.Mutex
	byID map[shared.ID]ordering.Variant
}

var _ ordering.VariantRepository = (*OrderingVariantRepo)(nil)

func NewOrderingVariantRepo() *OrderingVariantRepo {
	return &OrderingVariantRepo{byID: map[shared.ID]ordering.Variant{}}
}

func (r *OrderingVariantRepo) ByID(ctx context.Context, variant shared.ID) (ordering.Variant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.byID[variant]
	if !ok {
		return ordering.Variant{}, fmt.Errorf("variant %s: %w", variant, ordering.ErrVariantNotFound)
	}
	return v, nil
}

func (r *OrderingVariantRepo) Save(ctx context.Context, v ordering.Variant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[v.Variant] = v
	return nil
}

type AcceptedQuoteRepo struct {
	mu   sync.Mutex
	byID map[shared.ID]ordering.AcceptedQuote
}

var _ ordering.AcceptedQuoteRepository = (*AcceptedQuoteRepo)(nil)

func NewAcceptedQuoteRepo() *AcceptedQuoteRepo {
	return &AcceptedQuoteRepo{byID: map[shared.ID]ordering.AcceptedQuote{}}
}

func (r *AcceptedQuoteRepo) ByID(ctx context.Context, quote shared.ID) (ordering.AcceptedQuote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	q, ok := r.byID[quote]
	if !ok {
		return ordering.AcceptedQuote{}, fmt.Errorf("accepted quote %s: %w", quote, ordering.ErrAcceptedQuoteNotFound)
	}
	return q, nil
}

func (r *AcceptedQuoteRepo) Save(ctx context.Context, q ordering.AcceptedQuote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[q.Quote] = q
	return nil
}
