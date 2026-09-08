package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/duongsy/portage/internal/domain/catalog"
)

// MerchantRepo implements catalog.MerchantRepository over a map.
type MerchantRepo struct {
	mu           sync.Mutex
	byID         map[catalog.MerchantID]*catalog.Merchant
	FailSaveWith error // set in a test to make Save fail
}

// Compile-time proof that the adapter satisfies the port. If the interface
// grows a method, this line stops the build — before any test runs.
//
// [PHP] Không có `implements` để compiler kiểm; dòng này là cách Go nói
// [PHP] "tôi khẳng định *MerchantRepo thoả MerchantRepository, kiểm đi".
var _ catalog.MerchantRepository = (*MerchantRepo)(nil)

func NewMerchantRepo() *MerchantRepo {
	return &MerchantRepo{byID: map[catalog.MerchantID]*catalog.Merchant{}}
}

func (r *MerchantRepo) ByID(ctx context.Context, id catalog.MerchantID) (*catalog.Merchant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("merchant %s: %w", id, catalog.ErrMerchantNotFound)
	}
	return m, nil
}

func (r *MerchantRepo) BySite(ctx context.Context, site catalog.Hostname) (*catalog.Merchant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.byID {
		if m.Site() == site {
			return m, nil
		}
	}
	return nil, fmt.Errorf("merchant at %s: %w", site, catalog.ErrMerchantNotFound)
}

// All returns the merchants sorted by id. Ids are UUIDv7, so that IS the order
// they were added in — the same order Postgres gives with ORDER BY id, which is
// what makes the two adapters interchangeable in a test.
func (r *MerchantRepo) All(ctx context.Context) ([]*catalog.Merchant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*catalog.Merchant, 0, len(r.byID))
	for _, m := range r.byID {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID().String() < out[j].ID().String() })
	return out, nil
}

func (r *MerchantRepo) Save(ctx context.Context, m *catalog.Merchant) error {
	if r.FailSaveWith != nil {
		return r.FailSaveWith
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[m.ID()] = m
	return nil
}

// Len is for tests: how many merchants are stored.
func (r *MerchantRepo) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byID)
}
