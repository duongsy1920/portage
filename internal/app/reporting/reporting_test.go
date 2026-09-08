package reportingapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

var now = time.Date(2026, 9, 6, 10, 0, 0, 0, time.UTC)

type world struct {
	summaries *memory.OrderSummaryRepo
	names     *memory.ProductNameRepo
	worklist  *memory.ProductWorklistRepo
	p         *reportingapp.Projector
}

func newWorld(t *testing.T) *world {
	t.Helper()
	w := &world{
		summaries: memory.NewOrderSummaryRepo(),
		names:     memory.NewProductNameRepo(),
		worklist:  memory.NewProductWorklistRepo(),
	}
	w.p = reportingapp.NewProjector(reportingapp.Deps{
		UoW: memory.UnitOfWork{}, Summaries: w.summaries, Names: w.names, Worklist: w.worklist,
	})
	return w
}

func vnd(minor int64) contracts.MoneyV1 { return contracts.MoneyV1{Minor: minor, Currency: "VND"} }

// The whole life of an order, replayed through the projector — five contexts
// writing one row. This is the test that says the read model actually
// reassembles what the write side split apart.
func TestProjector_buildsOneRowFromFiveContexts(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	order, customer, product, variant, quote := ordering.NewOrderID(), shared.NewID(), shared.NewID(), shared.NewID(), shared.NewID()

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(w.p.OnProductPublished(ctx, contracts.ProductPublishedV1{ID: product.String(), Name: "Air Trainer 90", At: now}))
	must(w.p.OnOrderPlaced(ctx, contracts.OrderPlacedV1{
		ID: order.String(), Quote: quote.String(), Product: product.String(), Variant: variant.String(),
		Customer: customer.String(), Total: vnd(5393720), Deposit: vnd(2696860), At: now,
	}))

	s, err := w.summaries.ByOrder(ctx, order)
	if err != nil {
		t.Fatal(err)
	}
	if s.ProductName != "Air Trainer 90" || s.Status != ordering.StatusAwaitingDeposit || s.Tracking != reportingapp.TrackingNone {
		t.Fatalf("after order_placed: %+v", s)
	}
	if s.Total.String() != "5393720 VND" || s.Deposit.String() != "2696860 VND" {
		t.Fatalf("money = %s / %s", s.Total, s.Deposit)
	}

	must(w.p.OnDepositPaid(ctx, contracts.DepositPaidV1{ID: order.String(), Amount: vnd(2696860), At: now.Add(time.Minute)}))
	must(w.p.OnPurchaseConfirmed(ctx, contracts.PurchaseConfirmedV1{Order: order.String(), Reference: "NK-20260905-001", At: now.Add(2 * time.Minute)}))
	must(w.p.OnOrderPurchased(ctx, contracts.OrderPurchasedV1{ID: order.String(), At: now.Add(3 * time.Minute)}))
	must(w.p.OnParcelExpected(ctx, contracts.ParcelExpectedV1{Order: order.String(), Reference: "NK-20260905-001", At: now.Add(4 * time.Minute)}))
	must(w.p.OnParcelReceived(ctx, contracts.ParcelReceivedV1{Order: order.String(), At: now.Add(5 * time.Minute)}))
	must(w.p.OnBatchShipped(ctx, contracts.BatchShippedV1{
		Allocations: []contracts.AllocationV1{{Order: order.String(), ChargeableG: 2500}}, At: now.Add(6 * time.Minute),
	}))
	must(w.p.OnOrderShipped(ctx, contracts.OrderShippedV1{ID: order.String(), At: now.Add(7 * time.Minute)}))
	must(w.p.OnBalancePaid(ctx, contracts.BalancePaidV1{ID: order.String(), Amount: vnd(2696860), At: now.Add(8 * time.Minute)}))
	must(w.p.OnOrderDelivered(ctx, contracts.OrderDeliveredV1{ID: order.String(), At: now.Add(9 * time.Minute)}))

	s, _ = w.summaries.ByOrder(ctx, order)
	switch {
	case s.Status != ordering.StatusDelivered:
		t.Errorf("status = %s", s.Status)
	case s.Tracking != reportingapp.TrackingShipped:
		t.Errorf("tracking = %s", s.Tracking)
	case !s.DepositPaid || !s.BalancePaid:
		t.Errorf("payments = %v / %v", s.DepositPaid, s.BalancePaid)
	case s.ShopReference != "NK-20260905-001":
		t.Errorf("shop reference = %q", s.ShopReference)
	case !s.DeliveredAt.Equal(now.Add(9 * time.Minute)):
		t.Errorf("delivered at %v", s.DeliveredAt)
	}
	if w.summaries.Len() != 1 {
		t.Errorf("%d rows for one order", w.summaries.Len())
	}
}

