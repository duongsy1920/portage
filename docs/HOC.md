# HOC.md — lộ trình học Go + DDD bằng chính project này

File này không dạy lại Go hay DDD. Nó nói **đọc gì, theo thứ tự nào, và làm gì
để biết là mình thật sự hiểu**.

Anh đang có sẵn 6 năm PHP/Symfony. Nên lộ trình này viết theo hướng *"cái này ở
Symfony là gì, và vì sao ở đây nó khác"* — không phải học lại từ đầu.

**Quy tắc học duy nhất trong file này:**

> Mỗi buổi, đọc xong phải **sửa được một dòng và làm test đỏ**. Đọc mà không
> phá được cái gì thì chưa phải hiểu, mới là quen mặt.

---

## Bốn file, bốn câu hỏi khác nhau

| File | Trả lời | Đọc khi nào |
|---|---|---|
| [SETUP.md](SETUP.md) | *"máy tôi chạy nó thế nào, code nằm ở đâu, quy ước gì"* | trước tiên, và tra lại mỗi ngày |
| [GO-CHO-PHP.md](GO-CHO-PHP.md) | *"cú pháp này ở PHP là gì"* | tra khi vướng, đừng đọc một mạch |
| [WALKTHROUGH.md](WALKTHROUGH.md) | *"code chạy ra sao"* — file nào, dòng nào | xương sống của lộ trình dưới |
| [DDD.md](DDD.md) | *"vì sao lại thiết kế thế"* — khái niệm | đọc **sau** khi đã thấy code |
| [FLOW-ORDER.md](FLOW-ORDER.md) | *"một đơn hàng đi qua những đâu"* | đọc trước buổi 5, và trước khi phỏng vấn |

Thứ tự sai hay gặp nhất: đọc DDD.md trước. Đừng. Khái niệm không có code đi kèm
thì đọc xong quên ngay. Ở đây **luôn xem code trước, đọc lý do sau**.

---

## Buổi 1 — Go không có gì lạ, chỉ khác chỗ để lỗi

**Đọc:** WALKTHROUGH §1 (từ `curl` tới handler) + GO-CHO-PHP §1–4.

**Code:** `internal/domain/shared/money.go`, `decimal.go`, `id.go`.

Ba thứ khác PHP nhiều nhất, và đều nằm trong ba file trên:

```
   PHP                              Go
   ───                              ──
   throw new Exception              return err        ← lỗi là GIÁ TRỊ, không phải nhánh ngầm
   $obj->method()                   func (m Money) …  ← method gắn vào kiểu, không nằm trong class
   float cho tiền                   int64 minor units ← 0.1 + 0.2 != 0.3, và tiền thì không được sai
```

**Làm:**

1. `go test ./internal/domain/shared/` — 46 test, dưới một giây.
2. Mở `money.go`, đổi `roundHalfUp` thành làm tròn xuống. Chạy test. **Đếm xem
   mấy test đỏ** — rồi khôi phục.
3. Viết một hàm nhỏ trong `money_test.go` cộng hai `Money` khác tiền tệ. Xem
   nó bị từ chối thế nào và ở đâu.

**Biết là hiểu khi:** trả lời được *"vì sao `Money` không có setter?"* mà không
nhìn code.

---

## Buổi 2 — Value Object và Entity: cái gì có ID, cái gì không

**Đọc:** DDD.md §11 (Value Object), §12 (Entity), §13 (Invariant), §10
(Primitive Obsession).

**Code:** `shared/weight.go`, `catalog/freeshipping.go`, `catalog/merchant.go`.

Câu hỏi phân biệt, học một lần dùng mãi:

```
   Hai tờ 50 000 ₫ có khác nhau không?   → không  → VALUE OBJECT (Money, Weight)
   Hai khách tên Nguyễn Văn A?           → có     → ENTITY  (có ID riêng)
```

Ở Symfony anh hay thấy `Entity` cho **mọi thứ** vì Doctrine map như thế. Ở đây
`Money` không phải entity và không bao giờ vào bảng riêng — nó là hai cột
`amount_minor` + `currency` trong bảng của thằng sở hữu nó.

