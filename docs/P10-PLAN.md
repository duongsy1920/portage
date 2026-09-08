# P10-PLAN.md — đặt hộ có ghi tên (operator accept + place)

**Quyết định đã chốt 06/09/2026.** Console operator phải làm trọn vòng, không
rơi ra curl ở giữa. Nghĩa là hai route đang `requireCustomer` phải mở cho
operator — nhưng **có ghi tên người đặt hộ**, không giả làm khách.

Đây là việc mở lại một quyết định đã đóng có chủ đích ở P9/T1. Đọc
WALKTHROUGH §18e để biết vì sao nó từng bị đóng, rồi mới sửa.

---

## 1. Hai điều đã kiểm trong code — đừng làm theo mô tả cũ

**`OnBehalfOf` ở `CancelOrder` KHÔNG phải tiền lệ đặt hộ.** Nó là **kiểm quyền
sở hữu**: adapter điền nó từ token **của chính khách**, và để zero khi operator
gọi. Khách tự giới hạn mình, không phải nhân viên khai tên người khác. Ngược
hướng hoàn toàn — đừng copy hình dạng của nó.

**Danh tính khách chỉ vào hệ thống ở MỘT chỗ.**

```
pricing.Quote          id · product · lane · breakdown · issuedAt · expiresAt · status   ← không có customer
ordering.AcceptedQuote quote · product · total · deposit                                  ← không có customer
ordering.PlaceOrder    quote · variant · CUSTOMER                                         ← ĐÂY, lấy từ token
```

Hệ quả: **accept không cần biết khách là ai.** Nó chỉ là "đồng ý cái giá này".
Nên đừng thêm attribution vào `pricing` — không có gì để ghi tên vào cả.

---

## 2. Thiết kế

### 2a. Attribution nằm ở ĐƠN, không ở báo giá

Một field, một cột, một chỗ:

```go
// ordering/order.go
type OrderDetails struct {
	Quote    shared.ID
	Product  shared.ID
	Variant  shared.ID
	Customer shared.ID
	Total    shared.Money
	Deposit  shared.Money

	// PlacedBy: nhân viên đã đặt hộ. ZERO nghĩa là khách tự đặt bằng chìa của
	// chính họ — đó là hành vi có từ trước, nên mọi dòng cũ đã đúng sẵn,
	// không cần backfill.
	PlacedBy shared.OperatorID
}
```

Vì sao ở đơn chứ không ở báo giá: đơn là **cam kết** (có tiền đi kèm), báo giá
chỉ là một cái giá. "Ai bấm ok trên một cái giá" là bằng chứng yếu hơn nhiều so
với "ai tạo ra cái đơn", và đằng nào sau đó còn một khoản cọc 50 % chuyển khoản
làm bằng chứng mạnh nhất.

Zero là "khách tự đặt" — hợp quy ước 9 (zero value phải an toàn), và đúng nghĩa
đen: chưa từng có ai đặt hộ thì không có tên nào để ghi.

### 2b. Luật giữ lại lời hứa của T1

T1 xoá `customer_id` khỏi body với lời hứa: *"không còn field nào để đặt hộ tên
người khác"*. P10 đưa field đó trở lại, nên lời hứa phải được phát biểu lại:

> **`customer_id` trong body chỉ operator đọc được. Khách gửi nó là 403.**

```go
// adapter/http/orders.go — POST /orders, giờ là requireAny
if customer, isCustomer := customerOf(r); isCustomer {
	// Khách đặt cho chính mình. Gửi kèm customer_id là đang thử đặt hộ
	// người khác — từ chối THẲNG, đừng im lặng bỏ qua: im lặng nghĩa là
	// client tưởng nó có tác dụng.
	if req.CustomerID != "" {
		writeError(w, errForbidden)
		return
	}
	cmd.Customer = customer
} else {
	operator, _ := operatorOf(r)
	if req.CustomerID == "" {
		writeError(w, catalog.ErrCustomerRequired) // 400 customer_required
		return
	}
	cmd.Customer = parse(req.CustomerID)
	cmd.PlacedBy = operator          // ← ghi tên
}
```

**Test bắt buộc, không có nó thì cả thay đổi này là một lỗ hổng:**

```
TestPlaceOrder_customerCannotOrderInSomebodyElsesName
    chìa khách A + customer_id của khách B  →  403, và KHÔNG có đơn nào được tạo
```

### 2c. Accept: đổi đúng một dòng

```go
mux.HandleFunc("POST /quotes/{id}/accept", requireAny(s.acceptQuote))   // was requireCustomer
```

Không thêm attribution. `Quote` không có chỗ để ghi, và thêm chỗ nghĩa là một
migration + đổi wire format cho một thông tin mà đơn hàng ngay sau đó đã ghi
tốt hơn.

Cập nhật lại comment ở `auth_test.go`
(`TestAuth_customerOnlyRoutesRefuseOperator`) — sau P10 chỉ còn **`POST /orders`
với khách** là ca "khách đặt cho chính mình", còn accept mở cho cả hai.

---

## 3. File phải sửa

