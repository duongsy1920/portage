package procurement_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

var now = time.Date(2026, 9, 5, 14, 0, 0, 0, time.UTC)

func usd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.USD)
}

func details() procurement.TaskDetails {
	return procurement.TaskDetails{Order: shared.NewID(), Product: shared.NewID(), Variant: shared.NewID(), Currency: shared.USD}
}

func open(t *testing.T) *procurement.PurchaseTask {
	t.Helper()
	task, err := procurement.OpenTask(details(), now)
	if err != nil {
		t.Fatal(err)
	}
	task.PullEvents()
	return task
}

func TestOpenTask_validatesEveryField(t *testing.T) {
	task, err := procurement.OpenTask(details(), now)
	if err != nil || task.Status() != procurement.TaskOpen || task.OpenedAt() != now {
		t.Fatalf("task = %+v, %v", task, err)
	}
	if evs := task.PullEvents(); len(evs) != 1 || evs[0].EventName() != "procurement.purchase_task_opened" {
		t.Fatalf("events = %v", evs)
	}
	base := details()
	for name, mutate := range map[string]func(*procurement.TaskDetails){
		"Order":    func(d *procurement.TaskDetails) { d.Order = shared.ID{} },
		"Product":  func(d *procurement.TaskDetails) { d.Product = shared.ID{} },
		"Variant":  func(d *procurement.TaskDetails) { d.Variant = shared.ID{} },
		"Currency": func(d *procurement.TaskDetails) { d.Currency = shared.Currency{} },
	} {
		d := base
		mutate(&d)
		if _, err := procurement.OpenTask(d, now); !errors.Is(err, procurement.ErrInvalidTask) {
			t.Errorf("zero %s: %v", name, err)
		}
	}
	// Subject is the 5th field and the ONLY optional one: it is what to buy in
	// the shop's words, and a task with an empty subject is still a real task
	// (the variant row may not have been relayed yet). Add a required field
	// and this guard fails until the loop above covers it.
	if f := reflect.TypeOf(base).NumField(); f != 5 {
		t.Fatalf("TaskDetails has %d fields; the loop checks 4 required + Subject", f)
	}
	withoutSubject := details()
	withoutSubject.Subject = procurement.Subject{}
	if _, err := procurement.OpenTask(withoutSubject, now); err != nil {
		t.Errorf("an empty Subject must still open a task: %v", err)
	}
}

func TestTask_confirmNeedsARealReceipt(t *testing.T) {
	task := open(t)
	op := shared.NewOperatorID()
	good := procurement.PurchaseReceipt{Reference: " NK-123456 ", Paid: usd("163.22"), PaidBy: op}

	for name, r := range map[string]procurement.PurchaseReceipt{
		"no reference":   {Paid: usd("163.22"), PaidBy: op},
		"zero paid":      {Reference: "NK-1", Paid: shared.Zero(shared.USD), PaidBy: op},
		"wrong currency": {Reference: "NK-1", Paid: shared.MustParseMoney("4000000", shared.VND), PaidBy: op},
		"no operator":    {Reference: "NK-1", Paid: usd("163.22")},
	} {
		if err := task.Confirm(r, now); err == nil {
			t.Errorf("%s: accepted", name)
		}
		if task.Status() != procurement.TaskOpen || len(task.PullEvents()) != 0 {
			t.Fatalf("%s: a refused confirm changed the task", name)
		}
	}
	if err := task.Confirm(procurement.PurchaseReceipt{Paid: usd("1"), PaidBy: op}, now); !errors.Is(err, procurement.ErrEmptyReference) {
		t.Errorf("no reference: %v", err)
	}
	if err := task.Confirm(procurement.PurchaseReceipt{Reference: "x", Paid: shared.MustParseMoney("1", shared.VND), PaidBy: op}, now); !errors.Is(err, procurement.ErrPaidCurrency) {
		t.Errorf("wrong currency: %v", err)
	}
	if err := task.Confirm(procurement.PurchaseReceipt{Reference: "x", Paid: usd("1")}, now); !errors.Is(err, shared.ErrOperatorRequired) {
		t.Errorf("no operator: %v", err)
	}

	later := now.Add(time.Hour)
	if err := task.Confirm(good, later); err != nil {
		t.Fatal(err)
	}
	if task.Status() != procurement.TaskConfirmed || task.Receipt().Reference != "NK-123456" || task.Receipt().Paid != usd("163.22") || task.ClosedAt() != later {
		t.Fatalf("task = %+v", task.Snapshot())
	}
	evs := task.PullEvents()
	if len(evs) != 1 {
		t.Fatalf("events = %v", evs)
	}
	ev := evs[0].(procurement.PurchaseConfirmed)
	if ev.Order != task.Order() || ev.Reference != "NK-123456" || ev.Paid != usd("163.22") || ev.By != op {
		t.Fatalf("event = %+v", ev)
	}
	if err := task.Confirm(good, later); !errors.Is(err, procurement.ErrTaskNotOpen) {
		t.Errorf("confirm twice: %v", err)
	}
	if err := task.Fail("too late", later); !errors.Is(err, procurement.ErrTaskNotOpen) {
		t.Errorf("fail after confirm: %v", err)
	}
}

