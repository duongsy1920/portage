# Project: portage

Hệ mua hộ Mỹ → Việt Nam. Repo Go, DDD, 5 bounded context. Luật của repo nằm **trong
repo** (`rules` dưới đây) và **thắng** agent trên mọi thứ về repo — trừ hai guardrail
và cách ly (CONTRACT §1).

```yaml
name:      portage
path:      /home/duongvantiensy/portage
rules:     CLAUDE.md
docs:      docs/
commands:
  test:    go test ./...
  fmt:     gofmt -l .
  lint:    go vet ./...
  smoke:   bash scripts/smoke.sh     # file không có bit thực thi trong repo; cần PORTAGE_DSN (và PORTAGE_PORT nếu 8080 bận); ghi hộ theo lời anh 25/09
branches:
  default:   main
  protected: [main]
  prefix:    agent/
releases:
  - {channel: plan, since: 2026-09-25, note: "anh quyết sau khi xem demo ở cửa commit (ADR-007); ghi hộ theo lời anh trong chat"}
```

## Ghi chú cho agent khi đọc luật của repo này

Không chép gì từ `CLAUDE.md` vào đây. Chỉ trỏ tới chỗ đọc:

- Luật làm việc bắt buộc, "Ba bài học đã trả giá", "Trước khi bảo xong" → `CLAUDE.md`.
- Vì sao code trông như vậy: `docs/SETUP.md` §9, đọc 2–3 đợt cuối.
- Việc còn nợ (ứng viên task): `CLAUDE.md` mục "Còn nợ".
- Quyết định đã HOÃN, **đừng đề xuất lại**: `CLAUDE.md` mục "Quyết định 08/09".

## Corrections

*(chỉ người viết)*
