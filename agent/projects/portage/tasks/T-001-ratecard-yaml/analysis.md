# analysis — T-001 ratecard-yaml

*Run 1 — 2026-09-23 12:16 · `challenge: true` → mục câu hỏi. Người trả lời 23/09: đồng ý
hết. Run 2 — 2026-09-25 14:01 → mục kết luận ở dưới, mọi `file:dòng` đã đọc lại hôm nay
(repo có 2 commit mới từ 23/09; số dòng `wire.go` không đổi, `CLAUDE.md` "Còn nợ" dời
xuống :137–138).*

## Câu hỏi trước khi kết luận

Bốn câu, mỗi câu đổi kế hoạch tuỳ trả lời. Trả lời **một dòng** ở `## Corrections` của
`TASK.md`: *"đồng ý hết"*, hoặc *"Q2: JSON"*, *"Q4: bỏ hằng"*…

### Q1. File chứa **những gì** — chỉ bốn mức cước, hay mọi số nghiệp vụ?
*(Tự trả lời trước nếu muốn học — đề nghị của agent ở dưới.)*
**Đề nghị:** mọi số trong `wire.go:279–366`, ba nhóm: `quote` (thuế 8,81 % · phí 10 % ·
sàn 500 000 ₫ · cọc 50 % · hạn 48 h), `classes` (category → GoodsClass, `:297–303`),
`lane` (mã `us_forwarder` · divisor 5000 · bước 500 g · bốn mức cước 9/10/12/14 USD · phụ
phí pin 3 USD · duty bundled, `:354–366`). **Không** đưa tỷ giá 26 000 (`:378`).
**Vì sao:** ba nhóm đó cùng đổi khi đổi hãng bay hay đổi bang kho — comment ở `wire.go:276`
gọi chúng là "the set of constants every quote is built with". Tỷ giá đổi **theo ngày** và đã
có đường riêng: `POST /fx` → `SetExchangeRateHandler` (`wire.go:378` chính là seed đi qua
handler ấy). Hai chu kỳ đổi khác nhau thì hai chỗ chứa khác nhau.
**Đã bỏ:** chỉ đưa bốn mức cước (nghĩa đen của "ratecard") — vì đổi thuế hay phí vẫn phải
build lại, nợ chỉ trả một nửa. Đưa cả tỷ giá — vì thành hai nguồn sự thật cho một con số
mà `POST /fx` đang đổi lúc chạy.

### Q2. Định dạng và dependency — YAML (thêm thư viện) hay JSON (stdlib)?
*(Tự trả lời trước nếu muốn học — đề nghị của agent ở dưới.)*
**Đề nghị:** YAML, parse bằng `gopkg.in/yaml.v3` trong package mới
`internal/adapter/config`. `go.mod` thêm **một** dependency trực tiếp (hiện có hai: uuid,
pgx). Mọi con số trong file là **chuỗi** (`sales_tax: "8.81"`, `floor: "500000"`) và đi qua
`shared.ParsePercent` / `shared.ParseMoney` — không bao giờ qua `float64`.
**Vì sao:** file này do **người vận hành sửa tay**, cần dòng chú thích kiểu
`# 9.00 — bảng giá hãng bay 09/2026` — đúng tinh thần "mọi con số nói được nó ở đâu ra"
(`docs/UI-GUIDE.md`); JSON không có comment. Tên nợ trong `CLAUDE.md` đã là `.yaml`. Đặt ở
`adapter/` nên luật vàng không bị đụng — allowlist ở `decisions_test.go:226–229` chỉ canh
`internal/domain/`. Số là chuỗi vì quy ước 2 (`int64` minor units, không float).
**Đã bỏ:** JSON + `encoding/json` (không thêm dep) — mất comment, và người vận hành phải
nhớ bọc số trong ngoặc kép để né float. Biến môi trường — hơn 15 giá trị, không diễn tả được
`classes` và `lane` có cấu trúc.

