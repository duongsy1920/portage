# P9 — Nguyên tắc, business rule và plan chi tiết (cho session tiếp theo)

> Viết ngày 2026-09-06, ngay sau khi P8 xong (201 test xanh, smoke 20 event qua 5 context).
> Mục đích: một session khác (Claude bên Windows, hoặc chính bạn) làm được P9 **mà không phải
> hỏi lại** — mọi quyết định đã chốt nằm ở đây, mọi chỗ còn mở được đánh dấu ❓ kèm đề xuất.
> Đọc kèm: [SETUP.md §6](SETUP.md) (10 quy ước code), [WALKTHROUGH.md](WALKTHROUGH.md) §14–§17
> (khuôn của một context: domain → contracts → codec → app → memory → postgres → http → wire → test vàng → docs).

---

## 0. Cách làm — không đổi so với P5–P8

| Quy tắc | Cụ thể |
|---|---|
| **TDD** | test đỏ trước, code sau, `gofmt -w` → `go vet ./...` → `go test ./...` (unit) → với `PORTAGE_TEST_DSN` (tích hợp) → `scripts\smoke.ps1` |
| **Không commit** | để working tree, báo lại; chủ repo commit |
| **Style** | không one-liner struct/func; comment `// [PHP] …` tiếng Việt đối chiếu Symfony ở chỗ khái niệm mới |
| **Khuôn một tính năng** | domain (VO/aggregate + sentinel mỗi lý do) → `contracts/*V1` nếu qua ranh giới → `eventcodec` (mapper + decoder + hàng trong `codec_test.go`) → `app` (handler `Deps` + `mustHave`; `mutate` nếu ≥ 3 handler cùng hình) → `adapter/memory` → `adapter/postgres` (migration `000N_*.sql` + repo + test round trip `reflect.DeepEqual` snapshot) → `adapter/http` (route + `errorTable` + test) → `wire` (Graph + `Subscribe`) → `wire_test.go` (`wholeFlow` kéo dài) → `smoke.ps1` → docs |
| **Docs bắt buộc mỗi task** | WALKTHROUGH.md §18+ (theo mẫu §17: bản đồ → domain → app → http → postgres → test vàng → transcript → bài học), SETUP.md (§4 cây, §5d curl, §6 "đã viết gì" + số test, §7 checkbox, §9 đợt N), DDD.md (mục liên quan + PHẦN VI), README.md |
| **Guard kiến trúc** | 7 guard `internal/domain/decisions_test.go` phải xanh: domain không import adapter/DB/HTTP/SDK; context không import nhau; không `float`, không `time.Now`, không tên hãng, không `Set*` |
| **Kỹ thuật cần nhớ** | luôn `-timeout`; giết ssh không giết `go test` (`Get-Process *.test`); patch theo anchor → **fetch file mới trước**, script vá phải `print SKIP` từng anchor thay vì `assert`; đổi tên hằng trạng thái nếu đụng tên kiểu event (`Status*` / `*Event`) |

**Definition of Done cho mỗi task:** unit + tích hợp xanh · test vàng `wholeFlow` mở rộng (nếu chạm flow) · smoke thật chạy · docs 4 file cập nhật · SETUP §9 có "đợt N" ghi quyết định và bug đã gặp.

---

## 1. Nguyên tắc kiến trúc đã chốt (áp cho mọi task P9)

1. **Ranh giới context là tuyệt đối.** Không import chéo, không FK chéo, id sang context khác là `shared.ID` trần. Nói chuyện bằng event qua outbox → `contracts.*V1` → projection bên nghe.
2. **Save → PullEvents → Outbox.Append, cùng transaction.** Handler không phát event; chỉ append.
3. **At-least-once ⇒ mọi consumer idempotent** (upsert, hoặc "đã làm rồi" nuốt đúng sentinel như `Reactor.alreadyDone`). Sai thứ tự ⇒ giữ dòng dở hoặc bỏ qua có chủ đích, **không** lỗi mù.
4. **Eventual consistency có mã trạng thái**: `listing_not_found`, `quote_not_accepted`, danh sách rỗng trước relay — tài liệu API nói thẳng.
5. **Payload là hợp đồng**: mapper viết tay; guard `TestEncode_contractCoversEveryDomainEvent` quét `domain/*/events.go`; decoder chỉ khi có consumer.
6. **Tiền không float, chia tiền dùng `shared.Allocate`**, làm tròn half-up ở `Money.Mul`, cân làm tròn LÊN.
7. **`now` chỉ vào khi có gì được ghi**; `Clock` là port; test dùng `clock.Fixed`.
8. **Cấu hình nghiệp vụ là value object dựng ở `wire`** (`quotePolicy()`, `goodsClasses()`), không env var rải rác. Cấu hình *hạ tầng* (DSN, token bootstrap, API key) mới là cờ/env.
9. **Adapter làm đúng ba việc**: decode/validate bằng `Parse*` của domain, chuẩn hoá locale, map lỗi qua `errorTable`. Không quy tắc nghiệp vụ ở adapter.
10. **Ghi trả id (201) hoặc 204; đọc trả view.** Ngoại lệ phải có lý do ghi trong comment (như `POST /batches/{id}/ship` trả phần chia).
11. **Một aggregate một transaction**, nới lỏng chỉ khi trạng thái bên kia là sổ sách suy ra và có comment giải thích (`logisticsapp.AddParcelHandler`).
12. **Chống lặp theo hai nửa**: kiểm ở app (`ByQuote`, `ByOrder`) **và** `UNIQUE` ở DB.

