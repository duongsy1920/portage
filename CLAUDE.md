# CLAUDE.md — đọc file này trước khi làm gì

File này được Claude Code nạp tự động mỗi phiên mở trong `D:\portage`. Nó thay
cho việc phải kể lại bối cảnh từ đầu.

---

## Người dùng

Dương Sỹ (Tony) — PHP/Software Engineer, 6 năm backend, đang làm ở Fastboy
Marketing. **Đang học Go và DDD từ số 0**, dựa trên nền Symfony/Doctrine.

Project này vừa là **sản phẩm thật** (hệ thống mua hộ Mỹ → Việt Nam của anh ấy)
vừa là **bài học**, và sẽ đưa vào CV.

## Cách làm việc — bắt buộc

| Luật | Chi tiết |
|---|---|
| **Giải thích bằng tiếng Việt** | mọi phân tích, mọi câu trả lời trong chat, mọi file `.md` trong `docs/` |
| **Comment trong code bằng tiếng Anh** | project lên CV, người khác đọc |
| **Comment `// [PHP]`** | thêm khi cú pháp Go khác PHP rõ rệt — anh ấy sẽ tự bỏ trước khi commit |
| **KHÔNG tự commit** | để nguyên trong working tree cho anh ấy review. Chỉ commit khi được bảo thẳng |
| **Git identity** | repo này đã set `--local` là `duongsy1920@gmail.com` — **không** dùng mail công ty `sy_duong@fastboy.net` |

## Ba bài học đã trả giá — đừng lặp lại

1. **Đề xuất đổi thiết kế thì phải tự phản biện TRƯỚC.** Đã từng đề xuất thay
   `listingProv` bằng `classification` rồi tự rút lại — thu hẹp phạm vi không
   làm một cái dấu chính xác hơn, chỉ để hở ba field. Anh ấy nói thẳng: *"bạn
   cứ đề xuất mà không suy nghĩ kĩ như này thì thật là tốn thời gian ghê"*.
   → Đáng tin khi **kiểm chứng**, không đáng tin khi **phát biểu**. Viết phản
   biện mạnh nhất cho chính đề xuất của mình trước khi mở miệng.

2. **Script vá phải in SKIP từng anchor và ĐẾM LẠI sau khi chạy.** Đã có lần
   `perl -0pi` báo "PATCH 2 chỗ" mà không đổi gì cả. Từ đó: mỗi anchor in rõ
   HIT/SKIP, và đếm lại sau khi chạy rồi mới báo cáo. (SETUP.md §9 đợt 13, 14)

3. **Chạy một lần không chứng minh được gì về lần thứ hai.** `smoke.ps1` từng
   xanh chỉ vì database lúc đó tình cờ rỗng; lần chạy thứ hai 401 sạch.
   (SETUP.md §9 đợt 15)

## Trước khi bảo "xong"

Tự kiểm ba thứ: **lộ thông tin** (secret trong log/response?), **nhất quán**
(số trong docs có khớp code không?), **luồng người dùng thật** (chạy lại smoke,
không chỉ chạy test). Đừng để phần QA rơi vào tay anh ấy.

---

## Project là gì

Mua hộ xuyên biên giới: khách Việt Nam dán link một sản phẩm ở web Mỹ (Nike,
The North Face, Sony…), mình mua hộ, gom lô tại **kho Denver, Colorado**, ship
về tận nhà.

**Số nghiệp vụ đã chốt** (đừng tự đổi):

```
   thuế bán hàng Colorado   8,81 %
   tỷ giá                   × 26 000 ₫/USD
   cọc                      50 % giá trị đơn
   phí dịch vụ              max(10 %, sàn 500 000 ₫)
   quy đổi thể tích         mm³ / 5000, bước cân 500 g
   hạn báo giá              48 giờ
```

**Số vàng** — smoke script phải luôn in ra đúng những con số này:

```
quote … total 5393720 VND, deposit 2696860
variance -2.50 USD          (quoted 163.22+25.00, actual 163.22+27.50)
```

**Ràng buộc pháp lý:** điều khoản của Nike (và phần lớn shop) **cấm cào trang**
và cấm mua để bán lại qua affiliate feed. Nên hệ thống **không tự fetch trang
shop**: khách/operator nhập tay, hoặc `adapter/openai` đọc từ URL. Đây là ràng
buộc thiết kế, không phải chú thích.

