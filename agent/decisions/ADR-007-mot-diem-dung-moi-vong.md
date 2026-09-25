# ADR-007 — Một điểm dừng mỗi vòng, đặt sau khi có thứ để xem

**Ngày:** 2026-09-25 · **Trạng thái:** chấp nhận (người duyệt trong chat: *"oke"* cho "nấc 1") ·
**Bối cảnh:** CONTRACT §3.1 (`releases`), §5.0; DEC-008, DEC-009; T-001 chạy tới `verify`

## Quyết định
1. **Kênh `plan` nhả cho Portage** — một dòng `releases` trong `PROJECT.md`, ghi hộ theo lời
   người. Agent đi thẳng intake → analyze → plan → implement → verify → review, không dừng giữa.
2. **Điểm dừng duy nhất của một vòng là cửa commit.** `review.md` có thêm mục **Xem thử**: lệnh,
   URL, bấm gì, ảnh chụp (`tasks/<id>/shots/`) khi là việc UI. Người xem demo rồi mới quyết.
3. **Đổi hướng không mở task mới.** Tại cửa commit, *"sửa X"* → dòng `## Corrections` → `plan`
   viết mục *Sửa theo đổi hướng* → implement → verify → review chạy lại cùng nhánh → dừng lại
   ở cửa commit.
4. **Challenge không chặn** khi kênh `plan` đã nhả: câu hỏi và đề nghị vẫn ghi vào `analysis.md`,
   agent làm theo đề nghị của mình; chỉ dừng khi thiếu thông tin chỉ người có, kèm mặc định.
5. **`/agent run <id>`** chạy nối routine tới điểm dừng kế; sau `approve`/`answer` cũng chạy nối.
   Mỗi routine vẫn Step 0 riêng, run record riêng, R1–R4 riêng.
6. `commit` và `push` **vẫn giữ**. Merge không bao giờ là kênh.
7. Làm rõ ADR-006: skill gọi agent trong harness (`.claude/skills/agent/`) thuộc về agent như
   `agent/` — sửa nó không làm bẩn tree của task. Tìm ra vì chính ADR này phải sửa file đó
   trong lúc T-001 đang ở giữa vòng.

## Vì sao
Người nói thẳng: *"mọi thứ hầu như tôi không thể quyết định được vì khi dựng lên UI rồi thì tôi
mới xem demo rồi tôi sẽ quyết định chỉnh sửa sau."* Cửa plan đòi quyết một thứ chưa nhìn thấy;
câu *"đồng ý hết 4 đề nghị"* ở T-001 là chữ ký hình thức, không phải quyết định — một cửa chỉ
sinh chữ ký hình thức thì không bảo vệ gì. Cửa phải nằm ở chỗ người **có thể** quyết. Cách làm
việc thật của người là xem → chỉnh → xem lại; mô hình phải là vòng đó, với hoàn tác rẻ (một nhánh,
một commit) thay cho duyệt trước.

## Đã cân nhắc
- Giữ cửa plan nhưng viết plan dễ hình dung hơn — thứ người cần là chạy thử, không phải đọc hay hơn.
- Nhả cả `commit` và `push` ngay — `review` chưa chạy lần nào, chưa có dữ liệu về độ đúng của đề
  nghị agent; nới trước khi có bằng chứng là đúng cái bẫy ROLE ghi cho chữ "senior" (DEC-019).
  Đây là **nấc 2**, mở sau 5 task qua cửa commit với ≤ 1 lần đổi hướng mỗi task.
- `BACKLOG.md` + `/agent next` để người không gõ lệnh — **nấc 3**, sau nấc 2.

## Đánh đổi
Agent sẽ có lúc chọn thứ người đổi lại sau khi xem. Chi phí là một vòng đổi hướng trên nhánh,
rẻ hơn một cửa duyệt mà người không thể dùng. Mỗi lần đổi hướng phải thành một dòng luật
(`## Corrections`, design system) để lần sau agent tự quyết đúng chỗ đó — không ghi lại thì vòng
lặp không học được gì.

## Xem lại khi
Có 5 task qua cửa commit: đếm số lần đổi hướng mỗi task từ run log; ≤ 1 → mở nấc 2 (cần sửa
CONTRACT §3.1 để `commit` nhả được, và một ADR). Một task phải revert → siết ngay về nấc 1.
