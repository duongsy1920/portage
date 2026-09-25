---
name: analyze
description: Hiểu domain bị đụng trước khi có bất kỳ kế hoạch nào. Gọi tên aggregate, invariant, context bị chạm. Nếu task có challenge thì hỏi 3–5 câu rồi DỪNG chờ người trả lời — đó là điểm của challenge.
---

# analyze — hiểu đúng trước khi kết luận

## Dòng chi phối cả file
**Câu hỏi hỏi được ở đây mà để tới lúc đã viết code là lỗi của routine này. Câu hỏi
không kèm đề nghị cũng là lỗi của routine này.**

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

## Đầu vào
`TASK.md` với `status: TODO` (lần đầu) hoặc `status: IN_PROGRESS, step: analyze` (sau khi
người trả lời câu hỏi — câu trả lời nằm ở `## Corrections` của TASK.md).

## Step 1 — Đọc vùng code liên quan
Dùng `code.search` với các từ trong Goal. `file.read` file trúng, và file luật/docs khớp vùng đó
(đã làm ở 0.1). Ghi mọi file đã đọc vào `docs_read`.

## Step 2 — Nếu `challenge: true` và chưa có câu trả lời
Viết `analysis.md` với mục `## Câu hỏi trước khi kết luận`: 3–5 câu, mỗi câu phải đổi
hướng kế hoạch tuỳ câu trả lời — không hỏi câu nào mà trả lời kiểu gì cũng làm như nhau.
**Mỗi câu đi kèm đề nghị của agent** (CONTRACT §5.0), viết theo mẫu:

```
### Q1. <câu hỏi>
*(Tự trả lời trước nếu muốn học — đề nghị của agent ở dưới.)*
**Đề nghị:** <phương án cụ thể>
**Vì sao:** <lý do kiểm được, trỏ file:dòng hoặc mục trong file luật>
**Đã bỏ:** <phương án khác> — vì <lý do>
```

Người đọc và trả lời **một dòng** trong `## Corrections`: *"đồng ý hết"*, hoặc *"Q2: làm
X thay vì Y"*. Dòng *"Tự trả lời trước nếu muốn học"* là để challenge vẫn dạy được —
người chọn nghĩ trước hay đọc ngay, agent không quyết thay. Khung câu hỏi (dùng cái nào
hợp, không phải cả năm):

1. Thứ đang sửa thuộc **aggregate / module** nào? Ai là cửa vào?
2. **Invariant** nào bị đụng, và nó đang được giữ ở đâu (domain, tầng app, hay chưa ở đâu)?
3. Đây là **luật nghiệp vụ** hay **chi tiết vận hành**? Hệ quả: lỗi thì ai sai — người viết
   code hay người dùng — và file luật của project nói gì về hai trường hợp đó?
4. **Chuyển trạng thái** nào hợp lệ? Trạng thái nào đã qua điểm không quay đầu?
5. Có **quyết định đã đóng** trong file luật/docs mà việc này đang mở lại không?

Rồi, tuỳ kênh `plan` của project (ADR-007):
- **`releases` chưa có `plan`** → `status: WAITING_HUMAN`, `waiting_for: question`, run record,
  **dừng**. Chưa viết mục kết luận (Step 3).
- **`releases` có `plan`** → **không dừng.** Ghi ngay dưới mỗi câu một dòng
  *"Làm theo đề nghị (kênh plan đã nhả · <ngày>)"*, đi tiếp Step 3 với đề nghị của mình làm
  câu trả lời. Người đọc các câu này cùng lúc với demo ở cửa commit; muốn đổi thì nói ở đó.
  **Ngoại lệ duy nhất vẫn dừng:** câu hỏi về thông tin chỉ người có (số nghiệp vụ mới, ưu tiên
  kinh doanh, ai được phép gì) — CONTRACT §5.0 — và phải kèm mặc định sẽ dùng nếu không ai trả lời.

## Step 3 — Viết analysis.md
(Không challenge, hoặc đã có câu trả lời trong `## Corrections`.) Mục cố định:

```markdown
## Đụng vào cái gì
- aggregate / module: … (file:dòng)
- invariant: … — hiện giữ ở …
- context / module khác bị chạm: … — qua đường nào (import? event? id trần?)
## Luật của project áp vào đây
- <trỏ mục trong PROJECT.rules; không chép>
## Điều ngoài luật, hoặc quyết định đã đóng bị chạm
- <hoặc: không có>
## Câu trả lời của người (nếu challenge)
- <chép nguyên văn từ ## Corrections, kèm ngày>
## Gợi ý kiến thức riêng của project (nếu có điều chưa ghi ở đâu)
- <một câu khẳng định — người chốt mới ghi vào projects/<p>/knowledge/>
```

Mọi khẳng định về code phải có `file:dòng`. Không có thì viết `chưa xác định`.

## Step 4 — Đổi trạng thái
`status: IN_PROGRESS`, `step: plan`.

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
