# `balance` trên bảng đọc đơn: một con số, một nguồn, không còn phép trừ trong JavaScript

Mọi dòng dưới đây mô tả một thay đổi trên nhánh. Chưa merge. Merge là của anh.

## What changed and why
Từ đợt 21, hai màn hình tự tính "phần còn lại" = tổng − cọc bằng `BigInt` trong `words.js`, vì đợt ấy
cấm đụng Go. Giờ `summaryView` của bảng đọc (`GET /orders`, `GET /me/orders`) có `balance`, tính lúc đọc
bằng đúng `shared.Money.Sub` từ hai cột đã nằm cùng dòng — không lưu, không migration, không có số thứ
ba phải giữ đồng bộ. `balanceOf` và ba chỗ gọi của nó biến mất: phép tính của domain không còn bản chép
ở màn hình, đúng quy tắc "mọi con số nói được nó ở đâu ra". Một dòng bảng đọc hỏng (hai tiền tệ khác
nhau, điều `PlaceOrder` không cho phép) cho ra trường **thiếu** thay vì panic; màn hình in "—". Test
"cùng một đơn, hai cửa" trong `wholeFlow` khẳng định số của bảng đọc bằng số của ordering. Làm đúng
`docs/UI-NEXT-PLAN.md` §B, không mở rộng sang §A/§C.

## Files
- `internal/adapter/http/summaries.go` — +13 −2
- `internal/adapter/http/summaries_test.go` — +22 (1 test)
- `internal/platform/wire/wire_test.go` — +18 (khẳng định trong `wholeFlow`, 2 struct)
- `web/app/words.js` — +1 −14 (xoá `balanceOf`) · `customer.js` — +2 −2 · `staff.js` — +3 −3
- `CLAUDE.md` — −1 dòng nợ · `docs/UI-NEXT-PLAN.md` — §B ghi đã làm · `docs/SETUP.md` — +12 (Đợt 23)

## Gate
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

## Rollback
Chưa có commit: `git checkout main && git branch -D agent/T-002-balance-field`. Sau commit: `git revert` một commit.

## Left for you
- **Bấm thử `staff.html` một lượt** (không có script trình duyệt trong repo): hàng đợi thu tiền và phiếu thu
  phải hiện đúng "phần còn lại 2.696.860 ₫" như trước — số giờ đến từ API. Nhìn giống hệt là **đúng**.
- Vòng này không có ảnh chụp: máy này không có puppeteer, và màn hình không đổi hình theo thiết kế.

## Xem thử
```
# 1. Hai test nói đúng điều task hứa (không cần Postgres)
go test -count=1 -run 'TestOrders_balanceIsTotalMinusDeposit|TestMemory_runsTheWholeFlow' -v ./internal/adapter/http ./internal/platform/wire
#   --- PASS: TestOrders_balanceIsTotalMinusDeposit      ← /me/orders và /orders đều có balance 2696860 VND
#   --- PASS: TestMemory_runsTheWholeFlow                ← bảng đọc == GET /orders/{id} cho cùng một đơn

# 2. Không còn phép trừ nào trong JS
grep -rn balanceOf web/app/          # (không có dòng nào)

# 3. Trên màn hình: chạy như mọi khi, đi hết luồng trong console.html rồi mở staff.html
go run ./cmd/api -web ./web          # http://localhost:8080/ui/console.html → 26 bước → staff.html, hàng đợi
                                     # "thu tiền": phần còn lại 2.696.860 ₫ — cùng số, giờ từ trường balance
```

## Compare
`agent/T-002-balance-field` so với `main` (613e5ca)

---

## Review — vì sao, không chỉ cái gì

### Idiom của ngôn ngữ project
- `summaries.go:71` `if balance, err := s.Total.Sub(s.Deposit); err == nil { … }` — nuốt lỗi có chủ ý, và
  comment ở `:43` nói vì sao (dòng hỏng → trường thiếu, thấy được). Đề nghị: giữ; nếu muốn chặt hơn thì
  một log ở projector khi chiếu `order_placed` với hai tiền tệ khác nhau, không phải ở view.
- `wire_test.go:297` tìm dòng theo `order_id` thay vì `[0]` — `wholeFlow` có hai đơn của cùng khách,
  lấy `[0]` sẽ xanh nhầm đơn. Đề nghị: giữ.
- `summaries_test.go:57` bảng hai route với `as` là hàm — cùng mẫu keyed fields như T-001. Đề nghị: giữ.

### Mô hình domain
- Không đụng aggregate. `Balance` ở view là **dẫn xuất** từ hai cột cùng dòng của read model, đúng như
  `CustomerOrder.Balance()` là dẫn xuất từ hai field của aggregate — hai bên cùng một phép, không bên nào
  lưu. Đề nghị: giữ; ghi vào `knowledge/` của project: "dẫn xuất trên bảng đọc tính lúc đọc trừ khi cần
  lọc/sắp theo nó".
- Không có trạng thái mới, không có event mới, không có migration. Đề nghị: giữ.

### Kỹ thuật
- Tương thích: trường **thêm**, không đổi trường cũ; client cũ không hỏng. `omitempty` chỉ vắng khi dòng
  hỏng. Đề nghị: giữ.
- JS: `money(undefined)` in "—" (`api.js:124`) nên chỗ gọi không cần guard. `journeyOf` không có test trong
  repo (đợt 21 kiểm bằng script ngoài) — `node --check` chỉ bắt cú pháp. Đề nghị: giữ ở task này; một
  file test `node --test` cho `words.js` là task nhỏ đáng làm (nợ mới).
- Chạy thật: smoke đi qua Postgres tạm xanh, `TestPostgres_runsTheWholeFlow` PASS có khẳng định mới.
  Đề nghị: giữ.

### Vượt plan
- Không có. 9 file đúng danh sách. `PROJECT.md` sửa một chữ `bash` trong `commands.smoke` là trạng thái
  agent, không phải thay đổi của task.

### Kết luận — đề nghị
**Đề nghị: commit.**
