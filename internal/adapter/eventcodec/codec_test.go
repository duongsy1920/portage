package eventcodec_test

import (
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/eventcodec"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	at   = time.Date(2026, 9, 4, 17, 0, 0, 0, time.UTC)
	atS  = "2026-09-04T17:00:00Z"
	mid  = catalog.NewMerchantID()
	pid  = catalog.NewProductID()
	oid  = catalog.NewProductID()
	host = catalog.MustParseHostname("www.example.com")
	usd  = func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	over = catalog.MustFreeShippingOver(usd("50.00"))
	box  = shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))
	vid  = catalog.NewVariantID()
	qid  = pricing.NewQuoteID()
	vnd  = func(s string) shared.Money { return shared.MustParseMoney(s, shared.VND) }
	lane = pricing.MustParseLaneCode("us_forwarder")
	oid2 = ordering.NewOrderID()
	cust = shared.NewID()
	vrt  = shared.NewID()
	tid  = procurement.NewTaskID()
	op   = shared.NewOperatorID()
	prc  = logistics.NewParcelID()
	bid  = logistics.NewBatchID()
)

type m = map[string]any

// The payload IS the contract (Published Language, DDD.md §7): a consumer in
// another context cannot load our aggregate — guard 7 forbids the import — so
// what is in the JSON is all it will ever know. Every key below is therefore
// written by hand, like EventName() is, and this table is the contract's text.
var contract = []struct {
	ev   shared.Event
	want m
}{
	{catalog.MerchantRegistered{ID: mid, Name: "Example Sports", Site: host, Currency: shared.USD, At: at},
		m{"id": mid.String(), "name": "Example Sports", "site": "www.example.com", "currency": "USD", "at": atS}},
	{catalog.MerchantRenamed{ID: mid, From: "A", To: "B", At: at},
		m{"id": mid.String(), "from": "A", "to": "B", "at": atS}},
	{catalog.MerchantSourcingEnabled{ID: mid, Mode: catalog.SourcedByCustomer, At: at},
		m{"id": mid.String(), "mode": "customer", "at": atS}},
	{catalog.MerchantSourcingDisabled{ID: mid, Mode: catalog.SourcedByFeed, At: at},
		m{"id": mid.String(), "mode": "feed", "at": atS}},
	{catalog.MerchantFreeShippingChanged{ID: mid, From: catalog.NoFreeShipping(), To: over, At: at},
		m{"id": mid.String(), "from": m{"kind": "never"}, "to": m{"kind": "over", "threshold": m{"minor": 5000.0, "currency": "USD"}}, "at": atS}},
	{catalog.MerchantSuspended{ID: mid, Reason: "banned", At: at},
		m{"id": mid.String(), "reason": "banned", "at": atS}},
	{catalog.MerchantReinstated{ID: mid, At: at},
		m{"id": mid.String(), "at": atS}},
	{catalog.ProductAdded{ID: pid, Merchant: mid, Category: catalog.MustParseCategoryCode("footwear"), Name: "Air Trainer 90", At: at},
		m{"id": pid.String(), "merchant": mid.String(), "category": "footwear", "name": "Air Trainer 90", "at": atS}},
	{catalog.ProductMeasured{ID: pid, Parcel: box, Verified: true, At: at},
		m{"id": pid.String(), "parcel": m{"weight_g": 1250.0, "length_mm": 340.0, "width_mm": 230.0, "height_mm": 130.0}, "verified": true, "at": atS}},
	{catalog.ProductRepriced{ID: pid, From: usd("150.00"), To: usd("160.00"), At: at},
		m{"id": pid.String(), "from": m{"minor": 15000.0, "currency": "USD"}, "to": m{"minor": 16000.0, "currency": "USD"}, "at": atS}},
	{catalog.ProductFlaggedDuplicate{ID: pid, Of: oid, Reason: "same page url", At: at},
		m{"id": pid.String(), "of": oid.String(), "reason": "same page url", "at": atS}},
	{catalog.ProductDuplicateCleared{ID: pid, Of: oid, Reason: "checked", At: at},
		m{"id": pid.String(), "of": oid.String(), "reason": "checked", "at": atS}},
	{catalog.ProductPublished{ID: pid, Merchant: mid, Category: catalog.MustParseCategoryCode("footwear"), Name: "Air Trainer 90",
		Source: catalog.MustParseSourceURL("https://www.example.com/t/air-trainer-90/abc"), Price: usd("160.00"), Parcel: box, At: at},
		m{"id": pid.String(), "merchant": mid.String(), "category": "footwear", "name": "Air Trainer 90",
			"source": "https://www.example.com/t/air-trainer-90/abc",
			"price":  m{"minor": 16000.0, "currency": "USD"},
			"parcel": m{"weight_g": 1250.0, "length_mm": 340.0, "width_mm": 230.0, "height_mm": 130.0}, "at": atS}},
	// A size as a shop writes it, slashes and all — the payload must not touch it.
	{catalog.VariantAdded{ID: pid, Variant: vid, Size: "M 8 / W 9.5", Color: "black", MerchantRef: "EX-AT90-8-BLK", At: at},
		m{"product": pid.String(), "variant": vid.String(), "size": "M 8 / W 9.5", "color": "black",
			"merchant_ref": "EX-AT90-8-BLK", "at": atS}},
	{catalog.CategoryDefined{Code: catalog.MustParseCategoryCode("footwear"), Estimate: box, Restrictions: []catalog.Restriction{catalog.RestrictionMagnet}, At: at},
		m{"code": "footwear", "estimate": m{"weight_g": 1250.0, "length_mm": 340.0, "width_mm": 230.0, "height_mm": 130.0}, "restrictions": []any{"magnet"}, "at": atS}},
	{catalog.ProductRetired{ID: pid, Reason: "discontinued", At: at},
		m{"id": pid.String(), "reason": "discontinued", "at": atS}},

	// ── pricing ──────────────────────────────────────────────────────────────
	{pricing.QuoteIssuedEvent{ID: qid, Product: pid.ID, Lane: lane, TotalVND: vnd("5328720"), Deposit: vnd("2664360"), Estimated: false, ExpiresAt: at.Add(48 * time.Hour), At: at},
		m{"id": qid.String(), "product": pid.String(), "lane": "us_forwarder", "total": m{"minor": 5328720.0, "currency": "VND"},
			"deposit": m{"minor": 2664360.0, "currency": "VND"}, "estimated": false, "expires_at": "2026-09-06T17:00:00Z", "at": atS}},
	{pricing.QuoteAcceptedEvent{ID: qid, Product: pid.ID, TotalVND: vnd("5328720"), Deposit: vnd("2664360"), At: at},
		m{"id": qid.String(), "product": pid.String(), "total": m{"minor": 5328720.0, "currency": "VND"}, "deposit": m{"minor": 2664360.0, "currency": "VND"}, "at": atS}},
	{pricing.QuoteExpiredEvent{ID: qid, Product: pid.ID, At: at},
		m{"id": qid.String(), "product": pid.String(), "at": atS}},

	// ── ordering ─────────────────────────────────────────────────────────────
	{ordering.OrderPlaced{ID: oid2, Quote: qid.ID, Product: pid.ID, Variant: vrt, Customer: cust, Total: vnd("5393720"), Deposit: vnd("2696860"), At: at},
		m{"id": oid2.String(), "quote": qid.String(), "product": pid.String(), "variant": vrt.String(), "customer": cust.String(),
			"total": m{"minor": 5393720.0, "currency": "VND"}, "deposit": m{"minor": 2696860.0, "currency": "VND"}, "at": atS}},
	{ordering.DepositPaid{ID: oid2, Quote: qid.ID, Product: pid.ID, Variant: vrt, Amount: vnd("2696860"), At: at},
		m{"id": oid2.String(), "quote": qid.String(), "product": pid.String(), "variant": vrt.String(), "amount": m{"minor": 2696860.0, "currency": "VND"}, "at": atS}},
	{ordering.OrderPurchased{ID: oid2, At: at}, m{"id": oid2.String(), "at": atS}},
	{ordering.OrderPurchaseFailed{ID: oid2, Reason: "sold out", At: at}, m{"id": oid2.String(), "reason": "sold out", "at": atS}},
	{ordering.OrderShipped{ID: oid2, At: at}, m{"id": oid2.String(), "at": atS}},
	{ordering.BalancePaid{ID: oid2, Amount: vnd("2696860"), At: at}, m{"id": oid2.String(), "amount": m{"minor": 2696860.0, "currency": "VND"}, "at": atS}},
	{ordering.OrderDelivered{ID: oid2, At: at}, m{"id": oid2.String(), "at": atS}},
	{ordering.OrderCancelled{ID: oid2, Reason: "customer asked", Refund: vnd("2696860"), Forfeited: false, At: at},
		m{"id": oid2.String(), "reason": "customer asked", "refund": m{"minor": 2696860.0, "currency": "VND"}, "forfeited": false, "at": atS}},

	// ── procurement ──────────────────────────────────────────────────────────
	{procurement.PurchaseTaskOpened{ID: tid, Order: oid2.ID, Product: pid.ID, Variant: vrt, At: at},
		m{"id": tid.String(), "order": oid2.String(), "product": pid.String(), "variant": vrt.String(), "at": atS}},
	{procurement.PurchaseConfirmed{ID: tid, Order: oid2.ID, Reference: "NK-123456", Paid: usd("163.22"), By: op, At: at},
		m{"id": tid.String(), "order": oid2.String(), "reference": "NK-123456", "paid": m{"minor": 16322.0, "currency": "USD"}, "by": op.String(), "at": atS}},
	{procurement.PurchaseFailed{ID: tid, Order: oid2.ID, Reason: "sold out", At: at},
		m{"id": tid.String(), "order": oid2.String(), "reason": "sold out", "at": atS}},

	// ── pricing (lane) + logistics ───────────────────────────────────────────
	{pricing.LaneDefinedEvent{Code: lane, Name: "US forwarder", Divisor: 5000, Step: shared.Grams(500), Currency: shared.USD, At: at},
		m{"code": "us_forwarder", "name": "US forwarder", "divisor": 5000.0, "step_g": 500.0, "currency": "USD", "at": atS}},
	{logistics.ParcelExpectedEvent{ID: prc, Order: oid2.ID, Reference: "NK-1", At: at},
		m{"id": prc.String(), "order": oid2.String(), "reference": "NK-1", "at": atS}},
	{logistics.ParcelReceivedEvent{ID: prc, Order: oid2.ID, Actual: box, By: op, At: at},
		m{"id": prc.String(), "order": oid2.String(), "actual": m{"weight_g": 1250.0, "length_mm": 340.0, "width_mm": 230.0, "height_mm": 130.0}, "by": op.String(), "at": atS}},
	{logistics.BatchOpened{ID: bid, Lane: "us_forwarder", At: at}, m{"id": bid.String(), "lane": "us_forwarder", "at": atS}},
	{logistics.BatchClosedEvent{ID: bid, Parcels: 2, At: at}, m{"id": bid.String(), "parcels": 2.0, "at": atS}},
	{logistics.BatchShippedEvent{ID: bid, Lane: "us_forwarder", Freight: usd("87.50"),
		Allocations: []logistics.Allocation{{Parcel: prc.ID, Order: oid2.ID, Chargeable: shared.Grams(2500), Freight: usd("29.17")}}, At: at},
		m{"id": bid.String(), "lane": "us_forwarder", "freight": m{"minor": 8750.0, "currency": "USD"},
			"allocations": []any{m{"parcel": prc.String(), "order": oid2.String(), "chargeable_g": 2500.0, "freight": m{"minor": 2917.0, "currency": "USD"}}}, "at": atS}},
}

