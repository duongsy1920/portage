# ADR-006 — Trạng thái của agent không tính là tree bẩn, và không vào commit của task

**Ngày:** 2026-09-25 · **Trạng thái:** chấp nhận (người duyệt trong chat: *"đồng ý ADR-006"*) ·
**Bối cảnh:** CONTRACT §4, §8 dòng 0.2, R1 — tìm ra ở lần `plan` đầu tiên sau khi `agent/` được commit

## Quyết định
Mọi phép kiểm tree trong CONTRACT §4 **bỏ qua `agent/`**: "sạch" là không file đã theo dõi
nào **ngoài `agent/`** bị sửa; "mọi file đang sửa nằm trong plan" cũng chỉ xét ngoài `agent/`.
Đổi lại, `git.commit` của một task **không bao giờ** stage file dưới `agent/` (R1 thêm một
vế); `review` stage đúng các file trong "File sẽ đụng". Trạng thái agent do người commit
riêng, khi muốn.

## Vì sao
ADR-005 giả định trạng thái của agent "luôn ở dạng chưa theo dõi". Sai từ commit `ea04696`:
`agent/` nằm trong repo, nên **mỗi run** làm bẩn file đã theo dõi (`TASK.md`, `analysis.md`,
run record), và `implement` sẽ tự chặn mình ở bước tạo nhánh. Trạng thái agent là sản phẩm
phụ của mọi run, không phải việc dở của người — luật tree bẩn sinh ra để bảo vệ việc dở của
người, không phải để chặn chính agent.

## Đã cân nhắc
- Commit trạng thái agent lên `main` trước mỗi lần `implement` — chạy được hôm nay, nhưng
  mỗi task thêm một commit thủ công chỉ để mở đường, và quên là chặn.
- Commit trạng thái agent **cùng** commit của task — PR lẫn file `agent/`, và merge kéo
  trạng thái từ nhánh task về `main` theo thứ tự merge, không theo thứ tự run.
- Tách `agent/` ra repo riêng — đúng hướng dài hạn (DEC-020), nhưng người đã chọn chưa tách
  ở phase này.

## Đánh đổi
Trạng thái agent và code của task đi hai đường commit; ai đọc lịch sử phải biết điều đó.
`review` phải stage theo danh sách, không `git add -A` — thêm một phép so, nhưng phép so đó
chính là T7 nhìn từ phía index.

## Xem lại khi
Tách `agent/` ra repo riêng, hoặc có harness thứ hai: lúc đó vế "bỏ qua `agent/`" thành
"bỏ qua thư mục của agent", và đường commit trạng thái có thể tự động.