**Làm:**

1. Xoá một dòng validate trong `NewWeight` (cho phép cân âm). Chạy
   `go test ./internal/domain/...`. Đếm test đỏ, khôi phục.
2. Tìm trong `merchant.go` chỗ **duy nhất** một `Merchant` được tạo ra. Vì sao
   không có `merchant := &Merchant{}` ở đâu khác?

**Biết là hiểu khi:** chỉ ra được ba `Value Object` trong repo và nói vì sao
mỗi cái không cần ID.

---

## Buổi 3 — Aggregate: ranh giới của một transaction

**Đọc:** DDD.md §14 (Aggregate), §15 (Domain Event), §29 (vòng đời một
aggregate).

**Code:** `catalog/product.go` (đọc kỹ `Publish`), `catalog/events.go`.

`Product.Publish()` có **năm** điều kiện. Đọc từng cái và tự hỏi: *"cái này để
ở đâu nếu viết theo kiểu Symfony?"* — câu trả lời thường là "trong Service", và
đó chính là chỗ luật nghiệp vụ bị rải mỏng ra rồi mất.

**Ba câu hỏi để chọn ranh giới aggregate**, học thuộc:

```
   1. Cái gì phải ĐÚNG ngay lập tức, không được chờ?      → cùng một aggregate
   2. Cái gì chấp nhận đúng sau vài giây?                 → aggregate khác, nối bằng event
   3. Một transaction chạm mấy aggregate?                 → MỘT. Nhiều hơn là thiết kế sai.
```

**Làm:**

1. `go test ./internal/domain/catalog/` rồi mở `product_test.go` — 11 test cho
   một aggregate. Xem mỗi test canh **một** điều kiện.
2. Bỏ điều kiện `unverified` trong `Publish`. Chạy `go test ./...` (cả repo).
   Đếm xem có bao nhiêu test ở **bao nhiêu package** đỏ. Khôi phục.
3. Trả lời: vì sao `Variant` không phải aggregate root riêng?

---

## Buổi 4 — Port & Adapter: cái làm project này khác một CRUD

**Đọc:** DDD.md §16 (Repository), §20 (Ports & Adapters), §21 (Dependency
Rule). Rồi WALKTHROUGH §4, §6.

**Code:** `domain/catalog/repository.go` (port), `adapter/memory/*.go` và
`adapter/postgres/*.go` (hai adapter cho **cùng** một port).

Đây là chỗ Symfony và project này khác nhau nhất:

```
   Symfony   Entity dính Doctrine (#[ORM\Column] nằm ngay trong class nghiệp vụ)
   Portage   struct nghiệp vụ SẠCH; phần map xuống DB nằm riêng ở adapter/postgres
```

Đổi lại được cái gì? **262 trên 292 test chạy dưới 2 giây, không cần Docker,
không cần mạng, không cần API key.** Đó không phải khoe — đó là lý do anh sửa
được code mà không sợ.

**Làm:**

1. Chạy `go test ./internal/domain/` — đây là 7 **test canh kiến trúc**. Mở
   `decisions_test.go` và đọc guard 6.
2. Thêm `import "database/sql"` vào một file bất kỳ trong `internal/domain/`.
   Chạy lại. Xem guard bắt anh. Khôi phục.
3. Trả lời: `Clock` là port. Vì sao *"mấy giờ rồi"* lại là một dependency?

---

## Buổi 5 — Một đơn hàng đi hết năm context

**Đọc:** [FLOW-ORDER.md](FLOW-ORDER.md) — **cả file**, một mạch. Rồi DDD.md §6
(Bounded Context), §7 (Context Map), §27 (Outbox), §25 (Eventual Consistency).

**Làm trước khi đọc code:**

```powershell
docker compose up -d
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1
```

Nhìn 20 event chạy qua 5 context trên màn hình. Rồi mở
`internal/platform/wire/subscribe.go` — **31 dòng** đó chính là cái vừa chạy.

**Ba ý phải nắm:**