---

## 2. Business rule toàn hệ — bảng tra một chỗ

Mọi rule dưới đây **đã là code có test**. P9 không được phá; task nào chạm phải ghi rõ.

| Miền | Rule | Ở đâu |
|---|---|---|
| Catalog | Publish cần: draft, ≥ 1 variant, listing **verified** (operator), parcel **verified** (operator đo), không nghi trùng | `Product.Publish` |
| Catalog | Cùng URL lần 2 → tạo mới + cờ nghi trùng (phương án C); operator clear có lý do | `AddProductHandler`, `ClearDuplicateFlag` |
| Catalog | Category không giữ thuế/HS/divisor (của lane) | CATALOG.md §6 |
| Pricing | `chargeable = max(actual, volumetric = mm³/divisor) làm tròn LÊN theo step`; parcel = đo thật, không thì mặc định category (`Estimated=true`) | `Calculate` |
| Pricing | `subtotal = item + item×tax + freight(class, chargeable) + surcharge(battery) + duty(lane)`; `total = subtotal×fx + max(subtotal×fx×margin%, floor)`; `deposit = total×50 %` | `Calculate`, `wire.quotePolicy()` |
| Pricing | Quote là **ảnh chụp** 48 h; accept trễ → `expired` + event + từ chối (commit rồi mới báo) | `Quote.Accept`, `AcceptQuoteHandler` |
| Pricing | Class không biết → `standard` (rẻ nhất; lộ ở đối soát) | `Classification.ClassOf` |
| Pricing | Variance = (quotedGoods + quotedFreight) − (actualGoods + actualFreight), USD; chỉ khi đủ hai actual | `Reconciliation.Variance` |
| Ordering | Một quote (đã accept) → **một** đơn; app `ByQuote` + DB `UNIQUE(quote)` | `PlaceOrderHandler`, `0003` |
| Ordering | Cọc phải **đúng** số; balance phải đúng số; balance chỉ khi `in_transit`; deliver cần balance đã trả | `PayDeposit/PayBalance/Deliver` |
| Ordering | **Cancel**: awaiting → hoàn 0; deposited/purchase_failed → hoàn **đủ** cọc; purchased/in_transit → **mất cọc**; delivered → từ chối | `CustomerOrder.Cancel` |
| Ordering | Điểm không thể quay đầu = `ConfirmPurchase` (từ `procurement.purchase_confirmed`) | `Reactor.OnPurchaseConfirmed` |
| Procurement | Task chỉ mở từ `deposit_paid` (không có POST tạo); một đơn một task (`ByOrder` + `UNIQUE("order")`) | `OpenTaskHandler`, `0004` |
| Procurement | Biên nhận: reference ≠ "", paid > 0, **đúng tiền tệ shop**, có operator | `PurchaseTask.Confirm` |
| Procurement | ACL: `ErrManualPurchase` → task open cho người; ok → confirm ngay; !ok → fail ngay; lỗi khác → retry | `OpenTaskHandler` |
| Logistics | Parcel chỉ từ `purchase_confirmed`; một đơn một parcel; receive cần operator + đủ 3 cạnh | `ExpectParcelHandler`, `Parcel.Receive` |
| Logistics | Batch chỉ mở trên lane có `LaneRule`; chỉ nhận parcel `received`; một parcel/đơn một lần; close cần ≥ 1; ship cần closed + freight > 0 | `OpenBatch/AddParcel/Close/Ship` |
| Logistics | Chia cước theo **chargeable weight**, cent thừa theo largest remainder, tổng đúng hoá đơn; phần chia **đóng băng** trong event | `ByChargeableWeight`, `shared.Allocate` |
| Ordering ← Logistics | `batch_shipped` → mỗi đơn `in_transit`; đơn đã huỷ trong thùng → bỏ qua (hàng của mình) | `Reactor.OnBatchShipped` |
| HTTP | 400 không hiểu · 404 không có · 409 nghiệp vụ từ chối · 500 bug; số tiền theo `Accept-Language` (`vi`: `1.234,56`) | `errors.go`, `normalizeAmount` |

