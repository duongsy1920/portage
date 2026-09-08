// Package ordering is the bounded context that owns the CUSTOMER's side of a
// purchase: the order, its money (deposit now, balance later), its lifecycle
// from "placed" to "delivered", and the one rule the whole business model
// rests on — what a cancellation refunds, and when it stops refunding
// (DDD.md §26, §31).
//
// It knows a quote only as a number it was told (pricing.quote_accepted), a
// product and a variant only as ids, and a purchase only as an event from
// procurement. Nothing here imports another context (guard 7).
//
// [PHP] Bundle "Ordering": một Entity CustomerOrder với state machine viết tay
// [PHP] (không dùng Symfony Workflow) — mỗi chuyển trạng thái là một method có
// [PHP] tên nghiệp vụ, trả error thay vì ném LogicException.
package ordering

import (
	"fmt"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// OrderID identifies an order.
type OrderID struct {
	shared.ID
}

func NewOrderID() OrderID {
	return OrderID{shared.NewID()}
}

func ParseOrderID(s string) (OrderID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return OrderID{}, fmt.Errorf("order id: %w", err)
	}
	return OrderID{id}, nil
}

// OrderStatus is the lifecycle. Read it as a line with one fork:
//
//	awaiting_deposit → deposited → purchased → in_transit → delivered
//	                       └──────→ purchase_failed
//	any of the first five ───────→ cancelled          (with a Refund)
//
// "purchased" is the POINT OF NO RETURN (DDD.md §26): before it, cancelling
// refunds everything; after it, the deposit is forfeited — we are holding
// goods we bought for this customer.
type OrderStatus string

const (
	StatusAwaitingDeposit OrderStatus = "awaiting_deposit"
	StatusDeposited       OrderStatus = "deposited"
	StatusPurchased       OrderStatus = "purchased"
	StatusPurchaseFailed  OrderStatus = "purchase_failed"
	StatusInTransit       OrderStatus = "in_transit"
	StatusDelivered       OrderStatus = "delivered"
	StatusCancelled       OrderStatus = "cancelled"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusAwaitingDeposit, StatusDeposited, StatusPurchased, StatusPurchaseFailed, StatusInTransit, StatusDelivered, StatusCancelled:
		return true
	}
	return false
}

// Refund is what a cancellation gives back — and whether something was kept.
// A zero Refund (no amount) means nothing had been paid; Forfeited means the
// deposit was paid and is NOT coming back.
//
// [PHP] Một #[ORM\Embeddable] hai cột; là value object nên so sánh bằng ==.
type Refund struct {
	amount    shared.Money
	forfeited bool
}

func (r Refund) Amount() shared.Money {
	return r.amount
}

func (r Refund) Forfeited() bool {
	return r.forfeited
}

func (r Refund) IsZero() bool {
	return r == Refund{}
}

// OrderDetails is what placing an order takes (convention 10). Total and
// Deposit come from the accepted quote — ordering never recomputes a price.
type OrderDetails struct {
	Quote    shared.ID
	Product  shared.ID
	Variant  shared.ID // the customer's choice: size/colour, as catalog's id
	Customer shared.ID
	Total    shared.Money // home currency, from the quote
	Deposit  shared.Money // what is due now, from the quote
}

// CustomerOrder is the AGGREGATE ROOT of ordering.
type CustomerOrder struct {
	shared.Events

	id           OrderID
	quote        shared.ID
	product      shared.ID
	variant      shared.ID
	customer     shared.ID
	total        shared.Money
	deposit      shared.Money
	status       OrderStatus
	balancePaid  bool
	refund       Refund
	cancelReason string
	placedAt     time.Time
}

// PlaceOrder creates an order awaiting its deposit. The numbers are the
// quote's; the only arithmetic here is checking they make sense together.
func PlaceOrder(d OrderDetails, now time.Time) (*CustomerOrder, error) {
	bad := func(why string) error {
		return fmt.Errorf("place order: %s: %w", why, ErrInvalidOrder)
	}
	switch {
	case d.Quote.IsZero():
		return nil, bad("no quote")
	case d.Product.IsZero():
		return nil, bad("no product")
	case d.Variant.IsZero():
		return nil, bad("no variant")
	case d.Customer.IsZero():
		return nil, bad("no customer")
	case !d.Total.IsValid() || !d.Total.IsPositive():
		return nil, bad("total must be positive")
	case !d.Deposit.IsValid() || !d.Deposit.IsPositive():
		return nil, bad("deposit must be positive")
	case d.Deposit.Currency() != d.Total.Currency():
		return nil, fmt.Errorf("place order: deposit %s vs total %s: %w", d.Deposit, d.Total, shared.ErrCurrencyMismatch)
	case d.Deposit.Minor() > d.Total.Minor():
		return nil, bad("deposit exceeds total")
	}
	o := &CustomerOrder{
		id: NewOrderID(), quote: d.Quote, product: d.Product, variant: d.Variant, customer: d.Customer,
		total: d.Total, deposit: d.Deposit, status: StatusAwaitingDeposit, placedAt: now,
	}
	o.Record(OrderPlaced{ID: o.id, Quote: o.quote, Product: o.product, Variant: o.variant, Customer: o.customer, Total: o.total, Deposit: o.deposit, At: now})
	return o, nil
}

