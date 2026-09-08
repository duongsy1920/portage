package eventcodec

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
)

// ErrNoDecoder: nobody reads this event yet. Decoders are written when a
// consumer needs one — a decoder nobody calls is a contract nobody tests.
var ErrNoDecoder = errors.New("no decoder for event")

// The V1 types live in internal/contracts (the Published Language); this
// file only knows which wire name decodes into which of them.
// decoders maps a wire name to the V1 struct it decodes into. Adding a
// consumer for an event means adding one line here.
var decoders = map[string]func([]byte) (any, error){
	"catalog.product_published":      into[contracts.ProductPublishedV1],
	"catalog.product_measured":       into[contracts.ProductMeasuredV1],
	"catalog.product_repriced":       into[contracts.ProductRepricedV1],
	"catalog.product_retired":        into[contracts.ProductRetiredV1],
	"catalog.category_defined":       into[contracts.CategoryDefinedV1],
	"catalog.variant_added":          into[contracts.VariantAddedV1],       // procurement: WHICH SIZE to buy
	"pricing.quote_accepted":         into[contracts.QuoteAcceptedV1],      // ordering places orders on it
	"catalog.merchant_registered":    into[contracts.MerchantRegisteredV1], // procurement: which currency a shop bills in
	"ordering.deposit_paid":          into[contracts.DepositPaidV1],        // procurement opens a purchase task
	"procurement.purchase_confirmed": into[contracts.PurchaseConfirmedV1],  // ordering: point of no return
	"procurement.purchase_failed":    into[contracts.PurchaseFailedV1],     // ordering: compensate
	"pricing.lane_defined":           into[contracts.LaneDefinedV1],        // logistics: how the lane counts weight
	"ordering.order_placed":          into[contracts.OrderPlacedV1],        // pricing: which quote → reconciliation
	"logistics.batch_shipped":        into[contracts.BatchShippedV1],       // ordering: in transit; pricing: actual freight

	// The reporting read model (internal/app/reporting) listens to the rest
	// of an order's life. No other context wants these: they exist so a
	// screen can be one SELECT instead of five (DDD.md §24).
	"ordering.order_purchased":  into[contracts.OrderPurchasedV1],
	"ordering.order_shipped":    into[contracts.OrderShippedV1],
	"ordering.balance_paid":     into[contracts.BalancePaidV1],
	"ordering.order_delivered":  into[contracts.OrderDeliveredV1],
	"ordering.order_cancelled":  into[contracts.OrderCancelledV1],
	"logistics.parcel_expected": into[contracts.ParcelExpectedV1],
	"logistics.parcel_received": into[contracts.ParcelReceivedV1],
}

// Decode turns an outbox row back into its V1 struct. The caller type-switches
// on the result; an unknown or unconsumed name is ErrNoDecoder.
//
// [PHP] `into[T]` là hàm generic — Go 1.18+. Tương đương một closure
// [PHP] `fn(string $json): T` cho từng class, sinh ra từ một template.
func Decode(name string, payload []byte) (any, error) {
	dec, ok := decoders[name]
	if !ok {
		return nil, fmt.Errorf("decode %s: %w", name, ErrNoDecoder)
	}
	msg, err := dec(payload)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return msg, nil
}

func into[T any](payload []byte) (any, error) {
	var v T
	if err := json.Unmarshal(payload, &v); err != nil {
		return nil, err
	}
	return v, nil
}
