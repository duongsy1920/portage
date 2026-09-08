package procurement

import (
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Procurement's events. Ordering listens for the two outcomes — that is the
// saga's next step (DDD.md §26): confirmed moves the order past the point of
// no return, failed asks it to compensate.

type PurchaseTaskOpened struct {
	ID      TaskID
	Order   shared.ID
	Product shared.ID
	Variant shared.ID
	At      time.Time
}

func (PurchaseTaskOpened) EventName() string {
	return "procurement.purchase_task_opened"
}

func (e PurchaseTaskOpened) OccurredAt() time.Time {
	return e.At
}

type PurchaseConfirmed struct {
	ID        TaskID
	Order     shared.ID
	Reference string
	Paid      shared.Money
	By        shared.OperatorID
	At        time.Time
}

func (PurchaseConfirmed) EventName() string {
	return "procurement.purchase_confirmed"
}

func (e PurchaseConfirmed) OccurredAt() time.Time {
	return e.At
}

type PurchaseFailed struct {
	ID     TaskID
	Order  shared.ID
	Reason string
	At     time.Time
}

func (PurchaseFailed) EventName() string {
	return "procurement.purchase_failed"
}

func (e PurchaseFailed) OccurredAt() time.Time {
	return e.At
}
