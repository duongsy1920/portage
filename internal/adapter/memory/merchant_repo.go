package memory

import (
	"context"
	"fmt"
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
