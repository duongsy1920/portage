# plan — T-001 ratecard-yaml

*Viết 2026-09-25 từ `analysis.md` (Q1–Q4 đã duyệt 23/09) và từ code đọc lại hôm nay.*

## Mục tiêu
Đưa các số nghiệp vụ của báo giá và cước — hiện là hằng trong
`internal/platform/wire/wire.go` — ra file cấu hình `config/ratecard.yaml`, để người vận
hành đổi bảng giá mà không sửa code, không build lại.

## Hình dạng sau khi xong — để đọc bước cho nhanh
- **`config/ratecard.yaml`** (mới): ba nhóm `quote` · `classes` · `lanes` (danh sách, hôm
  nay một phần tử). Mọi số là **chuỗi**; mỗi số một dòng chú thích "từ đâu". Không có tỷ giá.
- **`internal/adapter/config`** (package mới): `Load(path)` và `Parse(io.Reader)` trả
  `config.Pricing{Policy, Classification, Lanes}`. Thư viện yaml **chỉ** ở đây. Không tự
  validate — gọi `shared.ParsePercent/ParseMoney`, `pricing.New*`, và bọc lỗi bằng **tên
  trường** (`ratecard: quote.deposit: …`). Khoá lạ trong file → lỗi (`KnownFields(true)`),
  để gõ sai tên trường không âm thầm rơi về mặc định.
- **`wire.Postgres(ctx, clk, dsn, ratecard string)`**: nạp file **trước** khi kết nối
  (rẻ, cục bộ, hỏng thì hỏng ngay và nói đúng file). `wire.Memory()` **giữ** hằng làm
  fixture — tên hàm `quotePolicy()` · `goodsClasses()` · `seedPricing()` **giữ nguyên** để
  bốn chỗ trong `docs/SETUP.md` trỏ tới chúng không thành lỗi thời; chỉ đổi comment.
- **Cờ `-ratecard`** ở `cmd/api` và `cmd/worker`, **mặc định `config/ratecard.yaml`**
  (tương đối cwd). Mọi lệnh trong docs và smoke đều chạy từ gốc repo (`go run ./cmd/api`
  chỉ chạy được từ đó), nên không lệnh nào phải đổi; production truyền đường dẫn tuyệt đối.
  Memory mode không đọc file — ghi rõ trong usage của cờ.
- **Một test cột hai nguồn**: fixture trong `wire.Memory` phải bằng từng giá trị của
  `config/ratecard.yaml`. Đổi một bên mà quên bên kia → đỏ.

## Bước
1. **Dependency.** `go get gopkg.in/yaml.v3` → `go.mod`, `go.sum` (dep trực tiếp thứ ba).
2. **File cấu hình.** Viết `config/ratecard.yaml` với đúng giá trị hiện tại của
   `wire.go:279–287` (quote), `:299–303` (classes), `:354–366` (lane). Chú thích nguồn
   từng số (SETUP §7). Tỷ giá **không** vào file.
3. **Package `internal/adapter/config`.** `ratecard.go`: struct yaml (chuỗi cho mọi số),
   `Parse` (decoder `KnownFields(true)`), `Load` (mở file, bọc lỗi `ratecard <path>: …`).
   Lane: `code` qua `pricing.ParseLaneCode`; `currency` USD/VND; `duty: bundled`, hoặc
   `duty: itemised` + `duty_rate`. Thiếu `lanes` hay danh sách rỗng → lỗi.
4. **Test của package** `ratecard_test.go`, **bảng** (table-driven — repo chưa có cái nào,
   `learn/PLAN-BO-SUNG.md` đang nợ mẫu này): một file đúng → mọi giá trị khớp; mỗi dòng
   sai một trường (cọc 150, mức cước âm, thiếu một mức, khoá lạ, `lanes: []`, số không
   bọc ngoặc kép) → lỗi **chứa tên trường**.
