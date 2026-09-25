# Bảng giá ra `config/ratecard.yaml`: api và worker đọc số nghiệp vụ từ file, fixture giữ cho dev/test

Mọi dòng dưới đây mô tả một thay đổi trên nhánh. Chưa merge. Merge là của anh.

## What changed and why
Nợ từ đợt 9: thuế, phí, cọc, hạn báo giá, bảng phân loại và lane của hãng bay là hằng trong
`wire.go`, đổi một số là build lại. Giờ `wire.Postgres` nhận đường dẫn file và nạp nó **trước** khi
kết nối database, qua package mới `internal/adapter/config`: yaml chỉ ở adapter, mọi số là chuỗi
đi qua `shared.Parse*`, mọi ràng buộc vẫn nằm ở constructor của `pricing`; adapter chỉ thêm **tên
trường** vào lỗi. File sai thì api và worker không khởi động và nói đúng trường (acceptance 3).
`wire.Memory` giữ hằng làm fixture để dev và test không cần file, và `TestFixtureMatchesTheRateCardFile`
cột hai nguồn lại: đổi một bên mà quên bên kia là đỏ (acceptance 4). Tỷ giá không vào file — nó là
trạng thái đổi hàng ngày qua `POST /fx`. Cờ `-ratecard` có mặc định `config/ratecard.yaml` nên
smoke, CI và lệnh trong docs không đổi. Cả bốn câu challenge (Q1–Q4 trong `analysis.md`) làm theo
đề nghị của agent, anh đã đồng ý hôm 23/09.

## Files
- `config/ratecard.yaml` — mới, 37 dòng
- `internal/adapter/config/ratecard.go` — mới, 262 dòng
- `internal/adapter/config/ratecard_test.go` — mới, 157 dòng (bảng 11 dòng + 3 test)
- `internal/platform/wire/wire.go` — +60 −37
- `internal/platform/wire/wire_test.go` — +66 −2 (2 test mới, 2 lời gọi đổi)
- `cmd/api/main.go` — +3 −1 · `cmd/worker/main.go` — +2 −1
- `go.mod` — +3 (yaml.v3 trực tiếp; 2 indirect là dep test của nó, không vào binary) · `go.sum` — +9
- `CLAUDE.md` — +5 −1 · `docs/SETUP.md` — +30 −10 (§7 trỏ file, bỏ dòng nợ, Đợt 22)

## Gate
## Verify 2026-09-25T15:19:26+07:00 — lần 3 — XANH

Kết quả ghi bằng script đọc exit code **sau** khi chạy (sửa lỗi của lần 2). `PORTAGE_TEST_DSN` bỏ
khỏi môi trường khi chạy `commands.test` (lý do ở lần 1, §5.4).

| lệnh | exit | thời gian | ghi chú |
|---|---|---|---|
| `gofmt -l .` | 0 | 0.1s | 0 file lệch ngoài learn/ |
| `go vet ./...` | 0 | 0.1s | 0 dòng output |
| `go test ./...` | 0 | 0.5s | 23 package ok · 0 dòng FAIL |
| build | n/a (no command) | | |

Đếm `-v`: **265** `--- PASS` (mốc trước task 259, +6) · 33 `--- SKIP` · 31 subtest PASS,
0 subtest FAIL · test mới PASS 6/6 · guard `TestDecision_*` 7/7.

Chưa kiểm:
- Acceptance 1, phần tích hợp Postgres — not run (33 SKIP; DSN của shell trỏ cổng bị chiếm) —
  chạy được khi có Postgres rảnh hoặc tạm.
- Acceptance 2 (smoke in số vàng) — not run (`scripts/smoke.sh` không có trong `PROJECT.commands`;
  cổng 5433 bị chiếm) — **Left for you**: thêm `commands.smoke` và cho một Postgres.
- Acceptance 3, 4, 5 — PASS bằng bảng 11 dòng, test thiếu file, fixture = file, 7 guard.
- Acceptance 6 — kiểm bằng mắt như lần 1, không đổi.

