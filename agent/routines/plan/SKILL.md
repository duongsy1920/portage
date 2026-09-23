---
name: plan
description: Viết kế hoạch từ analysis.md — bước, file sẽ đụng, test sẽ thêm, mục "Không làm", mục "Tự phản biện" — rồi dừng chờ duyệt. Kế hoạch được duyệt là kế hoạch được làm, không hơn không kém.
---

# plan — đề xuất, không quyết

## Dòng chi phối cả file
**Không có mục "Tự phản biện" thì kế hoạch chưa xong.** Học từ một đề xuất đã phải rút
lại vì đưa ra mà chưa nghĩ kỹ.

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
`TASK.md` với `status: IN_PROGRESS, step: plan`, và `analysis.md` đã có.

## Step 1 — Viết plan.md

```markdown
## Mục tiêu
<chép Goal từ TASK.md>
## Bước
1. <một bước — một việc kiểm được; mỗi bước nêu file sẽ đụng>
2. …
## File sẽ đụng — danh sách ĐÓNG
- <đường dẫn> — <sửa gì>
(implement không được đụng file ngoài danh sách này. Cần thêm → về hỏi.)
## Test sẽ thêm hoặc đổi
- <tên test, file> — chứng minh acceptance nào
## Không làm
- <những thứ "hiển nhiên nên làm luôn" mà cố ý không làm, và vì sao>
## Tự phản biện
- Cách hiển nhiên hơn là: …
- Không chọn vì: … (phải là lý do kiểm được, không phải sở thích)
- Chỗ kế hoạch này có thể sai: …
- **Đề nghị: làm theo kế hoạch này.** (hoặc: *"Đề nghị: KHÔNG làm, vì …"* — một kế hoạch
  kết luận không nên làm vẫn là kế hoạch xong việc)
## Rủi ro
- <cái gì có thể hỏng, và biết bằng cách nào>
## Acceptance (chép từ TASK.md)
- [ ] …
```

## Step 2 — Ghi digest của plan
Tính SHA-256 của `plan.md`, ghi `plan_digest: <hex>` vào frontmatter TASK.md. `implement`
sẽ so lại — đây là cách kiểm **T7** bằng máy: plan đã duyệt không được đổi.

## Step 3 — Cổng plan
- `PROJECT.releases` có `channel: plan` → tự ghi
  `approvals: [{gate: plan, by: release, at: <ISO>, note: "released since <since>"}]`,
  `status: IN_PROGRESS`, `step: implement`. Ghi vào run record là **đã đi qua kênh đã nhả**.
- Không → `status: WAITING_HUMAN`, `waiting_for: plan-approval`, dừng. Nói với người
  đúng ba thứ: kế hoạch ở đâu, mục "Không làm" nói gì, mục "Tự phản biện" nói gì.

## mode: learning
Thay mục **Bước** bằng `## Gợi ý cho người`: hướng, chỗ nên bắt đầu, câu hỏi để tự kiểm.
Không có bước máy làm. Vẫn có "Không làm" và "Tự phản biện".

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
