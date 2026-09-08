# HOC.md — lộ trình học Go + DDD bằng chính project này

File này không dạy lại Go hay DDD. Nó nói **đọc gì, theo thứ tự nào, và làm gì
để biết là mình thật sự hiểu**.

Anh đang có sẵn 6 năm PHP/Symfony. Nên lộ trình này viết theo hướng *"cái này ở
Symfony là gì, và vì sao ở đây nó khác"* — không phải học lại từ đầu.

**Quy tắc học duy nhất trong file này:**

> Mỗi buổi, đọc xong phải **sửa được một dòng và làm test đỏ**. Đọc mà không
> phá được cái gì thì chưa phải hiểu, mới là quen mặt.

---

## Sáu chỗ đọc, sáu câu hỏi khác nhau

| File | Trả lời | Đọc khi nào |
|---|---|---|
| [SETUP.md](SETUP.md) | *"máy tôi chạy nó thế nào, code nằm ở đâu, quy ước gì"* | trước tiên, và tra lại mỗi ngày |
| [GO-CHO-PHP.md](GO-CHO-PHP.md) | *"cú pháp này ở PHP là gì"* | tra khi vướng, đừng đọc một mạch |
| [WALKTHROUGH.md](WALKTHROUGH.md) | *"code chạy ra sao"* — file nào, dòng nào | xương sống của lộ trình dưới |
| [DDD.md](DDD.md) | *"vì sao lại thiết kế thế"* — khái niệm | đọc **sau** khi đã thấy code |
| [FLOW-ORDER.md](FLOW-ORDER.md) | *"một đơn hàng đi qua những đâu"* | đọc trước buổi 5, và trước khi phỏng vấn |
| [UI-GUIDE.md](UI-GUIDE.md) | *"bấm cái gì để thấy nó chạy"* | buổi 0, trước khi đọc dòng code nào |
| [SETUP.md](SETUP.md) §9 | *"vì sao code lại trông như thế này"* | khi vướng một quyết định lạ; đọc 2–3 đợt cuối |

Thứ tự sai hay gặp nhất: đọc DDD.md trước. Đừng. Khái niệm không có code đi kèm
thì đọc xong quên ngay. Ở đây **luôn xem code trước, đọc lý do sau**.

`SETUP.md` xuất hiện hai lần trong bảng vì nó là hai file khác nhau đóng chung
một bìa: phần đầu là môi trường và quy ước, còn **§9 là nhật ký 18 đợt review**
— mỗi đợt một bảng *"quyết định hoặc bug → chỗ nó nằm"*. Khi một dòng code trông
vô lý, §9 thường là chỗ giải thích thẳng nhất, hơn cả DDD.md.

---

## Buổi 0 — Bấm trước, đọc sau

Nửa tiếng này tiết kiệm cho anh vài buổi. Đọc code mà chưa có bản đồ trong đầu
thì mỗi file là một hòn đảo; bấm qua UI trước thì mọi thứ đọc sau đó có chỗ để
gắn vào.

**Bấm, theo đúng thứ tự này:**

1. **`web/flow.html`** — mở bằng double-click, không cần build, **không gọi API**.
   26 bước của một đơn, mỗi bước hiện màn hình người dùng thật sự thấy, request,
   event ghi vào outbox, và ai nghe event đó. Bấm hỏng cũng không sao, nó là mô
   phỏng.
2. **Hai màn hình thật** — `go run ./cmd/api -web ./web` rồi mở
   `http://localhost:8080/ui/`. Mở trang **khách** ở một tab và trang **nhân viên** ở tab
   khác, rồi gửi một link: bạn sẽ thấy việc xuất hiện bên kia, checklist bốn bước, và thông
   báo bắn qua lại. Đó là vòng đời một đơn nhìn từ phía người dùng.
3. **Bảng kiểm API** — cùng địa chỉ, trang `console.html`. Tab **Runner** chạy 26 bước bằng
   request thật, và cột **Nhật ký gọi** cho thấy từng request với từng mã lỗi. Đây là công
   cụ cho lập trình viên, không phải UI cho người dùng — phân biệt được hai thứ đó cũng là
   một bài học.
4. **[UI-GUIDE.md](UI-GUIDE.md)** — bốn màn hình là gì, bấm gì trên từng cái, 10 nhánh rẽ
   nên thử bằng tay, và bảng triệu chứng suy ra nguyên nhân khi có gì đó không chạy.

**Ba thứ chỉ bấm mới thấy, đọc code rất khó nhận ra:**

