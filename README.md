# Portage

Hệ thống đặt hàng hộ xuyên biên giới (Mỹ → Việt Nam): khách đặt trên web của
mình, mình mua hộ ở web Mỹ (Nike, The North Face, Sony…), gom lô tại kho
Denver, ship về tận nhà khách ở VN.

Viết bằng **Go**, kiến trúc **DDD**. Vừa là sản phẩm thật, vừa là project học.

## Tài liệu

| File | Nội dung |
|---|---|
| [docs/SETUP.md](docs/SETUP.md) | dựng môi trường, cấu trúc thư mục, quy ước, lệnh hàng ngày |
| [docs/DDD.md](docs/DDD.md) | sổ tay DDD — mỗi khái niệm giải thích bằng chính code trong repo, đối chiếu Symfony |
| [docs/CATALOG.md](docs/CATALOG.md) | thiết kế bounded context `catalog`: nguồn dữ liệu, affiliate, các quyết định đã gỡ/thêm; `pricing` nghe gì từ nó |
| [docs/GO-CHO-PHP.md](docs/GO-CHO-PHP.md) | cú pháp Go tra nhanh cho người viết PHP; comment `// [PHP]` trong code |
| [docs/WALKTHROUGH.md](docs/WALKTHROUGH.md) | **đọc code theo thứ tự**: §0–§13 một request đi hết 5 tầng; §14 `pricing`; §15 `ordering` — cọc 50 %, điểm không thể quay đầu; §16 `procurement` — ACL, saga; §17 `logistics` — cân thật, chia cước, Quote vs Actual; §18 auth · §19 sweep (việc không ai gọi) · §20 read model (một bảng, năm nguồn) · §21 ACL cho AI; thứ tự đọc 93 file; hành trình ba ngày |
| [docs/HOC.md](docs/HOC.md) | **lộ trình học**: 8 buổi, mỗi buổi đọc gì → sửa gì cho test đỏ → biết là hiểu khi nào; 15 câu tự kiểm tra; nói gì (và **không** nói gì) khi phỏng vấn |
| [docs/FLOW-ORDER.md](docs/FLOW-ORDER.md) | **một đơn từ đầu tới cuối**: mỗi bước ai gọi, ai nghe, bảng nào đổi, mã lỗi nào gặp; nhánh rẽ (huỷ, hết size, event tới hai lần); bảng "trạng thái → làm được gì" cho 5 aggregate |
| [web/console.html](web/console.html) | **console thật**: `go run ./cmd/api -web ./web` rồi mở `localhost:8080/ui/` — bấm 26 bước gọi API thật, hoặc thao tác tay như khách/nhân viên |
| [docs/UI-GUIDE.md](docs/UI-GUIDE.md) | **hướng dẫn thao tác UI từng bước** để chạy thử luồng, kèm 10 nhánh rẽ nên thử và bảng chẩn lỗi |
| [web/flow.html](web/flow.html) | **xem luồng bằng mắt**: mở bằng double-click, bấm qua 26 bước một đơn — màn hình + request + event + ai nghe, outbox pending→sent, đối soát cuối cùng |
| [docs/P9-PLAN.md](docs/P9-PLAN.md) | plan gốc của P9 (**đã xong hết 06/09**) — giữ lại để thấy mỗi quyết định được chốt thế nào trước khi code |

## Chạy

