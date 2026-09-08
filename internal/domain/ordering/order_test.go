package ordering_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

var now = time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

func vnd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.VND)
}

func details() ordering.OrderDetails {
	return ordering.OrderDetails{
		Quote: shared.NewID(), Product: shared.NewID(), Variant: shared.NewID(), Customer: shared.NewID(),
		Total: vnd("5393720"), Deposit: vnd("2696860"),
	}
}

func placed(t *testing.T) *ordering.CustomerOrder {
	t.Helper()
	o, err := ordering.PlaceOrder(details(), now)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func names(evs []shared.Event) []string {
	out := make([]string, 0, len(evs))
	for _, e := range evs {
		out = append(out, e.EventName())
	}
	return out
}

func TestPlaceOrder_recordsAndValidates(t *testing.T) {
	o := placed(t)
	if o.Status() != ordering.StatusAwaitingDeposit || o.Balance() != vnd("2696860") || o.PlacedAt() != now {
		t.Fatalf("order = %s, balance %s", o.Status(), o.Balance())
	}
	if got := names(o.PullEvents()); !reflect.DeepEqual(got, []string{"ordering.order_placed"}) {
		t.Fatalf("events = %v", got)
	}

	// every field is validated (convention 10): zero one at a time
	base := details()
	for name, mutate := range map[string]func(*ordering.OrderDetails){
		"Quote":    func(d *ordering.OrderDetails) { d.Quote = shared.ID{} },
		"Product":  func(d *ordering.OrderDetails) { d.Product = shared.ID{} },
		"Variant":  func(d *ordering.OrderDetails) { d.Variant = shared.ID{} },
		"Customer": func(d *ordering.OrderDetails) { d.Customer = shared.ID{} },
		"Total":    func(d *ordering.OrderDetails) { d.Total = shared.Money{} },
		"Deposit":  func(d *ordering.OrderDetails) { d.Deposit = shared.Zero(shared.VND) },
	} {
		d := base
		mutate(&d)
		if _, err := ordering.PlaceOrder(d, now); !errors.Is(err, ordering.ErrInvalidOrder) {
			t.Errorf("zero %s: got %v, want ErrInvalidOrder", name, err)
		}
	}
	if f := reflect.TypeOf(base).NumField(); f != 6 {
		t.Fatalf("OrderDetails has %d fields; the loop above checks 6 — add the new one", f)
	}
	d := base
	d.Deposit = vnd("5393721")
	if _, err := ordering.PlaceOrder(d, now); !errors.Is(err, ordering.ErrInvalidOrder) {
		t.Errorf("deposit > total: %v", err)
	}
	d = base
	d.Deposit = shared.MustParseMoney("100.00", shared.USD)
	if _, err := ordering.PlaceOrder(d, now); !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Errorf("deposit in another currency: %v", err)
	}
}

// The happy path, end to end, with the exact-amount rule on both payments.
func TestOrder_lifecycleToDelivered(t *testing.T) {
	o := placed(t)
	o.PullEvents()

	if err := o.PayDeposit(vnd("1000000"), now); !errors.Is(err, ordering.ErrWrongAmount) {
		t.Fatalf("partial deposit: %v", err)
	}
	if err := o.ConfirmPurchase(now); !errors.Is(err, ordering.ErrNotDeposited) {
		t.Fatalf("purchase before deposit: %v", err)
	}
	if err := o.PayDeposit(vnd("2696860"), now); err != nil {
		t.Fatal(err)
	}
	if err := o.PayDeposit(vnd("2696860"), now); !errors.Is(err, ordering.ErrNotAwaitingDeposit) {
		t.Fatalf("deposit twice: %v", err)
	}
	if err := o.Ship(now); !errors.Is(err, ordering.ErrNotPurchased) {
		t.Fatalf("ship before purchase: %v", err)
	}
	if err := o.ConfirmPurchase(now); err != nil {
		t.Fatal(err)
	}
	if err := o.PayBalance(o.Balance(), now); !errors.Is(err, ordering.ErrNotInTransit) {
		t.Fatalf("balance before transit: %v", err)
	}
	if err := o.Ship(now); err != nil {
		t.Fatal(err)
	}
	if err := o.Deliver(now); !errors.Is(err, ordering.ErrBalanceUnpaid) {
		t.Fatalf("deliver with balance unpaid: %v", err)
	}
	if err := o.PayBalance(vnd("1"), now); !errors.Is(err, ordering.ErrWrongAmount) {
		t.Fatalf("wrong balance: %v", err)
	}
	if err := o.PayBalance(vnd("2696860"), now); err != nil {
		t.Fatal(err)
	}
	if err := o.PayBalance(vnd("2696860"), now); !errors.Is(err, ordering.ErrWrongAmount) {
		t.Fatalf("balance twice: %v", err)
	}
	if err := o.Deliver(now); err != nil {
		t.Fatal(err)
	}
	if err := o.Deliver(now); !errors.Is(err, ordering.ErrAlreadyDelivered) {
		t.Fatalf("deliver twice: %v", err)
	}
	if _, err := o.Cancel("changed my mind", now); !errors.Is(err, ordering.ErrAlreadyDelivered) {
		t.Fatalf("cancel after delivery: %v", err)
	}
	want := []string{"ordering.deposit_paid", "ordering.order_purchased", "ordering.order_shipped", "ordering.balance_paid", "ordering.order_delivered"}
	if got := names(o.PullEvents()); !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %v", got)
	}
}

