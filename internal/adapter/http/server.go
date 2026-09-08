// Package httpapi is the HTTP ADAPTER: the edge where bytes become domain
// values and domain refusals become status codes. It does exactly three
// things (DDD.md §30, tầng 1) and nothing else:
//
//  1. decode JSON and turn strings into value objects with the domain's own
//     Parse* functions — the adapter never validates business rules itself;
//  2. normalise what only the edge can know: the request's LOCALE, so that
//     "150,50" from a Vietnamese user reaches ParseMoney as "150.50";
//  3. map every error to a status and a stable code (errors.go).
//
// There is no framework. net/http since Go 1.22 routes by method and path
// pattern ("POST /products/{id}/publish") and exposes path values, which is
// the endpoints need.
//
// [PHP] Đây là Controller + FormType + Serializer + ExceptionListener của
// [PHP] Symfony gộp lại, viết tay, ~300 dòng. Không có annotation route: bảng
// [PHP] route nằm ngay trong NewHandler, đọc là thấy hết.
package httpapi

import (
	"net/http"

	"github.com/duongsy/portage/internal/app"
	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	logisticsapp "github.com/duongsy/portage/internal/app/logistics"
	orderingapp "github.com/duongsy/portage/internal/app/ordering"
	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	procurementapp "github.com/duongsy/portage/internal/app/procurement"
	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/platform/auth"
)

// Deps is every context the API fronts. One HTTP server, several bounded
// contexts: the routes of each are grouped in their own file, and a route
// only ever calls the use cases of ITS context — /quotes never touches a
// catalog repository.
type Deps struct {
	Catalog     catalogapp.Deps
	Pricing     pricingapp.Deps
	Ordering    orderingapp.Deps
	Procurement procurementapp.Deps
	Logistics   logisticsapp.Deps
	Reporting   reportingapp.Deps

	// Extractor reads a shop page with a language model (adapter/openai).
	// It is an ANTI-CORRUPTION LAYER, so it is a port with three adapters:
	// the real client, a Fake for dev and tests, and Unavailable when no
	// key is configured — which answers 503 rather than inventing data.
	Extractor catalogapp.ListingExtractor

	// Auth says who is calling. It is a port (platform/auth): the tests hand
	// it a map, production hands it a table. NewHandler panics without one —
	// a nil verifier would mean an API with no door.
	Auth auth.Verifier

	// Tokens cuts new keys — the write half of the same port. It is a separate
	// interface because one route issues and thirty only verify, and the
	// thirty must not be able to.
	Tokens auth.Issuer

	// Registry is the ADMIN half: listing and revoking keys. A third
	// interface for the same reason Issuer is a second one — only the
	// key-management routes may revoke.
	Registry auth.Registry
}

// server holds the use cases. It is unexported: the outside world sees only
// the http.Handler NewHandler returns.
type server struct {
	// catalog
	register *catalogapp.RegisterMerchantHandler
	add      *catalogapp.AddProductHandler
	publish  *catalogapp.PublishProductHandler
	define   *catalogapp.DefineCategoryHandler
	variant  *catalogapp.AddVariantHandler
	confirm  *catalogapp.ConfirmListingHandler
	measure  *catalogapp.MeasureProductHandler
	fromURL  *catalogapp.DraftFromURLHandler
	// the read side of GET /categories and GET /merchants: the two closed sets
	// a form has to offer instead of asking for a uuid
	categories catalog.CategoryRepository
	merchants  catalog.MerchantRepository
	worklist   reportingapp.ProductWorklistRepository // the two screens' queue
	// pricing
	issue  *pricingapp.IssueQuoteHandler
	accept *pricingapp.AcceptQuoteHandler
	quotes pricing.QuoteRepository // the read side of GET /quotes/{id}
	// ordering
	place   *orderingapp.PlaceOrderHandler
	deposit *orderingapp.PayDepositHandler
	balance *orderingapp.PayBalanceHandler
	cancel  *orderingapp.CancelOrderHandler
	deliver *orderingapp.DeliverOrderHandler
	orders  ordering.OrderRepository // the read side of GET /orders/{id}
	// procurement
	confirmTask_ *procurementapp.ConfirmTaskHandler
	failTask_    *procurementapp.FailTaskHandler
	tasks        procurement.TaskRepository // the buyer's list and GET /purchase-tasks/{id}
	// pricing, operator side
	defineLane_     *pricingapp.DefineLaneHandler
	rates           *pricingapp.SetExchangeRateHandler
	reconciliations pricing.ReconciliationRepository
	// logistics
	receive     *logisticsapp.ReceiveParcelHandler
	openBatch_  *logisticsapp.OpenBatchHandler
	addParcel   *logisticsapp.AddParcelHandler
	closeBatch_ *logisticsapp.CloseBatchHandler
	shipBatch_  *logisticsapp.ShipBatchHandler
	parcels     logistics.ParcelRepository
	batches     logistics.BatchRepository
	// reporting — the READ side (DDD.md §24)
	summaries reportingapp.OrderSummaryRepository

	// auth
	tokens   auth.Issuer
	registry auth.Registry
	clock    app.Clock
}