```bash
go run ./cmd/api                    # API :8080, in-memory — token dev: "dev-operator" / "dev-customer"
curl -H "Authorization: Bearer dev-operator" localhost:8080/purchase-tasks   # MỌI route đều cần token (không có → 401)
docker compose up -d                # Postgres 16 (cần Docker Desktop)
go run ./cmd/api    -dsn "postgres://portage:portage@localhost:5432/portage?sslmode=disable" \
                    -bootstrap-operator-token "$(openssl rand -hex 32)"   # chìa đầu tiên, CHỈ khi api_tokens rỗng
go run ./cmd/worker -dsn "postgres://portage:portage@localhost:5432/portage?sslmode=disable"
go run ./cmd/api -web ./web          # + console trong trình duyệt tại http://localhost:8080/ui/   # + sweep quote hết hạn mỗi phút (-sweep)
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1   # cả flow trên binary thật: paste → publish → relay → quote → accept
go test ./...                       # 262 test < 2 giây; set PORTAGE_TEST_DSN để chạy cả 30 test tích hợp
go test -cover ./...   # kèm độ phủ
go vet ./...           # soi lỗi tĩnh
gofmt -l .             # liệt kê file chưa format (rỗng = sạch)
```


## API — 37 route

Mọi route đều ở sau một cửa: `Authorization: Bearer <token>`. Không token →
**401**; đúng token sai loại → **403**.

| Route | Ai | Ghi chú |
|---|---|---|
| `POST /tokens` | operator | cắt chìa; token trả về **đúng một lần** |
| `GET /tokens` · `DELETE /tokens/{hash}` | operator | xem/thu hồi; chìa operator cuối cùng → 409 |
| `POST /categories` · `/merchants` · `/lanes` · `/fx` | operator | dữ liệu tham chiếu; `/fx` **không phát event** |
| `POST /products` | cả hai | provenance suy từ token: operator → verified, khách → nháp |
| `POST /products/from-url` | cả hai | AI đọc trang → nháp `feed`; không key → **503** |
| `POST /products/{id}/{variants,confirm-listing,measure,publish}` | operator | bốn bước tới publish |
| `POST /quotes` · `GET /quotes/{id}` | cả hai | báo giá là **ảnh chụp**, hạn 48 h |
| `POST /quotes/{id}/accept` · `POST /orders` | **khách** | chấp giá và đặt đơn là việc của khách |
| `GET /me/orders` | **khách** | đơn của chính mình — "me" là token, không phải tham số |
| `GET /orders?status=` | operator | hàng đợi việc; status gõ sai → **400** |
| `GET /orders/{id}` · `POST /orders/{id}/cancel` | chủ đơn hoặc operator | người lạ → **404**, không phải 403 |
| `POST /orders/{id}/{deposit,balance,deliver}` | operator | tới khi có cổng thanh toán |
| `GET /purchase-tasks` · `/purchase-tasks/{id}` · `POST …/{confirm,fail}` | operator | danh sách việc mua |
| `GET /parcels` · `/parcels/{id}` · `POST /parcels/{id}/receive` | operator | kho: cân **thật** |
| `POST /batches` · `GET /batches/{id}` · `POST …/{parcels,close,ship}` | operator | gom lô, chia cước |
| `GET /reconciliations/{order}` | operator | Quote vs Actual → `variance` |

Đi hết một đơn qua các route này, kèm event nào ai nghe và hỏng thì sao:
[docs/FLOW-ORDER.md](docs/FLOW-ORDER.md).

## Trạng thái

- **Shared Kernel** (`internal/domain/shared`): `Money`, `Currency`, `Rate`,
  `ExchangeRate`, `Weight`, `Dimensions`, `ParcelSpec`, `ID` (UUID v7), `OperatorID`, `Events`.
- **Bounded context `catalog`** (`internal/domain/catalog`) — v1 xong: aggregate
  `Merchant` (vòng đời) và `Product` ⊃ `Variant` (draft → published → retired,
  3 `Provenance`, cờ nghi trùng), `CategoryPolicy`, `FreeShipping`, `SourceURL`;
  15 domain event; 3 repository interface.
- **Bounded context `pricing`** (`internal/domain/pricing`, 05/09): `ShippingLane`
  (divisor, bước cân, `RateCard` 4 loại hàng, phụ thu pin, `DutyPolicy`), `QuotePolicy`
  (thuế 8,81 %, lãi 10 %/sàn 500k, cọc 50 %, 48 h), `Calculate` (domain service thuần),
  aggregate `Quote` = ảnh chụp có hạn; projection `Listing`/`CategoryProfile` dựng từ
  event của catalog. Pricing **không import, không FK** sang catalog.
