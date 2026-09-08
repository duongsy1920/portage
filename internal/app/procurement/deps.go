// Package procurementapp holds the use cases of procurement: open a purchase
// task when an order is deposited (and try the shop's ACL at once), and let
// the buyer close it — confirmed with a receipt, or failed with a reason.
//
// [PHP] MessageHandlers của bundle Procurement; OpenTask là handler nhận
// [PHP] DepositPaid từ Ordering qua Messenger.
package procurementapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/app"
	"github.com/duongsy/portage/internal/domain/procurement"
)

// Deps is everything the procurement use cases are wired to (convention 10).
// ACL is the port to the shops — Manual today.
type Deps struct {
	Clock  app.Clock
	UoW    app.UnitOfWork
	Outbox app.Outbox
	Tasks  procurement.TaskRepository
	Shops  procurement.ShopRepository
	Items  procurement.ItemRepository
	// Variants: which size each variant id means, so the buyer's screen can
	// say "M 8 / W 9.5" where the order only says a uuid.
	Variants procurement.VariantRepository
	ACL      procurement.MerchantACL
}

func mustHave(handler string, deps map[string]any) {
	app.MustHave("procurementapp: "+handler, deps)
}

// mutate: load → act → save → outbox, once (see orderingapp.Deps.mutate).
func (d Deps) mutate(ctx context.Context, id procurement.TaskID, fn func(t *procurement.PurchaseTask) error) error {
	return d.UoW.InTx(ctx, func(ctx context.Context) error {
		t, err := d.Tasks.ByID(ctx, id)
		if err != nil {
			return err
		}
		if err := fn(t); err != nil {
			return err
		}
		if err := d.Tasks.Save(ctx, t); err != nil {
			return fmt.Errorf("save purchase task: %w", err)
		}
		if err := d.Outbox.Append(ctx, t.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}
