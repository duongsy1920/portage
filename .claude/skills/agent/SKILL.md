---
name: agent
description: Chạy một routine của Coding Agent cho một task. Dùng khi người gõ /agent <routine> <task-id>, hoặc nói "chạy analyze cho T-001", "intake việc này", "duyệt plan T-001", "từ chối T-002". Đây là phần duy nhất phụ thuộc harness; mọi luật và cách làm nằm trong agent/.
---

# /agent — cách gọi Coding Agent từ Claude Code

Bạn là **operator session** của Coding Agent (CONTRACT §6.1). Bạn không có luật riêng.
Mọi thứ bạn được và không được làm nằm trong `agent/`. Việc của skill này là mở đúng
file và làm theo.

## Cú pháp

```
/agent <routine> <task-id>      intake | analyze | plan | implement | verify | review
/agent intake                   task mới — task-id sinh trong routine
/agent approve <gate> <task-id> [note]     gate: plan | commit | push   ← việc của NGƯỜI, bạn ghi hộ
/agent reject <task-id> [note]
/agent answer <task-id>         người vừa trả lời câu hỏi trong chat — ghi vào ## Corrections của TASK.md
/agent status                   liệt kê task theo status
```

## Thứ tự đọc, không đổi

1. `agent/CONTRACT.md` — trọn, mỗi lần. Đây là ranh giới; đọc lại rẻ hơn vi phạm.
2. `agent/routines/<routine>/SKILL.md` — và **làm theo đúng Step 0 của nó trước**.
3. Không đọc `ROLE.md`, `CAPABILITIES.md` trọn mỗi lần — routine sẽ trỏ khi cần.

## `approve` · `reject` · `answer` — bạn là tay, người là đầu

Đây là ba việc **chỉ người được quyết** (CONTRACT §6.1). Bạn làm phần cơ học, để lại vết:

- `approve <gate> <id> [note]`: mở `TASK.md`, thêm vào `approvals`:
  `{gate: <gate>, by: <git config user.name của project>, at: <ISO giờ máy>, note: <note hoặc ~>}`;
  đổi `status: IN_PROGRESS`, `step:` = bước kế tiếp (`plan`→`implement`, `commit`→ giữ
  `review`, `push`→ giữ `review`), xoá `waiting_for`. Ghi một run record `routine: approve`.
  **Chỉ làm khi người vừa nói rõ trong chat.** Không suy ra từ "ok" mơ hồ — hỏi lại.
- `reject <id> [note]`: chỉ hợp lệ khi `status: WAITING_HUMAN` (T4). Đổi `status: REJECTED`,
  ghi note, run record. Nhánh: xoá **chỉ khi** không có commit; có commit thì để lại và ghi
  vào `review.md` mục *Left for you*.
- `answer <id>`: chép nguyên văn câu trả lời của người vào `## Corrections` của `TASK.md`
  kèm ngày; đổi `status: IN_PROGRESS`, giữ `step`. Run record.

## Không bao giờ

- Chạy hai routine liên tiếp vì "tiện". Người gọi từng bước — đó là thiết kế Phase 1.
- Tự ghi `approvals`. Tự sửa `CONTRACT.md`, `PROJECT.md`, `PAUSED`, `## Corrections`.
- Chạy routine khi `agent/PAUSED` tồn tại.

## Nếu thiếu

- Không có `agent/projects/<tên>/PROJECT.md` cho project hiện tại → nói vậy, đưa schema
  từ CONTRACT §2, **không đoán**.
- Task-id không tồn tại → liệt kê task đang có.
