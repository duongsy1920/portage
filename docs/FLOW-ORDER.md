# FLOW-ORDER.md — một đơn hàng, từ cái link tới đồng bạc cuối

File này trả lời **một** câu hỏi, thật kỹ: *"khách dán một cái link, rồi chuyện
gì xảy ra?"*

Đọc kèm:

| File | Trả lời |
|---|---|
| [WALKTHROUGH.md](WALKTHROUGH.md) | đọc **code** theo thứ tự — file nào, dòng nào |
| [DDD.md](DDD.md) | **khái niệm**: Aggregate, Event, ACL, CQRS, Saga… |
| file này | **nghiệp vụ chạy**: mỗi bước ai gọi, ai nghe, bảng nào đổi, hỏng thì sao |

Mỗi bước dưới đây có đúng sáu ô:

```
   request  →  handler  →  method của aggregate  →  event
                                                      ↓
                                          ai nghe  ·  bảng nào đổi  ·  lỗi có thể gặp
```

---

## 0. Bản đồ

Năm bounded context, không context nào import context nào. Chúng chỉ nói
chuyện bằng **event** đi qua bảng `outbox` và một cái `Relay` (DDD.md §27).

```
   CATALOG ──product_published──▶ PRICING ──quote_accepted──▶ ORDERING
      │                              ▲                           │
      │                              │                     deposit_paid
      └──merchant_registered────▶ PROCUREMENT ◀──────────────────┘
                                     │
                            purchase_confirmed
                                     │
                    ┌────────────────┼────────────────┐
                    ▼                ▼                ▼
                ORDERING        LOGISTICS          PRICING
              (điểm không       (chờ thùng)      (giá THẬT đã trả)
               quay đầu)            │
                                batch_shipped
                                     │
                        ┌────────────┴────────────┐
                        ▼                         ▼
                    ORDERING                   PRICING
                  (in_transit)            (cước THẬT) → variance

              …và MỌI event ở trên ──▶ REPORTING (order_summaries)
```

**REPORTING** là tầng đọc (CQRS, DDD.md §24): nó nghe cả năm context và ráp
lại thành **một dòng phẳng** cho màn hình "đơn của tôi". Nó không quyết định
gì cả — không có invariant nào để giữ.

Bảng định tuyến thật: `internal/platform/wire/subscribe.go`, **35 dòng**. Đọc
từ trên xuống chính là vòng đời một đơn.

Một event có thể có nhiều người nghe. `procurement.purchase_confirmed` có
**bốn**: ordering (qua điểm không quay đầu), logistics (chờ thùng), pricing
(ghi giá thật), reporting (mã đơn ở shop).

---

## 1. Bước 1–7 — CATALOG: từ cái link thành một sản phẩm bán được

### Bước 0 (một lần): mở cửa

```
POST /tokens {kind:"customer"}          operator cắt chìa → token trả về ĐÚNG MỘT LẦN
```

Mọi request từ đây đều cần `Authorization: Bearer <token>`. Không có → **401**.
Đúng token sai loại → **403**. Chi tiết: WALKTHROUGH §18.

### Bước 1 — đăng ký shop

```
POST /merchants {name, site, currency, sourcing}     operator
  → catalogapp.RegisterMerchantHandler
  → catalog.RegisterMerchant(details, now)           ← AGGREGATE ROOT thứ nhất
  → event catalog.merchant_registered
      nghe: procurement (Shop projection — shop này tính tiền bằng gì)
      bảng: merchants
      lỗi:  400 invalid_hostname · 400 unknown_sourcing · 400 currency_required
```

`currency` ở đây là **luật**, không phải trang trí: bước 2 sẽ từ chối một sản
phẩm có giá khác tiền tệ của shop (409 `price_currency`).

### Bước 2 — khách dán link