// THE rule of the deposit model (DDD.md §26): what a cancellation refunds
// depends only on whether we have bought the goods yet.
func TestOrder_cancelRefundDependsOnThePointOfNoReturn(t *testing.T) {
	type step func(o *ordering.CustomerOrder)
	deposit := func(o *ordering.CustomerOrder) { _ = o.PayDeposit(vnd("2696860"), now) }
	purchase := func(o *ordering.CustomerOrder) { _ = o.ConfirmPurchase(now) }
	fail := func(o *ordering.CustomerOrder) { _ = o.FailPurchase("sold out in US 9", now) }
	ship := func(o *ordering.CustomerOrder) { _ = o.Ship(now) }

	cases := []struct {
		name      string
		steps     []step
		refund    string
		forfeited bool
	}{
		{"awaiting deposit: nothing paid, nothing back", nil, "0", false},
		{"deposited: full refund", []step{deposit}, "2696860", false},
		{"purchase failed: full refund (our problem)", []step{deposit, fail}, "2696860", false},
		{"purchased: deposit forfeited", []step{deposit, purchase}, "0", true},
		{"in transit: deposit forfeited", []step{deposit, purchase, ship}, "0", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o := placed(t)
			for _, s := range c.steps {
				s(o)
			}
			o.PullEvents()
			refund, err := o.Cancel("customer asked", now)
			if err != nil {
				t.Fatal(err)
			}
			if refund.Amount() != vnd(c.refund) || refund.Forfeited() != c.forfeited {
				t.Fatalf("refund = %s forfeited=%v, want %s %v", refund.Amount(), refund.Forfeited(), c.refund, c.forfeited)
			}
			if o.Status() != ordering.StatusCancelled || o.Refund() != refund || o.CancelReason() != "customer asked" {
				t.Fatalf("order after cancel = %s %+v %q", o.Status(), o.Refund(), o.CancelReason())
			}
			evs := o.PullEvents()
			if len(evs) != 1 {
				t.Fatalf("events = %v", names(evs))
			}
			ev := evs[0].(ordering.OrderCancelled)
			if ev.Refund != refund.Amount() || ev.Forfeited != c.forfeited {
				t.Fatalf("event = %+v", ev)
			}
			// a cancelled order is closed for everything
			if err := o.PayDeposit(vnd("2696860"), now); !errors.Is(err, ordering.ErrOrderCancelled) {
				t.Errorf("pay after cancel: %v", err)
			}
			if _, err := o.Cancel("again", now); !errors.Is(err, ordering.ErrOrderCancelled) {
				t.Errorf("cancel twice: %v", err)
			}
		})
	}

	o := placed(t)
	if _, err := o.Cancel("   ", now); !errors.Is(err, ordering.ErrEmptyReason) {
		t.Fatalf("cancel without a reason: %v", err)
	}
	if err := o.FailPurchase("", now); !errors.Is(err, ordering.ErrNotDeposited) {
		t.Fatalf("fail before deposit: %v", err)
	}
}

func TestOrder_snapshotRoundTripAndRejectsImpossibleShapes(t *testing.T) {
	o := placed(t)
	_ = o.PayDeposit(vnd("2696860"), now)
	_ = o.ConfirmPurchase(now)
	_ = o.Ship(now)
	_ = o.PayBalance(vnd("2696860"), now)
	back, err := ordering.OrderFromSnapshot(o.Snapshot())
	if err != nil || !reflect.DeepEqual(back.Snapshot(), o.Snapshot()) {
		t.Fatalf("round trip: %v\n got %+v\nwant %+v", err, back.Snapshot(), o.Snapshot())
	}
	refund, _ := o.Cancel("late", now)
	back, err = ordering.OrderFromSnapshot(o.Snapshot())
	if err != nil || back.Refund() != refund || back.Status() != ordering.StatusCancelled {
		t.Fatalf("cancelled round trip: %v, %+v", err, back)
	}
	if len(back.PullEvents()) != 0 {
		t.Fatal("rehydration must not record events")
	}

	good := placed(t).Snapshot()
	for name, mutate := range map[string]func(*ordering.OrderSnapshot){
		"no id":                 func(s *ordering.OrderSnapshot) { s.ID = ordering.OrderID{} },
		"no variant":            func(s *ordering.OrderSnapshot) { s.Variant = shared.ID{} },
		"deposit > total":       func(s *ordering.OrderSnapshot) { s.Deposit = vnd("9999999") },
		"unknown status":        func(s *ordering.OrderSnapshot) { s.Status = "lost" },
		"refund on live order":  func(s *ordering.OrderSnapshot) { s.Refund = vnd("1") },
		"cancelled, no reason":  func(s *ordering.OrderSnapshot) { s.Status = ordering.StatusCancelled; s.Refund = vnd("0") },
		"balance paid too soon": func(s *ordering.OrderSnapshot) { s.BalancePaid = true },
		"delivered unpaid":      func(s *ordering.OrderSnapshot) { s.Status = ordering.StatusDelivered },
	} {
		s := good
		mutate(&s)
		if _, err := ordering.OrderFromSnapshot(s); !errors.Is(err, ordering.ErrInvalidSnapshot) {
			t.Errorf("%s: got %v, want ErrInvalidSnapshot", name, err)
		}
	}
}