```
   1. Worker là một BƯỚC RIÊNG. Không phải hàm được gọi — là một process khác,
      chạy sau. Sáu lượt relay của một đơn ra đúng 5/2/2/2/6/3 event.
   2. Bảng đọc TRỐNG cho tới khi relay chạy. Nên `POST /quotes` ngay sau publish
      trả 404, và đó là thiết kế, không phải lỗi.
   3. Một event có NHIỀU người nghe. `purchase_confirmed` có bốn.
```

**Biết là hiểu khi:** chỉ được trên màn hình chỗ nào là *"đã ghi nhưng chưa ai
biết"*, và chỗ nào là *"đã có người nghe rồi"*.

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

Đổi lại được cái gì? **272 trên 304 test chạy dưới 2 giây, không cần Docker,
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
docker compose up -d                                        # Windows
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1
```

```bash
docker compose up -d                                        # Linux
PORTAGE_DSN=postgres://portage:portage@localhost:5433/portage?sslmode=disable \
  PORTAGE_PORT=8081 bash scripts/smoke.sh
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

**Đọc thêm, ngay sau buổi này:** WALKTHROUGH **§22**. Đó là ví dụ mới nhất và
cụ thể nhất của đúng bài học trên: một chữ của shop, chữ `size`, phải đi qua ba
context mà không context nào được đọc bảng của context kia. Buổi 9 mổ kỹ nó.

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

WALKTHROUGH §10: **101 file, theo thứ tự**, mỗi dòng ghi rõ file → khái niệm →
đọc thêm ở đâu. Toàn bộ code không tính test là 18 132 dòng kể cả comment.

Đừng đọc một mạch. Chia:

```
   file 1–35    ngày 1: một request đi hết 5 tầng, tới dòng sent_at trong Postgres
   file 36–56   ngày 2: context thứ hai, báo giá
   file 57–86   ngày 3: ba context còn lại, chia cước, đối soát
   file 87–93   P9: cửa auth, sweep, read model, ACL cho AI
   file 94–97   variant/size: một chữ của shop đi qua ba context (§22)
   file 98–101  bảng đọc worklist + hai màn hình theo vai (§23)
```

---

## Buổi 9 — Ba thứ chỉ học được khi CHẠY THẬT

Bảy buổi trên học từ code đang có. Buổi này học từ hai lần code đang có **hoá ra
sai**, và cả hai đều chỉ lộ ra khi bấm thật hoặc chạy thật, không phải khi đọc.

### (a) Thêm một event vào hệ thống đang chạy — WALKTHROUGH §22

**Đọc:** WALKTHROUGH §22 cả tám mục, rồi CATALOG.md phần *"`size` là chữ của shop"*.

Bắt đầu từ một câu hỏi rất thường: *khách nhập size kiểu `M 8 / W 9.5` thì có
đỡ được không?* Trả lời được phần dễ ngay, `size` là string tự do. Nhưng lần
theo nó thì lộ ra hai lỗ hở lớn hơn nhiều:

```
   POST /orders      nhận BẤT KỲ uuid làm variant_id → khách trả cọc xong,
                     người đi mua nhận một task không thể thực hiện
   màn hình đi mua   bốn cái uuid và một mã tiền tệ, không ai cầm vào shop được
```

Cùng một nguyên nhân: `size` chưa bao giờ ra khỏi catalog. Và không sửa bằng một
câu `JOIN`: guard 7 chặn được việc context này **import** context kia, còn việc
đọc bảng của nhau thì không guard nào bắt được — nó chỉ sai vì cùng một lý do,
và dùng chung một database hôm nay là chuyện tiện, không phải chuyện được phép.
Nên catalog phải **kể**: `catalog.variant_added`.

Phần đáng học nhất là **hai người nghe chép khác nhau**:

```
   procurement   giữ 5 field (size, màu, mã shop…)   vì có NGƯỜI phải đọc
   ordering      giữ 2 field (variant, product)      vì luật chỉ cần "có thật? của ai?"
```

Projection giữ dữ liệu **không ai đọc** sẽ mốc mà không test nào bắt. Đó là lý
do ordering **cố tình** không giữ `size`.

**Làm:**

1. Mở `internal/domain/ordering/variant.go`. Thêm field `Size string` vào. Chạy
   `go test ./...`. Nó **vẫn xanh** — và đó chính là vấn đề. Tự trả lời: sáu
   tháng sau ai biết field đó còn đúng hay không?
2. Trong `internal/app/procurement/open_task.go`, đổi nhánh
   `ErrVariantNotFound` thành `return err`. Chạy test. Đọc kỹ test đỏ: nó đang
   nói gì về việc huỷ một đơn **đã trả cọc** vì một cuộc đua vài trăm ms?
