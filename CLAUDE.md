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

## Trạng thái (15/09/2026 — P9 xong, variant/size, migrate race, hai màn hình theo vai, P10 xong)

```
5 bounded context + 1 tầng đọc · 37 domain event · 41 route · 35 dòng Subscribe
312 test (311 PASS + 1 SKIP cố ý) · 278 test chạy < 2 giây không cần Docker
7 test canh kiến trúc bằng go/ast · 13 migration
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

**Vừa xong — hai màn hình theo vai.** `web/{customer,staff}.html` là hai trang React
(thư viện vendor trong `web/vendor/`, **không có bước build**) cho hai công việc thật, thay
vì một nút cho mỗi endpoint như bảng kiểm cũ. Kèm bảng đọc thứ hai `product_worklist`,
event `catalog.listing_confirmed`, và hai trường của *yêu cầu* trên bản nháp:
`RequestedBy` (ai đang đợi) và `RequestedVariant` (họ xin size nào, bằng chữ của họ).
Hai quy tắc của màn hình: không chữ nào của máy ra tới UI (`web/app/words.js`), và mọi con
số nói được nó ở đâu ra. Đọc `docs/WALKTHROUGH.md` §23 và `docs/UI-GUIDE.md`.

**Vừa xong — migrate race (đợt 18):** `Migrate` tạo bảng `schema_migrations`
**ngoài** advisory lock, và `CREATE TABLE IF NOT EXISTS` của PostgreSQL không
nguyên tử: hai process khởi động cùng lúc trên database trống thì một cái vỡ ở
`pg_type_typname_nsp_index`. Chỉ hiện khi database **cold**, nên Windows chưa
bao giờ thấy. Tìm ra bằng lần chạy thật đầu tiên của `scripts/smoke.sh` trên
Linux. Có test hồi quy dùng database dùng-một-lần.

**Vừa xong — P10 (đợt 20, 15/09):** operator đặt hộ có ghi tên (`docs/P10-PLAN.md`,
`WALKTHROUGH.md` §24). `POST /orders` mở cho cả hai loại chìa: khách đặt cho mình,
operator đặt hộ phải kèm `customer_id` trong body (400 `customer_required` nếu
thiếu, 403 nếu khách tự gửi kèm — đang thử đặt hộ người khác). `ordering.OrderDetails.PlacedBy`
(`shared.OperatorID`, zero = khách tự đặt) ghi lại ai đặt hộ; `POST /quotes/{id}/accept`
mở cho cả hai vì báo giá không có chủ để gán. Migration `0012_placed_by.sql`, nullable,
không backfill. **Phần tuỳ chọn cũng đã làm luôn** (cùng đợt 20, sau khi hỏi lại):
`PlacedBy` giờ hiện trên bảng đọc `reporting` (`0013_summary_placed_by.sql`) và trên
CẢ HAI màn hình — `web/app/customer.js` (dòng "nhân viên đặt hộ bạn") và
`web/app/staff.js` (dòng "đơn đặt hộ (nhân viên)" trong hàng đợi thu tiền).
`go test ./...` xanh hai lần (272→278 test không cần Docker, +6; 311 PASS + 1 SKIP
với `PORTAGE_TEST_DSN`, +7). `scripts/smoke.sh` thật cũng xanh hai lần sau cả hai
đợt việc, đúng số vàng mọi lần — nhưng qua một Postgres **tạm**, vì cổng 5433 cố
định của máy Linux này đang bị dự án khác chiếm (`rift-db-1`); xem SETUP §9 đợt 20.

**Vừa xong — làm lại giao diện (đợt 21, 23/09):** theo `docs/UI-REDESIGN-PLAN.md`, không đổi Go
hay API. Xem mục **Giao diện** bên dưới và SETUP §9 đợt 21.

**Còn nợ:** trường `balance` trong `GET /orders` (màn hình đang tự trừ tổng − cọc) ·
`config/ratecard.yaml` (bảng giá đang là hằng trong `wire.go`) ·
một `adapter/merchant/<shop>.go` thật (chưa shop nào cho API) · cổng thanh toán ·
chạy lại `scripts/smoke.sh` qua `portage-postgres` thật (không phải bản tạm) khi
cổng 5433 rảnh lại.

**Quyết định 08/09 — HOÃN, đừng đề xuất lại:** một đơn = một variant = **một cái**. Không
có `Quantity`, và một kiện không chứa được nhiều đơn. Anh biết hai chỗ đó là đơn giản hoá và
chọn để nguyên: *"sau này chạy thật xem thực tế lãi lỗ như nào"*. Thứ sẽ trả lời chính là
bảng `reconciliations` (DDD.md §28) — `variance` âm liên tục qua nhiều đơn nghĩa là bảng giá
cước hoặc phép đo đang sai, và lúc đó mới có dữ liệu để quyết đổi mô hình. Chi tiết ba chỗ
đơn giản hoá: kích thước đo trên **một hộp giày** chứ không phải thùng gom; lô hàng
**không được đo**, chỉ nhận hoá đơn của hãng bay rồi chia; cân quy đổi tính trên từng kiện,
không trên pallet.

---

## Docs — đọc cái nào khi nào

| File | Trả lời |
|---|---|
| `docs/HOC.md` | **lộ trình học, buổi 0 tới buổi 9** (buổi 0 = bấm UI trước khi đọc code) + 17 câu tự kiểm tra + nói gì khi phỏng vấn |
| `docs/SETUP.md` | môi trường, cây thư mục, lệnh hàng ngày, **§9 nhật ký 21 đợt review** |
| `docs/WALKTHROUGH.md` | đọc code theo thứ tự — 101 file, §0–§23 (§23 = hai màn hình theo vai) |
| `docs/DDD.md` | sổ tay khái niệm, đối chiếu Symfony |
| `docs/FLOW-ORDER.md` | một đơn từ đầu tới cuối: ai gọi, ai nghe, bảng nào đổi, hỏng thì sao |
| `docs/CATALOG.md` | vì sao không dùng affiliate feed, provenance theo nhóm |
| `docs/GO-CHO-PHP.md` | cú pháp Go tra nhanh cho người viết PHP |
| `docs/UI-REDESIGN-PLAN.md` | plan của đợt 21 (đã xong) — làm lại giao diện bằng hai plugin thiết kế |
| `docs/UI-NEXT-PLAN.md` | **plan đợt sau, chưa làm**: đưa trí nhớ phiên của màn hình về bảng đọc (migration `0014`), `balance`, hai route đọc; §7 chờ anh chốt |
| `docs/P10-PLAN.md` | plan gốc của P10 (đã xong, đợt 20) — operator đặt hộ có ghi tên; giữ để thấy cách chốt quyết định |
| `docs/P9-PLAN.md` | plan gốc của P9 (đã xong hết) — giữ để thấy cách chốt quyết định |
| `learn/CURRICULUM.md` | **loạt video học, 24 tập**: Mùa 1 Go (10), Mùa 2 DDD (6), Mùa 3 Portage (8). Kèm bốn cửa kiểm ở §12 |
| `learn/HUB-PLAN.md` | **plan chưa làm**: trang web để đi hết lộ trình học (bản đồ 24 tập, phát tập bằng `@remotion/player`, bản in, tự kiểm tra, đối chiếu roadmap.sh); §8 chờ anh chốt |
| `docs/UI-GUIDE.md` | **bốn màn hình và bấm gì trên từng cái**: trang khách, trang nhân viên, mô phỏng, bảng kiểm API; 10 nhánh rẽ nên thử; bảng hỏng-thì-xem |
| `docs/CODING-AGENT.md` | **thiết kế Coding Agent** (đồng nghiệp AI dùng lại được cho mọi project): audit Portage → ai-employees → mô hình → Phase 1. Đã dựng ở `agent/` |

**`docs/SETUP.md` §9 là chỗ quan trọng nhất khi tiếp tục việc**: 21 đợt review,
mỗi đợt một bảng "quyết định / bug → chỗ nó nằm". Đọc 2–3 đợt cuối là nắm được
vì sao code hiện tại trông như vậy.

## Loạt video học — `learn/`

Thư mục `learn/` là một project Remotion **tách hẳn** khỏi code Go (không import
gì của nhau). 24 tập, mỗi tập một video và một bản in A3:

```
learn/out/mua1-go/       10 tập — Go cho người viết PHP
learn/out/mua2-ddd/       6 tập — DDD, bằng ví dụ Symfony đã quen
learn/out/mua3-portage/   8 tập — vì sao project thật này trông như vậy
```

**Thêm một tập** = một `spec.ts` + vài cảnh + một dòng trong `src/videos/index.ts`.
Không sửa `Root.tsx`, không sửa script nào.

**Bốn cửa kiểm trước khi báo xong** (`tsc` xanh không chứng minh được gì):

```bash
cd learn
./scripts/verify-numbers.sh        # 11 con số vs repo Go thật
python3 scripts/verify-snippets.py # mọi dòng code lên hình có thật trong file nó khai
python3 scripts/check-overflow.py  # 110 cảnh, không cảnh nào tràn khung
./scripts/render-plates.sh         # 24 bản in → PNG @2x + PDF A3
```

Chi tiết: `learn/CURRICULUM.md`.

## Coding Agent — `agent/`

Thư mục `agent/` là **Phase 1 của Coding Agent** (`docs/CODING-AGENT.md`): bộ file Markdown,
không runtime, không `go.mod`, harness (Claude Code) đọc và làm theo. Nó **không phải** một
phần của Portage — Portage chỉ là project đầu tiên nó làm việc trên (`agent/projects/portage/`).
Thêm project khác = thêm một `PROJECT.md`, không sửa gì khác.

```
/agent intake                    tạo task từ mô tả trong chat
/agent analyze T-001             rồi plan → implement → verify → review, người gọi từng bước
/agent approve plan T-001 [note] duyệt một cửa (plan · commit · push) — chỉ người quyết
./agent/scripts/verify-agent.sh  16 check cơ học, đỏ là chưa xong
```

Hai luật chi phối cách agent nói với anh: **mọi điểm dừng đến kèm đề nghị** (anh duyệt, không
thiết kế — `agent/CONTRACT.md` §5.0) và **gate hỏng đi qua cổng plan kèm đề nghị sửa**
(ADR-004). Ba công tắc của anh: file `agent/PAUSED`, mục `## Corrections` ở đuôi mọi file,
và `releases` trong `PROJECT.md`. Đọc `agent/README.md` trước khi gọi.