func (o *CustomerOrder) ID() OrderID {
	return o.id
}

func (o *CustomerOrder) Quote() shared.ID {
	return o.quote
}

func (o *CustomerOrder) Product() shared.ID {
	return o.product
}

func (o *CustomerOrder) Variant() shared.ID {
	return o.variant
}

func (o *CustomerOrder) Customer() shared.ID {
	return o.customer
}

func (o *CustomerOrder) Total() shared.Money {
	return o.total
}

func (o *CustomerOrder) Deposit() shared.Money {
	return o.deposit
}

// Balance is what remains after the deposit — due when the parcel is on its way.
func (o *CustomerOrder) Balance() shared.Money {
	b, _ := o.total.Sub(o.deposit) // same currency, deposit ≤ total: PlaceOrder guaranteed it
	return b
}

func (o *CustomerOrder) Status() OrderStatus {
	return o.status
}

func (o *CustomerOrder) BalancePaid() bool {
	return o.balancePaid
}

// Refund is meaningful only once cancelled.
func (o *CustomerOrder) Refund() Refund {
	return o.refund
}

func (o *CustomerOrder) CancelReason() string {
	return o.cancelReason
}

func (o *CustomerOrder) PlacedAt() time.Time {
	return o.placedAt
}

// PayDeposit records the customer's first payment. The amount must be EXACTLY
// the deposit: a partial deposit is not a smaller commitment, it is no
// commitment, and an overpayment is a refund nobody asked for.
func (o *CustomerOrder) PayDeposit(amount shared.Money, now time.Time) error {
	if o.status != StatusAwaitingDeposit {
		return fmt.Errorf("pay deposit on %s: %w", o.id, o.refuse(ErrNotAwaitingDeposit))
	}
	if amount != o.deposit {
		return fmt.Errorf("pay deposit on %s: got %s, due %s: %w", o.id, amount, o.deposit, ErrWrongAmount)
	}
	o.status = StatusDeposited
	o.Record(DepositPaid{ID: o.id, Quote: o.quote, Product: o.product, Variant: o.variant, Amount: amount, At: now})
	return nil
}

// ConfirmPurchase is procurement telling us the goods were bought. From here
// on the customer's deposit is at stake — the point of no return.
func (o *CustomerOrder) ConfirmPurchase(now time.Time) error {
	if o.status != StatusDeposited {
		return fmt.Errorf("confirm purchase on %s: %w", o.id, o.refuse(ErrNotDeposited))
	}
	o.status = StatusPurchased
	o.Record(OrderPurchased{ID: o.id, At: now})
	return nil
}

// FailPurchase is procurement telling us it could not buy (sold out, price
// jumped). The order waits for a decision — cancel with a full refund, or a
// new order for another variant. Nothing is decided here.
func (o *CustomerOrder) FailPurchase(reason string, now time.Time) error {
	if o.status != StatusDeposited {
		return fmt.Errorf("fail purchase on %s: %w", o.id, o.refuse(ErrNotDeposited))
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("fail purchase on %s: %w", o.id, ErrEmptyReason)
	}
	o.status = StatusPurchaseFailed
	o.Record(OrderPurchaseFailed{ID: o.id, Reason: strings.TrimSpace(reason), At: now})
	return nil
}

// Ship is logistics telling us the parcel left the US warehouse.
func (o *CustomerOrder) Ship(now time.Time) error {
	if o.status != StatusPurchased {
		return fmt.Errorf("ship %s: %w", o.id, o.refuse(ErrNotPurchased))
	}
	o.status = StatusInTransit
	o.Record(OrderShipped{ID: o.id, At: now})
	return nil
}

// PayBalance is the second half, due once the parcel is on its way.
func (o *CustomerOrder) PayBalance(amount shared.Money, now time.Time) error {
	if o.status != StatusInTransit {
		return fmt.Errorf("pay balance on %s: %w", o.id, o.refuse(ErrNotInTransit))
	}
	if o.balancePaid {
		return fmt.Errorf("pay balance on %s: already paid: %w", o.id, ErrWrongAmount)
	}
	if amount != o.Balance() {
		return fmt.Errorf("pay balance on %s: got %s, due %s: %w", o.id, amount, o.Balance(), ErrWrongAmount)
	}
	o.balancePaid = true
	o.Record(BalancePaid{ID: o.id, Amount: amount, At: now})
	return nil
}

// Deliver closes the order: in transit, balance paid, parcel at the door.
func (o *CustomerOrder) Deliver(now time.Time) error {
	if o.status != StatusInTransit {
		return fmt.Errorf("deliver %s: %w", o.id, o.refuse(ErrNotInTransit))
	}
	if !o.balancePaid {
		return fmt.Errorf("deliver %s: %w", o.id, ErrBalanceUnpaid)
	}
	o.status = StatusDelivered
	o.Record(OrderDelivered{ID: o.id, At: now})
	return nil
}