```
   1. Context KHÔNG import nhau. Chúng nói chuyện bằng event.
   2. Event ghi CÙNG transaction với dữ liệu (outbox) — không bao giờ mất.
   3. Người nhận có thể nhận HAI LẦN, và SAI THỨ TỰ. Nên ai cũng phải idempotent.
```

**Làm:**

1. Trong `smoke.ps1`, bỏ dòng `Start-Sleep` sau publish. Chạy. Xem
   `listing_not_found`. Đó **không phải bug** — đó là eventual consistency có
   mã trạng thái.
2. Mở `internal/app/pricing/projector.go`. Chỉ ra chỗ nào làm cho nó gọi hai
   lần cũng như một lần.

---

## Buổi 6 — Tiền: Quote vs Actual

**Đọc:** DDD.md §28, WALKTHROUGH §14 và §17.

**Code:** `pricing/calc.go` (Domain Service thuần), `pricing/quote.go`,
`shared/allocate.go`.

Đây là phần **nghiệp vụ thật của anh**, không phải bài tập:

```
   quoted  163.22 + 25.00 = 188.22 USD    mình ĐOÁN
   actual  163.22 + 27.50 = 190.72 USD    thật
   ─────────────────────────────────
   variance                −2.50 USD      ← ÂM = mình chịu
```

`variance` **luôn âm** qua nhiều đơn = bảng giá cước đang sai, không phải xui.
Đó là câu hỏi mà bảng `reconciliations` sinh ra để trả lời.

**Làm:**

1. Mở `shared/allocate_test.go`. Chia 87.50 cho hai thùng. Thử tự tính bằng
   `freight * ratio` rồi làm tròn từng cái — xem lệch mấy cent.
2. Đổi `divisor` của lane từ 5000 thành 6000 trong `wire.go`. Chạy `go test
   ./internal/platform/wire/`. Xem số vàng đổi thế nào và test nào bắt được.

---

## Buổi 7 — Ba thứ P9 thêm vào (và ba bài học lớn nhất)

**Đọc:** WALKTHROUGH §18 (auth), §19 (sweep), §20 (read model), §21 (ACL cho AI).

Ba câu, mỗi câu đáng nhớ hơn cả đoạn code sinh ra nó:

```
   §18   "Ai đang gọi" KHÔNG phải quy tắc nghiệp vụ — nó là việc của BIÊN,
         cùng lý do với Clock. Nên auth ở platform/, không ở domain/.

   §19   Có loại việc KHÔNG AI GỌI — chỉ thời gian đẩy. Và đó là lý do rõ nhất
         để Clock là port: 48 giờ trôi qua trong một dòng test.

   §20   Màn hình cần dữ liệu của bốn context → ĐỪNG JOIN. JOIN biến bốn
         bounded context thành một cục không tách nổi. Dựng một bảng bằng event.
```

**Làm:**

1. Trong `internal/app/reporting/projector.go`, xoá dòng `if s.Status == ""`.
   Chạy `go test ./internal/app/reporting/`. Đọc kỹ test đỏ — nó đang nói với
   anh điều gì về việc "đoán" một trạng thái?
2. Gọi `POST /products/from-url` trên bản Postgres không có `OPENAI_API_KEY`.
   Xem 503. Rồi tự hỏi: nếu nó rơi về `Fake` thì hỏng chỗ nào?

---

## Buổi 8 — Đọc hết repo

WALKTHROUGH §10: **93 file, theo thứ tự**, mỗi dòng ghi rõ file → khái niệm →
đọc thêm ở đâu. Khoảng 12 000 dòng kể cả comment.

Đừng đọc một mạch. Chia:

```
   file 1–35    ngày 1: một request đi hết 5 tầng, tới dòng sent_at trong Postgres
   file 36–56   ngày 2: context thứ hai, báo giá
   file 57–86   ngày 3: ba context còn lại, chia cước, đối soát
   file 87–93   P9: cửa auth, sweep, read model, ACL cho AI
```

---

## Tự kiểm tra — trả lời không nhìn code

Trả lời trôi chảy 15 câu này là đủ để nói về project trong phỏng vấn.

