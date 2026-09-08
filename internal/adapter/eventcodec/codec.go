// Package eventcodec turns a domain event into the bytes that go into the
// outbox — and therefore onto the wire to other contexts.
//
// The payload is a CONTRACT (Published Language, DDD.md §7): a consumer in
// another context cannot load our aggregate (guard 7 forbids the import), so
// what is in the JSON is all it will ever know. That is why every key here is
// written by hand, exactly like EventName() is — deriving the shape from the
// Go struct with json.Marshal would let a field rename change the wire format
// while everything still compiles. Here, a changed event means a mapper that
// does not compile, and a new event means a red test
// (TestEncode_contractCoversEveryDomainEvent) until someone writes its row.
//
// Only Encode exists. Decode arrives with the first consumer that needs it.
//
// [PHP] Đây là Serializer normalizer viết tay cho từng class event — thay cho
// [PHP] việc để Serializer phản chiếu (reflection) property. Cùng lý do người
// [PHP] ta dùng Normalizer riêng khi JSON là hợp đồng với hệ thống khác.
package eventcodec

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

var ErrUnknownEvent = errors.New("unknown event type")

// Envelope is one outbox row before it is written: the stable name, when it
// happened, and the payload as JSON.
type Envelope struct {
	Name       string
	OccurredAt time.Time
	Payload    json.RawMessage
}

// m is a payload object. Keys are the contract; values are primitives or
// nested m — never a domain type, so json.Marshal has nothing to guess.
type m = map[string]any

// Encode maps one event to its envelope. An event this package does not know
// is an error: an unknown shape on the wire is worse than an error in the log.
func Encode(ev shared.Event) (Envelope, error) {
	if ev == nil {
		return Envelope{}, fmt.Errorf("encode nil event: %w", ErrUnknownEvent)
	}
	body, err := payload(ev)
	if err != nil {
		return Envelope{}, err
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return Envelope{}, fmt.Errorf("encode %s: %w", ev.EventName(), err)
	}
	return Envelope{Name: ev.EventName(), OccurredAt: ev.OccurredAt(), Payload: raw}, nil
}

