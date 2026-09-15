# HOC.md — lộ trình học Go + DDD bằng chính project này

File này không dạy lại Go hay DDD từ sách. Nó nói **đọc gì, theo thứ tự nào, và làm gì để
biết mình đã thật sự hiểu**.

Người đọc được giả định là một lập trình viên PHP/Symfony có kinh nghiệm, **chưa viết Go và
chưa làm DDD**. Nên mỗi buổi đều có một cột *"ở Symfony cái này là gì"*, và mọi ví dụ đều là
code đang chạy trong repo, không phải code minh hoạ.

**Quy tắc học duy nhất trong file này:**

> Mỗi buổi, đọc xong phải **sửa được một dòng và làm test đỏ**. Đọc mà không phá được cái gì
> thì chưa phải hiểu, mới là quen mặt.

---

## Trước buổi 0 — mười lăm phút: 12 từ và 1 hình

Tài liệu của repo dùng khoảng mười hai từ chuyên môn lặp đi lặp lại. Đọc bảng này một lần
trước, để khi gặp chúng ở các buổi sau bạn không phải dừng lại tra.

| Từ | Nghĩa trong repo này | Bạn sẽ gặp nó ở |
|---|---|---|
| **Bounded context** | một vùng nghiệp vụ có từ vựng riêng. Ở đây là 5 thư mục con của `internal/domain/`: `catalog`, `pricing`, `ordering`, `procurement`, `logistics` | buổi 5 |
| **Aggregate** (và **aggregate root**) | một cụm dữ liệu phải luôn đúng cùng nhau, với **một** cửa duy nhất để sửa. `Product` và các `Variant` của nó là một aggregate; `Product` là root | buổi 3 |
| **Value Object** | một giá trị không có danh tính: hai tờ 50 000 ₫ là như nhau. `Money`, `Weight` | buổi 2 |
| **Entity** | một thứ có danh tính (ID) và vòng đời: hai khách cùng tên vẫn là hai người. `Merchant`, `CustomerOrder` | buổi 2 |
| **Invariant** | một luật luôn phải đúng, không có ngoại lệ. *"Cân nặng không âm"*. Constructor từ chối tạo ra thứ vi phạm | buổi 2 |
| **Domain event** | một việc **đã xảy ra**, đặt tên ở thì quá khứ: `deposit_paid`, `product_published` | buổi 3, 5 |
| **Outbox** | bảng ghi event, ghi **cùng transaction** với dữ liệu, để không bao giờ có chuyện "lưu xong mà event mất" | buổi 5 |
| **Worker / relay** | một tiến trình **riêng** đọc bảng outbox và đưa từng event tới các vùng đang nghe. Nó không phải một hàm được gọi; nó chạy sau | buổi 0, 5 |
| **Projection** (bảng chiếu) | bản sao **mỏng** mà một vùng giữ về dữ liệu của vùng khác, dựng từ event. `pricing` giữ `Listing` là bản chiếu của `Product` | buổi 5 |
| **Read model** (bảng đọc) | một bảng phẳng cho **màn hình**, dựng từ event của nhiều vùng. `order_summaries` | buổi 7 |
| **Port / Adapter** | port là một interface nói *"tôi cần gì"* (`Clock`, `MerchantRepository`). Adapter là cài đặt *"bằng gì"*: RAM hay Postgres | buổi 4 |
| **Idempotent** · **eventual consistency** | nhận cùng một event hai lần thì kết quả như một lần · hai vùng khớp nhau **sau một lúc**, không phải tức thì | buổi 5 |

Và một hình. Đây là toàn bộ hệ thống, nhìn từ trên cao:

```
   khách dán link ─▶ CATALOG ─product_published─▶ PRICING ─quote_accepted─▶ ORDERING
                                                     ▲                          │ deposit_paid
                                                     │                          ▼
                                                     │ (giá thật)         PROCUREMENT ─▶ người đi mua
                                                     │◀───────────────────────┘ purchase_confirmed
                                                     │ (cước thật)                │
                                                     │◀──── LOGISTICS ◀───────────┘  kho cân, gom lô, ship
                                                     ▼
                                              variance = đã báo − thật

   Mọi mũi tên là một EVENT đi qua bảng outbox, do WORKER chuyển.
   Không vùng nào gọi hàm của vùng khác. Không vùng nào đọc bảng của vùng khác.
```

