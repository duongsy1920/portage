package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/duongsy/portage/internal/domain/catalog"
)

// CategoryRepo implements catalog.CategoryRepository over a map keyed by the
// natural code. Save replaces the whole policy — it is a value object.
type CategoryRepo struct {
	mu     sync.Mutex
	byCode map[catalog.CategoryCode]catalog.CategoryPolicy
}

var _ catalog.CategoryRepository = (*CategoryRepo)(nil)

func NewCategoryRepo() *CategoryRepo {
	return &CategoryRepo{byCode: map[catalog.CategoryCode]catalog.CategoryPolicy{}}
}

func (r *CategoryRepo) ByCode(ctx context.Context, code catalog.CategoryCode) (catalog.CategoryPolicy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byCode[code]
	if !ok {
		return catalog.CategoryPolicy{}, fmt.Errorf("category %s: %w", code, catalog.ErrCategoryNotFound)
	}
	return p, nil
}

func (r *CategoryRepo) All(ctx context.Context) ([]catalog.CategoryPolicy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]catalog.CategoryPolicy, 0, len(r.byCode))
	for _, p := range r.byCode {
		out = append(out, p)
	}
	// By code, because Postgres says ORDER BY code and the two adapters have
	// to agree. Go map iteration is deliberately randomised, so without this
	// the form that lists categories picks a different default on every load —
	// a select box that reshuffles under the person using it.
	sort.Slice(out, func(i, j int) bool { return out[i].Code().String() < out[j].Code().String() })
	return out, nil
}

func (r *CategoryRepo) Save(ctx context.Context, p catalog.CategoryPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byCode[p.Code()] = p
	return nil
}
