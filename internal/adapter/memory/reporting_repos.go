package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The read model's stores, in memory. Same shape as every other memory repo,
// with one difference worth noticing: OrderSummary is a plain struct, so a map
// of VALUES is enough — no pointers, no aliasing, no way for a caller to
// mutate a stored row by accident.

type OrderSummaryRepo struct {
	mu      sync.Mutex
	byOrder map[ordering.OrderID]reportingapp.OrderSummary
}

var _ reportingapp.OrderSummaryRepository = (*OrderSummaryRepo)(nil)

func NewOrderSummaryRepo() *OrderSummaryRepo {
	return &OrderSummaryRepo{byOrder: map[ordering.OrderID]reportingapp.OrderSummary{}}
}

func (r *OrderSummaryRepo) ByOrder(ctx context.Context, id ordering.OrderID) (reportingapp.OrderSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.byOrder[id]
	if !ok {
		return reportingapp.OrderSummary{}, fmt.Errorf("order summary %s: %w", id, reportingapp.ErrSummaryNotFound)
	}
	return s, nil
}

func (r *OrderSummaryRepo) ByCustomer(ctx context.Context, customer shared.ID) ([]reportingapp.OrderSummary, error) {
	return r.filter(func(s reportingapp.OrderSummary) bool { return s.Customer == customer }), nil
}

func (r *OrderSummaryRepo) ByStatus(ctx context.Context, status ordering.OrderStatus) ([]reportingapp.OrderSummary, error) {
	return r.filter(func(s reportingapp.OrderSummary) bool { return s.Status == status }), nil
}

func (r *OrderSummaryRepo) ByProduct(ctx context.Context, product shared.ID) ([]reportingapp.OrderSummary, error) {
	return r.filter(func(s reportingapp.OrderSummary) bool { return s.Product == product }), nil
}

func (r *OrderSummaryRepo) All(ctx context.Context) ([]reportingapp.OrderSummary, error) {
	return r.filter(func(reportingapp.OrderSummary) bool { return true }), nil
}

func (r *OrderSummaryRepo) Save(ctx context.Context, s reportingapp.OrderSummary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byOrder[s.Order] = s
	return nil
}

func (r *OrderSummaryRepo) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byOrder)
}

// filter walks the map and sorts NEWEST FIRST, the order a "my orders" screen
// wants. Map iteration in Go is deliberately randomised, so a list repository
// that did not sort would return a different order every call — and a test
// written against it would pass or fail by luck.
func (r *OrderSummaryRepo) filter(keep func(reportingapp.OrderSummary) bool) []reportingapp.OrderSummary {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]reportingapp.OrderSummary, 0, len(r.byOrder))
	for _, s := range r.byOrder {
		if keep(s) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if a, b := out[i].PlacedAt, out[j].PlacedAt; !a.Equal(b) {
			return a.After(b)
		}
		return out[i].Order.String() < out[j].Order.String() // a tie still has ONE order
	})
	return out
}

type ProductNameRepo struct {
	mu   sync.Mutex
	byID map[shared.ID]string
}

var _ reportingapp.ProductNames = (*ProductNameRepo)(nil)

func NewProductNameRepo() *ProductNameRepo {
	return &ProductNameRepo{byID: map[shared.ID]string{}}
}

func (r *ProductNameRepo) Name(ctx context.Context, product shared.ID) (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name, ok := r.byID[product]
	return name, ok, nil
}

func (r *ProductNameRepo) Save(ctx context.Context, product shared.ID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[product] = name
	return nil
}

// ProductWorklistRepo is the in-memory store for the worklist read model.
type ProductWorklistRepo struct {
	mu        sync.Mutex
	byProduct map[shared.ID]reportingapp.WorklistItem
}

var _ reportingapp.ProductWorklistRepository = (*ProductWorklistRepo)(nil)

func NewProductWorklistRepo() *ProductWorklistRepo {
	return &ProductWorklistRepo{byProduct: map[shared.ID]reportingapp.WorklistItem{}}
}

func (r *ProductWorklistRepo) ByProduct(ctx context.Context, product shared.ID) (reportingapp.WorklistItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.byProduct[product]
	if !ok {
		return reportingapp.WorklistItem{}, fmt.Errorf("worklist item %s: %w", product, reportingapp.ErrWorklistItemNotFound)
	}
	return w, nil
}

// Open is unfinished work, oldest first: a queue, so the thing that has been
// waiting longest is the thing a person sees at the top.
func (r *ProductWorklistRepo) Open(ctx context.Context) ([]reportingapp.WorklistItem, error) {
	return r.filterWorklist(func(w reportingapp.WorklistItem) bool { return !w.Published }), nil
}

func (r *ProductWorklistRepo) ByRequester(ctx context.Context, customer shared.ID) ([]reportingapp.WorklistItem, error) {
	if customer.IsZero() {
		// Nobody is "the zero customer". Answering with every unrequested row
		// would hand one person somebody else's list.
		return []reportingapp.WorklistItem{}, nil
	}
	return r.filterWorklist(func(w reportingapp.WorklistItem) bool { return w.RequestedBy == customer }), nil
}

func (r *ProductWorklistRepo) Save(ctx context.Context, item reportingapp.WorklistItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byProduct[item.Product] = item
	return nil
}

func (r *ProductWorklistRepo) filterWorklist(keep func(reportingapp.WorklistItem) bool) []reportingapp.WorklistItem {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []reportingapp.WorklistItem{}
	for _, w := range r.byProduct {
		if keep(w) {
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].AddedAt.Equal(out[j].AddedAt) {
			return out[i].AddedAt.Before(out[j].AddedAt)
		}
		return out[i].Product.String() < out[j].Product.String() // ties break stably
	})
	return out
}
