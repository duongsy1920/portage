# UI-GUIDE.md — bốn màn hình, và bấm cái gì trên từng cái

> Repo có **bốn** trang. Hai trang để **làm việc**, hai trang để **hiểu hệ thống**. Lẫn
> lộn giữa chúng là lý do bản đầu tiên khó dùng, nên đây là chỗ nói rõ trang nào cho ai.

| Trang | Cho ai | Trả lời câu gì |
|---|---|---|
| `/ui/customer.html` | khách | *"gửi link món tôi muốn, giá bao nhiêu, hàng đâu rồi"* |
| `/ui/staff.html` | nhân viên | *"còn việc gì phải làm, và bước tiếp theo là gì"* |
| `/ui/flow.html` | người học | *"hệ thống chạy ra sao"* — mô phỏng, **không gọi API** |
| `/ui/console.html` | lập trình viên | *"request nào, mã lỗi gì"* — bảng kiểm API |

`/ui/` là trang chọn, và `http://localhost:8080/` cũng nhảy về đó.

---

## 0. Chạy lên

```bash
go run ./cmd/api -web ./web          # in-memory: không cần Postgres, mất dữ liệu khi tắt
```

Mở `http://localhost:8080/ui/`. Chế độ này để sẵn hai chìa khoá `dev-operator` và
`dev-customer`, và có một relay chạy trong process mỗi 200 ms.

Muốn dữ liệu còn sau khi tắt:

```bash
docker compose up -d
go run ./cmd/api -dsn "postgres://portage:portage@localhost:5432/portage?sslmode=disable" \
  -web ./web -bootstrap-operator-token chia-dau-tien
go run ./cmd/worker -dsn "postgres://portage:portage@localhost:5432/portage?sslmode=disable"
```

Ở chế độ `-dsn` **phải** chạy `cmd/worker`, vì relay nằm ở đó. Không có worker thì bảng đọc
không bao giờ đầy, và mọi màn hình sẽ trống dù request đã thành công.

### Trước khi khách dán được link: phải có shop

Khách chọn shop từ một danh sách, không tự gõ. Danh sách đó rỗng cho tới khi nhân viên thêm
shop, nên bước đầu tiên trên một hệ thống mới là:

```bash
curl -X POST localhost:8080/merchants \
  -H 'Authorization: Bearer dev-operator' -H 'Content-Type: application/json' \
  -d '{"name":"www.nike.com","site":"www.nike.com","currency":"USD","sourcing":["operator","customer"]}'
```

`sourcing` phải có `customer`, không thì khách không gửi được link cho shop đó. Trang khách
lọc danh sách theo đúng điều này, **và server cũng chặn**: gửi thẳng bằng `curl` cho một shop
chỉ nhận `operator` sẽ nhận 409 `sourcing_not_allowed`. Một danh sách đã lọc không phải là
kiểm soát — `curl` không đọc màn hình.

---

## 1. Chìa khoá

Cuối mỗi trang có một dòng **Chìa khoá đang dùng: …**, bấm *Đổi chìa* thì nó mở ra ô nhập.
Nó thay cho đăng nhập: mọi request gắn kèm nó, nên
"hàng của tôi" và "đơn của tôi" không bao giờ trả về của người khác. Nó cũng là cách hệ
thống ghi lại **ai** đã xác nhận sản phẩm và **ai** đã nhận tiền.

| Chìa | Dùng ở đâu | Sai thì |
|---|---|---|
| `dev-operator` | trang nhân viên | 401 nếu trống, 403 nếu dùng chìa khách |
| `dev-customer` | trang khách | 403 khi vào route của nhân viên |

Ở `-dsn` thì không có chìa sẵn: chìa đầu tiên do `-bootstrap-operator-token` cắt, và nó chỉ
nổ khi bảng chìa còn rỗng. Chìa cho khách thì operator phát qua `POST /tokens`.

---

## 2. Trang khách — ba việc

