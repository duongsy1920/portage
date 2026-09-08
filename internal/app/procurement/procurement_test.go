package procurementapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/adapter/merchant"
	procurementapp "github.com/duongsy/portage/internal/app/procurement"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/clock"
)

var now = time.Date(2026, 9, 5, 14, 0, 0, 0, time.UTC)

type world struct {
	tasks    *memory.TaskRepo
	variants *memory.VariantRepo
	outbox   *memory.Outbox
	deps     procurementapp.Deps
}

func newWorld(acl procurement.MerchantACL) *world {
	w := &world{tasks: memory.NewTaskRepo(), variants: memory.NewVariantRepo(), outbox: memory.NewOutbox()}
	w.deps = procurementapp.Deps{
		Clock:    clock.FixedAt(now),
		UoW:      memory.UnitOfWork{},
		Outbox:   w.outbox,
		Tasks:    w.tasks,
		Shops:    memory.NewShopRepo(),
		Items:    memory.NewItemRepo(),
		Variants: w.variants,
		ACL:      acl,
	}
	return w
}

var merchantID, productID, variantID = shared.NewID(), shared.NewID(), shared.NewID()

const sourceURL = "https://www.example.com/t/air-trainer-90/abc"

// What catalog says when the operator types a size in. procurement keeps it
// so the buyer's screen can say "M 8 / W 9.5" instead of a uuid.
func variantAdded() contracts.VariantAddedV1 {
	return contracts.VariantAddedV1{
		Product:     productID.String(),
		Variant:     variantID.String(),
		Size:        "M 8 / W 9.5",
		Color:       "black",
		MerchantRef: "EX-AT90-8-BLK",
		At:          now,
	}
}

func (w *world) seedCatalog(t *testing.T) {
	t.Helper()
	p := procurementapp.NewProjector(w.deps)
	ctx := context.Background()
	for range 2 { // idempotent
		if err := p.OnMerchantRegistered(ctx, contracts.MerchantRegisteredV1{ID: merchantID.String(), Name: "Example Sports", Site: "www.example.com", Currency: "USD", At: now}); err != nil {
			t.Fatal(err)
		}
		if err := p.OnProductPublished(ctx, contracts.ProductPublishedV1{ID: productID.String(), Merchant: merchantID.String(), Category: "footwear", Name: "Air Trainer 90", Source: sourceURL,
			Price: contracts.MoneyV1{Minor: 15000, Currency: "USD"}, Parcel: contracts.ParcelV1{WeightG: 1250, LengthMM: 340, WidthMM: 230, HeightMM: 130}, At: now}); err != nil {
			t.Fatal(err)
		}
		if err := p.OnVariantAdded(ctx, variantAdded()); err != nil {
			t.Fatal(err)
		}
	}
}

func deposit(order shared.ID) contracts.DepositPaidV1 {
	return contracts.DepositPaidV1{ID: order.String(), Quote: shared.NewID().String(), Product: productID.String(), Variant: variantID.String(),
		Amount: contracts.MoneyV1{Minor: 2696860, Currency: "VND"}, At: now}
}

func names(evs []shared.Event) []string {
	out := []string{}
	for _, e := range evs {
		out = append(out, e.EventName())
	}
	return out
}

