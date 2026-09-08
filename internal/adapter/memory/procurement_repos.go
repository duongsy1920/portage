package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The procurement ports, in memory.

type TaskRepo struct {
	mu      sync.Mutex
	byID    map[procurement.TaskID]*procurement.PurchaseTask
	byOrder map[shared.ID]procurement.TaskID
}

var _ procurement.TaskRepository = (*TaskRepo)(nil)

func NewTaskRepo() *TaskRepo {
	return &TaskRepo{byID: map[procurement.TaskID]*procurement.PurchaseTask{}, byOrder: map[shared.ID]procurement.TaskID{}}
}

func (r *TaskRepo) ByID(ctx context.Context, id procurement.TaskID) (*procurement.PurchaseTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("purchase task %s: %w", id, procurement.ErrTaskNotFound)
	}
	return t, nil
}

func (r *TaskRepo) ByOrder(ctx context.Context, order shared.ID) (*procurement.PurchaseTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byOrder[order]
	if !ok {
		return nil, fmt.Errorf("purchase task for order %s: %w", order, procurement.ErrTaskNotFound)
	}
	return r.byID[id], nil
}

func (r *TaskRepo) Open(ctx context.Context) ([]*procurement.PurchaseTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*procurement.PurchaseTask
	for _, t := range r.byID {
		if t.Status() == procurement.TaskOpen {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].OpenedAt().Before(out[j].OpenedAt()) || (out[i].OpenedAt().Equal(out[j].OpenedAt()) && out[i].ID().String() < out[j].ID().String())
	})
	return out, nil
}

func (r *TaskRepo) Save(ctx context.Context, t *procurement.PurchaseTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[t.ID()] = t
	r.byOrder[t.Order()] = t.ID()
	return nil
}

type ShopRepo struct {
	mu   sync.Mutex
	byID map[shared.ID]procurement.Shop
}

var _ procurement.ShopRepository = (*ShopRepo)(nil)

func NewShopRepo() *ShopRepo {
	return &ShopRepo{byID: map[shared.ID]procurement.Shop{}}
}

func (r *ShopRepo) ByID(ctx context.Context, merchant shared.ID) (procurement.Shop, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.byID[merchant]
	if !ok {
		return procurement.Shop{}, fmt.Errorf("shop %s: %w", merchant, procurement.ErrShopNotFound)
	}
	return s, nil
}

func (r *ShopRepo) Save(ctx context.Context, s procurement.Shop) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[s.Merchant] = s
	return nil
}

// VariantRepo answers "which size does this variant id mean" — procurement's
// dictionary from catalog.variant_added.
type VariantRepo struct {
	mu   sync.Mutex
	byID map[shared.ID]procurement.Variant
}

var _ procurement.VariantRepository = (*VariantRepo)(nil)

func NewVariantRepo() *VariantRepo {
	return &VariantRepo{byID: map[shared.ID]procurement.Variant{}}
}

func (r *VariantRepo) ByID(ctx context.Context, variant shared.ID) (procurement.Variant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.byID[variant]
	if !ok {
		return procurement.Variant{}, fmt.Errorf("variant %s: %w", variant, procurement.ErrVariantNotFound)
	}
	return v, nil
}

func (r *VariantRepo) Save(ctx context.Context, v procurement.Variant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[v.Variant] = v
	return nil
}

type ItemRepo struct {
	mu        sync.Mutex
	byProduct map[shared.ID]procurement.Item
}

var _ procurement.ItemRepository = (*ItemRepo)(nil)

func NewItemRepo() *ItemRepo {
	return &ItemRepo{byProduct: map[shared.ID]procurement.Item{}}
}

func (r *ItemRepo) ByProduct(ctx context.Context, product shared.ID) (procurement.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := r.byProduct[product]
	if !ok {
		return procurement.Item{}, fmt.Errorf("item %s: %w", product, procurement.ErrItemNotFound)
	}
	return i, nil
}

func (r *ItemRepo) Save(ctx context.Context, i procurement.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byProduct[i.Product] = i
	return nil
}
