// Package procurement is the bounded context that BUYS: for every deposited
// order there is one PurchaseTask — "go to this shop, buy this variant of
// this product, for this order" — that ends confirmed (we have an order
// number and paid a real price) or failed (sold out, price jumped).
//
// This is the context that touches the outside world (DDD.md §22): the
// shop's website, its cart, its order confirmation. All of that sits behind
// the MerchantACL port; today's adapter is an operator with a browser typing
// what happened into POST /purchase-tasks/{id}/confirm. When a shop gets an
// API, the domain does not change a line.
//
// [PHP] Bundle "Procurement": Entity PurchaseTask + một interface
// [PHP] MerchantGateway (ACL) với implementation ManualGateway.
package procurement

import (
	"fmt"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

type TaskID struct {
	shared.ID
}

func NewTaskID() TaskID {
	return TaskID{shared.NewID()}
}

func ParseTaskID(s string) (TaskID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return TaskID{}, fmt.Errorf("purchase task id: %w", err)
	}
	return TaskID{id}, nil
}

// TaskStatus: open → confirmed | failed. Terminal states are terminal — a
// failed purchase is not retried on the same task; the order decides what
// happens next (cancel with a full refund, or a new order).
type TaskStatus string

const (
	TaskOpen      TaskStatus = "open"
	TaskConfirmed TaskStatus = "confirmed"
	TaskFailed    TaskStatus = "failed"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskOpen, TaskConfirmed, TaskFailed:
		return true
	}
	return false
}

// PurchaseReceipt is what the shop gave us back: their order reference and
// what we ACTUALLY paid — the first "Actual" of Quote vs Actual (DDD.md §28).
// The price may differ from the quoted item price (sale, tax rounding); the
// difference is ours to see at reconciliation, never silently overwritten.
type PurchaseReceipt struct {
	Reference string       // the shop's order number
	Paid      shared.Money // in the shop's currency, all-in (item + tax + domestic shipping)
	PaidBy    shared.OperatorID
}

// Subject is WHAT to buy, in words, copied onto the task when it opens.
//
// It is descriptive, not an invariant: a one-size product has no variant label
// and an old product may have no page. Copying it (rather than looking it up
// when the screen is drawn) follows the same reasoning as Currency below — a
// work order should say what was ordered at the time it was ordered.
type Subject struct {
	ProductName  string
	VariantLabel string // "M 8 / W 9.5 · black"; empty for a single nameless form
	VariantRef   string // the shop's own code for this exact variant, when it has one
	Source       string // the page to buy from
}

// TaskDetails is what opening a task takes (convention 10). Everything comes
// from ordering.deposit_paid plus procurement's own projections of catalog:
// procurement is TOLD what to buy.
//
// The first four fields are REQUIRED — a task without them cannot be worked.
// Subject is what a human reads and may be empty; the reflection guard in
// task_test.go pins that split.
type TaskDetails struct {
	Order    shared.ID
	Product  shared.ID
	Variant  shared.ID
	Currency shared.Currency // the shop's; what Paid must be in
	Subject  Subject         // what to buy, for the person who has to buy it
}

// PurchaseTask is the AGGREGATE ROOT of procurement.
type PurchaseTask struct {
	shared.Events

	id       TaskID
	order    shared.ID
	product  shared.ID
	variant  shared.ID
	currency shared.Currency
	subject  Subject
	status   TaskStatus
	receipt  PurchaseReceipt
	reason   string
	openedAt time.Time
	closedAt time.Time
}

// OpenTask creates the task for a deposited order.
func OpenTask(d TaskDetails, now time.Time) (*PurchaseTask, error) {
	bad := func(why string) error {
		return fmt.Errorf("open purchase task: %s: %w", why, ErrInvalidTask)
	}
	switch {
	case d.Order.IsZero():
		return nil, bad("no order")
	case d.Product.IsZero():
		return nil, bad("no product")
	case d.Variant.IsZero():
		return nil, bad("no variant")
	case d.Currency.IsZero():
		return nil, bad("no currency")
	}
	t := &PurchaseTask{
		id: NewTaskID(), order: d.Order, product: d.Product, variant: d.Variant, currency: d.Currency,
		subject: d.Subject, status: TaskOpen, openedAt: now,
	}
	t.Record(PurchaseTaskOpened{ID: t.id, Order: t.order, Product: t.product, Variant: t.variant, At: now})
	return t, nil
}

func (t *PurchaseTask) ID() TaskID {
	return t.id
}

func (t *PurchaseTask) Order() shared.ID {
	return t.order
}

func (t *PurchaseTask) Product() shared.ID {
	return t.product
}

func (t *PurchaseTask) Variant() shared.ID {
	return t.variant
}

func (t *PurchaseTask) Currency() shared.Currency {
	return t.currency
}

// Subject is what to buy, in the shop's words, as it was when the task opened.
func (t *PurchaseTask) Subject() Subject {
	return t.subject
}

func (t *PurchaseTask) Status() TaskStatus {
	return t.status
}

