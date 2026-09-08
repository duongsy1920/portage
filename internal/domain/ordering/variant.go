package ordering

import "github.com/duongsy/portage/internal/domain/shared"

// Variant is what ordering knows about a purchasable form: only that it
// exists, and which product it belongs to. A PROJECTION of
// catalog.variant_added, like AcceptedQuote is one of pricing's event.
//
// It exists for one check. The customer sends a variant id in POST /orders,
// and until now nothing could tell whether that id was even real — ordering
// may not read catalog's tables (guard 7), so any uuid was accepted and the
// buyer found out at the shop. Two fields are enough to close that.
//
// Deliberately NOT here: the size and the colour. Ordering has no rule that
// needs the words, and a projection that carries data nobody reads goes stale
// silently. Procurement keeps those, because its buyer has to say them out
// loud (procurement.Variant).
type Variant struct {
	Variant shared.ID
	Product shared.ID
}
