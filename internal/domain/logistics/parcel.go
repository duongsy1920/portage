// Package logistics is the bounded context that MOVES things: a parcel per
// purchased order arrives at the US warehouse, is weighed (the ACTUAL of
// Quote vs Actual, DDD.md §28), joins a consolidation batch, and leaves when
// the batch ships — at which point the carrier's one invoice is split over
// the orders in the box (FreightAllocator, DDD.md §17).
//
// [PHP] Bundle "Logistics": hai Entity Parcel/ConsolidationBatch và một
// [PHP] service FreightAllocator — cái service "chia tiền" mà nếu viết trong
// [PHP] Controller sẽ không ai test được.
package logistics

import (
	"fmt"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

type ParcelID struct {
	shared.ID
}

func NewParcelID() ParcelID {
	return ParcelID{shared.NewID()}
}

func ParseParcelID(s string) (ParcelID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return ParcelID{}, fmt.Errorf("parcel id: %w", err)
	}
	return ParcelID{id}, nil
}

// ParcelStatus: expected (bought, on its way to our warehouse) → received
// (on our scale) → batched (in a box) → shipped (the box left).
type ParcelStatus string

const (
	ParcelExpected ParcelStatus = "expected"
	ParcelReceived ParcelStatus = "received"
	ParcelBatched  ParcelStatus = "batched"
	ParcelShipped  ParcelStatus = "shipped"
)

func (s ParcelStatus) IsValid() bool {
	switch s {
	case ParcelExpected, ParcelReceived, ParcelBatched, ParcelShipped:
		return true
	}
	return false
}

// ParcelDetails is what expecting a parcel takes: which order it is for and
// the shop's reference — the string printed on the box, how the warehouse
// matches a delivery to an order.
type ParcelDetails struct {
	Order     shared.ID
	Reference string
}

// Parcel is an AGGREGATE ROOT: one physical box for one order.
type Parcel struct {
	shared.Events

	id         ParcelID
	order      shared.ID
	reference  string
	status     ParcelStatus
	actual     shared.ParcelSpec // what our scale said; zero until received
	receivedBy shared.OperatorID
	batch      BatchID
	expectedAt time.Time
	receivedAt time.Time
}

// ExpectParcel: procurement bought it, so a box will arrive.
func ExpectParcel(d ParcelDetails, now time.Time) (*Parcel, error) {
	d.Reference = strings.TrimSpace(d.Reference)
	switch {
	case d.Order.IsZero():
		return nil, fmt.Errorf("expect parcel: no order: %w", ErrInvalidParcel)
	case d.Reference == "":
		return nil, fmt.Errorf("expect parcel: no reference: %w", ErrInvalidParcel)
	}
	p := &Parcel{id: NewParcelID(), order: d.Order, reference: d.Reference, status: ParcelExpected, expectedAt: now}
	p.Record(ParcelExpectedEvent{ID: p.id, Order: p.order, Reference: p.reference, At: now})
	return p, nil
}

func (p *Parcel) ID() ParcelID {
	return p.id
}

func (p *Parcel) Order() shared.ID {
	return p.order
}

func (p *Parcel) Reference() string {
	return p.reference
}

func (p *Parcel) Status() ParcelStatus {
	return p.status
}

// Actual is the measured parcel; zero until Receive.
func (p *Parcel) Actual() shared.ParcelSpec {
	return p.actual
}

func (p *Parcel) ReceivedBy() shared.OperatorID {
	return p.receivedBy
}

func (p *Parcel) Batch() BatchID {
	return p.batch
}

func (p *Parcel) ExpectedAt() time.Time {
	return p.expectedAt
}

func (p *Parcel) ReceivedAt() time.Time {
	return p.receivedAt
}

