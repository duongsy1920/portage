# plan — T-002 balance-field

## Mục tiêu
Bảng đọc đơn (`GET /orders`, `GET /me/orders`) trả thêm `balance` = tổng − cọc, để hai màn hình không
còn tự chép phép tính của domain sang JavaScript.

## Bước
1. `internal/adapter/http/summaries.go`: thêm `Balance *moneyView` (`json:"balance,omitempty"`) sau
   `Deposit`; trong `summaryViewOf`, `s.Total.Sub(s.Deposit)` → có thì gán, lỗi (chỉ khi dòng hỏng)
   thì để trống. Comment nói vì sao dẫn xuất và vì sao con trỏ.
2. `internal/adapter/http/summaries_test.go`: `summaryJSON` thêm `Balance *money`; test mới
   `TestOrders_balanceIsTotalMinusDeposit`: seed 5 393 720 / 2 696 860 → cả `/me/orders` lẫn
   `/orders` có `balance.amount == "2696860"`.
3. `internal/platform/wire/wire_test.go`: `ov` và `summaryRow` thêm `Balance`; sau khi đơn
   `in_transit`, đọc `/me/orders`, dòng của đơn ấy có `balance` bằng `ov.Balance` của `GET /orders/{id}`
   — đó là `CustomerOrder.Balance()` cho cùng một đơn (acceptance 2).
4. `web/app/words.js`: xoá `balanceOf` và comment của nó; `journeyOf` dùng `o.balance`.
   `web/app/customer.js`, `web/app/staff.js`: bỏ import, ba chỗ gọi dùng `o.balance`.
   `node --check` cho ba file.
5. Docs: `CLAUDE.md` bỏ dòng nợ `balance`; `docs/UI-NEXT-PLAN.md` §B ghi "đã làm (T-002)";
   `docs/SETUP.md` thêm Đợt 23 ngắn.
6. `gofmt -l .` sạch. Test và smoke là của `verify`.

## File sẽ đụng — danh sách ĐÓNG
- `internal/adapter/http/summaries.go` — trường + phép trừ
- `internal/adapter/http/summaries_test.go` — struct đọc JSON + 1 test
- `internal/platform/wire/wire_test.go` — hai struct + một khẳng định trong `wholeFlow`
- `web/app/words.js` — xoá `balanceOf`, `journeyOf` đọc `o.balance`
- `web/app/customer.js` — import + 1 chỗ gọi
- `web/app/staff.js` — import + 2 chỗ gọi
- `CLAUDE.md` — "Còn nợ"
- `docs/UI-NEXT-PLAN.md` — §B
- `docs/SETUP.md` — Đợt 23
(implement không được đụng file ngoài danh sách này. Cần thêm → về hỏi.)

## Test sẽ thêm hoặc đổi
- `TestOrders_balanceIsTotalMinusDeposit` (mới, http) — acceptance 1.
- `wholeFlow` (đổi) — acceptance 2: số của bảng đọc bằng số của ordering cho cùng đơn.
- Mốc: 265 `--- PASS` không DSN (đo 25/09 sau T-001); kỳ vọng 266.

## Không làm
- **Không** lưu `balance` vào `order_summaries`, **không** migration 0014: dẫn xuất từ hai cột cùng dòng.
- **Không** làm §A và §C của UI-NEXT-PLAN (trí nhớ phiên về bảng đọc, hai route đọc) — task khác.
- **Không** đụng `GET /orders/{id}` — đã có `balance`.
- **Không** đụng `web/flow.html`/`portage.js` (mô phỏng, có `money()` riêng, không gọi API).
- **Không** giữ `balanceOf` làm fallback: hai nguồn cho một con số là đúng thứ task này xoá.

## Tự phản biện
- **Cách hiển nhiên hơn:** projector ghi `balance_minor` lúc chiếu `order_placed`, view đọc cột.
  Không chọn vì: thêm migration + một cột phải giữ đồng bộ, đổi lại không có truy vấn nào lọc/sắp
  theo balance. Tính lúc đọc rẻ hơn và không có trạng thái để lệch.
- **Cách khác:** giữ `balanceOf` ở JS và chỉ thêm trường ở API. Không chọn vì: UI-NEXT-PLAN §B đã
  chốt xoá, và một con số hai nguồn là lỗi đợi ngày xảy ra.
- **Chỗ có thể sai:** (1) `journeyOf` được kiểm bằng script `node` ngoài repo ở đợt 21 — không có
  test JS trong repo, `node --check` chỉ bắt lỗi cú pháp; (2) `wholeFlow` có thể có hơn một đơn của
  cùng khách → phải tìm dòng theo `order_id`, không lấy `[0]`.
- **Đề nghị: làm theo kế hoạch này.**

## Rủi ro
- Smoke chạy được (Postgres tạm còn sống, `commands.smoke` đã khai) — nếu container đã tắt, verify ghi
  `not run` kèm lý do.

## Acceptance (chép từ TASK.md)
- [ ] `GET /me/orders` và `GET /orders` có `balance`; đơn mẫu → `"2696860"`
- [ ] Luồng đầu-cuối: `balance` bảng đọc == `balance` của `GET /orders/{id}`
- [ ] `balanceOf` không còn trong `web/app/`; ba chỗ gọi dùng `o.balance`; `node --check` sạch
- [ ] `go test ./...` xanh, ≥ 265 `--- PASS`; smoke đúng số vàng
- [ ] Không migration, không đổi bảng
- [ ] `CLAUDE.md` bỏ dòng nợ; `UI-NEXT-PLAN.md` §B ghi đã làm
