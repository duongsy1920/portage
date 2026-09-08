# Thao tác trên UI để chạy thử cả luồng

> Hai màn hình, hai việc khác nhau:
>
> | | Dùng khi |
> |---|---|
> | `web/flow.html` | **hiểu** luồng — mô phỏng, không gọi API, mở bằng double-click |
> | `web/console.html` | **chạy thử** luồng — gọi API thật, phải mở qua `http://localhost:8080/ui/` |
>
> File này nói về cái thứ hai. Đã kiểm 06/09/2026: 26 bước xanh, kết thúc ở `variance -2.50 USD`.

---

## 0. Chạy lên

**Cách nhanh nhất — in-memory, một lệnh.** Không cần Docker. Dữ liệu mất khi tắt, và relay outbox
chạy ngay trong `cmd/api` mỗi 200 ms nên không cần `cmd/worker`:

```bash
cd /d/portage
go run ./cmd/api -web ./web
# → portage api: console at http://localhost:8080/ui/console.html
```

**Giống production hơn — Postgres, ba terminal.** Bảng vẫn còn sau khi tắt, và relay do worker thật làm:

```bash
docker compose up -d                                   # terminal 1 (một lần)
DSN="postgres://portage:portage@localhost:5432/portage?sslmode=disable"
go run ./cmd/api    -dsn "$DSN" -web ./web             # terminal 2
go run ./cmd/worker -dsn "$DSN" -every 300ms           # terminal 3
```

Mở **http://localhost:8080/ui/console.html** (hoặc chỉ `localhost:8080`, nó tự chuyển).

> **Đừng double-click `console.html`.** Mở bằng `file://` thì module JS không nạp được và fetch
> không cùng origin — console sẽ hiện dải đỏ nhắc đúng chuyện này. `flow.html` thì double-click được.

Cờ `-web` chỉ thêm một file server ở `/ui/` **cạnh** API, cùng cổng. Cùng origin nên không có
CORS, và token trong `fetch()` đi qua đúng cái `authenticate()` mà `curl` đi qua.

---

## 1. Chìa khoá (dải trên cùng)

Mọi route đều cần `Authorization: Bearer …`; không có token là **401**, sai loại người là **403**.

| Chế độ | Token operator | Token khách |
|---|---|---|
| in-memory | `dev-operator` (đã điền sẵn) | `dev-customer` (đã điền sẵn) |
| Postgres | chạy api thêm `-bootstrap-operator-token <chuỗi-bạn-tự-đặt>` rồi dán chuỗi đó vào | tab **Chìa khoá** → **Phát chìa mới** (kind `customer`) → **Dùng làm token khách** |

Bấm **Lưu + kiểm**. Góc phải phải hiện `API ok · token operator ok` với đèn xanh. Nếu đỏ/vàng thì
xem bảng ở §7. Token được nhớ trong `localStorage`, reload không phải điền lại.

---

## 2. Cách nhanh: tab **Chạy luồng**

Bấm **Chạy tới hết** — 26 bước tự chạy, mỗi bước một (hoặc vài) request thật. Xong thì:

- cột trái: 26 dấu ✓, thanh tiến độ đầy
- bước 26 hiện `variance -2.50 USD`
- panel **Id của lần chạy này** có đủ merchant / product / variant / quote / order / task / parcel / batch

Muốn học thì bấm **Chạy bước này** từng bước một, và đọc:

- dòng **hint** — bước đó chứng minh điều gì
- khối **Đã gửi / Nhận về** — JSON thật, không phải ví dụ
- các dòng `›` — khi bước đang chờ relay hoặc đang thử lại

Bấm vào một bước ở cột trái chỉ **nhảy tới xem**, không chạy lại. **Đặt lại** xoá session và sinh
`run` mới (mỗi lần chạy phải có `site` và `source_url` khác nhau, vì `merchants.site` là UNIQUE và
URL trùng sẽ bị cắm cờ nghi trùng làm publish thất bại).

