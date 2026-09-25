# UI-NEXT-PLAN.md — đợt sau của giao diện: đưa trí nhớ của màn hình về server

> Plan viết 23/09/2026, ngay sau đợt 21 (làm lại giao diện, `docs/UI-REDESIGN-PLAN.md`),
> **chưa làm gì**. Người thực hiện có thể là một agent khác: file này phải đủ để bắt tay làm mà
> không đọc lại cuộc trò chuyện. Anh Sỹ chốt các mục ở **§7**. Mục nào chưa trả lời thì làm theo
> **đề nghị** ghi ở đó.

---

## 0. Đọc trước

| Đọc | Vì sao |
|---|---|
| `CLAUDE.md` | luật làm việc; mục **Giao diện** tóm đợt 21 |
| `docs/SETUP.md` §9 đợt 21 | ba lượt của đợt trước, mỗi bug tìm ra và cách sửa |
| `design-system/portage/MASTER.md` §5 | luật của dải hành trình: ô chỉ ghi số API thật sự gửi |
| `docs/UI-GUIDE.md` §3 (d)(e), §5 | luồng kho, và ba dòng "hỏng thì xem" nói về chỗ màn hình chỉ nhớ trong phiên |

---

## 1. Vấn đề: màn hình đang nhớ thay cho server

Đợt 21 không được đụng Go, nên ba chỗ thiếu của API được vá **ở trình duyệt**:

| Màn hình đang làm | Vì API thiếu | Hậu quả |
|---|---|---|
| nhớ các việc đi mua đã xong, các kiện, các lô **trong phiên** (`seen` trong `staff.js`) | `GET /purchase-tasks` chỉ trả việc mở, `GET /parcels` chỉ trả kiện chưa bay, **không có** route liệt kê lô | tải lại trang là mất: dải của nhân viên mất số ở ba ô giữa, nhóm "vừa xong" biến mất |
| đọc lại `GET /purchase-tasks/{id}` sau khi mua để có ngày và số tiền mua | bảng đọc đơn không lưu hai thứ đó | chỉ có số cho việc **chính tay mình vừa làm** |
| tính phần còn lại = tổng − cọc (`balanceOf` trong `words.js`) | `summaryView` không có `balance` | phép tính của domain bị chép sang JS |
| dải của **khách** ghi "xong" ở Đã mua, Kho Denver, Đã bay | khách chỉ đọc được `/me/orders` | khách không biết hàng về kho ngày nào, bay ngày nào |
| hằng `LANE = "us_forwarder"`, `FREIGHT_CURRENCY = "USD"` trong hai file JS | chưa có route đọc tuyến bay | thêm tuyến thứ hai là màn hình sai |

**Sự thật đã kiểm trong code (23/09):** `reporting.Projector` **đã nghe** đúng các event mang dữ liệu
còn thiếu, chỉ chưa ghi lại:

```go
// internal/app/reporting/projector.go
OnPurchaseConfirmed(m contracts.PurchaseConfirmedV1)  // m.At, m.Paid, m.Reference (đang lưu mỗi Reference)
OnParcelReceived(m contracts.ParcelReceivedV1)        // m.At, m.Actual (cân + hộp)  — đang chỉ đổi Tracking
OnBatchShipped(m contracts.BatchShippedV1)            // m.At, m.Allocations[i].Freight — đang chỉ đổi Tracking
```

Nên phần lớn đợt này là **một migration cho bảng đọc** và vài field trong `summaryView`, **không có
event mới, không đổi domain**.

---

## 2. Ràng buộc: giữ nguyên từ đợt 21

1. Không chữ máy nào ra màn hình; mọi mã đi qua `words.js`.
2. Mọi con số nói được nó từ đâu ra; ô không có số thì không bịa.
3. Không bước build, không CDN cho `web/`.
4. Số vàng: báo giá `5.393.720 ₫`, cọc `2.696.860 ₫`, `variance -2.50 USD`.
5. **Mới:** số nội bộ (số tiền mua thật, phần cước của hãng bay) **chỉ** cho nhân viên, cùng cách
   `ShopReference` đang làm (`summaryViewOf(s, staff bool)`). Khách thấy ngày và cân, không thấy
   giá vốn: lộ giá vốn là lộ biên lãi.
6. Luật vàng của repo: `internal/domain/` không đổi import; `go test ./...` thêm test, không bớt.

---

## 3. Việc đề nghị

