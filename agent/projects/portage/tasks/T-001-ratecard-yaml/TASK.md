---
id:          T-001
project:     portage
slug:        ratecard-yaml
status:      DONE
verify:      passed
plan_digest: be0bb8cf6e85a4f40dfcb9afe90c1f1c7d2df97ea3a0cd44265350c4e2cd5c5e
challenge:   true
mode:        employee
branch:      agent/T-001-ratecard-yaml
created:     2026-09-23T12:13:00+07:00
approvals:
  - {gate: plan, by: Sy Duong, at: 2026-09-25T14:14:31+07:00, note: ~}
  - {gate: plan, by: release, at: 2026-09-25T15:15:41+07:00, note: "sửa sau gate hỏng — released since 2026-09-25 (ADR-007)"}
  - {gate: plan, by: release, at: 2026-09-25T15:18:37+07:00, note: "sửa lần 2 sau gate hỏng — released since 2026-09-25 (ADR-007)"}
  - {gate: commit, by: Sy Duong, at: 2026-09-25T15:32:31+07:00, note: ~}
  - {gate: push, by: Sy Duong, at: 2026-09-25T15:41:20+07:00, note: ~}
docs_read:
  - {path: CLAUDE.md, at: 2026-09-23}
  - {path: docs/SETUP.md, at: 2026-09-23}
  - {path: docs/UI-GUIDE.md, at: 2026-09-23}
  - {path: internal/platform/wire/wire.go, at: 2026-09-23}
  - {path: internal/platform/wire/wire_test.go, at: 2026-09-23}
  - {path: internal/domain/pricing/policy.go, at: 2026-09-23}
  - {path: internal/domain/pricing/lane.go, at: 2026-09-23}
  - {path: internal/domain/decisions_test.go, at: 2026-09-23}
  - {path: cmd/api/main.go, at: 2026-09-23}
  - {path: scripts/smoke.sh, at: 2026-09-23}
  - {path: go.mod, at: 2026-09-23}
  - {path: cmd/worker/main.go, at: 2026-09-25}
  - {path: internal/app/pricing/define_lane.go, at: 2026-09-25}
  - {path: internal/app/logistics/parcels.go, at: 2026-09-25}
  - {path: internal/platform/wire/subscribe.go, at: 2026-09-25}
  - {path: internal/domain/shared/rate.go, at: 2026-09-25}
  - {path: internal/domain/shared/money.go, at: 2026-09-25}
  - {path: .github/workflows/ci.yml, at: 2026-09-25}
  - {path: scripts/smoke.ps1, at: 2026-09-25}
  - {path: docs/UI-REDESIGN-PLAN.md, at: 2026-09-25}
  - {path: docs/UI-NEXT-PLAN.md, at: 2026-09-25}
  - {path: internal/domain/shared/weight.go, at: 2026-09-25}
  - {path: internal/domain/shared/currency.go, at: 2026-09-25}
---

## Goal
Đưa các số nghiệp vụ của báo giá và cước — hiện là hằng trong
`internal/platform/wire/wire.go` — ra file cấu hình `config/ratecard.yaml`, để người vận
hành đổi bảng giá mà không sửa code, không build lại.

## Context
Nợ ghi từ đợt 9: `CLAUDE.md` mục "Còn nợ" và `docs/SETUP.md` mục "Còn nợ trong code"
(dòng 852) và "Còn nợ sau P9" (dòng 1331). Hằng nằm ở `wire.go:279–286` (`quotePolicy`:
phí 10 % sàn 500 000 ₫, thuế 8,81 %, cọc 50 %, hạn 48 h), `:297–303` (`goodsClasses`),
`:354–366` (lane `us_forwarder`: divisor 5000, bước 500 g, bốn mức cước 9/10/12/14 USD,
phụ phí pin 3 USD), `:378` (tỷ giá seed 26 000). Chưa có thư mục `config/`.

## Acceptance
- [ ] `go test ./...` xanh, số test không giảm so với trước task *(đề nghị)*
- [ ] `scripts/smoke.sh` in đúng số vàng: `total 5393720 VND, deposit 2696860` và
      `variance -2.50 USD` *(đề nghị)*
