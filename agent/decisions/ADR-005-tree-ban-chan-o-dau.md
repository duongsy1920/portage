# ADR-005 — Tree bẩn chỉ chặn lúc tạo nhánh; sau đó diff chưa commit là việc đang làm

**Ngày:** 2026-09-23 · **Trạng thái:** chấp nhận · **Bối cảnh:** CONTRACT §4, §8 dòng 0.2 — tìm ra ở lần chạy `intake` đầu tiên

## Quyết định
Dòng 0.2 của Step 0 phân biệt theo routine. `intake` · `analyze` · `plan` không ghi vào repo
nên tree bẩn không chặn — chỉ cần đứng ở `default` hoặc nhánh của task, và ghi vào run record.
`implement` đòi tree **sạch** đúng một lần: lúc `git.branch`. Từ đó `verify` và `review` chấp
nhận tree bẩn với một điều kiện: mọi file đang sửa nằm trong "File sẽ đụng" của `plan.md`.

## Vì sao
Bản đầu ghi *"working tree sạch"* cho **mọi** routine. Chạy thật mới thấy nó tự mâu thuẫn:
`implement` viết code nhưng **không commit** (commit là kênh giữ), nên tới `verify` tree
**luôn** bẩn — pipeline không bao giờ qua được bước 4. Và lần `intake` đầu tiên đã bị chặn bởi
chính file thiết kế đang sửa dở của người — một routine chỉ ghi vào `agent/` không có lý do gì
để từ chối vì repo có việc dở.

## Đã cân nhắc
- Giữ "sạch cho mọi routine" và bắt `implement` commit ngay — mở kênh commit không qua duyệt,
  vi phạm §3.1.
- Bỏ hẳn điều kiện tree — mất lớp bảo vệ duy nhất chống việc tạo nhánh kéo theo việc dở của
  người vào diff của agent.

## Đánh đổi
Dòng 0.2 dài hơn năm dòng khác, và `verify`/`review` phải so tree với `plan.md` — thêm một
phép so, nhưng phép so đó chính là T7 nhìn từ phía đĩa.

## Xem lại khi
Có harness thứ hai hoặc chạy tự động (Phase 6): lúc đó có thể cần `git stash` có kiểm soát
thay cho `BLOCKED`, và phải quyết lại xem ai được stash.
