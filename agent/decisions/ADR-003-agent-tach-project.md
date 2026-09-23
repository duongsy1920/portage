# ADR-003 — Agent (dùng chung) tách khỏi Project (riêng từng repo)

**Ngày:** 2026-09-23 · **Trạng thái:** chấp nhận · **Bối cảnh:** `docs/CODING-AGENT.md` §3.1, §5.2

## Quyết định
Mọi thứ đổi theo repo đi qua **một** file `projects/<tên>/PROJECT.md` với schema đóng
(CONTRACT §2). Phần còn lại của `agent/` **không chứa một chữ nào của riêng project
nào** — không tên ngôn ngữ, không lệnh test, không tên file luật, không đường dẫn tuyệt
đối. Luật của một repo **ở lại** repo và **thắng** agent trên mọi thứ về repo (thang ưu
tiên CONTRACT §1 — học từ ROLE §0 của `ai-employees`).

## Vì sao
Yêu cầu gốc: dùng lại cho mọi project sau, kể cả project viết bằng ngôn ngữ khác. Cách duy
nhất kiểm được
lời hứa đó là làm cho phần dùng chung **không thể** biết project là gì.

## Ba lời hứa kiểm được
P1 — `scripts/verify-agent.sh` đỏ nếu có tên ngôn ngữ/công cụ/đường dẫn tuyệt đối ngoài
`projects/`. P2 — Phase 3: một task trên repo khác ngôn ngữ tới `DONE` với
`git diff agent/ -- ':!agent/projects'` rỗng. P3 — schema `PROJECT.md` đóng; thêm trường
phải qua ADR mới.

## Đã cân nhắc
Đặt luật project vào agent (một `CONTRACT.md` to cho tất cả) — lệch ngay lần thêm project
thứ hai. Để agent tự đoán lệnh test từ repo — đoán sai là chạy thứ không ai bảo chạy.

## Đánh đổi
Người phải viết `PROJECT.md` trước khi agent làm được gì. Agent không "tự cắm" vào repo lạ.

## Xem lại khi
Có trường nào trong `PROJECT.md` mà ≥ 2 project cần **khác nghĩa** — lúc đó schema đã sai.