Đầu trang là tiêu đề và ba số đếm sống (việc chờ bạn, món đang xử lý, đơn đang chạy). Từ 1024 px
trở lên trang chia hai cột: **trái** là ô gửi link, đứng yên khi cuộn; **phải** là việc, hàng và
đơn. Trên điện thoại thì một cột. Trên cùng là **Việc đang chờ bạn** (chỉ hiện khi có): món
nhân viên đã xử lý xong, báo giá chưa đồng ý, cọc hoặc phần còn lại chưa chuyển. Mỗi dòng có
nút *Xem* cuộn mượt tới đúng món đó, làm nó sáng lên một nhịp, và đặt con trỏ vào ô đầu tiên,
nên đi bằng bàn phím cũng tới.
Số trên huy hiệu cạnh chữ "Portage" là số dòng ở đây.

### (a) Gửi link

Điền năm ô. Bốn ô đầu là shop, ngành hàng, link, tên, giá. Ô thứ năm là **size hoặc màu bạn
muốn**, và nó quan trọng hơn trông có vẻ: nhân viên đọc đúng dòng đó để biết mua cái nào.
Bỏ trống là họ phải hỏi lại.

Viết đúng chữ trên trang shop, **kể cả chữ cái của hệ size**:

```
M   nam            W   nữ            Y   thiếu niên        C   trẻ nhỏ
S/M/L  áo          256GB  dung lượng
```

Cùng số 1 mà `1Y` và `1C` là hai đôi khác nhau. Trang nào ghi hai hệ cùng lúc, ví dụ
`M 8 / W 9.5`, thì copy nguyên cả dòng. Đây là lý do ô này là **chữ tự do** chứ không phải
danh sách chọn: không bảng size nào phủ hết được.

Tiền tệ **khoá theo shop**, không chọn được. Shop bán bằng USD thì giá phải là USD.

### (b) Chờ, rồi xin báo giá

Ngay sau khi gửi, món hiện trong **Hàng của tôi** với trạng thái *"nhân viên đang nhập
size"*. Trang tự cập nhật mỗi 4 giây, không cần tải lại, và sẽ bật thông báo khi có tiến
triển. Bốn trạng thái bạn sẽ thấy lần lượt:

```
nhân viên đang nhập size  →  đang kiểm đúng sản phẩm  →  đang cân và đo  →  chờ đăng bán
```

Khi thành **sẵn sàng báo giá** thì có nút *Xin báo giá*, và ô chọn size đã chọn sẵn đúng cái
bạn yêu cầu.

### (c) Đọc báo giá rồi đồng ý

Báo giá tách thành từng dòng, mỗi dòng nói rõ nó ở đâu ra:

```
Tiền hàng trên web          giá bạn thấy trên trang shop
Thuế bán hàng bên Mỹ        shop Mỹ tính thêm khi thanh toán, mình trả hộ
Cước bay, tính trên 2500 g  KHÔNG theo cân thật: lấy số lớn hơn giữa cân thật và cân
                            quy đổi từ kích thước hộp, rồi làm tròn lên bước 500 g
Cộng lại bên Mỹ             ba dòng trên, vẫn bằng tiền của shop
Quy ra tiền Việt            theo tỷ giá đã KHOÁ vào báo giá này
Phí dịch vụ mua hộ          phần của bên mình
Tổng bạn trả                đã gồm hết, không phát sinh sau
Cọc trước                   một nửa tổng
```

Bấm *Đồng ý và đặt hàng* là tạo đơn. Báo giá giữ **48 giờ**, hết thì xin lại cái mới.

Món đang được báo giá nổi lên thành một **phiếu** (nền trắng, có viền); các món khác chỉ là
dòng trong danh sách.

Đơn hiện ở **Đơn của tôi**, mỗi đơn một **dải hành trình** kiểu nhãn vận đơn:

```
┌──────────────┬──────────────┬─────────┬─────────────┬─────────┬─────────┐
│ Báo giá      │ Đã cọc       │ Đã mua  │ Kho Denver  │ Đã bay  │ Đã giao │
│ 5.393.720 ₫  │ 2.696.860 ₫  │         │             │         │         │
│ đặt 23/09    │ cần chuyển   │         │             │         │         │
└──────────────┴──────────────┴─────────┴─────────────┴─────────┴─────────┘
  ━━━━━━━━━━━━━━━━━━━✈- - - - - - - - - - - - - - - - - - - - - - - - - - -   (đường bay)
Chờ bạn chuyển cọc. Việc của bạn: chuyển cọc cho nhân viên. Bên mình ứng trước…
```

Ô đổ xanh đặc là bước đơn đang ở; ô viền **sọc đỏ–xanh** là bước đang chờ chính bạn. Dưới nhãn là
đường bay: một chiếc máy bay đứng ở bước hiện tại, và **bay** sang ô mới khi đơn chuyển bước
(trang tự hỏi lại mỗi 4 giây nên không cần tải lại). Dải gộp lại hai câu mà trước đây nằm ở hai cột: *đơn đang ở đâu* (tiền và cam
kết) và *kiện hàng* (cái hộp nằm đâu). Câu dưới dải nói cả hai.

Ô nào đã qua mà **chỉ ghi "xong"** là vì màn hình khách không được đọc số của bước đó: số tiền
mua thật, cân ở kho và phần cước là số nội bộ, chỉ route của nhân viên trả. Màn hình không bịa số
để lấp. Trên màn nhân viên cùng ô đó có số thật.
Đơn huỷ hoặc shop không bán được thì dải dừng ở bước cuối đã tới, ô đó có dấu ✕, và câu dưới
dải nói tiền được hoàn thế nào. Màn hình hẹp hơn 560 px thì dải dựng dọc và ẩn đường bay.

Nếu một nhân viên đặt đơn này hộ bạn (P10 — ví dụ bạn gọi điện đặt qua nhân viên thay vì tự
bấm), dòng "nhân viên đặt hộ bạn" hiện ngay dưới tên sản phẩm.

Cọc và tiền còn lại thì chuyển cho nhân viên; **chưa có cổng thanh toán**, nên họ xác nhận
tay trong hệ thống.

---

## 3. Trang nhân viên — năm hàng chờ

Làm cho màn rộng trước. Đầu trang là tiêu đề *Bàn làm việc* và số việc đang chờ. Từ 1024 px trở
lên là hai cột: **trái** là ba hàng đợi, mỗi hàng có số đếm, số chuyển xanh khi có việc
(*Món chờ xử lý* · *Việc đi mua* · *Kho Denver* · *Tiền chờ thu* · *Chờ giao*); **phải** là phiếu đang mở. Hẹp hơn
thì xếp chồng, bấm một dòng là cuộn xuống phiếu.

Bạn không phải đi tìm việc kế tiếp: khi một việc rời hàng đợi (đăng bán xong, thu tiền xong),
phiếu tự mở việc đầu tiên còn lại, theo đúng thứ tự công việc chạy: làm cho bán được → đi mua →
cân và gom lô → thu tiền → giao. Riêng sau *Đã mua xong* thì phiếu **ở lại** để bạn thấy dải hành trình của đơn chuyển
sang *Kho Denver*, kèm nút *Mở việc kế tiếp*.

### (a) Món chờ xử lý

Mỗi món khách gửi là một phiếu, có **checklist bốn bước đánh số**. Chỉ một bước sáng lên, ba
bước còn lại mờ. Bước đã xong **hiện thứ nó đã ghi** thay vì thành nút xám, vì "Đã có
1Y · black" mới là bằng chứng bước đó xảy ra.