```
POST /products {name, merchant_id, category, source_url, price, currency,
                requested_variant}                                        khách HOẶC operator
  → catalogapp.AddProductHandler
  → catalog.AddProduct(details, now)                 ← AGGREGATE ROOT thứ hai
  → event catalog.product_added  (mang cả link, giá, ai yêu cầu, xin size nào)
      nghe: reporting (dòng đầu tiên của hàng chờ việc)
      bảng: products
      lỗi:  409 merchant_inactive · 409 sourcing_not_allowed · 409 price_currency
            404 category_not_found · 400 invalid_source_url
```

Thứ tự kiểm có ý nghĩa: shop còn hoạt động không → **nguồn này có được phép không** → giá
có đúng tiền tệ → ngành hàng có tồn tại. Hai câu đầu là *"bạn có được phép hỏi không"*, hai
câu sau là *"bạn hỏi gì"*. `sourcing_not_allowed` là 409 chứ không phải 403: chìa khoá của
khách hợp lệ, chỉ là shop đó không nhận link do khách gửi.

Hai trường của **yêu cầu**, không phải của sản phẩm: `requested_by` lấy từ chìa khoá (khách
nào đang đợi), và `requested_variant` là size họ xin, bằng chữ của họ. Cả hai cố tình không
nằm trong `Provenance` — provenance trả lời *dữ liệu từ đâu tới và ai bảo đảm*, hai trường
này trả lời *ai đang đợi và họ xin gì*. `requested_variant` là một điều mong: shop có thể
không bán size đó, nên operator vẫn phải kiểm rồi tạo `Variant` thật.

**Ba luật chỉ tầng app kiểm được** (vì cần nhiều hơn một aggregate):

1. shop phải tồn tại và đang hoạt động;
2. giá phải đúng tiền tệ của shop;
3. category phải tồn tại.

**Ai dán quyết định `Provenance`** — và điều này quan trọng hơn nó trông:

```
   token operator  →  SourcedByOperator  →  Verified() = true
   token khách     →  SourcedByCustomer  →  Verified() = false
   POST /products/from-url (AI)  →  SourcedByFeed  →  Verified() = false
```

`Product.Publish()` **từ chối** listing chưa `Verified()`. Nên một cái link
khách dán, hay một dòng AI đoán ra, **không thể** thành giá bán cho tới khi có
người thật xác nhận.

Dán trùng link đã có → vẫn 201, nhưng sản phẩm mới bị **cắm cờ nghi trùng**
(`ProductFlaggedDuplicate`) — CATALOG.md §7 phương án C: không chặn khách, chỉ
đánh dấu cho operator.

### Bước 2b — hoặc để máy đọc hộ (P9/T5)

```
POST /products/from-url {merchant_id, url, category}    khách HOẶC operator
  → catalogapp.DraftFromURLHandler
  → catalogapp.ListingExtractor (PORT) → adapter/openai
  → …rồi đi tiếp đúng đường bước 2
      lỗi:  503 extractor_unavailable (không có key / model chết / trả rác)
            400 malformed_amount (model trả "about $150")
```

`category` **vẫn bắt buộc** dù model có trả `category_hint`: category quyết
thuế và cân nặng ước lượng — tiền thật. Gợi ý của máy trả về cho người xem, nó
không tự chọn gì cả.

### Bước 3–5 — ba bước của operator

```
POST /products/{id}/variants  {size, color, merchant_ref}  → 201   Variant (ENTITY con của Product)
      event catalog.variant_added   (size là CHỮ TỰ DO: "M 8 / W 9.5", "42 EU", "XS")
      nghe: ordering (có thật? của ai?), procurement (size/màu/mã shop để in ra)
      lỗi:  409 duplicate_variant   (khoá bỏ hoa/thường và MỌI khoảng trắng)
            409 unnamed_variant     (không size không màu → chỉ được là variant duy nhất)
POST /products/{id}/confirm-listing           → 204   Provenance listing → Verified
POST /products/{id}/measure   {weight_g, …}   → 204   ParcelSpec đo THẬT
      event catalog.product_measured
      nghe: pricing (Listing projection)
```

