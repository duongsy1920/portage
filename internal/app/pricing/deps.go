// Package pricingapp holds the use cases of the pricing context: issue and
// accept quotes, and the Projector that keeps pricing's copy of catalog data
// up to date from catalog's events. The two things it decides on its own —
// which lane, which rate to freeze — are orchestration; every number is the
// domain's (pricing.Calculate).
//
// [PHP] MessageHandlers của bundle Pricing. Projector là handler nhận event từ
// [PHP] bundle Catalog qua Messenger và ghi vào bảng của Pricing.
package pricingapp

import (
	"github.com/duongsy/portage/internal/app"
	"github.com/duongsy/portage/internal/domain/pricing"
)

// Deps is everything the pricing use cases are wired to (convention 10).
// Policy and Classification are business CONFIGURATION, not services: they are
// value objects loaded once at start-up and frozen into every quote.
type Deps struct {
	Clock  app.Clock
	UoW    app.UnitOfWork
	Outbox app.Outbox

	Lanes           pricing.LaneRepository
	Quotes          pricing.QuoteRepository
	Listings        pricing.ListingRepository
	Profiles        pricing.CategoryProfileRepository
	Rates           pricing.ExchangeRates
	Reconciliations pricing.ReconciliationRepository // Quote vs Actual, per order

	Policy         pricing.QuotePolicy
	Classification pricing.Classification
}

func mustHave(handler string, deps map[string]any) {
	app.MustHave("pricingapp: "+handler, deps)
}
