package catalog

import (
	"context"
	"errors"
)

var (
	ErrMerchantNotFound = errors.New("merchant not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrProductNotFound  = errors.New("product not found")
)

// MerchantRepository is the PORT through which the rest of the system loads
// and stores merchants. The domain declares what it needs; adapter/postgres
// will implement it. Nothing here knows SQL exists.
//
// Two rules, from DDD.md §16:
//
//   - It works in whole aggregates. ByID returns a *Merchant with every
//     invariant intact, never a row or a partial view.
//   - It is not a query API. There is no ByStatusAndCurrencyAddedAfter(...);
//     screens that need lists and filters get a read model (CQRS, §24).
//
// ctx carries cancellation and deadlines from the caller (an HTTP request
// that gave up, a worker shutting down) down to the database call. It is
// the first parameter by Go convention, on every method that can block.
//
// [PHP] Đây là interface Repository của DDD — KHÔNG kế thừa gì cả. So với
// [PHP] `MerchantRepository extends ServiceEntityRepository`: bên Symfony,
// [PHP] domain biết Doctrine tồn tại; ở đây domain chỉ khai báo interface,
// [PHP] Postgres cắm vào từ internal/adapter. `context.Context` không có bản
// [PHP] tương đương trong PHP-FPM vì mỗi request là một process sống độc lập.
type MerchantRepository interface {
	// ByID returns ErrMerchantNotFound (wrapped) when there is no such merchant.
	ByID(ctx context.Context, id MerchantID) (*Merchant, error)

	// BySite finds the merchant that owns a shop hostname — how a product
	// reference the customer supplies is tied back to who we buy from.
	BySite(ctx context.Context, site Hostname) (*Merchant, error)

	// All returns every merchant in the order they were added, for the screen
	// that lets a person PICK a shop instead of typing its id. Small by
	// nature, like CategoryRepository.All: shops are added one at a time by
	// hand, because each one is a decision about who we are willing to buy
	// from and in which currency.
	All(ctx context.Context) ([]*Merchant, error)

	// Save persists a new or changed merchant. It does NOT publish the
	// merchant's events: the application layer pulls them after Save
	// returns and writes them to the outbox in the same transaction.
	Save(ctx context.Context, m *Merchant) error
}

// CategoryRepository stores CategoryPolicy values keyed by their natural code.
// Save replaces the whole policy under its code — the value object has no
// partial updates (see CategoryPolicy).
type CategoryRepository interface {
	// ByCode returns ErrCategoryNotFound (wrapped) when the code is unknown.
	ByCode(ctx context.Context, code CategoryCode) (CategoryPolicy, error)

	// All returns every policy, for the operator screen that edits them and
	// for pricing to load once at start-up. The set is small by nature.
	All(ctx context.Context) ([]CategoryPolicy, error)

	Save(ctx context.Context, p CategoryPolicy) error
}

// ProductRepository loads and stores Product aggregates — whole, with their
// variants. There is no VariantRepository on purpose (DDD.md §14, rule 1).
type ProductRepository interface {
	// ByID returns ErrProductNotFound (wrapped) when there is no such product.
	ByID(ctx context.Context, id ProductID) (*Product, error)

	// BySource returns every product recorded for a page URL — the first
	// question duplicate detection asks (CATALOG.md §7, option C). The app
	// layer decides whether to flag; the domain only remembers the flag.
	BySource(ctx context.Context, source SourceURL) ([]*Product, error)

	// Save persists a new or changed product and its variants. It does NOT
	// publish the product's events (see MerchantRepository.Save).
	Save(ctx context.Context, p *Product) error
}