func TestEncode_matchesTheContract(t *testing.T) {
	for _, c := range contract {
		env, err := eventcodec.Encode(c.ev)
		if err != nil {
			t.Errorf("%s: %v", c.ev.EventName(), err)
			continue
		}
		if env.Name != c.ev.EventName() || !env.OccurredAt.Equal(at) {
			t.Errorf("%s: envelope = %s at %v", c.ev.EventName(), env.Name, env.OccurredAt)
		}
		var got m
		if err := json.Unmarshal(env.Payload, &got); err != nil {
			t.Errorf("%s: payload is not JSON: %v: %s", c.ev.EventName(), err, env.Payload)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s:\n got  %v\n want %v", c.ev.EventName(), got, c.want)
		}
	}
}

type strayEvent struct{}

func (strayEvent) EventName() string     { return "stray" }
func (strayEvent) OccurredAt() time.Time { return at }

// An event the codec does not know is a bug at the boundary, not something to
// serialise "as best we can" — an unknown shape on the wire is worse than an
// error in the log.
func TestEncode_rejectsUnknownEvent(t *testing.T) {
	if _, err := eventcodec.Encode(strayEvent{}); !errors.Is(err, eventcodec.ErrUnknownEvent) {
		t.Fatalf("got %v, want ErrUnknownEvent", err)
	}
	if _, err := eventcodec.Encode(nil); err == nil {
		t.Fatal("nil event must be an error")
	}
}

