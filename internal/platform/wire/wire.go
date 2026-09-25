// Package wire builds the object graph — every adapter, plugged into every
// port of every context — in ONE place, so cmd/api, cmd/worker and the tests
// all get the same graph and cannot drift apart. There are two graphs: Memory
// for a dev run and for tests, Postgres for everything that has to survive a
// restart.
//
// The graph also carries the BUSINESS CONFIGURATION pricing quotes with (the
// sales-tax rate, our margin, the deposit share, how long a quote lives, which
// category is which goods class, the forwarder's lane and price list). The
// Postgres graph reads them from config/ratecard.yaml through adapter/config;
// the Memory graph keeps the same numbers as a FIXTURE below, and
// TestFixtureMatchesTheRateCardFile keeps the two equal. Either way they are
// value objects, frozen into every quote — SETUP.md §7.
//
// [PHP] Đây là services.yaml + config/packages/*.yaml dưới dạng hai hàm Go.
// [PHP] Không autowire: mỗi dependency được tạo và cắm ngay trước mắt, và
// [PHP] chọn adapter nào là một dòng `if` trong main(), không phải env-var
// [PHP] tên "doctrine.dbal.url" ẩn sau ba lớp config.
package wire

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/duongsy/portage/internal/adapter/config"
	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/adapter/merchant"
	"github.com/duongsy/portage/internal/adapter/openai"
	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/app"
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
	"github.com/duongsy/portage/internal/worker"
)

// Graph is the wired system: one Deps per bounded context, sharing the clock,
// the unit of work and the outbox — and the outbox's read side for the relay.
type Graph struct {
	Catalog     catalogapp.Deps
	Pricing     pricingapp.Deps
	Ordering    orderingapp.Deps
	Procurement procurementapp.Deps
	Logistics   logisticsapp.Deps
	Reporting   reportingapp.Deps // the READ side: order_summaries, written only by events
	Source      worker.Source     // the same outbox the handlers write to, seen from the relay

	// Extractor reads a shop page with a language model. Memory wires a Fake;
	// Postgres wires the real client when OPENAI_API_KEY is set, and
	// openai.Unavailable when it is not — never the Fake.
	Extractor catalogapp.ListingExtractor

	// Auth is the door: who may call the API. Memory wires fixed dev tokens;
	// Postgres wires the api_tokens table.
	Auth auth.Verifier

	// Tokens cuts new keys. It is the SAME object as Auth in both wirings —
	// one store, two halves of the port — but a separate field, because
	// almost every caller must only be able to verify.
	Tokens auth.Issuer

	// Registry is the same object again — listing and revoking keys.
	Registry auth.Registry
}