### 26 bước: kết quả đúng trông như thế nào

| # | Bước | Đúng thì thấy |
|---|---|---|
| 1 | Đăng ký shop | `shop 01a0…` · site `www.example-<run>.com` |
| 2 | Khách dán link | `product 01a0…` — gửi bằng token **khách**, nên provenance là customer |
| 3 | Thêm variant | `variant 01a0…` — phát `catalog.variant_added`, nên relay #1 giao **5** event chứ không phải 4 |
| 4 | Xác nhận listing | `204` |
| 5 | Cân và đo thật | `204` |
| 6 | Publish | `published` |
| 7 | Chờ worker relay #1 | `đã chờ` |
| 8 | Xin báo giá | `quote 01a0…`; nếu relay chưa kịp sẽ thấy `còn listing_not_found (lần 1/12)` rồi mới xanh |
| 9 | Khách đọc báo giá | `branded · 2500 g · tổng 5.393.720 ₫ · cọc 2.696.860 ₫` |
| 10 | Khách đồng ý | `accepted` |
| 11 | Chờ worker relay #2 | `đã chờ` |
| 12 | Đặt hàng | `đơn 01a0…`; có thể thấy `còn quote_not_accepted` một hai lần |
| 13 | Thu cọc 50 % | `đã thu 2696860 ₫` |
| 14 | Chờ procurement mở việc mua | `task 01a0… · shop tính bằng USD` sau 1–3 lần thử |
| 14b | Cột **Mua gì** ở bảng việc | `Air Trainer 90 · US 9 · black`, mã `EX-AT90-9-BLK`, kèm link shop |
| 15 | Người mua xem việc | `1 việc đang mở` |
| 16 | Mua xong, nhập biên nhận | `NK-<run>-001 · đã trả 163.22 USD` |
| 17 | Chờ 4 listener | `đơn purchased · thùng 01a0… đang chờ về kho` |
| 18 | Cân thùng ở kho | `204` |
| 19 | Mở lô hàng | `lô 01a0…` |
| 20 | Xếp thùng, dán kín | `lô closed` |
| 21 | Ship lô | `chia: đơn 01a0… · 2500 g · 27.50 USD` |
| 22 | Chờ ordering nghe batch_shipped | `đơn in_transit` |
| 23 | Khách trả nốt | `đã thu nốt 2696860 ₫` |
| 24 | Giao tận nhà | `delivered` |
| 25 | Chờ màn hình khách | `Air Trainer 90 · delivered · đã bay` |
| 26 | Đối soát | `variance -2.50 USD (ước tính 163.22+25.00 · thật 163.22+27.50)` |

Ba bước ⚙ (7, 11) chỉ **chờ** rồi để bước sau tự thử lại; bốn bước ⚙ còn lại (14, 17, 22, 25)
**poll** một endpoint đọc thật, nên bạn thấy đúng khoảnh khắc projection được ghi.

---

## 3. Cách chậm: thao tác tay như người dùng thật

Vẫn nên chạy runner một lượt trước cho có dữ liệu, rồi tự bấm để cảm nhận. Thứ tự:

**Tab Khách → panel “Dán link sản phẩm”**
1. `Shop (merchant_id)` cần có id — chạy bước 1 ở runner, hoặc dán id cũ.
2. Sửa `Link` cho khác lần trước (nếu trùng, publish sẽ bị 409 `suspected_duplicate`).
3. **Gửi cho nhân viên** → góc phải panel hiện `product …`.

**Tab Nhân viên → “Sản phẩm nháp → publish”**
4. Bấm **Publish** *trước* — cố ý sai, để thấy `409 no_variants`.
5. **Thêm variant** → **Publish** lại → `409 unverified`.
6. **Xác nhận listing** → **Cân 1250 g** → **Publish** → xanh.
   *(Panel này không đọc được trạng thái sản phẩm: API không có `GET /products/{id}`. Trạng thái chỉ
   hiện ra qua chính mã lỗi — đó là CQRS, đọc là việc của read model.)*

