---
name: verify
description: Chạy mọi lệnh kiểm chứng PROJECT.md khai (test, fmt, lint, build), ghi exit code và thời gian, và nói rõ acceptance nào CHƯA được lệnh nào chứng minh. Gate hỏng thì để nhánh y nguyên, chẩn đoán, đề nghị cách sửa, chờ duyệt — không tự thử cách khác.
---

# verify — nói cái đã chạy, cái chưa chạy, và exit code

## Dòng chi phối cả file
**"Tests passed" trần là lỗi. Câu đúng có ba phần: chạy gì · kết quả gì · CHƯA chạy gì.**

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

## Đầu vào
`TASK.md` với `status: IN_PROGRESS, step: verify`, có diff trên nhánh của task.

## Step 1 — Chạy từng lệnh có trong PROJECT.commands
Thứ tự: `fmt` → `lint` → `build` → `test`. Với mỗi lệnh **có khai**: `shell.run`, ghi exit
code, thời gian, số dòng output. Lệnh **không khai** → ghi `n/a (no command)` — không bỏ
trống, không tự đoán lệnh thay.

Exit ≠ 0 → lấy **dòng lỗi đầu tiên** (một dòng, không phải cả output, không phải tóm tắt
của agent), cho qua `secret.scan`; bị đánh dấu → ghi `first failing line withheld
(<loại>)` và trỏ vị trí log.

## Step 2 — Chưa kiểm gì
Bắt buộc. Đọc lại **Acceptance** trong TASK.md; với mỗi dòng, ghi lệnh nào ở Step 1 chứng
minh nó, hoặc `not run (<lý do>)`. Acceptance cần môi trường không có (database, dịch vụ
ngoài, tay người bấm) → ghi rõ tên môi trường thiếu và ai/khi nào chạy được.

## Step 3 — Ghi kết quả
Vào `TASK.md`, mục mới `## Verify <ISO>`:

```
| lệnh | exit | thời gian | ghi chú |
|---|---|---|---|
| <commands.fmt>  | 0 | 0.4s | — |
| <commands.lint> | 0 | 1.1s | — |
| <commands.test> | 0 | 1.7s | <số package/case nếu output cho biết; không thì "not measured"> |
| <commands.build>| n/a (no command) | | |

Chưa kiểm:
- <acceptance> — not run (<lý do>) — chạy được khi <điều kiện>
```

Frontmatter: `verify: passed` nếu **mọi** lệnh đã chạy exit 0; `verify: gate-failed` nếu
có lệnh ≠ 0. Lệnh `n/a` không tính là hỏng.

## Step 4 — Gate hỏng là run thành công (CONTRACT §5.3)
**Để nhánh y nguyên.** Không revert, không amend, không thử cách khác trong run này. Rồi
làm đúng CONTRACT §5.0 — **đến với đề nghị**: đọc dòng lỗi và diff, viết vào `plan.md`
mục `## Sửa đề nghị sau gate hỏng <ISO>`: chẩn đoán (lỗi ở file:dòng nào, vì sao) · **đề
nghị** sửa gì · vì sao · cách khác đã bỏ. Tính lại `plan_digest`. `status: WAITING_HUMAN`,
`waiting_for: plan-approval`. Người duyệt một dòng → `implement` chạy phần sửa.

Không chẩn đoán được → vẫn đề nghị: *"Đề nghị: đọc `<file>` cùng nhau — tôi không tìm
được nguyên nhân sau khi xem <chỗ đã xem>."* Đó vẫn là đề nghị, không phải câu hỏi trống.

## Step 5 — Passed
`status: IN_PROGRESS`, `step: review`.

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