| # | Bước | Bạn làm gì | Vì sao có bước này |
|---|---|---|---|
| 1 | Khách mua size nào | ô Size **đã điền sẵn** chữ khách viết; mở trang shop kiểm rồi sửa theo chữ của shop nếu khác | không có size thì không đặt được đơn, và mua nhầm hệ size là đổi cả đơn |
| 2 | Xác nhận đúng sản phẩm | bấm một nút | chữ ký của bạn: khách dán link thì chưa ai kiểm |
| 3 | Cân và đo thật | bốn số, **điền sẵn theo hộp mẫu của ngành hàng** | trang shop không công bố cân đóng gói; cước tính theo số này |
| 4 | Đăng bán | bấm một nút | từ đây khách mới xin được báo giá |

Ba điều đáng biết ở đây:

- Số cân điền sẵn là **hộp mẫu của ngành hàng**, chỉ để đỡ gõ. Phải sửa thành số cân thật của
  hộp trước mặt. Ô đó có dòng tính luôn cân quy đổi để bạn thấy số nào sẽ được dùng.
- Yêu cầu gốc của khách **vẫn hiện** sau khi bạn tạo size. Nếu lệch nhau, ví dụ khách xin
  `US 9` mà shop chỉ có `US 9.5`, thì đó là chuyện phải nói với khách, không phải chuyện tự
  quyết.
- Khách **không ghi** size thì thẻ nói rõ, và lời khuyên là hỏi lại chứ đừng đoán.

### (b) Tiền chờ thu

Phiếu có dải hành trình của đơn và đúng số cần thu. Thu cọc trước khi đi mua, thu phần còn lại
khi hàng đã bay; phần còn lại là tổng trừ cọc, đúng phép tính `CustomerOrder.Balance()`. **Chỉ bấm khi tiền đã thực sự
vào tài khoản**, và phải đúng số: thiếu hay thừa một đồng hệ thống đều không nhận, vì cọc
thiếu không phải một cam kết nhỏ hơn và cọc thừa là một khoản phải hoàn mà không ai xin.

Đơn nào một nhân viên đặt hộ (không phải khách tự bấm) có dòng "đơn đặt hộ (nhân viên)" dưới
tên sản phẩm — cùng thông tin `PlacedBy` với màn hình khách, không giấu ai đặt.

### (c) Việc đi mua

Việc chỉ mở sau khi cọc vào. Mỗi phiếu có đủ thứ để cầm đi mua: tên, size, mã của shop, và
link để mở, cùng dải hành trình của đơn (ô *Đã mua* có sọc: đó là việc của bạn). Nhập lại hai thứ sau khi mua xong:

| Ô | Vì sao cần |
|---|---|
| Mã đơn ở shop | để đối chiếu khi kiện về kho, và để khiếu nại nếu shop giao sai |
| Đã trả thật | số **thật** đã trả, không phải số đã báo khách; hệ thống cần cả hai để biết đơn lời hay lỗ |

Việc đã mua rời hàng đợi (API chỉ trả việc đang mở), nhưng trong phiên đang mở nó vẫn nằm dưới
nhãn riêng *Vừa mua xong trong phiên này*, nên số đếm ở đầu hàng luôn khớp số dòng đang chờ.
Phiếu vừa mua đọc lại việc đó một lần, nên ô *Đã mua* trên dải có ngày mua và số tiền thật.

### (d) Kho Denver

Kiện xuất hiện ở đây ngay khi một món mua xong, với trạng thái *chờ shop giao tới kho*.

1. **Hộp tới kho:** mở kiện, cân và đo thật (bốn số), bấm *Lưu số cân kiện*. Đây là số tính cước
   **thật**, so với số lúc báo giá để biết lời lỗ.
2. **Xếp lô:** bấm *Xếp vào lô đang gom*, hoặc *Mở lô mới và xếp vào* nếu chưa có lô nào đang gom.
   Phiếu chuyển sang lô.
3. **Dán kín lô** khi đủ chuyến. Dán rồi thì không thêm kiện được nữa.
4. **Xác nhận lô đã bay:** nhập tổng cước trên hoá đơn của hãng bay (USD). Hệ thống chia số đó cho
   từng kiện theo cân tính cước, và phiếu hiện ngay phần của mỗi đơn. Lô vừa bay ở lại dưới nhãn
   *Lô vừa bay trong phiên này*.

