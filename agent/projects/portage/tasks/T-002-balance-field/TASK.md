---
id:          T-002
project:     portage
slug:        balance-field
status:      DONE
verify:      passed
plan_digest: 34f3f8de87544e7d428f631d47e27f3718aa83b9ac3563788cffd016fa8f5779
challenge:   false
mode:        employee
branch:      agent/T-002-balance-field
created:     2026-09-25T15:51:01+07:00
approvals:
  - {gate: plan, by: release, at: 2026-09-25T15:52:51+07:00, note: "released since 2026-09-25 (ADR-007)"}
  - {gate: commit, by: Sy Duong, at: 2026-09-25T17:00:32+07:00, note: ~}
  - {gate: push, by: Sy Duong, at: 2026-09-25T17:01:14+07:00, note: ~}
docs_read:
  - {path: CLAUDE.md, at: 2026-09-25}
  - {path: docs/UI-NEXT-PLAN.md, at: 2026-09-25}
  - {path: docs/SETUP.md, at: 2026-09-25}
  - {path: docs/WALKTHROUGH.md, at: 2026-09-25}
  - {path: internal/app/reporting/summary.go, at: 2026-09-25}
  - {path: internal/app/reporting/projector.go, at: 2026-09-25}
  - {path: internal/adapter/http/summaries.go, at: 2026-09-25}
  - {path: internal/adapter/http/summaries_test.go, at: 2026-09-25}
  - {path: internal/adapter/http/orders.go, at: 2026-09-25}
  - {path: internal/adapter/http/quotes.go, at: 2026-09-25}
  - {path: internal/domain/ordering/order.go, at: 2026-09-25}
  - {path: internal/domain/shared/money.go, at: 2026-09-25}
  - {path: internal/platform/wire/wire_test.go, at: 2026-09-25}
  - {path: web/app/words.js, at: 2026-09-25}
  - {path: web/app/customer.js, at: 2026-09-25}
  - {path: web/app/staff.js, at: 2026-09-25}
  - {path: web/app/api.js, at: 2026-09-25}
---

## Goal
Bảng đọc đơn (`GET /orders`, `GET /me/orders`) trả thêm trường `balance` = tổng − cọc, để hai màn
hình không còn tự chép phép tính của domain sang JavaScript.

## Context
Nợ ghi ở `CLAUDE.md:146` ("Còn nợ") từ đợt 21, khi màn hình đọc `o.balance` mà `summaryView` không có
trường đó; đợt ấy cấm đụng Go nên sửa tạm bằng `balanceOf()` trong `web/app/words.js:208`. Thiết kế
đã có ở `docs/UI-NEXT-PLAN.md` §B: `balance = Total − Deposit` tính ở adapter bằng `shared.Money.Sub`,
xoá `balanceOf` và các chỗ gọi, test số trong JSON bằng `CustomerOrder.Balance()`. `GET /orders/{id}`
(view của ordering) **đã có** `balance` (`internal/adapter/http/orders.go:186`); thiếu là ở bảng đọc
`reporting` (`summaries.go:22–51`).

## Acceptance
- [ ] `GET /me/orders` và `GET /orders` có `balance: {amount, currency}` cho mọi dòng; với đơn mẫu
      5 393 720 / 2 696 860 thì `balance.amount == "2696860"` *(đề nghị)*
- [ ] Trong luồng đầu-cuối, `balance` của bảng đọc bằng đúng `balance` của `GET /orders/{id}` cho cùng
      một đơn — tức bằng `CustomerOrder.Balance()` *(đề nghị)*
- [ ] `balanceOf` không còn trong `web/app/`; ba chỗ gọi dùng `o.balance`; `node --check` sạch cho
      mọi module bị đụng *(đề nghị)*
- [ ] `go test ./...` xanh, số test không giảm (mốc 265 `--- PASS` không DSN); smoke đúng số vàng *(đề nghị)*
- [ ] Không migration, không đổi bảng `order_summaries`: `balance` là giá trị **dẫn xuất**, không lưu *(đề nghị)*
- [ ] `CLAUDE.md` "Còn nợ" bỏ dòng `balance`; `UI-NEXT-PLAN.md` §B ghi đã làm *(đề nghị)*

## Constraints
- `CLAUDE.md` quy ước 2 (tiền là `int64`, không float), 8 (adapter gọi domain, không tự tính lại luật).
- Hai quy tắc màn hình (`CLAUDE.md` mục "hai màn hình theo vai"): không chữ máy ra UI, mọi con số nói được
  nó ở đâu ra.
- Quyết định 08/09: không mở lại.

## Learning
- DDD: giá trị **dẫn xuất** trên read model — tính lúc đọc hay lưu lúc chiếu? Vì sao ở đây tính lúc đọc.
- Go: `(Money, error)` trong một view — lỗi không thể xảy ra thì xử lý thế nào cho không panic mà cũng
  không im lặng.

## Corrections
*(chỉ người viết)*

## Verify 2026-09-25T15:54:28+07:00 — XANH

Môi trường: Postgres tạm cổng 5436 (T-001 để lại), `PORTAGE_TEST_DSN` và `PORTAGE_DSN` trỏ vào đó (giá trị
không ghi), `PORTAGE_PORT=18080`, shim `psql` cho smoke. Verdict ghi bằng script đọc exit code sau khi chạy.

| lệnh | exit | thời gian | ghi chú |
|---|---|---|---|
| `gofmt -l .` | 0 | 0.1s | 0 file lệch ngoài learn/ |
| `go vet ./...` | 0 | 0.1s | 0 dòng output |
| `go test ./...` (-count=1, có DSN) | 0 | 10.0s | 23 package ok · 0 dòng FAIL |
| `bash scripts/smoke.sh` | 0 | 7.5s | total 5393720 VND, deposit 2696860 variance -2.50 USD |
| build | n/a (no command) | | |

Đếm `-v` (có DSN): **298** `--- PASS` · 1 SKIP · 0 FAIL · `TestOrders_balanceIsTotalMinusDeposit` 1/1 ·
--- PASS: TestMemory_runsTheWholeFlow --- PASS: TestPostgres_runsTheWholeFlow. Mốc không DSN sau T-001 là 265; có DSN sau T-001 là 297 → kỳ vọng 298.

Chưa kiểm:
- Acceptance 3, phần `node --check` — đã chạy ở implement (3 module sạch), verify không có lệnh JS trong
  `PROJECT.commands` → ghi lại từ implement, không đo lại.
- Màn hình bằng mắt: không có script trình duyệt trong repo; số hiện ra **giống hệt trước** theo thiết kế
  (chỉ đổi nguồn của số). Left for you: bấm `staff.html` một lượt.