Gate đỏ hai lần trước khi xanh, cả hai là lỗi của agent trong **test**, không phải trong code:
(1) hàng "no lanes" tự dùng khoá lạ nên bị `KnownFields` bắt trước; (2) thêm trường vào struct bảng
mà 10 hàng cũ còn viết theo vị trí — và lần đó record ghi "passed" **trước** khi đọc exit code, đã
sửa lại record cho đúng. Cả ba lần verify có trong `TASK.md`.

## Rollback
Chưa có commit. Bỏ toàn bộ: `git checkout main && git branch -D agent/T-001-ratecard-yaml` rồi xoá
hai thư mục mới `config/` và `internal/adapter/config/`. Sau khi commit: `git revert <hash>` một commit.

## Left for you
- **Smoke chưa chạy.** `scripts/smoke.sh` không có trong `PROJECT.commands`, và cổng 5433 của máy
  này đang thuộc `rift-db-1`. Thêm `smoke: scripts/smoke.sh` vào `agent/projects/portage/PROJECT.md`
  và cho một Postgres (tạm cũng được, như đợt 20) — verify lần sau sẽ chạy và số vàng sẽ đi qua
  **đúng đường đọc file**. Cùng lý do, `TestPostgres_runsTheWholeFlow` mới SKIP, chưa PASS.
- **5 comment `[PHP]`** trong `ratecard.go` — quy ước của anh là tự bỏ trước khi commit.
- **Máy Windows có một `internal/adapter/config/` chưa commit** (docs/UI-REDESIGN-PLAN.md:303).
  Bản đó và bản này trùng tên package; giữ bản nào là của anh, bản này có hồ sơ đầy đủ.
- `docs/CODING-AGENT.md` thiếu DEC-024 (ADR-006), DEC-025 (ADR-007) — thêm ở lần commit trạng thái agent.

## Xem thử
Mọi lệnh chạy từ gốc repo, không cần Postgres, không cần Docker (`bin/` đã nằm trong .gitignore; đã
chạy thử, kết quả chép ở dưới). DSN dưới là địa chỉ giả tới cổng 1 — không gì lắng nghe — để chứng
minh file được kiểm **trước** khi chạm database:

```
go build -o bin/api ./cmd/api && go build -o bin/worker ./cmd/worker

# 1. Thiếu file → từ chối ngay, nêu tên file, KHÔNG chạm database
bin/api -dsn "postgres://127.0.0.1:1/none?sslmode=disable" -ratecard does-not-exist.yaml
#   portage api: ratecard does-not-exist.yaml: open does-not-exist.yaml: no such file or directory

# 2. Sai một trường (cọc 150 %) → nêu đúng trường và luật
sed 's/deposit: "50"/deposit: "150"/' config/ratecard.yaml > bin/bad.yaml
bin/api -dsn "postgres://127.0.0.1:1/none?sslmode=disable" -ratecard bin/bad.yaml
#   portage api: ratecard bin/bad.yaml: quote: deposit 150.0000%: want 0 < deposit ≤ 100%: invalid pricing policy

# 3. Gõ sai tên khoá → không âm thầm rơi về mặc định
sed 's/deposit: "50"/deposit_pct: "50"/' config/ratecard.yaml > bin/typo.yaml
bin/api -dsn "postgres://127.0.0.1:1/none?sslmode=disable" -ratecard bin/typo.yaml
#   ... yaml: unmarshal errors:  line 17: field deposit_pct not found in type config.quoteSection

# 4. Worker cùng luật
bin/worker -dsn "postgres://127.0.0.1:1/none?sslmode=disable" -ratecard does-not-exist.yaml

# 5. Mặc định hiện trong -h;  6. Chế độ memory không cần file, số vàng như cũ
bin/api -h 2>&1 | grep -A1 ratecard
go run ./cmd/api -web ./web        # rồi mở http://localhost:8080/ui/customer.html, gửi link → 5.393.720 ₫
```
Có Postgres: `go run ./cmd/api -dsn "$DSN"` không cần thêm gì (mặc định `config/ratecard.yaml`);
đổi `standard: "9.00"` thành `"90.00"` rồi khởi động lại, báo giá **mới** đổi, báo giá **đã phát** giữ số cũ.

## Compare
`agent/T-001-ratecard-yaml` so với `main` (42ac518)

---

## Review — vì sao, không chỉ cái gì