func TestTask_failNeedsAReason(t *testing.T) {
	task := open(t)
	if err := task.Fail("  ", now); !errors.Is(err, procurement.ErrEmptyReason) {
		t.Fatalf("empty reason: %v", err)
	}
	if err := task.Fail(" sold out in US 9 ", now); err != nil {
		t.Fatal(err)
	}
	if task.Status() != procurement.TaskFailed || task.Reason() != "sold out in US 9" {
		t.Fatalf("task = %+v", task.Snapshot())
	}
	ev := task.PullEvents()[0].(procurement.PurchaseFailed)
	if ev.Order != task.Order() || ev.Reason != "sold out in US 9" {
		t.Fatalf("event = %+v", ev)
	}
}

func TestTask_snapshotRoundTripAndRejectsImpossibleShapes(t *testing.T) {
	task := open(t)
	if back, err := procurement.TaskFromSnapshot(task.Snapshot()); err != nil || !reflect.DeepEqual(back.Snapshot(), task.Snapshot()) {
		t.Fatalf("open round trip: %v", err)
	}
	_ = task.Confirm(procurement.PurchaseReceipt{Reference: "NK-1", Paid: usd("163.22"), PaidBy: shared.NewOperatorID()}, now)
	back, err := procurement.TaskFromSnapshot(task.Snapshot())
	if err != nil || !reflect.DeepEqual(back.Snapshot(), task.Snapshot()) || len(back.PullEvents()) != 0 {
		t.Fatalf("confirmed round trip: %v", err)
	}

	good := open(t).Snapshot()
	for name, mutate := range map[string]func(*procurement.TaskSnapshot){
		"no id":                   func(s *procurement.TaskSnapshot) { s.ID = procurement.TaskID{} },
		"no currency":             func(s *procurement.TaskSnapshot) { s.Currency = shared.Currency{} },
		"unknown status":          func(s *procurement.TaskSnapshot) { s.Status = "lost" },
		"confirmed, no receipt":   func(s *procurement.TaskSnapshot) { s.Status = procurement.TaskConfirmed },
		"failed, no reason":       func(s *procurement.TaskSnapshot) { s.Status = procurement.TaskFailed; s.ClosedAt = now },
		"open with a reason":      func(s *procurement.TaskSnapshot) { s.Reason = "sold out" },
		"open but already closed": func(s *procurement.TaskSnapshot) { s.ClosedAt = now },
	} {
		s := good
		mutate(&s)
		if _, err := procurement.TaskFromSnapshot(s); !errors.Is(err, procurement.ErrInvalidSnapshot) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}