// Receive is the warehouse moment: the box on OUR scale. This number, not
// the category default and not the shop's listing, is what the carrier will
// bill — and what reconciliation compares with the quote.
func (p *Parcel) Receive(actual shared.ParcelSpec, by shared.OperatorID, now time.Time) error {
	if p.status != ParcelExpected {
		return fmt.Errorf("receive parcel %s: status %s: %w", p.id, p.status, ErrParcelNotExpected)
	}
	if actual.IsZero() {
		return fmt.Errorf("receive parcel %s: %w", p.id, shared.ErrIncompleteParcelSpec)
	}
	if by.IsZero() {
		return fmt.Errorf("receive parcel %s: %w", p.id, shared.ErrOperatorRequired)
	}
	p.status, p.actual, p.receivedBy, p.receivedAt = ParcelReceived, actual, by, now
	p.Record(ParcelReceivedEvent{ID: p.id, Order: p.order, Actual: actual, By: by, At: now})
	return nil
}

// AssignToBatch: the box goes into a consolidation batch. No event: the
// batch announces its contents when it ships.
func (p *Parcel) AssignToBatch(batch BatchID) error {
	if p.status != ParcelReceived {
		return fmt.Errorf("batch parcel %s: status %s: %w", p.id, p.status, ErrParcelNotReceived)
	}
	if batch.IsZero() {
		return fmt.Errorf("batch parcel %s: %w", p.id, ErrInvalidBatch)
	}
	p.status, p.batch = ParcelBatched, batch
	return nil
}

// MarkShipped follows the batch out of the door.
func (p *Parcel) MarkShipped() error {
	if p.status != ParcelBatched {
		return fmt.Errorf("ship parcel %s: status %s: %w", p.id, p.status, ErrParcelNotBatched)
	}
	p.status = ParcelShipped
	return nil
}

// Item is the parcel as a batch sees it.
func (p *Parcel) Item() BatchItem {
	return BatchItem{Parcel: p.id.ID, Order: p.order, Actual: p.actual}
}

type ParcelSnapshot struct {
	ID         ParcelID
	Order      shared.ID
	Reference  string
	Status     ParcelStatus
	Actual     shared.ParcelSpec
	ReceivedBy shared.OperatorID
	Batch      BatchID
	ExpectedAt time.Time
	ReceivedAt time.Time
}

func (p *Parcel) Snapshot() ParcelSnapshot {
	return ParcelSnapshot{
		ID: p.id, Order: p.order, Reference: p.reference, Status: p.status, Actual: p.actual,
		ReceivedBy: p.receivedBy, Batch: p.batch, ExpectedAt: p.expectedAt, ReceivedAt: p.receivedAt,
	}
}

func ParcelFromSnapshot(s ParcelSnapshot) (*Parcel, error) {
	bad := func(why string) error {
		return fmt.Errorf("parcel snapshot %s: %s: %w", s.ID, why, ErrInvalidSnapshot)
	}
	received := s.Status != ParcelExpected
	switch {
	case s.ID.IsZero():
		return nil, fmt.Errorf("parcel snapshot: no id: %w", ErrInvalidSnapshot)
	case s.Order.IsZero(), s.Reference == "":
		return nil, bad("missing order or reference")
	case !s.Status.IsValid():
		return nil, bad(fmt.Sprintf("status %q", s.Status))
	case received && (s.Actual.IsZero() || s.ReceivedBy.IsZero() || s.ReceivedAt.IsZero()):
		return nil, bad("received without a measurement")
	case !received && (!s.Actual.IsZero() || !s.ReceivedBy.IsZero() || !s.Batch.IsZero()):
		return nil, bad("expected parcel with receipt data")
	case (s.Status == ParcelBatched || s.Status == ParcelShipped) && s.Batch.IsZero():
		return nil, bad("batched without a batch")
	}
	return &Parcel{
		id: s.ID, order: s.Order, reference: s.Reference, status: s.Status, actual: s.Actual,
		receivedBy: s.ReceivedBy, batch: s.Batch, expectedAt: s.ExpectedAt, receivedAt: s.ReceivedAt,
	}, nil
}