- [ ] Api khởi động với file cấu hình thì dùng số trong file; file thiếu, không parse
      được, hoặc số sai (âm, cọc > 100 %, thiếu một mức cước) → api **không** khởi động và
      in một dòng lỗi nói rõ **trường nào** sai *(đề nghị)*
- [ ] Không số nghiệp vụ nào đổi giá trị — bảng "Số nghiệp vụ đã chốt" trong `CLAUDE.md`
      *(đề nghị)*
- [ ] `internal/domain/` không import thêm gì ngoài allowlist; 7 guard trong
      `internal/domain/decisions_test.go` vẫn xanh *(đề nghị)*
- [ ] `CLAUDE.md` "Còn nợ" và hai dòng nợ trong `docs/SETUP.md` được gỡ; số trong docs khớp
      code *(đề nghị)*

## Constraints
- `CLAUDE.md` "Quy ước code": luật 1 (input lập trình viên sai → panic; input người dùng
  sai → error), luật 7 (`Clock` là port), luật 10 (`Deps` + `mustHave`).
- `CLAUDE.md` "Luật vàng": `internal/domain/` chỉ stdlib + allowlist.
- `CLAUDE.md` "Số nghiệp vụ đã chốt": không đổi. Quyết định 08/09 (một đơn = một variant =
  một cái): không mở lại.
- `CLAUDE.md` "Ràng buộc pháp lý": không liên quan tới task này, ghi để khẳng định đã đọc.

## Learning
- DDD: ranh giới **cấu hình** (đọc ở adapter, đổi theo môi trường) và **value object**
  (`QuotePolicy`, `RateCard` — bất biến, được chụp vào từng Quote). Domain không biết file.
- Go: `flag` vs biến môi trường vs `go:embed`; parse cấu hình có xác thực ở biên — đối chiếu
  Symfony `config/packages/*.yaml` + `Configuration::getConfigTreeBuilder()`.

## Corrections
*(chỉ người viết)*

- 2026-09-23 — Sy Duong, trong chat: *"đồng ý hết 4 đề nghị"* → Q1 · Q2 · Q3 · Q4 của
  `analysis.md` giữ nguyên như đề nghị. (agent ghi hộ qua `/agent answer`)

## Verify 2026-09-25T14:36:04+07:00

Chạy trên nhánh `agent/T-001-ratecard-yaml`, máy Linux, shell có sẵn `PORTAGE_TEST_DSN`.

| lệnh | exit | thời gian | ghi chú |
|---|---|---|---|
| `gofmt -l .` | 0 | 0.1s | 0 file lệch |
| `go vet ./...` | 0 | 0.1s | 0 dòng output |
| `go test ./...` | **1** | 0.6s | 20 package ok · 3 FAIL: `adapter/config`, `adapter/postgres`, `platform/wire` |
| build | n/a (no command) | | |

Dòng lỗi đầu tiên: `--- FAIL: TestParse_rejectsOneWrongFieldByName (0.00s)` (package `adapter/config`).

**Kiểm lại theo §5.4** — cùng lệnh, bỏ `PORTAGE_TEST_DSN` khỏi môi trường: exit 1, 0.4s,
22 package ok · **1 FAIL** (`adapter/config`, cùng test trên, subtest `no_lanes`). Hai package
`postgres` và `wire` đỏ ở lần đầu là do **môi trường**: dòng lỗi giữ lại (có chuỗi kết nối);
nội dung là xác thực thất bại trên `127.0.0.1:5433`, cổng đang thuộc container `rift-db-1`
của dự án khác (SETUP §9 đợt 20). Không phải lỗi của thay đổi này.

Đếm không DSN, `-v`: **264** `--- PASS` (mốc trước task 259, +5: sáu test mới trừ một đỏ) ·
33 `--- SKIP` (test tích hợp Postgres) · 30 subtest PASS. Test mới PASS:
`TestParse_readsEveryValueExactly`, `TestParse_unquotedNumbersStayExact`,
`TestLoad_namesTheFileOnEveryError`, `TestPostgres_refusesAMissingRateCardBeforeConnecting`,
`TestFixtureMatchesTheRateCardFile`. 7 guard `TestDecision_*` PASS.

