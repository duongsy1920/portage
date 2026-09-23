# ADR-001 — Lưu trữ bằng file Markdown + frontmatter, không database

**Ngày:** 2026-09-23 · **Trạng thái:** chấp nhận · **Bối cảnh đầy đủ:** `docs/CODING-AGENT.md` §5.5

## Quyết định
Mọi trạng thái (task, duyệt, run, nợ học) là file Markdown có frontmatter YAML, trong
`agent/`. Không SQLite, không Postgres.

## Vì sao
Người dùng là một; đọc bằng mắt là yêu cầu số một; mọi thay đổi trạng thái tự có audit
log qua Git; sống qua đổi harness vì file là file. `SETUP.md §9` của project đầu tiên
chứng minh Markdown chịu được 20 đợt.

## Đã cân nhắc
SQLite — không diff được, không review được trong PR. Postgres — cần server; project đầu
tiên từng bị chiếm cổng.

## Đánh đổi
Không có khoá ghi. Hai run song song trên cùng task sẽ ghi đè nhau.

## Xem lại khi
Có run chạy song song (Phase 6). Lúc đó: file lock, hoặc SQLite làm **chỉ mục dựng lại
được** từ file — file vẫn là nguồn sự thật.