**Tab Khách → “Báo giá”**
7. **Xin báo giá** ngay → có thể `404 listing_not_found` (worker chưa relay). Bấm lại sau 1 giây.
8. Đọc bảng: 150.00 + 13.22 thuế + 25.00 cước = 188.22 USD → ×26.000 → 4.893.720 ₫ + phí 500.000 ₫
   = **5.393.720 ₫**, cọc **2.696.860 ₫**.
9. **Đồng ý báo giá**.

**Tab Nhân viên → “Một đơn hàng”** (đơn do runner đặt, hoặc dán `order_id`)
10. **Tải đơn** → `awaiting_deposit`. Sửa `Số tiền` thành `1000` → **Thu cọc** → `409 wrong_amount`.
11. Trả lại `2696860` → **Thu cọc** → `deposited`.

**Tab Nhân viên → “Việc cần đi mua”**
12. **Tải danh sách** (rỗng thì đợi worker). Bấm **Chọn** ở dòng của mình để điền `task_id`.
    Đọc cột **Mua gì**: tên sản phẩm, size/màu, mã của shop và link gốc. Bốn chữ đó
    procurement **chép** vào task lúc mở — không join sang catalog (WALKTHROUGH §22).
13. Sửa `Tiền tệ` thành `VND` → **Xác nhận đã mua** → `409 paid_currency`. Trả lại `USD` → xanh.

**Tab Nhân viên → “Kho: thùng chưa bay”**
14. **Tải danh sách** → một thùng `expected` mang mã đơn shop. Bấm **Cân**.

**Tab Nhân viên → “Lô hàng”**
15. **Mở lô us_forwarder** → **Xếp thùng vào lô** → **Ship lô** → `409 batch_not_closed`.
16. **Dán kín lô** → **Ship lô** → bảng phần chia cước hiện `2500 g → 27.50 USD`.

**Tab Nhân viên → “Một đơn hàng”**
17. **Tải đơn** → `in_transit` (sau relay) → **Thu nốt** → **Đã giao** → `delivered`.

**Tab Khách → “Đơn của tôi”** → **Tải danh sách**: khách thấy trạng thái + vận chuyển, **không** thấy
mã đơn shop. So với **Nhân viên → “Hàng đợi theo trạng thái”**: cùng một bảng đọc, operator thấy thêm cột đó.

**Tab Nhân viên → “Đối soát Quote vs Actual”** → **Tải đối soát** → `-2.50 USD`.

---

## 4. Mười nhánh rẽ nên thử tay

| Thử | Kết quả đúng |
|---|---|
| Xoá token operator ở dải trên → **Lưu + kiểm** → bấm bất cứ nút nào | `401 unauthenticated` |
| Dán token **khách** vào ô operator → **Tải danh sách** ở “Việc cần đi mua” | `403 forbidden` |
| Xin báo giá ngay sau Publish | `404 listing_not_found`, bấm lại thì xanh |
| Đặt hàng ngay sau khi đồng ý báo giá (runner bước 12) | `409 quote_not_accepted` rồi tự xanh |
| Runner **Chạy tới hết** hai lần liên tiếp, cùng session | bước 12 `409 quote_already_used` — một quote một đơn |
| Sửa `variant_id` trong panel đặt hàng thành một uuid tự bịa | `409 variant_unknown` — ordering chỉ nhận variant nó đã nghe qua `catalog.variant_added` |
| Dán `variant_id` của một sản phẩm khác | `409 variant_not_for_product` |
| Thu cọc lệch một đồng | `409 wrong_amount` |
| “Không mua được” với lý do `hết size US 9` | đơn về `purchase_failed`; **Huỷ đơn** lúc này hoàn **đủ** 2.696.860 ₫ |
| **Huỷ đơn** sau khi đã `purchased` | hoàn `0 ₫`, `forfeited = true` — mất cọc |
| **Đã giao** khi chưa **Thu nốt** | `409 balance_unpaid` |
| Tab Chìa khoá → **Thu hồi** chìa operator cuối cùng | bị từ chối; phát chìa mới trước rồi mới thu hồi |

