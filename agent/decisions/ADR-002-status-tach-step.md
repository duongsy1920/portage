# ADR-002 — `status` (bóng ở đâu) tách khỏi `step` (đang làm gì)

**Ngày:** 2026-09-23 · **Trạng thái:** chấp nhận · **Bối cảnh:** `docs/CODING-AGENT.md` §5.4

## Quyết định
`TASK.md` có hai trường: `status` (7 giá trị đóng — TODO · IN_PROGRESS · WAITING_HUMAN
· DONE · BLOCKED · REJECTED · FAILED) và `step` (routine kế tiếp). Thêm `verify`
(not-run · passed · gate-failed) làm trường riêng.

## Vì sao
Danh sách 8 trạng thái tuần tự của prompt trộn hai câu hỏi: *ai đang cầm bóng* và
*routine nào đang chạy*. Hệ quả: "chờ duyệt" xuất hiện hai lần mà một enum không diễn tả
được, và người phải đọc 8 giá trị để biết có việc chờ mình không. Với hai trường, đọc
`status` là biết.

## Đã cân nhắc
Giữ 8 trạng thái tuần tự — đơn giản hơn về lý thuyết, nhưng project đầu tiên chạy 20 đợt
**không có trạng thái nào** và vẫn ổn, vì người luôn biết bóng ở đâu. Trạng thái phải mô
tả đúng điều đó.

## Đánh đổi
Hai trường thay một; máy đọc phải kiểm cặp (status, step) hợp lệ.

## Xem lại khi
Có routine chạy tự động nối tiếp (Phase 6) — lúc đó `step` có thể cần thành hàng đợi.
