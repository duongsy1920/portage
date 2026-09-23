# CAPABILITIES — agent cần làm được gì

Routine gọi **tên khả năng**, không gọi tên công cụ. Ở Phase 1 harness là bản cài đặt
duy nhất, nên file này là **danh sách kiểm** ("routine này dùng gì?"), chưa phải bảng
ánh xạ. Nó thành bảng ánh xạ khi có harness thứ hai — và chỉ khi đó.

| Khả năng | Làm gì | Không bao giờ |
|---|---|---|
| `file.read` | đọc một file dưới `PROJECT.path` hoặc `agent/` | đọc file bí mật vào đầu ra |
| `file.write` | ghi một file — ghi tạm rồi đổi tên cho file mà crash có thể cắt nửa | ghi ngoài `PROJECT.path` và `agent/`; ghi `CONTRACT.md`, `PROJECT.md`, `PAUSED`, `## Corrections` |
| `code.search` | tìm định danh, chuỗi, mẫu trong `PROJECT.path` | — |
| `shell.run` | chạy **một** lệnh ghi trong `PROJECT.commands.*`, đọc exit code, thời gian, output; agent không hiểu nghĩa lệnh (CONTRACT §2) | chạy lệnh không nằm trong `commands.*` mà có ghi/xoá; chạy migration; báo "passed" khi exit ≠ 0 |
| `git.branch` | tạo/chuyển nhánh `<prefix><id>-<slug>` từ `branches.default` | tạo trên tree bẩn; chuyển khi người đang ở nhánh của họ |
| `git.diff` | đọc lại toàn bộ thay đổi trước khi commit | commit diff chưa đọc |
| `git.commit` | commit trên nhánh làm việc, sau `approvals.commit` | `--amend` lên commit đã push; message chứa secret/log/trace |
| `git.push` | push nhánh làm việc, sau `approvals.push` hoặc `releases.push` | `--force`; push nhánh `protected` |
| `secret.scan` | quét diff và mọi text sắp ghi ra, tìm token/key/connection string | bỏ qua vì "chắc không có" |

**Không có** `agent.execute`: ở Phase 1 agent *là* thứ đang chạy, không gọi agent khác.
**Không có** `browser.*`, `notify.*`: không có việc nào cần.

Một routine dùng khả năng không có trong bảng này là routine cần sửa, không phải bảng
cần thêm — cho tới khi có ≥ 3 run record chứng minh nhu cầu.

## Corrections

*(chỉ người viết)*
