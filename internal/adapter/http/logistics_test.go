package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

func (a *api) seedLogistics(t *testing.T) (shoe, jacket *logistics.Parcel) {
	t.Helper()
	ctx := context.Background()
	if err := a.laneRules.Save(ctx, logistics.LaneRule{Code: "us_forwarder", Divisor: 5000, Step: shared.Grams(500)}); err != nil {
		t.Fatal(err)
	}
	mk := func(ref string) *logistics.Parcel {
		p, err := logistics.ExpectParcel(logistics.ParcelDetails{Order: shared.NewID(), Reference: ref}, now)
		if err != nil {
			t.Fatal(err)
		}
		p.PullEvents()
		if err := a.parcels.Save(ctx, p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	return mk("NK-1"), mk("TNF-9")
}

func TestLogistics_receiveBoxShip(t *testing.T) {
	a := newAPI()
	shoe, jacket := a.seedLogistics(t)
	operator := map[string]string{"X-Operator-ID": shared.NewOperatorID().String()}

	var list []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(a.call(t, "GET", "/parcels", "", nil).Body.Bytes(), &list)
	if len(list) != 2 || list[0].Status != "expected" {
		t.Fatalf("pending = %+v", list)
	}

	shoeBox := `{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}`
	expectError(t, a.call(t, "POST", "/parcels/"+shoe.ID().String()+"/receive", shoeBox, a.asCustomer()), 403, "forbidden")
	expectError(t, a.call(t, "POST", "/parcels/"+logistics.NewParcelID().String()+"/receive", shoeBox, operator), 404, "parcel_not_found")
	if rec := a.call(t, "POST", "/parcels/"+shoe.ID().String()+"/receive", shoeBox, operator); rec.Code != http.StatusNoContent {
		t.Fatalf("receive = %d %s", rec.Code, rec.Body)
	}
	expectError(t, a.call(t, "POST", "/parcels/"+shoe.ID().String()+"/receive", shoeBox, operator), 409, "parcel_not_expected")
	if rec := a.call(t, "POST", "/parcels/"+jacket.ID().String()+"/receive", `{"weight_g":900,"length_mm":400,"width_mm":320,"height_mm":180}`, operator); rec.Code != http.StatusNoContent {
		t.Fatalf("receive jacket = %d %s", rec.Code, rec.Body)
	}

	expectError(t, a.call(t, "POST", "/batches", `{"lane":"sea_freight"}`, nil), 404, "lane_rule_not_found")
	batch := idOf(t, a.call(t, "POST", "/batches", `{"lane":"us_forwarder"}`, nil))
	expectError(t, a.call(t, "POST", "/batches/"+batch+"/close", "", nil), 409, "batch_empty")
	for _, p := range []*logistics.Parcel{shoe, jacket} {
		if rec := a.call(t, "POST", "/batches/"+batch+"/parcels", `{"parcel_id":"`+p.ID().String()+`"}`, nil); rec.Code != http.StatusNoContent {
			t.Fatalf("add = %d %s", rec.Code, rec.Body)
		}
	}
	expectError(t, a.call(t, "POST", "/batches/"+batch+"/parcels", `{"parcel_id":"`+shoe.ID().String()+`"}`, nil), 409, "duplicate_parcel")
	expectError(t, a.call(t, "POST", "/batches/"+batch+"/ship", `{"freight":"87.50","currency":"USD"}`, nil), 409, "batch_not_closed")
	if rec := a.call(t, "POST", "/batches/"+batch+"/close", "", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("close = %d %s", rec.Code, rec.Body)
	}
	expectError(t, a.call(t, "POST", "/batches/"+batch+"/ship", `{"freight":"0","currency":"USD"}`, nil), 409, "invalid_freight")
	a.outbox.Drain()
	rec := a.call(t, "POST", "/batches/"+batch+"/ship", `{"freight":"87,50","currency":"USD"}`, map[string]string{"Accept-Language": "vi"})
	if rec.Code != http.StatusOK {
		t.Fatalf("ship = %d %s", rec.Code, rec.Body)
	}
	var allocs []struct {
		OrderID     string `json:"order_id"`
		ChargeableG int64  `json:"chargeable_g"`
		Freight     money  `json:"freight"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &allocs)
	if len(allocs) != 2 || allocs[0].OrderID != shoe.Order().String() || allocs[0].ChargeableG != 2500 || allocs[0].Freight.Amount != "29.17" || allocs[1].Freight.Amount != "58.33" {
		t.Fatalf("allocations = %+v", allocs)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "logistics.batch_shipped" {
		t.Fatalf("outbox = %v", evs)
	}
	var bv struct {
		Status  string `json:"status"`
		Freight *money `json:"freight"`
	}
	_ = json.Unmarshal(a.call(t, "GET", "/batches/"+batch, "", nil).Body.Bytes(), &bv)
	if bv.Status != "shipped" || bv.Freight == nil || bv.Freight.Amount != "87.50" {
		t.Fatalf("batch view = %+v", bv)
	}
	_ = json.Unmarshal(a.call(t, "GET", "/parcels", "", nil).Body.Bytes(), &list)
	if len(list) != 0 {
		t.Fatalf("still pending: %+v", list)
	}
	expectError(t, a.call(t, "GET", "/batches/"+logistics.NewBatchID().String(), "", nil), 404, "batch_not_found")
}

func TestLanesAndReconciliation(t *testing.T) {
	a := newAPI()
	body := `{"code":"sea_freight","name":"Sea, 6 weeks","divisor":6000,"step_g":1000,"currency":"USD",
		"rates_per_kg":{"standard":"4.00","branded":"4.50","electronics":"5.00","sensitive":"6.00"},"duty_rate_percent":"5"}`
	a.outbox.Drain()
	if rec := a.call(t, "POST", "/lanes", body, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("define lane = %d %s", rec.Code, rec.Body)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "pricing.lane_defined" {
		t.Fatalf("outbox = %v", evs)
	}
	lane, err := a.lanes.ByCode(context.Background(), pricing.MustParseLaneCode("sea_freight"))
	if err != nil || lane.Divisor() != 6000 || !lane.DutyPolicy().Itemised() {
		t.Fatalf("lane = %+v, %v", lane, err)
	}
	expectError(t, a.call(t, "POST", "/lanes", `{"code":"Sea Freight","currency":"USD"}`, nil), 400, "invalid_lane_code")
	expectError(t, a.call(t, "POST", "/lanes", `{"code":"x","name":"x","divisor":5000,"step_g":500,"currency":"USD","rates_per_kg":{"standard":"0","branded":"1","electronics":"1","sensitive":"1"}}`, nil), 400, "invalid_rate")

	order := shared.NewID()
	expectError(t, a.call(t, "GET", "/reconciliations/"+order.String(), "", nil), 404, "reconciliation_not_found")
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	if err := a.reconciliations.Save(context.Background(), pricing.Reconciliation{Order: order, Quote: shared.NewID(),
		QuotedGoods: usd("163.22"), QuotedFreight: usd("25.00"), QuotedChargeable: shared.Grams(2500),
		ActualGoods: usd("163.22"), ActualFreight: usd("27.50"), ActualChargeable: shared.Grams(2500)}); err != nil {
		t.Fatal(err)
	}
	var v struct {
		Complete bool   `json:"complete"`
		Variance *money `json:"variance"`
	}
	_ = json.Unmarshal(a.call(t, "GET", "/reconciliations/"+order.String(), "", nil).Body.Bytes(), &v)
	if !v.Complete || v.Variance == nil || v.Variance.Amount != "-2.50" {
		t.Fatalf("reconciliation view = %+v", v)
	}
}