// payload is the contract, one case per event. Text value objects go through
// their canonical String(); the three structured ones have helpers below.
//
// [PHP] `switch e := ev.(type)` = type switch — thay cho instanceof theo chuỗi.
// [PHP] Bên trong mỗi case, `e` đã có kiểu cụ thể, không cần ép kiểu thêm.
func payload(ev shared.Event) (m, error) {
	switch e := ev.(type) {
	// ── Merchant ──────────────────────────────────────────────────────────
	case catalog.MerchantRegistered:
		return m{"id": e.ID.String(), "name": e.Name, "site": e.Site.String(), "currency": e.Currency.Code(), "at": ts(e.At)}, nil
	case catalog.MerchantRenamed:
		return m{"id": e.ID.String(), "from": e.From, "to": e.To, "at": ts(e.At)}, nil
	case catalog.MerchantSourcingEnabled:
		return m{"id": e.ID.String(), "mode": string(e.Mode), "at": ts(e.At)}, nil
	case catalog.MerchantSourcingDisabled:
		return m{"id": e.ID.String(), "mode": string(e.Mode), "at": ts(e.At)}, nil
	case catalog.MerchantFreeShippingChanged:
		return m{"id": e.ID.String(), "from": freeShipping(e.From), "to": freeShipping(e.To), "at": ts(e.At)}, nil
	case catalog.MerchantSuspended:
		return m{"id": e.ID.String(), "reason": e.Reason, "at": ts(e.At)}, nil
	case catalog.MerchantReinstated:
		return m{"id": e.ID.String(), "at": ts(e.At)}, nil

	// ── Product ───────────────────────────────────────────────────────────
	case catalog.ProductAdded:
		return m{"id": e.ID.String(), "merchant": e.Merchant.String(), "category": e.Category.String(),
			"name": e.Name, "source": e.Source.String(), "price": money(e.Price),
			"sourced_by": string(e.SourcedBy), "requested_by": idOrEmpty(e.RequestedBy), "at": ts(e.At)}, nil
	case catalog.ListingConfirmed:
		return m{"id": e.ID.String(), "by": e.By.String(), "at": ts(e.At)}, nil
	case catalog.ProductMeasured:
		return m{"id": e.ID.String(), "parcel": parcel(e.Parcel), "verified": e.Verified, "at": ts(e.At)}, nil
	case catalog.ProductRepriced:
		return m{"id": e.ID.String(), "from": money(e.From), "to": money(e.To), "at": ts(e.At)}, nil
	case catalog.ProductFlaggedDuplicate:
		return m{"id": e.ID.String(), "of": e.Of.String(), "reason": e.Reason, "at": ts(e.At)}, nil
	case catalog.ProductDuplicateCleared:
		return m{"id": e.ID.String(), "of": e.Of.String(), "reason": e.Reason, "at": ts(e.At)}, nil
	case catalog.ProductPublished:
		return m{"id": e.ID.String(), "merchant": e.Merchant.String(), "category": e.Category.String(),
			"name": e.Name, "source": e.Source.String(), "price": money(e.Price),
			"parcel": parcel(e.Parcel), "at": ts(e.At)}, nil
	case catalog.VariantAdded:
		return m{"product": e.ID.String(), "variant": e.Variant.String(), "size": e.Size,
			"color": e.Color, "merchant_ref": e.MerchantRef, "at": ts(e.At)}, nil
	case catalog.ProductRetired:
		return m{"id": e.ID.String(), "reason": e.Reason, "at": ts(e.At)}, nil

	// ── pricing ───────────────────────────────────────────────────────────
	case pricing.QuoteIssuedEvent:
		return m{"id": e.ID.String(), "product": e.Product.String(), "lane": e.Lane.String(),
			"total": money(e.TotalVND), "deposit": money(e.Deposit), "estimated": e.Estimated,
			"expires_at": ts(e.ExpiresAt), "at": ts(e.At)}, nil
	case pricing.QuoteAcceptedEvent:
		return m{"id": e.ID.String(), "product": e.Product.String(), "total": money(e.TotalVND), "deposit": money(e.Deposit), "at": ts(e.At)}, nil
	case pricing.QuoteExpiredEvent:
		return m{"id": e.ID.String(), "product": e.Product.String(), "at": ts(e.At)}, nil
	case pricing.LaneDefinedEvent:
		return m{"code": e.Code.String(), "name": e.Name, "divisor": e.Divisor, "step_g": e.Step.Grams(), "currency": e.Currency.Code(), "at": ts(e.At)}, nil

	// ── Ordering ──────────────────────────────────────────────────────────
	case ordering.OrderPlaced:
		return m{"id": e.ID.String(), "quote": e.Quote.String(), "product": e.Product.String(), "variant": e.Variant.String(),
			"customer": e.Customer.String(), "total": money(e.Total), "deposit": money(e.Deposit), "at": ts(e.At)}, nil
	case ordering.DepositPaid:
		return m{"id": e.ID.String(), "quote": e.Quote.String(), "product": e.Product.String(), "variant": e.Variant.String(),
			"amount": money(e.Amount), "at": ts(e.At)}, nil
	case ordering.OrderPurchased:
		return m{"id": e.ID.String(), "at": ts(e.At)}, nil
	case ordering.OrderPurchaseFailed:
		return m{"id": e.ID.String(), "reason": e.Reason, "at": ts(e.At)}, nil
	case ordering.OrderShipped:
		return m{"id": e.ID.String(), "at": ts(e.At)}, nil
	case ordering.BalancePaid:
		return m{"id": e.ID.String(), "amount": money(e.Amount), "at": ts(e.At)}, nil
	case ordering.OrderDelivered:
		return m{"id": e.ID.String(), "at": ts(e.At)}, nil
	case ordering.OrderCancelled:
		return m{"id": e.ID.String(), "reason": e.Reason, "refund": money(e.Refund), "forfeited": e.Forfeited, "at": ts(e.At)}, nil

	// ── Procurement ───────────────────────────────────────────────────────
	case procurement.PurchaseTaskOpened:
		return m{"id": e.ID.String(), "order": e.Order.String(), "product": e.Product.String(), "variant": e.Variant.String(), "at": ts(e.At)}, nil
	case procurement.PurchaseConfirmed:
		return m{"id": e.ID.String(), "order": e.Order.String(), "reference": e.Reference, "paid": money(e.Paid), "by": e.By.String(), "at": ts(e.At)}, nil
	case procurement.PurchaseFailed:
		return m{"id": e.ID.String(), "order": e.Order.String(), "reason": e.Reason, "at": ts(e.At)}, nil

	// ── Logistics ─────────────────────────────────────────────────────────
	case logistics.ParcelExpectedEvent:
		return m{"id": e.ID.String(), "order": e.Order.String(), "reference": e.Reference, "at": ts(e.At)}, nil
	case logistics.ParcelReceivedEvent:
		return m{"id": e.ID.String(), "order": e.Order.String(), "actual": parcel(e.Actual), "by": e.By.String(), "at": ts(e.At)}, nil
	case logistics.BatchOpened:
		return m{"id": e.ID.String(), "lane": e.Lane, "at": ts(e.At)}, nil
	case logistics.BatchClosedEvent:
		return m{"id": e.ID.String(), "parcels": e.Parcels, "at": ts(e.At)}, nil
	case logistics.BatchShippedEvent:
		allocs := make([]m, 0, len(e.Allocations))
		for _, a := range e.Allocations {
			allocs = append(allocs, m{"parcel": a.Parcel.String(), "order": a.Order.String(), "chargeable_g": a.Chargeable.Grams(), "freight": money(a.Freight)})
		}
		return m{"id": e.ID.String(), "lane": e.Lane, "freight": money(e.Freight), "allocations": allocs, "at": ts(e.At)}, nil

	// ── Category ──────────────────────────────────────────────────────────
	case catalog.CategoryDefined:
		restrictions := make([]string, 0, len(e.Restrictions))
		for _, r := range e.Restrictions {
			restrictions = append(restrictions, string(r))
		}
		return m{"code": e.Code.String(), "estimate": parcel(e.Estimate), "restrictions": restrictions, "at": ts(e.At)}, nil

	default:
		return nil, fmt.Errorf("encode %T (%s): %w", ev, ev.EventName(), ErrUnknownEvent)
	}
}

