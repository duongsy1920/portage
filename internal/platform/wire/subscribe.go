package wire

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/adapter/eventcodec"
	logisticsapp "github.com/duongsy/portage/internal/app/logistics"
	orderingapp "github.com/duongsy/portage/internal/app/ordering"
	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	procurementapp "github.com/duongsy/portage/internal/app/procurement"
	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/worker"
)

// Subscribe is the event ROUTING TABLE of the system: which context hears
// which event. It belongs to the composition root because it is the one
// place that may know both the wire format (eventcodec) and every consumer.
//
// Today's table, read top to bottom, IS the saga of DDD.md §26:
//
//	catalog.*                    → pricing (listings, profiles) and procurement (shops, items)
//	pricing.quote_accepted       → ordering (accepted quotes)
//	ordering.deposit_paid        → procurement opens a purchase task (and asks the shop's ACL)
//	procurement.purchase_*       → ordering: past the point of no return, or compensate
//	procurement.purchase_confirmed → logistics expects a parcel; pricing records the ACTUAL goods cost
//	pricing.lane_defined         → logistics (how the lane counts weight)
//	logistics.batch_shipped      → ordering: in transit; pricing records the ACTUAL freight
//	ordering.order_placed        → pricing: which quote an order is on (Quote vs Actual)
//	mọi event ở trên             → reporting: order_summaries, bảng của màn hình
//
// One event may have several listeners (product_published has two); the Bus
// calls them in order, and a failing one holds the row back for all.
//
// [PHP] messenger.yaml `routing:` + các #[AsMessageHandler] gom lại thành một
// [PHP] bảng đọc được. Handler nào nhận message nào không còn phải grep.
func Subscribe(bus *worker.Bus, g Graph) {
	projector := pricingapp.NewProjector(g.Pricing)
	bus.Subscribe("catalog.product_published", on(projector.OnProductPublished))
	bus.Subscribe("catalog.product_measured", on(projector.OnProductMeasured))
	bus.Subscribe("catalog.product_repriced", on(projector.OnProductRepriced))
	bus.Subscribe("catalog.product_retired", on(projector.OnProductRetired))
	bus.Subscribe("catalog.category_defined", on(projector.OnCategoryDefined))

	orders := orderingapp.NewProjector(g.Ordering)
	bus.Subscribe("pricing.quote_accepted", on(orders.OnQuoteAccepted))
	bus.Subscribe("catalog.variant_added", on(orders.OnVariantAdded)) // so PlaceOrder can refuse an id we never issued

	shops := procurementapp.NewProjector(g.Procurement)
	bus.Subscribe("catalog.merchant_registered", on(shops.OnMerchantRegistered))
	bus.Subscribe("catalog.product_published", on(shops.OnProductPublished)) // second listener on the same event
	bus.Subscribe("catalog.variant_added", on(shops.OnVariantAdded))         // WHICH SIZE the buyer must ask for
	buyer := procurementapp.NewOpenTaskHandler(g.Procurement)
	bus.Subscribe("ordering.deposit_paid", on(buyer.OnDepositPaid))

	reactor := orderingapp.NewReactor(g.Ordering)
	bus.Subscribe("procurement.purchase_confirmed", on(reactor.OnPurchaseConfirmed))
	bus.Subscribe("procurement.purchase_failed", on(reactor.OnPurchaseFailed))

	warehouse := logisticsapp.NewExpectParcelHandler(g.Logistics)
	bus.Subscribe("procurement.purchase_confirmed", on(warehouse.OnPurchaseConfirmed)) // second listener
	lanes := logisticsapp.NewProjector(g.Logistics)
	bus.Subscribe("pricing.lane_defined", on(lanes.OnLaneDefined))
	bus.Subscribe("logistics.batch_shipped", on(reactor.OnBatchShipped))

	reconciler := pricingapp.NewReconciler(g.Pricing) // Quote vs Actual, three feeds
	bus.Subscribe("ordering.order_placed", on(reconciler.OnOrderPlaced))
	bus.Subscribe("procurement.purchase_confirmed", on(reconciler.OnPurchaseConfirmed)) // third listener
	bus.Subscribe("logistics.batch_shipped", on(reconciler.OnBatchShipped))

	// The READ side (DDD.md §24). It listens to ALL FIVE contexts, which is
	// the shape CQRS predicts: the write side splits into contexts that each
	// guard their own invariants, and the read side joins them back together
	// for a screen. Every handler is idempotent and order-tolerant — any of
	// them may be the event that creates the row.
	screens := reportingapp.NewProjector(g.Reporting)
	bus.Subscribe("catalog.product_published", on(screens.OnProductPublished)) // third listener on this one
	bus.Subscribe("ordering.order_placed", on(screens.OnOrderPlaced))
	bus.Subscribe("ordering.deposit_paid", on(screens.OnDepositPaid))
	bus.Subscribe("ordering.order_purchased", on(screens.OnOrderPurchased))
	bus.Subscribe("ordering.order_shipped", on(screens.OnOrderShipped))
	bus.Subscribe("ordering.balance_paid", on(screens.OnBalancePaid))
	bus.Subscribe("ordering.order_delivered", on(screens.OnOrderDelivered))
	bus.Subscribe("ordering.order_cancelled", on(screens.OnOrderCancelled))
	bus.Subscribe("procurement.purchase_confirmed", on(screens.OnPurchaseConfirmed)) // fourth listener
	bus.Subscribe("logistics.parcel_expected", on(screens.OnParcelExpected))
	bus.Subscribe("logistics.parcel_received", on(screens.OnParcelReceived))
	bus.Subscribe("logistics.batch_shipped", on(screens.OnBatchShipped))
}

// on adapts a typed consumer method to the bus: decode the payload into its
// V1 contract, hand the struct over. The decode is the only place bytes
// become a type, so a consumer never parses JSON — and a payload the codec
// does not recognise fails loudly instead of arriving half-empty.
//
// [PHP] Generic (Go 1.18+): `on[T]` sinh một closure cho từng kiểu message,
// [PHP] thay cho một argument resolver phản chiếu kiểu tham số của handler.
func on[T any](handle func(context.Context, T) error) worker.Handler {
	return func(ctx context.Context, e worker.Entry) error {
		msg, err := eventcodec.Decode(e.Name, e.Payload)
		if err != nil {
			return err
		}
		m, ok := msg.(T)
		if !ok {
			return fmt.Errorf("%s decodes to %T, handler wants %T", e.Name, msg, *new(T))
		}
		return handle(ctx, m)
	}
}