- **Bounded context `ordering`** (`internal/domain/ordering`, 05/09): aggregate `CustomerOrder`
  — state machine 7 trạng thái, cọc phải đúng số, `Cancel` trả `Refund` theo **điểm không thể
  quay đầu** (chưa mua: hoàn đủ; đã mua: mất cọc); projection `AcceptedQuote` từ event của pricing.
- **Bounded context `procurement`** (`internal/domain/procurement`, 05/09): aggregate `PurchaseTask`
  (open → confirmed | failed), `PurchaseReceipt` = số tiền **thật** đã trả (Actual đầu tiên của Quote vs
  Actual), port `MerchantACL` với adapter `merchant.Manual` (một con người mua) — **Anti-Corruption Layer**
  đầu tiên; projection `Shop`/`Item` từ event của catalog.
- **Bounded context `logistics`** (`internal/domain/logistics`, 06/09): `Parcel` (cân **thật** ở kho),
  `ConsolidationBatch` (gom lô, `Ship` chia hoá đơn qua `FreightAllocator` — Domain Service dạng interface,
  `ByChargeableWeight`), `shared.Allocate` chia tiền đúng từng cent; **`pricing.Reconciliation`** = Quote vs
  Actual mỗi đơn từ 3 context, `GET /reconciliations/{order}` → `variance`.
- **Tầng app**: `catalogapp` 7 use case (kể cả 3 bước operator: variant, confirm-listing,
  measure), `pricingapp` (`IssueQuote`, `AcceptQuote`, `DefineLane`, `Projector`, `Reconciler`), `orderingapp`
  (`PlaceOrder` với luật một-quote-một-đơn, thanh toán, `Cancel`, vòng đời, `Reactor` nghe procurement + logistics;
  `Deps.mutate`), `procurementapp` (`OpenTask` hỏi ACL với 4 nhánh, `ConfirmTask`/`FailTask`) và `logisticsapp`
  (`ExpectParcel`, `ReceiveParcel`, `Open/Add/Close/ShipBatch`) đi qua port
  `Clock` / `UnitOfWork` / `Outbox`; **Published Language** `internal/contracts` (*V1);
  **adapter in-memory** (`internal/adapter/memory`) và `platform/clock` cài các port đó.
- **Auth** (`internal/platform/auth` + `adapter/http/auth.go`, 06/09): bearer token, port `Verifier`
  (30 route chỉ đọc) tách khỏi `Issuer` (một route phát chìa), adapter `Static` (RAM) và `postgres.TokenRepo`
  (lưu **hash** SHA-256, `revoked_at` chứ không DELETE). Middleware bọc **cả mux** — fail-closed, route thêm
  ngày mai đã ở sau cửa. `X-Operator-ID` và `customer_id` **biến mất**: ai xác nhận listing, đơn thuộc về ai,
  đều suy từ token chứ không từ thứ người gọi tự khai. 401 ≠ 403, và chủ đơn sai → **404** (403 sẽ xác nhận đơn có thật).
  Sống ở `platform/` chứ không ở `domain/` vì *"ai đang gọi"* là việc của **biên**, cùng lý do với `Clock`.
- **HTTP** (`internal/adapter/http`, package `httpapi`) + `cmd/api`: 37 endpoint sau cửa auth (bảng đầy đủ ở trên: `/tokens`, catalog, `/lanes`, `/fx`,
  `/quotes` + read model, `/orders` + deposit/balance/cancel/deliver, `/me/orders`, `/purchase-tasks`,
  `/parcels` + receive, `/batches` + parcels/close/ship, `/reconciliations/{order}`), chuẩn hoá số tiền
  theo `Accept-Language`, một bảng lỗi → 400/401/403/404/409.