// ts is the one time format on the wire: RFC 3339, UTC, nanoseconds when any.
// idOrEmpty keeps a zero id out of the wire as "" rather than a uuid of all
// zeros. A consumer that parses "00000000-0000-0000-0000-000000000000" gets a
// perfectly valid id for a customer who does not exist, and then addresses a
// notification to nobody. Same rule the reporting repository follows with NULL.
func idOrEmpty(id shared.ID) string {
	if id.IsZero() {
		return ""
	}
	return id.String()
}

func ts(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// money: minor units and the ISO code — the same two things Money stores.
func money(x shared.Money) m {
	return m{"minor": x.Minor(), "currency": x.Currency().Code()}
}

// parcel: grams and millimetres, exactly as ParcelSpec holds them.
func parcel(p shared.ParcelSpec) m {
	d := p.Dimensions()
	return m{"weight_g": p.Weight().Grams(), "length_mm": d.LengthMM(), "width_mm": d.WidthMM(), "height_mm": d.HeightMM()}
}

// freeShipping: the kind, plus the threshold only when there is one.
func freeShipping(f catalog.FreeShipping) m {
	out := m{"kind": string(f.Kind())}
	if threshold, ok := f.Threshold(); ok {
		out["threshold"] = money(threshold)
	}
	return out
}