Lô không được đo cả khối: nó chỉ nhận hoá đơn của hãng bay rồi chia (quyết định 08/09 trong
`CLAUDE.md`). Với số của smoke (cân 1250 g, hộp 340×230×130, cước lô 27.50 USD) bảng đối chiếu ra
đúng `variance -2.50 USD`.

### (e) Tiền chờ thu → Chờ giao

Khi lô đã bay, đơn hiện lại ở *Tiền chờ thu* với **phần còn lại**. Thu xong thì nó sang *Chờ giao*;
bấm *Đã giao tận tay khách* khi kiện tới tay khách là đơn khép lại, và dải của cả hai màn hình
đầy đủ sáu ô.

---

## 4. Mười nhánh rẽ nên thử tay

| Thử | Kết quả đúng |
|---|---|
| Xoá chìa khoá rồi bấm bất cứ nút nào | 401, và trang nói rõ chưa có chìa |
| Dán chìa khách vào trang nhân viên | 403 |
| Gửi link cho một shop chỉ nhận `operator` (bằng `curl`) | 409 `sourcing_not_allowed` — chìa đúng, shop không nhận nguồn đó |
| Gửi link với giá `115.00` mà chọn VND | *"Số tiền không đúng dạng, tiền Việt không có phần thập phân"* |
| Gửi **cùng một link** hai lần | vẫn nhận, nhưng cái thứ hai bị cắm cờ nghi trùng và sẽ bị chặn ở bước Đăng bán |
| Bấm Đăng bán khi chưa có size | *"Phải có ít nhất một size trước khi đăng bán"* |
| Nhập lại **đúng size và màu** đã có | *"Size và màu này đã có rồi"* |
| Thêm một size không tên khi đã có size khác | *"Sản phẩm đã có size khác, nên size mới phải có tên"* |
| Thu cọc lệch một đồng | *"Phải đúng số tiền, không thiếu không thừa"* |
| Xin báo giá ngay sau khi Đăng bán | có thể thấy *"chưa đăng bán xong, vài giây nữa thử lại"* — đó là relay chưa chạy, không phải lỗi |
| Đồng ý báo giá rồi đặt luôn | trang tự thử lại vài nhịp nếu relay chưa giao `quote_accepted` |

Ba dòng cuối là **nhất quán sau cùng** hiện ra ở UI. Chúng không phải bug, và cách trang xử
lý là tự thử lại chứ không báo đỏ.

---

## 5. Hỏng thì xem bảng này

| Triệu chứng | Nguyên nhân thường gặp |
|---|---|
| Trang trắng, console báo module lỗi | mở bằng `file://`. Phải qua `http://…/ui/` |
| Danh sách shop rỗng | chưa thêm shop, hoặc shop không có `customer` trong `sourcing` |
| Mọi nút trả 401 | chìa sai, hoặc `-dsn` mà chưa cắt chìa nào |
| Gửi link xong nhân viên không thấy | ở `-dsn` mà chưa chạy `cmd/worker`; bảng đọc chỉ đầy sau relay |
| `merchant_not_found` với id vừa dùng | api vừa khởi động lại; chế độ in-memory mất hết dữ liệu |
| Đơn vẫn "chờ cọc" sau khi thu | thu ở trang nhân viên, khách chỉ xem |
| Số cân điền sẵn không phải 1250 | nó là hộp mẫu của **ngành hàng bạn chọn**, không phải của món này |
| Chữ hiện bằng font hệ thống, dấu tiếng Việt lệch | thiếu file trong `web/vendor/fonts/`, hoặc `app/fonts.css` trỏ sai; font không bao giờ tải từ mạng |
| Trên **trang khách**, các ô *Đã mua*, *Kho Denver*, *Đã bay* chỉ ghi "xong", không có số | đúng như thiết kế: số cân, số tiền mua thật, phần cước là của nội bộ, chỉ route của nhân viên trả; màn hình khách không bịa |
| Trên **trang nhân viên**, các ô đó trống sau khi tải lại trang | ngày mua, cân và cước chỉ được đọc lại trong phiên đang làm; `GET /purchase-tasks` và `GET /parcels` chỉ trả việc chưa xong |
| Mua xong mà việc biến khỏi hàng *Việc đi mua* sau khi tải lại trang | cùng lý do: trang chỉ nhớ việc đã mua **trong phiên** |