Nếu bạn nắm được hình này, mọi thứ đọc sau đó đều có chỗ để gắn vào.

---

## Bảy chỗ đọc, bảy câu hỏi khác nhau

| File | Trả lời | Đọc khi nào |
|---|---|---|
| [UI-GUIDE.md](UI-GUIDE.md) | *"bấm cái gì để thấy nó chạy"* | buổi 0, trước khi đọc dòng code nào |
| [GO-CHO-PHP.md](GO-CHO-PHP.md) | *"cú pháp này ở PHP là gì"* | buổi 1 đọc §1–§9 **theo thứ tự**; sau đó tra khi vướng |
| [SETUP.md](SETUP.md) | *"máy tôi chạy nó thế nào, code nằm ở đâu, quy ước gì"* | buổi 1, rồi tra lại mỗi ngày |
| [WALKTHROUGH.md](WALKTHROUGH.md) | *"code chạy ra sao"*: file nào, dòng nào | xương sống của các buổi 4–9 |
| [DDD.md](DDD.md) | *"vì sao lại thiết kế thế"*: khái niệm | đọc **sau** khi đã thấy code của khái niệm đó |
| [FLOW-ORDER.md](FLOW-ORDER.md) | *"một đơn hàng đi qua những đâu"* | buổi 5, và trước khi phỏng vấn |
| [SETUP.md §9](SETUP.md) | *"vì sao code lại trông như thế này"*: nhật ký 20 đợt review | khi vướng một quyết định trông vô lý |

Thứ tự sai hay gặp nhất là đọc DDD.md trước. Khái niệm không có code đi kèm thì đọc xong quên
ngay. Ở đây **luôn xem code trước, đọc lý do sau**.

---

## Buổi 0 — Bấm trước, đọc sau

Nửa tiếng này tiết kiệm cho bạn vài buổi. Đọc code mà chưa có bản đồ trong đầu thì mỗi file
là một hòn đảo. Bấm qua giao diện trước thì mọi thứ đọc sau đó có chỗ để gắn vào.

**Bấm, theo đúng thứ tự này:**

1. **`web/flow.html`**: mở bằng double-click, không cần build, **không gọi API**. Nó mô
   phỏng 26 bước của một đơn. Mỗi bước hiện màn hình người dùng thấy, request gửi đi, event
   ghi vào outbox, và ai nghe event đó. Bấm hỏng cũng không sao.
2. **Hai màn hình thật**: chạy `go run ./cmd/api -web ./web` rồi mở `http://localhost:8080/ui/`.
   Mở trang **khách** ở một tab và trang **nhân viên** ở tab khác. Gửi một link từ trang
   khách. Bạn sẽ thấy việc xuất hiện bên trang nhân viên, checklist bốn bước, và thông báo
   bắn qua lại. Đó là vòng đời một đơn nhìn từ phía người dùng.
3. **Bảng kiểm API**: cùng địa chỉ, trang `console.html`. Tab **Runner** chạy 26 bước bằng
   request thật, cột **Nhật ký gọi** cho thấy từng request và mã lỗi. Đây là công cụ cho lập
   trình viên, không phải giao diện cho người dùng. Phân biệt được hai thứ đó cũng là một
   bài học.
4. **[UI-GUIDE.md](UI-GUIDE.md)**: bốn màn hình là gì, bấm gì trên từng cái, 10 nhánh rẽ nên
   thử bằng tay, và bảng triệu chứng suy ra nguyên nhân khi có gì không chạy.

**Ba điều chỉ bấm mới thấy, đọc code rất khó nhận ra:**

1. **Worker là một bước riêng.** Nó không phải hàm được gọi; nó là một tiến trình khác, chạy
   sau. Trong mô phỏng, sáu lượt relay của một đơn đưa đi đúng 5 / 2 / 2 / 2 / 6 / 3 event,
   khớp log thật.
