# analysis — T-002 balance-field

*Run 2026-09-25. `challenge: false`: cách làm đã chốt ở `docs/UI-NEXT-PLAN.md` §B (23/09), đây là
kiểm lại §B với code hôm nay rồi kết luận.*

## Đụng vào cái gì
- **aggregate / module:** không đụng aggregate nào. Việc nằm ở **tầng đọc**: `reportingapp.OrderSummary`
  (`internal/app/reporting/summary.go:53–69`, có `Total`, `Deposit`) và view HTTP `summaryView`
  (`internal/adapter/http/summaries.go:22–51`, dựng ở `summaryViewOf` `:53–79`). `viewOf`
  (`quotes.go:69`) đổi `shared.Money` thành `{amount, currency}`.
- **invariant:** `CustomerOrder.Balance()` = `total.Sub(deposit)` (`ordering/order.go:196–199`), với
  `deposit ≤ total` do `PlaceOrder` bảo đảm. Bảng đọc nhận `Total`/`Deposit` từ event `order_placed`
  (`reporting/projector.go:103–122`) nên cùng bất biến. `shared.Money.Sub` (`money.go:215–220`) chỉ
  lỗi khi khác tiền tệ — trên một dòng bảng đọc hợp lệ, không xảy ra.
- **context / module khác bị chạm:** hai màn hình `web/app/customer.js:310`, `staff.js:273,624`, và
  `words.js:208` (`balanceOf`) + `:246` (`journeyOf` dùng nó). `GET /orders/{id}` của ordering
  (`orders.go:186`) đã có `balance` — đó là "cùng một đơn" để so trong luồng đầu-cuối
  (`wire_test.go:288–296`: có sẵn `ov` với `Total`, chỉ cần thêm `Balance` và một lần `myOrders`).

## Luật của project áp vào đây
- `CLAUDE.md` quy ước 2: tính bằng `shared.Money.Sub` (int64), không float — cả ở Go lẫn khi bỏ `BigInt` ở JS.
- Quy ước 8: adapter **gọi** `Money.Sub`, không viết lại phép trừ.
- "Mọi con số nói được nó ở đâu ra": số hiện trên màn hình giờ đến từ API, không từ một phép tính trong JS.
- `UI-NEXT-PLAN.md` §B là quyết định đã có; task này làm đúng §B, không mở rộng sang §A/§C.

## Điều ngoài luật, hoặc quyết định đã đóng bị chạm
- Không có. Tính lúc **đọc**, không lưu: `balance` là dẫn xuất từ hai cột đã có; lưu thêm là một
  migration (0014) và một chỗ nữa phải giữ đồng bộ, đổi lại không được gì (đã bỏ).
- `Sub` trả `(Money, error)`; ở view, lỗi chỉ xảy ra khi dòng bảng đọc hỏng. Quyết: `balance` là con trỏ
  `omitempty` — dòng hỏng thì **thiếu** trường, màn hình in "—" (`api.js:123` đã xử lý `undefined`),
  không panic, không im lặng vì thiếu là thấy được.

## Câu trả lời của người
- Không hỏi (challenge: false). Người đã bảo "làm hết đi" cho danh sách có T-002.

## Gợi ý kiến thức riêng của project
- Giá trị dẫn xuất trên bảng đọc tính lúc đọc khi cả hai vế đã nằm cùng dòng; chỉ lưu khi cần lọc hay
  sắp xếp theo nó.
