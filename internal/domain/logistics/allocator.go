package logistics

import (
	"fmt"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Allocation is one order's share of a batch's freight.
type Allocation struct {
	Parcel     shared.ID
	Order      shared.ID
	Chargeable shared.Weight // what the carrier billed this parcel on
	Freight    shared.Money  // this order's share of the invoice
}

// FreightAllocator is the DOMAIN SERVICE of DDD.md §17: how one invoice for
// one carton becomes N lines for N orders. By chargeable weight? by volume?
// by value? Each rule favours a different customer, which is why it is an
// interface and not a formula inlined in Ship.
type FreightAllocator interface {
	Allocate(freight shared.Money, items []BatchItem) ([]Allocation, error)
}

// LaneRule is what logistics knows about a lane: the two constants the
// carrier bills with. A PROJECTION of pricing.lane_defined — pricing owns
// the price list, logistics only needs how weight is counted.
type LaneRule struct {
	Code    string
	Divisor int64
	Step    shared.Weight
}

// ByChargeableWeight splits in proportion to what the carrier actually
// bills each parcel on — max(actual, volumetric) rounded up to the step —
// so the customer with the big light box pays for the air in it, exactly
// as the carrier charges us. Cents reconcile by shared.Allocate.
type ByChargeableWeight struct {
	Divisor int64
	Step    shared.Weight
}

var _ FreightAllocator = ByChargeableWeight{}

func (r ByChargeableWeight) Allocate(freight shared.Money, items []BatchItem) ([]Allocation, error) {
	if r.Divisor <= 0 || r.Step.IsZero() {
		return nil, fmt.Errorf("allocate: lane rule divisor %d step %s: %w", r.Divisor, r.Step, ErrLaneRuleNotFound)
	}
	weights := make([]int64, len(items))
	chargeable := make([]shared.Weight, len(items))
	for i, it := range items {
		chargeable[i] = shared.ChargeableWeight(it.Actual.Weight(), it.Actual.Dimensions(), r.Divisor, r.Step)
		weights[i] = chargeable[i].Grams()
	}
	parts, err := shared.Allocate(freight, weights)
	if err != nil {
		return nil, err
	}
	out := make([]Allocation, len(items))
	for i, it := range items {
		out[i] = Allocation{Parcel: it.Parcel, Order: it.Order, Chargeable: chargeable[i], Freight: parts[i]}
	}
	return out, nil
}