2. **Bảng chiếu trống cho tới khi relay chạy.** Vì thế `POST /quotes` ngay sau khi đăng bán
   trả 404 `listing_not_found`. Đó là thiết kế, không phải lỗi. Trang khách xử lý bằng cách
   tự thử lại.
3. **Một event có nhiều người nghe.** `purchase_confirmed` có bốn: ordering, logistics,
   pricing, reporting.

**Biết là hiểu khi:** chỉ được trên màn hình chỗ nào là *"đã ghi nhưng chưa ai biết"* và chỗ
nào là *"đã có người nghe rồi"*.

---

## Buổi 1 — Go không có gì lạ, chỉ khác chỗ để lỗi

**Đọc trước khi mở file `.go` nào:** [GO-CHO-PHP.md](GO-CHO-PHP.md) §1 tới §9, **theo thứ
tự**, khoảng một tiếng rưỡi. §9 (vòng đời chương trình Go so với PHP-FPM) là mục quan trọng
nhất của cả file: PHP dựng lại thế giới mỗi request, Go dựng một lần rồi phục vụ mọi request
trên đó. Hiểu điều này thì `main()`, `context.Context`, goroutine, và data race đều tự sáng.

**Rồi mở code:** `internal/domain/shared/money.go`, `decimal.go`, `id.go`. Ba file này có
nhiều comment `// [PHP]` nhất trong repo. Đọc chúng ngay tại chỗ cú pháp xuất hiện.

Ba thứ khác PHP nhiều nhất, và cả ba đều nằm trong ba file trên:

```
   PHP                              Go
   ───                              ──
   throw new Exception              return err        lỗi là GIÁ TRỊ trả về, không phải nhánh ngầm
   $obj->method()                   func (m Money) …  method gắn vào kiểu, viết tách rời struct
   float cho tiền                   int64 minor units 0.1 + 0.2 != 0.3, và tiền thì không được sai
```

**Làm:**

1. `go test ./internal/domain/shared/`. Xem số test và thời gian chạy.
2. Mở `money.go`, tìm `divRoundHalfUp` trong `decimal.go`, đổi làm tròn thành làm tròn
   xuống. Chạy test lại. **Đếm xem mấy test đỏ**, rồi khôi phục.
3. Viết một hàm test nhỏ trong `money_test.go` cộng hai `Money` khác tiền tệ. Xem nó bị từ
   chối thế nào và ở đâu.

**Ở Symfony:** `Money` ở đây là `Brick\Money` viết tay và không biết Doctrine tồn tại. Không
có `#[ORM\Embeddable]` trên struct; phần map xuống cột nằm ở `adapter/postgres`.

**Biết là hiểu khi:** trả lời được *"vì sao `Money` không có setter?"* mà không nhìn code.

---

## Buổi 2 — Value Object và Entity: cái gì có ID, cái gì không

**Đọc:** DDD.md §11 (Value Object), §12 (Entity), §13 (Invariant), §10 (Primitive Obsession).

**Code:** `shared/weight.go`, `catalog/freeshipping.go`, `catalog/merchant.go`.

Câu hỏi phân biệt, học một lần dùng mãi:

```
   Hai tờ 50 000 ₫ có khác nhau không?   → không  → VALUE OBJECT (Money, Weight)
   Hai khách tên Nguyễn Văn A?           → có     → ENTITY  (có ID riêng)
```

Ở Symfony bạn quen thấy `Entity` cho **mọi thứ** vì Doctrine map như thế. Ở đây `Money` không
phải entity và không bao giờ có bảng riêng. Nó là hai cột `amount_minor` và `currency` trong
bảng của thứ sở hữu nó.

**Làm:**

1. Xoá dòng kiểm `grams < 0` trong `NewWeight`. Chạy `go test ./internal/domain/...`. Đếm
   test đỏ, khôi phục.
2. Tìm trong `merchant.go` chỗ **duy nhất** một `Merchant` được tạo ra. Vì sao không có
   `merchant := &Merchant{}` ở đâu khác?

**Biết là hiểu khi:** chỉ ra được ba Value Object trong repo và nói vì sao mỗi cái không cần ID.