Tất cả những lỗi trên là **409 — nghiệp vụ từ chối**, không phải bug. Xem `docs/FLOW-ORDER.md` §9
hoặc tab **Nhánh rẽ** trong `flow.html` để biết vì sao mỗi cái tồn tại.

---

## 5. Đọc kết quả ở đâu

- **Tab Call log** — mọi request console gửi: status, thời gian, body đã gửi và nhận. Bấm một dòng để mở JSON.
- **Terminal worker** (`cmd/worker`) — event đi ra theo thứ tự, đúng số lượng mỗi lượt: 4 / 2 / 2 / 2 / 6 / 3.
- **Postgres** — nhìn thẳng vào bảng:

```bash
docker compose exec postgres psql -U portage -d portage -c \
  "select id, event_name, sent_at is not null as sent from outbox order by id desc limit 12;"
docker compose exec postgres psql -U portage -d portage -c \
  "select \"order\", status, tracking, shop_reference from order_summaries order by placed_at desc limit 5;"
docker compose exec postgres psql -U portage -d portage -c \
  "select \"order\", quoted_goods_minor, quoted_freight_minor, actual_goods_minor, actual_freight_minor from reconciliations;"
```

- **`scripts/smoke.ps1`** — cùng luồng đó nhưng không cần trình duyệt; dùng khi muốn biết lỗi ở UI hay ở API.

---

## 6. Hỏng thì xem bảng này

| Triệu chứng | Nguyên nhân |
|---|---|
| Dải đỏ “đang mở bằng `file://`” | mở qua `http://localhost:8080/ui/console.html`, không double-click |
| Đèn đỏ `Không gọi được API` | `cmd/api` chưa chạy, hoặc chạy cổng khác — sửa ô `API base` |
| Đèn vàng `token operator sai (401)` | in-memory dùng `dev-operator`; Postgres phải `-bootstrap-operator-token` |
| Đèn vàng `403` | ô operator đang dán token khách |
| Bước 8 thử hết 12 lần vẫn `listing_not_found` | Postgres mode mà **quên chạy `cmd/worker`** |
| Bước 14 không thấy task | cũng là worker; hoặc bước 13 chưa thu cọc thành công |
| `409 suspected_duplicate` khi Publish | `source_url` trùng lần chạy trước — bấm **Đặt lại** |
| `400 invalid_hostname` ở bước 1 | ô site bị sửa thành chuỗi không phải hostname |
| Sửa file trong `web/` mà UI không đổi | Ctrl+F5 (file server trả theo `Last-Modified`) |
| `404` ở `/ui/...` | thiếu cờ `-web ./web`, hoặc chạy `go run` từ thư mục khác |

---

## 7. Source FE

```
web/
├── flow.html            mô phỏng để hiểu luồng (một file, không phụ thuộc)
├── console.html         khung của console thật
└── app/
    ├── console.css      token màu/chữ, dùng chung nhận diện với flow.html
    ├── api.js           CHỖ DUY NHẤT gọi fetch: gắn token, dịch lỗi, ghi call log, poll/retry
    ├── steps.js         26 bước = 26 hàm gọi API thật (dùng cho runner)
    └── main.js          tabs, runner, các màn hình, wiring
```

Không npm, không build, không framework — cùng tinh thần với backend (`net/http` viết tay).
Sửa file rồi Ctrl+F5 là thấy. `cmd/api -web <dir>` là toàn bộ phần server của việc này: một
`http.FileServer` ở `/ui/` cạnh API, cộng một redirect từ `/`.