// Receipt is meaningful only once confirmed.
func (t *PurchaseTask) Receipt() PurchaseReceipt {
	return t.receipt
}

// Reason is meaningful only once failed.
func (t *PurchaseTask) Reason() string {
	return t.reason
}

func (t *PurchaseTask) OpenedAt() time.Time {
	return t.openedAt
}

func (t *PurchaseTask) ClosedAt() time.Time {
	return t.closedAt
}

// Confirm records a successful purchase: the shop's reference and what we
// paid. This is what flips the order past its point of no return, so the
// receipt must be real — a reference and a positive amount in the shop's
// currency, from a named operator.
func (t *PurchaseTask) Confirm(r PurchaseReceipt, now time.Time) error {
	if t.status != TaskOpen {
		return fmt.Errorf("confirm task %s: status %s: %w", t.id, t.status, ErrTaskNotOpen)
	}
	r.Reference = strings.TrimSpace(r.Reference)
	switch {
	case r.Reference == "":
		return fmt.Errorf("confirm task %s: %w", t.id, ErrEmptyReference)
	case !r.Paid.IsValid() || !r.Paid.IsPositive():
		return fmt.Errorf("confirm task %s: paid %s: %w", t.id, r.Paid, ErrInvalidTask)
	case r.Paid.Currency() != t.currency:
		return fmt.Errorf("confirm task %s: paid %s, shop bills in %s: %w", t.id, r.Paid, t.currency, ErrPaidCurrency)
	case r.PaidBy.IsZero():
		return fmt.Errorf("confirm task %s: %w", t.id, shared.ErrOperatorRequired)
	}
	t.status, t.receipt, t.closedAt = TaskConfirmed, r, now
	t.Record(PurchaseConfirmed{ID: t.id, Order: t.order, Reference: r.Reference, Paid: r.Paid, By: r.PaidBy, At: now})
	return nil
}

// Fail records that we could not buy. The reason travels to ordering, which
// decides the compensation (DDD.md §26).
func (t *PurchaseTask) Fail(reason string, now time.Time) error {
	if t.status != TaskOpen {
		return fmt.Errorf("fail task %s: status %s: %w", t.id, t.status, ErrTaskNotOpen)
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("fail task %s: %w", t.id, ErrEmptyReason)
	}
	t.status, t.reason, t.closedAt = TaskFailed, reason, now
	t.Record(PurchaseFailed{ID: t.id, Order: t.order, Reason: reason, At: now})
	return nil
}

// TaskSnapshot is PurchaseTask, flat — for the repository only.
type TaskSnapshot struct {
	ID        TaskID
	Order     shared.ID
	Product   shared.ID
	Variant   shared.ID
	Currency  shared.Currency
	Subject   Subject
	Status    TaskStatus
	Reference string
	Paid      shared.Money // zero Money unless confirmed
	PaidBy    shared.OperatorID
	Reason    string
	OpenedAt  time.Time
	ClosedAt  time.Time // zero unless closed
}

func (t *PurchaseTask) Snapshot() TaskSnapshot {
	return TaskSnapshot{
		ID: t.id, Order: t.order, Product: t.product, Variant: t.variant, Currency: t.currency,
		Subject: t.subject, Status: t.status,
		Reference: t.receipt.Reference, Paid: t.receipt.Paid, PaidBy: t.receipt.PaidBy, Reason: t.reason,
		OpenedAt: t.openedAt, ClosedAt: t.closedAt,
	}
}

func TaskFromSnapshot(s TaskSnapshot) (*PurchaseTask, error) {
	bad := func(why string) error {
		return fmt.Errorf("purchase task snapshot %s: %s: %w", s.ID, why, ErrInvalidSnapshot)
	}
	switch {
	case s.ID.IsZero():
		return nil, fmt.Errorf("purchase task snapshot: no id: %w", ErrInvalidSnapshot)
	case s.Order.IsZero(), s.Product.IsZero(), s.Variant.IsZero(), s.Currency.IsZero():
		return nil, bad("missing reference")
	case !s.Status.IsValid():
		return nil, bad(fmt.Sprintf("status %q", s.Status))
	case s.Status == TaskConfirmed && (s.Reference == "" || !s.Paid.IsValid() || s.PaidBy.IsZero() || s.ClosedAt.IsZero()):
		return nil, bad("confirmed without a receipt")
	case s.Status == TaskFailed && (s.Reason == "" || s.ClosedAt.IsZero()):
		return nil, bad("failed without a reason")
	case s.Status == TaskOpen && (s.Reference != "" || s.Paid.IsValid() || s.Reason != "" || !s.ClosedAt.IsZero()):
		return nil, bad("open task with an outcome")
	}
	return &PurchaseTask{
		id: s.ID, order: s.Order, product: s.Product, variant: s.Variant, currency: s.Currency,
		subject: s.Subject, status: s.Status,
		receipt: PurchaseReceipt{Reference: s.Reference, Paid: s.Paid, PaidBy: s.PaidBy}, reason: s.Reason,
		openedAt: s.OpenedAt, closedAt: s.ClosedAt,
	}, nil
}