---

## Buổi 3 — Aggregate: ranh giới của một transaction

**Đọc:** DDD.md §14 (Aggregate), §15 (Domain Event), §29 (vòng đời một aggregate).

**Code:** `catalog/product.go` (đọc kỹ `Publish`), `catalog/events.go`.

`Product.Publish()` có **năm** điều kiện. Đọc từng cái và tự hỏi: *"cái này để ở đâu nếu viết
theo kiểu Symfony?"*. Câu trả lời thường là "trong một Service". Đó chính là chỗ luật nghiệp
vụ bị rải mỏng ra rồi mất.

**Ba câu hỏi để chọn ranh giới aggregate**, học thuộc:

```
   1. Cái gì phải ĐÚNG ngay lập tức, không được chờ?      → cùng một aggregate
   2. Cái gì chấp nhận đúng sau vài giây?                 → aggregate khác, nối bằng event
   3. Một transaction chạm mấy aggregate?                 → MỘT. Nhiều hơn là thiết kế sai.
```

**Làm:**

1. `go test ./internal/domain/catalog/`, rồi mở `product_test.go`. Xem mỗi test canh **một**
   điều kiện của `Publish`.
2. Bỏ điều kiện `listingProv.Verified()` trong `Publish`. Chạy `go test ./...` cho cả repo.
   Đếm xem có bao nhiêu test ở **bao nhiêu package** đỏ. Khôi phục.
3. Trả lời: vì sao `Variant` không phải aggregate root riêng, và vì sao không có
   `VariantRepository`?

**Biết là hiểu khi:** giải thích được vì sao `CustomerOrder` và `PurchaseTask` là hai
aggregate ở hai context, dù một cái sinh ra cái kia.

---

## Buổi 4 — Port & Adapter: cái làm project này khác một CRUD

**Đọc:** DDD.md §16 (Repository), §20 (Ports & Adapters), §21 (Dependency Rule). Rồi
WALKTHROUGH §4 và §6.

**Code:** `domain/catalog/repository.go` (port), `adapter/memory/*.go` và
`adapter/postgres/*.go` (hai adapter cho **cùng** một port), `app/ports.go`.

Đây là chỗ Symfony và project này khác nhau nhất:

```
   Symfony   Entity dính Doctrine: #[ORM\Column] nằm ngay trong class nghiệp vụ
   Portage   struct nghiệp vụ SẠCH; phần map xuống DB nằm riêng ở adapter/postgres
```

Đổi lại được gì? **278 trên 312 test chạy dưới 2 giây, không cần Docker, không cần mạng,
không cần API key.** Đó là lý do bạn sửa được code mà không sợ.

**Làm:**

1. `go test ./internal/domain/ -v`. Đây là 7 **test canh kiến trúc**. Mở `decisions_test.go`
   và đọc guard 6 (domain chỉ import stdlib và allowlist).
2. Thêm `import "database/sql"` vào một file bất kỳ trong `internal/domain/`. Chạy lại. Xem
   guard bắt bạn. Khôi phục.
3. Trả lời: `Clock` là port. Vì sao *"mấy giờ rồi"* lại là một dependency cần cắm từ ngoài?

**Ở Symfony:** `Clock` là `ClockInterface` (PSR-20). `UnitOfWork.InTx` là
`$em->wrapInTransaction(fn)`. Khác ở chỗ Go không có container: `main()` cắm tay từng cái.

---

## Buổi 5 — Một đơn hàng đi hết năm context

**Đọc:** [FLOW-ORDER.md](FLOW-ORDER.md) **cả file, một mạch**. Rồi DDD.md §6 (Bounded
Context), §7 (Context Map), §27 (Outbox), §25 (Eventual Consistency).

**Làm trước khi đọc code:** chạy cả vòng đời một đơn trên binary thật.

```bash
docker compose up -d
bash scripts/smoke.sh                                       # Linux
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1  # Windows
```

Nhìn 20 event chạy qua 5 context trên màn hình. Rồi mở
`internal/platform/wire/subscribe.go`. **35 dòng** đó chính là cái vừa chạy: event nào →
ai nghe.

