---
name: intake
description: Nhận một việc từ mô tả trong chat và viết nó thành TASK.md có goal, context, acceptance, constraints, learning. Không phân tích, không lên kế hoạch — chỉ ghi lại đúng điều người muốn để analyze có thứ để đọc.
---

# intake — biến một câu nói thành một task đọc lại được

## Dòng chi phối cả file
**Task không có acceptance là task không kiểm được là xong.** Thiếu acceptance thì hỏi,
không tự bịa.

## Step 0 — năm dòng cố định (CONTRACT §8). Làm trước mọi thứ, đúng thứ tự

```
0.0  agent/PAUSED tồn tại? Rỗng, hoặc có dòng ghi tên routine này → ghi run record
     `skipped-paused`, thoát. Không đụng file đó.
0.1  Đọc projects/<p>/PROJECT.md → file PROJECT.rules → PROJECT.docs (index + file khớp vùng
     task đụng) → ## Corrections của mọi file sắp đụng. Ghi docs_read có ngày vào TASK.md.
0.2  Xác nhận đang ở PROJECT.path, rồi tuỳ routine (CONTRACT §4): intake/analyze/plan — đứng ở
     default hoặc nhánh của task, tree bẩn không chặn nhưng ghi vào run record · implement lúc
     tạo nhánh — tree sạch ngoài agent/, đang ở default · implement (nhánh đã có)/verify/review —
     đang ở nhánh của task, mọi file đang sửa ngoài agent/ nằm trong "File sẽ đụng" của plan.md.
     Sai → BLOCKED, dừng.
0.3  Đọc TASK.md → kiểm T1–T7 (CONTRACT §6.2) cho bước sắp làm. Vi phạm → BLOCKED kèm mã.
0.4  Đọc logs/runs/*-<task-id>-* → biết run trước dừng ở đâu. Đừng làm lại việc đã xong.
```

*(Ở intake chưa có TASK.md nên 0.3 và 0.4 bỏ qua; 0.1 và 0.2 vẫn làm.)*

## Đầu vào
Mô tả việc của người trong chat. Có thể kèm: `challenge` (hỏi trước khi kết luận),
`mode: learning` (người tự viết code).

## Step 1 — Sinh id và slug
- `id` = `T-` + số thứ tự 3 chữ số = (số thư mục trong `projects/<p>/tasks/`) + 1.
- `slug` = 2–4 từ không dấu, gạch nối, lấy từ goal. Thư mục: `tasks/<id>-<slug>/`.

## Step 2 — Soạn acceptance, không hỏi xin
Người mô tả việc bằng lời; **agent soạn** acceptance đo được từ lời đó theo best practice
(lệnh nào phải xanh, hành vi nào phải thấy được, con số nào không được đổi — đọc file luật
của project để biết con số nào là bất biến). Đánh dấu `(đề nghị)` sau từng dòng agent tự
soạn. Người sửa trong `## Corrections` nếu muốn; không sửa nghĩa là đồng ý.

Chỉ hỏi khi **goal** thật sự mơ hồ tới mức hai cách hiểu dẫn tới hai việc khác nhau — và
khi hỏi, nêu cách hiểu agent sẽ dùng nếu không có trả lời (CONTRACT §5.0). Không hỏi về
cách làm — đó là việc của `analyze` và `plan`.

## Step 3 — Ghi TASK.md (`file.write`)
Frontmatter theo CONTRACT §6, `status: TODO`, `step: analyze`, `verify: not-run`,
`branch: <prefix><id>-<slug>`, `created` = giờ máy ISO. Thân:

```markdown
## Goal
<một câu — điều muốn có sau khi xong>
## Context
<vì sao làm; trỏ tới chỗ trong repo hoặc file luật nói về việc này, nếu có>
## Acceptance
- [ ] <mỗi dòng một điều kiểm được bằng lệnh hoặc bằng mắt>
## Constraints
- <luật của project áp lên việc này — TRỎ tới PROJECT.rules, không chép>
## Learning
- <người muốn học gì qua việc này — để review rút đúng bài>
## Corrections
*(chỉ người viết)*
```

## Step 4 — Đổi trạng thái
`status: TODO` (đã là mặc định). Người sẽ gọi `analyze`.

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