// The table above is only a contract if it is COMPLETE. This reads every
// context's events.go with go/ast, collects every type that has an EventName
// method, and demands a row for each — add an event, forget the mapper, and
// this goes red without anyone having to remember.
func TestEncode_contractCoversEveryDomainEvent(t *testing.T) {
	files, err := filepath.Glob("../../domain/*/events.go") // catalog, pricing, ... — every context
	if err != nil || len(files) == 0 {
		t.Fatalf("found no events.go — wrong path? (%v)", err)
	}
	fset := token.NewFileSet()
	declared := map[string]bool{}
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name.Name != "EventName" {
				continue
			}
			if id, ok := fn.Recv.List[0].Type.(*ast.Ident); ok {
				declared[id.Name] = true
			}
		}
	}
	if len(declared) < 15 {
		t.Fatalf("found only %d event types — wrong path?", len(declared))
	}
	covered := map[string]bool{}
	for _, c := range contract {
		covered[reflect.TypeOf(c.ev).Name()] = true
	}
	for name := range declared {
		if !covered[name] {
			t.Errorf("event %s has no row in the contract table (and probably no mapper)", name)
		}
	}
	for name := range covered {
		if !declared[name] {
			t.Errorf("contract table names %s, which events.go does not declare", name)
		}
	}
}
