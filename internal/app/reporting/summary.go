// Package reportingapp is the READ side of the system (DDD.md §24, CQRS).
//
// It is the only package here with NO domain package behind it, and that is
// the point: a read model has no invariants to protect. Nothing in it can be
// wrong in a business sense, because nothing in it DECIDES anything — it only
// remembers what the five write-side contexts announced.
//
// Two rules keep it honest:
//
//  1. It is written ONLY by events. It never reads another context's tables,
//     which is what stops "one screen" turning into a join across five
//     bounded contexts and a shared database nobody can change.
//  2. Every handler is idempotent and order-tolerant. The relay is
//     at-least-once, and a projector that needed events in order would be a
//     projector that breaks the first time a row is retried.
//
// [PHP] Một bảng "denormalised" mà bình thường mình sẽ build bằng Doctrine
// [PHP] query khổng lồ với 6 JOIN. Ở đây nó được ghi dần bằng event, nên màn
// [PHP] hình "đơn của tôi" là đúng MỘT câu SELECT không JOIN.
package reportingapp

import (
	"context"
	"errors"
	"time"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// ErrSummaryNotFound: no row for that order — it was never placed, or the
// relay has not delivered order_placed yet (eventual consistency, §25).
var ErrSummaryNotFound = errors.New("order summary not found")

// Tracking is where the customer's box is. It is NOT the order's status:
// ordering owns that, and it moves on payment and purchase, not on the
// warehouse's internal steps.
type Tracking string

const (
	TrackingNone     Tracking = "none"     // nothing bought yet
	TrackingExpected Tracking = "expected" // bought; the warehouse is waiting for it
	TrackingReceived Tracking = "received" // arrived and weighed in Denver
	TrackingShipped  Tracking = "shipped"  // on a plane in a consolidated batch
)

// OrderSummary is one row of the "my orders" screen: everything five contexts
// know about one order, flat.
//
// Every field is exported and plain. There are no methods, no constructor and
// no validation — a projection is data, and pretending otherwise would invite
// business rules into the one place that must not have any.
type OrderSummary struct {
	Order    ordering.OrderID
	Customer shared.ID
	Product  shared.ID
	Variant  shared.ID
	Quote    shared.ID

	ProductName string // filled from catalog.product_published, may be empty
	Status      ordering.OrderStatus
	Tracking    Tracking

	Total   shared.Money
	Deposit shared.Money
	Refund  shared.Money // zero unless cancelled

	DepositPaid bool
	BalancePaid bool
	Forfeited   bool

	ShopReference string // the buyer's order number at the shop

	PlacedAt    time.Time
	DeliveredAt time.Time // zero until delivered
	CancelledAt time.Time // zero unless cancelled
	UpdatedAt   time.Time
}

// OrderSummaryRepository is the read model's store. Save is an upsert: the
// projector reads a row, changes a field, writes it back.
//
// ByCustomer and ByStatus are the two screens, and they are the whole reason
// the table exists — "my orders" and "everything waiting for a deposit" are
// one index scan each here, and a five-context join anywhere else.
type OrderSummaryRepository interface {
	ByOrder(ctx context.Context, id ordering.OrderID) (OrderSummary, error) // ErrSummaryNotFound
	ByCustomer(ctx context.Context, customer shared.ID) ([]OrderSummary, error)
	ByStatus(ctx context.Context, status ordering.OrderStatus) ([]OrderSummary, error)
	ByProduct(ctx context.Context, product shared.ID) ([]OrderSummary, error) // only for the late-name backfill
	All(ctx context.Context) ([]OrderSummary, error)
	Save(ctx context.Context, s OrderSummary) error
}

// ProductNames is the read model's second, tiny store: product id → name, so
// a summary can show "Air Trainer 90" instead of a UUID.
//
// It is separate from OrderSummary because the two arrive in either order. A
// product is published long before it is ordered — usually. "Usually" is not
// a guarantee, so the projector writes the name into new rows from here, and
// backfills existing rows when a name turns up late.
type ProductNames interface {
	Name(ctx context.Context, product shared.ID) (string, bool, error)
	Save(ctx context.Context, product shared.ID, name string) error
}