// Memory wires the in-memory adapters and the dev seed. Nothing survives the
// process; that is the point — a dev run and a test start from the same three
// categories, one lane and one rate every time.
func Memory(clk app.Clock) Graph {
	outbox := memory.NewOutbox()
	devStatic := DevTokens()
	items := memory.NewItemRepo() // shared with the ACL Router below
	uow := memory.UnitOfWork{}
	g := Graph{
		Catalog: catalogapp.Deps{
			Clock:      clk,
			UoW:        uow,
			Merchants:  memory.NewMerchantRepo(),
			Categories: memory.NewCategoryRepo(),
			Products:   memory.NewProductRepo(),
			Outbox:     outbox,
		},
		Pricing: pricingapp.Deps{
			Clock:           clk,
			UoW:             uow,
			Outbox:          outbox,
			Lanes:           memory.NewLaneRepo(),
			Quotes:          memory.NewQuoteRepo(),
			Listings:        memory.NewListingRepo(),
			Profiles:        memory.NewProfileRepo(),
			Rates:           memory.NewExchangeRates(),
			Reconciliations: memory.NewReconciliationRepo(),
			Policy:          quotePolicy(),
			Classification:  goodsClasses(),
		},
		Ordering: orderingapp.Deps{
			Clock:    clk,
			UoW:      uow,
			Outbox:   outbox,
			Orders:   memory.NewOrderRepo(),
			Quotes:   memory.NewAcceptedQuoteRepo(),
			Variants: memory.NewOrderingVariantRepo(),
		},
		Logistics: logisticsapp.Deps{
			Clock:   clk,
			UoW:     uow,
			Outbox:  outbox,
			Parcels: memory.NewParcelRepo(),
			Batches: memory.NewBatchRepo(),
			Lanes:   memory.NewLaneRuleRepo(),
		},
		Procurement: procurementapp.Deps{
			Clock:    clk,
			UoW:      uow,
			Outbox:   outbox,
			Tasks:    memory.NewTaskRepo(),
			Shops:    memory.NewShopRepo(),
			Items:    items,
			Variants: memory.NewVariantRepo(),
			// Router, not Manual directly: adding a shop with an API is one
			// entry in ByMerchant here, and nothing above changes (DDD.md §22).
			ACL: merchant.Router{Items: items}, // no shop has an adapter yet → every task goes to a person
		},
		Reporting: reportingapp.Deps{
			UoW:       uow,
			Summaries: memory.NewOrderSummaryRepo(),
			Names:     memory.NewProductNameRepo(),
			Worklist:  memory.NewProductWorklistRepo(),
		},
		// A Fake extractor in memory: `go run ./cmd/api` must work with no API
		// key, no network and no bill. Postgres does NOT fall back to it.
		Extractor: openai.Fake{Draft: catalogapp.ListingDraft{
			Name: "Air Trainer 90", Price: "150.00", Currency: "USD", CategoryHint: "footwear",
		}},
		Source:   outbox,
		Auth:     devStatic,
		Tokens:   devStatic,
		Registry: devStatic,
	}
	if err := seed(context.Background(), g, []pricing.LaneDetails{fixtureLane()}); err != nil {
		log.Fatalf("wire: seed: %v", err) // programmer data; a typo stops the process (convention 1)
	}
	return g
}

// The two subjects the dev tokens stand for. They are FIXED, not generated, so
// a smoke script and a handler test can predict what a run produces — and they
// are valid UUID v7 values, because shared.ParseID accepts nothing else.
const (
	DevOperatorID = "01920000-0000-7000-8000-000000000001"
	DevCustomerID = "01920000-0000-7000-8000-000000000002"

	// Tokens for a MEMORY run only. wire.Postgres never sees them: a fixed
	// credential in a real database is a back door, so there the first
	// operator arrives through cmd/api -bootstrap-operator-token.
	DevOperatorToken = "dev-operator"
	DevCustomerToken = "dev-customer"
)

// DevTokens is the auth.Verifier a memory run and every handler test use — the
// dev half of the port platform/auth declares. postgres.TokenRepo is the other
// half, and answers identically.
func DevTokens() *auth.Static {
	return auth.NewStatic(map[string]auth.Principal{
		DevOperatorToken: auth.MustPrincipal(auth.Operator, mustID(DevOperatorID)),
		DevCustomerToken: auth.MustPrincipal(auth.Customer, mustID(DevCustomerID)),
	})
}

func mustID(s string) shared.ID {
	id, err := shared.ParseID(s)
	if err != nil {
		panic(fmt.Sprintf("wire: %q is not a v7 uuid: %v", s, err))
	}
	return id
}

