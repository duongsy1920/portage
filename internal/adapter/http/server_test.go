package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpapi "github.com/duongsy/portage/internal/adapter/http"
	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/adapter/merchant"
	"github.com/duongsy/portage/internal/adapter/openai"
	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	logisticsapp "github.com/duongsy/portage/internal/app/logistics"
	orderingapp "github.com/duongsy/portage/internal/app/ordering"
	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	procurementapp "github.com/duongsy/portage/internal/app/procurement"
	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
	"github.com/duongsy/portage/internal/platform/clock"
)

var now = time.Date(2026, 9, 4, 16, 0, 0, 0, time.UTC)

// api is the whole system behind one http.Handler — in memory, no port opened.
type api struct {
	h          http.Handler
	operatorID shared.OperatorID
	customerID shared.ID
	auth       *auth.Static
	clock      *clock.Fixed
	merchants  *memory.MerchantRepo
	categories *memory.CategoryRepo
	products   *memory.ProductRepo
	outbox     *memory.Outbox
	// pricing
	lanes    *memory.LaneRepo
	quotes   *memory.QuoteRepo
	listings *memory.ListingRepo
	profiles *memory.ProfileRepo
	rates    *memory.ExchangeRates
	// ordering
	orders         *memory.OrderRepo
	acceptedQuotes *memory.AcceptedQuoteRepo
	variants       *memory.OrderingVariantRepo
	// procurement
	tasks *memory.TaskRepo
	// logistics + reconciliation
	parcels         *memory.ParcelRepo
	batches         *memory.BatchRepo
	laneRules       *memory.LaneRuleRepo
	reconciliations *memory.ReconciliationRepo
	// reporting (read model)
	summaries *memory.OrderSummaryRepo
	names     *memory.ProductNameRepo
	worklist  *memory.ProductWorklistRepo
}

// newAPI wires the whole API onto in-memory adapters. newAPIWith takes the
// Deps apart first, for the handful of tests that need a different adapter —
// an extractor that fails, or none at all.
func newAPI() *api {
	return newAPIWith(nil)
}

func newAPIWith(tweak func(*httpapi.Deps)) *api {
	a := &api{
		clock:           clock.FixedAt(now),
		operatorID:      shared.NewOperatorID(),
		customerID:      shared.NewID(),
		merchants:       memory.NewMerchantRepo(),
		categories:      memory.NewCategoryRepo(),
		products:        memory.NewProductRepo(),
		outbox:          memory.NewOutbox(),
		lanes:           memory.NewLaneRepo(),
		quotes:          memory.NewQuoteRepo(),
		listings:        memory.NewListingRepo(),
		profiles:        memory.NewProfileRepo(),
		rates:           memory.NewExchangeRates(),
		orders:          memory.NewOrderRepo(),
		acceptedQuotes:  memory.NewAcceptedQuoteRepo(),
		variants:        memory.NewOrderingVariantRepo(),
		tasks:           memory.NewTaskRepo(),
		parcels:         memory.NewParcelRepo(),
		batches:         memory.NewBatchRepo(),
		laneRules:       memory.NewLaneRuleRepo(),
		reconciliations: memory.NewReconciliationRepo(),
		summaries:       memory.NewOrderSummaryRepo(),
		names:           memory.NewProductNameRepo(),
		worklist:        memory.NewProductWorklistRepo(),
	}
	a.auth = auth.NewStatic(map[string]auth.Principal{
		devOperatorToken: auth.MustPrincipal(auth.Operator, a.operatorID.ID),
		devCustomerToken: auth.MustPrincipal(auth.Customer, a.customerID),
	})
	margin, _ := pricing.NewMarginPolicy(shared.MustParsePercent("10"), shared.MustParseMoney("500000", shared.VND))
	policy, _ := pricing.NewQuotePolicy(pricing.QuotePolicyDetails{
		SalesTax: shared.MustParsePercent("8.81"), Margin: margin, Deposit: shared.MustParsePercent("50"), TTL: 48 * time.Hour,
	})
	d := httpapi.Deps{
		Catalog: catalogapp.Deps{
			Clock: a.clock, UoW: memory.UnitOfWork{},
			Merchants: a.merchants, Categories: a.categories, Products: a.products, Outbox: a.outbox,
		},
		Pricing: pricingapp.Deps{
			Clock: a.clock, UoW: memory.UnitOfWork{}, Outbox: a.outbox,
			Lanes: a.lanes, Quotes: a.quotes, Listings: a.listings, Profiles: a.profiles, Rates: a.rates, Reconciliations: a.reconciliations,
			Policy: policy,
		},
		Ordering: orderingapp.Deps{
			Clock: a.clock, UoW: memory.UnitOfWork{}, Outbox: a.outbox, Orders: a.orders, Quotes: a.acceptedQuotes,
			Variants: a.variants,
		},
		Procurement: procurementapp.Deps{
			Clock: a.clock, UoW: memory.UnitOfWork{}, Outbox: a.outbox, Tasks: a.tasks, Shops: memory.NewShopRepo(), Items: memory.NewItemRepo(), ACL: merchant.Manual{},
			Variants: memory.NewVariantRepo(),
		},
		Logistics: logisticsapp.Deps{
			Clock: a.clock, UoW: memory.UnitOfWork{}, Outbox: a.outbox, Parcels: a.parcels, Batches: a.batches, Lanes: a.laneRules,
		},
		Reporting: reportingapp.Deps{
			UoW: memory.UnitOfWork{}, Summaries: a.summaries, Names: a.names, Worklist: a.worklist,
		},
		Auth:   a.auth,
		Tokens: a.auth, Registry: a.auth, // one store, all three halves of the port
		Extractor: openai.Fake{Draft: catalogapp.ListingDraft{
			Name: "Air Trainer 90", Price: "150.00", Currency: "USD", CategoryHint: "footwear",
		}},
	}
	if tweak != nil {
		tweak(&d)
	}
	a.h = httpapi.NewHandler(d)
	return a
}