**Go**

1. Vì sao tiền là `int64` chứ không `float64`?
2. `error` là giá trị — điều đó đổi cách viết code thế nào so với exception?
3. `interface` ở Go được khai báo ở đâu: nơi CẦN hay nơi CÀI? Vì sao?
4. Vì sao `internal/` là một thư mục có ý nghĩa với compiler?

**DDD**

5. Aggregate root là gì, và ranh giới của nó quyết định điều gì về transaction?
6. Value Object khác Entity ở đúng một điểm — điểm nào?
7. Bounded Context: hai context cùng có "Product", chúng có phải cùng một thứ?
8. Domain Event khác một message queue thường ở chỗ nào?
9. Anti-Corruption Layer bảo vệ khỏi cái gì? Kể hai cái trong repo này.
10. CQRS: vì sao read model **không có** invariant?

**Hệ thống**

11. Outbox giải quyết vấn đề gì mà "lưu xong rồi publish" không giải quyết được?
12. At-least-once nghĩa là subscriber phải có tính chất gì?
13. Vì sao khách hỏi đơn người khác trả **404** chứ không phải 403?
14. `variance` âm nghĩa là gì, và bao nhiêu đơn âm liên tiếp thì phải xem lại
    bảng giá?
15. Tính năng AI tắt (không có API key) thì API trả gì, và vì sao **không**
    được rơi về Fake?

---

## Nói gì khi phỏng vấn

Nói **đúng** phần đã làm. Đoạn dưới là sự thật, kiểm chứng được bằng repo:

> Một hệ thống mua hộ xuyên biên giới viết bằng Go theo DDD: **năm bounded
> context** (catalog, pricing, ordering, procurement, logistics) cộng một tầng
> đọc, nối nhau bằng **36 domain event qua outbox** và một Published Language —
> không context nào import context nào. **Bảy aggregate root**, hai domain
> service, **hai anti-corruption layer** (một cho API shop, một cho mô hình
> ngôn ngữ đọc trang web), auth bằng bearer token với port ở tầng biên chứ
> không ở domain, và một **read model dựng chỉ bằng event** cho màn hình khách.
> **292 test**, trong đó 262 chạy dưới 2 giây không cần Docker vì domain không
> import gì ngoài stdlib — và có **7 test canh kiến trúc** bằng `go/ast` khiến
> vi phạm dependency rule là build đỏ. Vòng đời một đơn chạy hết trên **cả**
> in-memory và Postgres bằng **cùng một test**, và trên **binary thật** bằng
> một smoke script trong CI.

Đừng nói: "có microservices" (không có — một binary), "dùng AI để tự động hoá
mua hàng" (không — AI chỉ đọc trang, người vẫn phải xác nhận), "production-
ready" (chưa — chưa có cổng thanh toán).

**Thứ đáng kể nhất trong project này không phải là code, mà là những chỗ đã cố
tình KHÔNG làm** — và anh giải thích được vì sao. Ví dụ:

- không cào trang shop (điều khoản cấm) → khách/operator nhập, hoặc AI đọc từ URL;
- không tự động huỷ đơn khi shop hết hàng → con người quyết định;
- không cho AI xác nhận listing → `Publish` vẫn đòi một người thật;
- không rơi về Fake khi thiếu API key → 503, vì tính năng tắt phải trông như đang tắt.

---

## Sau khi học xong: ba việc tiếp theo đáng làm

1. **`config/ratecard.yaml`** — bảng giá cước đang là hằng số trong `wire.go`.
   Đưa ra file là bài tập tốt về "dữ liệu ≠ code" (DDD §23).
2. **Cổng thanh toán thật** — webhook → `PayDeposit`/`PayBalance`. Command đã
   có sẵn, chỉ thiếu adapter. Đây là bài tập Port & Adapter chuẩn nhất.
3. **Đọc lại `variance` sau 20 đơn thật** — nếu luôn âm thì `RateCard` sai.
   Đó là lúc project bắt đầu trả lời câu hỏi kinh doanh, không chỉ câu hỏi kỹ
   thuật.
