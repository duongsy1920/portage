// Package merchant holds the Anti-Corruption Layer adapters (DDD.md §22): one
// per way of buying at a shop, all behind procurement.MerchantACL.
//
// Today there is one, Manual: it cannot buy. It exists so the use case has a
// port to call NOW, so the flow "task opened → try to buy → fall back to a
// person" is real code with a real test, and so the day a shop exposes an
// API the change is one new file here and one line in wire — not a new
// branch in the use case.
//
// [PHP] `ManualMerchantGateway implements MerchantGatewayInterface` — cái
// [PHP] implementation "null object" bạn hay đăng ký trong services.yaml trước
// [PHP] khi có API thật.
package merchant

import (
	"context"

	"github.com/duongsy/portage/internal/domain/procurement"
)

// Manual is the ACL for every shop we buy from by hand.
type Manual struct{}

var _ procurement.MerchantACL = Manual{}

// Purchase always answers ErrManualPurchase: a person has to do it. The task
// stays open on the buyer's list and is closed through the API.
func (Manual) Purchase(ctx context.Context, task *procurement.PurchaseTask) (procurement.PurchaseReceipt, bool, string, error) {
	return procurement.PurchaseReceipt{}, false, "", procurement.ErrManualPurchase
}