---

## 3. Task P9 — thứ tự đề xuất

Thứ tự: **T1 → T3 → T4 → T2 → T7 → T5 → T6 → T8**. T1 trước vì mọi route sau đó cần principal;
T2 sau T1 vì `GET /me/orders` cần biết "me". T5/T6 độc lập, để cuối vì đụng thế giới ngoài.

### T1 — Auth thật: thay `X-Operator-ID` và `customer_id` (ưu tiên cao)

**Mục tiêu.** Mọi request biết *ai* gọi; header `X-Operator-ID` và field `customer_id` biến mất.

**Business rule.**
- Hai loại principal: `operator` (nhân viên: catalog/kho/mua/thanh toán) và `customer` (khách). Chưa có "admin".
- Xác thực bằng `Authorization: Bearer <token>`. Thiếu/sai → **401** `unauthenticated`. Đúng loại sai quyền → **403** `forbidden`.
- Quyền theo route:

| Route | Ai |
|---|---|
| `POST /categories`, `POST /lanes`, `POST /fx` (T4) | operator |
| `POST /merchants` | operator |
| `POST /products` | operator **hoặc** customer; **`sourced_by` bỏ khỏi body** — suy từ principal (`operator` → `SourcedByOperator`, `customer` → `SourcedByCustomer`) |
| `POST /products/{id}/{variants,confirm-listing,measure,publish}` | operator (operator id = principal) |
| `POST /quotes`, `GET /quotes/{id}` | operator hoặc customer |
| `POST /quotes/{id}/accept`, `POST /orders` | customer (**`customer_id` bỏ khỏi body**, = principal) |
| `GET /orders/{id}` | chủ đơn, hoặc operator |
| `POST /orders/{id}/cancel` | chủ đơn, hoặc operator |
| `POST /orders/{id}/{deposit,balance,deliver}` | operator (thanh toán do nhân viên xác nhận cho tới khi có cổng thanh toán) |
| `GET /purchase-tasks*`, `POST /purchase-tasks/**` | operator |
| `GET /parcels*`, `POST /parcels/**`, `/batches/**` | operator |
| `GET /reconciliations/{order}` | operator |
| `GET /me/orders` (T2) | customer |

- Khách đọc/huỷ đơn **không phải của mình** → **404** `order_not_found` (không lộ sự tồn tại). ❓ Đề xuất 404; 403 cũng chấp nhận được nếu ghi rõ lý do.
- Token lưu **hash SHA-256**, không lưu thô; có `revoked_at`.

**Thiết kế.**
- `internal/platform/auth/`: `Kind` (`operator`|`customer`), `Principal{Kind, ID shared.ID}`, port `Verifier interface{ Verify(ctx, token string) (Principal, error) }`, `ErrUnauthenticated`. Adapter `Static{tokens map[string]Principal}` (dev + test). `Principal` **không** vào domain — domain nhận `shared.OperatorID`/`shared.ID` như hiện nay.
- `internal/adapter/postgres/token_repo.go` + `migrations/0006_auth.sql`: `api_tokens(token_hash text PK, kind text, subject uuid, label text, created_at, revoked_at)`; `Verify` = hash → select → chưa revoke.
- `internal/adapter/http/auth.go`: middleware `authenticate(v auth.Verifier) func(http.Handler) http.Handler` (Bearer → `Principal` vào ctx qua key unexported); `principalOf(r)`, `requireOperator(h)`, `requireCustomer(h)`, `requireAny(h)`; `operatorOf(r)` đọc từ principal. Xoá `operatorHeader`. `errorTable`: `auth.ErrUnauthenticated → 401 unauthenticated`, `errForbidden → 403 forbidden`.
- Handler app đổi chữ ký: `catalogapp.AddProduct.SourcedBy` do adapter điền từ principal (app không đổi); `orderingapp.PlaceOrder.Customer` từ principal; `CancelOrder` thêm kiểm chủ đơn **ở adapter** (đọc order, so `Customer()`), hoặc thêm `Customer shared.ID` vào command và để handler so — ❓ đề xuất: handler so (`ErrNotOwner` → adapter map 404) để luật không nằm ở adapter.
- `wire.Graph.Auth auth.Verifier`. Memory: `Static` với `dev-operator` / `dev-customer`, in ra log lúc khởi động. Postgres: `postgres.TokenRepo`; **bootstrap**: cờ `cmd/api -bootstrap-operator-token <token>` tạo token operator đầu tiên nếu bảng rỗng (không seed token cố định vào DB thật). Tạo token khách: `POST /tokens {kind:"customer"}` operator-only → trả token thô **một lần**.
- `cmd/worker` không đổi.