Ai "confirm" là **chủ token**, không phải một header người gọi tự điền (§18e).

`variant_added` tồn tại vì hai context khác cần biết size mà **không được đọc bảng
của catalog**: ordering để từ chối một id nó chưa từng nghe, procurement để người
đi mua đọc được "US 9 · black" thay vì một uuid (WALKTHROUGH §22).

### Bước 6 — publish

```
POST /products/{id}/publish   → 204
  → catalog.Product.Publish(now)
      NĂM điều kiện, tất cả ở DOMAIN:
        1. đang là draft            → 409 not_draft
        2. có ít nhất một variant   → 409 no_variants
        3. listing đã Verified      → 409 unverified
        4. có giá                   → 400 price_required
        5. không bị cắm cờ trùng    → 409 suspected_duplicate
  → event catalog.product_published (mang Price + ParcelSpec)
      nghe: pricing (Listing), procurement (Item + link gốc), reporting (tên sản phẩm)  ← BA người nghe
      bảng: products
```

### Bước 7 — worker chuyển tin

Chưa có gì tới pricing cho tới khi `Relay` chạy. Gọi `POST /quotes` ngay lập
tức → **404 `listing_not_found`**. Đó không phải bug: đó là **eventual
consistency có mã trạng thái** (DDD.md §25).

---

## 2. Bước 8–9 — PRICING: báo giá là một bức ảnh

### Bước 8 — báo giá

```
POST /quotes {product_id, lane}      khách hoặc operator
  → pricingapp.IssueQuoteHandler
  → pricing.IssueQuote(inputs, now)                  ← AGGREGATE "ảnh chụp"
  → event pricing.quote_issued
      bảng: quotes (từng cột: từng dòng tiền, tỷ giá, cân tính cước)
      lỗi:  404 listing_not_found · 409 listing_inactive · 409 no_exchange_rate · 409 nothing_to_weigh
```

**Số vàng** (giày 150 USD, thùng 1250 g / 34×23×13 cm, lane `us_forwarder`):

```
   item            150.00 USD
   sales_tax        13.22        8,81 % — kho Denver, Colorado
   freight          25.00        cân tính cước 2500 g × 10.00 USD/kg (branded)
   ────────────────────────
   subtotal        188.22 USD
   × fx 26 000   4 893 720 VND
   + service_fee   500 000       max(10 %, sàn 500k)
   ────────────────────────
   TOTAL         5 393 720 VND
   deposit       2 696 860       50 %
```

**Cân tính cước** = `roundUp(max(cân thật, quy đổi thể tích), bước 500 g)`,
quy đổi = `dài×rộng×cao (mm³) / 5000`. Thùng to mà nhẹ vẫn phải trả tiền cho
chỗ nó chiếm.

Quote **đóng băng** cả tỷ giá lẫn bảng giá đang dùng. Đổi tỷ giá ngày mai
(`POST /fx`) → quote hôm nay **không đổi một chữ**. Đó là nửa "Quote" của
Quote vs Actual (DDD.md §28).

### Bước 8b — quote không ai trả lời

```
(không có request nào — THỜI GIAN đẩy)
  → worker.Sweeper mỗi phút → pricingapp.ExpireQuotesHandler
  → pricing.Quote.Expire(now)      chỉ quote còn `issued` và đã quá 48 h
  → event pricing.quote_expired
```

Chi tiết: WALKTHROUGH §19. Đây là use case duy nhất trong hệ **không có người
gọi**.

### Bước 9 — khách chấp giá

```
POST /quotes/{id}/accept    KHÁCH (operator gọi → 403)
  → pricing.Quote.Accept(now)
  → event pricing.quote_accepted
      nghe: ordering (AcceptedQuote projection)
      lỗi:  409 quote_expired (quá 48 h — và trạng thái expired VẪN ĐƯỢC LƯU)
            409 quote_not_issued (bấm lần hai)
```

