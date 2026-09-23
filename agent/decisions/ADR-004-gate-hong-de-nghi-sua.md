# ADR-004 — Gate hỏng đi qua cổng plan kèm đề nghị sửa, không hỏi trống

**Ngày:** 2026-09-23 · **Trạng thái:** chấp nhận · **Bối cảnh:** CONTRACT §5.0, §5.3, T7

## Quyết định
Khi `verify` thấy gate đỏ, routine **không** đặt `WAITING_HUMAN/question`. Nó chẩn đoán,
viết mục *đề nghị sửa* vào `plan.md` (chẩn đoán · sửa gì · vì sao · cách khác đã bỏ), tính
lại `plan_digest`, và đặt `WAITING_HUMAN/plan-approval`. Người duyệt một dòng, `implement`
làm phần sửa. "Gặp việc ngoài plan khi implement" đi đúng đường này.

## Vì sao
Người dùng agent để **bớt phải nghĩ về giải pháp**. Một câu *"test đỏ, anh muốn làm gì?"*
đẩy việc thiết kế về phía người — đúng điều CONTRACT §5.0 cấm. Và nếu có hai đường (question
cho test đỏ, plan-approval cho việc ngoài plan) thì T7 có một đường tắt: sửa code sau test
đỏ mà `plan.md` không đổi, digest vẫn khớp, plan đã duyệt không còn mô tả cái đang làm.

## Đã cân nhắc
- `WAITING_HUMAN/question` kèm đề nghị trong câu hỏi — thoả §5.0 nhưng vẫn để T7 hở, vì
  câu trả lời "ok sửa đi" không đổi `plan.md`.
- `verify` tự sửa rồi chạy lại — vi phạm "gate hỏng là run thành công" và "không thử cách
  khác"; xoá mất bằng chứng về thay đổi không chạy.

## Đánh đổi
`verify` phải đọc diff và chẩn đoán — việc nặng hơn "chạy lệnh, ghi exit code". Chấp nhận:
chẩn đoán sai thì người thấy ngay ở cổng plan; im lặng thì không ai thấy gì.

## Xem lại khi
Có routine chạy tự động nối tiếp (Phase 6): lúc đó chuỗi verify → sửa → verify có thể lặp,
cần giới hạn số lần đề nghị trước khi `BLOCKED`.
