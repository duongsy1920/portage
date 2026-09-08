package logistics

import (
	"fmt"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

type BatchID struct {
	shared.ID
}

func NewBatchID() BatchID {
	return BatchID{shared.NewID()}
}

func ParseBatchID(s string) (BatchID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return BatchID{}, fmt.Errorf("batch id: %w", err)
	}
	return BatchID{id}, nil
}

// BatchStatus: open (taking parcels) → closed (taped, waiting for the
// carrier) → shipped (invoice in hand, freight split).
type BatchStatus string

const (
	BatchOpen    BatchStatus = "open"
	BatchClosed  BatchStatus = "closed"
	BatchShipped BatchStatus = "shipped"
)

func (s BatchStatus) IsValid() bool {
	switch s {
	case BatchOpen, BatchClosed, BatchShipped:
		return true
	}
	return false
}

// BatchItem is one parcel inside a batch: what the allocator needs to know.
type BatchItem struct {
	Parcel shared.ID
	Order  shared.ID
	Actual shared.ParcelSpec
}

// ConsolidationBatch is an AGGREGATE ROOT: one carton going out on one lane,
// several customers' parcels inside. "Gom lô, chờ ~4 tuần" (DDD.md §31).
type ConsolidationBatch struct {
	shared.Events

	id          BatchID
	lane        string // the lane's code, as logistics was told (LaneRule)
	status      BatchStatus
	items       []BatchItem
	freight     shared.Money // what the carrier charged, once shipped
	allocations []Allocation
	openedAt    time.Time
	closedAt    time.Time
	shippedAt   time.Time
}

func OpenBatch(lane string, now time.Time) (*ConsolidationBatch, error) {
	lane = strings.TrimSpace(lane)
	if lane == "" {
		return nil, fmt.Errorf("open batch: no lane: %w", ErrInvalidBatch)
	}
	b := &ConsolidationBatch{id: NewBatchID(), lane: lane, status: BatchOpen, openedAt: now}
	b.Record(BatchOpened{ID: b.id, Lane: lane, At: now})
	return b, nil
}

func (b *ConsolidationBatch) ID() BatchID {
	return b.id
}

func (b *ConsolidationBatch) Lane() string {
	return b.lane
}

func (b *ConsolidationBatch) Status() BatchStatus {
	return b.status
}

// Items returns a copy: the batch owns its list.
func (b *ConsolidationBatch) Items() []BatchItem {
	return append([]BatchItem(nil), b.items...)
}

func (b *ConsolidationBatch) Freight() shared.Money {
	return b.freight
}

func (b *ConsolidationBatch) Allocations() []Allocation {
	return append([]Allocation(nil), b.allocations...)
}

func (b *ConsolidationBatch) OpenedAt() time.Time {
	return b.openedAt
}

func (b *ConsolidationBatch) ClosedAt() time.Time {
	return b.closedAt
}

func (b *ConsolidationBatch) ShippedAt() time.Time {
	return b.shippedAt
}

// AddParcel puts a RECEIVED parcel (it has a measurement) into an open batch.
// One parcel once; one order once — two boxes for one order is a different
// design and would be a second parcel on the order, not a duplicate here.
func (b *ConsolidationBatch) AddParcel(item BatchItem) error {
	if b.status != BatchOpen {
		return fmt.Errorf("add to batch %s: status %s: %w", b.id, b.status, ErrBatchNotOpen)
	}
	if item.Parcel.IsZero() || item.Order.IsZero() || item.Actual.IsZero() {
		return fmt.Errorf("add to batch %s: parcel without an order or a measurement: %w", b.id, ErrInvalidParcel)
	}
	for _, have := range b.items {
		if have.Parcel == item.Parcel || have.Order == item.Order {
			return fmt.Errorf("add to batch %s: parcel %s / order %s: %w", b.id, item.Parcel, item.Order, ErrDuplicateParcel)
		}
	}
	b.items = append(b.items, item)
	return nil
}