Chỗ tinh tế: **accept muộn vừa đổi trạng thái vừa bị từ chối**. Tầng app
commit transaction rồi mới trả lỗi — nếu rollback thì quote sẽ "sống lại" thành
`issued` và khách bấm lại được vào giá cũ.

---

## 3. Bước 10–11 — ORDERING: cọc 50 %

### Bước 10 — đặt hàng

```
POST /orders {quote_id, variant_id}      KHÁCH (không còn field customer_id)
  → orderingapp.PlaceOrderHandler
  → ordering.PlaceOrder(details, now)                ← AGGREGATE ROOT
  → event ordering.order_placed
      nghe: pricing (Reconciliation: đơn này dựa trên quote nào), reporting
      bảng: orders (UNIQUE(quote))
      lỗi:  409 quote_not_accepted (chưa accept HOẶC relay chưa chạy)
            409 quote_already_used  (một quote một đơn)
            409 variant_unknown     (id chưa từng phát HOẶC variant_added chưa relay tới)
            409 variant_not_for_product (variant có thật, nhưng của sản phẩm khác)
```

`variant_id` là **thứ duy nhất khách tự chọn** trong request này, nên là thứ duy
nhất được kiểm: nó phải có trong bảng chiếu của ordering, và phải thuộc đúng sản
phẩm mà quote đã báo giá. Thứ tự kiểm: quote → một-quote-một-đơn → variant.

Luật *"một quote một đơn"* được canh **hai lần**: use case hỏi `ByQuote` trước,
và `UNIQUE(quote)` trong Postgres bắt trường hợp hai request cùng lúc. Luật
nghiệp vụ ở tầng app; DB là cái lưới cuối.

Chủ đơn là **chủ token**. Không còn field nào để đặt hộ tên người khác.

### Bước 11 — cọc

```
POST /orders/{id}/deposit {amount, currency}    operator (tới khi có cổng thanh toán)
  → ordering.CustomerOrder.PayDeposit(amount, now)
      SỐ PHẢI ĐÚNG CHÍNH XÁC → 409 wrong_amount
  → event ordering.deposit_paid
      nghe: procurement (mở việc mua), reporting
```

`"2.696.860"` với `Accept-Language: vi` = 2 696 860 ₫. Chuẩn hoá theo locale
nằm ở **adapter**, domain không đoán (WALKTHROUGH §1c).

---

## 4. Bước 12–13 — PROCUREMENT: đi mua, và điểm không thể quay đầu

### Bước 12 — hệ thống mở một việc

```
(không có request — event ordering.deposit_paid đẩy)
  → procurementapp.OpenTaskHandler
  → hỏi PORT procurement.MerchantACL   ← ANTI-CORRUPTION LAYER (DDD.md §22)
        adapter merchant.Router chọn theo shop:
          shop có API   → adapter riêng của shop
          shop chưa có  → merchant.Manual → ErrManualPurchase
  → dựng Subject: tên sản phẩm + size/màu + mã shop + link gốc
        CHÉP từ hai bảng chiếu của chính procurement (Item, Variant), không JOIN
        thiếu dòng Variant (chưa relay) → task VẪN mở, chỉ là không có size
  → procurement.OpenTask(details, now)               ← AGGREGATE ROOT
  → event procurement.purchase_task_opened
      bảng: purchase_tasks (UNIQUE("order") — một đơn một việc)
```

`OpenTask` **idempotent theo đơn**: relay at-least-once có thể gửi
`deposit_paid` hai lần, và hai việc mua cho một đơn nghĩa là mua hai đôi giày.

`Subject` được **chép** vào task, không tra lại lúc đọc: việc đã giao cho người
thì không đổi nội dung, kể cả khi shop rename sản phẩm ngày mai.

