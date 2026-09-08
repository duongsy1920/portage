package ordering

import (
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Ordering's events. DepositPaid is the one procurement waits for: it carries
// product and variant so the buyer knows WHAT to buy without asking anyone.

type OrderPlaced struct {
	ID       OrderID
	Quote    shared.ID
	Product  shared.ID
	Variant  shared.ID
	Customer shared.ID
	Total    shared.Money
	Deposit  shared.Money
	At       time.Time
}

func (OrderPlaced) EventName() string {
	return "ordering.order_placed"
}

func (e OrderPlaced) OccurredAt() time.Time {
	return e.At
}

type DepositPaid struct {
	ID      OrderID
	Quote   shared.ID
	Product shared.ID
	Variant shared.ID
	Amount  shared.Money
	At      time.Time
}

func (DepositPaid) EventName() string {
	return "ordering.deposit_paid"
}

func (e DepositPaid) OccurredAt() time.Time {
	return e.At
}

type OrderPurchased struct {
	ID OrderID
	At time.Time
}

func (OrderPurchased) EventName() string {
	return "ordering.order_purchased"
}

func (e OrderPurchased) OccurredAt() time.Time {
	return e.At
}

type OrderPurchaseFailed struct {
	ID     OrderID
	Reason string
	At     time.Time
}

func (OrderPurchaseFailed) EventName() string {
	return "ordering.order_purchase_failed"
}

func (e OrderPurchaseFailed) OccurredAt() time.Time {
	return e.At
}

type OrderShipped struct {
	ID OrderID
	At time.Time
}

func (OrderShipped) EventName() string {
	return "ordering.order_shipped"
}

func (e OrderShipped) OccurredAt() time.Time {
	return e.At
}

type BalancePaid struct {
	ID     OrderID
	Amount shared.Money
	At     time.Time
}

func (BalancePaid) EventName() string {
	return "ordering.balance_paid"
}

func (e BalancePaid) OccurredAt() time.Time {
	return e.At
}

type OrderDelivered struct {
	ID OrderID
	At time.Time
}

func (OrderDelivered) EventName() string {
	return "ordering.order_delivered"
}

func (e OrderDelivered) OccurredAt() time.Time {
	return e.At
}

type OrderCancelled struct {
	ID        OrderID
	Reason    string
	Refund    shared.Money
	Forfeited bool
	At        time.Time
}

func (OrderCancelled) EventName() string {
	return "ordering.order_cancelled"
}

func (e OrderCancelled) OccurredAt() time.Time {
	return e.At
}
