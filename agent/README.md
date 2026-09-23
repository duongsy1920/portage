# Coding Agent

Một đồng nghiệp lập trình do AI vận hành, làm việc trên **nhiều project** của một
người, dưới sự kiểm soát của người đó. Thiết kế đầy đủ và lý do của từng quyết định:
`../docs/CODING-AGENT.md`. File này chỉ nói cách dùng.

## Nó là gì trên đĩa

Không có runtime. Không có database. Agent là **bộ file Markdown** mà harness (hôm nay
là Claude Code) đọc và làm theo. Mọi trạng thái nằm trên đĩa, đọc được bằng mắt, diff
được bằng Git. Đóng Claude, đổi máy, đổi harness — không mất gì.

```
agent/
├── ROLE.md · CONTRACT.md · CAPABILITIES.md    ai · luật · cần làm được gì
├── routines/<tên>/SKILL.md                    sáu bước, người gọi từng bước
├── projects/<tên>/PROJECT.md                  RANH GIỚI DÙNG LẠI: thứ duy nhất đổi theo project
├── projects/<tên>/tasks/<id>/                 trạng thái + bằng chứng của một việc
├── knowledge/ · learning/debt.md · decisions/ kiến thức chung · nợ học · ADR
├── logs/runs/                                 một file mỗi lần chạy, chỉ ghi thêm
└── scripts/verify-agent.sh                    audit cơ học — đỏ là chưa xong
```

## Dùng thế nào

Trong Claude Code, mở repo chứa thư mục này, gõ:

```
/agent <routine> <task-id>        ví dụ: /agent analyze T-001
/agent intake                     tạo task mới từ mô tả trong chat
```

Sáu routine, theo thứ tự: `intake → analyze → plan → implement → verify → review`.
Agent **dừng chờ người** sau `plan`, sau `review` (trước commit, trước push), khi `analyze`
có `challenge`, và khi `verify` thấy gate đỏ. Mỗi lần dừng đều **đến kèm đề nghị + lý do +
phương án đã bỏ** (CONTRACT §5.0) — anh duyệt, từ chối, hay đổi hướng bằng một dòng; anh
không phải nghĩ ra giải pháp. Người duyệt bằng cách bảo agent
*"duyệt plan T-001"* — agent ghi một dòng vào `approvals` của `TASK.md` với tên anh và
ghi chú, rồi mới đi tiếp.

## Ba công tắc của người

- **`PAUSED`** — tạo file rỗng ở đây: mọi routine dừng. Ghi id routine từng dòng: chỉ
  dừng những cái đó. Xoá file: chạy lại. **Không routine nào được đụng file này.**
- **`## Corrections`** ở đuôi mọi file — anh viết một dòng có ngày, dòng đó **thắng**
  file nó nằm trong, từ lần chạy sau. Đây là cách agent học ý anh.
- **`releases`** trong `PROJECT.md` — anh nhả một kênh (plan hoặc push) cho một project
  khi đã đủ tin. Merge không bao giờ là kênh.

## Thêm một project

Tạo `projects/<tên>/PROJECT.md` theo schema trong `CONTRACT.md` §2. **Không sửa gì khác.**
Nếu phải sửa `routines/` để chạy được trên project mới, đó là bug của agent, không phải
việc của project.

## Trước khi bảo "xong" một task

```
./agent/scripts/verify-agent.sh
```