- **Sweep theo giờ** (`internal/worker/sweep.go` + `pricingapp.ExpireQuotesHandler`, 06/09): quote quá
  `TTL` 48 h chuyển `expired` mà không cần ai bấm gì — use case đầu tiên **không có người gọi**, chỉ thời
  gian đẩy, nên cũng là lý do rõ nhất để `Clock` là port (48 h trôi trong một dòng test). Mỗi quote **một
  transaction** (lô không phải invariant), danh sách đọc ngoài tx còn quyết định đọc lại trong tx, `LIMIT`
  chặn một lượt chứ không bỏ việc, và index partial `quotes_open_idx` chỉ lớn theo số quote **còn mở**.
  Chạy trong `cmd/worker -sweep` (và trong `cmd/api` chế độ memory, nơi không có worker riêng).
- **Read model** (`internal/app/reporting` + `order_summaries`, 06/09): màn hình "đơn của tôi" cần
  dữ liệu của **bốn context** (tên sản phẩm, trạng thái đơn, mã đơn ở shop, thùng tới đâu), nên nó là
  **một bảng phẳng ghi CHỈ bằng event** — CQRS về lưu trữ, không JOIN qua bounded context. Package duy
  nhất không có domain: read model không có invariant nào để giữ. Idempotent và **chịu sai thứ tự**
  (event nào cũng được tạo dòng "khung"). `GET /me/orders` — "me" là token, không phải tham số đường dẫn.
- **ACL thứ hai cho AI** (`internal/adapter/openai`, 06/09): khách dán link → mô hình đọc → **nháp**
  `SourcedByFeed`. Port trả **string thô** để domain quyết cái gì hợp lệ; máy được *tạo* nháp, không
  được *xác nhận* — `Publish` vẫn đòi một con người. Không SDK, không cào trang (điều khoản của shop).
  Không `OPENAI_API_KEY` trên Postgres → **503**, tuyệt đối không rơi về Fake: tính năng tắt phải
  trông như đang tắt. `merchant.Router` làm điều tương tự cho ACL shop: thêm shop có API = một dòng ở `wire`.
- **Postgres** (`internal/adapter/postgres`): repository qua `Snapshot()`, transaction
  trong `ctx`, outbox cùng transaction, migration SQL nhúng. **Worker**
  (`internal/worker` + `cmd/worker`): đọc outbox → publish → `sent_at`, at-least-once;
  `wire.Subscribe` là bảng định tuyến 31 dòng event → consumer — đọc từ trên xuống chính là vòng đời
  một đơn: catalog → pricing → ordering → procurement → logistics → ordering + pricing (đối soát).
  Payload event là hợp đồng viết tay (`internal/adapter/eventcodec`, Encode + Decode).
- **7 test canh quyết định** (`internal/domain/decisions_test.go`) — kiến trúc
  được kiểm tra bằng `go/ast`, vi phạm là build đỏ.
- 292 test; 262 unit < 2 giây, 30 test tích hợp cần `PORTAGE_TEST_DSN`; một test vàng chạy **cả vòng đời**
  (paste → publish → quote `5 393 720 ₫` → accept → order → cọc → task mua → confirm `163.22 USD` → `purchased`
  → cân thật → gom lô → ship `27.50 USD` → `in_transit` → trả nốt → `delivered` → đối soát `variance −2.50 USD`)
  trên **cả** memory và Postgres; `scripts/smoke.ps1` (Windows) và `scripts/smoke.sh` (CI, Linux) chạy đúng flow đó trên
  **binary thật** + Postgres thật + bearer token thật — 20 event, 5 context, và bản `.sh` **assert** số vàng chứ không chỉ in ra.
  Chưa có: bảng giá thật trong `config/`, ACL API thật của một shop (chưa shop nào cho API), cổng thanh toán.