**Ba ý phải nắm:**

```
   1. Context KHÔNG import nhau, KHÔNG đọc bảng của nhau. Chúng nói chuyện bằng event.
   2. Event ghi CÙNG transaction với dữ liệu (outbox), nên không bao giờ mất.
   3. Người nhận có thể nhận HAI LẦN, và SAI THỨ TỰ. Nên ai cũng phải idempotent.
```

**Làm:**

1. Trong `scripts/smoke.sh`, bỏ đoạn chờ sau bước publish. Chạy. Xem `listing_not_found`.
   Đó **không phải bug**: đó là eventual consistency có mã trạng thái.
2. Mở `internal/app/pricing/projector.go`. Chỉ ra chỗ nào làm cho nó gọi hai lần cũng như một lần.
3. Mở `internal/domain/pricing/listing.go`. `Product` ở đây là `shared.ID`, không phải
   `catalog.ProductID`. Vì sao pricing không được gọi tên kiểu của catalog?

**Đọc thêm ngay sau buổi này:** WALKTHROUGH §22. Đó là ví dụ cụ thể nhất của bài học trên:
một chữ của shop, chữ `size`, phải đi qua ba context mà không context nào được đọc bảng của
context kia. Buổi 9 mổ kỹ nó.

---

## Buổi 6 — Tiền: Quote vs Actual

**Đọc:** DDD.md §28, WALKTHROUGH §14 và §17.

**Code:** `pricing/calc.go` (một hàm thuần, không clock, không repository), `pricing/quote.go`,
`pricing/reconciliation.go`, `shared/allocate.go`.

Đây là phần **nghiệp vụ thật**, không phải bài tập:

```
   quoted  163.22 + 25.00 = 188.22 USD    mình ĐOÁN lúc báo giá
   actual  163.22 + 27.50 = 190.72 USD    thật: shop đúng, hãng bay đắt hơn 2.50
   ─────────────────────────────────
   variance                −2.50 USD      ← ÂM = mình chịu
```

`variance` **luôn âm** qua nhiều đơn nghĩa là bảng giá cước đang sai, không phải xui. Bảng
`reconciliations` sinh ra để trả lời đúng câu hỏi đó.

**Làm:**

1. Mở `shared/allocate_test.go`. Chia 87.50 USD cho hai thùng 2500 g và 5000 g. Thử tự tính
   bằng `freight × tỷ lệ` rồi làm tròn từng cái. Xem lệch mấy cent, và vì sao `Allocate`
   không lệch.
2. Đổi `divisor` của lane từ 5000 thành 6000 trong `internal/platform/wire/wire.go`. Chạy
   `go test ./internal/platform/wire/`. Xem số vàng đổi thế nào và test nào bắt được.

**Biết là hiểu khi:** giải thích được vì sao `Quote` phải **chụp ảnh** tỷ giá và bảng giá
lúc phát hành, thay vì trỏ tới "tỷ giá hiện tại".

---

## Buổi 7 — Ai đang gọi, việc không ai gọi, và màn hình cần bốn context

**Đọc:** WALKTHROUGH §18 (auth), §19 (sweep), §20 (read model), §21 (ACL cho AI).

Ba câu, mỗi câu đáng nhớ hơn cả đoạn code sinh ra nó:

```
   §18   "Ai đang gọi" không phải quy tắc nghiệp vụ. Nó là việc của BIÊN, cùng lý do với Clock.
         Nên auth ở platform/, không ở domain/. Domain chỉ nhận một shared.OperatorID.

   §19   Có loại việc KHÔNG AI GỌI: chỉ thời gian đẩy. Quote quá 48 giờ phải thành expired dù
         không ai bấm gì. Đây là lý do rõ nhất để Clock là port: 48 giờ trôi trong một dòng test.

   §20   Màn hình cần dữ liệu của bốn context → ĐỪNG JOIN. JOIN biến bốn bounded context thành
         một cục không tách nổi. Dựng một bảng phẳng bằng event thay vào đó.
```

**Làm:**