// Cancel is THE rule of the deposit model (DDD.md §26). What comes back
// depends only on where the order is:
//
//	awaiting_deposit                 nothing was paid   → refund nothing, nothing kept
//	deposited, purchase_failed       not bought yet     → refund the deposit in full
//	purchased, in_transit            goods are ours now → deposit forfeited
//	delivered                        too late           → ErrAlreadyDelivered
//	cancelled                        already            → ErrOrderCancelled
//
// It returns the Refund so the caller can act on it (pay it out) without
// reading the aggregate back.
func (o *CustomerOrder) Cancel(reason string, now time.Time) (Refund, error) {
	if strings.TrimSpace(reason) == "" {
		return Refund{}, fmt.Errorf("cancel %s: %w", o.id, ErrEmptyReason)
	}
	var refund Refund
	switch o.status {
	case StatusAwaitingDeposit:
		refund = Refund{amount: shared.Zero(o.total.Currency())}
	case StatusDeposited, StatusPurchaseFailed:
		refund = Refund{amount: o.deposit}
	case StatusPurchased, StatusInTransit:
		refund = Refund{amount: shared.Zero(o.total.Currency()), forfeited: true}
	case StatusDelivered:
		return Refund{}, fmt.Errorf("cancel %s: %w", o.id, ErrAlreadyDelivered)
	default:
		return Refund{}, fmt.Errorf("cancel %s: %w", o.id, ErrOrderCancelled)
	}
	o.status = StatusCancelled
	o.refund = refund
	o.cancelReason = strings.TrimSpace(reason)
	o.Record(OrderCancelled{ID: o.id, Reason: o.cancelReason, Refund: refund.amount, Forfeited: refund.forfeited, At: now})
	return refund, nil
}

// refuse picks the sentinel for a wrong-state call: a cancelled or delivered
// order says so instead of "not deposited" — the caller learns the real reason.
func (o *CustomerOrder) refuse(expected error) error {
	switch o.status {
	case StatusCancelled:
		return fmt.Errorf("status %s: %w", o.status, ErrOrderCancelled)
	case StatusDelivered:
		return fmt.Errorf("status %s: %w", o.status, ErrAlreadyDelivered)
	}
	return fmt.Errorf("status %s: %w", o.status, expected)
}

// OrderSnapshot is CustomerOrder, flat — for the repository only.
type OrderSnapshot struct {
	ID           OrderID
	Quote        shared.ID
	Product      shared.ID
	Variant      shared.ID
	Customer     shared.ID
	Total        shared.Money
	Deposit      shared.Money
	Status       OrderStatus
	BalancePaid  bool
	Refund       shared.Money // zero Money when not cancelled
	Forfeited    bool
	CancelReason string
	PlacedAt     time.Time
}

func (o *CustomerOrder) Snapshot() OrderSnapshot {
	return OrderSnapshot{
		ID: o.id, Quote: o.quote, Product: o.product, Variant: o.variant, Customer: o.customer,
		Total: o.total, Deposit: o.deposit, Status: o.status, BalancePaid: o.balancePaid,
		Refund: o.refund.amount, Forfeited: o.refund.forfeited, CancelReason: o.cancelReason, PlacedAt: o.placedAt,
	}
}

func OrderFromSnapshot(s OrderSnapshot) (*CustomerOrder, error) {
	bad := func(why string) error {
		return fmt.Errorf("order snapshot %s: %s: %w", s.ID, why, ErrInvalidSnapshot)
	}
	switch {
	case s.ID.IsZero():
		return nil, fmt.Errorf("order snapshot: no id: %w", ErrInvalidSnapshot)
	case s.Quote.IsZero(), s.Product.IsZero(), s.Variant.IsZero(), s.Customer.IsZero():
		return nil, bad("missing reference")
	case !s.Total.IsValid(), !s.Deposit.IsValid(), s.Deposit.Currency() != s.Total.Currency(), s.Deposit.Minor() > s.Total.Minor():
		return nil, bad("inconsistent money")
	case !s.Status.IsValid():
		return nil, bad(fmt.Sprintf("status %q", s.Status))
	case s.Status == StatusCancelled && (s.CancelReason == "" || !s.Refund.IsValid()):
		return nil, bad("cancelled without reason or refund")
	case s.Status != StatusCancelled && (s.Refund.IsValid() || s.Forfeited || s.CancelReason != ""):
		return nil, bad("refund on a live order")
	case s.BalancePaid && s.Status != StatusInTransit && s.Status != StatusDelivered && s.Status != StatusCancelled:
		return nil, bad("balance paid before transit")
	case s.Status == StatusDelivered && !s.BalancePaid:
		return nil, bad("delivered with balance unpaid")
	}
	return &CustomerOrder{
		id: s.ID, quote: s.Quote, product: s.Product, variant: s.Variant, customer: s.Customer,
		total: s.Total, deposit: s.Deposit, status: s.Status, balancePaid: s.BalancePaid,
		refund: Refund{amount: s.Refund, forfeited: s.Forfeited}, cancelReason: s.CancelReason, placedAt: s.PlacedAt,
	}, nil
}