// The relay is at-least-once, so every handler must survive the same event
// twice — and a RETRY of an old event must not drag the row backwards.
func TestProjector_isIdempotentAndDoesNotGoBackwards(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	order := ordering.NewOrderID()
	placed := contracts.OrderPlacedV1{
		ID: order.String(), Quote: shared.NewID().String(), Product: shared.NewID().String(),
		Variant: shared.NewID().String(), Customer: shared.NewID().String(),
		Total: vnd(5393720), Deposit: vnd(2696860), At: now,
	}

	for range 2 {
		if err := w.p.OnOrderPlaced(ctx, placed); err != nil {
			t.Fatal(err)
		}
	}
	if w.summaries.Len() != 1 {
		t.Fatalf("%d rows after the same event twice", w.summaries.Len())
	}

	if err := w.p.OnOrderDelivered(ctx, contracts.OrderDeliveredV1{ID: order.String(), At: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	// A stale redelivery of deposit_paid: it may set its own flag again, but
	// the row's UpdatedAt must not move back to the older timestamp.
	if err := w.p.OnDepositPaid(ctx, contracts.DepositPaidV1{ID: order.String(), Amount: vnd(1), At: now}); err != nil {
		t.Fatal(err)
	}
	s, _ := w.summaries.ByOrder(ctx, order)
	if !s.UpdatedAt.Equal(now.Add(time.Hour)) {
		t.Errorf("updated_at = %v, want the newest event's time", s.UpdatedAt)
	}
}

// Order-tolerance: nothing guarantees order_placed arrives first. An event for
// an unknown order writes a FRAME, and order_placed fills it in later.
func TestProjector_toleratesEventsBeforeOrderPlaced(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	order, product := ordering.NewOrderID(), shared.NewID()

	// Out of order on purpose: the box ships before the read model has ever
	// heard of the order.
	if err := w.p.OnBatchShipped(ctx, contracts.BatchShippedV1{
		Allocations: []contracts.AllocationV1{{Order: order.String()}}, At: now.Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	s, err := w.summaries.ByOrder(ctx, order)
	if err != nil {
		t.Fatalf("the frame must exist: %v", err)
	}
	if s.Tracking != reportingapp.TrackingShipped {
		t.Errorf("tracking = %s", s.Tracking)
	}
	if s.Status != "" {
		t.Errorf("status = %q — only order_placed may set the first status, guessing would be a lie", s.Status)
	}

	if err := w.p.OnOrderPlaced(ctx, contracts.OrderPlacedV1{
		ID: order.String(), Quote: shared.NewID().String(), Product: product.String(),
		Variant: shared.NewID().String(), Customer: shared.NewID().String(),
		Total: vnd(100), Deposit: vnd(50), At: now,
	}); err != nil {
		t.Fatal(err)
	}
	s, _ = w.summaries.ByOrder(ctx, order)
	if s.Status != ordering.StatusAwaitingDeposit || s.Tracking != reportingapp.TrackingShipped {
		t.Fatalf("after the late order_placed: %+v", s)
	}

	// And the product's name arriving LAST still reaches the row.
	if s.ProductName != "" {
		t.Fatalf("no name could be known yet, got %q", s.ProductName)
	}
	if err := w.p.OnProductPublished(ctx, contracts.ProductPublishedV1{ID: product.String(), Name: "Air Trainer 90", At: now}); err != nil {
		t.Fatal(err)
	}
	if s, _ = w.summaries.ByOrder(ctx, order); s.ProductName != "Air Trainer 90" {
		t.Errorf("the late name was not backfilled: %q", s.ProductName)
	}
}

// The two screens the table exists for.
func TestSummaries_byCustomerAndByStatus(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	mine, theirs := shared.NewID(), shared.NewID()

	place := func(customer shared.ID, at time.Time) ordering.OrderID {
		t.Helper()
		id := ordering.NewOrderID()
		if err := w.p.OnOrderPlaced(ctx, contracts.OrderPlacedV1{
			ID: id.String(), Quote: shared.NewID().String(), Product: shared.NewID().String(),
			Variant: shared.NewID().String(), Customer: customer.String(),
			Total: vnd(100), Deposit: vnd(50), At: at,
		}); err != nil {
			t.Fatal(err)
		}
		return id
	}
	older := place(mine, now)
	newer := place(mine, now.Add(time.Hour))
	place(theirs, now.Add(2*time.Hour))

	rows, err := w.summaries.ByCustomer(ctx, mine)
	if err != nil || len(rows) != 2 {
		t.Fatalf("ByCustomer = %d rows, %v — a customer must see only their own", len(rows), err)
	}
	if rows[0].Order != newer || rows[1].Order != older {
		t.Errorf("list is not newest-first: %v", rows)
	}

	if err := w.p.OnDepositPaid(ctx, contracts.DepositPaidV1{ID: newer.String(), Amount: vnd(50), At: now.Add(2 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	deposited, _ := w.summaries.ByStatus(ctx, ordering.StatusDeposited)
	if len(deposited) != 1 || deposited[0].Order != newer {
		t.Errorf("ByStatus(deposited) = %v", deposited)
	}
	waiting, _ := w.summaries.ByStatus(ctx, ordering.StatusAwaitingDeposit)
	if len(waiting) != 2 {
		t.Errorf("ByStatus(awaiting_deposit) = %d rows", len(waiting))
	}
	if all, _ := w.summaries.All(ctx); len(all) != 3 {
		t.Errorf("All = %d rows", len(all))
	}
	if _, err := w.summaries.ByOrder(ctx, ordering.NewOrderID()); !errors.Is(err, reportingapp.ErrSummaryNotFound) {
		t.Errorf("unknown order = %v", err)
	}
}

// A payload the projector cannot parse is refused BEFORE any row is touched —
// a half-written summary is worse than none, because nothing later repairs it.
func TestProjector_refusesBadPayloadsWithoutWriting(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()

	if err := w.p.OnOrderPlaced(ctx, contracts.OrderPlacedV1{ID: "not-a-uuid"}); err == nil {
		t.Error("a bad order id must be refused")
	}
	if err := w.p.OnOrderPlaced(ctx, contracts.OrderPlacedV1{
		ID: ordering.NewOrderID().String(), Quote: shared.NewID().String(), Product: shared.NewID().String(),
		Variant: shared.NewID().String(), Customer: "nonsense", Total: vnd(1), Deposit: vnd(1), At: now,
	}); err == nil {
		t.Error("a bad customer id must be refused")
	}
	if w.summaries.Len() != 0 {
		t.Errorf("%d rows written by refused events", w.summaries.Len())
	}
}
