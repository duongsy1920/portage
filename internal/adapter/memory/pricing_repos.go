package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The pricing ports, in memory. Same shape as the catalog ones: a map, a
// mutex, the domain's sentinel on a miss.

type LaneRepo struct {
	mu     sync.Mutex
	byCode map[pricing.LaneCode]pricing.ShippingLane
}

var _ pricing.LaneRepository = (*LaneRepo)(nil)

func NewLaneRepo() *LaneRepo {
	return &LaneRepo{byCode: map[pricing.LaneCode]pricing.ShippingLane{}}
}

func (r *LaneRepo) ByCode(ctx context.Context, code pricing.LaneCode) (pricing.ShippingLane, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byCode[code]
	if !ok {
		return pricing.ShippingLane{}, fmt.Errorf("lane %s: %w", code, pricing.ErrLaneNotFound)
	}
	return l, nil
}

func (r *LaneRepo) All(ctx context.Context) ([]pricing.ShippingLane, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]pricing.ShippingLane, 0, len(r.byCode))
	for _, l := range r.byCode {
		out = append(out, l)
	}
	return out, nil
}

func (r *LaneRepo) Save(ctx context.Context, lane pricing.ShippingLane) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byCode[lane.Code()] = lane
	return nil
}

type QuoteRepo struct {
	mu   sync.Mutex
	byID map[pricing.QuoteID]*pricing.Quote
}

var _ pricing.QuoteRepository = (*QuoteRepo)(nil)

func NewQuoteRepo() *QuoteRepo {
	return &QuoteRepo{byID: map[pricing.QuoteID]*pricing.Quote{}}
}

func (r *QuoteRepo) ByID(ctx context.Context, id pricing.QuoteID) (*pricing.Quote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	q, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("quote %s: %w", id, pricing.ErrQuoteNotFound)
	}
	return q, nil
}

func (r *QuoteRepo) Save(ctx context.Context, q *pricing.Quote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[q.ID()] = q
	return nil
}

// IssuedBefore answers the expiry sweep. The order is the deadline, oldest
// first, so a limited pass always takes the quotes that have been waiting
// longest — and so two runs over the same data see the same order.
func (r *QuoteRepo) IssuedBefore(ctx context.Context, t time.Time, limit int) ([]*pricing.Quote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*pricing.Quote
	for _, q := range r.byID {
		if q.Status() == pricing.QuoteIssued && q.ExpiresAt().Before(t) {
			out = append(out, q)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if a, b := out[i].ExpiresAt(), out[j].ExpiresAt(); !a.Equal(b) {
			return a.Before(b)
		}
		return out[i].ID().String() < out[j].ID().String() // a tie still has ONE order
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (r *QuoteRepo) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byID)
}

type ListingRepo struct {
	mu        sync.Mutex
	byProduct map[shared.ID]pricing.Listing
}

var _ pricing.ListingRepository = (*ListingRepo)(nil)

func NewListingRepo() *ListingRepo {
	return &ListingRepo{byProduct: map[shared.ID]pricing.Listing{}}
}

func (r *ListingRepo) ByProduct(ctx context.Context, product shared.ID) (pricing.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byProduct[product]
	if !ok {
		return pricing.Listing{}, fmt.Errorf("listing for product %s: %w", product, pricing.ErrListingNotFound)
	}
	return l, nil
}

func (r *ListingRepo) Save(ctx context.Context, l pricing.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byProduct[l.Product] = l
	return nil
}

func (r *ListingRepo) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byProduct)
}

type ProfileRepo struct {
	mu     sync.Mutex
	byCode map[string]pricing.CategoryProfile
}

var _ pricing.CategoryProfileRepository = (*ProfileRepo)(nil)

func NewProfileRepo() *ProfileRepo {
	return &ProfileRepo{byCode: map[string]pricing.CategoryProfile{}}
}

func (r *ProfileRepo) ByCode(ctx context.Context, code string) (pricing.CategoryProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byCode[code]
	if !ok {
		return pricing.CategoryProfile{}, fmt.Errorf("category profile %s: %w", code, pricing.ErrProfileNotFound)
	}
	return p, nil
}

func (r *ProfileRepo) Save(ctx context.Context, p pricing.CategoryProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byCode[p.Code] = p
	return nil
}

// ExchangeRates keeps one current rate per currency pair.
type ExchangeRates struct {
	mu    sync.Mutex
	rates map[[2]shared.Currency]shared.ExchangeRate
}

var _ pricing.ExchangeRates = (*ExchangeRates)(nil)

func NewExchangeRates() *ExchangeRates {
	return &ExchangeRates{rates: map[[2]shared.Currency]shared.ExchangeRate{}}
}

func (r *ExchangeRates) Current(ctx context.Context, from, to shared.Currency) (shared.ExchangeRate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rate, ok := r.rates[[2]shared.Currency{from, to}]
	if !ok {
		return shared.ExchangeRate{}, fmt.Errorf("%s→%s: %w", from, to, pricing.ErrNoExchangeRate)
	}
	return rate, nil
}

func (r *ExchangeRates) Set(ctx context.Context, rate shared.ExchangeRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rates[[2]shared.Currency{rate.From(), rate.To()}] = rate
	return nil
}
