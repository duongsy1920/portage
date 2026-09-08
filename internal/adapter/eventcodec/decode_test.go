package eventcodec_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/adapter/eventcodec"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The read side of the Published Language: a consumer in another context
// gets a V1 struct — never a catalog type, which it may not import (guard 7).
// Encode → Decode must hand back exactly what was announced.
func TestDecode_roundTripsWhatEncodeWrote(t *testing.T) {
	pub := catalog.ProductPublished{ID: pid, Merchant: mid, Category: catalog.MustParseCategoryCode("footwear"),
		Name: "Air Trainer 90", Source: catalog.MustParseSourceURL("https://www.example.com/t/x"),
		Price: usd("160.00"), Parcel: box, At: at}
	got := decode(t, pub).(contracts.ProductPublishedV1)
	if got.Source != "https://www.example.com/t/x" {
		t.Errorf("ProductPublishedV1.Source = %q", got.Source)
	}
	if got.ID != pid.String() || got.Merchant != mid.String() || got.Category != "footwear" || got.Name != "Air Trainer 90" ||
		got.Price != (contracts.MoneyV1{Minor: 16000, Currency: "USD"}) ||
		got.Parcel != (contracts.ParcelV1{WeightG: 1250, LengthMM: 340, WidthMM: 230, HeightMM: 130}) ||
		!got.At.Equal(at) {
		t.Errorf("ProductPublishedV1 = %+v", got)
	}

	meas := decode(t, catalog.ProductMeasured{ID: pid, Parcel: box, Verified: true, At: at}).(contracts.ProductMeasuredV1)
	if meas.ID != pid.String() || !meas.Verified || meas.Parcel.WeightG != 1250 {
		t.Errorf("ProductMeasuredV1 = %+v", meas)
	}

	rep := decode(t, catalog.ProductRepriced{ID: pid, From: usd("150.00"), To: usd("160.00"), At: at}).(contracts.ProductRepricedV1)
	if rep.From.Minor != 15000 || rep.To.Minor != 16000 || rep.To.Currency != "USD" {
		t.Errorf("ProductRepricedV1 = %+v", rep)
	}

	ret := decode(t, catalog.ProductRetired{ID: pid, Reason: "discontinued", At: at}).(contracts.ProductRetiredV1)
	if ret.ID != pid.String() || ret.Reason != "discontinued" {
		t.Errorf("ProductRetiredV1 = %+v", ret)
	}

	// A size with spaces and a slash must arrive byte for byte: the buyer reads it.
	va := decode(t, catalog.VariantAdded{ID: pid, Variant: catalog.NewVariantID(),
		Size: "M 8 / W 9.5", Color: "black", MerchantRef: "EX-AT90-8-BLK", At: at}).(contracts.VariantAddedV1)
	if va.Product != pid.String() || va.Size != "M 8 / W 9.5" || va.Color != "black" || va.MerchantRef != "EX-AT90-8-BLK" {
		t.Errorf("VariantAddedV1 = %+v", va)
	}

	def := decode(t, catalog.CategoryDefined{Code: catalog.MustParseCategoryCode("footwear"), Estimate: box,
		Restrictions: []catalog.Restriction{catalog.RestrictionMagnet}, At: at}).(contracts.CategoryDefinedV1)
	if def.Code != "footwear" || def.Estimate.WeightG != 1250 || len(def.Restrictions) != 1 || def.Restrictions[0] != "magnet" {
		t.Errorf("CategoryDefinedV1 = %+v", def)
	}

	// pricing → ordering
	acc := decode(t, pricing.QuoteAcceptedEvent{ID: qid, Product: pid.ID, TotalVND: vnd("5393720"), Deposit: vnd("2696860"), At: at}).(contracts.QuoteAcceptedV1)
	if acc.ID != qid.String() || acc.Product != pid.String() || acc.Total != (contracts.MoneyV1{Minor: 5393720, Currency: "VND"}) || acc.Deposit.Minor != 2696860 {
		t.Errorf("QuoteAcceptedV1 = %+v", acc)
	}

	// catalog → procurement (the shop's currency), ordering → procurement, procurement → ordering
	reg := decode(t, catalog.MerchantRegistered{ID: mid, Name: "Example Sports", Site: host, Currency: shared.USD, At: at}).(contracts.MerchantRegisteredV1)
	if reg.ID != mid.String() || reg.Currency != "USD" || reg.Site != "www.example.com" {
		t.Errorf("MerchantRegisteredV1 = %+v", reg)
	}
	dep := decode(t, ordering.DepositPaid{ID: oid2, Quote: qid.ID, Product: pid.ID, Variant: vrt, Amount: vnd("2696860"), At: at}).(contracts.DepositPaidV1)
	if dep.ID != oid2.String() || dep.Product != pid.String() || dep.Variant != vrt.String() || dep.Amount.Minor != 2696860 {
		t.Errorf("DepositPaidV1 = %+v", dep)
	}
	conf := decode(t, procurement.PurchaseConfirmed{ID: tid, Order: oid2.ID, Reference: "NK-1", Paid: usd("163.22"), By: op, At: at}).(contracts.PurchaseConfirmedV1)
	if conf.Order != oid2.String() || conf.Reference != "NK-1" || conf.Paid.Minor != 16322 || conf.By != op.String() {
		t.Errorf("PurchaseConfirmedV1 = %+v", conf)
	}
	fail := decode(t, procurement.PurchaseFailed{ID: tid, Order: oid2.ID, Reason: "sold out", At: at}).(contracts.PurchaseFailedV1)
	if fail.Order != oid2.String() || fail.Reason != "sold out" {
		t.Errorf("PurchaseFailedV1 = %+v", fail)
	}

	// pricing → logistics, ordering → pricing, logistics → ordering + pricing
	ln := decode(t, pricing.LaneDefinedEvent{Code: lane, Name: "US forwarder", Divisor: 5000, Step: shared.Grams(500), Currency: shared.USD, At: at}).(contracts.LaneDefinedV1)
	if ln.Code != "us_forwarder" || ln.Divisor != 5000 || ln.StepG != 500 || ln.Currency != "USD" {
		t.Errorf("LaneDefinedV1 = %+v", ln)
	}
	pl := decode(t, ordering.OrderPlaced{ID: oid2, Quote: qid.ID, Product: pid.ID, Variant: vrt, Customer: cust, Total: vnd("5393720"), Deposit: vnd("2696860"), At: at}).(contracts.OrderPlacedV1)
	if pl.ID != oid2.String() || pl.Quote != qid.String() || pl.Total.Minor != 5393720 {
		t.Errorf("OrderPlacedV1 = %+v", pl)
	}
	sh := decode(t, logistics.BatchShippedEvent{ID: bid, Lane: "us_forwarder", Freight: usd("87.50"),
		Allocations: []logistics.Allocation{{Parcel: prc.ID, Order: oid2.ID, Chargeable: shared.Grams(2500), Freight: usd("29.17")}}, At: at}).(contracts.BatchShippedV1)
	if sh.ID != bid.String() || sh.Freight.Minor != 8750 || len(sh.Allocations) != 1 || sh.Allocations[0].Order != oid2.String() || sh.Allocations[0].ChargeableG != 2500 || sh.Allocations[0].Freight.Minor != 2917 {
		t.Errorf("BatchShippedV1 = %+v", sh)
	}
}

func decode(t *testing.T, ev shared.Event) any {
	t.Helper()
	env, err := eventcodec.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := eventcodec.Decode(env.Name, env.Payload)
	if err != nil {
		t.Fatalf("Decode(%s): %v", env.Name, err)
	}
	return msg
}

// Decoders exist for the events some consumer needs — not for every event.
// An event without a decoder is not an error of the event; it is a message
// nobody reads yet, and Decode says so with ErrNoDecoder.
func TestDecode_unknownAndMalformed(t *testing.T) {
	if _, err := eventcodec.Decode("catalog.merchant_renamed", []byte(`{}`)); !errors.Is(err, eventcodec.ErrNoDecoder) {
		t.Errorf("no consumer reads merchant_renamed yet: got %v", err)
	}
	if _, err := eventcodec.Decode("nobody.ever", []byte(`{}`)); !errors.Is(err, eventcodec.ErrNoDecoder) {
		t.Errorf("unknown name: got %v", err)
	}
	if _, err := eventcodec.Decode("catalog.product_published", []byte(`{"id": `)); err == nil {
		t.Error("malformed payload must be an error")
	}
}