### A. Bảng đọc đơn nhớ ba mốc của kiện (migration `0014_summary_milestones.sql`)

Thêm vào `reporting.OrderSummary`, tất cả nullable, **không backfill** (cùng luật `0012`/`0013`):

| Field | Ghi từ event | Khách thấy? |
|---|---|---|
| `BoughtAt time.Time` | `purchase_confirmed.At` | có |
| `PaidAtShop shared.Money` | `purchase_confirmed.Paid` | **không** |
| `ReceivedAt time.Time`, `WeighedG int64` | `parcel_received.At`, `.Actual` | có |
| `ShippedAt time.Time` | `batch_shipped.At` | có |
| `FreightShare shared.Money` | phần của đơn trong `batch_shipped.Allocations` | **không** |

Chịu sai thứ tự như mọi handler của projector (§20 của WALKTHROUGH): event nào cũng tạo được dòng.
Test giao ba event ngược chiều.

### B. `summaryView` thêm `balance` — *đã làm, T-002 (25/09), SETUP §9 đợt 23*

`balance` = `Total − Deposit`, tính ở adapter bằng đúng `shared.Money.Sub`. Xoá `balanceOf` khỏi
`words.js` và hai chỗ gọi. Test: số trong JSON bằng số `CustomerOrder.Balance()` cho cùng một đơn.

### C. Hai route đọc cho trang nhân viên

| Route | Để làm gì |
|---|---|
| `GET /batches?status=open,closed` | bỏ cách tìm lô qua `batch_id` của kiện; lô rỗng vừa mở cũng thấy được |
| `GET /purchase-tasks?status=confirmed&since=…` | nhóm "vừa mua xong" sống qua lần tải lại |

Cả hai `requireOperator`. Trạng thái lạ là 400, không phải danh sách rỗng (cùng lý do như
`GET /orders?status=`).

### D. Màn hình dùng dữ liệu mới, và bỏ trí nhớ phiên

- `journeyOf(o, viewer)` đọc thẳng các field mới của đơn; tham số `facts` bỏ đi.
- `staff.js`: xoá `seen`, `taskInfo`, `factsFor`, lần đọc lại `GET /purchase-tasks/{id}` sau khi mua.
- Trang khách: ba ô giữa có ngày và cân (không có giá vốn).
- `LANE`, `FREIGHT_CURRENCY`: đọc từ tuyến bay nếu T-001 đã có route đọc; chưa có thì giữ hằng, ghi
  rõ là nợ (xem D3 ở §7).

### E. Đưa script kiểm bằng trình duyệt vào repo (chờ D2)

Ba lượt của đợt 21 cho thấy mọi bug màn hình đều do script Chrome bắt, không phải `go test`. Script
đó đang nằm ở thư mục tạm. Đề nghị `scripts/ui-e2e/` (tách khỏi `web/`, có `package.json` riêng với
`puppeteer-core`, dùng Chrome có sẵn trên máy): đi hết luồng tới lúc giao, quét chữ máy, đo độ trễ
sau "Lưu size này", kiểm `variance -2.50 USD`, chạy hai lần trên cùng một process.

### F. Việc nhỏ đã thấy mà chưa làm

- Ảnh chụp thật trong `docs/UI-GUIDE.md` (hiện chỉ có sơ đồ ASCII).
- Trên điện thoại, bảng báo giá có cột tên hẹp; thử đưa số tiền xuống dưới tên ở ≤ 480 px.
- Toast "Có tiến triển" ở trang khách hiện một lần cho mỗi bước; nhân viên làm liền bốn bước thì
  khách nhận bốn toast. Gộp thành một nếu cách nhau dưới 3 giây.

---

## 4. Thứ tự làm

Mỗi giai đoạn kết thúc với `go test ./...` xanh và hai màn hình vẫn đi hết luồng.

| # | Việc | Đụng | Xong khi |
|---|---|---|---|
| N1 | A: migration + projector + repo cả hai adapter | `internal/app/reporting`, `adapter/postgres`, `adapter/memory`, migration `0014` | test projector giao event ngược chiều vẫn đủ field; test Postgres với `PORTAGE_TEST_DSN` |
| N2 | B + field mới lên `summaryView` (lọc theo `staff`) | `adapter/http/summaries.go` | test: khách **không** nhận `paid_at_shop`, `freight_share` |
| N3 | C: hai route đọc | `adapter/http`, `app/logistics`, `app/procurement` | test 400 cho trạng thái lạ |
| N4 | D: màn hình | `web/app/{words,ui,staff,customer}.js` | tải lại trang nhân viên giữa chừng mà dải vẫn đủ số |
| N5 | E (nếu D2 = có) | `scripts/ui-e2e/` | chạy hai lần xanh trên cùng process |
| N6 | Docs | `UI-GUIDE` §3/§5, `MASTER` §5, `WALKTHROUGH` (mục mới), `SETUP` §9 đợt 22, `CLAUDE.md` số test/route/migration | số trong docs khớp code |