// Postgres reads the rate card, connects, migrates, seeds (upserts, so a
// second start changes nothing) and wires the Postgres adapters. A wrong DSN
// or a wrong rate card fails HERE, at start-up — never on the first request.
// The rate card is read first: it is local and cheap, and an operator who
// mistyped a number should not have to wait for a database to find out.
// The returned close releases the pool.
func Postgres(ctx context.Context, clk app.Clock, dsn, ratecard string) (g Graph, closeFn func(), err error) {
	rc, err := config.Load(ratecard)
	if err != nil {
		return Graph{}, nil, err
	}
	pool, err := postgres.Connect(ctx, dsn)
	if err != nil {
		return Graph{}, nil, err
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		pool.Close()
		return Graph{}, nil, err
	}
	outbox := postgres.NewOutbox(pool)
	tokenRepo := postgres.NewTokenRepo(pool)
	items := postgres.NewItemRepo(pool) // shared with the ACL Router below
	uow := postgres.NewUnitOfWork(pool)
	// No key, no extractor — and no silent fallback to the Fake: a feature
	// that is off must answer 503, not 201 with an invented price.
	var extractor catalogapp.ListingExtractor = openai.Unavailable{}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		extractor = openai.New(key)
	}
	g = Graph{
		Catalog: catalogapp.Deps{
			Clock:      clk,
			UoW:        uow,
			Merchants:  postgres.NewMerchantRepo(pool),
			Categories: postgres.NewCategoryRepo(pool),
			Products:   postgres.NewProductRepo(pool),
			Outbox:     outbox,
		},
		Pricing: pricingapp.Deps{
			Clock:           clk,
			UoW:             uow,
			Outbox:          outbox,
			Lanes:           postgres.NewLaneRepo(pool),
			Quotes:          postgres.NewQuoteRepo(pool),
			Listings:        postgres.NewListingRepo(pool),
			Profiles:        postgres.NewProfileRepo(pool),
			Rates:           postgres.NewExchangeRates(pool),
			Reconciliations: postgres.NewReconciliationRepo(pool),
			Policy:          rc.Policy,
			Classification:  rc.Classification,
		},
		Ordering: orderingapp.Deps{
			Clock:    clk,
			UoW:      uow,
			Outbox:   outbox,
			Orders:   postgres.NewOrderRepo(pool),
			Quotes:   postgres.NewAcceptedQuoteRepo(pool),
			Variants: postgres.NewOrderingVariantRepo(pool),
		},
		Logistics: logisticsapp.Deps{
			Clock:   clk,
			UoW:     uow,
			Outbox:  outbox,
			Parcels: postgres.NewParcelRepo(pool),
			Batches: postgres.NewBatchRepo(pool),
			Lanes:   postgres.NewLaneRuleRepo(pool),
		},
		Procurement: procurementapp.Deps{
			Clock:    clk,
			UoW:      uow,
			Outbox:   outbox,
			Tasks:    postgres.NewTaskRepo(pool),
			Shops:    postgres.NewShopRepo(pool),
			Items:    items,
			Variants: postgres.NewVariantRepo(pool),
			ACL:      merchant.Router{Items: items},
		},
		Reporting: reportingapp.Deps{
			UoW:       uow,
			Summaries: postgres.NewOrderSummaryRepo(pool),
			Names:     postgres.NewProductNameRepo(pool),
			Worklist:  postgres.NewProductWorklistRepo(pool),
		},
		Extractor: extractor,
		Source:    outbox,
		Auth:      tokenRepo,
		Tokens:    tokenRepo,
		Registry:  tokenRepo,
	}
	if err := seed(ctx, g, rc.Lanes); err != nil {
		pool.Close()
		return Graph{}, nil, fmt.Errorf("wire: seed: %w", err)
	}
	return g, pool.Close, nil
}

// ── business configuration: the FIXTURE ──────────────────────────────────────
//
// Production reads these numbers from config/ratecard.yaml (adapter/config);
// the three functions below are the same numbers for the Memory graph, so a
// dev run and every test start without a file. They are the test's data, not
// the operator's: a typo here is a programmer error, hence Must-style panics
// (convention 1) — and TestFixtureMatchesTheRateCardFile fails the build the
// day one side is changed without the other.