// The two callers a handler test can be. The ids are fields on api so a test
// can assert "this listing was confirmed by the token's operator" and name the
// value it expects.
const (
	devOperatorToken = "test-operator"
	devCustomerToken = "test-customer"
)

func (a *api) asOperator() map[string]string {
	return map[string]string{"Authorization": "Bearer " + devOperatorToken}
}

func (a *api) asCustomer() map[string]string {
	return map[string]string{"Authorization": "Bearer " + devCustomerToken}
}

// call performs one request and returns the recorded response.
//
// If the caller did not name itself, the OPERATOR token is added. Almost every
// test in this package exercises a route's own logic, not the door, and
// spelling out a token in three hundred places would bury what each test is
// about. A test that cares about the door uses callAnon; a test that needs the
// other kind passes asCustomer.
func (a *api) call(t *testing.T, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	if _, named := headers["Authorization"]; !named {
		merged := map[string]string{"Authorization": "Bearer " + devOperatorToken}
		for k, v := range headers {
			merged[k] = v
		}
		headers = merged
	}
	return a.callAnon(t, method, path, body, headers)
}

// callAnon sends exactly the headers it is given and adds nothing. It is how
// the 401 tests ask a question that call would answer for them.
func (a *api) callAnon(t *testing.T, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// expectError asserts status and the stable error code; the message is the
// domain's own text and is only checked to be non-empty.
func expectError(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body %s", rec.Code, status, rec.Body)
	}
	var e errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("error body is not JSON: %v: %s", err, rec.Body)
	}
	if e.Error.Code != code || e.Error.Message == "" {
		t.Fatalf("error = %+v, want code %q with a message", e.Error, code)
	}
}

func idOf(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.ID == "" {
		t.Fatalf("no id in %s (%v)", rec.Body, err)
	}
	return out.ID
}

const merchantJSON = `{
	"name": "Example Sports",
	"site": "www.example.com",
	"currency": "USD",
	"free_shipping": {"kind": "over", "threshold": "50.00"},
	"sourcing": ["operator"]
}`

