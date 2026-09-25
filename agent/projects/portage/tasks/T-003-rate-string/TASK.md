---
id:          T-003
project:     portage
slug:        rate-string
status:      IN_PROGRESS
step:        implement
verify:      not-run
plan_digest: 9272ddd49bfd0c3f7388ce9b88be27c5397475672a1d21853b9fc2acd5446db8
challenge:   false
mode:        learning
branch:      agent/T-003-rate-string
created:     2026-09-25T15:58:21+07:00
approvals:
  - {gate: plan, by: release, at: 2026-09-25T15:58:21+07:00, note: "released since 2026-09-25 (ADR-007); mode learning — người tự implement"}
docs_read:
  - {path: CLAUDE.md, at: 2026-09-25}
  - {path: internal/domain/shared/rate.go, at: 2026-09-25}
  - {path: internal/domain/shared/rate_test.go, at: 2026-09-25}
---

## Goal
`shared.Rate.String()` in cho **người** đọc: `8.81%`, `30%`, `-2.5%`, `0.0001%` — bỏ số 0 thừa ở phần
thập phân, thay vì `8.8100%` / `30.0000%` như hôm nay. **Anh tự viết code** (`mode: learning`); agent
phân tích, gợi ý, và review code của anh.

## Context
Nợ học ghi từ T-001 (`agent/learning/debt.md`): thông báo lỗi của `config/ratecard.yaml` in
`deposit 150.0000%: want 0 < deposit ≤ 100%` — người vận hành đọc thấy lạ. `Rate.String()` ở
`internal/domain/shared/rate.go:65–72` in cố định bốn chữ số thập phân vì ppm = phần trăm × 10⁴.
Bảng test ở `rate_test.go:16–20` khẳng định đúng dạng cũ nên sẽ đỏ trước, xanh sau — đó là bài.

## Acceptance
- [ ] `Rate.String()`: `8.81%`, `30%`, `0.0001%`, `-2.5%`, `108.81%` (Plus1 của 8.81) — bảng test ở
      `rate_test.go` đổi kỳ vọng và thêm dòng `Plus1` *(đề nghị)*
- [ ] Không đi qua `float64`: vẫn số nguyên + chuỗi (quy ước 2) *(đề nghị)*
- [ ] `go test ./...` xanh, không test nào khác đổi kỳ vọng (không chỗ nào ngoài `rate_test.go` khẳng định
      chuỗi `0000%` — đã kiểm bằng grep 25/09) *(đề nghị)*
- [ ] Thông báo lỗi của `config.Parse` với cọc 150 giờ in `deposit 150%` *(đề nghị)*

## Constraints
- `CLAUDE.md` quy ước 2 (không float), 9 (zero value an toàn: `Rate{}` phải in `0%`).
- `Rate.String()` chỉ dùng trong thông báo lỗi và test (đã kiểm: không view HTTP hay cột DB nào ghi chuỗi
  này; `pricing_repos.go:57` ghi `PPM()`). Đổi định dạng không đổi API hay dữ liệu.

## Learning
- Go: `Stringer`, `fmt.Sprintf` với `%d`/`%04d`, `strings.TrimRight` — và vì sao KHÔNG dùng `%g`/float.
- Go: sửa một bảng test có kỳ vọng cũ — đỏ trước, xanh sau; thêm một dòng cho `Plus1`.
- DDD: value object đổi cách **hiển thị** mà không đổi **giá trị** — `PPM()` là sự thật, `String()` là lời.

## Corrections
*(chỉ người viết)*