1. Trong `internal/app/reporting/projector.go`, xoá dòng `if s.Status == ""`. Chạy
   `go test ./internal/app/reporting/`. Đọc kỹ test đỏ: nó đang nói gì về việc "đoán" một
   trạng thái cho khách đang nhìn màn hình?
2. Chạy `cmd/api` với `-dsn` mà không đặt `OPENAI_API_KEY`. Gọi `POST /products/from-url`.
   Xem 503. Rồi tự hỏi: nếu nó rơi về bản `Fake` thì hỏng chỗ nào?
3. Gọi `GET /orders/{id}` bằng chìa của khách khác. Xem 404. Vì sao không phải 403?

---

## Buổi 8 — Đọc hết repo

WALKTHROUGH §10 liệt kê **101 file theo thứ tự**, mỗi dòng ghi rõ file → khái niệm → đọc
thêm ở đâu. Đừng đọc một mạch. Chia:

```
   file 1–35    ngày 1: một request đi hết 5 tầng, tới dòng sent_at trong Postgres
   file 36–56   ngày 2: context thứ hai, báo giá
   file 57–86   ngày 3: ba context còn lại, chia cước, đối soát
   file 87–93   cửa auth, sweep, read model, ACL cho AI           (WALKTHROUGH §18–§21)
   file 94–97   variant/size: một chữ của shop đi qua ba context   (§22)
   file 98–101  bảng đọc worklist + hai màn hình theo vai          (§23)
```

Sau đó đọc WALKTHROUGH §24: đặt hộ có ghi tên. Đây là lần đầu repo **mở lại một quyết định
đã đóng** (`POST /orders` từng chỉ dành cho khách), và cách nó được mở lại là bài học riêng:
lời hứa cũ được phát biểu lại thành một câu mới có test canh.

---

## Buổi 9 — Ba thứ chỉ học được khi CHẠY THẬT

Tám buổi trên học từ code đang có. Buổi này học từ ba lần code đang có **hoá ra sai**, và
cả ba đều chỉ lộ ra khi bấm thật hoặc chạy thật, không phải khi đọc.

### (a) Thêm một event vào hệ thống đang chạy — WALKTHROUGH §22

**Đọc:** WALKTHROUGH §22 cả tám mục, rồi CATALOG.md phần *"`size` là chữ của shop"*.

Bắt đầu từ một câu hỏi rất thường: *khách nhập size kiểu `M 8 / W 9.5` thì có đỡ được
không?* Trả lời được phần dễ ngay: `size` là chuỗi tự do. Nhưng lần theo nó thì lộ ra hai
lỗ hở lớn hơn nhiều:

```
   POST /orders      nhận BẤT KỲ uuid làm variant_id → khách trả cọc xong,
                     người đi mua nhận một việc không thể thực hiện
   màn hình đi mua   bốn cái uuid và một mã tiền tệ; không ai cầm vào shop được
```

Cùng một nguyên nhân: `size` chưa bao giờ ra khỏi catalog. Và không sửa bằng một câu `JOIN`:
guard 7 chặn việc context này **import** context kia, còn việc đọc bảng của nhau thì không
guard nào bắt được. Nó chỉ sai vì cùng một lý do. Nên catalog phải **kể**: event
`catalog.variant_added`.

Phần đáng học nhất là **hai người nghe chép khác nhau**:

```
   procurement   giữ 5 field (size, màu, mã shop…)   vì có NGƯỜI phải đọc
   ordering      giữ 2 field (variant, product)      vì luật chỉ cần "có thật? của ai?"
```

Một bảng chiếu giữ dữ liệu **không ai đọc** sẽ mốc mà không test nào bắt. Đó là lý do
ordering **cố tình** không giữ `size`.

**Làm:**

1. Mở `internal/domain/ordering/variant.go`. Thêm field `Size string`. Chạy `go test ./...`.
   Nó **vẫn xanh**, và đó chính là vấn đề. Tự trả lời: sáu tháng sau ai biết field đó còn
   đúng hay không?