// Close tapes the box: nothing goes in after this.
func (b *ConsolidationBatch) Close(now time.Time) error {
	if b.status != BatchOpen {
		return fmt.Errorf("close batch %s: status %s: %w", b.id, b.status, ErrBatchNotOpen)
	}
	if len(b.items) == 0 {
		return fmt.Errorf("close batch %s: %w", b.id, ErrBatchEmpty)
	}
	b.status, b.closedAt = BatchClosed, now
	b.Record(BatchClosedEvent{ID: b.id, Parcels: len(b.items), At: now})
	return nil
}

// Ship records the carrier's invoice for the whole box and splits it over
// the orders inside with the given allocator — the moment each order learns
// its ACTUAL freight. The split is frozen into the event ordering and pricing
// consume; recomputing it later with another rule would be rewriting history.
func (b *ConsolidationBatch) Ship(freight shared.Money, alloc FreightAllocator, now time.Time) ([]Allocation, error) {
	if b.status != BatchClosed {
		return nil, fmt.Errorf("ship batch %s: status %s: %w", b.id, b.status, ErrBatchNotClosed)
	}
	if !freight.IsValid() || !freight.IsPositive() {
		return nil, fmt.Errorf("ship batch %s: freight %s: %w", b.id, freight, ErrInvalidFreight)
	}
	if alloc == nil {
		return nil, fmt.Errorf("ship batch %s: no allocator: %w", b.id, ErrInvalidBatch)
	}
	allocations, err := alloc.Allocate(freight, b.items)
	if err != nil {
		return nil, fmt.Errorf("ship batch %s: %w", b.id, err)
	}
	b.status, b.freight, b.allocations, b.shippedAt = BatchShipped, freight, allocations, now
	b.Record(BatchShippedEvent{ID: b.id, Lane: b.lane, Freight: freight, Allocations: append([]Allocation(nil), allocations...), At: now})
	return b.Allocations(), nil
}

type BatchSnapshot struct {
	ID          BatchID
	Lane        string
	Status      BatchStatus
	Items       []BatchItem
	Freight     shared.Money
	Allocations []Allocation
	OpenedAt    time.Time
	ClosedAt    time.Time
	ShippedAt   time.Time
}

func (b *ConsolidationBatch) Snapshot() BatchSnapshot {
	return BatchSnapshot{
		ID: b.id, Lane: b.lane, Status: b.status, Items: b.Items(), Freight: b.freight, Allocations: b.Allocations(),
		OpenedAt: b.openedAt, ClosedAt: b.closedAt, ShippedAt: b.shippedAt,
	}
}

func BatchFromSnapshot(s BatchSnapshot) (*ConsolidationBatch, error) {
	bad := func(why string) error {
		return fmt.Errorf("batch snapshot %s: %s: %w", s.ID, why, ErrInvalidSnapshot)
	}
	switch {
	case s.ID.IsZero():
		return nil, fmt.Errorf("batch snapshot: no id: %w", ErrInvalidSnapshot)
	case s.Lane == "":
		return nil, bad("no lane")
	case !s.Status.IsValid():
		return nil, bad(fmt.Sprintf("status %q", s.Status))
	case s.Status != BatchOpen && (len(s.Items) == 0 || s.ClosedAt.IsZero()):
		return nil, bad("closed without parcels or a time")
	case s.Status == BatchShipped && (!s.Freight.IsValid() || len(s.Allocations) != len(s.Items) || s.ShippedAt.IsZero()):
		return nil, bad("shipped without an invoice or a full allocation")
	case s.Status != BatchShipped && (s.Freight.IsValid() || len(s.Allocations) != 0):
		return nil, bad("allocation on an unshipped batch")
	}
	return &ConsolidationBatch{
		id: s.ID, lane: s.Lane, status: s.Status, items: append([]BatchItem(nil), s.Items...), freight: s.Freight,
		allocations: append([]Allocation(nil), s.Allocations...), openedAt: s.OpenedAt, closedAt: s.ClosedAt, shippedAt: s.ShippedAt,
	}, nil
}