// The manual path: the shop cannot be bought from by API, so the task stays
// open on the buyer's list; the same event twice opens nothing new.
func TestOpenTask_manualShopLeavesTheTaskOpen(t *testing.T) {
	w := newWorld(merchant.Manual{})
	ctx := context.Background()
	open := procurementapp.NewOpenTaskHandler(w.deps)
	order := shared.NewID()

	if err := open.OnDepositPaid(ctx, deposit(order)); !errors.Is(err, procurement.ErrItemNotFound) {
		t.Fatalf("before catalog is projected the event must be retried later: %v", err)
	}
	w.seedCatalog(t)
	for range 2 {
		if err := open.OnDepositPaid(ctx, deposit(order)); err != nil {
			t.Fatal(err)
		}
	}
	tasks, _ := w.tasks.Open(ctx)
	if len(tasks) != 1 || tasks[0].Order() != order || tasks[0].Currency() != shared.USD || tasks[0].Status() != procurement.TaskOpen {
		t.Fatalf("open tasks = %d %+v", len(tasks), tasks)
	}
	if got := names(w.outbox.Drain()); len(got) != 1 || got[0] != "procurement.purchase_task_opened" {
		t.Fatalf("outbox = %v", got)
	}
	// The buyer's screen: a name, a size, the shop's own code and the page to
	// buy from — frozen at open time, so a later rename cannot change what
	// this task said to buy.
	if s := tasks[0].Subject(); s.ProductName != "Air Trainer 90" || s.VariantLabel != "M 8 / W 9.5 · black" || s.VariantRef != "EX-AT90-8-BLK" || s.Source != sourceURL {
		t.Fatalf("subject = %+v", s)
	}

	// the buyer closes it
	id := tasks[0].ID()
	confirm := procurementapp.NewConfirmTaskHandler(w.deps)
	if err := confirm.Handle(ctx, procurementapp.ConfirmTask{Task: id, Receipt: procurement.PurchaseReceipt{Reference: "NK-1", Paid: shared.MustParseMoney("4000000", shared.VND), PaidBy: shared.NewOperatorID()}}); !errors.Is(err, procurement.ErrPaidCurrency) {
		t.Fatalf("paid in the wrong currency: %v", err)
	}
	if err := confirm.Handle(ctx, procurementapp.ConfirmTask{Task: id, Receipt: procurement.PurchaseReceipt{Reference: "NK-1", Paid: shared.MustParseMoney("163.22", shared.USD), PaidBy: shared.NewOperatorID()}}); err != nil {
		t.Fatal(err)
	}
	if got := names(w.outbox.Drain()); len(got) != 1 || got[0] != "procurement.purchase_confirmed" {
		t.Fatalf("outbox = %v", got)
	}
	if err := procurementapp.NewFailTaskHandler(w.deps).Handle(ctx, procurementapp.FailTask{Task: id, Reason: "late"}); !errors.Is(err, procurement.ErrTaskNotOpen) {
		t.Fatalf("fail after confirm: %v", err)
	}
	if left, _ := w.tasks.Open(ctx); len(left) != 0 {
		t.Fatalf("still open: %d", len(left))
	}
}

// apiShop is a fake ACL: the shop answers by itself. It proves the use case
// handles all three answers of the port without a person.
type apiShop struct {
	ok     bool
	reason string
	err    error
}

func (s apiShop) Purchase(ctx context.Context, task *procurement.PurchaseTask) (procurement.PurchaseReceipt, bool, string, error) {
	if s.err != nil {
		return procurement.PurchaseReceipt{}, false, "", s.err
	}
	if !s.ok {
		return procurement.PurchaseReceipt{}, false, s.reason, nil
	}
	return procurement.PurchaseReceipt{Reference: "API-1", Paid: shared.MustParseMoney("163.22", task.Currency()), PaidBy: shared.NewOperatorID()}, true, "", nil
}

func TestOpenTask_anAPIShopClosesTheTaskAtOnce(t *testing.T) {
	ctx := context.Background()
	for name, c := range map[string]struct {
		acl    apiShop
		status procurement.TaskStatus
		events []string
	}{
		"bought":   {apiShop{ok: true}, procurement.TaskConfirmed, []string{"procurement.purchase_task_opened", "procurement.purchase_confirmed"}},
		"sold out": {apiShop{ok: false, reason: "sold out"}, procurement.TaskFailed, []string{"procurement.purchase_task_opened", "procurement.purchase_failed"}},
	} {
		t.Run(name, func(t *testing.T) {
			w := newWorld(c.acl)
			w.seedCatalog(t)
			order := shared.NewID()
			if err := procurementapp.NewOpenTaskHandler(w.deps).OnDepositPaid(ctx, deposit(order)); err != nil {
				t.Fatal(err)
			}
			task, _ := w.tasks.ByOrder(ctx, order)
			if task.Status() != c.status {
				t.Fatalf("status = %s", task.Status())
			}
			got := names(w.outbox.Drain())
			if len(got) != 2 || got[0] != c.events[0] || got[1] != c.events[1] {
				t.Fatalf("events = %v", got)
			}
		})
	}

	// the shop's API is down: nothing is saved, the relay will retry
	w := newWorld(apiShop{err: errors.New("502 from shop")})
	w.seedCatalog(t)
	if err := procurementapp.NewOpenTaskHandler(w.deps).OnDepositPaid(ctx, deposit(shared.NewID())); err == nil {
		t.Fatal("an unreachable shop must surface as an error")
	}
	if open, _ := w.tasks.Open(ctx); len(open) != 0 || len(w.outbox.Drain()) != 0 {
		t.Fatal("a failed attempt must leave no task and no event")
	}
}