2. Trong `internal/app/procurement/open_task.go`, đổi nhánh `ErrVariantNotFound` thành
   `return err`. Chạy test. Đọc kỹ test đỏ: nó đang nói gì về việc huỷ một đơn **đã trả
   cọc** vì một cuộc đua vài trăm mili giây?
3. Gọi `POST /orders` với một uuid tự bịa. Xem `409 variant_unknown`. Rồi tìm trong
   `web/app/portage.js` chỗ mã đó nằm trong danh sách `retryOn`, và tự trả lời vì sao nó
   **thử lại được** mà `variant_not_for_product` thì không.

### (b) Một giao diện không có test là một giao diện chưa ai chạy — WALKTHROUGH §23h

**Đọc:** WALKTHROUGH §23, cả chín mục.

`go test ./...` xanh, `curl` đúng, và năm lỗi vẫn còn nguyên trong hai màn hình: `style`
truyền chuỗi thay vì object, gọi component như hàm nên sai quy tắc hooks, `<label>` không nối
với `<input>`, dòng tiền `0.00` vẫn hiện vì so chuỗi với `"0"`, và ngành hàng mặc định đổi mỗi
lần tải vì adapter in-memory duyệt map Go. Cả năm chỉ lộ ra khi cho Chrome bấm hết luồng.

**Làm:**

1. Đọc §23h. Tự trả lời: trong năm lỗi đó, lỗi nào `go vet` bắt được? (Không cái nào.)
2. Mở `web/app/ui.js`, đổi một chỗ `<${Field} ...${{...}} />` thành `${Field({...})}`. Mở
   trang, xem console. Đó là React error #310, và là lý do component phải được dùng như
   component chứ không như hàm.

### (c) Chạy xanh nhiều lần không chứng minh gì về lần đầu — SETUP §9 đợt 18

**Đọc:** SETUP.md §9 đợt 18, rồi `internal/adapter/postgres/migrate.go`.

`Migrate` từng tạo bảng sổ sách `schema_migrations` bằng `CREATE TABLE IF NOT EXISTS`
**ngoài** advisory lock. Câu đó **không nguyên tử** trong PostgreSQL: hai tiến trình song
song đều thấy bảng chưa có, rồi một cái vỡ khi chèn kiểu dữ liệu của bảng, với lỗi
`duplicate key value violates unique constraint "pg_type_typname_nsp_index"`.

`cmd/api` và `cmd/worker` khởi động cùng lúc, nên đây là đường đi thật của một lần deploy
đầu tiên. Bug này sống sót qua gần 300 test và rất nhiều lần smoke xanh, vì:

```
   database đã migrate  →  CREATE TABLE IF NOT EXISTS là no-op  →  không có gì để đua
   database TRỐNG       →  hai tiến trình cùng tạo bảng         →  vỡ
```

Bài học đắt nhất của cả repo gọn trong một câu: **"chạy xanh nhiều lần" không chứng minh
được gì về lần chạy ĐẦU TIÊN.**

**Làm:**

1. Đọc `TestMigrate_survivesTwoProcessesOnAColdDatabase`. Trả lời: vì sao nó phải tự tạo
   một database **dùng một lần** thay vì dùng `portage_test`?
2. Dịch câu `CREATE TABLE` trong `migrate.go` trở lại ra ngoài transaction. Chạy riêng test
   đó vài lần. Xem một test **xác suất** trông như thế nào, và tự trả lời vì sao chiều
   "không bao giờ đỏ oan trên code đúng" mới là chiều quan trọng.

---

## Tự kiểm tra — trả lời không nhìn code

Trả lời trôi chảy 17 câu này là đủ để nói về project trong phỏng vấn.

**Go**

1. Vì sao tiền là `int64` chứ không `float64`?
2. `error` là giá trị trả về. Điều đó đổi cách viết code thế nào so với exception?
3. `interface` ở Go được khai báo ở đâu: nơi CẦN hay nơi CÀI? Vì sao?
4. Vì sao thư mục `internal/` có ý nghĩa với compiler?

**DDD**