### Bước 13 — người mua xác nhận

```
GET  /purchase-tasks                       operator — danh sách việc, cũ nhất trước
      mỗi dòng: product_name · variant_label · variant_ref · source (Subject đã chép)
POST /purchase-tasks/{id}/confirm {reference, paid, currency}
  → procurement.PurchaseTask.Confirm(receipt, now)
      biên nhận gắt: mã đơn không rỗng, tiền > 0, ĐÚNG tiền tệ của shop, có người
  → event procurement.purchase_confirmed
      nghe: BỐN người —
            ordering  → OrderPurchased  (ĐIỂM KHÔNG THỂ QUAY ĐẦU)
            logistics → ExpectParcel    (kho chờ một thùng)
            pricing   → ActualGoods     (tiền THẬT đã trả: 163.22 USD)
            reporting → shop_reference
      lỗi:  409 purchase_task_not_open · 409 paid_currency
```

Hoặc:

```
POST /purchase-tasks/{id}/fail {reason:"hết size US 10"}
  → event procurement.purchase_failed
      nghe: ordering → OrderPurchaseFailed
```

**Điểm không thể quay đầu** là luật nghiệp vụ có tiền thật đằng sau:

```
   huỷ TRƯỚC purchase_confirmed  →  refund = TOÀN BỘ cọc,  forfeited = false
   huỷ SAU  purchase_confirmed   →  refund = 0,            forfeited = true
   purchase_failed rồi mới huỷ   →  refund = TOÀN BỘ cọc   (lỗi của mình, không phải của khách)
```

Đây là nửa "compensation" của Saga (DDD.md §26). Nửa "điều phối" cố tình
**không tự động**: sau `purchase_failed`, con người quyết định hoàn tiền hay
đặt lại (P9-PLAN §4, câu 7).

---

## 5. Bước 14–18 — LOGISTICS: cân thật, gom lô, chia cước

```
GET  /parcels                                   operator — thùng đang chờ
POST /parcels/{id}/receive {weight_g, …}        → CÂN THẬT (Actual thứ hai)
      event logistics.parcel_received → reporting (tracking)
POST /batches {lane}                            mở một lô
POST /batches/{id}/parcels {parcel_id}          bỏ thùng vào lô — HAI aggregate, MỘT transaction
POST /batches/{id}/close                        chốt lô  (rỗng → 409 batch_empty)
POST /batches/{id}/ship {freight, currency}     → 200 [allocation]
      event logistics.batch_shipped
        nghe: ordering (→ in_transit), pricing (cước THẬT), reporting
```

**Chia cước** là chỗ dễ sai nhất trong cả hệ, nên nó là một Domain Service
riêng (`FreightAllocator`) với một thuật toán riêng (`shared.Allocate`):

```
   hoá đơn carrier   87.50 USD cho cả lô
   thùng A  2500 g  →  29.17
   thùng B  5000 g  →  58.33
   ───────────────────────────
   tổng                87.50   ← ĐÚNG TỚI TỪNG CENT
```

Chia theo tỷ lệ rồi làm tròn từng cái sẽ lệch 1 cent. `Allocate` chia sàn rồi
phát phần dư theo **largest remainder** — tổng luôn khớp hoá đơn.

Phần chia được **đóng băng vào event** `batch_shipped`, không tính lại sau: đổi
`LaneRule` tháng sau không được làm đổi số tiền đã tính cho lô tháng này.

---

## 6. Bước 19–21 — ORDERING đóng lại

```
GET  /orders/{id}                     → "in_transit" sau khi relay batch_shipped
POST /orders/{id}/balance {amount}    → 204   (409 not_in_transit nếu sớm quá)
POST /orders/{id}/deliver             → 204   (409 balance_unpaid nếu chưa trả nốt)
      event ordering.order_delivered → reporting
```

Vòng đời đầy đủ của `CustomerOrder`:

```
awaiting_deposit ─deposit─▶ deposited ─purchase_confirmed─▶ purchased
                                │                              │
                        purchase_failed                   batch_shipped
                                ▼                              ▼
                        purchase_failed                    in_transit ─balance+deliver─▶ delivered
                                                   
        …và từ bất kỳ trạng thái nào trước delivered: ─cancel─▶ cancelled
```

---

## 7. Đối soát: Quote vs Actual

```
GET /reconciliations/{order}        operator
```

```
   quoted  163.22 + 25.00 = 188.22 USD    (mình ĐOÁN lúc báo giá)
   actual  163.22 + 27.50 = 190.72 USD    (thật: shop đúng, carrier đắt hơn 2.50)
   ─────────────────────────────────────
   variance                −2.50 USD      ← DẤU ÂM = MÌNH CHỊU
```

**Cách đọc dấu:** `variance = quoted − actual`.

```
   variance ÂM   →  thực tế đắt hơn dự tính  →  ăn vào lãi
   variance DƯƠNG→  thực tế rẻ hơn           →  lãi thêm
```

Ở đây −2.50 USD ≈ 65 000 ₫, và phí dịch vụ 500 000 ₫ đã che nó. Nhưng nếu
`variance` **luôn âm** qua nhiều đơn thì bảng giá cước đang sai, không phải
xui — và đó chính là câu hỏi mà bảng này sinh ra để trả lời.

Một dòng `reconciliations` được ghi bởi **ba** context: ordering (quote nào),
procurement (giá hàng thật), logistics (cước thật). Chỉ khi đủ cả ba thì
`Complete() = true`.

---

## 8. Màn hình: một bảng, năm nguồn

```
GET /me/orders                    KHÁCH — đơn của chính mình, mới nhất trước
GET /orders?status=deposited      OPERATOR — hàng đợi việc
```

Một dòng `order_summaries`, ghi **chỉ bằng event**, từ năm context:

| Cột | Ai nói |
|---|---|
| `product_name` | catalog (`product_published`) |
| `status`, `total`, `deposit`, `refund` | ordering |
| `shop_reference` | procurement (`purchase_confirmed`) — **chỉ operator thấy** |
| `tracking` (none → expected → received → shipped) | logistics |
| `deposit_paid`, `balance_paid` | ordering |

Vì sao không JOIN năm bảng lúc đọc? Vì năm bảng đó thuộc **năm bounded
context**, và một câu JOIN qua chúng biến năm context thành một cục không tách
được nữa. Chi tiết: WALKTHROUGH §20.

`tracking` cố tình **không** gộp vào `status`: ordering sở hữu status (tiền và
việc mua), còn kho sở hữu chuyện cái thùng đi tới đâu.

---

## 9. Nhánh rẽ — chỗ hệ thống thật khác hệ thống demo

| Tình huống | Chuyện gì xảy ra | Vì sao |
|---|---|---|
| **Accept muộn** (>48 h) | 409 `quote_expired`, và quote **được lưu** thành expired | rollback thì khách bấm lại được vào giá cũ |
| **Không ai accept** | sweep theo giờ đổi sang `expired` | không có event "quote già đi" để nghe (§19) |
| **Đặt hàng trước khi relay chạy** | 409 `quote_not_accepted` | eventual consistency **có mã trạng thái**, không phải bug |
| **Hết size** | `purchase_failed` → đơn về `purchase_failed`; huỷ hoàn **đủ** | lỗi của mình |
| **Huỷ sau khi đã mua** | refund 0, `forfeited: true` | tiền đã ra khỏi túi mình rồi |
| **Khách xem đơn người khác** | **404**, không phải 403 | 403 xác nhận đơn đó có thật (§18d) |
| **Event tới hai lần** | không có gì đổi lần thứ hai | mọi projector idempotent (upsert theo khoá) |
| **Event tới sai thứ tự** | read model tạo dòng "khung" rồi điền dần | relay at-least-once **không** bảo đảm thứ tự giữa các luồng |
| **Outbox ghi lỗi** | cả transaction rollback — không lưu, không phát | `Save` và `Outbox.Append` **cùng một** transaction (§27) |
| **Worker chết giữa chừng** | event gửi lại lần sau | đánh dấu `sent_at` SAU khi publish (at-least-once) |
| **Không có OPENAI_API_KEY** | 503 `extractor_unavailable` | tính năng tắt phải **trông như** đang tắt |
| **Thu hồi chìa operator cuối** | 409 `last_operator_key` | API không ai quản được không phải API an toàn hơn |

