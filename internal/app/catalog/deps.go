// Package catalogapp holds the USE CASES of the catalog context — the
// application layer of DDD.md §30. A handler here opens the transaction, asks
// the clock, loads aggregates, lets the DOMAIN decide, saves, and hands the
// recorded events to the outbox. It contains no business rule of its own;
// the two rules it does enforce (ErrMerchantInactive, ErrPriceCurrency) span
// two aggregates, which is exactly what a use case is for.
//
// The package is named catalogapp, not catalog, so files here can import the
// domain package without aliasing every time. The directory follows
// convention 8: internal/app/<context>/.
//
// [PHP] Đây là các MessageHandler của Symfony Messenger — mỗi use case một
// [PHP] class, __invoke() nhận Command. Khác ở chỗ không có bus tự route:
// [PHP] adapter/http gọi thẳng handler.Handle().
package catalogapp

import (
	"errors"

	"github.com/duongsy/portage/internal/app"
	"github.com/duongsy/portage/internal/domain/catalog"
)

var (
	// ErrMerchantInactive: the merchant exists but is suspended — nothing new
	// is listed for a shop we cannot buy from. A rule across two aggregates
	// (Merchant, Product), so it lives here, not in either of them.
	ErrMerchantInactive = errors.New("merchant is not active")

	// ErrPriceCurrency: a product's price must be in its merchant's currency.
	// Product cannot know the merchant's currency; the use case does.
	ErrPriceCurrency = errors.New("price is not in the merchant's currency")
)

// Deps is everything the catalog use cases are wired to (convention 10).
// Each handler checks the fields it needs at construction and panics on nil:
// wiring is programmer input, and a missing dependency is a start-up bug,
// not a request-time error.
//
// [PHP] Bằng với constructor injection + autowiring của Symfony; ở đây main()
// [PHP] điền struct này bằng tay. Nil = "service không tồn tại" lúc compile
// [PHP] container bên Symfony — nên ở đây cũng phải nổ ngay lúc khởi động.
type Deps struct {
	Clock      app.Clock
	UoW        app.UnitOfWork
	Merchants  catalog.MerchantRepository
	Categories catalog.CategoryRepository
	Products   catalog.ProductRepository
	Outbox     app.Outbox
}

// mustHave is app.MustHave with this package's name in the message.
func mustHave(handler string, deps map[string]any) {
	app.MustHave("catalogapp: "+handler, deps)
}
