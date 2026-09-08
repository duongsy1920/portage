package catalogapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
)

// RegisterMerchantHandler is the use case behind "operator adds a shop".
// The command IS catalog.MerchantDetails: wrapping it in a second struct with
// the same fields would be a copy with a different name.
type RegisterMerchantHandler struct {
	deps Deps
}

func NewRegisterMerchantHandler(d Deps) *RegisterMerchantHandler {
	mustHave("RegisterMerchantHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Merchants": d.Merchants, "Outbox": d.Outbox,
	})
	return &RegisterMerchantHandler{deps: d}
}

// Handle runs the request lifecycle of DDD.md §30. Read the ORDER: the domain
// decides, the aggregate is saved, and only then are its events pulled into
// the outbox — inside the same transaction. A save that fails publishes
// nothing, so no worker ever acts on a merchant that does not exist.
//
// [PHP] Tương đương __invoke(RegisterMerchant $cmd) trong một MessageHandler:
// [PHP] $em->wrapInTransaction(function () { ... persist ... dispatch events })
func (h *RegisterMerchantHandler) Handle(ctx context.Context, d catalog.MerchantDetails) (catalog.MerchantID, error) {
	now := h.deps.Clock.Now() // the clock is read HERE, once, and passed down

	var id catalog.MerchantID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		m, err := catalog.RegisterMerchant(d, now) // ① the domain decides
		if err != nil {
			return err // a business refusal, returned untouched
		}
		if err := h.deps.Merchants.Save(ctx, m); err != nil { // ② save first
			return fmt.Errorf("save merchant: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, m.PullEvents()); err != nil { // ③ then pull
			return fmt.Errorf("outbox: %w", err)
		}
		id = m.ID()
		return nil
	})
	return id, err
}
