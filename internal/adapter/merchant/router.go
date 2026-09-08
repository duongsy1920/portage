package merchant

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Router picks the ACL adapter for the shop a task belongs to, and falls back
// to Manual for every shop nobody has written an adapter for.
//
// It exists because "does this shop have an API?" is a fact about ONE SHOP,
// and the alternatives are both worse:
//
//	an if in OpenTaskHandler   puts a deployment fact in a use case, and grows
//	                           a branch per shop forever
//	one ACL that knows shops   makes every shop's outage everyone's outage, and
//	                           every new shop a change to shared code
//
// Router is itself a MerchantACL, so nothing above it knows it exists: the use
// case still calls Purchase and still gets a receipt, a refusal, or an error.
// Adding a shop is one entry in the map, in wire — the composition root, which
// is the only place that is allowed to know which adapters exist at all.
//
// [PHP] Một "gateway registry" nhỏ: `MerchantGatewayInterface` với
// [PHP] `RouterGateway` chọn service theo merchant, mặc định về ManualGateway.
type Router struct {
	// Items answers "which shop is this product from" — procurement's own
	// projection, not catalog's table (the ACL must not reach into another
	// context any more than the use case may).
	Items procurement.ItemRepository

	// ByMerchant holds the shops that have an adapter. A shop that is absent
	// is not an error: it is a shop we buy from by hand.
	ByMerchant map[shared.ID]procurement.MerchantACL

	// Fallback is what an unmapped shop gets. Nil means Manual, which is the
	// answer that keeps the task on a person's list instead of losing it.
	Fallback procurement.MerchantACL
}

var _ procurement.MerchantACL = Router{}

func (r Router) Purchase(ctx context.Context, task *procurement.PurchaseTask) (procurement.PurchaseReceipt, bool, string, error) {
	acl, err := r.aclFor(ctx, task)
	if err != nil {
		return procurement.PurchaseReceipt{}, false, "", err
	}
	return acl.Purchase(ctx, task)
}

func (r Router) aclFor(ctx context.Context, task *procurement.PurchaseTask) (procurement.MerchantACL, error) {
	fallback := r.Fallback
	if fallback == nil {
		fallback = Manual{}
	}
	if r.Items == nil || len(r.ByMerchant) == 0 {
		return fallback, nil // nothing to route to; do not even look it up
	}

	item, err := r.Items.ByProduct(ctx, task.Product())
	if err != nil {
		// The projection has not caught up (the relay is eventual, §25). Not
		// knowing the shop is not a reason to fail the purchase: hand it to a
		// person, which is what would have happened anyway.
		if errors.Is(err, procurement.ErrItemNotFound) {
			return fallback, nil
		}
		return nil, fmt.Errorf("route purchase for product %s: %w", task.Product(), err)
	}
	if acl, ok := r.ByMerchant[item.Merchant]; ok {
		return acl, nil
	}
	return fallback, nil
}