**Test bắt buộc.** middleware: thiếu token 401, token lạ 401, đúng loại 200, sai loại 403; mỗi route bảo vệ có test 401; khách đọc đơn người khác → 404; `wire_test.wholeFlow` chuyển sang Bearer (helper `s.as(kind)`); `smoke.ps1` dùng `Authorization`; token hash round trip Postgres; revoke → 401.

**Docs.** WALKTHROUGH §18 "Ai đang gọi — auth là adapter, danh tính là port"; SETUP §5d curl với `-H "Authorization: Bearer dev-operator"`; DDD §20 (port `Verifier`), §30 tầng 1; README.

**Ước lượng:** 1 buổi. **Rủi ro:** đụng nhiều test HTTP hiện có (14 file) — làm helper `a.as("operator")` trước, đổi từng file.

### T2 — Read model bảng riêng: `GET /me/orders`, `GET /orders?status=`

**Mục tiêu.** Màn hình "đơn của tôi" và "tất cả đơn theo trạng thái" đọc **một bảng bẹt**, dựng từ event của **năm** context — CQRS đúng nghĩa (DDD.md §24).

**Business rule.**
- Khách chỉ thấy đơn của mình; operator thấy tất cả, lọc theo `status`.
- Một dòng `order_summaries` mỗi đơn: `order, customer, product_name, variant (size/colour nếu có), status, total, deposit, deposit_paid, balance_paid, shop_reference, tracking (none|expected|received|batched|shipped), placed_at, delivered_at, cancelled_at, refund`.
- **Chỉ event ghi vào bảng này**; không đọc bảng ghi của context nào. Idempotent, chịu sai thứ tự (dòng "khung" tạo bởi event đầu tiên tới, kể cả khi đó không phải `order_placed`).

**Thiết kế.**
- Package `internal/app/reporting` (`reportingapp`) — **không có domain**: read model không có invariant. `OrderSummary` struct exported; port `OrderSummaryRepository{ByOrder, ByCustomer(customer) []OrderSummary, ByStatus(status) []OrderSummary, Save}`; `Projector` với `On*` cho: `catalog.product_published` (tên sản phẩm → bảng phụ `product_names`), `ordering.order_placed`, `ordering.deposit_paid`, `procurement.purchase_confirmed` (reference), `ordering.order_purchased`, `logistics.parcel_expected/parcel_received` (tracking), `logistics.batch_shipped`, `ordering.order_shipped`, `ordering.balance_paid`, `ordering.order_delivered`, `ordering.order_cancelled`.
- Contracts mới: `OrderPurchasedV1`, `OrderShippedV1`, `BalancePaidV1`, `OrderDeliveredV1`, `OrderCancelledV1`, `ParcelExpectedV1`, `ParcelReceivedV1` (+ decoder mỗi cái; guard codec đã có).
- Postgres `0007_reporting.sql`: `order_summaries` (PK order; index customer; index status), `product_names`. Memory repo.
- HTTP `GET /me/orders` (customer), `GET /orders?status=deposited` (operator); view = struct tags JSON. `wire.Graph.Reporting`; `Subscribe` +11 dòng.

**Test.** projector: mỗi event hai lần → một dòng; `batch_shipped` trước `order_placed` → dòng khung rồi điền; `wholeFlow`: sau mỗi relay, `GET /me/orders` phản chiếu trạng thái đúng; Postgres round trip.

**Docs.** WALKTHROUGH §19 "Read model thật — một bảng, năm nguồn"; DDD §24 chuyển ✅ hoàn toàn; SETUP.

**Ước lượng:** 1 buổi.

### T3 — Quét quote hết hạn (worker)