// NewHandler wires the routes to the use cases. The same Deps that the tests
// fill with memory adapters, main() fills with real ones — the handler does
// not know the difference.
func NewHandler(d Deps) http.Handler {
	if d.Auth == nil {
		panic("httpapi: NewHandler needs an auth.Verifier — an API with no door is a bug, not a configuration")
	}
	if d.Reporting.Summaries == nil {
		panic("httpapi: NewHandler needs Reporting.Summaries — the read routes would panic on the first request instead")
	}
	if d.Registry == nil {
		panic("httpapi: NewHandler needs an auth.Registry — GET /tokens would panic on the first request instead")
	}
	// A missing extractor is not a panic: it is a feature that is off. The
	// route then answers 503, which is the truth, instead of the handler
	// being nil and the process dying on the first request.
	extractor := d.Extractor
	if extractor == nil {
		extractor = unavailableExtractor{}
	}
	s := &server{
		register:        catalogapp.NewRegisterMerchantHandler(d.Catalog),
		add:             catalogapp.NewAddProductHandler(d.Catalog),
		publish:         catalogapp.NewPublishProductHandler(d.Catalog),
		define:          catalogapp.NewDefineCategoryHandler(d.Catalog),
		variant:         catalogapp.NewAddVariantHandler(d.Catalog),
		confirm:         catalogapp.NewConfirmListingHandler(d.Catalog),
		measure:         catalogapp.NewMeasureProductHandler(d.Catalog),
		fromURL:         catalogapp.NewDraftFromURLHandler(d.Catalog, extractor),
		categories:      d.Catalog.Categories,
		merchants:       d.Catalog.Merchants,
		worklist:        d.Reporting.Worklist,
		issue:           pricingapp.NewIssueQuoteHandler(d.Pricing),
		accept:          pricingapp.NewAcceptQuoteHandler(d.Pricing),
		quotes:          d.Pricing.Quotes,
		place:           orderingapp.NewPlaceOrderHandler(d.Ordering),
		deposit:         orderingapp.NewPayDepositHandler(d.Ordering),
		balance:         orderingapp.NewPayBalanceHandler(d.Ordering),
		cancel:          orderingapp.NewCancelOrderHandler(d.Ordering),
		deliver:         orderingapp.NewDeliverOrderHandler(d.Ordering),
		orders:          d.Ordering.Orders,
		confirmTask_:    procurementapp.NewConfirmTaskHandler(d.Procurement),
		failTask_:       procurementapp.NewFailTaskHandler(d.Procurement),
		tasks:           d.Procurement.Tasks,
		defineLane_:     pricingapp.NewDefineLaneHandler(d.Pricing),
		rates:           pricingapp.NewSetExchangeRateHandler(d.Pricing),
		reconciliations: d.Pricing.Reconciliations,
		receive:         logisticsapp.NewReceiveParcelHandler(d.Logistics),
		openBatch_:      logisticsapp.NewOpenBatchHandler(d.Logistics),
		addParcel:       logisticsapp.NewAddParcelHandler(d.Logistics),
		closeBatch_:     logisticsapp.NewCloseBatchHandler(d.Logistics),
		shipBatch_:      logisticsapp.NewShipBatchHandler(d.Logistics),
		parcels:         d.Logistics.Parcels,
		batches:         d.Logistics.Batches,
		summaries:       d.Reporting.Summaries,
		tokens:          d.Tokens,
		registry:        d.Registry,
		clock:           d.Catalog.Clock,
	}
	// Every route says WHO may call it on the same line as the path, so the
	// permission table and the route table cannot drift apart (auth.go):
	//
	//	requireOperator  staff only
	//	requireCustomer  the customer only — staff have no name to act under
	//	requireAny       both kinds; the handler narrows further if it must
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tokens", requireOperator(s.issueToken))
	mux.HandleFunc("GET /tokens", requireOperator(s.listTokens))
	mux.HandleFunc("DELETE /tokens/{hash}", requireOperator(s.revokeToken))
	mux.HandleFunc("POST /categories", requireOperator(s.defineCategory))
	mux.HandleFunc("POST /merchants", requireOperator(s.registerMerchant))
	mux.HandleFunc("POST /products", requireAny(s.addProduct))
	mux.HandleFunc("POST /products/from-url", requireAny(s.draftFromURL))
	mux.HandleFunc("GET /product-queue", requireOperator(s.listProductQueue))
	mux.HandleFunc("GET /me/products", requireCustomer(s.listMyProducts))
	mux.HandleFunc("GET /categories", requireAny(s.listCategories))
	mux.HandleFunc("GET /merchants", requireAny(s.listMerchants))
	mux.HandleFunc("POST /products/{id}/variants", requireOperator(s.addVariant))
	mux.HandleFunc("POST /products/{id}/confirm-listing", requireOperator(s.confirmListing))
	mux.HandleFunc("POST /products/{id}/measure", requireOperator(s.measureProduct))
	mux.HandleFunc("POST /products/{id}/publish", requireOperator(s.publishProduct))
	mux.HandleFunc("POST /quotes", requireAny(s.issueQuote))
	mux.HandleFunc("GET /quotes/{id}", requireAny(s.getQuote))
	mux.HandleFunc("POST /quotes/{id}/accept", requireCustomer(s.acceptQuote))
	// The customer places their own order: with customer_id gone from the
	// body there is no field left for staff to order in somebody else's name.
	mux.HandleFunc("POST /orders", requireCustomer(s.placeOrder))
	// Read and cancel are "owner or operator": the wrapper lets both kinds
	// in, and the ownership rule stays where it belongs — in the handler
	// (ordering.ErrNotOwner) for cancel, in the read itself for GET.
	mux.HandleFunc("GET /me/orders", requireCustomer(s.myOrders))
	mux.HandleFunc("GET /orders", requireOperator(s.listOrders))
	mux.HandleFunc("GET /orders/{id}", requireAny(s.getOrder))
	mux.HandleFunc("POST /orders/{id}/cancel", requireAny(s.cancelOrder))
	// Payments are confirmed by staff until there is a payment gateway.
	mux.HandleFunc("POST /orders/{id}/deposit", requireOperator(s.payDeposit))
	mux.HandleFunc("POST /orders/{id}/balance", requireOperator(s.payBalance))
	mux.HandleFunc("POST /orders/{id}/deliver", requireOperator(s.deliverOrder))
	mux.HandleFunc("GET /purchase-tasks", requireOperator(s.listOpenTasks))
	mux.HandleFunc("GET /purchase-tasks/{id}", requireOperator(s.getTask))
	mux.HandleFunc("POST /purchase-tasks/{id}/confirm", requireOperator(s.confirmTask))
	mux.HandleFunc("POST /purchase-tasks/{id}/fail", requireOperator(s.failTask))
	mux.HandleFunc("POST /lanes", requireOperator(s.defineLane))
	mux.HandleFunc("POST /fx", requireOperator(s.setRate))
	mux.HandleFunc("GET /reconciliations/{order}", requireOperator(s.getReconciliation))
	mux.HandleFunc("GET /parcels", requireOperator(s.listPendingParcels))
	mux.HandleFunc("GET /parcels/{id}", requireOperator(s.getParcel))
	mux.HandleFunc("POST /parcels/{id}/receive", requireOperator(s.receiveParcel))
	mux.HandleFunc("POST /batches", requireOperator(s.openBatch))
	mux.HandleFunc("GET /batches/{id}", requireOperator(s.getBatch))
	mux.HandleFunc("POST /batches/{id}/parcels", requireOperator(s.addParcelToBatch))
	mux.HandleFunc("POST /batches/{id}/close", requireOperator(s.closeBatch))
	mux.HandleFunc("POST /batches/{id}/ship", requireOperator(s.shipBatch))

	// authenticate wraps the WHOLE mux rather than each route: a route added
	// below tomorrow is behind it before anyone remembers to put it there.
	return authenticate(d.Auth)(mux)
}

// idResponse is the body of every 201: the new aggregate's id, nothing else.
// The client that wants to read the thing back asks a read model (CQRS,
// DDD.md §24) — there is no GET on a write endpoint.
type idResponse struct {
	ID string `json:"id"`
}