---

## 5. Kiểm chứng

1. `gofmt -l . && go vet ./... && go test ./...`, và với `PORTAGE_TEST_DSN`.
2. Đi hết luồng trên màn hình tới lúc giao, **hai lần** trên cùng process; `GET /reconciliations/{id}`
   của mỗi đơn là `-2.50 USD`.
3. **Tải lại trang nhân viên ở từng bước** (sau mua, sau cân, sau bay): dải không mất số nào.
4. Soi JSON của `/me/orders`: không có `paid_at_shop`, `freight_share`.
5. `scripts/smoke.sh` / `smoke.ps1` vẫn in đúng số vàng (migration mới không được làm vỡ luồng cũ).
6. Chụp cả hai theme ở 390 và 1440 px **với dữ liệu sạch**, rồi soi (bài học của đợt 21: kiểm đo
   được không thay được việc nhìn).

---

## 6. Tự phản biện

**"Cứ để màn hình nhớ, đơn giản hơn."** Không: hai nguồn sự thật cho cùng một con số, và nguồn ở
trình duyệt mất khi tải lại. Đợt 21 chọn cách đó **chỉ vì** bị cấm đụng Go; lý do đó không còn.

**"Thêm field vào bảng đọc là nới rộng read model."** Đúng, nhưng read model sinh ra để làm đúng
việc này (DDD.md §24: một SELECT không JOIN cho một màn hình). Projector đã nghe ba event đó rồi;
bỏ qua dữ liệu trong payload mới là lạ.

**"Sao không cho khách thấy luôn phần cước?"** Vì phần cước của hãng bay cộng với giá mua thật là
giá vốn. Khách đã có báo giá từng dòng; giá vốn là của mình. Ràng buộc 5 ở §2.

**Rủi ro lớn nhất: `FreightShare` sai khi một lô chia lại.** Hiện `batch_shipped` chỉ phát một lần
cho mỗi lô, nên không có "chia lại". Nếu sau này có, handler phải **ghi đè**, không cộng dồn. Viết
test cho chính câu đó ngay ở N1.

**E (script trình duyệt) đi ngược tinh thần "không npm".** Tinh thần đó là cho `web/`: màn hình chạy
không cần mạng. Script kiểm là công cụ của người phát triển, như `learn/` đã có `package.json` riêng.
Vẫn để anh quyết (D2), vì nó thêm một toolchain Node vào quy trình kiểm của một repo Go.

---

## 7. Cần anh chốt (im lặng = làm theo đề nghị)

| # | Câu hỏi | Đề nghị | Đã bỏ |
|---|---|---|---|
| D1 | Khách có thấy **cân thật** ở kho không? | **có**: cước của khách tính theo cân, họ có quyền biết hộp nặng bao nhiêu | chỉ ngày: che một số khách có lý do muốn biết |
| D2 | Đưa script kiểm Chrome vào repo (§3.E)? | **có**, ở `scripts/ui-e2e/` với `package.json` riêng | để ngoài repo: lần sau lại viết lại từ đầu |
| D3 | `LANE`/`FREIGHT_CURRENCY` hằng trong JS? | **giữ** tới khi T-001 xong (tuyến bay ở file hay database còn chờ Q5) | thêm `GET /lanes` ngay: có thể phải làm lại theo T-001 |
| D4 | `flow.html` làm lại theo hệ thiết kế mới? | **để sau** đợt này, như D3 của plan trước | làm cùng: gấp đôi khối lượng |

---

## 8. Working tree lúc viết plan

Việc T-001 (ratecard) vẫn **chưa commit** và **không thuộc** đợt giao diện: `agent/projects/portage/…`,
`agent/logs/runs/…`, `config/`, `internal/adapter/config/` (còn 2 test đỏ), `go.mod`, `go.sum`.
Commit của đợt 21 cố ý **không** gồm các file đó.

---

## Corrections

*(chỉ anh viết. Một dòng có ngày ở đây thắng mọi thứ ở trên.)*
