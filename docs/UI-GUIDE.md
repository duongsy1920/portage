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

Cuối mỗi trang có một thẻ **Chìa khoá**. Nó thay cho đăng nhập: mọi request gắn kèm nó, nên
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

Đơn hiện ở **Đơn của tôi**, hai cột giữa trả lời hai câu khác nhau: *đơn đang ở đâu* là
chuyện tiền và cam kết, *kiện hàng* là chuyện cái hộp đang nằm đâu. Chúng đổi vào những lúc
khác nhau nên tách hai cột.

Cọc và tiền còn lại thì chuyển cho nhân viên; **chưa có cổng thanh toán**, nên họ xác nhận
tay trong hệ thống.

---

## 3. Trang nhân viên — ba hàng chờ

### (a) Việc cần làm

Mỗi món khách gửi là một thẻ, có **checklist bốn bước đánh số**. Chỉ một bước sáng lên, ba
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

### (b) Đơn chờ thu tiền

Thu cọc trước khi đi mua, thu phần còn lại khi hàng đã bay. **Chỉ bấm khi tiền đã thực sự
vào tài khoản**, và phải đúng số: thiếu hay thừa một đồng hệ thống đều không nhận, vì cọc
thiếu không phải một cam kết nhỏ hơn và cọc thừa là một khoản phải hoàn mà không ai xin.

### (c) Việc đi mua

Việc chỉ mở sau khi cọc vào. Mỗi phiếu có đủ thứ để cầm đi mua: tên, size, mã của shop, và
link để mở. Nhập lại hai thứ sau khi mua xong:

| Ô | Vì sao cần |
|---|---|
| Mã đơn ở shop | để đối chiếu khi kiện về kho, và để khiếu nại nếu shop giao sai |
| Đã trả thật | số **thật** đã trả, không phải số đã báo khách; hệ thống cần cả hai để biết đơn lời hay lỗ |

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
└── app/
    ├── portage.js        chỗ DUY NHẤT gọi API: gắn chìa, dịch lỗi thành mã
    ├── words.js          từ điển: mọi mã của máy → chữ người đọc được
    ├── ui.js             component dùng chung: Top, Card, Field, Steps, Toasts, usePoll
    ├── screens.css       thiết kế của hai trang làm việc
    ├── customer.js       trang khách
    ├── staff.js          trang nhân viên
    └── api.js, main.js, steps.js, console.css   bảng kiểm API cũ
```

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