5. **`wire.go`.** `Postgres` thêm tham số `ratecard string`, `config.Load` trước
   `postgres.Connect`; `Policy`/`Classification` của Deps lấy từ file (`:225–226`);
   `seed(ctx, g, lanes)` và `seedPricing(ctx, deps, lanes)` nhận danh sách lane, lặp
   `DefineLane`; `Memory` truyền `[]pricing.LaneDetails{fixtureLane()}` — `fixtureLane()`
   là phần lane tách ra từ `seedPricing` hiện tại. Sửa comment `:275–277`: hằng giờ là
   **fixture cho dev/test**, production đọc file, test ở bước 6 giữ hai cái bằng nhau.
6. **`wire_test.go`.** `:345` và `:386` truyền `rateCardFile = "../../../config/ratecard.yaml"`.
   Thêm `TestPostgres_refusesAMissingRateCardBeforeConnecting` (DSN sai **và** file sai →
   lỗi phải nói về file, không nói về DSN; chạy được không cần Postgres) và
   `TestFixtureMatchesTheRateCardFile` (nạp file thật, so policy, ba category + một
   category lạ qua `ClassOf`, và lane trong repo `g.Pricing.Lanes` sau seed).
7. **`cmd/api/main.go`, `cmd/worker/main.go`.** Cờ `-ratecard` mặc định
   `config/ratecard.yaml`, truyền vào `wire.Postgres`. Lỗi → `log.Fatalf` như DSN
   (`api:79`, `worker:50`). Không thêm gì cho memory mode.
8. **Docs.** `CLAUDE.md:137–138` bỏ dòng nợ, thêm một dòng vào "Số nghiệp vụ đã chốt":
   nguồn là `config/ratecard.yaml`, fixture trong `wire.go`, một test giữ hai cái bằng
   nhau. `docs/SETUP.md`: bỏ checkbox `:852`; cột nguồn ở bảng §7 (`:818–825`) trỏ thêm
   file; thêm **Đợt 22 (25/09)** ngắn ở §9 theo mẫu các đợt trước.
9. **Định dạng.** `gofmt -l .` sạch trước khi rời tay. Test và smoke là việc của `verify`.

## File sẽ đụng — danh sách ĐÓNG
- `go.mod`, `go.sum` — thêm `gopkg.in/yaml.v3`
- `config/ratecard.yaml` — **tạo mới**
- `internal/adapter/config/ratecard.go` — **tạo mới**
- `internal/adapter/config/ratecard_test.go` — **tạo mới**
- `internal/platform/wire/wire.go` — chữ ký `Postgres`, `seed`/`seedPricing` nhận lane, comment
- `internal/platform/wire/wire_test.go` — hai lời gọi cũ, hai test mới
- `cmd/api/main.go` — cờ `-ratecard`
- `cmd/worker/main.go` — cờ `-ratecard`
- `CLAUDE.md` — "Còn nợ", "Số nghiệp vụ đã chốt"
- `docs/SETUP.md` — `:852`, bảng §7, Đợt 22
(implement không được đụng file ngoài danh sách này. Cần thêm → về hỏi.)

## Test sẽ thêm hoặc đổi
- `internal/adapter/config/ratecard_test.go` (mới, bảng) — acceptance 3: file sai → lỗi
  nói rõ trường; acceptance 4: file đúng → từng giá trị bằng số đã chốt.
- `wire_test.go`: `TestPostgres_refusesAMissingRateCardBeforeConnecting` (mới) —
  acceptance 3, không cần Postgres. `TestFixtureMatchesTheRateCardFile` (mới) — acceptance
  4, chặn hai nguồn lệch nhau. `TestPostgres_runsTheWholeFlow`, `TestPostgres_badDSNFailsAtStartUp`
  (đổi lời gọi) — acceptance 1.
- Mốc đo trước task: `go test ./... -v | grep -c '^--- PASS'` = **259** (không có
  `PORTAGE_TEST_DSN`, 2026-09-25). Acceptance 1 là số này không giảm.
- `decisions_test.go` không đổi — guard allowlist tự chứng minh acceptance 5.

## Không làm
- **Không** đổi bất kỳ giá trị nào. File ghi đúng số hôm nay; `TestFixtureMatchesTheRateCardFile`
  và số vàng của smoke là bằng chứng.
