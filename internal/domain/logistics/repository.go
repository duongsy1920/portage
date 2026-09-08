package logistics

import (
	"context"

	"github.com/duongsy/portage/internal/domain/shared"
)

// The PORTS of logistics.

type ParcelRepository interface {
	ByID(ctx context.Context, id ParcelID) (*Parcel, error)        // ErrParcelNotFound
	ByOrder(ctx context.Context, order shared.ID) (*Parcel, error) // ErrParcelNotFound
	// Pending lists parcels not yet shipped (expected, received, batched), oldest first: the warehouse screen.
	Pending(ctx context.Context) ([]*Parcel, error)
	Save(ctx context.Context, p *Parcel) error
}

type BatchRepository interface {
	ByID(ctx context.Context, id BatchID) (*ConsolidationBatch, error) // ErrBatchNotFound
	Save(ctx context.Context, b *ConsolidationBatch) error
}

// LaneRuleRepository stores logistics' projection of pricing's lanes.
type LaneRuleRepository interface {
	ByCode(ctx context.Context, code string) (LaneRule, error) // ErrLaneRuleNotFound
	Save(ctx context.Context, r LaneRule) error
}
