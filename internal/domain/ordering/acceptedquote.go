package ordering

import "github.com/duongsy/portage/internal/domain/shared"

// AcceptedQuote is what ordering knows about a quote — a PROJECTION from
// pricing.quote_accepted, the same way pricing.Listing is a projection from
// catalog. Ordering never asks pricing for a price; it is TOLD, once, when the
// customer said yes. Plain exported fields: a record, not an aggregate.
//
// [PHP] Bảng riêng của bundle Ordering, MessageHandler cập nhật khi nhận
// [PHP] QuoteAccepted từ Pricing. Không join sang bảng quotes.
type AcceptedQuote struct {
	Quote   shared.ID
	Product shared.ID
	Total   shared.Money
	Deposit shared.Money
}
