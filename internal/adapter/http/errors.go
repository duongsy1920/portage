package httpapi

import (
	"errors"
	"log"
	"net/http"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

var (
	errBadJSON        = errors.New("malformed json")
	errInvalidRequest = errors.New("invalid request")

	// Revoking the last operator key would leave an API nobody can
	// administer — recoverable only by editing the database by hand.
	errLastOperator = errors.New("the last active operator key cannot be revoked")
)

// errorResponse is the body of every non-2xx answer. code is a STABLE slug a
// client can switch on; message is the domain's own text, for humans and logs.
type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type mapping struct {
	target error
	status int
	code   string
}

// errorTable is the ONE place a domain or app error becomes an HTTP answer.
// Three conversations with the client, three status codes:
//
//	400  the request cannot be understood — fix the request
//	404  the request points at something that does not exist
//	409  the request is well-formed and the BUSINESS says no — fix the state
//
// Anything not in the table is a bug or an outage: 500, generic message,
// details in the log only.
//
// [PHP] Một ExceptionListener duy nhất với bảng exception → status, thay cho
// [PHP] rải rác try/catch trong từng Controller.
var errorTable = []mapping{
	// 400 — cannot be understood
	{errBadJSON, http.StatusBadRequest, "bad_json"},
	{errInvalidRequest, http.StatusBadRequest, "invalid_request"},
	{errLastOperator, http.StatusConflict, "last_operator_key"},
	{shared.ErrMalformedAmount, http.StatusBadRequest, "malformed_amount"},
	{shared.ErrUnknownCurrency, http.StatusBadRequest, "unknown_currency"},
	{shared.ErrInvalidID, http.StatusBadRequest, "invalid_id"},
	{catalog.ErrInvalidHostname, http.StatusBadRequest, "invalid_hostname"},
	{catalog.ErrInvalidCategoryCode, http.StatusBadRequest, "invalid_category_code"},
	{catalog.ErrInvalidSourceURL, http.StatusBadRequest, "invalid_source_url"},
	{catalog.ErrEmptyName, http.StatusBadRequest, "empty_name"},
	{catalog.ErrEmptyReason, http.StatusBadRequest, "empty_reason"},
	{catalog.ErrUnknownSourcing, http.StatusBadRequest, "unknown_sourcing"},
	{catalog.ErrCurrencyRequired, http.StatusBadRequest, "currency_required"},
	{catalog.ErrFreeShipCurrency, http.StatusBadRequest, "free_shipping_currency"},
	{catalog.ErrNegativeThreshold, http.StatusBadRequest, "negative_threshold"},
	{catalog.ErrPriceRequired, http.StatusBadRequest, "price_required"},
	{catalog.ErrNegativePrice, http.StatusBadRequest, "negative_price"},
	{catalog.ErrOperatorRequired, http.StatusBadRequest, "operator_required"},
	{catalog.ErrProvenanceRequired, http.StatusBadRequest, "provenance_required"},
	{catalog.ErrMerchantRequired, http.StatusBadRequest, "merchant_required"},
	{shared.ErrIncompleteParcelSpec, http.StatusBadRequest, "incomplete_parcel_spec"},
	{catalog.ErrUnknownRestriction, http.StatusBadRequest, "unknown_restriction"},
	{shared.ErrNegativeWeight, http.StatusBadRequest, "negative_weight"},
	{pricing.ErrInvalidLaneCode, http.StatusBadRequest, "invalid_lane_code"},
	{ordering.ErrEmptyReason, http.StatusBadRequest, "empty_reason"},
	{ordering.ErrInvalidOrder, http.StatusBadRequest, "invalid_order"},
	{procurement.ErrEmptyReason, http.StatusBadRequest, "empty_reason"},
	{procurement.ErrEmptyReference, http.StatusBadRequest, "empty_reference"},
	{procurement.ErrInvalidTask, http.StatusBadRequest, "invalid_purchase_task"},
	// Two sentinels, two codes: pricing.ErrInvalidRate is a price in a rate
	// card, shared.ErrInvalidRate is an exchange rate (POST /fx). Same name,
	// different packages and different fixes — one code for both would tell
	// the client to look in the wrong field.
	{pricing.ErrInvalidRate, http.StatusBadRequest, "invalid_rate"},
	{shared.ErrInvalidRate, http.StatusBadRequest, "invalid_exchange_rate"},
	{pricing.ErrInvalidLane, http.StatusBadRequest, "invalid_lane"},
	{logistics.ErrInvalidParcel, http.StatusBadRequest, "invalid_parcel"},
	{logistics.ErrInvalidBatch, http.StatusBadRequest, "invalid_batch"},
	// POST /tokens: the kind is validated by auth.NewPrincipal, so the adapter
	// keeps no list of kinds of its own to fall out of step with.
	{auth.ErrUnknownKind, http.StatusBadRequest, "unknown_principal_kind"},
	{auth.ErrNoSubject, http.StatusBadRequest, "no_subject"},

	// 404 — points at nothing
	{catalog.ErrMerchantNotFound, http.StatusNotFound, "merchant_not_found"},
	{catalog.ErrCategoryNotFound, http.StatusNotFound, "category_not_found"},
	{catalog.ErrProductNotFound, http.StatusNotFound, "product_not_found"},
	{pricing.ErrLaneNotFound, http.StatusNotFound, "lane_not_found"},
	{pricing.ErrQuoteNotFound, http.StatusNotFound, "quote_not_found"},
	{pricing.ErrListingNotFound, http.StatusNotFound, "listing_not_found"}, // not (yet) published, or the relay has not run
	{ordering.ErrOrderNotFound, http.StatusNotFound, "order_not_found"},
	{procurement.ErrTaskNotFound, http.StatusNotFound, "purchase_task_not_found"},
	{pricing.ErrReconciliationNotFound, http.StatusNotFound, "reconciliation_not_found"},
	{logistics.ErrParcelNotFound, http.StatusNotFound, "parcel_not_found"},
	{logistics.ErrBatchNotFound, http.StatusNotFound, "batch_not_found"},
	{logistics.ErrLaneRuleNotFound, http.StatusNotFound, "lane_rule_not_found"},

	// 409 — well-formed, refused by the business
	{catalog.ErrNotDraft, http.StatusConflict, "not_draft"},
	{catalog.ErrNoVariants, http.StatusConflict, "no_variants"},
	{catalog.ErrUnverified, http.StatusConflict, "unverified"},
	{catalog.ErrSuspectedDuplicate, http.StatusConflict, "suspected_duplicate"},
	{catalog.ErrDuplicateVariant, http.StatusConflict, "duplicate_variant"},
	{catalog.ErrInvalidDuplicate, http.StatusConflict, "invalid_duplicate"},
	{catalogapp.ErrMerchantInactive, http.StatusConflict, "merchant_inactive"},
	{catalogapp.ErrPriceCurrency, http.StatusConflict, "price_currency"},
	{shared.ErrCurrencyMismatch, http.StatusConflict, "currency_mismatch"},
	{pricing.ErrListingInactive, http.StatusConflict, "listing_inactive"},
	{pricing.ErrNothingToWeigh, http.StatusConflict, "nothing_to_weigh"},
	{pricing.ErrNoExchangeRate, http.StatusConflict, "no_exchange_rate"},
	{pricing.ErrQuoteExpired, http.StatusConflict, "quote_expired"},
	{pricing.ErrQuoteNotIssued, http.StatusConflict, "quote_not_issued"},
	{pricing.ErrQuoteStillValid, http.StatusConflict, "quote_still_valid"},
	{ordering.ErrQuoteNotAccepted, http.StatusConflict, "quote_not_accepted"}, // not accepted, or accepted and not yet relayed
	{ordering.ErrQuoteAlreadyUsed, http.StatusConflict, "quote_already_used"},
	// A size the customer did not get from us, or one that belongs to another
	// shoe. "unknown" is retryable: catalog.variant_added may not be relayed yet.
	{ordering.ErrVariantUnknown, http.StatusConflict, "variant_unknown"},
	{ordering.ErrVariantNotForProduct, http.StatusConflict, "variant_not_for_product"},
	{catalog.ErrUnnamedVariant, http.StatusConflict, "unnamed_variant"},
	{ordering.ErrWrongAmount, http.StatusConflict, "wrong_amount"},
	{ordering.ErrNotAwaitingDeposit, http.StatusConflict, "not_awaiting_deposit"},
	{ordering.ErrNotDeposited, http.StatusConflict, "not_deposited"},
	{ordering.ErrNotPurchased, http.StatusConflict, "not_purchased"},
	{ordering.ErrNotInTransit, http.StatusConflict, "not_in_transit"},
	{ordering.ErrBalanceUnpaid, http.StatusConflict, "balance_unpaid"},
	{ordering.ErrAlreadyDelivered, http.StatusConflict, "already_delivered"},
	{ordering.ErrOrderCancelled, http.StatusConflict, "order_cancelled"},
	{procurement.ErrTaskNotOpen, http.StatusConflict, "purchase_task_not_open"},
	{procurement.ErrPaidCurrency, http.StatusConflict, "paid_currency"},
	{logistics.ErrParcelNotExpected, http.StatusConflict, "parcel_not_expected"},
	{logistics.ErrParcelNotReceived, http.StatusConflict, "parcel_not_received"},
	{logistics.ErrParcelNotBatched, http.StatusConflict, "parcel_not_batched"},
	{logistics.ErrBatchNotOpen, http.StatusConflict, "batch_not_open"},
	{logistics.ErrBatchNotClosed, http.StatusConflict, "batch_not_closed"},
	{logistics.ErrBatchEmpty, http.StatusConflict, "batch_empty"},
	{logistics.ErrDuplicateParcel, http.StatusConflict, "duplicate_parcel"},
	{logistics.ErrInvalidFreight, http.StatusConflict, "invalid_freight"},

	// 401 / 403 — who is calling (auth.go). Two codes because they are two
	// different conversations: 401 says "I do not know you", 403 says "I know
	// you, this is not yours". Collapsing them would make the API an oracle
	// for guessing which tokens are real.
	{auth.ErrUnauthenticated, http.StatusUnauthorized, "unauthenticated"},
	{errForbidden, http.StatusForbidden, "forbidden"},

	// A customer asking after somebody else's order is told exactly what a
	// customer asking after a non-existent order is told. 404, not 403: a 403
	// would confirm the order exists (P9-PLAN §4, câu 1).
	{ordering.ErrNotOwner, http.StatusNotFound, "order_not_found"},

	// 503 — the request is fine, the help is not there. A feature that is
	// off must LOOK off: falling back to a fake extractor would answer 201
	// with an invented name and price (P9-PLAN §4, câu 6).
	{catalogapp.ErrExtractorUnavailable, http.StatusServiceUnavailable, "extractor_unavailable"},
}

// writeError answers with the mapped status and code. errors.Is walks the
// %w chain, so a wrapped sentinel ("product \"x\": product has no variants")
// still finds its row.
func writeError(w http.ResponseWriter, err error) {
	var body errorResponse
	for _, m := range errorTable {
		if errors.Is(err, m.target) {
			body.Error.Code, body.Error.Message = m.code, err.Error()
			writeJSON(w, m.status, body)
			return
		}
	}
	log.Printf("httpapi: unmapped error: %v", err)
	body.Error.Code, body.Error.Message = "internal", "internal error"
	writeJSON(w, http.StatusInternalServerError, body)
}
