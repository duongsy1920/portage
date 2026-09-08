package wire_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"testing"
	"time"

	httpapi "github.com/duongsy/portage/internal/adapter/http"
	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/adapter/postgres/pgtest"
	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
	"github.com/duongsy/portage/internal/platform/clock"
	"github.com/duongsy/portage/internal/platform/wire"
	"github.com/duongsy/portage/internal/worker"
)

var now = time.Date(2026, 9, 4, 19, 0, 0, 0, time.UTC)

// system is a wired graph behind one http.Handler plus its relay — the two
// processes of a deployment (api, worker) in one test, driven by hand.
type system struct {
	t     *testing.T
	h     http.Handler
	relay *worker.Relay
}

func newSystem(t *testing.T, g wire.Graph) *system {
	t.Helper()
	bus := worker.NewBus()
	wire.Subscribe(bus, g)
	relay := worker.New(worker.Deps{Source: g.Source, Publisher: bus, UoW: g.Catalog.UoW, Clock: g.Catalog.Clock})
	return &system{t: t, h: httpapi.NewHandler(httpapi.Deps{Catalog: g.Catalog, Pricing: g.Pricing, Ordering: g.Ordering, Procurement: g.Procurement, Logistics: g.Logistics, Reporting: g.Reporting, Extractor: g.Extractor, Auth: g.Auth, Tokens: g.Tokens, Registry: g.Registry}), relay: relay}
}

// as names a caller by kind. The golden flow goes through the same door as a
// real client would: a bearer token, not an operator-id header the caller
// writes for itself.
func (s *system) as(kind string) string {
	if kind == "customer" {
		return wire.DevCustomerToken
	}
	return wire.DevOperatorToken
}

func (s *system) call(method, path, body string, token string) *httptest.ResponseRecorder {
	s.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.h.ServeHTTP(rec, req)
	return rec
}

func (s *system) expect(rec *httptest.ResponseRecorder, status int) *httptest.ResponseRecorder {
	s.t.Helper()
	if rec.Code != status {
		s.t.Fatalf("status = %d, want %d; body %s", rec.Code, status, rec.Body)
	}
	return rec
}

func idOf(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.ID == "" {
		t.Fatalf("no id in %s", rec.Body)
	}
	return out.ID
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var e struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &e)
	return e.Error.Code
}

