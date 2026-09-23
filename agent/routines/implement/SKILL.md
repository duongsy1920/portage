---
name: implement
description: Làm đúng kế hoạch đã duyệt, trên nhánh riêng của task, chỉ đụng file trong danh sách đóng của plan.md. Gặp việc ngoài plan thì đề nghị sửa plan và dừng — không tự nới, không tự làm thêm vì "hiển nhiên là nên".
---

# implement — đúng plan, không hơn không kém

## Dòng chi phối cả file
**Diff phải khớp danh sách "File sẽ đụng" của plan.md. Một file ngoài danh sách là một
lần phải quay về xin.**

## Step 0 — năm dòng cố định (CONTRACT §8). Làm trước mọi thứ, đúng thứ tự

```
0.0  agent/PAUSED tồn tại? Rỗng, hoặc có dòng ghi tên routine này → ghi run record
     `skipped-paused`, thoát. Không đụng file đó.
0.1  Đọc projects/<p>/PROJECT.md → file PROJECT.rules → PROJECT.docs (index + file khớp vùng
     task đụng) → ## Corrections của mọi file sắp đụng. Ghi docs_read có ngày vào TASK.md.
0.2  Xác nhận đang ở PROJECT.path, rồi tuỳ routine (CONTRACT §4): intake/analyze/plan — đứng ở
     default hoặc nhánh của task, tree bẩn không chặn nhưng ghi vào run record · implement lúc
     tạo nhánh — tree sạch, đang ở default · implement (nhánh đã có)/verify/review — đang ở nhánh
     của task, mọi file đang sửa nằm trong "File sẽ đụng" của plan.md. Sai → BLOCKED, dừng.
0.3  Đọc TASK.md → kiểm T1–T7 (CONTRACT §6.2) cho bước sắp làm. Vi phạm → BLOCKED kèm mã.
0.4  Đọc logs/runs/*-<task-id>-* → biết run trước dừng ở đâu. Đừng làm lại việc đã xong.
```

*(Ở 0.3 kiểm chặt: **T1** — `approvals` có `plan` hoặc `releases` có `plan`; **T7** —
SHA-256 của `plan.md` hiện tại phải bằng `plan_digest` trong TASK.md. Lệch một → BLOCKED
kèm mã invariant, không đụng code.)*

## Đầu vào
`TASK.md` với `status: IN_PROGRESS, step: implement`, `plan.md` đã duyệt.

## Step 1 — Nhánh
- Nhánh `branch` của task chưa có → đang ở `branches.default`, tree sạch (0.2 đã xác nhận)
  → `git.branch` tạo từ đó. Đây là điểm **duy nhất** đòi tree sạch (CONTRACT §4).
- Đã có → chuyển sang. Đang ở nhánh khác không phải của task → BLOCKED (CONTRACT §4).
- File nào trong **File sẽ đụng** được đánh dấu *tạo mới* mà đã tồn tại ở dạng chưa theo
  dõi → BLOCKED: đó là file người đang viết (CONTRACT §4).

## Step 2 — Làm từng bước của plan
Theo đúng thứ tự mục **Bước**. Mỗi bước: chỉ đụng file ghi trong **File sẽ đụng**. Comment
trong code theo quy ước của file luật project (0.1 đã đọc).

**Gặp việc ngoài plan** (file chưa liệt kê, bước thiếu, giả định trong plan sai): **dừng
tại đó**. Không sửa `plan.md` (T7). Viết vào `plan.md` một mục mới
`## Phát hiện khi làm — đề nghị sửa plan` theo đúng CONTRACT §5.0: phát hiện gì · **đề
nghị** thêm/bỏ bước nào · vì sao · phương án khác đã bỏ. Tính lại `plan_digest`, đặt
`status: WAITING_HUMAN`, `waiting_for: plan-approval`. Code đã viết tới đó **để nguyên
trên nhánh** — không revert. Người duyệt phần sửa là làm tiếp từ chỗ dừng.

## Step 3 — Không chạy test ở đây
Đó là việc của `verify`, để kết quả gate có một chỗ ghi duy nhất. Được phép chạy
`commands.fmt` để code đúng định dạng trước khi rời tay.

## Step 4 — Đổi trạng thái
`status: IN_PROGRESS`, `step: verify`. Run record: `changed` liệt kê **mọi** file đã
đụng dạng `<project>:<đường dẫn>` — đây là thứ `review` so với plan.

## mode: learning
**Không chạy Step 1–2.** Người tự viết code trên nhánh của task. Run record ghi
`skipped: mode learning — người tự làm`, đổi `step: verify`, dừng.

## Cuối run — bắt buộc (CONTRACT §7)

1. Kiểm **R1–R4**. Sai một → `status: failed` trong record, dù đã làm gì.
2. Ghi **đúng một** file `logs/runs/<YYYY-MM-DD>-<HHMM>-<task-id>-<routine>.md` theo schema
   §7. `changed` liệt kê đường dẫn tương đối; file của repo ghi dạng `<project>:<đường dẫn>`.
3. Nói với người **3–8 dòng**: đã làm gì, dừng ở đâu, họ cần làm gì tiếp (nếu có).

## Không bao giờ ghi ra

Secret · giá trị biến môi trường · diff hay stack trace vào run record · tên công cụ hay
ngôn ngữ vào file này (gọi khả năng, đọc `PROJECT.commands`) · một con số chưa đo (dùng
từ vựng CONTRACT §5.5).

## Corrections

*(chỉ người viết — một dòng có ngày ở đây thắng file này từ lần chạy sau)*