5. Aggregate root là gì, và ranh giới của nó quyết định điều gì về transaction?
6. Value Object khác Entity ở đúng một điểm. Điểm nào?
7. Bounded Context: hai context cùng có "Product". Chúng có phải cùng một thứ?
8. Domain Event khác một message queue thường ở chỗ nào?
9. Anti-Corruption Layer bảo vệ khỏi cái gì? Kể hai cái trong repo này.
10. CQRS: vì sao read model **không có** invariant?

**Hệ thống**

11. Outbox giải quyết vấn đề gì mà "lưu xong rồi publish" không giải quyết được?
12. At-least-once nghĩa là subscriber phải có tính chất gì?
13. Vì sao khách hỏi đơn người khác trả **404** chứ không phải 403?
14. `variance` âm nghĩa là gì, và bao nhiêu đơn âm liên tiếp thì phải xem lại bảng giá?
15. Tính năng AI tắt (không có API key) thì API trả gì, và vì sao **không** được rơi về Fake?
16. Hai context nghe **cùng một event**: vì sao một bên chép 5 field mà bên kia chỉ chép 2?
    Chép thừa thì hỏng chỗ nào?
17. Hai `CREATE TABLE IF NOT EXISTS` chạy song song trong PostgreSQL thì ra gì, và vì sao lỗi
    đó **chỉ** hiện trên database trống?

---

## Nói gì khi phỏng vấn

Nói **đúng** phần đã làm. Đoạn dưới là sự thật, kiểm chứng được bằng repo:

> Một hệ thống mua hộ xuyên biên giới viết bằng Go theo DDD: **năm bounded context** (catalog,
> pricing, ordering, procurement, logistics) cộng một tầng đọc, nối nhau bằng **37 domain
> event qua outbox** và một Published Language. Không context nào import context nào, không
> context nào đọc bảng của context nào. **Bảy aggregate root**, hai domain service, **hai
> anti-corruption layer** (một cho API shop, một cho mô hình ngôn ngữ đọc trang web), auth
> bằng bearer token với port ở tầng biên chứ không ở domain, và **hai read model dựng chỉ
> bằng event** cho màn hình khách và nhân viên. **312 test**, trong đó 278 chạy dưới 2 giây
> không cần Docker vì domain không import gì ngoài stdlib, và **7 test canh kiến trúc** bằng
> `go/ast` khiến vi phạm dependency rule là build đỏ. Vòng đời một đơn chạy hết trên **cả**
> in-memory và Postgres bằng **cùng một test**, và trên **binary thật** bằng một smoke script
> trong CI.

Đừng nói: "có microservices" (không có, một binary), "dùng AI để tự động hoá mua hàng"
(không, AI chỉ đọc trang, người vẫn phải xác nhận), "production-ready" (chưa, chưa có cổng
thanh toán).

**Thứ đáng kể nhất trong project này không phải là code, mà là những chỗ đã cố tình KHÔNG
làm**, và bạn giải thích được vì sao:

- không cào trang shop (điều khoản cấm) → khách hoặc nhân viên nhập, hoặc AI đọc từ URL;
- không tự động huỷ đơn khi shop hết hàng → con người quyết định;
- không cho AI xác nhận listing → `Publish` vẫn đòi một người thật;
- không rơi về Fake khi thiếu API key → 503, vì tính năng tắt phải trông như đang tắt;
- không có `Quantity`: một đơn là một cái. Chưa có dữ liệu thật để biết đơn giản hoá này
  có sai không; bảng `reconciliations` sẽ trả lời.

---

## Sau khi học xong: ba việc tiếp theo đáng làm

1. **`config/ratecard.yaml`**: bảng giá cước đang là hằng số trong `wire.go`. Đưa ra file là
   bài tập tốt về "dữ liệu khác code" (DDD.md §23).
2. **Cổng thanh toán thật**: webhook → `PayDeposit` / `PayBalance`. Command đã có sẵn, chỉ
   thiếu adapter. Đây là bài tập Port & Adapter chuẩn nhất.
3. **Đọc lại `variance` sau 20 đơn thật**: nếu luôn âm thì `RateCard` sai. Đó là lúc project
   bắt đầu trả lời câu hỏi kinh doanh, không chỉ câu hỏi kỹ thuật.