### Q3. File thiếu hoặc số sai — `panic` hay `error`? Ai truyền đường dẫn?
*(Tự trả lời trước nếu muốn học — đề nghị của agent ở dưới.)*
**Đề nghị:** **`error`**. `config.Load(path)` trả lỗi **có tên trường**
(`quote.deposit: "150": want 0 < deposit ≤ 100%`); `wire.Postgres` nhận thêm đường dẫn và
trả lỗi lên; `cmd/api` `log.Fatalf` như đang làm với DSN (`cmd/api/main.go:79`). Đường dẫn
qua cờ mới `-ratecard config/ratecard.yaml`, cùng kiểu với `-dsn` (`main.go:39–44`). Ba
constructor domain đã trả `error` sẵn — `NewMarginPolicy` (`policy.go:18`), `NewQuotePolicy`
(`policy.go:72`), `NewRateCard` (`lane.go:84`) — dùng nguyên, bỏ `Must*` trên đường đọc file.
**Vì sao:** luật 1 `CLAUDE.md`: input của **người dùng** sai → `error`. Comment ở
`wire.go:277` — *"A wrong constant here is a programmer error, hence Must-style panics"* —
đúng khi số là hằng trong code, **sai** ngay khi số đến từ file người vận hành sửa. Task này
đổi đúng chỗ đó, nên comment ấy phải đổi theo.
**Đã bỏ:** giữ `panic`/`Must*` — người vận hành nhận stack trace thay vì "trường nào sai".
Biến môi trường cho đường dẫn — `cmd/api` chưa dùng env ở đâu, thêm một cách cấu hình thứ
hai không có lý do.

### Q4. Test và chế độ dev đang dựa vào hằng — giữ hằng làm fallback không?
*(Tự trả lời trước nếu muốn học — đề nghị của agent ở dưới.)*
**Đề nghị:** `wire.Memory()` **giữ** hằng, đổi tên rõ là fixture (`fixtureQuotePolicy`,
`fixtureLane`) kèm comment *"phải bằng config/ratecard.yaml"*; `wire.Postgres()` **bắt buộc**
đọc file, không fallback. Thêm **một** test trong `wire_test.go`: nạp `config/ratecard.yaml`
thật rồi so từng giá trị với fixture — hai nguồn không thể lệch nhau mà test còn xanh.
**Vì sao:** chỉ **một** file test gọi `wire.Memory` (`wire_test.go`); năm file test tự dựng
`QuotePolicy` (`repos_test.go:350`, `pricing_test.go:33`, `server_test.go:103`, …) — không
đụng. `scripts/smoke.sh:41` chạy api với `-dsn` → đi qua `Postgres()` → **số vàng được kiểm
qua đúng đường đọc file**. Dev `-dsn ""` vẫn chạy không cần file. `go test ./...` vẫn không
cần thư mục ngoài package trừ đúng một test cột hai nguồn.
**Đã bỏ:** `go:embed` một file mặc định — `go:embed` không đọc được `../../config/`, buộc
dời file khỏi `config/` (trái tên nợ). Bỏ hẳn hằng — dev mode và `wire_test.go` phải tìm file
bằng đường dẫn tương đối theo package, mong manh; hai nguồn được **cột bằng test** tốt hơn
một nguồn mà test không đụng được.

---

## Đụng vào cái gì
- **aggregate / module:** `pricing`. Ba value object và một aggregate: `QuotePolicy`
  (`policy.go:65`, dựng ở `:72`), `MarginPolicy` (`:13`, `:18`), `RateCard` (`lane.go:80`,
  `:84`), `Classification` (`lane.go:287`, `:291`), aggregate `ShippingLane` (`lane.go:177`).
  Cửa vào hiện tại là composition root `internal/platform/wire`: `quotePolicy()`
  (`wire.go:278`), `goodsClasses()` (`:298`), `seedPricing()` (`:350`), gắn vào
  `pricingapp.Deps` ở `Postgres()` (`:225–226`) và `Memory()` (`:99–100`); seed chạy ở
  `:266` và `:146`. Task chỉ đổi **nguồn** của các giá trị ấy, không đổi kiểu hay cửa vào.