3. Gọi `POST /orders` với một uuid tự bịa. Xem `409 variant_unknown`. Rồi tìm
   trong `web/app/steps.js` chỗ mã đó được đưa vào danh sách `retryOn`, và tự
   trả lời vì sao nó **thử lại được** mà `variant_not_for_product` thì không.

### (b) Một UI không có test là một UI chưa ai chạy — WALKTHROUGH §23h

**Đọc:** WALKTHROUGH §23, cả chín mục.

`go test ./...` xanh, `curl` đúng, và năm lỗi vẫn còn nguyên trong hai màn hình: `style`
truyền chuỗi thay vì object, gọi component như hàm nên sai quy tắc hooks, `<label>` không
nối với `<input>`, dòng tiền `0.00` vẫn hiện vì so chuỗi với `"0"`, và ngành hàng mặc định
đổi mỗi lần tải vì adapter in-memory duyệt map Go. Cả năm chỉ lộ ra khi cho Chrome bấm hết
luồng.

**Làm:**

1. Đọc §23h. Tự trả lời: trong năm lỗi đó, lỗi nào `go vet` bắt được? (Không cái nào.)
2. Mở `web/app/ui.js`, đổi `<${Field} ...${{...}} />` ở một chỗ thành `${Field({...})}`.
   Mở trang, xem console. Đó là React error #310, và là lý do component phải được dùng như
   component chứ không như hàm.

### (c) Chạy xanh nhiều lần không chứng minh gì về lần đầu — SETUP §9 đợt 18

**Đọc:** SETUP.md §9 đợt 18, rồi `internal/adapter/postgres/migrate.go`.

`Migrate` tạo bảng sổ sách `schema_migrations` bằng `CREATE TABLE IF NOT EXISTS`,
**ngoài** advisory lock. Câu đó **không nguyên tử** trong PostgreSQL: hai câu
song song đều thấy bảng chưa có, rồi một cái vỡ khi chèn row type của bảng —
`duplicate key value violates unique constraint "pg_type_typname_nsp_index"`.

`cmd/api` và `cmd/worker` khởi động cùng lúc, nên đây là đường đi thật của một
lần deploy đầu tiên. Bug này sống sót qua 291 test và qua rất nhiều lần smoke
xanh, vì:

```
   database đã migrate  →  CREATE TABLE IF NOT EXISTS là no-op  →  không có gì để đua
   database TRỐNG       →  hai process cùng tạo bảng            →  vỡ
```

Bài học đắt nhất của cả repo này gọn trong một câu: **"chạy xanh nhiều lần"
không chứng minh được gì về lần chạy ĐẦU TIÊN.**

**Làm:**

1. Đọc `TestMigrate_survivesTwoProcessesOnAColdDatabase`. Trả lời: vì sao nó
   phải tự tạo một database **dùng-một-lần** thay vì dùng `portage_test`?
2. Dịch câu `CREATE TABLE` trong `migrate.go` trở lại ra ngoài transaction. Chạy
   riêng test đó vài lần. Xem một test **xác suất** trông như thế nào, và tự trả
   lời vì sao chiều "không bao giờ đỏ oan trên code đúng" mới là chiều quan trọng.

---

## Tự kiểm tra — trả lời không nhìn code

Trả lời trôi chảy 17 câu này là đủ để nói về project trong phỏng vấn.

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
16. Hai context nghe **cùng một event**: vì sao một bên chép 5 field mà bên kia
    chỉ chép 2? Chép thừa thì hỏng chỗ nào?
17. Hai `CREATE TABLE IF NOT EXISTS` chạy song song trong PostgreSQL thì ra gì,
    và vì sao lỗi đó **chỉ** hiện trên database trống?

---

## Nói gì khi phỏng vấn

Nói **đúng** phần đã làm. Đoạn dưới là sự thật, kiểm chứng được bằng repo:

> Một hệ thống mua hộ xuyên biên giới viết bằng Go theo DDD: **năm bounded
> context** (catalog, pricing, ordering, procurement, logistics) cộng một tầng
> đọc, nối nhau bằng **37 domain event qua outbox** và một Published Language —
> không context nào import context nào. **Bảy aggregate root**, hai domain
> service, **hai anti-corruption layer** (một cho API shop, một cho mô hình
> ngôn ngữ đọc trang web), auth bằng bearer token với port ở tầng biên chứ
> không ở domain, và một **read model dựng chỉ bằng event** cho màn hình khách.
> **304 test**, trong đó 272 chạy dưới 2 giây không cần Docker vì domain không
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
