package logistics

import (
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Logistics' events. BatchShippedEvent is the one others wait for: ordering
// moves each order to in_transit, pricing learns each order's ACTUAL freight.

type ParcelExpectedEvent struct {
	ID        ParcelID
	Order     shared.ID
	Reference string
	At        time.Time
}

func (ParcelExpectedEvent) EventName() string {
	return "logistics.parcel_expected"
}

func (e ParcelExpectedEvent) OccurredAt() time.Time {
	return e.At
}

type ParcelReceivedEvent struct {
	ID     ParcelID
	Order  shared.ID
	Actual shared.ParcelSpec
	By     shared.OperatorID
	At     time.Time
}

func (ParcelReceivedEvent) EventName() string {
	return "logistics.parcel_received"
}

func (e ParcelReceivedEvent) OccurredAt() time.Time {
	return e.At
}

type BatchOpened struct {
	ID   BatchID
	Lane string
	At   time.Time
}

func (BatchOpened) EventName() string {
	return "logistics.batch_opened"
}

func (e BatchOpened) OccurredAt() time.Time {
	return e.At
}

type BatchClosedEvent struct {
	ID      BatchID
	Parcels int
	At      time.Time
}

func (BatchClosedEvent) EventName() string {
	return "logistics.batch_closed"
}

func (e BatchClosedEvent) OccurredAt() time.Time {
	return e.At
}

type BatchShippedEvent struct {
	ID          BatchID
	Lane        string
	Freight     shared.Money
	Allocations []Allocation
	At          time.Time
}

func (BatchShippedEvent) EventName() string {
	return "logistics.batch_shipped"
}

func (e BatchShippedEvent) OccurredAt() time.Time {
	return e.At
}
