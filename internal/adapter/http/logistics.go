package httpapi

import (
	"net/http"
	"time"

	logisticsapp "github.com/duongsy/portage/internal/app/logistics"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The logistics routes are the WAREHOUSE's screen: what is coming, weigh it,
// box it, ship the box with the carrier's invoice. Parcels are never created
// here — procurement.purchase_confirmed creates them through the relay.

type parcelView struct {
	ID         string     `json:"id"`
	OrderID    string     `json:"order_id"`
	Reference  string     `json:"reference"`
	Status     string     `json:"status"`
	Actual     *parcelOut `json:"actual,omitempty"`
	BatchID    string     `json:"batch_id,omitempty"`
	ExpectedAt time.Time  `json:"expected_at"`
	ReceivedAt *time.Time `json:"received_at,omitempty"`
}

type parcelOut struct {
	WeightG  int64 `json:"weight_g"`
	LengthMM int64 `json:"length_mm"`
	WidthMM  int64 `json:"width_mm"`
	HeightMM int64 `json:"height_mm"`
}

func parcelOutOf(s shared.ParcelSpec) *parcelOut {
	if s.IsZero() {
		return nil
	}
	d := s.Dimensions()
	return &parcelOut{WeightG: s.Weight().Grams(), LengthMM: d.LengthMM(), WidthMM: d.WidthMM(), HeightMM: d.HeightMM()}
}

func parcelViewOf(p *logistics.Parcel) parcelView {
	v := parcelView{ID: p.ID().String(), OrderID: p.Order().String(), Reference: p.Reference(), Status: string(p.Status()), Actual: parcelOutOf(p.Actual()), ExpectedAt: p.ExpectedAt()}
	if !p.Batch().IsZero() {
		v.BatchID = p.Batch().String()
	}
	if !p.ReceivedAt().IsZero() {
		at := p.ReceivedAt()
		v.ReceivedAt = &at
	}
	return v
}

// GET /parcels → 200 [parcelView]: everything not yet shipped, oldest first.
func (s *server) listPendingParcels(w http.ResponseWriter, r *http.Request) {
	parcels, err := s.parcels.Pending(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]parcelView, 0, len(parcels))
	for _, p := range parcels {
		out = append(out, parcelViewOf(p))
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /parcels/{id} → 200 parcelView.
func (s *server) getParcel(w http.ResponseWriter, r *http.Request) {
	id, err := logistics.ParseParcelID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.parcels.ByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, parcelViewOf(p))
}

// POST /parcels/{id}/receive {weight_g, length_mm, width_mm, height_mm} → 204.
// The operator is the bearer token: who put it on the scale.
func (s *server) receiveParcel(w http.ResponseWriter, r *http.Request) {
	id, err := logistics.ParseParcelID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req parcelRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	spec, err := req.spec()
	if err != nil {
		writeError(w, err)
		return
	}
	operator, ok := operatorOf(r)
	if !ok {
		// requireOperator already guaranteed this; the check stays so that
		// moving the route out from behind it fails loudly instead of
		// recording the zero operator.
		writeError(w, errForbidden)
		return
	}
	if err := s.receive.Handle(r.Context(), logisticsapp.ReceiveParcel{Parcel: id, Actual: spec, Operator: operator}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type openBatchRequest struct {
	Lane string `json:"lane"`
}

// POST /batches {"lane"} → 201 {"id"}.
func (s *server) openBatch(w http.ResponseWriter, r *http.Request) {
	var req openBatchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	id, err := s.openBatch_.Handle(r.Context(), req.Lane)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, idResponse{ID: id.String()})
}

type allocationView struct {
	ParcelID    string    `json:"parcel_id"`
	OrderID     string    `json:"order_id"`
	ChargeableG int64     `json:"chargeable_g"`
	Freight     moneyView `json:"freight"`
}

type batchView struct {
	ID          string           `json:"id"`
	Lane        string           `json:"lane"`
	Status      string           `json:"status"`
	Parcels     []string         `json:"parcels"`
	Freight     *moneyView       `json:"freight,omitempty"`
	Allocations []allocationView `json:"allocations,omitempty"`
	OpenedAt    time.Time        `json:"opened_at"`
	ClosedAt    *time.Time       `json:"closed_at,omitempty"`
	ShippedAt   *time.Time       `json:"shipped_at,omitempty"`
}

// GET /batches/{id} → 200 batchView.
func (s *server) getBatch(w http.ResponseWriter, r *http.Request) {
	id, err := logistics.ParseBatchID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	b, err := s.batches.ByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	v := batchView{ID: b.ID().String(), Lane: b.Lane(), Status: string(b.Status()), Parcels: []string{}, OpenedAt: b.OpenedAt()}
	for _, it := range b.Items() {
		v.Parcels = append(v.Parcels, it.Parcel.String())
	}
	if b.Freight().IsValid() {
		f := viewOf(b.Freight())
		v.Freight = &f
	}
	for _, a := range b.Allocations() {
		v.Allocations = append(v.Allocations, allocationView{ParcelID: a.Parcel.String(), OrderID: a.Order.String(), ChargeableG: a.Chargeable.Grams(), Freight: viewOf(a.Freight)})
	}
	if !b.ClosedAt().IsZero() {
		at := b.ClosedAt()
		v.ClosedAt = &at
	}
	if !b.ShippedAt().IsZero() {
		at := b.ShippedAt()
		v.ShippedAt = &at
	}
	writeJSON(w, http.StatusOK, v)
}

type addParcelRequest struct {
	ParcelID string `json:"parcel_id"`
}

// POST /batches/{id}/parcels {"parcel_id"} → 204.
func (s *server) addParcelToBatch(w http.ResponseWriter, r *http.Request) {
	id, err := logistics.ParseBatchID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req addParcelRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	parcel, err := logistics.ParseParcelID(req.ParcelID)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.addParcel.Handle(r.Context(), logisticsapp.AddParcelToBatch{Batch: id, Parcel: parcel}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /batches/{id}/close → 204.
func (s *server) closeBatch(w http.ResponseWriter, r *http.Request) {
	id, err := logistics.ParseBatchID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.closeBatch_.Handle(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type shipBatchRequest struct {
	Freight  string `json:"freight"`
	Currency string `json:"currency"`
}

// POST /batches/{id}/ship {"freight","currency"} → 200 [allocationView]: the
// split is returned because the packer wants to see it right now — and it is
// also frozen in the event ordering and pricing consume.
func (s *server) shipBatch(w http.ResponseWriter, r *http.Request) {
	id, err := logistics.ParseBatchID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req shipBatchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	cur, err := shared.CurrencyFromCode(req.Currency)
	if err != nil {
		writeError(w, err)
		return
	}
	freight, err := shared.ParseMoney(normalizeAmount(req.Freight, language(r)), cur)
	if err != nil {
		writeError(w, err)
		return
	}
	allocs, err := s.shipBatch_.Handle(r.Context(), logisticsapp.ShipBatch{Batch: id, Freight: freight})
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]allocationView, 0, len(allocs))
	for _, a := range allocs {
		out = append(out, allocationView{ParcelID: a.Parcel.String(), OrderID: a.Order.String(), ChargeableG: a.Chargeable.Grams(), Freight: viewOf(a.Freight)})
	}
	writeJSON(w, http.StatusOK, out)
}