## Lệnh hay dùng

```bash
go test ./...                     # 278 test, < 2 giây
gofmt -l . && go vet ./...        # phải sạch trước khi báo xong

# Có Postgres (34 test tích hợp)
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

Bốn trang, bốn vai (`/ui/` là trang chọn): **`customer.html`** và **`staff.html`** để *làm việc*,
**`flow.html`** để *hiểu* luồng (mô phỏng), **`console.html`** để *chạy thử* luồng (gọi API thật).
Cách thao tác từng bước: `docs/UI-GUIDE.md`.

**Hai trang làm việc (làm lại đợt 21, 23/09).** Hệ thiết kế "phong bì *par avion*": nền giấy
xanh-xám, mực xanh đen, sọc xanh–đỏ **chỉ** ở chân thanh trên cùng và chỗ "việc đang chờ bạn";
đỏ không bao giờ là lỗi. Be Vietnam Pro **tự host** (`web/vendor/fonts/`, `web/app/fonts.css`),
light + dark. Chỗ táo bạo duy nhất là **dải hành trình** sáu ô (`Journey` trong `ui.js`, logic
`journeyOf` trong `words.js`): ô nào API không gửi số thì để trống. Trang nhân viên là bàn làm
việc: năm hàng đợi có số đếm bên trái (món chờ xử lý · việc đi mua · kho Denver · tiền chờ thu ·
chờ giao), phiếu đang mở bên phải — đi được hết luồng tới lúc giao, kết ở `variance -2.50 USD`. Luật màu/chữ/bố cục:
`design-system/portage/MASTER.md`; plan gốc: `docs/UI-REDESIGN-PLAN.md`; nhật ký: SETUP §9 đợt 21.
`flow.html` **chưa** đổi (quyết định D3 của plan).

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