- **invariant:** mọi ràng buộc số đã nằm trong constructor domain và **ở lại đó**:
  `NewRateCard` bốn mức > 0, cùng tiền tệ (`lane.go:85–93`); `NewMarginPolicy` percent ≥ 0,
  sàn hợp lệ (`policy.go:19–24`); `NewQuotePolicy` thuế ≥ 0, có margin, 0 < cọc ≤ 100 %,
  ttl > 0 (`policy.go:73–82`); `NewShippingLane` có mã, có tên, divisor > 0, bước > 0, có
  rate card, phụ phí không âm và cùng tiền (`lane.go:178–197`); `NewClassification` class
  hợp lệ (`lane.go:291–299`). Loader mới **chỉ gọi** chúng và thêm tên trường vào lỗi.
  Invariant thứ hai, đã có sẵn: policy và lane được **chụp vào từng Quote** (`policy.go:63–64`)
  → đổi file rồi khởi động lại **không** đổi báo giá đã phát.
- **context / module khác bị chạm:**
  - `logistics`, qua event `pricing.lane_defined` (`subscribe.go:63` → `parcels.go:92`).
    Seed đi qua `DefineLaneHandler` (`define_lane.go:24–41`: dựng lane, `Save`, `Append`
    event) ở **mỗi** lần khởi động, nên đổi divisor/bước trong file = định nghĩa lại lane;
    logistics ghi đè `LaneRule` (`parcels.go:97–98`). Kiện đã tính cước không đổi.
  - **`cmd/worker` — điểm mới so với 23/09.** Worker cũng gọi `wire.Postgres`
    (`cmd/worker/main.go:48`), tức cũng dựng policy và cũng seed. Postgres đọc file (Q3/Q4)
    thì worker cũng cần đường dẫn → kéo theo `scripts/smoke.sh:41,43`,
    `scripts/smoke.ps1:26–27`, và CI chạy `smoke.sh` (`ci.yml:50–54`). Không phải câu hỏi
    mới — là hệ quả của đề nghị đã duyệt — nhưng `plan` phải liệt kê đủ các file này.
  - `wire_test.go:345` và `:386` gọi `Postgres` với chữ ký cũ; `:338`, `:400` gọi `Memory`.
  - Docs: `CLAUDE.md:137–138` (Còn nợ), `docs/SETUP.md:852` và `:1331`.

## Luật của project áp vào đây
- `CLAUDE.md:244` quy ước 1 — file sai là lỗi **người vận hành** → `error` (Q3). Riêng
  `Memory()` giữ `log.Fatalf` ở `wire.go:147`: fixture là dữ liệu của lập trình viên.
- `:245` quy ước 2 — số trong file là **chuỗi**, qua `shared.ParsePercent` (`rate.go:38`)
  và `shared.ParseMoney` (`money.go:102`); không `float64` ở bất kỳ đâu.
- `:247` quy ước 4 — đường đọc file dùng `New*`, không `Must*`.
- `:251` quy ước 8 — loader **không tự validate**; chỉ gọi `Parse*`/`New*` của domain.
- `:253` quy ước 10 — `Policy`, `Classification` vẫn vào `Deps`; `mustHave` giữ nguyên.
- `:255` luật vàng + allowlist `decisions_test.go:226–229` — thư viện yaml chỉ ở
  `internal/adapter/config`.
- `:43` "Trước khi bảo xong" và bài học 3 — smoke phải chạy lại, hai lần.
- "Số nghiệp vụ đã chốt" — giá trị trong file **bằng đúng** hằng hiện tại, không hơn kém.

## Điều ngoài luật, hoặc quyết định đã đóng bị chạm
- Không có. Quyết định 08/09 (một đơn = một variant = một cái) không liên quan tới task này.
- Comment `wire.go:277` (*"A wrong constant here is a programmer error, hence Must-style
  panics"*) sẽ **sai** sau task — phải đổi cùng code, không để lại.

## Câu trả lời của người
- 2026-09-23 — *"đồng ý hết 4 đề nghị"* → Q1 · Q2 · Q3 · Q4 giữ nguyên như đề nghị.

## Gợi ý kiến thức riêng của project
- Tỷ giá là **trạng thái** có endpoint riêng (`POST /fx`), không phải cấu hình; 26 000 ở
  `wire.go:378` chỉ là giá trị khởi động cho dev.
- Đổi bảng giá = khởi động lại process; báo giá và kiện đã tính **giữ số cũ** vì policy và
  lane được chụp vào Quote (`policy.go:63`).
- Hai process (`api`, `worker`) cùng dựng graph và cùng seed lane lúc khởi động
  (`cmd/api/main.go:77`, `cmd/worker/main.go:48`) — cấu hình nào cũng phải đến được cả hai.
