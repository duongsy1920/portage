package orderingapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	orderingapp "github.com/duongsy/portage/internal/app/ordering"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/clock"
)

var now = time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

type world struct {
	orders   *memory.OrderRepo
	quotes   *memory.AcceptedQuoteRepo
	variants *memory.OrderingVariantRepo
	outbox   *memory.Outbox
	deps     orderingapp.Deps
}

func newWorld() *world {
	w := &world{
		orders:   memory.NewOrderRepo(),
		quotes:   memory.NewAcceptedQuoteRepo(),
		variants: memory.NewOrderingVariantRepo(),
		outbox:   memory.NewOutbox(),
	}
	w.deps = orderingapp.Deps{
		Clock:    clock.FixedAt(now),
		UoW:      memory.UnitOfWork{},
		Outbox:   w.outbox,
		Orders:   w.orders,
		Quotes:   w.quotes,
		Variants: w.variants,
	}
	return w
}

func vnd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.VND)
}

var quote, product, variant = shared.NewID(), shared.NewID(), shared.NewID()

func accepted() contracts.QuoteAcceptedV1 {
	return contracts.QuoteAcceptedV1{ID: quote.String(), Product: product.String(),
		Total: contracts.MoneyV1{Minor: 5393720, Currency: "VND"}, Deposit: contracts.MoneyV1{Minor: 2696860, Currency: "VND"}, At: now}
}

// The size the customer picked, as catalog announced it. PlaceOrder accepts
// no other id, so every test that places an order seeds this too.
func variantAdded() contracts.VariantAddedV1 {
	return contracts.VariantAddedV1{
		Product: product.String(),
		Variant: variant.String(),
		Size:    "M 8 / W 9.5",
		Color:   "black",
		At:      now,
	}
}

// seed puts in place the two projections PlaceOrder reads: the accepted quote
// (from pricing) and the variant dictionary (from catalog). Both are upserts,
// so the loop also proves at-least-once delivery is harmless.
func (w *world) seed(t *testing.T) {
	t.Helper()
	proj := orderingapp.NewProjector(w.deps)
	ctx := context.Background()
	for range 2 {
		if err := proj.OnQuoteAccepted(ctx, accepted()); err != nil {
			t.Fatal(err)
		}
		if err := proj.OnVariantAdded(ctx, variantAdded()); err != nil {
			t.Fatal(err)
		}
	}
}

func eventNames(evs []shared.Event) []string {
	out := []string{}
	for _, e := range evs {
		out = append(out, e.EventName())
	}
	return out
}

