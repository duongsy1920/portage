package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/duongsy/portage/internal/domain/catalog"
)

// ProductRepo implements catalog.ProductRepository over a map.
type ProductRepo struct {
	mu           sync.Mutex
	byID         map[catalog.ProductID]*catalog.Product
	FailSaveWith error
}

var _ catalog.ProductRepository = (*ProductRepo)(nil)

func NewProductRepo() *ProductRepo {
	return &ProductRepo{byID: map[catalog.ProductID]*catalog.Product{}}
}

func (r *ProductRepo) ByID(ctx context.Context, id catalog.ProductID) (*catalog.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("product %s: %w", id, catalog.ErrProductNotFound)
	}
	return p, nil
}

// BySource matches on the URL exactly as stored. Smarter matching (same page,
// different tracking parameters) is the detector's job, not the store's.
func (r *ProductRepo) BySource(ctx context.Context, source catalog.SourceURL) ([]*catalog.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*catalog.Product
	for _, p := range r.byID {
		if p.Source() == source {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *ProductRepo) Save(ctx context.Context, p *catalog.Product) error {
	if r.FailSaveWith != nil {
		return r.FailSaveWith
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[p.ID()] = p
	return nil
}

func (r *ProductRepo) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byID)
}