---

## 10. Bảng "trạng thái → làm được gì"

### `catalog.Product`

| Trạng thái | Làm được | Không làm được |
|---|---|---|
| `draft` | thêm variant, confirm-listing, measure, đổi giá, publish | — |
| `published` | đổi giá, đo lại, retire | publish lại (409 `not_draft`) |
| `retired` | — | publish, báo giá (409 `listing_inactive`) |

### `pricing.Quote`

| Trạng thái | Làm được | Không làm được |
|---|---|---|
| `issued` | accept (trong 48 h), expire (sau 48 h) | — |
| `accepted` | đặt **một** đơn | accept lần hai (409 `quote_not_issued`) |
| `expired` | — | accept, đặt đơn |

### `ordering.CustomerOrder`

| Trạng thái | Làm được | Không làm được |
|---|---|---|
| `awaiting_deposit` | cọc (đúng số), huỷ → hoàn đủ | trả nốt, giao |
| `deposited` | huỷ → hoàn đủ; hệ thống mở việc mua | cọc lần hai |
| `purchased` | huỷ → **mất cọc** | quay lại `deposited` |
| `purchase_failed` | huỷ → hoàn **đủ** | giao |
| `in_transit` | trả nốt, rồi giao | huỷ hoàn tiền |
| `delivered` | — | mọi thứ (409 `already_delivered`) |
| `cancelled` | — | mọi thứ (409 `order_cancelled`) |

### `procurement.PurchaseTask`

| Trạng thái | Làm được | Không làm được |
|---|---|---|
| `open` | confirm (có biên nhận), fail (có lý do) | — |
| `confirmed` / `failed` | — | đóng lần hai (409 `purchase_task_not_open`) |

### `logistics.Parcel` / `ConsolidationBatch`

| Trạng thái | Làm được | Không làm được |
|---|---|---|
| parcel `expected` | receive (cân thật) | vào lô |
| parcel `received` | vào một lô đang mở | cân lần hai (409 `parcel_not_expected`) |
| parcel `batched` | đi theo lô | vào lô thứ hai (409 `duplicate_parcel`) |
| batch `open` | thêm thùng, đóng | ship (409 `batch_not_closed`) |
| batch `closed` | ship (có hoá đơn cước) | thêm thùng (409 `batch_not_open`) |
| batch `shipped` | — | mọi thứ |

---

## 11. Chạy thật cả flow trong 20 giây

```powershell
docker compose up -d
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1
```

```
merchant 01a0…/ product 01a0…
quote 01a0…: issued branded chargeable 2500 g, total 5393720 VND, deposit 2696860
order 01a0…: deposited; buyer's list: 1 open task(s)
order after purchase_confirmed: purchased; warehouse expects 1 parcel(s), ref NK-20260905-001
batch 01a0… shipped: order 01a0… billed on 2500 g → freight 27.50 USD
order at the end: delivered; quote vs actual: quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD
my orders: 1 row(s), Air Trainer 90 - delivered, tracking shipped, shop ref on customer view: ''
```

20 event, 5 context, 2 binary thật, 1 Postgres thật. `scripts/smoke.sh` là bản
Linux của đúng script này và chạy trong CI.
