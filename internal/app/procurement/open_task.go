package procurementapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// OpenTask: an order was deposited — buy it. Comes from ordering.deposit_paid.
type OpenTask struct {
	Order   shared.ID
	Product shared.ID
	Variant shared.ID
}

// OpenTaskHandler is the saga's second step (DDD.md §26): it opens the task,
// looks up which shop and which currency (procurement's own projections of
// catalog), then ASKS THE SHOP through the ACL:
//
//	ok            → the task is confirmed at once (an API bought it)
//	!ok + reason  → the task fails at once (the API said sold out)
//	ErrManualPurchase → the task stays OPEN for a person (today's path)
//	any other error   → returned; the relay retries the event later
//
// Idempotent: the same deposit_paid delivered twice finds the existing task
// (Tasks.ByOrder) and does nothing — at-least-once, as always.
type OpenTaskHandler struct {
	deps Deps
}

func NewOpenTaskHandler(d Deps) *OpenTaskHandler {
	mustHave("OpenTaskHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox,
		"Tasks": d.Tasks, "Shops": d.Shops, "Items": d.Items, "Variants": d.Variants, "ACL": d.ACL})
	return &OpenTaskHandler{deps: d}
}

func (h *OpenTaskHandler) Handle(ctx context.Context, cmd OpenTask) (procurement.TaskID, error) {
	now := h.deps.Clock.Now()
	var id procurement.TaskID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		if existing, err := h.deps.Tasks.ByOrder(ctx, cmd.Order); err == nil {
			id = existing.ID() // already opened: the event came twice
			return nil
		} else if !errors.Is(err, procurement.ErrTaskNotFound) {
			return err
		}
		item, err := h.deps.Items.ByProduct(ctx, cmd.Product)
		if err != nil {
			return err // product_published not projected yet: retry later
		}
		shop, err := h.deps.Shops.ByID(ctx, item.Merchant)
		if err != nil {
			return err
		}
		// What to buy, in words. A missing variant row is NOT fatal: the label
		// is what a human reads, and refusing to open the task would leave a
		// paid order with nobody assigned to buy it. An empty label shows up
		// on the buyer's screen as a product with no size, which is exactly
		// what it is.
		subject := procurement.Subject{ProductName: item.Name, Source: item.Source}
		if v, err := h.deps.Variants.ByID(ctx, cmd.Variant); err == nil {
			subject.VariantLabel, subject.VariantRef = v.Label(), v.MerchantRef
		} else if !errors.Is(err, procurement.ErrVariantNotFound) {
			return err
		}

		t, err := procurement.OpenTask(procurement.TaskDetails{
			Order: cmd.Order, Product: cmd.Product, Variant: cmd.Variant,
			Currency: shop.Currency, Subject: subject,
		}, now)
		if err != nil {
			return err
		}

		receipt, ok, reason, err := h.deps.ACL.Purchase(ctx, t)
		switch {
		case errors.Is(err, procurement.ErrManualPurchase):
			// a person will; the task stays open
		case err != nil:
			return fmt.Errorf("shop %s: %w", shop.Site, err)
		case ok:
			if err := t.Confirm(receipt, now); err != nil {
				return err
			}
		default:
			if err := t.Fail(reason, now); err != nil {
				return err
			}
		}

		if err := h.deps.Tasks.Save(ctx, t); err != nil {
			return fmt.Errorf("save purchase task: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, t.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		id = t.ID()
		return nil
	})
	return id, err
}

// OnDepositPaid adapts the wire message to the command — the subscription
// entry point wire.Subscribe uses.
func (h *OpenTaskHandler) OnDepositPaid(ctx context.Context, m contracts.DepositPaidV1) error {
	ids := make([]shared.ID, 3)
	for i, raw := range []string{m.ID, m.Product, m.Variant} {
		id, err := shared.ParseID(raw)
		if err != nil {
			return fmt.Errorf("deposit_paid %s: %w", m.ID, err)
		}
		ids[i] = id
	}
	_, err := h.Handle(ctx, OpenTask{Order: ids[0], Product: ids[1], Variant: ids[2]})
	return err
}
