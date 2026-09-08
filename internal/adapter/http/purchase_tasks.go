package httpapi

import (
	"net/http"
	"time"

	procurementapp "github.com/duongsy/portage/internal/app/procurement"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The procurement routes are the BUYER's screen: what is there to buy
// (open tasks), and closing a task with what happened at the shop. Tasks are
// never created here — ordering.deposit_paid creates them through the relay.

type taskView struct {
	ID        string `json:"id"`
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`

	// What to buy, in the shop's own words, frozen when the task opened.
	// Without these the buyer's screen is four uuids and a currency, and
	// nobody can walk into a shop with that (procurement.Subject).
	ProductName  string `json:"product_name,omitempty"`
	VariantLabel string `json:"variant_label,omitempty"` // "M 8 / W 9.5 · black"; empty for a one-size product
	VariantRef   string `json:"variant_ref,omitempty"`   // the shop's own code, when it has one
	Source       string `json:"source,omitempty"`        // the page to buy from

	Currency  string     `json:"currency"`
	Status    string     `json:"status"`
	Reference string     `json:"reference,omitempty"`
	Paid      *moneyView `json:"paid,omitempty"`
	Reason    string     `json:"reason,omitempty"`
	OpenedAt  time.Time  `json:"opened_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
}

func taskViewOf(t *procurement.PurchaseTask) taskView {
	subject := t.Subject()
	v := taskView{
		ID:        t.ID().String(),
		OrderID:   t.Order().String(),
		ProductID: t.Product().String(),
		VariantID: t.Variant().String(),

		ProductName:  subject.ProductName,
		VariantLabel: subject.VariantLabel,
		VariantRef:   subject.VariantRef,
		Source:       subject.Source,

		Currency:  t.Currency().Code(),
		Status:    string(t.Status()),
		Reference: t.Receipt().Reference,
		Reason:    t.Reason(),
		OpenedAt:  t.OpenedAt(),
	}
	if t.Receipt().Paid.IsValid() {
		paid := viewOf(t.Receipt().Paid)
		v.Paid = &paid
	}
	if !t.ClosedAt().IsZero() {
		closed := t.ClosedAt()
		v.ClosedAt = &closed
	}
	return v
}

// GET /purchase-tasks → 200 [taskView]: the open ones, oldest first.
func (s *server) listOpenTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.tasks.Open(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]taskView, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, taskViewOf(t))
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /purchase-tasks/{id} → 200 taskView.
func (s *server) getTask(w http.ResponseWriter, r *http.Request) {
	id, err := procurement.ParseTaskID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	t, err := s.tasks.ByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskViewOf(t))
}

type confirmTaskRequest struct {
	Reference string `json:"reference"`
	Paid      string `json:"paid"`
	Currency  string `json:"currency"`
}

// POST /purchase-tasks/{id}/confirm {"reference","paid","currency"} → 204.
// The bearer token says who bought; the domain refuses without an operator.
func (s *server) confirmTask(w http.ResponseWriter, r *http.Request) {
	id, err := procurement.ParseTaskID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req confirmTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	cur, err := shared.CurrencyFromCode(req.Currency)
	if err != nil {
		writeError(w, err)
		return
	}
	paid, err := shared.ParseMoney(normalizeAmount(req.Paid, language(r)), cur)
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
	if err := s.confirmTask_.Handle(r.Context(), procurementapp.ConfirmTask{
		Task: id, Receipt: procurement.PurchaseReceipt{Reference: req.Reference, Paid: paid, PaidBy: operator},
	}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type failTaskRequest struct {
	Reason string `json:"reason"`
}

// POST /purchase-tasks/{id}/fail {"reason"} → 204.
func (s *server) failTask(w http.ResponseWriter, r *http.Request) {
	id, err := procurement.ParseTaskID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req failTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := s.failTask_.Handle(r.Context(), procurementapp.FailTask{Task: id, Reason: req.Reason}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
