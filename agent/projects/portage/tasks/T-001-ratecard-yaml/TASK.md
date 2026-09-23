---
id:          T-001
project:     portage
slug:        ratecard-yaml
status:      IN_PROGRESS
step:        analyze
verify:      not-run
challenge:   true
mode:        employee
branch:      agent/T-001-ratecard-yaml
created:     2026-09-23T12:13:00+07:00
approvals:   []
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