## Trạng thái (08/09/2026 — P9 xong, đợt variant/size, đợt migrate race)

```
5 bounded context + 1 tầng đọc · 36 domain event · 37 route · 31 dòng Subscribe
293 test (292 PASS + 1 SKIP cố ý) · 263 test chạy < 2 giây không cần Docker
7 test canh kiến trúc bằng go/ast
```

| Context | Aggregate root |
|---|---|
| `catalog` | Merchant, Product ⊃ Variant |
| `pricing` | Quote, Reconciliation (Quote vs Actual) |
| `ordering` | CustomerOrder (cọc 50 %, điểm không thể quay đầu) |
| `procurement` | PurchaseTask + port MerchantACL |
| `logistics` | Parcel, ConsolidationBatch + FreightAllocator |
| `reporting` | *(không có domain — read model, không invariant nào)* |

**Vừa xong — variant/size (đợt 17):** `catalog.variant_added` là event thứ 36.
Trước đợt này `POST /orders` nhận **bất kỳ** uuid làm `variant_id`, và màn hình
người đi mua chỉ có bốn cái uuid. Giờ ordering giữ một bảng chiếu mỏng (variant
có thật? của sản phẩm nào?) và trả 409 `variant_unknown` / `variant_not_for_product`;
procurement giữ bảng chiếu đủ chữ rồi **chép** vào `PurchaseTask.Subject` lúc mở
việc (tên · size/màu · mã shop · link gốc). Chi tiết: `docs/SETUP.md` §9 đợt 17.

**Vừa xong — migrate race (đợt 18):** `Migrate` tạo bảng `schema_migrations`
**ngoài** advisory lock, và `CREATE TABLE IF NOT EXISTS` của PostgreSQL không
nguyên tử: hai process khởi động cùng lúc trên database trống thì một cái vỡ ở
`pg_type_typname_nsp_index`. Chỉ hiện khi database **cold**, nên Windows chưa
bao giờ thấy. Tìm ra bằng lần chạy thật đầu tiên của `scripts/smoke.sh` trên
Linux. Có test hồi quy dùng database dùng-một-lần.

**Đang làm — P10:** operator đặt hộ có ghi tên (`docs/P10-PLAN.md`). Anh ấy đã
chốt phương án này sau khi cân bốn lựa chọn; spec đã kiểm chứng với code thật,
**đọc nó trước khi sửa** — mô tả ban đầu của phương án dựa trên một tiền lệ không
tồn tại (`OnBehalfOf` là kiểm quyền sở hữu, không phải đặt hộ).

**Còn nợ:** `config/ratecard.yaml` (bảng giá đang là hằng trong `wire.go`) ·
một `adapter/merchant/<shop>.go` thật (chưa shop nào cho API) · cổng thanh toán.
(`scripts/smoke.sh` đã hết nợ: chạy thật trên Linux 08/09, xem SETUP §9 đợt 18.)

---

## Docs — đọc cái nào khi nào

| File | Trả lời |
|---|---|
| `docs/HOC.md` | **lộ trình học, buổi 0 tới buổi 9** (buổi 0 = bấm UI trước khi đọc code) + 17 câu tự kiểm tra + nói gì khi phỏng vấn |
| `docs/SETUP.md` | môi trường, cây thư mục, lệnh hàng ngày, **§9 nhật ký 18 đợt review** |
| `docs/WALKTHROUGH.md` | đọc code theo thứ tự — 97 file, §0–§22 (§22 = đợt variant/size) |
| `docs/DDD.md` | sổ tay khái niệm, đối chiếu Symfony |
| `docs/FLOW-ORDER.md` | một đơn từ đầu tới cuối: ai gọi, ai nghe, bảng nào đổi, hỏng thì sao |
| `docs/CATALOG.md` | vì sao không dùng affiliate feed, provenance theo nhóm |
| `docs/GO-CHO-PHP.md` | cú pháp Go tra nhanh cho người viết PHP |
| `docs/P10-PLAN.md` | **việc đang làm**: operator đặt hộ có ghi tên — file phải sửa, ba cái bẫy, test bắt buộc |
| `docs/P9-PLAN.md` | plan gốc của P9 (đã xong hết) — giữ để thấy cách chốt quyết định |
| `docs/UI-GUIDE.md` | **thao tác UI từng bước để chạy thử luồng**: dựng lên, token, 26 bước, 10 nhánh rẽ nên thử, bảng hỏng-thì-xem |