// Placing an order needs an ACCEPTED quote — known only through pricing's
// event — and a quote makes exactly one order.
func TestPlaceOrder_needsAnAcceptedQuoteAndUsesItOnce(t *testing.T) {
	w := newWorld()
	ctx := context.Background()
	place := orderingapp.NewPlaceOrderHandler(w.deps)
	cmd := orderingapp.PlaceOrder{Quote: quote, Variant: variant, Customer: shared.NewID()}

	if _, err := place.Handle(ctx, cmd); !errors.Is(err, ordering.ErrQuoteNotAccepted) {
		t.Fatalf("before the projector heard quote_accepted: %v", err)
	}
	w.seed(t)
	id, err := place.Handle(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := w.orders.ByID(ctx, id)
	if o.Total() != vnd("5393720") || o.Deposit() != vnd("2696860") || o.Product() != product || o.Status() != ordering.StatusAwaitingDeposit {
		t.Fatalf("order = %+v", o.Snapshot())
	}
	if got := eventNames(w.outbox.Drain()); len(got) != 1 || got[0] != "ordering.order_placed" {
		t.Fatalf("outbox = %v", got)
	}
	if _, err := place.Handle(ctx, cmd); !errors.Is(err, ordering.ErrQuoteAlreadyUsed) {
		t.Fatalf("same quote twice: %v", err)
	}
	if _, err := place.Handle(ctx, orderingapp.PlaceOrder{Quote: quote, Customer: cmd.Customer}); !errors.Is(err, ordering.ErrQuoteAlreadyUsed) {
		t.Fatalf("the one-order rule is checked before the domain validates: %v", err)
	}
	if w.orders.Len() != 1 {
		t.Fatalf("%d orders", w.orders.Len())
	}
}

// The variant is the one thing in the request the customer chose, so it is
// checked against catalog's own event instead of being believed: an id we
// never issued is refused, and so is a real variant of a DIFFERENT product —
// otherwise the buyer walks into the shop and asks for a size of another shoe.
func TestPlaceOrder_refusesAVariantItNeverIssued(t *testing.T) {
	w := newWorld()
	ctx := context.Background()
	w.seed(t)
	place := orderingapp.NewPlaceOrderHandler(w.deps)

	invented := shared.NewID()
	if _, err := place.Handle(ctx, orderingapp.PlaceOrder{Quote: quote, Variant: invented, Customer: shared.NewID()}); !errors.Is(err, ordering.ErrVariantUnknown) {
		t.Fatalf("an id we never issued: %v", err)
	}

	foreign := shared.NewID()
	if err := orderingapp.NewProjector(w.deps).OnVariantAdded(ctx, contracts.VariantAddedV1{
		Product: shared.NewID().String(),
		Variant: foreign.String(),
		Size:    "US 10",
		At:      now,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := place.Handle(ctx, orderingapp.PlaceOrder{Quote: quote, Variant: foreign, Customer: shared.NewID()}); !errors.Is(err, ordering.ErrVariantNotForProduct) {
		t.Fatalf("a variant of another product: %v", err)
	}
	if w.orders.Len() != 0 {
		t.Fatalf("%d orders after two refusals", w.orders.Len())
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Fatalf("a refused order left events: %v", eventNames(evs))
	}
}

// Deposit → purchase → ship → balance → deliver, through the handlers, with
// the exact-amount refusal rolling back and leaving no event.
func TestOrder_lifecycleThroughHandlers(t *testing.T) {
	w := newWorld()
	ctx := context.Background()
	w.seed(t)
	id, err := orderingapp.NewPlaceOrderHandler(w.deps).Handle(ctx, orderingapp.PlaceOrder{Quote: quote, Variant: variant, Customer: shared.NewID()})
	if err != nil {
		t.Fatal(err)
	}
	w.outbox.Drain()

	deposit := orderingapp.NewPayDepositHandler(w.deps)
	if err := deposit.Handle(ctx, orderingapp.Payment{Order: id, Amount: vnd("100")}); !errors.Is(err, ordering.ErrWrongAmount) {
		t.Fatalf("wrong deposit: %v", err)
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Fatalf("a refused payment left events: %v", eventNames(evs))
	}
	if err := deposit.Handle(ctx, orderingapp.Payment{Order: ordering.NewOrderID(), Amount: vnd("2696860")}); !errors.Is(err, ordering.ErrOrderNotFound) {
		t.Fatalf("unknown order: %v", err)
	}
	steps := []struct {
		name string
		run  func() error
	}{
		{"deposit", func() error { return deposit.Handle(ctx, orderingapp.Payment{Order: id, Amount: vnd("2696860")}) }},
		{"purchase", func() error { return orderingapp.NewConfirmPurchaseHandler(w.deps).Handle(ctx, id) }},
		{"ship", func() error { return orderingapp.NewShipOrderHandler(w.deps).Handle(ctx, id) }},
		{"balance", func() error {
			return orderingapp.NewPayBalanceHandler(w.deps).Handle(ctx, orderingapp.Payment{Order: id, Amount: vnd("2696860")})
		}},
		{"deliver", func() error { return orderingapp.NewDeliverOrderHandler(w.deps).Handle(ctx, id) }},
	}
	for _, s := range steps {
		if err := s.run(); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
	}
	want := []string{"ordering.deposit_paid", "ordering.order_purchased", "ordering.order_shipped", "ordering.balance_paid", "ordering.order_delivered"}
	got := eventNames(w.outbox.Drain())
	if len(got) != len(want) {
		t.Fatalf("events = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
	if _, err := orderingapp.NewCancelOrderHandler(w.deps).Handle(ctx, orderingapp.CancelOrder{Order: id, Reason: "late"}); !errors.Is(err, ordering.ErrAlreadyDelivered) {
		t.Fatalf("cancel after delivery: %v", err)
	}
}

// Cancel returns the refund the domain decided, and saves it.
func TestCancelOrder_returnsAndPersistsTheRefund(t *testing.T) {
	w := newWorld()
	ctx := context.Background()
	w.seed(t)
	id, _ := orderingapp.NewPlaceOrderHandler(w.deps).Handle(ctx, orderingapp.PlaceOrder{Quote: quote, Variant: variant, Customer: shared.NewID()})
	_ = orderingapp.NewPayDepositHandler(w.deps).Handle(ctx, orderingapp.Payment{Order: id, Amount: vnd("2696860")})
	_ = orderingapp.NewFailPurchaseHandler(w.deps).Handle(ctx, orderingapp.FailPurchase{Order: id, Reason: "sold out"})
	w.outbox.Drain()

	cancel := orderingapp.NewCancelOrderHandler(w.deps)
	if _, err := cancel.Handle(ctx, orderingapp.CancelOrder{Order: id}); !errors.Is(err, ordering.ErrEmptyReason) {
		t.Fatalf("no reason: %v", err)
	}
	refund, err := cancel.Handle(ctx, orderingapp.CancelOrder{Order: id, Reason: "sold out, customer declined the alternative"})
	if err != nil || refund.Amount() != vnd("2696860") || refund.Forfeited() {
		t.Fatalf("refund = %+v, %v", refund, err)
	}
	o, _ := w.orders.ByID(ctx, id)
	if o.Status() != ordering.StatusCancelled || o.Refund() != refund {
		t.Fatalf("order = %+v", o.Snapshot())
	}
	if got := eventNames(w.outbox.Drain()); len(got) != 1 || got[0] != "ordering.order_cancelled" {
		t.Fatalf("outbox = %v", got)
	}
}