- **Không** đưa tỷ giá vào file (Q1). **Không** hot-reload, **không** endpoint đọc/ghi cấu
  hình, **không** biến môi trường — không có yêu cầu nào cần.
- **Không** đổi tên `quotePolicy()`/`goodsClasses()`/`seedPricing()` dù giờ là fixture — vì
  `docs/SETUP.md:821,823,825,1129` trỏ tới chúng; đổi comment đủ nói sự thật.
- **Không** bắt buộc cờ `-ratecard` — mặc định cwd-relative đúng với cách mọi lệnh trong
  docs đang chạy; nhờ vậy `scripts/smoke.sh`, `smoke.ps1`, `docs/UI-GUIDE.md:30–32`,
  `docs/WALKTHROUGH.md:57`, `ci.yml` **không phải đụng**. (Analysis hôm nay đã nêu worker
  và smoke; đây là cách đáp mà không mở rộng danh sách file.)
- **Không** sửa lịch sử `docs/SETUP.md:1129` và `:1331` — là nhật ký của đợt cũ. Acceptance
  6 của TASK.md viết "hai dòng nợ" → đề nghị hiểu là `CLAUDE.md:137–138` + `SETUP.md:852`.
- **Không** đụng `cmd/api` memory mode, `web/`, `learn/`.

## Tự phản biện
- **Cách hiển nhiên hơn là:** chỉ đưa bốn mức cước ra file, hoặc giữ hằng và thêm cờ
  override từng số. Không chọn vì: nợ trong `CLAUDE.md` ghi "bảng giá đang là hằng" — đổi
  thuế hay phí vẫn phải build lại thì nợ chưa trả; và Q1 đã chốt phạm vi.
- **Cách hiển nhiên thứ hai:** cờ `-ratecard` **bắt buộc**, để không có "mặc định ngầm".
  Không chọn vì: mặc định hiện ra trong `-h` và trong thông báo lỗi khi thiếu file, nên
  không ngầm; còn bắt buộc thì kéo theo 5 file nữa (hai smoke, hai docs, CI) chỉ để nói lại
  điều `go run ./cmd/api` đã giả định: chạy từ gốc repo.
- **Chỗ kế hoạch này có thể sai:** (1) `TestFixtureMatchesTheRateCardFile` cần so
  `QuotePolicy` — nếu struct không so `==` được (có `MarginPolicy` chứa `Money`), so qua
  getter từng trường; (2) `yaml.v3` với số **không** bọc ngoặc kép vào trường `string` —
  tin là được, nhưng bảng test ở bước 4 có một dòng kiểm đúng chuyện này, nếu đỏ thì file
  bắt buộc ngoặc kép và usage nói rõ; (3) `go get` cần mạng lúc implement.
- **Đề nghị: làm theo kế hoạch này.**

## Rủi ro
- `go get` không có mạng → implement dừng ở bước 1, không tự xoay sang vendoring.
- Smoke: cổng 5433 của máy này đang bị `rift-db-1` chiếm (SETUP §9 đợt 20). `verify` sẽ
  ghi `not run` cho smoke nếu không có Postgres rảnh; biết bằng bảng Verify. Cách đã dùng
  ở đợt 20 (Postgres tạm trên cổng khác) vẫn dùng được.
- Hai process cùng `DefineLane` lúc khởi động — đã là như vậy hôm nay, không mới.

## Acceptance (chép từ TASK.md)
- [ ] `go test ./...` xanh, số test không giảm so với trước task — mốc **259** `--- PASS` không DSN
- [ ] `scripts/smoke.sh` in đúng số vàng: `total 5393720 VND, deposit 2696860` và `variance -2.50 USD`
- [ ] Api khởi động với file cấu hình thì dùng số trong file; file thiếu, không parse được,
      hoặc số sai → api **không** khởi động và in một dòng lỗi nói rõ **trường nào** sai
- [ ] Không số nghiệp vụ nào đổi giá trị — bảng "Số nghiệp vụ đã chốt" trong `CLAUDE.md`
- [ ] `internal/domain/` không import thêm gì ngoài allowlist; 7 guard `decisions_test.go` xanh
- [ ] `CLAUDE.md:137–138` và `docs/SETUP.md:852` gỡ dòng nợ; số trong docs khớp code
      *(sửa so với TASK.md: `:1331` là lịch sử, giữ)*