| # | File | Sửa gì |
|---|---|---|
| 1 | `domain/ordering/order.go` | `OrderDetails.PlacedBy` · field `placedBy` · `PlacedBy()` · `OrderSnapshot.PlacedBy` · `FromSnapshot` |
| 2 | `domain/ordering/events.go` | `OrderPlaced.PlacedBy` |
| 3 | `app/ordering/place_order.go` | `PlaceOrder.PlacedBy`, truyền xuống `OrderDetails` |
| 4 | `adapter/eventcodec/codec.go` | payload `order_placed` thêm `placed_by` — **xem bẫy §4** |
| 5 | `contracts/ordering_v1.go` | `OrderPlacedV1.PlacedBy string` |
| 6 | `adapter/postgres/migrations/0008_placed_by.sql` | `ALTER TABLE orders ADD COLUMN placed_by uuid` (nullable, không default) |
| 7 | `adapter/postgres/ordering_repos.go` | thêm cột vào `orderColumns` · Save ghi `nil` khi zero · scan `*string` |
| 8 | `adapter/http/orders.go` | `placeOrderRequest.CustomerID` + nhánh ở §2b |
| 9 | `adapter/http/server.go` | 2 route `requireCustomer` → `requireAny` |
| 10 | `adapter/http/errors.go` | sentinel + mã `customer_required` → 400 |

**Tuỳ chọn** (nếu muốn hiện lên màn hình): `app/reporting/summary.go` +
`projector.go` + `0007`-style cột `placed_by`, và `adapter/http/summaries.go`.
Nên hiện cho **cả hai** loại người xem — giấu ai đặt hộ đơn của chính mình thì
đi ngược lại toàn bộ lý do làm việc này.

---

## 4. Ba cái bẫy đã biết

**Zero UUID xuống DB và xuống event.** `shared.OperatorID{}.String()` ra một
UUID toàn số 0. Ghi nguyên nó vào cột `placed_by` hay vào payload event thì đọc
lên nó thành một id **thật** — và bỗng dưng có một nhân viên tên
`00000000-…` đặt hộ mọi đơn khách tự đặt. Đây đúng là lỗi đã gặp ở read model
P9/T2 (`idOrNil`). Xử lý:

```go
// codec.go — bỏ hẳn key khi zero, đừng ghi chuỗi rỗng cũng đừng ghi UUID 0
m := map[string]any{"id": …, "customer": …}
if !e.PlacedBy.IsZero() {
	m["placed_by"] = e.PlacedBy.String()
}

// ordering_repos.go — nil, không phải chuỗi
var placedBy *string
if !s.PlacedBy.IsZero() { v := s.PlacedBy.String(); placedBy = &v }
```

**Migration là `ADD COLUMN` nullable, không backfill.** Mọi đơn cũ đều là khách
tự đặt, và NULL đã nói đúng điều đó. Backfill bằng bất cứ giá trị nào là ghi
một chuyện không xảy ra vào sổ.

**`OrderSnapshot` đổi hình dạng.** `FromSnapshot` có test từ chối 8 hình dạng
hỏng (`order_test.go`) — thêm field phải kiểm lại danh sách đó còn đúng không,
và `postgres` round trip so bằng `reflect.DeepEqual(Snapshot())` nên sẽ đỏ ngay
nếu quên một đầu.

---

## 5. Test phải có

| Test | Ở đâu | Canh cái gì |
|---|---|---|
| `TestPlaceOrder_customerCannotOrderInSomebodyElsesName` | `adapter/http` | **quan trọng nhất** — 403, và không đơn nào được tạo |
| `TestPlaceOrder_operatorMustNameTheCustomer` | `adapter/http` | operator không gửi `customer_id` → 400 `customer_required` |
| `TestPlaceOrder_operatorPlacedOrderRecordsWhoDidIt` | `app/ordering` | `PlacedBy` = operator; khách tự đặt → `PlacedBy` zero |
| `TestOrderRepo_roundTripByIDAndByQuote` | `postgres` | thêm ca `placed_by` NULL **và** có giá trị |
| `TestAuth_customerOnlyRoutesRefuseOperator` | `adapter/http` | sửa: accept giờ nhận cả hai |
| `wholeFlow` | `platform/wire` | thêm một đơn operator đặt hộ; số vàng **không được đổi** |

Xong phải xanh: `gofmt -l .` rỗng · `go vet ./...` · **304 + 5 test** ·
7 decision guard · `smoke.ps1` in đúng `total 5393720`, `variance -2.50`.

---

## 6. Ghi lại sau khi xong

`docs/SETUP.md` §9 **Đợt 17** — bảng quyết định/bug như 16 đợt trước, và phải
có một dòng nói thẳng: *đây là lần mở lại một quyết định đã đóng, lý do là gì,
và lời hứa của T1 giờ được phát biểu lại thành câu nào.*

Cập nhật thêm: `WALKTHROUGH §18e` (đoạn "3 assertion đổi thành 403" không còn
đúng) · `DDD §30` bảng tầng · `README` bảng route · `FLOW-ORDER §3 bước 10`.