Chưa kiểm:
- Acceptance 1 (`go test` xanh, không giảm) — **đỏ 1 test** (xem đề nghị sửa trong plan.md); phần
  tích hợp Postgres — not run (33 SKIP; DSN của shell trỏ cổng bị chiếm) — chạy được khi có
  Postgres rảnh, hoặc Postgres tạm như đợt 20.
- Acceptance 2 (smoke in số vàng) — not run (`scripts/smoke.sh` không có trong `PROJECT.commands`,
  và cổng 5433 bị chiếm) — chạy được khi người thêm `commands.smoke` vào PROJECT.md **và** có Postgres.
- Acceptance 3 (file sai → không khởi động, lỗi nêu trường) — một nửa: test thiếu file PASS,
  bảng 11 dòng đỏ 1 dòng (lỗi của **dòng test**, không phải của code — xem plan).
- Acceptance 4 (không đổi giá trị) — PASS bằng `TestFixtureMatchesTheRateCardFile` và
  `TestParse_readsEveryValueExactly`.
- Acceptance 5 (domain không import thêm) — PASS, 7 guard.
- Acceptance 6 (docs gỡ nợ, số khớp) — kiểm bằng mắt: `CLAUDE.md` và `SETUP.md:852` không còn dòng
  nợ; `SETUP.md:1331` là lịch sử, giữ như plan.

## Verify 2026-09-25T15:18:37+07:00 — lần 2, sau sửa test — GATE ĐỎ

*Sửa lại: bản đầu của mục này ghi "passed" vì script ghi kết quả chạy trước khi đọc exit code —
lỗi của agent, không phải của code. Đây là bản đúng.*

`PORTAGE_TEST_DSN` bỏ khỏi môi trường khi chạy `commands.test` (lý do ở lần 1, §5.4).

| lệnh | exit | thời gian | ghi chú |
|---|---|---|---|
| `gofmt -l .` | 0 | 0.3s | 0 file lệch |
| `go vet ./...` | **1** | 0.3s | package `adapter/config_test` không biên dịch |
| `go test ./...` | **1** | 0.4s | 22 package ok · 1 build failed (`adapter/config`) |
| build | n/a (no command) | | |

Dòng lỗi đầu tiên: `internal/adapter/config/ratecard_test.go:103:68: too few values in struct
literal of type struct{name string; from string; to string; file string; want string}`.

Chưa kiểm: như lần 1. Bảng 11 dòng chưa chạy được (không biên dịch); 261 `--- PASS` là các test
khác, không đủ để nói gì về acceptance 3.

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

## Verify sau DONE 2026-09-25T15:47:58+07:00 — Postgres tạm, theo lệnh người ("làm hết đi")

Môi trường: Postgres 16 tạm bằng `docker run` trên cổng 5436 (5433 vẫn thuộc dự án khác), database
`portage` và `portage_test`; máy không có `psql` nên `scripts/smoke.sh` chạy với một shim `psql`
trỏ vào container. Giá trị DSN không ghi ở đây (R4).

| lệnh | exit | thời gian | ghi chú |
|---|---|---|---|
| `go test ./...` với `PORTAGE_TEST_DSN` | 0 | 10.7s | **297** `--- PASS` · 1 SKIP (cố ý) · 0 FAIL · `TestPostgres_runsTheWholeFlow` PASS |
| cùng lệnh, `-count=1` (không cache) | 0 | 12.8s | 297 PASS · 1 SKIP · 0 FAIL — lần thứ hai thật, không phải cache |
| `bash scripts/smoke.sh` lần 1 | 0 | 7.7s | `total 5393720 VND, deposit 2696860` · `variance -2.50 USD` |
| `bash scripts/smoke.sh` lần 2 | 0 | 6.8s | cùng số vàng |

Api trong smoke khởi động không có cờ `-ratecard`, tức đọc `config/ratecard.yaml` mặc định từ gốc
repo — số vàng đi qua **đúng đường đọc file**. Acceptance 1 và 2 giờ đã có bằng chứng; T-001 không
còn "Chưa kiểm" nào ngoài cổng 5433 thật (nợ cũ của máy, không phải của task).

Hai điều nhỏ lộ ra: `scripts/smoke.sh` không có bit thực thi trong repo (CI `chmod +x` trước khi
chạy; máy này gọi qua `bash`), và `go test` lần hai phải `-count=1` mới là lần chạy thật.
