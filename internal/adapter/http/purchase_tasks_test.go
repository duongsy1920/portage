package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

func (a *api) seedOpenTask(t *testing.T) *procurement.PurchaseTask {
	t.Helper()
	task, err := procurement.OpenTask(procurement.TaskDetails{
		Order:    shared.NewID(),
		Product:  shared.NewID(),
		Variant:  shared.NewID(),
		Currency: shared.USD,
		// What the buyer reads. OpenTask copies it as given: the app layer is
		// what builds it, from procurement's own variant projection.
		Subject: procurement.Subject{
			ProductName:  "Air Trainer 90",
			VariantLabel: "M 8 / W 9.5 · black",
			VariantRef:   "EX-AT90-8-BLK",
			Source:       taskSource,
		},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	task.PullEvents()
	if err := a.tasks.Save(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	return task
}

const taskSource = "https://www.example.com/t/air-trainer-90/abc"

func TestPurchaseTasks_listConfirmFail(t *testing.T) {
	a := newAPI()
	first, second := a.seedOpenTask(t), a.seedOpenTask(t)
	operator := map[string]string{"X-Operator-ID": shared.NewOperatorID().String()}

	rec := a.call(t, "GET", "/purchase-tasks", "", nil)
	var list []struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		ProductName string `json:"product_name"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || rec.Code != http.StatusOK || len(list) != 2 {
		t.Fatalf("GET /purchase-tasks = %d %s (%v)", rec.Code, rec.Body, err)
	}
	if list[0].ProductName != "Air Trainer 90" || list[1].ProductName != "Air Trainer 90" {
		t.Fatalf("the buyer's list must name the product, not only its id: %+v", list)
	}

	id := first.ID().String()
	expectError(t, a.call(t, "GET", "/purchase-tasks/"+procurement.NewTaskID().String(), "", nil), 404, "purchase_task_not_found")
	expectError(t, a.call(t, "POST", "/purchase-tasks/"+id+"/confirm", `{"reference":"NK-1","paid":"163.22","currency":"USD"}`, a.asCustomer()), 403, "forbidden")
	expectError(t, a.call(t, "POST", "/purchase-tasks/"+id+"/confirm", `{"reference":"","paid":"163.22","currency":"USD"}`, operator), 400, "empty_reference")
	expectError(t, a.call(t, "POST", "/purchase-tasks/"+id+"/confirm", `{"reference":"NK-1","paid":"4000000","currency":"VND"}`, operator), 409, "paid_currency")
	a.outbox.Drain()
	if rec := a.call(t, "POST", "/purchase-tasks/"+id+"/confirm", `{"reference":"NK-1","paid":"163.22","currency":"USD"}`, operator); rec.Code != http.StatusNoContent {
		t.Fatalf("confirm = %d %s", rec.Code, rec.Body)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "procurement.purchase_confirmed" {
		t.Fatalf("outbox = %v", evs)
	}
	expectError(t, a.call(t, "POST", "/purchase-tasks/"+id+"/confirm", `{"reference":"NK-1","paid":"163.22","currency":"USD"}`, operator), 409, "purchase_task_not_open")
	var v struct {
		Status       string `json:"status"`
		Reference    string `json:"reference"`
		Paid         *money `json:"paid"`
		ProductName  string `json:"product_name"`
		VariantLabel string `json:"variant_label"`
		VariantRef   string `json:"variant_ref"`
		Source       string `json:"source"`
	}
	_ = json.Unmarshal(a.call(t, "GET", "/purchase-tasks/"+id, "", nil).Body.Bytes(), &v)
	if v.Status != "confirmed" || v.Reference != "NK-1" || v.Paid == nil || v.Paid.Amount != "163.22" {
		t.Fatalf("view = %+v", v)
	}
	// A buyer can act on this: a name, a size, the shop's code and the page.
	if v.ProductName != "Air Trainer 90" || v.VariantLabel != "M 8 / W 9.5 · black" || v.VariantRef != "EX-AT90-8-BLK" || v.Source != taskSource {
		t.Fatalf("subject in the view = %+v", v)
	}

	id2 := second.ID().String()
	expectError(t, a.call(t, "POST", "/purchase-tasks/"+id2+"/fail", `{"reason":" "}`, nil), 400, "empty_reason")
	if rec := a.call(t, "POST", "/purchase-tasks/"+id2+"/fail", `{"reason":"sold out"}`, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("fail = %d %s", rec.Code, rec.Body)
	}
	_ = json.Unmarshal(a.call(t, "GET", "/purchase-tasks", "", nil).Body.Bytes(), &list)
	if len(list) != 0 {
		t.Fatalf("still open: %v", list)
	}
}