---

## 6. Source FE

```
web/
├── index.html            trang chọn màn hình
├── customer.html         trang khách        ─┐
├── staff.html            trang nhân viên    ├─ React, cùng dùng app/*.js
├── console.html          bảng kiểm API      ─┘ (vanilla, không React)
├── flow.html             mô phỏng, một file, không gọi API
├── vendor/               React + ReactDOM + htm, kèm trong repo
│   └── fonts/            Be Vietnam Pro + chữ số IBM Plex Mono (woff2, kèm OFL)
└── app/
    ├── portage.js        chỗ DUY NHẤT gọi API: gắn chìa, dịch lỗi thành mã
    ├── words.js          từ điển: mọi mã của máy → chữ người đọc được,
    │                     và journeyOf(): đơn → sáu ô của dải hành trình
    ├── ui.js             component dùng chung: Top, Section, Sheet, Problem, Field, Select,
    │                     Steps, Journey, Keys, Toasts + usePoll, useAction, useToasts
    ├── icons.js          icon SVG vẽ tay, không emoji
    ├── fonts.css         @font-face, dùng chung cho screens.css và console.css
    ├── screens.css       thiết kế của ba trang: chọn, khách, nhân viên
    ├── customer.js       trang khách
    ├── staff.js          trang nhân viên
    └── api.js, main.js, steps.js, console.css   bảng kiểm API cũ
```

Màu, chữ, bố cục và các mục đã bỏ từ gợi ý của công cụ thiết kế nằm ở
`design-system/portage/MASTER.md`. Ý chính: **phong bì par avion**. Nền giấy xanh-xám, mực xanh
đen; sọc xanh–đỏ chỉ ở chân thanh trên cùng và ở chỗ "việc đang chờ bạn". Đỏ **không bao giờ**
là lỗi (đã đo: ở dark, đỏ sọc và đỏ lỗi chỉ cách nhau 1.07:1), nên lỗi luôn có chữ, nền và icon.
Có cả light lẫn dark, theo cài đặt của hệ điều hành.

Hai quy tắc chạy suốt hai trang làm việc, và chúng là lý do bản đầu tiên bị viết lại:

1. **Không chữ nào của máy ra tới màn hình.** `issued`, `open`, `none`, `branded` có nghĩa
   với người viết API và vô nghĩa với người dùng. Tất cả đi qua `words.js`.
2. **Mỗi con số nói được nó ở đâu ra.** Nhìn một số mà phải đi hỏi thì màn hình sai, không
   phải người dùng sai.

**React nhưng không có bước build.** Ba file trong `web/vendor/` là React 18, ReactDOM và
`htm` (viết JSX bằng template literal). `cmd/api -web` serve thẳng thư mục này, nên không
cần `npm install`, không cần mạng, và sửa file là F5 thấy ngay. Đổi lại là không có
TypeScript và không dùng được hệ sinh thái npm — nếu sau này cần thì chuyển sang Vite,
backend không phải sửa gì.

**Thông báo giữa hai vai** không dùng websocket. Mỗi trang hỏi hàng chờ của mình mỗi 4 giây
rồi so với lần trước; khác thì bật toast và tăng số trên badge. Cách này thật thà với việc
bảng đọc vốn đã là nhất quán sau cùng: một cú push cũng sẽ tới trước khi projection kịp ghi.
