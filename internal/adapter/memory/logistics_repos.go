package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The logistics ports, in memory — and pricing's reconciliation store.

type ParcelRepo struct {
	mu      sync.Mutex
	byID    map[logistics.ParcelID]*logistics.Parcel
	byOrder map[shared.ID]logistics.ParcelID
}

var _ logistics.ParcelRepository = (*ParcelRepo)(nil)

func NewParcelRepo() *ParcelRepo {
	return &ParcelRepo{byID: map[logistics.ParcelID]*logistics.Parcel{}, byOrder: map[shared.ID]logistics.ParcelID{}}
}

func (r *ParcelRepo) ByID(ctx context.Context, id logistics.ParcelID) (*logistics.Parcel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("parcel %s: %w", id, logistics.ErrParcelNotFound)
	}
	return p, nil
}

func (r *ParcelRepo) ByOrder(ctx context.Context, order shared.ID) (*logistics.Parcel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byOrder[order]
	if !ok {
		return nil, fmt.Errorf("parcel for order %s: %w", order, logistics.ErrParcelNotFound)
	}
	return r.byID[id], nil
}

func (r *ParcelRepo) Pending(ctx context.Context) ([]*logistics.Parcel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*logistics.Parcel
	for _, p := range r.byID {
		if p.Status() != logistics.ParcelShipped {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ExpectedAt().Before(out[j].ExpectedAt()) || (out[i].ExpectedAt().Equal(out[j].ExpectedAt()) && out[i].ID().String() < out[j].ID().String())
	})
	return out, nil
}

func (r *ParcelRepo) Save(ctx context.Context, p *logistics.Parcel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[p.ID()] = p
	r.byOrder[p.Order()] = p.ID()
	return nil
}

type BatchRepo struct {
	mu   sync.Mutex
	byID map[logistics.BatchID]*logistics.ConsolidationBatch
}

var _ logistics.BatchRepository = (*BatchRepo)(nil)

func NewBatchRepo() *BatchRepo {
	return &BatchRepo{byID: map[logistics.BatchID]*logistics.ConsolidationBatch{}}
}

func (r *BatchRepo) ByID(ctx context.Context, id logistics.BatchID) (*logistics.ConsolidationBatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("batch %s: %w", id, logistics.ErrBatchNotFound)
	}
	return b, nil
}

func (r *BatchRepo) Save(ctx context.Context, b *logistics.ConsolidationBatch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[b.ID()] = b
	return nil
}

type LaneRuleRepo struct {
	mu     sync.Mutex
	byCode map[string]logistics.LaneRule
}

var _ logistics.LaneRuleRepository = (*LaneRuleRepo)(nil)

func NewLaneRuleRepo() *LaneRuleRepo {
	return &LaneRuleRepo{byCode: map[string]logistics.LaneRule{}}
}

func (r *LaneRuleRepo) ByCode(ctx context.Context, code string) (logistics.LaneRule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rule, ok := r.byCode[code]
	if !ok {
		return logistics.LaneRule{}, fmt.Errorf("lane rule %s: %w", code, logistics.ErrLaneRuleNotFound)
	}
	return rule, nil
}

func (r *LaneRuleRepo) Save(ctx context.Context, rule logistics.LaneRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byCode[rule.Code] = rule
	return nil
}

type ReconciliationRepo struct {
	mu      sync.Mutex
	byOrder map[shared.ID]pricing.Reconciliation
}

var _ pricing.ReconciliationRepository = (*ReconciliationRepo)(nil)

func NewReconciliationRepo() *ReconciliationRepo {
	return &ReconciliationRepo{byOrder: map[shared.ID]pricing.Reconciliation{}}
}

func (r *ReconciliationRepo) ByOrder(ctx context.Context, order shared.ID) (pricing.Reconciliation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.byOrder[order]
	if !ok {
		return pricing.Reconciliation{}, fmt.Errorf("reconciliation for order %s: %w", order, pricing.ErrReconciliationNotFound)
	}
	return rec, nil
}

func (r *ReconciliationRepo) Save(ctx context.Context, rec pricing.Reconciliation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byOrder[rec.Order] = rec
	return nil
}
