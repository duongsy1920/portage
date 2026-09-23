---
name: review
description: Review diff theo ba chiều (idiom ngôn ngữ · mô hình domain · kỹ thuật) và nói VÌ SAO; viết change brief làm thân PR; rút learning.md và nợ học; rồi commit và push — mỗi việc sau một cửa duyệt, hoặc kênh đã nhả.
---

# review — nói vì sao, rồi đi qua từng cửa

## Dòng chi phối cả file
**Câu mở đầu của review.md là bất biến: "Mọi dòng dưới đây mô tả một thay đổi trên nhánh.
Chưa merge. Merge là của anh."** Routine này có ba lần vào, mỗi lần một việc.

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

## Đầu vào và ba lần vào
`TASK.md` với `step: review`. Routine nhìn ba thứ để biết mình đang ở lần nào:

| `review.md` | `approvals` | commit trên nhánh | Lần này làm |
|---|---|---|---|
| chưa có | — | — | **A. Review** → chờ duyệt commit |
| có | có `commit` | chưa | **B. Commit** → chờ duyệt push, hoặc push nếu đã nhả |
| có | có `push` hoặc `releases.push` | có | **C. Push** → DONE |

## A. Review

### A1 — Đọc diff, so với plan
`git.diff` toàn bộ nhánh so với `branches.default`. **Đọc hết** — diff chưa đọc là thay
đổi đang đoán. So từng file với **File sẽ đụng** của plan.md: file ngoài danh sách → ghi
`## Vượt plan` trong review.md, và đó là một điểm **đề nghị revert** trừ khi có lý do.

### A2 — `secret.scan` toàn bộ diff
Bị đánh dấu → **không đi tới cửa commit.** Ghi rõ file:dòng và loại, đề nghị thay giá trị
bằng tên biến, `status: WAITING_HUMAN`, `waiting_for: plan-approval` (vì cần sửa code).

### A3 — Viết review.md
Nửa trên là **change brief** — thân PR sẵn dán. Nửa dưới là review có lý do.

```markdown
# <một dòng: đã đổi gì>

Mọi dòng dưới đây mô tả một thay đổi trên nhánh. Chưa merge. Merge là của anh.

## What changed and why
<3–6 câu. Vì sao, không chỉ cái gì. Trỏ acceptance nào được đáp>
## Files
<đường dẫn>, +a −b
## Gate
<chép bảng Verify từ TASK.md, kèm "Chưa kiểm">
## Rollback
<một dòng: revert commit <hash> / xoá nhánh — nói rõ cái nào>
## Left for you
<việc chỉ người làm được: chạy môi trường thiếu, quyết định kinh doanh, merge. Không có → "nothing">
## Compare
<nhánh> so với <default>

---

## Review — vì sao, không chỉ cái gì
### Idiom của ngôn ngữ project
- <nhận xét> — vì <lý do>, xem <file:dòng>. Đề nghị: <giữ / sửa thế nào>
### Mô hình domain
- aggregate/invariant có bị lách không · ranh giới module có bị xuyên không · luật của
  project (đã đọc ở 0.1) có bị vi phạm không
### Kỹ thuật
- lỗi xử lý thế nào · secret · tương thích ngược · điều gì hỏng được lúc chạy thật
### Vượt plan
- <hoặc: không có>
### Kết luận — đề nghị
**Đề nghị: commit** (hoặc: **Đề nghị: sửa X rồi mới commit**, kèm lý do). Một dòng.
```

Mỗi nhận xét phải có `file:dòng` và một **đề nghị** (CONTRACT §5.0). Nhận xét không có đề
nghị là than phiền, không phải review.

### A4 — learning.md, ≤ 12 dòng, mẫu cố định
```
Ngôn ngữ:  ≤3 — mỗi dòng trỏ file:dòng VỪA VIẾT trong task này
Domain:    ≤3
Kỹ thuật:  ≤2
Agent:     ≤2 — điều gì về chính agent lộ ra qua task này (câu hỏi lẽ ra hỏi sớm hơn, plan sai ở đâu)
Sai ở đâu: <hoặc: không>
Nợ mới:    <mỗi dòng → thêm vào learning/debt.md mục của project, kèm "từ <task> · <ngày>">
```
Không có dòng nào là lý thuyết chung. Không trỏ được vào code vừa viết thì bỏ.

### A5 — Nợ học
Thêm dòng **Nợ mới** vào `learning/debt.md`. Nợ nào trong đó **30 ngày** không ai đụng →
gợi xoá trong run record (không tự xoá).

### A6 — Cửa commit
`status: WAITING_HUMAN`, `waiting_for: commit-approval`. Nói với người: review.md ở đâu,
dòng **Kết luận — đề nghị** nói gì, và "Left for you" có gì.

## B. Commit (sau `approvals.commit`)
`git.commit` trên nhánh task. Message: dòng đầu của **What changed and why** · dòng trống
· `Task: <id>` · dòng trống · trailer theo quy ước của file luật project nếu có. Qua
`secret.scan`. Không trace, không log, không đường dẫn tuyệt đối.
Rồi: `releases` có `push` → làm **C** luôn trong run này, ghi rõ *"push qua kênh đã nhả"*.
Không → `status: WAITING_HUMAN`, `waiting_for: push-approval`.

## C. Push (sau `approvals.push` hoặc `releases.push`)
Kiểm **T3** rồi `git.push` nhánh task (không `--force`, không nhánh protected). Kiểm **T2**
(review.md · verify passed · approvals.commit) → `status: DONE`, xoá `step`, `waiting_for`.
Nói với người: nhánh đã push ở đâu, thân PR nằm ở nửa trên review.md.

## mode: learning
**A** review diff **của người** trên nhánh task, cùng mẫu. **B** và **C** không chạy — người
tự commit và push code của mình. Run record ghi rõ.

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