### Idiom của ngôn ngữ project
- `ratecard.go:131` `field()` bọc lỗi bằng `%w` nên `errors.Is(err, pricing.ErrInvalidPolicy)` vẫn
  đúng qua ba lớp bọc — đúng quy ước `%w` của repo. Đề nghị: giữ.
- `ratecard.go:243` dựng `ShippingLane` chỉ để validate rồi bỏ; seed dựng lại lần nữa qua
  `DefineLane`. Hai lần dựng, một lần thừa — nhưng nhờ đó lỗi lane hiện **kèm tên file** lúc nạp,
  thay vì lúc seed. Đề nghị: giữ, vì thông báo cho người vận hành đáng hơn một lần dựng value object.
- `ratecard.go:104` trả lỗi yaml nguyên: người vận hành thấy `config.quoteSection` — tên kiểu Go lộ
  ra ngoài. Đề nghị: giữ ở task này; nếu có người vận hành thật ngoài anh, ánh xạ về đường dẫn yaml
  là một task nhỏ riêng.
- `ratecard_test.go:95` bảng 11 hàng viết theo tên trường — mẫu table-driven đầu tiên của repo
  (`learn/PLAN-BO-SUNG.md` nợ mẫu này). Hàng theo vị trí đã vỡ một lần khi thêm trường: lý do đúng để
  luôn viết theo tên. Đề nghị: giữ.
- `wire.go:368` `fixtureLane()` dùng `Must*` và `panic` — đúng quy ước 1 vì là dữ liệu của lập trình
  viên; comment ở `wire.go:284` nói rõ vì sao vẫn panic dù production giờ trả error. Đề nghị: giữ.

### Mô hình domain
- Không aggregate hay invariant nào bị lách: năm constructor của `pricing` vẫn là nơi duy nhất nói
  "số này hợp lệ"; adapter không có `if` nào về nghiệp vụ. 7 guard `decisions_test.go` xanh → luật
  vàng còn nguyên. Đề nghị: giữ.
- `lanes` là danh sách; báo giá chọn lane theo mã (`internal/app/pricing/issue_quote.go:48`), nên
  thêm lane thứ hai vào file không đổi hành vi báo giá hiện tại. Đề nghị: giữ, ghi nhận là tiền đề cho
  màn hình bảng giá cước (D4 trong nợ UI).
- Ranh giới cấu hình ↔ trạng thái đặt đúng: tỷ giá ở ngoài file. Đề nghị: giữ.

### Kỹ thuật
- Đường lỗi: `Load` → `Parse` → `field` → `pricing.New*`; mọi lỗi mang tên file + tên trường; api
  và worker `log.Fatalf` như DSN. Đã chạy thử bốn trường hợp ở mục *Xem thử*. Đề nghị: giữ.
- `secret.scan` đánh dấu một chuỗi: `wire_test.go:35` `unreachableDSN` = `nobody:nothing` tới cổng 1.
  Là DSN giả có sẵn trong repo từ trước (dòng cũ `:386`), giờ gom một chỗ. Không phải credential.
  Đề nghị: giữ.
- Tương thích: chữ ký `wire.Postgres` đổi, hai caller trong repo đã đổi, không caller ngoài. Cờ mới
  có mặc định nên mọi lệnh cũ chạy như cũ từ gốc repo; chạy binary từ thư mục khác thì phải truyền
  `-ratecard` — thông báo lỗi nói đúng điều đó. Đề nghị: giữ.
- Chạy thật: hai process cùng `DefineLane` lúc khởi động — đã như vậy từ trước, không mới. Điều mới
  là **đổi file rồi khởi động lại = định nghĩa lại lane**; logistics nghe `lane_defined` và cập nhật.
  Đề nghị: ghi vào `projects/portage/knowledge/` (gợi ý ở `analysis.md`), anh chốt.

### Vượt plan
- Không có. 10 file đúng danh sách đóng; `agent/` và `.claude/skills/agent/` là trạng thái của agent.

### Kết luận — đề nghị
**Đề nghị: commit.** Sau khi anh bỏ 5 comment `[PHP]` nếu muốn bỏ trước commit — hoặc để agent bỏ
qua một dòng "sửa X".