**Rule.** Quote `issued` quá `expires_at` → `Expire(now)` → `pricing.quote_expired`. Chạy định kỳ trong `cmd/worker` (cờ `-sweep 1m`, batch 100). Idempotent: `Expire` từ chối quote không còn `issued` — handler bỏ qua `ErrQuoteNotIssued`. Ordering **không** cần nghe (đơn chỉ đặt trên quote đã accept).

**Thiết kế.** Port `pricing.QuoteRepository` thêm `IssuedBefore(ctx, t time.Time, limit int) ([]*Quote, error)` (memory + Postgres; index `quotes_open_idx` đã có). `pricingapp.ExpireQuotesHandler.Handle(ctx) (expired int, err error)`: `now` → `IssuedBefore(now)` → từng quote `Expire` → Save → outbox, **mỗi quote một transaction** (một quote hỏng không chặn cả lượt). `cmd/worker`: goroutine `ticker` gọi handler, log số quote đã hết hạn.

**Test.** memory: 3 quote (1 quá hạn, 1 còn hạn, 1 đã accept) → 1 expired, 1 event; lượt hai → 0. Postgres `IssuedBefore` round trip. `wire_test`: `clock.Advance(49h)` → sweep → `GET /quotes/{id}` = `expired`.

**Ước lượng:** 2 giờ.

### T4 — Endpoint tỷ giá (operator): `POST /fx`

**Rule.** `{from, to, rate}`; rate > 0, `from ≠ to`; ghi đè tỷ giá hiện tại (quote đã phát giữ tỷ giá cũ — ảnh chụp). **Không event** (không ai nghe) — nhưng ghi `updated_at` (đã có). ❓ Nếu muốn lịch sử: bảng `fx_rate_history` append-only — đề xuất **chưa**.

**Thiết kế.** `pricingapp.SetExchangeRateHandler` (mustHave `Rates`), route operator-only, `errorTable` thêm `shared.ErrInvalidRate → 400 invalid_rate` (kiểm tra tên sentinel thật trong `shared/rate.go`). Seed `wire` chuyển qua handler cho nhất quán.

**Ước lượng:** 1 giờ.

### T5 — `adapter/openai`: trích xuất listing từ link (ACL cho "web lạ")

**Mục tiêu.** Khách dán link → hệ thống điền nháp tên/giá/tiền tệ/category gợi ý; operator vẫn phải **confirm-listing** (rule catalog không đổi).

**Business rule.**
- Kết quả AI là **nháp chưa xác minh**: provenance = `SourcedByFeed` (máy đưa vào, chưa ai vouch) — **không** thêm `SourcingMode` mới (❓ nếu muốn phân biệt "feed" và "ai" trong báo cáo, thêm `SourcedByAI = "ai"` — đề xuất chưa; tránh chạm domain catalog).
- Không có API key → route trả **503** `extractor_unavailable`; **không** rơi về Fake trong Postgres mode. Memory mode dùng `Fake` để dev không cần key.
- Không bao giờ để nội dung trang web/AI **ghi đè** dữ liệu operator đã confirm: chỉ tạo **nháp mới**.
- Guard 6 đã cấm domain import SDK; adapter dùng `net/http` thuần (không SDK) để không thêm dependency.

**Thiết kế.** Port trong `catalogapp`: `ListingExtractor interface{ Extract(ctx, url catalog.SourceURL) (ListingDraft, error) }`, `ListingDraft{Name string; Price string; Currency string; CategoryHint string}`; `ErrExtractorUnavailable`. Use case `DraftFromURL{Merchant, URL}` → extractor → `AddProduct` (category: hint nếu tồn tại, không thì bắt buộc trong body ❓ đề xuất: body có `category` bắt buộc, hint chỉ để trả về cho UI). Adapter `internal/adapter/openai/`: `Client{Base, Key, Model, HTTP}` gọi `POST /v1/chat/completions` với `response_format: json_schema` (name, price, currency); `Fake{Draft}`; `Unavailable{}` (trả `ErrExtractorUnavailable`). Route `POST /products/from-url {merchant_id, url, category}` (customer/operator). `wire`: memory → `Fake`; Postgres → `Client` nếu `OPENAI_API_KEY` set, không thì `Unavailable`.

**Test.** `httptest.Server` giả OpenAI: 200 JSON hợp lệ → nháp; 200 JSON rác → lỗi rõ; 429/500 → lỗi có mã; timeout 10 s; `Unavailable` → 503 ở HTTP.