func (a *api) seedMerchant(t *testing.T) *catalog.Merchant {
	t.Helper()
	m, err := catalog.RegisterMerchant(catalog.MerchantDetails{
		Name: "Example Sports", Site: catalog.MustParseHostname("www.example.com"), Currency: shared.USD,
		// All three modes: most tests here are about a route's own logic, and a
		// shop that refused customer pastes would make them fail for a reason
		// that has nothing to do with what they are testing. The rule itself
		// gets its own test below.
		Sourcing: []catalog.SourcingMode{catalog.SourcedByOperator, catalog.SourcedByCustomer, catalog.SourcedByFeed},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	m.PullEvents()
	if err := a.merchants.Save(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	return m
}

func (a *api) seedCategory(t *testing.T) {
	t.Helper()
	c, err := catalog.NewCategoryPolicy(catalog.MustParseCategoryCode("footwear"),
		shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.categories.Save(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}

// ── POST /merchants ──────────────────────────────────────────────────────────

// Tầng 1 của DDD.md §30: JSON vào, value object ra, use case chạy, 201 + id.
func TestPOSTMerchants_created(t *testing.T) {
	a := newAPI()

	rec := a.call(t, "POST", "/merchants", merchantJSON, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	id, err := catalog.ParseMerchantID(idOf(t, rec))
	if err != nil {
		t.Fatalf("id in response: %v", err)
	}
	m, err := a.merchants.ByID(context.Background(), id)
	if err != nil {
		t.Fatalf("merchant not saved: %v", err)
	}
	if got := m.FreeShipping().String(); got != "free over 50.00 USD" {
		t.Errorf("FreeShipping = %q", got)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "catalog.merchant_registered" {
		t.Errorf("outbox = %v", evs)
	}
}

// The adapter — not the domain — decides what a comma means, and it decides by
// the request's language. Same bytes, two amounts. This is why ParseMoney
// refuses to guess (DDD.md §11d): somebody has to know the locale, and only
// the edge of the system does.
func TestPOSTMerchants_localeDecidesTheDecimalSeparator(t *testing.T) {
	body := strings.Replace(merchantJSON, `"50.00"`, `"50,00"`, 1)

	cases := map[string]struct {
		lang string
		want string
	}{
		"vi: comma is the decimal mark": {"vi-VN,vi;q=0.9", "free over 50.00 USD"},
		"en: comma groups thousands":    {"en-US", "free over 5000.00 USD"},
		"no header: en is the default":  {"", "free over 5000.00 USD"},
	}
	for name, c := range cases {
		a := newAPI()
		headers := map[string]string{}
		if c.lang != "" {
			headers["Accept-Language"] = c.lang
		}
		rec := a.call(t, "POST", "/merchants", body, headers)
		if rec.Code != http.StatusCreated {
			t.Fatalf("%s: status = %d, body %s", name, rec.Code, rec.Body)
		}
		id, _ := catalog.ParseMerchantID(idOf(t, rec))
		m, _ := a.merchants.ByID(context.Background(), id)
		if got := m.FreeShipping().String(); got != c.want {
			t.Errorf("%s: FreeShipping = %q, want %q", name, got, c.want)
		}
	}
}

// Every refusal has a status and a stable code; the message is the domain's.
func TestPOSTMerchants_errorsAreMapped(t *testing.T) {
	cases := map[string]struct {
		body   string
		status int
		code   string
	}{
		"broken json":        {`{"name": `, 400, "bad_json"},
		"unknown field":      {`{"nmae": "x"}`, 400, "bad_json"},
		"bad hostname":       {strings.Replace(merchantJSON, `"www.example.com"`, `"https://example.com"`, 1), 400, "invalid_hostname"},
		"unknown currency":   {strings.Replace(merchantJSON, `"USD"`, `"XXX"`, 1), 400, "unknown_currency"},
		"bad threshold":      {strings.Replace(merchantJSON, `"50.00"`, `"fifty"`, 1), 400, "malformed_amount"},
		"unknown kind":       {strings.Replace(merchantJSON, `"over"`, `"sometimes"`, 1), 400, "invalid_request"},
		"empty name":         {strings.Replace(merchantJSON, `"Example Sports"`, `"  "`, 1), 400, "empty_name"},
		"unknown sourcing":   {strings.Replace(merchantJSON, `"operator"`, `"telepathy"`, 1), 400, "unknown_sourcing"},
		"threshold currency": {strings.Replace(merchantJSON, `"USD"`, `"VND"`, 1), 400, "malformed_amount"}, // VND has no decimals: "50.00" is malformed for VND
	}
	for name, c := range cases {
		rec := newAPI().call(t, "POST", "/merchants", c.body, nil)
		t.Run(name, func(t *testing.T) { expectError(t, rec, c.status, c.code) })
	}
}

// ── POST /products ───────────────────────────────────────────────────────────

func productJSON(merchantID string) string {
	return `{
		"name": "Air Trainer 90",
		"merchant_id": "` + merchantID + `",
		"category": "footwear",
		"source_url": "https://www.example.com/t/air-trainer-90/abc",
		"price": "150.00",
		"currency": "USD"
	}`
}

func TestPOSTProducts_created(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)

	rec := a.call(t, "POST", "/products", productJSON(m.ID().String()), a.asCustomer())
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	id, err := catalog.ParseProductID(idOf(t, rec))
	if err != nil {
		t.Fatal(err)
	}
	p, err := a.products.ByID(context.Background(), id)
	if err != nil || p.Price().String() != "150.00 USD" || p.ListingProvenance().Verified() {
		t.Fatalf("saved product: %v, %v", p, err)
	}
}

// The caller's identity comes from the edge (a header today, real auth later)
// and becomes a Provenance the domain can trust.
// 400 = the request cannot be understood; 404 = points at nothing; 409 = the
// request is fine and the business says no. Three different conversations
// with the client, three different codes.
func TestPOSTProducts_400vs404vs409(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)
	ok := productJSON(m.ID().String())

	expectError(t, a.call(t, "POST", "/products", strings.Replace(ok, `"150.00"`, `"abc"`, 1), nil), 400, "malformed_amount")
	expectError(t, a.call(t, "POST", "/products", strings.Replace(ok, m.ID().String(), "not-an-id", 1), nil), 400, "invalid_id")
	expectError(t, a.call(t, "POST", "/products", productJSON(catalog.NewMerchantID().String()), nil), 404, "merchant_not_found")
	expectError(t, a.call(t, "POST", "/products", strings.Replace(ok, `"footwear"`, `"luggage"`, 1), nil), 404, "category_not_found")
	// A valid VND amount for a USD merchant: the adapter parses it fine, the use case refuses it.
	vnd := strings.NewReplacer(`"150.00"`, `"3900000"`, `"currency": "USD"`, `"currency": "VND"`).Replace(ok)
	expectError(t, a.call(t, "POST", "/products", vnd, nil), 409, "price_currency")

	if err := m.Suspend("account banned", now); err != nil {
		t.Fatal(err)
	}
	expectError(t, a.call(t, "POST", "/products", ok, nil), 409, "merchant_inactive")
}

// ── POST /products/{id}/publish ──────────────────────────────────────────────

func TestPOSTPublish(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)
	id := idOf(t, a.call(t, "POST", "/products", productJSON(m.ID().String()), nil))
	a.outbox.Drain()

	expectError(t, a.call(t, "POST", "/products/not-an-id/publish", "", nil), 400, "invalid_id")
	expectError(t, a.call(t, "POST", "/products/"+catalog.NewProductID().String()+"/publish", "", nil), 404, "product_not_found")
	expectError(t, a.call(t, "POST", "/products/"+id+"/publish", "", nil), 409, "no_variants")

	// Make it publishable the way later use cases will: variant, confirmed listing, measured.
	pid, _ := catalog.ParseProductID(id)
	p, _ := a.products.ByID(context.Background(), pid)
	verified := catalog.MustProvenance(catalog.SourcedByOperator, now, shared.NewOperatorID())
	if _, err := p.AddVariant(catalog.VariantDetails{Size: "US 9"}, now); err != nil {
		t.Fatal(err)
	}
	if err := p.ConfirmListing(verified, now); err != nil {
		t.Fatal(err)
	}
	if err := p.Measure(shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), verified, now); err != nil {
		t.Fatal(err)
	}
	p.PullEvents()

	rec := a.call(t, "POST", "/products/"+id+"/publish", "", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "catalog.product_published" {
		t.Errorf("outbox = %v", evs)
	}
	expectError(t, a.call(t, "POST", "/products/"+id+"/publish", "", nil), 409, "not_draft")
}

func TestUnknownRouteIs404(t *testing.T) {
	a := newAPI()

	// A path nobody ever defined.
	if rec := a.call(t, "GET", "/nope", "", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("GET /nope = %d", rec.Code)
	}

	// And a path that is missing ON PURPOSE: there is no GET /products/{id},
	// because reading an aggregate back is the read model's job, not the write
	// side's (WALKTHROUGH §20). If this ever starts answering 200, somebody
	// added a read route to the write side and this test is the objection.
	rec := a.call(t, "GET", "/products/"+shared.NewID().String(), "", nil)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /products/{id} = %d, want it to stay absent", rec.Code)
	}
}

// ── POST /categories ─────────────────────────────────────────────────────────

const categoryJSON = `{
	"code": "luggage",
	"estimate": {"weight_g": 3200, "length_mm": 700, "width_mm": 450, "height_mm": 280},
	"restrictions": ["battery"]
}`

func TestPOSTCategories(t *testing.T) {
	a := newAPI()

	rec := a.call(t, "POST", "/categories", categoryJSON, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	c, err := a.categories.ByCode(context.Background(), catalog.MustParseCategoryCode("luggage"))
	if err != nil || c.DefaultParcelSpec().Weight() != shared.Grams(3200) || !c.Restricted(catalog.RestrictionBattery) {
		t.Fatalf("saved category = %+v, %v", c, err)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "catalog.category_defined" {
		t.Errorf("outbox = %v", evs)
	}

	expectError(t, a.call(t, "POST", "/categories", strings.Replace(categoryJSON, `"luggage"`, `"Giày Dép!"`, 1), nil), 400, "invalid_category_code")
	expectError(t, a.call(t, "POST", "/categories", strings.Replace(categoryJSON, `"weight_g": 3200`, `"weight_g": 0`, 1), nil), 400, "incomplete_parcel_spec")
	expectError(t, a.call(t, "POST", "/categories", strings.Replace(categoryJSON, `"battery"`, `"cursed"`, 1), nil), 400, "unknown_restriction")
}
