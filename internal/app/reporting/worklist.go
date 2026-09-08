package reportingapp

import (
	"context"
	"errors"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// The worklist is the second read model, and it answers the question the write
// side cannot: "what is unfinished, and what is the next thing to do about it".
//
// Catalog knows a product's state, but only as the reason Publish said no. That
// is enough for an API and useless for a screen: a person cannot be shown a
// 409 and asked to guess which of four steps is missing. So the four facts get
// remembered here as they are announced, and both screens read them.
//
// Nothing here is an invariant. A row can be wrong only in the sense of being
// stale, which is what eventual consistency means, and the fix for stale is to
// wait — never to refuse.

// ErrWorklistItemNotFound: no row for that product, or product_added has not
// been relayed yet.
var ErrWorklistItemNotFound = errors.New("worklist item not found")

// Step is one of the four things a person does to make a pasted link sellable,
// in the order Publish demands them.
type Step string

const (
	StepVariant Step = "variant" // no purchasable size exists yet
	StepListing Step = "listing" // nobody has vouched for what this IS
	StepMeasure Step = "measure" // nobody has put the box on a scale
	StepPublish Step = "publish" // everything is ready; say the word
	StepDone    Step = ""        // published: it can be quoted
)

// WorklistItem is one pasted product and how far along it is.
type WorklistItem struct {
	Product  shared.ID
	Merchant shared.ID
	Category string
	Name     string

	// Source is the page to open, and Price is what the shop was asking when
	// the link came in. Both are here because a row without them is not
	// something a person can act on.
	Source string
	Price  shared.Money

	// SourcedBy is how it arrived: customer, operator or feed. RequestedBy is
	// who is waiting, and it is ZERO when an operator added the product on
	// spec. A screen must not promise to notify nobody.
	SourcedBy   string
	RequestedBy shared.ID

	// RequestedVariant is the size the customer asked for, in their words.
	// It is a WISH, and it stays on the row after the real variant exists so
	// the two can be compared: "asked for US 9, we created US 9.5" is a
	// conversation somebody needs to have, and hiding it prevents it.
	RequestedVariant string

	ListingConfirmed bool
	Measured         bool
	Published        bool

	// Variants is every size announced so far. The id is here because the
	// customer has to pick one to place an order, and the label because no
	// screen may join back to catalog to find out what the id means.
	Variants []WorklistVariant

	AddedAt   time.Time
	UpdatedAt time.Time
}

// WorklistVariant is one purchasable form, in the shop's own words.
type WorklistVariant struct {
	ID    shared.ID
	Label string
}

// HasVariant: something is buyable. Derived rather than stored, so it can
// never disagree with the list it summarises.
func (w WorklistItem) HasVariant() bool {
	return len(w.Variants) > 0
}

// FirstLabel is what a one-line summary shows, or "" for a product with no
// size yet.
func (w WorklistItem) FirstLabel() string {
	if len(w.Variants) == 0 {
		return ""
	}
	return w.Variants[0].Label
}

// NextStep is what to do next, or StepDone when there is nothing left.
//
// It lives here rather than in each screen for one reason: two screens that
// each decide "what is missing" will disagree the day a fifth step appears,
// and the disagreement will be silent. One function, one test.
func (w WorklistItem) NextStep() Step {
	switch {
	case w.Published:
		return StepDone
	case !w.HasVariant():
		return StepVariant
	case !w.ListingConfirmed:
		return StepListing
	case !w.Measured:
		return StepMeasure
	default:
		return StepPublish
	}
}

// ProductWorklistRepository is the store. Save is an upsert, like every
// projection here: read a row, change a field, write it back.
//
// Open and ByRequester are the two screens, and they are the whole reason the
// table exists: "what is left to do" and "where is the thing I asked for".
type ProductWorklistRepository interface {
	ByProduct(ctx context.Context, product shared.ID) (WorklistItem, error) // ErrWorklistItemNotFound
	Open(ctx context.Context) ([]WorklistItem, error)                       // not published, oldest first
	ByRequester(ctx context.Context, customer shared.ID) ([]WorklistItem, error)
	Save(ctx context.Context, item WorklistItem) error
}