// wholeFlow is the system's golden path, end to end, exactly as a curl
// session runs it (docs/WALKTHROUGH.md): paste → operator steps → publish →
// relay → quote → accept. Both wirings must pass it unchanged — that is the
// promise of Ports & Adapters, and this is the test that holds it.
func wholeFlow(t *testing.T, g wire.Graph) {
	t.Helper()
	s := newSystem(t, g)
	ctx := context.Background()

	// catalog: paste a product against a SEEDED category
	merchant := idOf(t, s.expect(s.call("POST", "/merchants", `{"name":"Example Sports","site":"www.example.com","currency":"USD","sourcing":["operator","customer"]}`, s.as("operator")), 201))
	product := idOf(t, s.expect(s.call("POST", "/products", `{"name":"Air Trainer 90","merchant_id":"`+merchant+`","category":"footwear",
		"source_url":"https://www.example.com/t/air-trainer-90/abc","price":"150.00","currency":"USD"}`, s.as("customer")), 201))
	if code := errorCode(t, s.call("POST", "/products", `{"name":"Bag","merchant_id":"`+merchant+`","category":"luggage",
		"source_url":"https://www.example.com/t/bag","price":"10.00","currency":"USD"}`, s.as("customer"))); code != "category_not_found" {
		t.Fatalf("luggage must not be seeded, got %q", code)
	}

	// pricing knows nothing yet: the relay has not run
	if code := errorCode(t, s.call("POST", "/quotes", `{"product_id":"`+product+`","lane":"us_forwarder"}`, s.as("operator"))); code != "listing_not_found" {
		t.Fatalf("quote before publish = %q, want listing_not_found", code)
	}

	// operator steps, then publish
	variant := idOf(t, s.expect(s.call("POST", "/products/"+product+"/variants", `{"size":"US 9","color":"black"}`, s.as("operator")), 201))
	s.expect(s.call("POST", "/products/"+product+"/confirm-listing", "", s.as("operator")), 204)
	s.expect(s.call("POST", "/products/"+product+"/measure", `{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}`, s.as("operator")), 204)
	s.expect(s.call("POST", "/products/"+product+"/publish", "", s.as("operator")), 204)

	// still nothing: published, but not yet DELIVERED — eventual consistency has a status code
	if code := errorCode(t, s.call("POST", "/quotes", `{"product_id":"`+product+`","lane":"us_forwarder"}`, s.as("operator"))); code != "listing_not_found" {
		t.Fatalf("quote before the relay = %q, want listing_not_found", code)
	}

	// the worker's pass: seed categories + merchant + product + measured + published
	sent, err := s.relay.RunOnce(ctx)
	if err != nil || sent < 5 {
		t.Fatalf("relay: sent %d, %v", sent, err)
	}
	if again, _ := s.relay.RunOnce(ctx); again != 0 {
		t.Fatalf("second pass re-sent %d rows", again)
	}

	// pricing: quote, read, accept
	quote := idOf(t, s.expect(s.call("POST", "/quotes", `{"product_id":"`+product+`","lane":"us_forwarder"}`, s.as("operator")), 201))
	var v struct {
		Status      string `json:"status"`
		Class       string `json:"class"`
		ChargeableG int64  `json:"chargeable_g"`
		Estimated   bool   `json:"estimated"`
		Home        struct {
			Total struct {
				Amount, Currency string
			} `json:"total"`
		} `json:"home"`
	}
	if err := json.Unmarshal(s.expect(s.call("GET", "/quotes/"+quote, "", s.as("operator")), 200).Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	// footwear → branded 10.00/kg; 2500 g chargeable; 150 + 13.22 tax + 25 freight = 188.22 USD × 26000 + 500 000 fee
	if v.Status != "issued" || v.Class != "branded" || v.ChargeableG != 2500 || v.Estimated || v.Home.Total.Amount != "5393720" || v.Home.Total.Currency != "VND" {
		t.Fatalf("quote view = %+v", v)
	}
	s.expect(s.call("POST", "/quotes/"+quote+"/accept", "", s.as("customer")), 204)

	// ordering knows nothing yet: quote_accepted has not been relayed
	orderBody := `{"quote_id":"` + quote + `","variant_id":"` + variant + `"}`
	if code := errorCode(t, s.call("POST", "/orders", orderBody, s.as("customer"))); code != "quote_not_accepted" {
		t.Fatalf("order before the relay = %q, want quote_not_accepted", code)
	}
	if sent, err := s.relay.RunOnce(ctx); err != nil || sent != 2 {
		t.Fatalf("relay after quoting: sent %d, %v — want quote_issued + quote_accepted", sent, err)
	}

	// A size we never issued is not an order: ordering checks the variant
	// against catalog's variant_added, not against the caller's word for it.
	invented := `{"quote_id":"` + quote + `","variant_id":"` + shared.NewID().String() + `"}`
	if code := errorCode(t, s.call("POST", "/orders", invented, s.as("customer"))); code != "variant_unknown" {
		t.Fatalf("order with an invented variant = %q, want variant_unknown", code)
	}

	// ordering: place, deposit (exact), read, cancel with full refund
	order := idOf(t, s.expect(s.call("POST", "/orders", orderBody, s.as("customer")), 201))
	if code := errorCode(t, s.call("POST", "/orders", orderBody, s.as("customer"))); code != "quote_already_used" {
		t.Fatalf("second order on one quote = %q", code)
	}
	if code := errorCode(t, s.call("POST", "/orders/"+order+"/deposit", `{"amount":"1000","currency":"VND"}`, s.as("operator"))); code != "wrong_amount" {
		t.Fatalf("partial deposit = %q", code)
	}
	s.expect(s.call("POST", "/orders/"+order+"/deposit", `{"amount":"2696860","currency":"VND"}`, s.as("operator")), 204)
	var ov struct {
		Status string `json:"status"`
		Total  struct {
			Amount string
		} `json:"total"`
		Refund *struct {
			Amount string
		} `json:"refund"`
	}
	if err := json.Unmarshal(s.expect(s.call("GET", "/orders/"+order, "", s.as("operator")), 200).Body.Bytes(), &ov); err != nil {
		t.Fatal(err)
	}
	if ov.Status != "deposited" || ov.Total.Amount != "5393720" || ov.Refund != nil {
		t.Fatalf("order view = %+v", ov)
	}

	// procurement: the buyer's list is empty until deposit_paid is relayed; then one task, in the SHOP's currency
	var tasks []struct {
		ID       string `json:"id"`
		OrderID  string `json:"order_id"`
		Currency string `json:"currency"`
	}
	_ = json.Unmarshal(s.expect(s.call("GET", "/purchase-tasks", "", s.as("operator")), 200).Body.Bytes(), &tasks)
	if len(tasks) != 0 {
		t.Fatalf("tasks before the relay = %+v", tasks)
	}
	if sent, err := s.relay.RunOnce(ctx); err != nil || sent != 2 {
		t.Fatalf("relay after deposit: sent %d, %v — want order_placed + deposit_paid", sent, err)
	}
	_ = json.Unmarshal(s.expect(s.call("GET", "/purchase-tasks", "", s.as("operator")), 200).Body.Bytes(), &tasks)
	if len(tasks) != 1 || tasks[0].OrderID != order || tasks[0].Currency != "USD" {
		t.Fatalf("tasks after the relay = %+v", tasks)
	}

	// The READ side (DDD.md §24): the same facts, reassembled for a screen.
	// GET /orders/{id} above reads ordering's own aggregate and can only ever
	// show what ordering knows; this row already carries the product's NAME,
	// which came from catalog.
	mine := s.myOrders(t)
	if len(mine) != 1 || mine[0].Order != order {
		t.Fatalf("my orders = %+v", mine)
	}
	if mine[0].Status != "deposited" || mine[0].ProductName != "Air Trainer 90" || !mine[0].DepositPaid {
		t.Fatalf("summary after the deposit = %+v", mine[0])
	}
	if mine[0].Tracking != "none" {
		t.Fatalf("nothing is bought yet, tracking = %q", mine[0].Tracking)
	}

	// the buyer bought it: the receipt is the first ACTUAL (163.22 vs 150 + tax quoted)
	s.expect(s.call("POST", "/purchase-tasks/"+tasks[0].ID+"/confirm", `{"reference":"NK-20260905-001","paid":"163.22","currency":"USD"}`, s.as("operator")), 204)
	_ = json.Unmarshal(s.expect(s.call("GET", "/orders/"+order, "", s.as("operator")), 200).Body.Bytes(), &ov)
	if ov.Status != "deposited" {
		t.Fatalf("order must not move before purchase_confirmed is relayed, got %s", ov.Status)
	}
	if sent, err := s.relay.RunOnce(ctx); err != nil || sent != 2 {
		t.Fatalf("relay after confirm: sent %d, %v — want purchase_task_opened + purchase_confirmed", sent, err)
	}
	_ = json.Unmarshal(s.expect(s.call("GET", "/orders/"+order, "", s.as("operator")), 200).Body.Bytes(), &ov)
	if ov.Status != "purchased" {
		t.Fatalf("order after purchase_confirmed = %s, want purchased", ov.Status)
	}

	// logistics: purchase_confirmed also told the warehouse to expect a box (same event, third listener)
	var parcels []struct {
		ID        string `json:"id"`
		OrderID   string `json:"order_id"`
		Reference string `json:"reference"`
		Status    string `json:"status"`
	}
	_ = json.Unmarshal(s.expect(s.call("GET", "/parcels", "", s.as("operator")), 200).Body.Bytes(), &parcels)
	if len(parcels) != 1 || parcels[0].OrderID != order || parcels[0].Reference != "NK-20260905-001" || parcels[0].Status != "expected" {
		t.Fatalf("parcels = %+v", parcels)
	}
	s.expect(s.call("POST", "/parcels/"+parcels[0].ID+"/receive", `{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}`, s.as("operator")), 204)
	batch := idOf(t, s.expect(s.call("POST", "/batches", `{"lane":"us_forwarder"}`, s.as("operator")), 201)) // lane rule came from pricing.lane_defined at seed
	s.expect(s.call("POST", "/batches/"+batch+"/parcels", `{"parcel_id":"`+parcels[0].ID+`"}`, s.as("operator")), 204)
	s.expect(s.call("POST", "/batches/"+batch+"/close", "", s.as("operator")), 204)
	var allocs []struct {
		OrderID string `json:"order_id"`
		Freight struct {
			Amount string
		} `json:"freight"`
	}
	_ = json.Unmarshal(s.expect(s.call("POST", "/batches/"+batch+"/ship", `{"freight":"27.50","currency":"USD"}`, s.as("operator")), 200).Body.Bytes(), &allocs)
	if len(allocs) != 1 || allocs[0].OrderID != order || allocs[0].Freight.Amount != "27.50" {
		t.Fatalf("allocations = %+v", allocs)
	}
	if sent, err := s.relay.RunOnce(ctx); err != nil || sent != 6 {
		t.Fatalf("relay after shipping: sent %d, %v — want order_purchased, parcel_expected, parcel_received, batch_opened, batch_closed, batch_shipped", sent, err)
	}

	// ordering heard batch_shipped: in transit → balance → delivered
	_ = json.Unmarshal(s.expect(s.call("GET", "/orders/"+order, "", s.as("operator")), 200).Body.Bytes(), &ov)
	if ov.Status != "in_transit" {
		t.Fatalf("order after batch_shipped = %s, want in_transit", ov.Status)
	}
	s.expect(s.call("POST", "/orders/"+order+"/balance", `{"amount":"2696860","currency":"VND"}`, s.as("operator")), 204)
	s.expect(s.call("POST", "/orders/"+order+"/deliver", "", s.as("operator")), 204)
	_ = json.Unmarshal(s.expect(s.call("GET", "/orders/"+order, "", s.as("operator")), 200).Body.Bytes(), &ov)
	if ov.Status != "delivered" {
		t.Fatalf("order at the end = %s, want delivered", ov.Status)
	}

	// pricing heard all three: Quote vs Actual — quoted 163.22 + 25.00, actual 163.22 + 27.50 → we ate 2.50
	var rv struct {
		Complete bool `json:"complete"`
		Variance *struct {
			Amount, Currency string
		} `json:"variance"`
	}
	_ = json.Unmarshal(s.expect(s.call("GET", "/reconciliations/"+order, "", s.as("operator")), 200).Body.Bytes(), &rv)
	if !rv.Complete || rv.Variance == nil || rv.Variance.Amount != "-2.50" || rv.Variance.Currency != "USD" {
		t.Fatalf("reconciliation = %+v", rv)
	}
	if sent, err := s.relay.RunOnce(ctx); err != nil || sent != 3 {
		t.Fatalf("relay at the end: sent %d, %v — want order_shipped + balance_paid + order_delivered", sent, err)
	}

	// One row, five contexts: ordering's status, catalog's name, procurement's
	// shop reference, logistics' tracking — one SELECT, no JOIN.
	mine = s.myOrders(t)
	if len(mine) != 1 || mine[0].Status != "delivered" || mine[0].Tracking != "shipped" || !mine[0].BalancePaid {
		t.Fatalf("summary at the end = %+v", mine[0])
	}
	if mine[0].ShopReference != "" {
		t.Fatal("the shop's own order number must not reach the customer's screen")
	}
	staff := s.listOrders(t, "delivered")
	if len(staff) != 1 || staff[0].ShopReference != "NK-20260905-001" {
		t.Fatalf("operator queue = %+v", staff)
	}
	if none := s.listOrders(t, "awaiting_deposit"); len(none) != 0 {
		t.Fatalf("awaiting_deposit = %+v, want none left", none)
	}
}

func TestMemory_runsTheWholeFlow(t *testing.T) {
	wholeFlow(t, wire.Memory(clock.FixedAt(now)))
}

// The same flow on Postgres: connect, migrate, seed, serve, relay. Skips
// without PORTAGE_TEST_DSN like every integration test.
func TestPostgres_runsTheWholeFlow(t *testing.T) {
	pool := pgtest.Pool(t) // skip without a DSN; hold the lock; start from an empty database
	g, closeFn, err := wire.Postgres(context.Background(), clock.FixedAt(now), pgtest.DSN(t))
	if err != nil {
		t.Fatalf("wire.Postgres: %v", err)
	}
	defer closeFn()

	// wire.Postgres seeds NO tokens, deliberately: a fixed credential in a
	// real database is a back door, which is why cmd/api has
	// -bootstrap-operator-token instead. So the test cuts its own two keys,
	// with the same subjects the memory graph uses — and wholeFlow below then
	// runs unchanged on both wirings, which is the promise it exists to hold.
	issueDevTokens(t, pool)

	wholeFlow(t, g)
}

func issueDevTokens(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	repo := postgres.NewTokenRepo(pool)
	for token, p := range map[string]auth.Principal{
		wire.DevOperatorToken: auth.MustPrincipal(auth.Operator, mustID(t, wire.DevOperatorID)),
		wire.DevCustomerToken: auth.MustPrincipal(auth.Customer, mustID(t, wire.DevCustomerID)),
	} {
		if err := repo.Issue(context.Background(), token, p, "test", now); err != nil {
			t.Fatalf("issue %q: %v", token, err)
		}
	}
}

func mustID(t *testing.T, s string) shared.ID {
	t.Helper()
	id, err := shared.ParseID(s)
	if err != nil {
		t.Fatalf("%q is not a v7 uuid: %v", s, err)
	}
	return id
}

func TestPostgres_badDSNFailsAtStartUp(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, _, err := wire.Postgres(ctx, clock.System{}, "postgres://nobody:nothing@127.0.0.1:1/none?sslmode=disable"); err == nil {
		t.Fatal("a wrong DSN must fail when wiring, not on the first request")
	}
}

// The expiry sweep, through the real wiring: a quote nobody answered stops
// being a quote when its deadline passes, and the customer is told so by the
// read model rather than by a stale price.
//
// Nothing publishes "this quote got old" — the sweep is the one consumer in
// the system driven by the clock alone (worker.Sweeper, cmd/worker -sweep).
// That is exactly why the clock is a port: 48 hours pass here in one line.
func TestSweep_expiresAQuoteNobodyAnswered(t *testing.T) {
	clk := clock.FixedAt(now)
	g := wire.Memory(clk)
	s := newSystem(t, g)
	ctx := context.Background()

	merchant := idOf(t, s.expect(s.call("POST", "/merchants", `{"name":"Example Sports","site":"www.example.com","currency":"USD","sourcing":["operator","customer"]}`, s.as("operator")), 201))
	product := idOf(t, s.expect(s.call("POST", "/products", `{"name":"Air Trainer 90","merchant_id":"`+merchant+`","category":"footwear",
		"source_url":"https://www.example.com/t/air-trainer-90/abc","price":"150.00","currency":"USD"}`, s.as("operator")), 201))
	s.expect(s.call("POST", "/products/"+product+"/variants", `{"size":"US 9","color":"black"}`, s.as("operator")), 201) // this test never orders, so the id is not needed
	s.expect(s.call("POST", "/products/"+product+"/measure", `{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}`, s.as("operator")), 204)
	s.expect(s.call("POST", "/products/"+product+"/publish", "", s.as("operator")), 204)
	if _, err := s.relay.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	quote := idOf(t, s.expect(s.call("POST", "/quotes", `{"product_id":"`+product+`","lane":"us_forwarder"}`, s.as("customer")), 201))
	if _, err := s.relay.RunOnce(ctx); err != nil { // drain quote_issued so the pass at the end sees only the expiry
		t.Fatal(err)
	}
	if got := quoteStatus(t, s, quote); got != "issued" {
		t.Fatalf("fresh quote = %q", got)
	}

	sweep := pricingapp.NewExpireQuotesHandler(g.Pricing, 100)
	if n, err := sweep.Handle(ctx); err != nil || n != 0 {
		t.Fatalf("sweep inside the TTL = %d, %v; a valid quote must be left alone", n, err)
	}

	clk.Advance(49 * time.Hour) // the policy's TTL is 48h
	n, err := sweep.Handle(ctx)
	if err != nil || n != 1 {
		t.Fatalf("sweep after the deadline = %d, %v", n, err)
	}
	if got := quoteStatus(t, s, quote); got != "expired" {
		t.Fatalf("quote after the sweep = %q, want expired", got)
	}

	// The customer answering now is refused by state, not by a date check in
	// the adapter: the quote is simply not issued any more.
	if code := errorCode(t, s.call("POST", "/quotes/"+quote+"/accept", "", s.as("customer"))); code != "quote_not_issued" {
		t.Fatalf("accept after the sweep = %q", code)
	}
	// And the expiry left the system the way every other decision does.
	if sent, err := s.relay.RunOnce(ctx); err != nil || sent != 1 {
		t.Fatalf("relay after the sweep: sent %d, %v — want one pricing.quote_expired", sent, err)
	}
}

func quoteStatus(t *testing.T, s *system, quote string) string {
	t.Helper()
	var v struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(s.expect(s.call("GET", "/quotes/"+quote, "", s.as("customer")), 200).Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	return v.Status
}

// summaryRow is the read model as a client sees it (httpapi.summaryView).
type summaryRow struct {
	Order         string `json:"order_id"`
	ProductName   string `json:"product_name"`
	Status        string `json:"status"`
	Tracking      string `json:"tracking"`
	ShopReference string `json:"shop_reference"`
	DepositPaid   bool   `json:"deposit_paid"`
	BalancePaid   bool   `json:"balance_paid"`
}

func (s *system) myOrders(t *testing.T) []summaryRow {
	t.Helper()
	return rowsOf(t, s.expect(s.call("GET", "/me/orders", "", s.as("customer")), 200).Body.Bytes())
}

func (s *system) listOrders(t *testing.T, status string) []summaryRow {
	t.Helper()
	return rowsOf(t, s.expect(s.call("GET", "/orders?status="+status, "", s.as("operator")), 200).Body.Bytes())
}

func rowsOf(t *testing.T, body []byte) []summaryRow {
	t.Helper()
	var out []summaryRow
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode summaries: %v — %s", err, body)
	}
	return out
}
