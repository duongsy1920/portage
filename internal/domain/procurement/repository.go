package procurement

import (
	"context"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Shop and Item are procurement's PROJECTIONS of catalog: which shop a product
// belongs to and what that shop bills in. Fed by catalog.merchant_registered
// and catalog.product_published — procurement never reads catalog's tables.
type Shop struct {
	Merchant shared.ID
	Name     string
	Site     string
	Currency shared.Currency
}

type Item struct {
	Product  shared.ID
	Merchant shared.ID
	Name     string
	Source   string // the page to buy from, as catalog announced it
}

// Variant is what procurement knows about one purchasable form, from
// catalog.variant_added: the size and colour IN THE SHOP'S OWN WORDS.
//
// This projection exists for one reason. Somebody has to walk into a shop and
// ask for a size, and the order names that size only by id — procurement may
// not read catalog's tables to translate it (guard 7). Without this, the
// buyer's screen shows a UUID.
type Variant struct {
	Variant     shared.ID
	Product     shared.ID
	Size        string
	Color       string
	MerchantRef string
}

// Label is the size and colour as one line for the person buying: "M 8 / W 9.5
// · black". Empty when the product has a single nameless form, which is legal
// (catalog.VariantDetails) — the caller then shows the product name alone.
func (v Variant) Label() string {
	switch {
	case v.Size != "" && v.Color != "":
		return v.Size + " · " + v.Color
	case v.Size != "":
		return v.Size
	default:
		return v.Color
	}
}

// The PORTS of procurement.

type ShopRepository interface {
	ByID(ctx context.Context, merchant shared.ID) (Shop, error) // ErrShopNotFound
	Save(ctx context.Context, s Shop) error
}

type ItemRepository interface {
	ByProduct(ctx context.Context, product shared.ID) (Item, error) // ErrItemNotFound
	Save(ctx context.Context, i Item) error
}

type VariantRepository interface {
	ByID(ctx context.Context, variant shared.ID) (Variant, error) // ErrVariantNotFound
	Save(ctx context.Context, v Variant) error
}

type TaskRepository interface {
	ByID(ctx context.Context, id TaskID) (*PurchaseTask, error) // ErrTaskNotFound
	// ByOrder finds the task for an order — one deposited order, one task.
	ByOrder(ctx context.Context, order shared.ID) (*PurchaseTask, error) // ErrTaskNotFound
	// Open lists tasks nobody has closed yet, oldest first: the buyer's to-do list.
	Open(ctx context.Context) ([]*PurchaseTask, error)
	Save(ctx context.Context, t *PurchaseTask) error
}

// MerchantACL is the Anti-Corruption Layer port (DDD.md §22): how we place an
// order at a shop, in OUR words. Purchase returns what the shop told us,
// already translated into a PurchaseReceipt, or an error in our vocabulary.
//
// The first adapter is a human: adapter/merchant.Manual, which cannot buy by
// itself and says so (ErrManualPurchase) — the operator buys in a browser and
// reports through POST /purchase-tasks/{id}/confirm. A real shop API becomes a
// second adapter behind the same interface; the domain and the use cases do
// not change.
//
// [PHP] Interface MerchantGatewayInterface với hai implementation, chọn bằng
// [PHP] services.yaml — chính là "port" ở đây.
type MerchantACL interface {
	// Purchase attempts to buy the task's variant. ok=false with a reason means
	// the shop refused (sold out); an error means we could not even ask.
	Purchase(ctx context.Context, task *PurchaseTask) (receipt PurchaseReceipt, ok bool, reason string, err error)
}