// quotePolicy is the fixture's quote policy (SETUP.md §7): Denver sales tax,
// 10 % margin with a 500 000 ₫ floor, half up front, 48 h.
func quotePolicy() pricing.QuotePolicy {
	margin, err := pricing.NewMarginPolicy(shared.MustParsePercent("10"), shared.MustParseMoney("500000", shared.VND))
	if err != nil {
		panic(err)
	}
	policy, err := pricing.NewQuotePolicy(pricing.QuotePolicyDetails{
		SalesTax: shared.MustParsePercent("8.81"),
		Margin:   margin,
		Deposit:  shared.MustParsePercent("50"),
		TTL:      48 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	return policy
}

// goodsClasses is the fixture's map of OUR categories onto the forwarder's
// price list. Anything not listed is "standard" — the cheapest class, so an
// omission shows up as a too-low quote at reconciliation, never as an
// overcharged customer.
func goodsClasses() pricing.Classification {
	c, err := pricing.NewClassification(map[string]pricing.GoodsClass{
		"footwear":    pricing.ClassBranded,
		"apparel":     pricing.ClassBranded,
		"electronics": pricing.ClassElectronics,
	})
	if err != nil {
		panic(err)
	}
	return c
}

// ── dev seed ─────────────────────────────────────────────────────────────────

// seed gives every fresh system the reference data the docs and demos use:
// the categories, the lanes it is handed (the rate card's, or the fixture's)
// and today's exchange rate. Categories are DEV data; a deployment defines
// them through the API. Lanes come from the rate card on purpose — they are
// the forwarder's price list, and that is configuration, not demo data.
func seed(ctx context.Context, g Graph, lanes []pricing.LaneDetails) error {
	if err := seedCategories(ctx, g.Catalog); err != nil {
		return err
	}
	return seedPricing(ctx, g.Pricing, lanes)
}

// seedCategories runs THROUGH the use case, so each start-up also announces
// the categories (catalog.category_defined) and pricing's projection fills
// itself once the relay runs. Repeating the announcement is harmless by
// design: consumers are idempotent.
func seedCategories(ctx context.Context, deps catalogapp.Deps) error {
	define := catalogapp.NewDefineCategoryHandler(deps)
	seed := []struct {
		code         string
		spec         shared.ParcelSpec
		restrictions []catalog.Restriction
	}{
		{"footwear", shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13)), nil},
		{"apparel", shared.MustParcelSpec(shared.Grams(900), shared.NewDimensionsCM(40, 32, 18)), nil},
		{"electronics", shared.MustParcelSpec(shared.Grams(400), shared.NewDimensionsCM(20, 18, 8)),
			[]catalog.Restriction{catalog.RestrictionBattery, catalog.RestrictionMagnet}},
	}
	for _, s := range seed {
		if err := define.Handle(ctx, catalogapp.DefineCategory{Code: s.code, Estimate: s.spec, Restrictions: s.restrictions}); err != nil {
			return fmt.Errorf("category %s: %w", s.code, err)
		}
	}
	return nil
}

// fixtureLane is the fixture's lane: the forwarder's price list (SETUP.md §7)
// as config/ratecard.yaml also states it.
func fixtureLane() pricing.LaneDetails {
	usd := func(s string) shared.Money {
		return shared.MustParseMoney(s, shared.USD)
	}
	rates, err := pricing.NewRateCard(usd("9.00"), usd("10.00"), usd("12.00"), usd("14.00"))
	if err != nil {
		panic(err)
	}
	return pricing.LaneDetails{
		Code:             pricing.MustParseLaneCode("us_forwarder"),
		Name:             "US forwarder, Denver → Vietnam, air",
		Divisor:          5000,
		Step:             shared.Grams(500),
		Rates:            rates,
		BatterySurcharge: usd("3.00"),
		Duty:             pricing.DutyBundled(),
	}
}

// seedPricing: every lane it is handed, and today's rate. A lane goes
// THROUGH DefineLane so it is announced — logistics keeps the divisor and the
// step to split freight — and so a restart with a changed rate card
// redefines the lane the same way POST would. The rate goes straight to the
// repository: nobody listens for rates.
func seedPricing(ctx context.Context, deps pricingapp.Deps, lanes []pricing.LaneDetails) error {
	define := pricingapp.NewDefineLaneHandler(deps)
	for _, lane := range lanes {
		if err := define.Handle(ctx, lane); err != nil {
			return fmt.Errorf("lane %s: %w", lane.Code, err)
		}
	}
	// Through the handler, not the repository: the seed takes the same road
	// POST /fx takes, so a rule added to that road applies to the seed too.
	if err := pricingapp.NewSetExchangeRateHandler(deps).Handle(ctx, shared.MustExchangeRate(shared.USD, shared.VND, "26000")); err != nil {
		return fmt.Errorf("fx: %w", err)
	}
	return nil
}