**Docs.** WALKTHROUGH §20 "ACL thứ hai: web lạ qua AI"; DDD §22 (ACL #2), §7 context map (adapter openai); SETUP §4 + biến `OPENAI_API_KEY`.

**Ước lượng:** nửa buổi. **Rủi ro:** không có key để smoke thật — smoke dùng Fake, ghi rõ.

### T6 — ACL API thật cho một shop (stretch, chỉ khi có API)

**Rule.** Cùng port `procurement.MerchantACL`; adapter `adapter/merchant/<shop>.go`; `wire` chọn adapter **theo shop** (map merchant → ACL; mặc định `Manual`). Biên nhận từ API vẫn qua `PurchaseTask.Confirm` (mọi rule biên nhận giữ). Lỗi mạng → retry (relay); từ chối nghiệp vụ → fail.

**Thiết kế.** `wire`: `procurement.MerchantACL` → `merchant.Router{byMerchant map[shared.ID]MerchantACL, fallback Manual}`; `OpenTaskHandler` không đổi. Test với `httptest`.

### T7 — `docs/FLOW-ORDER.md` — một đơn từ đầu tới cuối (docs, nên làm sớm)

Dàn bài (mỗi bước: request → handler → method aggregate → event → ai nghe → bảng nào đổi → mã lỗi có thể gặp):

1. Bản đồ 5 context, 16 dòng `Subscribe`, 20 event (chép transcript `smoke.ps1` §17i).
2. Bước 1–7 catalog (đăng ký shop → dán link → variant → confirm → measure → publish) — WALKTHROUGH §1, §14h.
3. Bước 8–9 pricing (quote → accept), số vàng 5 393 720.
4. Bước 10–11 ordering (đặt → cọc), luật một quote một đơn, cọc đúng số.
5. Bước 12–13 procurement (task mở → confirm 163.22), điểm không thể quay đầu.
6. Bước 14–18 logistics (expected → cân → batch → ship 27.50), chia cước.
7. Bước 19–21 ordering (in_transit → balance → delivered).
8. Đối soát: variance −2.50 và cách đọc dấu.
9. **Nhánh rẽ**: accept trễ; đặt hàng trước relay; hết size (`purchase_failed` → huỷ hoàn đủ); huỷ sau mua (mất cọc); event tới hai lần (idempotent ở đâu).
10. Bảng "status → hành động hợp lệ" cho 5 aggregate.

### T8 — Dọn dẹp

- Unit test cho `adapter/memory` repo của 4 context sau (coverage 20 % → ~70 %).
- `README` bảng route đầy đủ (31 + P9).
- CI: chạy `scripts/smoke.ps1` tương đương trên Linux (bash) với service Postgres — ❓ đề xuất viết `scripts/smoke.sh`.
- Xoá `docs/P9-PLAN.md` khi xong hoặc chuyển thành "đã làm" trong SETUP §7.

---

## 4. Những chỗ còn mở (❓) — đã có đề xuất, chốt rồi làm

| # | Câu hỏi | Đề xuất |
|---|---|---|
| 1 | Khách đọc đơn người khác: 403 hay 404? | **404** — không lộ sự tồn tại |
| 2 | Kiểm "chủ đơn" ở adapter hay handler? | **handler** (`Customer` vào command, `ErrNotOwner`) — luật không nằm ở adapter |
| 3 | Token operator đầu tiên trên Postgres? | cờ `-bootstrap-operator-token` chỉ khi bảng rỗng |
| 4 | Lịch sử tỷ giá? | chưa; `updated_at` đủ cho v1 |
| 5 | Provenance cho nháp AI: `feed` hay mode mới `ai`? | **`feed`** — không chạm domain catalog |
| 6 | Không có `OPENAI_API_KEY` trên Postgres mode? | 503 `extractor_unavailable`, không rơi về Fake |
| 7 | Ai quyết định sau `purchase_failed` (hoàn hay đơn mới)? | v1: khách/operator gọi `POST /orders/{id}/cancel` (hoàn đủ); saga tự động **không** làm — quyết định con người |
| 8 | Nhiều parcel một đơn / nhiều đơn một task? | **không** ở v1; `UNIQUE` giữ nguyên; ghi vào "Việc sau" |

---

## 5. Việc sau P9 (không làm bây giờ)

Cổng thanh toán thật (webhook → `PayDeposit/PayBalance`) · thông báo khách (email/Zalo) từ read model · nhiều parcel/đơn · đổi màu/size sau `purchase_failed` bằng đơn mới liên kết · báo cáo vốn lưu động (DDD §24) từ `order_summaries` + `reconciliations` · identity context thật (user, role) thay `Static` tokens.