**`docs/SETUP.md` §9 là chỗ quan trọng nhất khi tiếp tục việc**: 18 đợt review,
mỗi đợt một bảng "quyết định / bug → chỗ nó nằm". Đọc 2–3 đợt cuối là nắm được
vì sao code hiện tại trông như vậy.

## Lệnh hay dùng

```bash
go test ./...                     # 263 test, < 2 giây
gofmt -l . && go vet ./...        # phải sạch trước khi báo xong

# Có Postgres (28 test tích hợp)
docker compose up -d
PORTAGE_TEST_DSN="postgres://portage:portage@localhost:5432/portage_test?sslmode=disable" go test ./...

# Smoke thật: 2 binary + Postgres thật + bearer token thật
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1

# Console thật trong trình duyệt (cùng origin, không CORS)
go run ./cmd/api -web ./web      # rồi mở http://localhost:8080/ui/console.html
```

Chi tiết hơn: `docs/SETUP.md` §5.

## Quy ước code (đầy đủ ở SETUP.md §6)

1. Input của **lập trình viên** sai → `panic`; input của **người dùng** sai → `error`
2. Tiền là `int64` minor units, không bao giờ `float`
3. Làm tròn half-up, ra xa số 0
4. Constructor có hai dạng: `New*` trả `error`, `Must*` panic
5. Typed ID bằng embedding `shared.ID`
6. Domain **ghi** event, không **phát** event
7. `Clock` là port — domain không gọi `time.Now()`
8. Adapter không tự validate — gọi `Parse*` của domain
9. Zero value phải an toàn
10. Mọi dependency vào một struct `Deps` + `mustHave` panic lúc dựng

**Luật vàng:** `internal/domain/` chỉ import stdlib + allowlist
(`github.com/google/uuid`). CI và `internal/domain/decisions_test.go` cùng canh
— vi phạm là build đỏ.

---

## Giao diện

Hai màn hình, hai việc: **`flow.html`** để *hiểu* luồng (mô phỏng), **`console.html`** để *chạy thử*
luồng (gọi API thật). Cách thao tác từng bước: `docs/UI-GUIDE.md`.

**`web/console.html` + `web/app/{api,steps,main}.js`** — console thật, không npm/build/framework.
Chạy `go run ./cmd/api -web ./web` rồi mở `http://localhost:8080/ui/console.html` (cờ `-web` mount
một `http.FileServer` ở `/ui/` **cạnh** API: cùng origin nên không cần CORS, và bearer token trong
`fetch()` đi qua đúng `authenticate()` mà curl đi qua). Tab **Chạy luồng** bấm 26 bước gọi API thật,
tự thử lại đúng chỗ `listing_not_found` / `quote_not_accepted` và poll đúng chỗ có endpoint đọc —
nên nhìn thấy relay outbox thật. Bốn tab còn lại là màn hình khách / nhân viên / chìa khoá / call log.
Đã kiểm 06/09: 26 bước xanh, kết ở `variance -2.50 USD`.

**`web/flow.html`** — mở bằng double-click, không cần build, không gọi API. Một bản Portage mô
phỏng trong trình duyệt: bấm qua 26 bước của một đơn, mỗi bước hiện màn hình người dùng thật sự
thấy + request + event ghi vào outbox + ai nghe event đó. Rail bên phải là trạng thái sống
(ladder đơn hàng, sổ tiền, outbox pending→sent, số dòng từng bảng chiếu).

Ba thứ nó dạy được mà đọc code khó thấy: **worker là một bước riêng** (6 lượt relay ra đúng
5/2/2/2/6/3 event như log thật), **bảng đọc chỉ đầy sau relay** (nên `POST /quotes` sớm là 404),
và **một event nhiều listener** (`purchase_confirmed` có 4). Tab `Bản đồ event` là 31 dòng
`wire.Subscribe`; tab `Nhánh rẽ` là các đường hỏng kèm mã lỗi + bảng huỷ/hoàn cọc.

Số liệu trong đó lấy từ lần `scripts/smoke.ps1` đã verify (5.393.720 ₫ · 2.696.860 ₫ ·
163.22 + 27.50 USD → variance −2.50 USD), nên nếu sửa nghiệp vụ mà file này còn số cũ là docs
đã lệch code.

Mockup shadcn cũ (artifact, account cá nhân):
https://claude.ai/code/artifact/9ee4e9ef-c0ac-4044-83a4-ca0a50f4503e