## Sửa đề nghị sau gate hỏng 2026-09-25T14:36:04+07:00

**Chẩn đoán.** `TestParse_rejectsOneWrongFieldByName/no_lanes` đỏ. Dòng test đó tạo file
"không có lane" bằng cách đổi `lanes:` thành `lanes: []` **và** dời khối lane sang một khoá mới
`unused:` — nhưng `Parse` bật `KnownFields(true)` (đúng thiết kế: khoá lạ là lỗi), nên lỗi trả về
là *"field unused not found"* thay vì *"lanes: want at least one lane"*. **Code đúng, dòng test
sai**: nó dùng chính cái mà nó đang kiểm ở dòng `unknown key`. Lỗi ở `ratecard_test.go:99`
(hàng `no lanes`).

Nhân tiện thấy khi đọc lỗi: `Parse` bọc lỗi yaml bằng tiền tố `yaml:` trong khi thư viện đã tự
in `yaml:` → thông báo thành *"yaml: yaml: unmarshal errors"*. Nhỏ, nhưng người vận hành đọc.

**Đề nghị** (hai file, cả hai đã trong danh sách đóng — không thêm file):
1. `ratecard_test.go`: tách `good` thành `head + lanesBlock`; hàng `no lanes` dùng
   `head + "lanes: []\n"` thay vì thay chuỗi. Thêm trường `file` (tuỳ chọn) cho bảng để hàng
   nào cần thì đưa nguyên file thay vì cặp from/to.
2. `ratecard.go`: `fmt.Errorf("yaml: %w", err)` → trả `err` nguyên, thư viện đã nói đủ.

**Vì sao:** giữ `KnownFields(true)` — đó là hành vi cố ý (khoá lạ không được rơi về mặc định) và
đang được hàng `unknown key` chứng minh; sửa test cho đúng cái nó định kiểm.

**Đã bỏ:** tắt `KnownFields` cho hàng đó — bỏ đi đúng lớp bảo vệ mà task này thêm vào. Xoá hàng
`no lanes` — mất kiểm cho "file rỗng lane", là trường hợp thật khi người xoá nhầm.

Sau khi sửa: `implement` chạy lại hai bước 3–4 phần này, rồi `verify` chạy lại toàn bộ.
**Đề nghị: duyệt sửa này.**

## Sửa đề nghị sau gate hỏng 2026-09-25T15:18:37+07:00 — lần 2

**Chẩn đoán.** Mục sửa lần 1 thêm trường `file` vào struct của bảng test, nhưng 10 hàng cũ vẫn
viết theo **vị trí** (4 giá trị) → *"too few values in struct literal"* ở `ratecard_test.go:103–110`.
Test không biên dịch, nên `go vet` và `go test` cùng đỏ. Lỗi của agent khi implement lần 1: kiểm
biên dịch bằng `go build` — lệnh này **không** biên dịch file `_test.go`.

**Đề nghị** (một file, trong danh sách đóng): viết cả 11 hàng theo **tên trường** (`name:`,
`from:`, `to:`/`file:`, `want:`) — đọc được, thêm trường sau không vỡ. Tiện tay: gom chuỗi DSN
giả (tới cổng 1 của máy, không gì lắng nghe) trong `wire_test.go` thành một hằng `unreachableDSN`
dùng ở hai test, vì `secret.scan` đánh dấu nó hai lần và một chỗ khai dễ đọc hơn hai chỗ chép.
**Vì sao:** hàng theo tên là cách Go thường viết bảng test có trường tuỳ chọn. **Đã bỏ:** bỏ trường
`file`, dựng hàng "no lanes" bằng thay chuỗi ba bước — khó đọc hơn, và chính cách đó đã sai lần 1.

Kiểm biên dịch cho lần implement này: `go vet ./internal/adapter/config` (biên dịch cả test).
**Đề nghị: duyệt sửa này.**
