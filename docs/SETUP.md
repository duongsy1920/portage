# Portage — Sổ tay dựng môi trường & khởi tạo source

> File ghi chép cá nhân. Ghi lại **đúng những gì đã làm**, để sau này dựng lại
> trên máy khác không phải mò lại từ đầu.
>
> Cập nhật: 2026-09-04
>
> 📚 **Học DDD:** xem [DDD.md](DDD.md) — toàn bộ khái niệm, kiến trúc, bẫy
> thường gặp, giải thích bằng chính code của project này.
>
> 🔤 **Cú pháp Go:** xem [GO-CHO-PHP.md](GO-CHO-PHP.md) — bảng tra Go ↔ PHP.
> Trong code còn có comment `// [PHP]` giải thích tại chỗ; cách xoá hàng loạt
> nằm ở mục 9 của file đó.

---

## 0. Project này là gì

**Portage** — hệ thống đặt hàng hộ xuyên biên giới (Mỹ → Việt Nam).
Khách đặt trên web của mình, mình mua hộ ở các web Mỹ (Nike, The North Face,
Sony…), gom hàng, ship về tận nhà khách ở VN.

Viết bằng **Go**, kiến trúc **DDD (Domain-Driven Design)**.

Mục tiêu kép: vừa là sản phẩm thật, vừa là project học Go + DDD để đưa vào CV.

---

## 1. Máy đang có gì (kiểm tra 03/09/2026)

| Công cụ | Phiên bản | Ghi chú |
|---|---|---|
| Go | **1.27.0** | mới cài hôm nay |
| Git | 2.51.0.windows.1 | có sẵn |
| Docker Desktop | 29.7.2 | chạy Postgres 16 (`docker compose up -d`). **Phải bật Docker Desktop** trước — daemon tắt thì `docker` báo `npipe … cannot find the file` |
| winget | 1.29.290 | trình cài đặt của Windows |
| Node.js | có sẵn ở `D:\NodeJs` | **không thuộc stack**; chỉ dùng chạy script tính giá nháp |

> ⚠️ **Node KHÔNG cần cho project này.** Stack chính là Go + Postgres.
> Node chỉ tình cờ được dùng để chạy mấy file tính chi phí lúc bàn nghiệp vụ.

> ⚠️ **Máy không có `gcc`**, nên `go test -race` (race detector, cần cgo) **không
> chạy được trên Windows**. Không cần cài — CI trên Linux chạy thay
> (xem mục 5b). Trên máy chỉ cần `go test ./...`.

---

## 2. Cài Go trên Windows

### Cách đã dùng (nhanh nhất)

```powershell
winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements
```

### Cách thủ công (nếu không có winget)

Tải `.msi` ở <https://go.dev/dl/> rồi chạy. Bản đã dùng: `go1.27.0.windows-amd64.msi`

### ⚠️ Bẫy hay gặp: cài xong gõ `go` báo "not found"

Bộ cài **đã tự thêm** `C:\Program Files\Go\bin` vào PATH hệ thống,
**nhưng terminal đang mở không tự cập nhật.**

| Tình huống | Làm gì |
|---|---|
| Cách đúng (khuyến nghị) | **Đóng hẳn rồi mở lại** terminal |
| Cần dùng ngay | set PATH tạm cho phiên hiện tại (bên dưới) |

```powershell
# PowerShell
$env:Path += ";C:\Program Files\Go\bin"
```

```bash
# Git Bash
export PATH="$PATH:/c/Program Files/Go/bin"
```

Kiểm tra:

```bash
go version
# go version go1.27.0 windows/amd64
```

### Biến môi trường Go — không phải set tay, nhưng nên biết

| Biến | Giá trị trên máy này | Nghĩa |
|---|---|---|
| `GOROOT` | `C:\Program Files\Go` | nơi cài bản thân Go |
| `GOPATH` | `C:\Users\tony_\go` | thư viện tải về + binary do `go install` sinh ra |
| `GOBIN` | `%GOPATH%\bin` | nơi để file `.exe` sau `go install` |
| `GOMODCACHE` | `%GOPATH%\pkg\mod` | cache thư viện đã tải |

> **Khác Composer:** Go **không** có thư mục `vendor/` trong project.
> Thư viện nằm chung ở `GOMODCACHE`, mọi project dùng chung.
> Không có `vendor` nghĩa là repo nhẹ, không phải ignore hàng nghìn file.

Xem toàn bộ cấu hình: `go env`

---

## 3. Các lệnh đã chạy để khởi tạo source

```bash
# 1. Tạo thư mục project
mkdir -p /d/portage
cd /d/portage

# 2. Khởi tạo Go module   (tương đương `composer init`)
go mod init github.com/duongsy/portage
```

> **`go mod init` làm gì:** sinh file `go.mod` — vai trò như `composer.json`.
> Tham số truyền vào là **module path**, đóng vai trò namespace gốc của project.
> Quy ước đặt theo URL repo (`github.com/<user>/<repo>`) để sau này người khác
> `go get` được.
>
> ⚠️ Nếu username GitHub không phải `duongsy` thì sửa dòng đầu `go.mod`
> và sửa import trong các file `_test.go`.

```bash
# 3. Tạo cấu trúc thư mục DDD
mkdir -p cmd/api cmd/worker \
         internal/domain/shared \
         internal/domain/ordering internal/domain/catalog internal/domain/pricing \
         internal/domain/procurement internal/domain/logistics \
         internal/app internal/adapter/postgres internal/adapter/http \
         internal/adapter/openai internal/adapter/merchant internal/platform \
         config docs
```

```bash
# 4. Khởi tạo git
git init -b main
```

```bash
# 5. Thư viện ngoài đầu tiên (04/09): sinh UUID v7 cho ID của entity
go get github.com/google/uuid@v1.6.0
```

> **`go get` làm gì:** thêm dòng `require` vào `go.mod` (~ `composer.json`) và
> ghi checksum vào `go.sum` (~ `composer.lock`). **Commit cả hai file.**
> Thư viện được tải về `GOMODCACHE` dùng chung, không có `vendor/` trong repo.
> Máy khác clone về chạy `go mod download` (~ `composer install`).

---

## 3c. VS Code — cài để Ctrl+Click nhảy vào hàm (như Intelephense)

### Cần đúng hai thứ

| Thứ | Vai trò | Tương đương bên PHP |
|---|---|---|
| Extension **`golang.go`** | cầu nối giữa VS Code và Go | extension Intelephense |
| **`gopls`** | *language server* — thứ THẬT SỰ hiểu code | engine của Intelephense |

Nhiều người cài mỗi extension rồi tưởng hỏng — thiếu `gopls` thì Ctrl+Click
không chạy. Extension chỉ là vỏ.

```bash
# 1. Extension
code --install-extension golang.go

# 2. Language server + debugger  (cần PATH có Go — xem mục 2)
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

Cài xong ở `C:\Users\tony_\go\bin` (tức `GOBIN`). Thư mục này **đã có sẵn trong
PATH người dùng** nên VS Code tự tìm thấy, không phải set gì thêm.

Kiểm tra:

```bash
gopls version      # golang.org/x/tools/gopls v0.23.0
dlv version        # Delve Debugger  Version: 1.27.1
```

### Phím tắt hay dùng

| Phím | Việc | Intelephense gọi là gì |
|---|---|---|
| **Ctrl + Click** / `F12` | nhảy vào định nghĩa | Go to Definition |
| `Alt + ←` | quay lại chỗ cũ | Go Back |
| `Ctrl + Shift + F12` | tìm **mọi nơi đang gọi** hàm này | Find All References |
| `Shift + F12` | như trên, xem ngay tại chỗ | Peek References |
| `Ctrl + T` | tìm nhanh theo tên hàm/kiểu toàn project | Go to Symbol in Workspace |
| `Ctrl + Shift + O` | nhảy tới symbol trong file đang mở | Go to Symbol in File |
| `F2` | đổi tên **toàn project**, an toàn | Rename Symbol |
| `Ctrl + .` | gợi ý sửa nhanh (thêm import, tách biến…) | Quick Fix |
| `Ctrl + Space` | gợi ý code | Autocomplete |
| `Ctrl + K, Ctrl + I` | xem tài liệu của thứ đang trỏ | Hover |

> **Riêng của Go, không có bên PHP:** `Ctrl + Click` vào một **interface** sẽ
> hỏi *"đi tới định nghĩa hay đi tới các cài đặt?"* — vì Go không có
> `implements`, đây là cách duy nhất tìm ra struct nào thoả interface đó.

### Việc `gopls` tự làm mỗi khi lưu file

Cấu hình nằm ở `.vscode/settings.json` (đã tạo, và **`.gitignore` bỏ qua thư
mục `.vscode/`** nên không lên repo):

| Lưu file là nó tự | Tương đương gõ tay |
|---|---|
| format lại code | `gofmt -w .` |
| sắp xếp + xoá import thừa | *(PHP không có)* |
| chạy `go vet` cho package | `go vet ./…` |
| chạy `staticcheck` | `phpstan` |
| hiện lỗi biên dịch ngay khi gõ | — |

Có bật thêm **inlay hints** — VS Code chèn chữ mờ hiển thị tên tham số và kiểu
biến ngay trong dòng code. Rất hợp lúc mới học, vì thấy được `int64`, `Weight`,
`Money` mà không phải hover. Muốn tắt: sửa `go.inlayHints.*` thành `false`.

### Kiểm chứng không cần mở VS Code

`gopls` chạy được từ terminal — chính là thứ Ctrl+Click gọi phía sau:

```bash
# "Ctrl+Click tại dòng 29, cột 23 của weight_test.go"
gopls definition internal/domain/shared/weight_test.go:29:23
# → D:\portage\internal\domain\shared\weight.go:154:6-22: defined here as
#   func shared.ChargeableWeight(...) shared.Weight

# quét lỗi cả package, như Intelephense quét project
gopls check ./internal/domain/shared/*.go
```

---

## 3b. Git — tách danh tính cá nhân khỏi danh tính công ty

### Vấn đề

Trên máy này, cấu hình **global** đang là email công ty:

```
user.email = sy_duong@fastboy.net
user.name  = Sy Duong
```

Và các repo công ty (`gci-payment-portal-backend`, `gci-payment-backend`)
**không set local** — chúng ăn theo global.

> ⚠️ **Vì vậy TUYỆT ĐỐI không sửa global.** Sửa global là commit ở repo công ty
> sẽ mang email cá nhân. Global cứ để nguyên là email công ty.

### Cách làm: set `--local` cho từng repo cá nhân

```bash
cd /d/portage
git config --local user.email "duongsy1920@gmail.com"
git config --local user.name  "Sy Duong"
git config --local core.autocrlf input     # Go dùng LF, tránh CRLF của Windows
```

`--local` ghi vào `D:\portage\.git\config`, **chỉ có hiệu lực trong repo này**.
Nó thắng global. Repo khác không bị ảnh hưởng.

### Kiểm chứng — đừng tin, hãy kiểm tra

```bash
# Email git SẼ dùng trong repo hiện tại
git config user.email

# Email THẬT SỰ đã đi vào commit
git log -1 --format='%an <%ae>'
```

Quét nhanh mọi repo trên ổ D để chắc không nhầm chỗ nào:

```bash
for r in /d/*/; do
  if git -C "$r" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    printf "%-38s -> %s\n" "$(basename $r)" "$(git -C $r config user.email)"
  fi
done
```

Kết quả đúng phải là:

```
gci-payment-portal-backend  -> sy_duong@fastboy.net     (công ty)
gci-payment-backend         -> sy_duong@fastboy.net     (công ty)
portage                     -> duongsy1920@gmail.com    (cá nhân)
```

### Lỡ commit nhầm email rồi thì sao?

```bash
# Sửa commit gần nhất (CHƯA push)
git config --local user.email "duongsy1920@gmail.com"
git commit --amend --reset-author --no-edit

# Kiểm tra lại
git log -1 --format='%an <%ae>'
```

> Đã push rồi thì phải `git rebase` viết lại lịch sử — rắc rối và ảnh hưởng
> người khác. Nên **luôn set local NGAY sau `git init`**, trước commit đầu tiên.

### Nâng cao: tự động theo thư mục (`includeIf`)

Nếu sau này bạn gom hết project cá nhân vào một thư mục, ví dụ `D:\personal\`,
thì để git tự chọn danh tính, khỏi phải nhớ set tay mỗi lần.

Tạo `~/.gitconfig-personal`:

```ini
[user]
    email = duongsy1920@gmail.com
    name  = Sy Duong
```

Rồi thêm vào cuối `~/.gitconfig`:

```ini
[includeIf "gitdir/i:D:/personal/"]
    path = ~/.gitconfig-personal
```

Từ đó mọi repo nằm dưới `D:\personal\` tự dùng email cá nhân.
Dấu `/i` là không phân biệt hoa thường (cần trên Windows), và **dấu `/` cuối
đường dẫn là bắt buộc**.

Hiện tại chưa dùng cách này vì project cá nhân đang nằm rải rác ở `D:\`.

---

## 4. Cấu trúc thư mục — và vì sao chia như vậy

```
D:\portage
├── go.mod                      # ~ composer.json
├── go.sum                      # ~ composer.lock  (checksum thư viện)
├── README.md
├── .editorconfig               # LF, tab cho Go — mọi editor đọc file này
├── .github/workflows/ci.yml    # CI: gofmt, vet, test -race, kiểm tra import domain
├── cmd/                        # nơi để hàm main() — mỗi thư mục 1 chương trình
│   ├── api/                    #   ✅ web server: `go run ./cmd/api` (in-memory + relay trong process) hoặc `-dsn …` (Postgres)
│   └── worker/                 #   ✅ relay outbox + mọi subscriber + sweep quote hết hạn: `go run ./cmd/worker -dsn …` (bắt buộc -dsn)
├── internal/                   # Go CẤM module khác import thư mục tên `internal`
│   ├── domain/                 # TRÁI TIM — nghiệp vụ thuần, KHÔNG import gì bên ngoài
│   │   ├── shared/             #   Shared Kernel: Money, Weight, Rate, ID, Events
│   │   ├── catalog/            #   ✅ Merchant, Product ⊃ Variant, CategoryPolicy, Provenance — 15 event
│   │   ├── pricing/            #   ✅ ShippingLane, RateCard, DutyPolicy, QuotePolicy, Calculate, Quote — 3 event
│   │   ├── ordering/           #   ✅ CustomerOrder (cọc 50 %, Cancel/Refund), AcceptedQuote — 8 event
│   │   ├── procurement/        #   ✅ PurchaseTask, PurchaseReceipt, Shop/Item projection, PORT MerchantACL — 3 event
│   │   └── logistics/          #   ✅ Parcel, ConsolidationBatch, FreightAllocator (Domain Service), LaneRule — 5 event
│   ├── contracts/              # ✅ PUBLISHED LANGUAGE — DTO *V1 đi qua ranh giới context (không domain, không adapter)
│   ├── app/                    # Use case / điều phối — gọi domain, gọi repository
│   │   ├── ports.go            #   ✅ Clock, UnitOfWork, Outbox — cái tầng app CẦN
│   │   ├── deps.go             #   ✅ MustHave — kiểm dây nối lúc khởi động
│   │   └── <context>/          #   MỘT thư mục con cho MỖI bounded context
│   │       ├── catalog/        #   ✅ package `catalogapp` — 7 use case (tên khác thư mục để không trùng domain)
│   │       ├── pricing/        #   ✅ package `pricingapp` — IssueQuote, AcceptQuote, Projector (nghe catalog)
│   │       ├── ordering/       #   ✅ package `orderingapp` — PlaceOrder, Pay*, Cancel, lifecycle, Projector (nghe pricing), Reactor (nghe procurement)
│   │       ├── procurement/    #   ✅ package `procurementapp` — OpenTask (hỏi ACL), ConfirmTask, FailTask, Projector (nghe catalog)
│   │       ├── logistics/      #   ✅ package `logisticsapp` — ExpectParcel (nghe procurement), Receive, Open/Add/Close/ShipBatch, Projector (nghe pricing.lane_defined)
│   │       └── reporting/      #   ✅ package `reportingapp` — READ MODEL: order_summaries, KHÔNG có domain (không invariant nào)
│   │                           #   không để phẳng — sẽ phình thành một package 50 file
│   ├── adapter/                # THẾ GIỚI BÊN NGOÀI cắm vào đây
│   │   ├── http/               #   ✅ package `httpapi` — 37 route (5 context + read model + /lanes + /fx + /tokens), CẢ mux sau middleware auth, decode + locale, bảng lỗi → status
│   │   ├── memory/             #   ✅ repository (catalog + pricing) + outbox + unit-of-work in-memory (test & dev run)
│   │   ├── eventcodec/         #   ✅ Encode: event → JSON viết tay (hợp đồng); Decode: JSON → contracts.*V1; guard go/ast mọi context
│   │   ├── postgres/           #   ✅ pgx: repo qua Snapshot, UnitOfWork tx-trong-ctx, outbox, migrations/0001 … 0007 (0006 api_tokens, 0007 order_summaries)
│   │   │   └── pgtest/         #   ✅ helper test tích hợp: DSN, advisory lock, truncate
│   │   ├── merchant/           #   ✅ ACL cho shop: `Manual` (một con người) + `Router` chọn adapter theo shop — thêm shop = một dòng trong wire
│   │   └── openai/             #   ✅ ACL thứ hai: đọc trang shop bằng model → ListingDraft; `Fake` (dev), `Unavailable` (không có key → 503)
│   ├── worker/                 # ✅ Relay (đọc outbox → publish → MarkSent, at-least-once) + Bus + Sweeper (việc theo GIỜ, không theo event)
│   └── platform/               # config, logger, connection pool
│       ├── auth/               #   ✅ Principal + port Verifier/Issuer/Registry + Static (RAM); "ai đang gọi" là việc của BIÊN, không phải domain
│       ├── clock/              #   ✅ System (time.Now) và Fixed (test) — cài đặt app.Clock
│       └── wire/               #   ✅ Memory() / Postgres() → Graph{5 context + Reporting, Source, Auth/Tokens/Registry, Extractor}; Subscribe() = bảng định tuyến 31 dòng = vòng đời §31
├── scripts/smoke.ps1           # ✅ cả flow trên binary thật + Postgres thật (api + worker nền, bearer token thật, 8 request)
├── scripts/smoke.sh            # ✅ bản Linux của đúng script đó — CHẠY TRONG CI, và assert số vàng chứ không chỉ in ra
├── config/                     # bảng giá, thuế suất — DỮ LIỆU, không phải code (hiện là hằng trong wire.quotePolicy)
├── docker-compose.yml          # ✅ postgres:16, db `portage` (dev) + `portage_test` (test)
├── docker/initdb/              # ✅ SQL chạy một lần khi volume trống: tạo portage_test
└── docs/                       # SETUP (file này) · HOC (lộ trình học) · WALKTHROUGH (đọc code) · DDD (khái niệm) · FLOW-ORDER (một đơn) · CATALOG · GO-CHO-PHP · P9-PLAN
```

### Luật vàng: chiều mũi tên phụ thuộc

```
   cmd  ──▶  app  ──▶  domain
                ▲
   adapter ─────┘         domain KHÔNG BAO GIỜ trỏ ra ngoài
```

**`internal/domain/` không được import:**

- driver database
- framework HTTP
- SDK của OpenAI
- bounded context khác

**Được import:** thư viện chuẩn của Go, và thư viện *tiện ích thuần* trong
allowlist — hiện chỉ có `github.com/google/uuid` (tương đương `symfony/uid`).
Tiện ích ≠ framework: nó không biết DB, không biết HTTP, không có side effect.
Allowlist nằm trong `.github/workflows/ci.yml`; thêm gì phải ghi rõ lý do.

Lỡ import → test domain sẽ cần Docker, cần API key, chạy chậm.
Làm đúng → test domain chạy trong **vài mili-giây**, không cần gì cả.
CI **tự fail** nếu domain import thứ ngoài allowlist (xem 5b).

> **So với Symfony:** ở Symfony, `Entity` dính chặt Doctrine — `#[ORM\Column]`
> nằm ngay trong class nghiệp vụ. Ở đây thì ngược lại: struct nghiệp vụ hoàn
> toàn sạch, phần map xuống DB nằm riêng ở `adapter/postgres`.

### Vì sao tên là `internal/`?

Đây là quy ước **được chính trình biên dịch Go ép buộc**: package nằm dưới thư
mục tên `internal` chỉ import được bởi code trong cùng module. Không cần tài
liệu nhắc, không cần reviewer canh — compiler chặn thẳng.

---

## 5. Lệnh dùng hàng ngày

| Việc cần làm | Lệnh | Tương đương bên PHP |
|---|---|---|
| Chạy toàn bộ test | `go test ./...` | `vendor/bin/phpunit` |
| Test kèm chi tiết | `go test -v ./...` | `phpunit --testdox` |
| Test một package | `go test ./internal/domain/shared` | `phpunit tests/Shared` |
| Đo độ phủ test | `go test -cover ./...` | `phpunit --coverage-text` |
| Soi lỗi tĩnh | `go vet ./...` | `phpstan analyse` |
| Format code | `gofmt -w .` | `php-cs-fixer fix` |
| Biên dịch ra .exe | `go build ./cmd/api` | (PHP không có) |
| Chạy thẳng không build | `go run ./cmd/api` | `php bin/console …` |
| Thêm thư viện | `go get <module>` | `composer require` |
| Dọn go.mod cho gọn | `go mod tidy` | `composer update` |
| Tải thư viện sau khi clone | `go mod download` | `composer install` |
| Xem cấu hình | `go env` | `php -i` |

> `./...` nghĩa là "thư mục hiện tại và **mọi** thư mục con". Nhớ dấu ba chấm.

### 5b. CI — `.github/workflows/ci.yml`

Mỗi lần push lên GitHub, máy Linux của GitHub chạy 4 bước theo thứ tự:

| Bước | Lệnh | Bắt lỗi gì |
|---|---|---|
| gofmt | `gofmt -l .` phải rỗng | quên format |
| vet | `go vet ./...` | lỗi tĩnh (printf sai kiểu, copy lock, …) |
| test | `go test -race -count=1 -cover ./...` | test fail + **data race** (cái Windows không chạy được) |
| dependency rule | `go list -f '{{range .Imports}}…' ./internal/domain/...` | domain import thứ ngoài allowlist |

Bước cuối là **kiến trúc được kiểm tra bằng máy**: không ai phải nhớ luật
"domain không import driver DB" — lỡ vi phạm là build đỏ.

> **So với Symfony:** tương đương GitHub Actions chạy `php-cs-fixer --dry-run`,
> `phpstan`, `phpunit`, và `deptrac` (deptrac chính là bước dependency rule).

### 5d. Chạy API thật và bấm thử

```bash
# In-memory — không cần gì, Ctrl+C là mất dữ liệu
go run ./cmd/api                 # :8080
go run ./cmd/api -addr :9090     # đổi cổng

# Postgres — cần Docker Desktop đang chạy
docker compose up -d             # lần đầu kéo postgres:16 (~150 MB); có healthcheck
go run ./cmd/api    -dsn "postgres://portage:portage@localhost:5432/portage?sslmode=disable"
go run ./cmd/worker -dsn "postgres://portage:portage@localhost:5432/portage?sslmode=disable" -every 500ms
# worker còn quét quote quá 48 h (-sweep, mặc định 1 phút; -sweep 0 để tắt, -sweep-batch đổi số quote mỗi lượt)
go run ./cmd/worker -dsn "$DSN" -every 500ms -sweep 10s -sweep-batch 50
# terminal 3: curl như dưới → worker in ra từng event kèm payload
```

Xem dữ liệu thật: `docker compose exec postgres psql -U portage -d portage -c "select id, event_name, sent_at from outbox order by id;"`

Nhanh nhất — cả flow trên binary thật, tự dọn: `powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1`
(cần Docker; build `bin\api.exe` + `bin\worker.exe`, chạy nền, 8 request, in quote + worker log, tắt).

Bằng tay, theo đúng thứ tự nghiệp vụ (Git Bash / WSL; PowerShell dùng `curl.exe`
và escape khác — xem transcript ở WALKTHROUGH.md §1a):

**Từ 06/09 mọi route đều cần `Authorization: Bearer <token>`** (WALKTHROUGH §18).
Chạy memory thì token cố định; chạy Postgres thì chìa đầu tiên do
`-bootstrap-operator-token` cắt, chìa sau qua `POST /tokens`:

```bash
# Memory (go run ./cmd/api) — hai token dev in sẵn trong wire.go
OPTOK=dev-operator
CUSTOK=dev-customer
# khách thứ hai thì cắt thêm chìa (xem POST /tokens dưới) — memory hay Postgres đều vậy

# Postgres — chìa đầu tiên, CHỈ tạo được khi bảng api_tokens còn rỗng
OPTOK=$(openssl rand -hex 32)
go run ./cmd/api -dsn "$DSN" -bootstrap-operator-token "$OPTOK"
# rồi operator cắt chìa cho khách; token trả về MỘT LẦN, kho chỉ giữ hash
CUSTOK=$(curl -s -X POST localhost:8080/tokens -H "Authorization: Bearer $OPTOK" \
  -H 'Content-Type: application/json' -d '{"kind":"customer","label":"app"}' | jq -r .token)
# một khách THỨ HAI, để thấy 404-chứ-không-403 ở bước 6
CUSTOK2=$(curl -s -X POST localhost:8080/tokens -H "Authorization: Bearer $OPTOK" \
  -H 'Content-Type: application/json' -d '{"kind":"customer","label":"khach-khac"}' | jq -r .token)

# Không token → 401; đúng token sai loại → 403
curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8080/categories                              # → 401
curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8080/categories -H "Authorization: Bearer $CUSTOK"  # → 403
```

```bash
OP="Authorization: Bearer $OPTOK"
CU="Authorization: Bearer $CUSTOK"
CU2="Authorization: Bearer $CUSTOK2"

# 1. Đăng ký shop — dấu phẩy trong "50,00" là dấu THẬP PHÂN vì Accept-Language: vi
curl -s -i -X POST localhost:8080/merchants -H "$OP" -H 'Content-Type: application/json' -H 'Accept-Language: vi'   -d '{"name":"Example Sports","site":"www.example.com","currency":"USD",
       "free_shipping":{"kind":"over","threshold":"50,00"},"sourcing":["operator"]}'
# → 201 {"id":"01a0…"}          ← chép id này

# 2. Sản phẩm KHÁCH dán — không còn field "sourced_by": provenance suy từ TOKEN
curl -s -i -X POST localhost:8080/products -H "$CU" -H 'Content-Type: application/json'   -d '{"name":"Air Trainer 90","merchant_id":"<ID Ở TRÊN>","category":"footwear",
       "source_url":"https://www.example.com/t/air-trainer-90/abc","price":"150.00","currency":"USD"}'
# → 201 {"id":"01a0…"}   (gọi bằng $OP thay $CU thì listing đã verified luôn)

# 3. Publish — chưa có variant nên nghiệp vụ từ chối
curl -s -i -X POST localhost:8080/products/<ID SẢN PHẨM>/publish -H "$OP"
# → 409 {"error":{"code":"no_variants","message":"publish "Air Trainer 90": product has no variants"}}

# 4. Ba bước operator — không còn X-Operator-ID: operator LÀ chủ token
P=<ID SẢN PHẨM>
curl -s -i -X POST localhost:8080/products/$P/variants -H "$OP" -H 'Content-Type: application/json' -d '{"size":"US 9","color":"black"}'   # → 201
curl -s -i -X POST localhost:8080/products/$P/confirm-listing -H "$OP"                                                                      # → 204
curl -s -i -X POST localhost:8080/products/$P/measure -H "$OP" -H 'Content-Type: application/json' \
     -d '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}'                                                                       # → 204, outbox: product_measured
curl -s -i -X POST localhost:8080/products/$P/publish -H "$OP"                                                                                    # → 204, outbox: product_published

# 5. Báo giá — NGAY LẬP TỨC sau publish có thể là 404 listing_not_found: worker chưa relay (in-memory: đợi 200 ms)
curl -s -i -X POST localhost:8080/quotes -H "$CU" -H 'Content-Type: application/json' -d "{\"product_id\":\"$P\",\"lane\":\"us_forwarder\"}"
# → 201 {"id":"01a0…"}          ← Q   (hỏi giá thì token nào cũng được: $OP hay $CU)
curl -s localhost:8080/quotes/<Q> -H "$CU"
# → {"status":"issued","class":"branded","chargeable_g":2500,"estimated":false,
#    "lines":{"item":{"amount":"150.00","currency":"USD"},"sales_tax":{"amount":"13.22",…},"freight":{"amount":"25.00",…},…,"subtotal":{"amount":"188.22",…}},
#    "fx":"26000","home":{"subtotal":{"amount":"4893720","currency":"VND"},"service_fee":{"amount":"500000",…},"total":{"amount":"5393720",…},"deposit":{"amount":"2696860",…}}}
curl -s -i -X POST localhost:8080/quotes/<Q>/accept -H "$CU"                                                    # → 204; operator gọi → 403 (chấp giá là việc của KHÁCH); lần 2 → 409 quote_not_issued

# 6. Đặt hàng (05/09, chiều) — ordering chỉ biết quote sau khi worker relay pricing.quote_accepted (in-memory: 200 ms)
V=<ID VARIANT ở bước 4>   # không còn customer_id trong body: chủ đơn LÀ chủ token $CU
curl -s -i -X POST localhost:8080/orders -H "$CU" -H 'Content-Type: application/json' -d "{\"quote_id\":\"<Q>\",\"variant_id\":\"$V\"}"   # → 201 {"id"} ← O  (sớm quá → 409 quote_not_accepted)
curl -s -i -X POST localhost:8080/orders/<O>/deposit -H "$OP" -H 'Content-Type: application/json' -H 'Accept-Language: vi' -d '{"amount":"2.696.860","currency":"VND"}'   # → 204 (sai số → 409 wrong_amount)
curl -s localhost:8080/orders/<O> -H "$CU"     # → {"status":"deposited","total":{"amount":"5393720",…},"deposit":…,"balance":{"amount":"2696860",…},"balance_paid":false,…}
curl -s localhost:8080/orders/<O> -H "$CU2"    # → 404 order_not_found với token của khách KHÁC — không phải 403: 403 sẽ xác nhận đơn đó có thật (§18d)
curl -s -i -X POST localhost:8080/orders/<O>/cancel -H "$CU" -H 'Content-Type: application/json' -d '{"reason":"đổi ý"}'                            # → 204; GET thấy "refund":{"amount":"2696860"} — chưa mua → hoàn đủ

# 7. Đi mua (05/09, chiều muộn) — KHÔNG huỷ ở bước 6 mà để worker relay deposit_paid → procurement mở task
curl -s localhost:8080/purchase-tasks -H "$OP"                                                           # → [{"id":"<T>","order_id":"<O>","currency":"USD","status":"open",…}]  (rỗng = relay chưa chạy)
curl -s -i -X POST localhost:8080/purchase-tasks/<T>/confirm -H "$OP" -H 'Content-Type: application/json' \
     -d '{"reference":"NK-20260905-001","paid":"163.22","currency":"USD"}'                                # → 204 (khách gọi → 403; VND → 409 paid_currency)
# hoặc: curl -s -i -X POST localhost:8080/purchase-tasks/<T>/fail -H "$OP" -d '{"reason":"hết size US 10"}'   # → 204 → đơn về purchase_failed, huỷ sẽ hoàn ĐỦ
curl -s localhost:8080/orders/<O> -H "$CU"                                                               # → "status":"purchased" sau khi worker relay purchase_confirmed — ĐIỂM KHÔNG THỂ QUAY ĐẦU
curl -s -i -X POST localhost:8080/orders/<O>/cancel -H "$CU" -H 'Content-Type: application/json' -d '{"reason":"đổi ý sau khi đã mua"}'             # → 204; GET: "refund":{"amount":"0"},"forfeited":true

# 8. Kho (06/09) — KHÔNG huỷ ở bước 7; purchase_confirmed cũng bảo kho chờ một thùng
curl -s localhost:8080/parcels -H "$OP"                                                                   # → [{"id":"<P>","order_id":"<O>","reference":"NK-…","status":"expected"}]
curl -s -i -X POST localhost:8080/parcels/<P>/receive -H "$OP" -H 'Content-Type: application/json' \
     -d '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}'                                # → 204 — CÂN THẬT (Actual của §28)
curl -s -i -X POST localhost:8080/batches -H "$OP" -H 'Content-Type: application/json' -d '{"lane":"us_forwarder"}'   # → 201 {"id"} ← B  (lane lạ → 404 lane_rule_not_found)
curl -s -i -X POST localhost:8080/batches/<B>/parcels -H "$OP" -H 'Content-Type: application/json' -d '{"parcel_id":"<P>"}'   # → 204
curl -s -i -X POST localhost:8080/batches/<B>/close -H "$OP"                                              # → 204 (rỗng → 409 batch_empty)
curl -s -X POST localhost:8080/batches/<B>/ship -H "$OP" -H 'Content-Type: application/json' -d '{"freight":"27.50","currency":"USD"}'
# → 200 [{"order_id":"<O>","chargeable_g":2500,"freight":{"amount":"27.50","currency":"USD"}}]   ← phần chia cước, đóng băng vào batch_shipped
curl -s localhost:8080/orders/<O> -H "$CU"                                                               # → "status":"in_transit" sau khi worker relay batch_shipped
curl -s -i -X POST localhost:8080/orders/<O>/balance -H "$OP" -H 'Content-Type: application/json' -d '{"amount":"2696860","currency":"VND"}'   # → 204
curl -s -i -X POST localhost:8080/orders/<O>/deliver -H "$OP"                                            # → 204 → "delivered"
curl -s localhost:8080/reconciliations/<O> -H "$OP"

# 9. Màn hình (06/09, P9/T2) — MỘT bảng phẳng dựng từ event của năm context
curl -s "localhost:8080/me/orders" -H "$CU"                    # khách: đơn của CHÍNH MÌNH, mới nhất trước; không có shop_reference
curl -s "localhost:8080/orders?status=deposited" -H "$OP"      # operator: hàng đợi việc (gõ sai status → 400 invalid_order, KHÔNG phải list rỗng)

# 10. Tỷ giá + AI + quản chìa (06/09, P9/T4-T5)
curl -s -i -X POST localhost:8080/fx -H "$OP" -H 'Content-Type: application/json' -d '{"from":"USD","to":"VND","rate":"26000"}'   # → 204, KHÔNG phát event
curl -s -i -X POST localhost:8080/products/from-url -H "$CU" -H 'Content-Type: application/json' \
     -d "{\"merchant_id\":\"<ID SHOP>\",\"url\":\"https://www.example.com/t/x\",\"category\":\"footwear\"}"   # memory → Fake; Postgres không OPENAI_API_KEY → 503 extractor_unavailable
curl -s localhost:8080/tokens -H "$OP"                          # danh sách chìa: hash + kind + label + active (KHÔNG BAO GIỜ có token)
curl -s -i -X DELETE localhost:8080/tokens/<HASH> -H "$OP"      # → 204; chìa operator cuối cùng → 409 last_operator_key
# → {"complete":true,"quoted":{"goods":{"amount":"163.22"},"freight":{"amount":"25.00"},"chargeable_g":2500},
#    "actual":{"goods":{"amount":"163.22"},"freight":{"amount":"27.50"},"chargeable_g":2500},"variance":{"amount":"-2.50","currency":"USD"}}
```

Toàn bộ đường đi, từng file: WALKTHROUGH.md §14 (pricing), §15 (ordering), §16 (procurement), §17 (logistics + Quote vs Actual).

Thử phá: bỏ `Accept-Language` ở bước 1 → threshold thành **5000.00 USD** (en:
dấu phẩy là hàng nghìn). Gõ `"nmae"` thay `"name"` → 400 `bad_json`. Gửi lại
đúng request 2 → 201 nhưng sản phẩm mới bị **cắm cờ nghi trùng** (không thấy qua
API — chưa có GET; thấy trong test `TestAddProduct_flagsSuspectedDuplicate`).

| Status | Nghĩa | Ví dụ `code` |
|---|---|---|
| 400 | request không hiểu được → sửa request | `bad_json`, `malformed_amount`, `invalid_id`, `invalid_hostname`, `unknown_principal_kind`, `invalid_lane_code`, `incomplete_parcel_spec`, `empty_reason`, `empty_reference` |
| 401 / 403 | **ai đang gọi** → 401 tôi không biết anh; 403 tôi biết anh, cái này không phải của anh (§18d) | `unauthenticated`, `forbidden` |
| 404 | trỏ vào thứ không tồn tại | `merchant_not_found`, `category_not_found`, `product_not_found`, `lane_not_found`, `quote_not_found`, `listing_not_found` (chưa publish **hoặc** relay chưa chạy), `order_not_found`, `purchase_task_not_found`, `parcel_not_found`, `batch_not_found`, `lane_rule_not_found`, `reconciliation_not_found` |
| 409 | request đúng, **nghiệp vụ từ chối** → sửa trạng thái | `no_variants`, `unverified`, `not_draft`, `duplicate_variant`, `merchant_inactive`, `price_currency`, `no_exchange_rate`, `quote_expired`, `quote_not_issued`, `quote_not_accepted`, `quote_already_used`, `wrong_amount`, `not_awaiting_deposit`, `not_in_transit`, `balance_unpaid`, `already_delivered`, `order_cancelled`, `purchase_task_not_open`, `paid_currency`, `parcel_not_expected`, `parcel_not_received`, `batch_not_open`, `batch_not_closed`, `batch_empty`, `duplicate_parcel`, `invalid_freight` |
| 503 | request đúng, **thứ giúp mình thì không có** → thử lại sau, hoặc gõ tay | `extractor_unavailable` |
| 500 | bug hoặc sự cố; chi tiết chỉ ở log | `internal` |

Bảng đầy đủ: `internal/adapter/http/errors.go` — một `errorTable`, một `writeError`.

### 5e. Test tích hợp với Postgres

```bash
docker compose up -d
export PORTAGE_TEST_DSN="postgres://portage:portage@localhost:5432/portage_test?sslmode=disable"   # Git Bash
# PowerShell:  $env:PORTAGE_TEST_DSN = "postgres://portage:portage@localhost:5432/portage_test?sslmode=disable"
go test -count=1 -timeout 90s ./...
```

| Không có biến | Có biến |
|---|---|
| `internal/adapter/postgres`, `platform/wire` **tự skip** (`t.Skip`) — unit test không bao giờ cần Docker | chạy thật: connect → migrate → truncate → test |

Ba điều đã học bằng cách vấp:

1. **`go test ./...` chạy các package song song** — hai package cùng TRUNCATE một DB
   thì đỏ 1/3 lần. Không dùng `-p 1` (che bệnh); `pgtest.Pool(t)` giữ **advisory lock**
   của Postgres trên một connection riêng suốt test → package nào có khoá thì sở hữu DB,
   package khác chờ. Một DB, song song tuỳ ý.
2. **Luôn truyền `-timeout`.** Test treo mà không có timeout thì `go test` đợi 10 phút.
3. **Giết ssh/terminal không giết `go test`.** Process `postgres.test.exe`, `wire.test.exe`
   sống tiếp và **giữ khoá** → mọi lần chạy sau đều treo. Khi test tích hợp treo bất thường:
   ```powershell
   Get-Process | Where-Object { $_.ProcessName -match '\.test$' } | Stop-Process -Force
   docker compose exec -T postgres psql -U portage -d portage_test -c "select pid, state, query from pg_stat_activity where datname='portage_test';"
   ```

CI (`.github/workflows/ci.yml`) có `services: postgres` và set biến này → test tích hợp
luôn chạy trên GitHub.

### 5c. Bộ lệnh hay dùng — chép về dùng luôn

Những lệnh thật sự gõ hàng ngày trong project này. Nhóm theo việc cần làm.

#### Trước khi mở terminal mới

```bash
# Git Bash — nếu gõ `go` báo not found (xem mục 2)
export PATH="$PATH:/c/Program Files/Go/bin"
cd /d/portage
```

```powershell
# PowerShell
$env:Path += ";C:\Program Files\Go\bin"
cd D:\portage
```

#### Chạy test

```bash
go test ./...                              # tất cả, im lặng nếu pass
go test -v ./...                           # in tên từng test
go test ./internal/domain/shared           # một package
go test -run TestMoney_convertUSDtoVND ./...        # đúng MỘT test
go test -run 'TestMoney_.*' ./...                   # theo mẫu regex
go test -count=1 ./...                     # BỎ QUA cache, chạy lại thật
go test -failfast ./...                    # dừng ngay ở test fail đầu tiên
go test -list '.*' ./internal/domain/shared         # liệt kê tên test, không chạy
```

> ⚠️ Go **cache kết quả test**. Thấy `(cached)` nghĩa là nó không chạy lại vì
> code không đổi. Muốn ép chạy thật: `-count=1`.

#### Đo độ phủ test (coverage)

```bash
go test -cover ./...                       # số tổng
go test -coverprofile=cover.out ./internal/domain/shared

go tool cover -func=cover.out              # % theo TỪNG HÀM
go tool cover -func=cover.out | awk '$3 != "100.0%"'   # chỉ hàm chưa đủ
go tool cover -html=cover.out              # mở trình duyệt, tô màu dòng chưa test
```

`cover.out` đã nằm trong `.gitignore` (khớp `*.out`), không lo commit nhầm.

#### Kiểm tra chất lượng — chạy đủ bộ trước khi commit

```bash
gofmt -l .                                 # LIỆT KÊ file chưa format (rỗng = sạch)
gofmt -w .                                 # SỬA luôn
go vet ./...                               # soi lỗi tĩnh
go build ./...                             # compile hết, không tạo file
go test ./...
```

Gộp một dòng, dừng ngay khi có lỗi — chạy trước mỗi lần commit:

```bash
gofmt -l . && go vet ./... && go test -count=1 ./...
```

#### Kiểm tra Dependency Rule (luật kiến trúc quan trọng nhất)

```bash
# Domain đang import những gì?
go list -f '{{range .Imports}}{{println .}}{{end}}' ./internal/domain/... | sort -u

# Có gì ngoài allowlist không?  (rỗng = sạch)
go list -f '{{range .Imports}}{{println .}}{{end}}' ./internal/domain/... \
  | sort -u | grep '\.' | grep -Ev '^github\.com/google/uuid$'
```

> Chạy lệnh này **mỗi khi thêm import vào domain**. CI cũng chạy đúng lệnh đó
> (mục 5b) — chạy trước ở máy thì không phải chờ build đỏ mới biết.
>
> ⚠️ Đừng dùng `go list -deps` — nó in cả dependency **gián tiếp** (`runtime`,
> `os`, `syscall`… do `fmt` kéo theo) làm tưởng domain bẩn. `.Imports` mới là
> import trực tiếp.

#### Quản lý thư viện

```bash
go get github.com/xxx/yyy@v1.2.3           # thêm, ghim đúng phiên bản
go get -u ./...                            # nâng cấp (cẩn thận)
go mod tidy                                # dọn require thừa, bổ sung thiếu
go mod download                            # tải về sau khi clone
go list -m all                             # cây thư viện đang dùng
go mod why github.com/google/uuid          # VÌ SAO thư viện này có mặt
```

Sau `go get` nhớ commit **cả `go.mod` và `go.sum`**.

#### Đọc tài liệu ngay trong terminal

```bash
go doc ./internal/domain/shared                    # tất cả hàm exported
go doc ./internal/domain/shared Money              # một kiểu
go doc ./internal/domain/shared Money.Convert      # một method
go doc -all ./internal/domain/shared | less        # kèm toàn bộ comment
```

Đây là lý do comment trong code viết cẩn thận: `go doc` biến chúng thành tài
liệu, không cần công cụ ngoài.

#### Dọn dẹp khi thấy lạ

```bash
go clean -testcache                        # xoá cache test (khi nghi test cũ)
go clean -modcache                         # xoá cache thư viện (nặng, ít dùng)
go env GOMODCACHE                          # xem thư viện tải về nằm đâu
```

#### Git — dùng hàng ngày

```bash
git status --short                         # gọn hơn `git status`
git diff                                   # thay đổi chưa `add`
git diff --staged                          # đã `add`, chưa commit
git log --oneline | cat                    # lịch sử một dòng (| cat để khỏi kẹt pager)
git log -1 --format='%an <%ae>'            # KIỂM CHỨNG email của commit cuối

git add -A
git commit -m "..."
```

> ⚠️ Trong Git Bash, `git log` mở pager và **treo terminal**. Thêm `| cat`
> hoặc `--no-pager`: `git --no-pager log --oneline`.

#### Kiểm tra danh tính git đúng repo (mục 3b)

```bash
git config user.email                      # email repo hiện tại sẽ dùng

# quét mọi repo trên ổ D
for r in /d/*/; do
  if git -C "$r" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    printf "%-38s -> %s\n" "$(basename $r)" "$(git -C $r config user.email)"
  fi
done
```

#### Chạy chương trình (khi đã có `cmd/`)

```bash
go run ./cmd/api                           # chạy thẳng, không tạo file
go build -o bin/api.exe ./cmd/api          # build ra .exe
go build ./...                             # compile hết để kiểm tra
```

#### Bảng tra nhanh — cờ hay quên

| Cờ | Nghĩa |
|---|---|
| `./...` | thư mục này **và mọi thư mục con** |
| `-v` | in chi tiết |
| `-run <regex>` | chỉ chạy test khớp tên |
| `-count=1` | bỏ qua cache, chạy thật |
| `-race` | dò data race (**cần gcc — Windows không chạy được**, CI lo) |
| `-cover` | đo độ phủ |
| `-failfast` | dừng ở lỗi đầu tiên |
| `-short` | bỏ qua test dài (`testing.Short()`) |

---

## 6. Đã viết được gì (tính tới 08/09/2026)

```
internal/domain/
├── decisions_test.go     7 test canh quyết định (go/ast) — xem mục dưới
│
├── ../contracts/         catalog_v1.go — PUBLISHED LANGUAGE: MoneyV1, ParcelV1, ProductPublishedV1, …, CategoryDefinedV1
├── ../app/ports.go       Clock, UnitOfWork, Outbox — PORT tầng app;  deps.go — MustHave
├── ../app/catalog/       package catalogapp — 14 test, coverage 86%
│   ├── deps.go           Deps (6 dependency) + mustHave; ErrMerchantInactive, ErrPriceCurrency
│   ├── register_merchant.go   RegisterMerchantHandler — dạng đơn giản nhất
│   ├── add_product.go         AddProductHandler — 2 luật liên aggregate + phát hiện trùng
│   ├── publish_product.go     PublishProductHandler — load → Publish → Save → outbox
│   ├── define_category.go     DefineCategoryHandler — Save + thông báo CategoryDefined (seed cũng đi đường này)
│   └── add_variant.go · confirm_listing.go · measure_product.go — ba bước operator để tới publish (05/09)
├── ../app/logistics/     package logisticsapp — 1 test (cả flow kho), coverage 78,3%
│   ├── deps.go           Deps (Parcels, Batches, Lanes) + mutateParcel/mutateBatch
│   ├── parcels.go        ExpectParcelHandler (OnPurchaseConfirmed, idempotent theo đơn), ReceiveParcelHandler, Projector (OnLaneDefined)
│   └── batches.go        OpenBatch (từ chối lane không rule), AddParcel (2 aggregate/1 tx), CloseBatch, ShipBatch (allocator từ LaneRule)
├── ../app/procurement/   package procurementapp — 3 test, coverage 81,5%
│   ├── deps.go           Deps (Tasks, Shops, Items, Variants, ACL) + mutate
│   ├── open_task.go      OpenTaskHandler: idempotent theo đơn, tra projection, dựng Subject, hỏi ACL (3 nhánh), OnDepositPaid
│   ├── close_task.go     ConfirmTaskHandler, FailTaskHandler
│   └── projector.go      OnMerchantRegistered (Currency), OnProductPublished (+ source), OnVariantAdded (từ điển size)
├── ../app/ordering/      package orderingapp — 4 test, coverage 70,8%
│   ├── deps.go           Deps + mutate (khuôn load → act → Save → outbox viết một lần)
│   ├── place_order.go    ba luật liên aggregate: quote đã accept (projection), một quote một đơn (ByQuote), variant có thật và của đúng sản phẩm
│   ├── payments.go · cancel_order.go · lifecycle.go   PayDeposit, PayBalance, Cancel (trả Refund), ConfirmPurchase, FailPurchase, Ship, Deliver
│   ├── projector.go      OnQuoteAccepted → AcceptedQuote (upsert); OnVariantAdded → Variant (chỉ id + chủ)
│   └── reactor.go        OnPurchaseConfirmed/OnPurchaseFailed → ConfirmPurchase/FailPurchase; nuốt "đã làm rồi"
├── ../app/pricing/       package pricingapp — 7 test, coverage 72,9%
│   ├── define_lane.go    DefineLaneHandler — Save + thông báo lane_defined (seed và POST /lanes đi đường này)
│   ├── reconciler.go     Reconciler — OnOrderPlaced (copy phần quoted từ Quote), OnPurchaseConfirmed, OnBatchShipped
│   ├── deps.go           Deps: 5 repo + Rates + Policy/Classification (CẤU HÌNH nghiệp vụ, không phải service)
│   ├── issue_quote.go    gom listing/lane/profile/fx → pricing.IssueQuote → Save → outbox
│   ├── accept_quote.go   Accept trễ: commit expiry rồi mới trả ErrQuoteExpired
│   ├── expire_quotes.go  ExpireQuotesHandler — use case ĐẦU TIÊN không ai gọi, chỉ thời gian đẩy; MỖI quote một transaction
│   └── projector.go      nghe catalog.* → Listing / CategoryProfile — idempotent, chịu sai thứ tự
├── ../adapter/merchant/  3 test, 88,9% — manual.go (luôn ErrManualPurchase) + router.go (chọn ACL theo shop; không biết shop → giao cho người)
├── ../adapter/memory/    14 test, 80,8% — repo của CẢ 6 context + Outbox (Drain, Pending/MarkSent) + UnitOfWork; mỗi repo: miss trả sentinel của DOMAIN, Save thứ hai là upsert, list có thứ tự ổn định
├── ../adapter/eventcodec/ 5 test, 98,6% — Encode 36 event + Decode → V1 (21 tên) + guard go/ast quét domain/*/events.go
├── ../adapter/postgres/  28 test tích hợp, 79,8% — round trip repo 6 context + reconciliations + api_tokens + order_summaries, rollback thật, outbox
│   ├── postgres.go       Connect, querier, txKey, db(ctx), UnitOfWork.InTx
│   ├── migrate.go        embed migrations/*.sql; MỌI thứ trong một tx sau pg_advisory_xact_lock — kể cả CREATE TABLE schema_migrations (đợt 18)
│   ├── migrations/       0001_catalog.sql (merchants, categories, products, product_variants, outbox) · 0002_pricing.sql (lanes, fx_rates, listings, category_profiles, quotes) · 0003_ordering.sql (accepted_quotes, orders UNIQUE(quote)) · 0004_procurement.sql (purchase_tasks UNIQUE("order"), procurement_shops, procurement_items) · 0005_logistics.sql (lane_rules, parcels, batches + batch_items + batch_allocations, reconciliations) · 0006_auth.sql (api_tokens: token_hash PK, kind, subject, revoked_at) · 0007_reporting.sql (order_summaries, product_names) · 0008_variant_subject.sql (procurement_variants, ordering_variants, 4 cột Subject của purchase_tasks, procurement_items.source)
│   ├── *_repo.go         upsert ON CONFLICT, scan → Snapshot → FromSnapshot
│   ├── pricing_repos.go  LaneRepo, QuoteRepo (breakdown từng cột), ListingRepo, ProfileRepo, ExchangeRates
│   ├── ordering_repos.go OrderRepo (ByID, ByQuote), AcceptedQuoteRepo, OrderingVariantRepo (tên có tiền tố vì hai package cùng có VariantRepo)
│   ├── procurement_repos.go  TaskRepo (ByID, ByOrder, Open — 4 cột Subject), ShopRepo, ItemRepo (+ source), VariantRepo
│   ├── logistics_repos.go    ParcelRepo (ByID, ByOrder, Pending), BatchRepo (bảng con ghi cùng cha), LaneRuleRepo
│   ├── reconciliation_repo.go  ReconciliationRepo — mọi cột nullable, dòng lớn dần theo event
│   ├── outbox.go         Append (cùng tx), Pending (FOR UPDATE SKIP LOCKED), MarkSent
│   ├── token_repo.go     Verify/Issue/Revoke/IsEmpty — Verify so bằng subtle.ConstantTimeCompare, Revoke ghi revoked_at chứ không DELETE
│   └── pgtest/           Pool(t): skip không DSN, advisory lock, migrate, truncate MỌI bảng — hỏi pg_tables chứ không đọc danh sách viết tay (đợt 14)
├── ../platform/wire/     4 test, 94,9%
│   ├── wire.go           Graph{Catalog, Pricing, Ordering, Procurement, Logistics, Source, Auth, Tokens}; DevTokens() cho memory, TokenRepo cho Postgres; quotePolicy(), goodsClasses(); seed categories + lane (qua DefineLane) + fx; ACL = merchant.Manual
│   └── subscribe.go      Subscribe(bus, g) — 31 dòng định tuyến = vòng đời §31; purchase_confirmed có 4 listener; on[T] decode generic
├── ../worker/            9 test, 87,3% — RunOnce thứ tự, retry at-least-once, batch, Run/cancel, Bus; Sweeper: pass lỗi vẫn chạy tiếp, chờ nhịp đầu, panic thiếu Pass
├── ../adapter/http/      package httpapi — 37 test, coverage 80,7%   (taskView có product_name/variant_label/variant_ref/source)
│   ├── server.go         NewHandler(Deps{6 context, Auth/Tokens/Registry, Extractor}) → http.Handler; 37 route, authenticate() bọc CẢ mux
│   ├── decode.go         decodeJSON (DisallowUnknownFields), language(), normalizeAmount()
│   ├── errors.go         errorTable (~58 dòng) + writeError → 400/401/403/404/409/500/503
│   ├── auth.go           authenticate() bọc CẢ mux (fail-closed), requireOperator/Customer/Any, principalOf/operatorOf/customerOf/sourcingOf
│   ├── tokens.go         POST /tokens — operator cắt chìa; token trả về ĐÚNG MỘT LẦN, kho chỉ giữ hash
│   ├── categories.go · merchants.go · products.go · operator.go   POST catalog (tham chiếu, khách, operator)
│   ├── quotes.go         POST /quotes, GET /quotes/{id} (read model quoteView), POST /quotes/{id}/accept
│   ├── orders.go         POST /orders, GET /orders/{id} (orderView, refund chỉ khi huỷ), deposit/balance/cancel/deliver
│   ├── purchase_tasks.go GET /purchase-tasks (danh sách việc), GET /purchase-tasks/{id}, confirm/fail — KHÔNG có POST tạo task
│   ├── logistics.go      GET /parcels, receive, POST /batches, add/close/ship (ship trả [allocation])
│   └── lanes.go          POST /lanes (operator), GET /reconciliations/{order} (Quote vs Actual)
│   ├── fx.go             POST /fx — route DUY NHẤT không phát event (không ai giữ bản sao tỷ giá)
│   ├── from_url.go       POST /products/from-url — AI đọc trang → nháp SourcedByFeed; không key → 503
│   └── summaries.go      GET /me/orders (khách, "me" = token), GET /orders?status= (operator); shop_reference chỉ operator thấy
├── ../platform/clock/    System, Fixed(Advance)
├── ../app/reporting/     package reportingapp — 5 test, 79,8% — READ MODEL, không có domain: OrderSummary (struct phẳng, không method),
│   ├── summary.go        OrderSummary + port OrderSummaryRepository (ByOrder/ByCustomer/ByStatus/ByProduct/All/Save) + ProductNames
│   └── projector.go      12 handler nghe CẢ NĂM context; "khung": event nào cũng được TẠO dòng → chịu sai thứ tự
├── ../adapter/openai/    5 test, 80,9% — Client (net/http thuần, json_schema strict, timeout), Fake (dev/test), Unavailable (không key → 503)
├── ../platform/auth/     10 test, 88,8% — Principal (kind+id), HashToken (sha256, không salt), Static, NewToken (crypto/rand 32 byte)
├── ../../cmd/api/main.go   -addr, -dsn → wire.Memory (+ relay trong process) | wire.Postgres → serve
├── ../../cmd/worker/main.go -dsn bắt buộc → wire.Postgres → wire.Subscribe → Relay.Run + Sweeper quét quote hết hạn (-sweep)
├── ../../scripts/smoke.ps1  cả flow trên binary thật + Postgres thật
│
├── shared/               Shared Kernel — 46 test, coverage 90,7%
│   ├── allocate.go       Allocate — chia tiền floor + largest remainder, tổng đúng (moneyphp allocate)
│   ├── money.go          Money — cộng/trừ/nhân tỷ lệ/đổi tiền
│   ├── currency.go       Currency, CurrencyFromCode
│   ├── decimal.go        parseDecimal, addExact/subExact/mulExact, divRoundHalfUp (private)
│   ├── rate.go           Rate (tỷ lệ %), ExchangeRate (tỷ giá, chính xác 12 số lẻ)
│   ├── weight.go         Weight, Dimensions, trọng lượng quy đổi thể tích
│   ├── id.go             ID — UUID v7, sinh trong domain
│   ├── event.go          Event, Events — ghi nhận domain event
│   ├── operator.go       OperatorID + ErrOperatorRequired — danh tính nhân viên, mọi context đều hỏi "ai làm" (catalog và procurement cùng dùng)
│   ├── parcelspec.go     ParcelSpec — cân nặng + hộp; DỜI từ catalog (05/09) vì pricing cũng dùng
│   │   (catalog/)        + snapshot.go — MerchantSnapshot/ProductSnapshot, FromSnapshot; 5 test round trip
│   └── *_test.go         money 16 · rate 7 · weight 9 · id 6 · event 2 · operator 1 · helpers
│
├── catalog/              Bounded context đầu tiên — 47 test, coverage 89,1%
    ├── merchant.go       Merchant (aggregate root), MerchantID, MerchantStatus, SourcingMode, Hostname
    ├── freeshipping.go   FreeShipping — VO 3 trạng thái
    ├── category.go       CategoryPolicy (VO khoá tự nhiên), CategoryCode, Restriction
│   ├── provenance.go     Provenance — nguồn + thời điểm + ai xác nhận
    ├── sourceurl.go      SourceURL — link tham chiếu, không fetch
    ├── product.go        Product (aggregate root) — draft → published → retired
    ├── variant.go        Variant — entity con của Product; khoá chống trùng bỏ MỌI khoảng trắng; unnamed() cho luật "không tên phải là duy nhất"
│   ├── events.go         16 domain event (+ VariantAdded 08/09 mang size/màu/mã shop; ProductPublished mang Price + Parcel + Source; MerchantRegistered mang Currency)
│   ├── repository.go     MerchantRepository, CategoryRepository, ProductRepository (interface)
│   └── *_test.go         merchant 14 · category 9 · freeshipping 5 · product 11 · provenance 2 · sourceurl 1
│
└── pricing/              Bounded context thứ hai (05/09) — 14 test, coverage 73,6%
    ├── errors.go         11 sentinel
    ├── lane.go           LaneCode, GoodsClass, RateCard, DutyPolicy, LaneDetails, ShippingLane, Classification
    ├── policy.go         MarginPolicy (max %, sàn), QuotePolicyDetails, QuotePolicy
    ├── listing.go        Listing, CategoryProfile — projection từ event của catalog
    ├── calc.go           QuoteInputs, Breakdown, Calculate — DOMAIN SERVICE thuần
    ├── quote.go          Quote (aggregate root) — IssueQuote, Accept, Expire, QuoteSnapshot
    ├── events.go         QuoteIssuedEvent, QuoteAcceptedEvent, QuoteExpiredEvent
    ├── repository.go     LaneRepository, QuoteRepository, ListingRepository, CategoryProfileRepository, ExchangeRates
    └── *_test.go         lane 7 · quote 7 (số vàng: 5 328 720 / 5 211 720 ₫)

└── ordering/             Bounded context thứ ba (05/09, chiều) — 4 test, coverage 81,4%
    ├── errors.go         16 sentinel — mỗi lý do từ chối một tên (+ ErrVariantUnknown, ErrVariantNotForProduct)
    ├── order.go          OrderID, OrderStatus (Status*), Refund, OrderDetails, CustomerOrder (8 method), OrderSnapshot
    ├── acceptedquote.go  AcceptedQuote — projection từ pricing.quote_accepted
    ├── variant.go        Variant{Variant, Product} — projection từ catalog.variant_added; bản sao MỎNG nhất, không giữ size
    ├── events.go         OrderPlaced, DepositPaid, OrderPurchased, OrderPurchaseFailed, OrderShipped, BalancePaid, OrderDelivered, OrderCancelled
    ├── repository.go     OrderRepository (ByID, ByQuote), AcceptedQuoteRepository, VariantRepository
    └── order_test.go     vòng đời tới delivered · bảng Cancel 5 trạng thái · snapshot từ chối 8 hình dạng

└── procurement/          Bounded context thứ tư (05/09, chiều muộn) — 4 test, coverage 80,6%
    ├── errors.go         12 sentinel, kể cả ErrManualPurchase (câu trả lời của ACL "để người mua") và ErrVariantNotFound
    ├── task.go           TaskID, TaskStatus, PurchaseReceipt (Actual đầu tiên), Subject (mua gì, bằng chữ), TaskDetails (5 field, chỉ Subject không bắt buộc), PurchaseTask (Confirm/Fail), TaskSnapshot
    ├── events.go         PurchaseTaskOpened, PurchaseConfirmed, PurchaseFailed
    ├── repository.go     Shop/Item/Variant projection (Variant.Label() → "M 8 / W 9.5 · black"); TaskRepository (ByID, ByOrder, Open), ShopRepository, ItemRepository, VariantRepository, PORT MerchantACL
    └── task_test.go      validate từng field · biên nhận gắt (4 lý do) · fail cần lý do · snapshot từ chối 7 hình dạng

└── logistics/            Bounded context thứ năm (06/09) — 2 test, coverage 76,1%
    ├── errors.go         15 sentinel
    ├── parcel.go         ParcelID, ParcelStatus, ParcelDetails, Parcel (Receive/AssignToBatch/MarkShipped), ParcelSnapshot
    ├── batch.go          BatchID, BatchStatus, BatchItem, ConsolidationBatch (AddParcel/Close/Ship), BatchSnapshot
    ├── allocator.go      Allocation, FreightAllocator (PORT/Domain Service), LaneRule, ByChargeableWeight
    ├── events.go         ParcelExpectedEvent, ParcelReceivedEvent, BatchOpened, BatchClosedEvent, BatchShippedEvent{Allocations}
    ├── repository.go     ParcelRepository (Pending), BatchRepository, LaneRuleRepository
    └── logistics_test.go vòng đời parcel · ship chia 87.50 → 29.17/58.33 · snapshot từ chối
```

Chạy kiểm tra:

```bash
cd /d/portage
go test -cover ./...
# ok  github.com/duongsy/portage/internal/adapter/eventcodec coverage: 98.6%
# ok  github.com/duongsy/portage/internal/adapter/http      coverage: 80.7%
# ok  github.com/duongsy/portage/internal/adapter/memory    coverage: 80.8%
# ok  github.com/duongsy/portage/internal/adapter/merchant  coverage: 88.9%
# ok  github.com/duongsy/portage/internal/adapter/openai    coverage: 80.9%
# ok  github.com/duongsy/portage/internal/adapter/postgres  coverage: 79.8%   (cần PORTAGE_TEST_DSN)
# ok  github.com/duongsy/portage/internal/app/catalog       coverage: 77.0%
# ok  github.com/duongsy/portage/internal/app/logistics     coverage: 78.3%
# ok  github.com/duongsy/portage/internal/app/ordering      coverage: 70.8%
# ok  github.com/duongsy/portage/internal/app/pricing       coverage: 72.9%
# ok  github.com/duongsy/portage/internal/app/procurement   coverage: 81.5%
# ok  github.com/duongsy/portage/internal/app/reporting     coverage: 79.8%
# ok  github.com/duongsy/portage/internal/domain            (7 guard)
# ok  github.com/duongsy/portage/internal/domain/catalog    coverage: 89.1%
# ok  github.com/duongsy/portage/internal/domain/logistics  coverage: 76.1%
# ok  github.com/duongsy/portage/internal/domain/ordering   coverage: 81.4%
# ok  github.com/duongsy/portage/internal/domain/pricing    coverage: 73.6%
# ok  github.com/duongsy/portage/internal/domain/procurement coverage: 80.6%
# ok  github.com/duongsy/portage/internal/domain/shared     coverage: 90.7%
# ok  github.com/duongsy/portage/internal/platform/auth     coverage: 88.8%
# ok  github.com/duongsy/portage/internal/platform/wire     coverage: 94.9%
# ok  github.com/duongsy/portage/internal/worker            coverage: 87.3%
# tổng 292 test (08/09, hết P9 + variant/size + migrate race) — 30 bỏ qua khi không có DSN (28 postgres + 1 wire + 1 rollback cố ý), 262 còn lại < 2 giây không cần gì
```

### 🛡️ Test canh quyết định — `internal/domain/decisions_test.go`

Chín quy ước bên dưới **không chỉ nằm trong tài liệu**. Bảy trong số đó đã được
viết thành test, quét mã nguồn bằng `go/ast`. Vi phạm là **build đỏ**.

Lý do có nhóm test này: ngày 04/09, `CategoryPolicy` bị thêm `hsCode` và
`dutyRate` — **vài ngày sau khi đã ghi hai lần** rằng tuyến vận chuyển đang
dùng gộp thuế vào giá /kg và không tách dòng thuế. **Mọi unit test đều xanh.**
Thiết kế vẫn sai.

> Unit test hỏi *"hàm này trả đúng chưa?"*. Test canh quyết định hỏi
> *"code này còn tuân quy ước mình đã chốt không?"*. Hai câu hỏi khác nhau,
> và cái thứ hai mới là cái hay bị quên.

| # | Guard | Bắt cái gì |
|---|---|---|
| 1 | `noFloatingPointInDomain` | `float32`/`float64` lọt vào domain |
| 2 | `domainDoesNotReadTheClock` | `time.Now()`, `time.Since()` trong domain |
| 3 | `noBrandNamesInCode` | "Nike"/"Adidas"… trong **code** (comment thì được) |
| 4 | `catalogKnowsNothingAboutTax` | `duty`, `hsCode`, `vat`, `customs` quay lại catalog |
| 5 | `noSettersOnDomainTypes` | method `SetXxx()` trên aggregate |
| 6 | `domainImportsOnlyStdlibAndAllowlist` | driver DB, framework, SDK lọt vào domain |
| 7 | `boundedContextsDoNotImportEachOther` | catalog import pricing, v.v. |

**Đã kiểm chứng từng cái thật sự sập bẫy** — cố tình vi phạm, xem test đỏ, rồi
khôi phục. Guard không bao giờ đỏ là guard vô dụng.

Mỗi guard trong file đều ghi rõ ba thứ: **quyết định là gì**, **ghi ở tài liệu
nào**, và **làm gì nếu nó đỏ** — kể cả trường hợp *"quyết định đã đổi, sửa
guard đi"*. Đây là tripwire, không phải điều răn.

```bash
go test ./internal/domain/ -v      # chạy riêng nhóm guard
go test ./...                      # chạy chung, CI cũng chạy cái này
```

---

### Quy ước code đã chốt — đọc trước khi viết bounded context đầu tiên

Mỗi quy ước dưới đây đều **đã có code thật** trong `shared/` để soi.

**1. Hai loại constructor, phân biệt bằng cách xử lý input sai**

| Loại | Ví dụ | Input | Sai thì |
|---|---|---|---|
| Validating | `ParseMoney`, `ParsePercent`, `NewExchangeRate`, `NewWeight`, `CurrencyFromCode`, `ParseID`; catalog: `ParseHostname`, `ParseCategoryCode`, `ParseSourceURL`, `NewParcelSpec`, `NewProvenance`, `FreeShippingOver`, `RegisterMerchant`, `NewCategoryPolicy`, `AddProduct` | **không tin được** — form, CSV, dòng DB | trả `error` |
| Literal | `NewMoney`, `Grams`, `Kilos`, `NewDimensionsCM`, `RatePPM`, `Must*` (`MustParseHostname`, `MustParseCategoryCode`, `MustParcelSpec`, `MustProvenance`…) | **lập trình viên viết tay** | `panic` |

`panic` ở đây = "code của mình có bug" (như Go panic khi index âm), **không bao
giờ** là lỗi của người dùng. Adapter nhận input ngoài → luôn dùng loại validating.

> **So với Symfony:** `error` ~ `\DomainException` / `\InvalidArgumentException`
> (bắt được, xử lý được); `panic` ~ `\LogicException` (không nên bắt, sửa code).

**2. Tiền: `int64` đơn vị nhỏ nhất, phép nhân/chia qua `math/big`, làm tròn
half-up xa số 0**

`13.215 → 13.22` và `-13.215 → -13.22`, để hoàn tiền = đúng số thuế đã thu đổi
dấu. Tích trung gian đi qua `big.Int` nên không tràn; kết quả tràn `int64`
(92 triệu tỷ đô) là bug → panic, không trả `0.00`.

**3. Parse số nghiêm ngặt: chỉ nhận `-?digits[.digits]`**

Không dấu phẩy, không `+`, không `150.`, không `.5`. Lý do: `"150,50"` là 150,50
với người Việt nhưng là 15.050 với người Mỹ — domain mà đoán thì sai với một
trong hai. Chuẩn hoá theo locale là việc của adapter (form/API), **trước khi**
gọi vào domain. Một hàm `parseDecimal` dùng chung cho Money, Rate, ExchangeRate.

**4. Cân nặng: mọi bước làm tròn đều làm tròn LÊN**

Kể cả bước trung gian `VolumetricWeight`. Bug đã gặp: 1900,8 g bị cắt xuống 1900
→ rơi đúng mốc 100 g → tính thiếu một bậc cân. Chi tiết ở mục 9.

**5. ID = UUID v7, sinh trong domain, mỗi aggregate bọc kiểu riêng**

```go
type OrderID struct{ shared.ID }          // OrderID ≠ MerchantID với compiler
func NewOrderID() OrderID { return OrderID{shared.NewID()} }
```

Không có `private ?int $id = null` chờ `flush()`. Object có danh tính từ dòng
đầu constructor. v7 có timestamp ở đầu nên index B-tree ghi nối đuôi, không
rải rác như v4.

**6. Domain event: aggregate chỉ GHI, tầng app mới PHÁT**

Aggregate root nhúng `shared.Events`, gọi `o.Record(ev)` trong method đổi trạng
thái. Tầng app gọi `o.PullEvents()` **sau khi** repository lưu xong, ghi vào
outbox cùng transaction. Không bao giờ publish trước khi lưu.

**7. Domain không đọc đồng hồ**

Method nhận `now time.Time` qua tham số. Không `time.Now()` trong
`internal/domain/`. Tầng app lấy `now` từ `app.Clock` (`internal/app/ports.go`)
**một lần** ở đầu handler rồi truyền xuống; production cắm `clock.System`, test
cắm `clock.Fixed` (`internal/platform/clock/`). Test canh số 2 chặn `time.Now()`
lọt vào domain.

**8. `internal/app/<context>/`, không để phẳng**

Mỗi bounded context một thư mục con trong `app/`, song song với `domain/`.
Package đặt tên `<context>app` (`catalogapp`) vì `package catalog` sẽ trùng tên
với domain và mọi file phải alias import. Khuôn một use case: `now` → `InTx` →
(load) → domain → `Save` → `PullEvents` → `Outbox` — xem WALKTHROUGH.md §3, §8.

**9. `Money{}`, `Currency{}`, `ID{}` — zero value của Go là trạng thái "chưa
set", không phải giá trị hợp lệ**

Go luôn cho phép viết `shared.Money{}` — không cấm được. Nên mỗi VO có `IsZero()`
/ `IsValid()`, và constructor từ chối xây trên zero value (`NewMoney(5,
Currency{})` panic). Coi nó như cột nullable chưa gán.

**10. Constructor > 3 tham số → struct `XxxDetails` — kèm hai điều kiện**

```go
// ≤ 3 tham số: vị trí                 > 3 hoặc sẽ lớn: struct
NewParcelSpec(weight, dims)            RegisterMerchant(MerchantDetails{
NewExchangeRate(from, to, "26000")         Name: "…", Site: …, Currency: shared.USD,
                                           FreeShipping: …, Sourcing: …,
                                       }, now)
```

| | Tham số vị trí | Struct tham số |
|---|---|---|
| Thêm field | **lỗi biên dịch** ở mọi chỗ gọi — compiler bắt | mọi chỗ gọi vẫn compile, field mới **âm thầm** mang zero value |
| Đảo 2 tham số cùng kiểu | compiler **không** bắt | không xảy ra — có tên |
| Đọc call site | phải nhớ thứ tự | tự giải thích |

Compile-time mạnh hơn runtime, nên struct chỉ được dùng khi **đủ hai điều kiện**:

1. **Constructor validate mọi field bắt buộc** — zero value đi vào là `error`
   lớn tiếng, không phải merchant có lỗ. Field nào cho phép zero thì zero phải
   là **đáp án nghiệp vụ an toàn** (`FreeShipping{}` = shop luôn tính phí ship;
   `Sourcing` nil = chưa có cách lấy dữ liệu).
2. **Có test `Test…_everyFieldIsValidated`** — dùng `reflect` zero **từng field
   một** trên details hợp lệ và đòi constructor từ chối; field nào cho phép zero
   phải khai trong `zeroIsMeaningful` kèm lý do. Nó thay cho việc compiler không
   còn bắt.

   > ⚠️ Một test zero **cả struct** là KHÔNG đủ — đã kiểm chứng: thêm field
   > `Country` không validate, 91 test vẫn xanh, vì `MerchantDetails{}` chết ở
   > `Name` rỗng rồi return, không bao giờ chạy tới field mới.

`now time.Time` để **ngoài** struct: nó là "lúc nào", không phải chi tiết của
merchant. Symfony: DTO / named arguments; Go không có named arguments.

## 7. Việc tiếp theo

- [x] `git init` + commit đầu tiên
- [x] Shared kernel: review, sửa bug, thêm `ID`, `Events` (04/09)
- [x] CI, README, `.editorconfig` (04/09)
- [ ] Tạo repo GitHub **bằng tài khoản cá nhân**, `git remote add origin`, push.
      Xác nhận username đúng là `duongsy` — nếu không, sửa module path trong
      `go.mod` và mọi import `github.com/duongsy/portage/...`
- [x] `internal/domain/catalog` — Merchant, CategoryPolicy, Estimate, FreeShipping, repository interface (04/09)
- [x] `internal/domain/catalog` — Product, Variant, Provenance, SourceURL (04/09, chiều)
- [x] `internal/app/catalog` + `adapter/memory` + `platform/clock` — 3 use case, port/adapter in-memory (04/09, tối)
- [x] `internal/adapter/http` + `cmd/api` — 3 endpoint, locale, bảng lỗi; chạy thật (04/09, tối)
- [x] API snapshot/rehydrate cho aggregate — `snapshot.go` (04/09, đêm)
- [x] `eventcodec` — payload outbox viết tay (04/09, đêm)
- [x] `adapter/postgres` + `docker-compose.yml` + CI service Postgres (04/09, đêm)
- [x] `platform/wire` — `Memory`/`Postgres`, `cmd/api -dsn` (04/09, đêm)
- [x] `internal/worker` + `cmd/worker -dsn` — cùng `wire.Postgres`, không chế độ memory (04/09, đêm)
- [x] `internal/domain/pricing` — ShippingLane, RateCard, DutyPolicy, QuotePolicy, Calculate, Quote (05/09)
- [x] `internal/contracts` + `eventcodec.Decode` — Published Language; `pricingapp.Projector` nghe catalog (05/09)
- [x] `0002_pricing.sql` + 5 repo Postgres; `wire.Graph` + `wire.Subscribe`; `/quotes`; smoke thật (05/09)
- [x] Ba bước operator qua HTTP: variants / confirm-listing / measure (05/09)
- [x] **P9/T1** Auth thật — `platform/auth` (Principal, `Verifier`/`Issuer`), bearer token bọc CẢ mux,
      `X-Operator-ID` và `customer_id` biến mất, `POST /tokens`, `0006_auth.sql`, `-bootstrap-operator-token` (06/09)
- [x] `GET /tokens` + `DELETE /tokens/{hash}` — xem/thu hồi chìa; chìa operator cuối cùng → 409 (06/09, trả nợ T1)
- [x] **P9/T4** `POST /fx` — endpoint tỷ giá (lane đã có `POST /lanes` từ 06/09); route DUY NHẤT không phát event (06/09)
- [ ] `config/ratecard.yaml` với bảng giá **thật** (hiện là hằng trong `wire.quotePolicy`)
- [x] `internal/domain/ordering` — `CustomerOrder` + luật cọc 50 % + `Cancel`/`Refund` (§26), nghe `pricing.quote_accepted`; `/orders` (05/09, chiều)
- [x] `internal/domain/procurement` — `PurchaseTask`, port `MerchantACL` + `merchant.Manual`; nghe `deposit_paid`, phát `purchase_confirmed/failed` → `ordering.Reactor`; `/purchase-tasks` (05/09, chiều muộn)
- [x] **P9/T6** `merchant.Router` — chọn ACL theo shop, mặc định `Manual`; thêm shop có API = một dòng trong `wire` (06/09)
- [ ] `adapter/merchant/<shop>.go` — ACL **thật** cho một shop có API (chưa có shop nào cho API)
- [x] `internal/domain/logistics` — `Parcel`, `ConsolidationBatch`, `FreightAllocator`, `shared.Allocate`; `pricing.Reconciliation` = đối soát Quote vs Actual; `/lanes`, `/parcels`, `/batches`, `/reconciliations` (06/09)
- [x] **P9/T3** Quét quote hết hạn — `pricing.QuoteRepository.IssuedBefore`, `pricingapp.ExpireQuotesHandler`
      (mỗi quote một transaction), `worker.Sweeper`, `cmd/worker -sweep` + `cmd/api` memory (06/09)
- [x] **P9/T2** Read model bảng riêng — `internal/app/reporting`, `order_summaries` dựng từ event của NĂM context,
      `GET /me/orders` (khách) + `GET /orders?status=` (operator), `0007_reporting.sql` (06/09)
- [x] **P9/T5** `adapter/openai` — ACL thứ hai: đọc trang shop bằng model → nháp `SourcedByFeed`; không key → 503 (06/09)
- [x] **P9/T7** `docs/FLOW-ORDER.md` — một đơn từ đầu tới cuối: mỗi bước ai gọi, ai nghe, bảng nào đổi, hỏng thì sao (06/09)
- [x] **P9/T8** dọn dẹp — unit test repo memory (17,5 % → 80,8 %), `scripts/smoke.sh` + smoke trong CI, README bảng route (06/09)
- [x] **P9 XONG (06/09)** — T1 auth · T2 read model · T3 quét quote · T4 `POST /fx` · T5 `adapter/openai` · T6 ACL Router · T7 `docs/FLOW-ORDER.md` · T8 dọn dẹp
- [x] `docker-compose.yml` cho Postgres
- [x] Tầng adapter: postgres, http

### Số liệu nghiệp vụ đã chốt

| Tham số | Giá trị | Ghi chú |
|---|---|---|
| Tỷ giá | 26.000 ₫/USD | tỷ giá ngân hàng |
| Cọc khách trả trước | 50% giá trị đơn | |
| Lãi mục tiêu | 500.000 ₫/đơn | mức sàn tối thiểu |
| Kho Mỹ | Denver, Colorado | sales tax ~8,81% |
| Tuyến vận chuyển chính | dịch vụ gửi hàng ở Mỹ, trọn gói tận nhà VN | phí trả hết đầu Mỹ |
| Chia thể tích | 5000 (cm³/kg) | cần xác nhận lại — **tạm là seed** `wire.seedPricing`: 5000, bước 500 g, 9/10/12/14 USD/kg, phụ thu pin 3 USD |
| Thời hạn báo giá | 48 h | `wire.quotePolicy()` — chụp vào từng quote |
| Phân loại hàng | footwear, apparel → branded; electronics → electronics; còn lại standard | `wire.goodsClasses()` — thiếu dòng → báo giá THẤP, lộ ở đối soát |

### Số còn thiếu — cần hỏi nhà gửi hàng bên Mỹ

- [ ] Bảng giá theo **kg**, phân theo **loại hàng** (thường / hàng hiệu / điện tử / nhạy cảm)
- [ ] Bước làm tròn cân — 100 g hay 500 g?
- [ ] Số chia thể tích họ dùng — 5000 hay 6000?
- [ ] Có phụ thu hàng **pin lithium** không (tai nghe, đồng hồ…)?
- [ ] Trần giá trị mỗi kiện?

### Ghi chú rủi ro

Tuyến "gửi ở Mỹ, trả trọn gói, về tận nhà VN" hoạt động tốt ở quy mô **cá nhân**.
Ở quy mô kinh doanh (nhiều thùng/tháng, từ một địa chỉ gửi tới nhiều người nhận),
cách hải quan xử lý có thể khác. Hệ thống nên **đếm và cảnh báo** khi tần suất
vượt ngưỡng tự đặt, thay vì giả vờ rủi ro đó không tồn tại.

---

## 8. Dựng lại trên máy mới — bản rút gọn

```bash
winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements
# ĐÓNG rồi MỞ LẠI terminal

git clone <repo> portage
cd portage
go mod download        # tải thư viện        (~ composer install)
go test ./...          # xác nhận chạy được
```


---

## 9. Nhật ký review 04/09/2026 — bug đã tìm thấy và bài học

Review toàn bộ `shared/` bằng cách viết một file test "thăm dò" chạy trên bản
copy tạm, in ra giá trị thật thay vì đoán. Kết quả:

| # | Bug | Biểu hiện thật | Bài học |
|---|---|---|---|
| 1 | `VolumetricWeight` cắt xuống | thùng 32×27×11 cm = 1900,8 g → bill `1.900 kg`, đúng phải `2.000 kg` | trong chuỗi tính phí, **mỗi bước** phải làm tròn lên; một bước cắt xuống là mất một bậc cân |
| 2 | `ParseMoney` bỏ dấu phẩy | `"150,50"` → `15050.00 USD` | domain không đoán locale; nhận đúng một định dạng, còn lại từ chối |
| 3 | Tỷ giá 0 được chấp nhận | `3900000 VND` → `0.00 USD`, không lỗi | value object phải **không tạo được** ở trạng thái vô nghĩa |
| 4 | Không biểu diễn được VND→USD | `1/26000` cần 9 số lẻ, `ParsePercent` cho 4 | đừng tái dùng một kiểu (`Rate`) cho khái niệm khác (`ExchangeRate`) chỉ vì "cùng là số" |
| 5 | `ParsePercent("")` → 0% | thuế suất thiếu → hàng đi không thuế | chuỗi rỗng là lỗi, không phải số 0 |
| 6 | `Money.Mul` tràn `int64` | `1<<62 cent × 100%` → `0.00 USD` | tích trung gian phải qua `big.Int`; `Convert` đã làm, `Mul` thì không → không nhất quán |
| 7 | Mất dấu âm dưới 1% | `-0.5%` in ra `0.5000%` | `-5000/10000 == 0` trong số nguyên; format phần độ lớn rồi gắn dấu |
| 8 | `Currency{}` tạo được từ ngoài | `Money` với currency `""`, `Add` vẫn chạy | zero value của Go — xem quy ước 9 |
| 9 | `ExchangeRate` có `Source`, `AsOf string` exported | lai giữa VO và entity, `AsOf` phải là `time.Time` | provenance là việc của Pricing; shared kernel chỉ giữ cái *thật sự* chung |
| 10 | `Dimensions`, `ExchangeRate` không có getter | adapter/postgres không lưu được | VO đóng kín field, nhưng phải **đọc ra được** để persist |

Cách kiểm chứng: mỗi bug có một test tái hiện (đỏ) trước khi sửa (xanh).
Coverage 71% → 93%.

### Đợt 2 — review lại sau khi tổ chức lại source

Cùng cách làm: viết file test thăm dò, in giá trị thật, xoá file sau khi xong.
Ba lỗi còn sót, **cùng một họ**: *giá trị vượt biên thì xử lý thế nào*.

| # | Bug | Biểu hiện thật | Bài học |
|---|---|---|---|
| 11 | `Add`/`Sub`/`Sum` tràn `int64` âm thầm | `MaxInt64 VND + 1` → `-9223372036854775808 VND`, `err = nil` | **nhất quán trong cùng một kiểu**: `Mul`/`Convert` đã panic khi tràn, `Add` lại để wrap — hai cách xử lý cho cùng một loại lỗi |
| 12 | `Weight.Add`, `RoundUpTo`, `Kilos` tràn | `MaxInt64 g + 1 g` → cân nặng **âm**, `String()` in `-9223372036854775.-808 kg` | tràn ở đây phá **chính invariant** `NewWeight` dựng lên; và `String()` bỏ `abs64` vì "weight không bao giờ âm" → hai lỗ hổng nối nhau |
| 13 | `ParseID` nhận 4 dạng không chuẩn | `{...}`, `urn:uuid:...`, CHỮ HOA, không gạch nối — đều `err = nil` | mâu thuẫn với chính nguyên tắc đã áp cho `parseDecimal` (quy ước "domain không đoán"); và hai cách viết của một id sẽ **index thành hai key khác nhau** trong DB |
| 14 | `Money{}.Add(Money{})` lọt | trả `"0 "` — số tiền không có tiền tệ, `err = nil` | hai giá trị **cùng vô hiệu** thì "khớp" nhau; guard phải hỏi *"có hợp lệ không"* trước khi hỏi *"có khớp không"* |

**Cách sửa — một nguyên tắc chung cho cả bốn:**

```go
// decimal.go — số học int64 từ chối wrap, panic thay vì trả giá trị sai
func addExact(a, b int64) int64
func subExact(a, b int64) int64
func mulExact(a, b int64) int64
```

Chọn `panic` chứ không phải `error` là **có chủ đích**, theo đúng quy ước 2 của
project: tràn `int64` ở đây không phải lỗi người dùng mà là **bug của lập trình
viên** — 9,2 tỷ tỷ đơn vị nhỏ nhất là 92 triệu tỷ đô, hoặc nhiều lần toàn bộ
tiền đồng đang lưu hành. Không có nghiệp vụ nào chạm tới đó.

> **Bài học lớn nhất của đợt này:** ba trong bốn bug **không phải lỗi logic**,
> mà là **thiếu nhất quán** — cùng một loại tình huống, chỗ này xử lý kiểu này,
> chỗ kia kiểu khác. Khi review, câu hỏi hiệu quả nhất không phải *"đoạn này
> đúng chưa?"* mà là *"đoạn này có xử lý giống đoạn kia không?"*

Coverage 93,0% → **94,3%**, 38 → **42 test**.

### Đợt 3 — review `catalog` (04/09, chiều)

Review bản `catalog` đầu tiên (Merchant, FreeShipping, CategoryPolicy, 3 event).
Test xanh, coverage 88,7%, nhưng **thiết kế có 4 chỗ lệch với chính tài liệu**.

| # | Vấn đề | Biểu hiện | Sửa | Bài học |
|---|---|---|---|---|
| 15 | `volumetricDiv` trong `CategoryPolicy` | comment tự nhận là dữ liệu của hãng vận chuyển, nhưng lưu theo ngành hàng; test đóng đinh `luggage` = 6000 | gỡ; `ShippingLane` (pricing) giữ số chia + bước làm tròn | gỡ một field vì sai context → hỏi tiếp *"field nào cùng nguồn gốc?"*; thuế và số chia cùng là dữ liệu của lane |
| 16 | Thiếu `Estimate` | CATALOG.md §3 gọi cân nặng/hộp theo ngành hàng là "tài sản riêng", code không có → Pricing không quote được món lần đầu | thêm `Estimate{Weight, Dimensions}`, bắt buộc khi tạo category | tài liệu nói "quan trọng nhất" mà code không có field = tài liệu và code đang mô tả hai hệ thống khác nhau |
| 17 | `CategoryCode` là `type string`, không validate | `CategoryCode("Giày Dép!")` compile bình thường; `Hostname` cùng file thì chặt chẽ | thành `struct{ code string }`, chỉ tạo qua `ParseCategoryCode` (`^[a-z][a-z0-9_]*$`) | khoá tự nhiên đi vào URL/DB → validate như mọi VO; `type X string` chỉ hợp cho enum đóng có `IsValid()` |
| 18 | `Merchant` không tắt được | có `Rename`, `EnableSourcing`; không có `Suspend`, `DisableSourcing`, đổi `FreeShipping` | thêm `Suspend`/`Reinstate` + `MerchantStatus`, `DisableSourcing`, `ChangeFreeShipping`, 4 event | entity có vòng đời thì phải có **cả hai chiều**; event liên context đáng giá nhất là `MerchantSuspended`, không phải `MerchantRenamed` |
| 19 | `ErrFreeShipCurrency` dùng cho threshold âm | test cũng assert sai danh tính lỗi | `ErrNegativeThreshold` | mỗi lý do từ chối một sentinel; `errors.Is` bên gọi mới phân biệt được |
| 20 | 26 hàm/struct một dòng | ngược quy tắc style vừa chốt | mở rộng | — |

Đợt này **không có bug số học** — toàn bộ là *field đặt sai chỗ* hoặc *thiếu
chiều ngược*. Câu hỏi hiệu quả nhất khi review một context mới: *"dữ liệu này
**ai sở hữu**? Có context nào khác cũng có quyền nói về nó không?"* Có → nó đang
ở sai chỗ, hoặc sắp có hai nguồn sự thật.

Coverage catalog 88,7% → **92,5%**; tổng 78 test (43 shared + 28 catalog + 7 guard).

### Đợt 4 — `Product`/`Variant`/`Provenance` (04/09, chiều muộn)

Không phải review — là **thiết kế trước, TDD sau**: 3 câu hỏi treo trong
CATALOG.md được chốt (trùng → phương án C; provenance 3 nhóm; `Estimate` →
`ParcelSpec` dùng chung), viết test cho `Product` (9), `Provenance` (2),
`SourceURL` (1) rồi mới viết code. Hai điểm đáng ghi:

- **`now` chỉ đi kèm khi có gì được ghi lại** (event hoặc timestamp).
  `ConfirmListing(prov)` không nhận `now` — provenance đã mang thời điểm riêng.
  Tham số không dùng là tham số nói dối. (`ClearDuplicateFlag` lúc đầu cũng
  không nhận `now` — sửa ở đợt 5, vì hoá ra nó *có* thứ cần ghi lại.)
- **`Reprice` cùng giá vẫn cập nhật provenance** nhưng không phát event:
  operator xác nhận đúng giá cũ là thông tin mới (giá đã được kiểm), nhưng không
  có gì "xảy ra" để ai phải phản ứng.

Tổng **91 test** (44 shared + 40 catalog + 7 guard); catalog 91,3%.

### Đợt 5 — review chéo từ session Windows (04/09, tối)

Bên kia review `Product`, đưa 4 điểm. Ba điểm đúng, sửa ngay; một điểm đổi quyết
định đã chốt nên hỏi lại.

| # | Nhận xét | Kiểm chứng | Sửa |
|---|---|---|---|
| 21 | `ClearDuplicateFlag()` không có reason/now/event → cặp trùng bị gắn cờ lại mãi, không ai biết đã xét | đúng; và lệch chuẩn của chính codebase (`Reinstate` có event) | `ClearDuplicateFlag(reason, now)` + `ProductDuplicateCleared`; `FlagDuplicateOf` cũng nhận reason; **thêm**: aggregate nhớ cặp đã bác → gắn lại là no-op |
| 22 | `listingProv` gộp tên + merchant + category; chỉ category sai là âm thầm | đúng về kỹ thuật | **giữ `listing`** — xác nhận "món này là gì" là một hành động của operator, không tách; lý do đầy đủ ở CATALOG.md §4 |
| 23 | 4 từ cho một thứ: `ParcelSpec`, `Estimate()`, `Parcel()`, `parcelProv` | đúng | `CategoryPolicy.DefaultParcelSpec()`, `Product.ParcelSpec()`; `ParcelProvenance()` giữ vì "parcel" là tên nhóm |
| 24 | `Test…_zeroDetailsIsRejected` không guard được thứ nó tuyên bố | **tái hiện được**: thêm `Country` không validate → 91/91 vẫn xanh | test reflection zero từng field (`…_everyFieldIsValidated`) cho `MerchantDetails` và `ProductDetails`; sửa quy ước 10 |

> **Bài học:** một test "chứng minh" điều gì thì phải **cố tình phá** để xem nó
> đỏ. Test zero-struct chưa từng được phá thử — nên nó "xanh" theo nghĩa vô dụng.
> Bên Windows đã làm đúng việc đó với 7 guard; giờ áp cho mọi test canh.

Tổng **93 test** (44 shared + 42 catalog + 7 guard).

### Đợt 6 — tầng app + adapter in-memory (04/09, tối)

Không phải review; là bước "ghép các mảnh lại xem có chạy không". Thiết kế lấy
từ DDD.md §30; kết quả mô tả từng dòng ở [WALKTHROUGH.md](WALKTHROUGH.md).

| Điều học được | Ở đâu |
|---|---|
| Use case tồn tại để **ghép**: clock + auth → `Provenance`; command khác details đúng hai field | `add_product.go` |
| Lỗi tầng app chỉ có khi luật **đụng hai aggregate** — đúng hai cái: `ErrMerchantInactive`, `ErrPriceCurrency` | `deps.go` |
| Go không có method generic → `InTx(ctx, fn) error`, kết quả qua biến closure | `ports.go` |
| Wiring là input của lập trình viên → `Deps` + `mustHave` panic lúc khởi động | `deps.go`, test `…panicsOnMissingDeps` |
| "Save trước Pull" thành **test**: Save lỗi → outbox rỗng | `…saveFailurePublishesNothing` |
| Adapter in-memory là adapter thật, nằm ở `adapter/`, có `var _ Port = (*Adapter)(nil)` | `memory/*.go` |
| Khoảng trống cố ý: repo giữ con trỏ; snapshot/rehydrate đến cùng Postgres | `memory/memory.go` |

Tổng **108 test** (44 shared + 42 catalog + 7 guard + 10 app + 5 memory).

### Đợt 7 — `adapter/http` + `cmd/api`, chạy thật (04/09, tối)

Mục 5d là kết quả. Điều học được:

| Điều học được | Ở đâu |
|---|---|
| Adapter làm **đúng ba việc**: decode, chuẩn hoá locale, map lỗi. Validate là gọi `Parse*` của domain — không viết luật thứ hai | `products.go` |
| `"50,00"` là 50 hay 5000 **tuỳ `Accept-Language`** — đây là chỗ đoán; domain từ chối đoán từ sáng là để chỗ này quyết | `decode.go`, test `…localeDecidesTheDecimalSeparator` |
| 400 / 404 / 409 là **ba cuộc hội thoại**: sửa request / không có / sửa trạng thái. `ErrPriceCurrency` là 409 vì request đúng, chỉ nghiệp vụ không chịu | `errors.go` |
| Một bảng lỗi + `errors.Is` đi hết chuỗi `%w` → lỗi domain bọc mấy lớp vẫn tìm đúng dòng; không có trong bảng = 500 + log | `errors.go` |
| `DisallowUnknownFields`: gõ nhầm `"nmae"` là 400, không phải tên rỗng âm thầm | `decode.go` |
| `cmd/api/main.go` điền **cùng struct `Deps`** với test — khác đúng `clock.System`. Có Postgres thì đổi ở đây, handler không đổi | `main.go` |
| Test tôi viết sai một case (VND `"150.00"` chết ở adapter trước khi tới luật 409) — test cũng có bug, và RED đúng lý do mới là RED | `server_test.go` |
| Smoke thật: build `api.exe`, `curl` 6 request, id trả về là UUID v7 tăng dần nhìn được bằng mắt | WALKTHROUGH.md §1a |
| **Review chéo lần 3 (Windows):** chiều "Save xong, Outbox hỏng" chưa test nào phủ — memory không có transaction nên merchant vẫn nằm lại (`Len()==1`). Sửa: viết test rollback, thấy nó đỏ đúng lý do, rồi `t.Skip` với memory → là test đầu tiên của `postgres.UnitOfWork`; ghi KNOWN GAP #2 vào `memory.go` | `…outboxFailureRollsBackTheSave`, `memory.go` |

Tổng **116 test** (44 shared + 42 catalog + 7 guard + 10 app + 5 memory + 8 http).

### Đợt 8 — quyết định KHÔNG làm `cmd/worker` bây giờ, tách `wire()` (04/09, tối)

Đề xuất ban đầu: xây `internal/worker` (relay đọc outbox, bus in-process) và chạy
bằng cờ `cmd/api -worker`. Review chéo bác cả hai, lý lẽ đã kiểm và đúng:

| Không làm | Vì sao |
|---|---|
| Binary `cmd/worker` stub (in thông báo, thoát) | **stub tệ hơn không có gì**: deploy được, exit 0, outbox đầy dần mà không ai biết. Không có binary thì hỏng lúc **build** — ầm ĩ đúng lúc. Docs đã đánh 🔜, stub làm docs kém chính xác hơn. Thư mục rỗng vốn không nằm trong git nên chẳng "giữ chỗ" gì |
| Cờ `-worker` trong `cmd/api` | cam kết "một binary, đổi chế độ bằng cờ" **trước khi biết worker có hình dạng gì** — mà lúc này nó chẳng có outbox chung để đọc |
| `internal/worker` relay | cùng lý do: hình dạng worker (batch? at-least-once? serialize event thế nào?) chỉ rõ khi có bảng outbox thật |

Làm thay: tách phần ráp trong `cmd/api/main.go` thành **`wire(clock) catalogapp.Deps`**.
`main()` chỉ còn parse cờ + serve. Khi worker thật ra đời, `cmd/worker/main.go`
gọi lại đúng `wire()` — không chép, không lệch cấu hình. Kèm test đầu tiên của
`cmd/api`: `TestWire_servesTheSeededCatalog` — gọi `wire()`, bắn request, chứng
minh seed `footwear` hoạt động và `luggage` chưa seed là 404.

> **Bài học:** *"im lặng là chế độ hỏng tệ nhất."* Một thứ chưa làm được thì để
> nó **không tồn tại** (build đỏ, 404, panic lúc khởi động) hơn là tồn tại mà
> không làm gì. Cùng tinh thần với `mustHave` panic và với `Publish` từ chối.

Tổng **117 test** (+1 `cmd/api`).

### Đợt 9 — hạ tầng thật: Postgres, outbox, worker (04/09, đêm)

Bốn pha, mỗi pha một RED → GREEN. Quyết định và bài học:

| Quyết định / bài học | Ở đâu |
|---|---|
| **Memento** thay cho export field: `Snapshot()` đi ra, `FromSnapshot()` đi vào, không phát event, từ chối hình dạng không thể có (`ErrInvalidSnapshot`) | `catalog/snapshot.go`, `snapshot_test.go` |
| **Payload outbox viết tay (phương án B)** — lý lẽ bên Windows: `EventName()` đã viết tay để bảo vệ wire format, để `json.Marshal` suy payload là bảo vệ nửa trên bỏ ngỏ nửa dưới. Câu hỏi quyết định: *consumer đọc payload hay load aggregate?* — guard 7 cấm import → payload là tất cả nó có → **là hợp đồng** | `eventcodec/codec.go` |
| Guard "thêm event quên mapper là đỏ" phải **không cần nhớ**: `go/ast` đọc `events.go`, so với bảng hợp đồng | `codec_test.go` |
| Transaction **đi theo `ctx`** (`txKey{}` unexported); repo hỏi `db(ctx)`; handler không đổi một dòng | `postgres/postgres.go` |
| Test rollback từng skip **chạy thật và pass** — lời hứa §27 có bằng chứng | `TestUnitOfWork_rollsBackTheSaveWhenTheOutboxFails` |
| Migration = SQL nhúng + bảng version + advisory xact lock; không tool | `postgres/migrate.go` |
| Test tích hợp song song trên một DB → **advisory lock trong helper**, không `-p 1` | `postgres/pgtest` |
| `pool.Close()` khi còn connection acquired → **treo mãi**; giết ssh không giết `go test` → process zombie **giữ khoá** → mọi lần sau treo. Luôn `-timeout`; kiểm `Get-Process *.test` | §5e |
| Worker: at-least-once **có chủ đích** — publish trước, MarkSent sau; subscriber idempotent | `worker/worker.go` |
| `cmd/worker` bắt buộc `-dsn`, dùng cùng `wire.Postgres` với api | `cmd/worker/main.go` |
| Test của tôi sai một case (`"150.00"` VND chết ở adapter trước khi tới luật 409) và một lần đóng pool sai — **test và harness cũng có bug** | đợt 7, đợt 9 |

Smoke thật: `api -dsn` + `worker -dsn`, 4 request → 4 dòng outbox `sent = t`, worker log
đủ payload. Transcript: WALKTHROUGH.md §7.

Tổng **140 test** (44 shared + 47 catalog + 7 guard + 11 app + 5 memory + 8 http + 3 codec
+ 6 postgres + 3 wire + 6 worker; 1 skip cố ý khi không có Postgres).

### Đợt 10 — context thứ hai: `pricing` (05/09)

Mục tiêu: chứng minh bằng code câu hỏi DDD.md §6–§7 — *hai context nói chuyện bằng gì*.
Thứ tự làm: dời `ParcelSpec` → `shared` + `CategoryDefined` + `ProductPublished` mang
price/parcel → `contracts` → domain `pricing` (RED → GREEN từng file) → `pricingapp` →
codec + guard → ba use case operator + `/quotes` → `0002_pricing.sql` + repo → `wire.Graph`
+ `Subscribe` → test vàng hai graph → smoke thật. Toàn bộ đường đi: WALKTHROUGH.md §14.

| Quyết định / bài học | Ở đâu |
|---|---|
| Event **mang đủ trạng thái** consumer cần; guard 7 cấm load ngược → "chỉ gửi id" là coupling trá hình | `catalog/events.go` |
| Published Language là **package riêng** `internal/contracts` — không domain (pricing import nó, guard 7), không adapter (app import nó) | `contracts/catalog_v1.go` |
| Decoder chỉ viết khi có consumer (`ErrNoDecoder`) — decoder không ai gọi là hợp đồng không ai test | `eventcodec/decode.go` |
| Bảng định tuyến event ở **composition root** (`wire.Subscribe`) — chỗ duy nhất được biết cả wire format và mọi consumer; `on[T]` generic decode một chỗ | `wire/subscribe.go` |
| Projection: **idempotent** (upsert) + **chịu sai thứ tự** (measured trước publish → bỏ qua) + **từ vựng riêng** (`Classification`) — mỗi tính chất một test | `pricingapp/projector.go` |
| `Calculate` là hàm thuần (Domain Service §17); quyết định duy nhất (cân cái gì) phải lộ ra `Estimated` | `pricing/calc.go` |
| `Quote.Accept` trễ **vừa đổi trạng thái vừa từ chối** → tầng app commit rồi mới trả lỗi (cùng ý `Relay.RunOnce`) | `pricingapp/accept_quote.go` |
| `Policy`/`Classification` là **cấu hình nghiệp vụ** (VO) trong `Deps`, không phải interface; panic khi zero | `pricingapp/deps.go`, `wire.quotePolicy()` |
| Ranh giới context vẽ trong SQL: `listings.product uuid` trần, **không FK** sang catalog; breakdown lưu **từng cột** cho đối soát | `0002_pricing.sql` |
| `laneFrom` dựng lại qua `NewShippingLane` — dòng sửa tay vô nghĩa bị từ chối lúc load | `postgres/pricing_repos.go` |
| `cmd/api` in-memory chạy relay **trong process** (relay thật, không stub) — không thì `/quotes` mãi 404; Postgres để cho `cmd/worker` | `cmd/api/main.go` |
| **Bug của tôi:** patch theo anchor `catalog.MustParcelSpec` sau khi đã dời sang `shared.` → assert fail, file ghi nửa chừng (import thừa). Luôn fetch lại file trước khi patch | `codec_test.go` |
| **Bug của tôi:** getter `Duty()` đụng method `Duty(item)` — Go không overload; đặt tên theo vai `DutyPolicy()` | `pricing/lane.go` |
| **Bug của tôi:** struct ẩn danh trong test thiếu `json:"sales_tax"` → decode rỗng → test đỏ dù code đúng | `http/quotes_test.go` |
| Coverage `memory` 65 % → 38 %: `pricing_repos.go` được package khác dùng nhưng không có unit test riêng; coverage đo **theo package** | `memory/pricing_repos.go` |
| `TestMemory_/TestPostgres_servesTheSeededCatalog` → `wholeFlow`: cùng lời hứa Ports & Adapters, phạm vi giờ là **cả hệ** — kể cả hai lần 404 `listing_not_found` trước relay | `wire/wire_test.go` |

Smoke thật (`scripts/smoke.ps1`): quote `issued branded 2500 g, total 5 393 720 VND, deposit
2 696 860`; worker log #11–#16 từ `merchant_registered` tới `pricing.quote_accepted`.

Tổng **172 test** (44 shared + 47 catalog + 14 pricing + 7 guard + 14 catalogapp + 4 pricingapp
+ 5 memory + 5 codec + 12 http + 11 postgres + 3 wire + 6 worker).

### Đợt 11 — context thứ ba: `ordering` (05/09, chiều)

Mục tiêu: DDD.md §26 (Saga, điểm không thể quay đầu) và §31 (vòng đời đơn) thành code có test.
Thứ tự: domain (`CustomerOrder`, `Refund`, bảng `Cancel`, snapshot; RED → GREEN) → `QuoteAcceptedV1`
+ decoder → codec 8 event → `orderingapp` (`mutate`) → memory repo → `0003_ordering.sql` + repo Postgres
→ HTTP `/orders` → `wire.Graph.Ordering` + `Subscribe` → test vàng kéo dài. WALKTHROUGH.md §15.

| Quyết định / bài học | Ở đâu |
|---|---|
| **Bug của tôi:** hằng trạng thái `OrderPurchased` trùng tên kiểu event `OrderPurchased` — Go một namespace/package → đổi hằng thành `Status*` (đúng như code mẫu §26 đã viết) | `ordering/order.go` |
| Cọc phải **đúng số**: thiếu = không cam kết, thừa = khoản hoàn không ai xin — luật của aggregate | `PayDeposit` |
| `Cancel` trả `Refund` **và** ghi vào aggregate — bảng 5 trạng thái, mỗi dòng một test con | `order_test.go` |
| `refuse()` chọn sentinel theo trạng thái thật (đã huỷ → `ErrOrderCancelled`) — lý do thật, không phải lý do gần nhất | `order.go` |
| `Deps.mutate`: khuôn load → act → Save → outbox viết một lần khi có 7 handler cùng hình | `orderingapp/deps.go` |
| Luật "một quote một đơn" hai nửa: `Orders.ByQuote` (app) + `orders.quote UNIQUE` (DB) | `place_order.go`, `0003` |
| `quote_not_accepted` có hai nghĩa (chưa accept / relay chưa chạy) — như `listing_not_found` | test vàng |
| `Money.IsPositive` thêm vào shared kernel **kèm test**; test file là package `shared_test` → phải viết `shared.Money` (lỗi vet đầu tiên nhắc) | `shared/money.go` |
| Số tiền thanh toán qua `normalizeAmount` như mọi số tiền ở biên: `"2.696.860"` vi = 2 696 860 | `http/orders.go` |

Tổng **183 test** (45 shared + 47 catalog + 14 pricing + 4 ordering + 7 guard + 14 catalogapp + 4 pricingapp
+ 3 orderingapp + 5 memory + 5 codec + 13 http + 13 postgres + 3 wire + 6 worker).

### Đợt 12 — context thứ tư: `procurement`, ACL đầu tiên (05/09, chiều muộn)

Mục tiêu: DDD.md §22 (Anti-Corruption Layer) thành một port có adapter thật, và khép kín saga §26:
`deposit_paid` → task → `purchase_confirmed/failed` → đơn đi tiếp hoặc bù trừ. Thứ tự: domain
(`PurchaseTask`, `PurchaseReceipt`, port `MerchantACL`; RED → GREEN) → `shared.ErrOperatorRequired`
→ `MerchantRegistered.Currency` + 4 contract + codec/decoder → `merchant.Manual` → `procurementapp`
(3 nhánh ACL, fake `apiShop`) → `orderingapp.Reactor` → memory/Postgres repo → `/purchase-tasks`
→ `Subscribe` 5 dòng → test vàng tới `forfeited` → smoke 12 event. WALKTHROUGH.md §16.

| Quyết định / bài học | Ở đâu |
|---|---|
| ACL = **port bằng từ vựng của mình**, adapter đầu tiên là **một con người** (`Manual` trả `ErrManualPurchase`) — use case gọi port từ ngày đầu, API thật sau này là thêm file + một dòng wire | `procurement.MerchantACL`, `adapter/merchant/manual.go` |
| Ba câu trả lời của port → ba nhánh + một nhánh "shop sập → trả lỗi, relay thử lại, không lưu gì"; fake `apiShop` 10 dòng chứng minh cả bốn | `procurementapp/open_task.go`, `procurement_test.go` |
| Biên nhận **gắt** (reference, tiền > 0, đúng tiền tệ shop, có operator) vì `purchase_confirmed` đẩy đơn qua điểm không thể quay đầu | `PurchaseTask.Confirm` |
| Procurement cần tiền tệ của shop → `MerchantRegistered` mang `Currency`; đổi struct + codec + contract + bảng test **cùng lúc**, guard bắt nếu quên | `catalog/events.go`, `codec_test.go` |
| `ErrOperatorRequired` lên shared kernel; catalog alias **cùng giá trị** → `errors.Is` và bảng lỗi HTTP cũ không đổi | `shared/operator.go` |
| `Reactor.alreadyDone` nuốt `ErrNotDeposited`: chỉ Reactor biết "đã làm rồi" ≠ "không làm được" | `orderingapp/reactor.go` |
| Một event hai listener (`product_published` → pricing + procurement); Bus gọi theo thứ tự, lỗi một cái giữ dòng cho cả hai | `wire/subscribe.go` |
| Không có `POST /purchase-tasks`: luật "không mua khi chưa cọc" giữ bằng endpoint **không tồn tại** | `http/purchase_tasks.go` |
| `purchase_tasks."order" UNIQUE` (`order` là từ khoá SQL → phải quote) là nửa DB của "một đơn một task" | `0004_procurement.sql` |
| **Bug của tôi (x2):** script vá theo anchor dừng giữa chừng (`assert`) nhưng `wput` vẫn đẩy file cũ → "undefined: procurement.ErrManualPurchase". Sửa: script in `SKIP` từng anchor, không dừng; kiểm output trước khi đẩy | đợt 12 |

Smoke thật (`scripts/smoke.ps1`): 12 event #23–#34 qua 4 context; đơn `purchased` sau `purchase_confirmed`,
huỷ sau đó → `refund 0 VND, forfeited=True`. Transcript: WALKTHROUGH.md §16h.

Tổng **192 test** (45 shared + 47 catalog + 14 pricing + 4 ordering + 4 procurement + 7 guard + 14 catalogapp
+ 4 pricingapp + 3 orderingapp + 2 procurementapp + 5 memory + 5 codec + 14 http + 15 postgres + 3 wire + 6 worker).

### Đợt 13 — context thứ năm: `logistics`, Quote vs Actual khép (06/09, đêm)

Mục tiêu: vòng đời §31 chạy hết bằng code — cân thật, gom lô, chia cước, đối soát. Thứ tự:
`shared.Allocate` → domain `logistics` (Parcel, Batch, `FreightAllocator`, `LaneRule`; RED → GREEN)
→ pricing `LaneDefined` + `DefineLane` + `Reconciliation`/`Reconciler` → contracts/codec (3 decoder mới)
→ `logisticsapp` → `Reactor.OnBatchShipped` → memory/Postgres (`0005`) → 10 route → `Subscribe` +6 dòng
→ test vàng tới `delivered` + variance → smoke 20 event. WALKTHROUGH.md §17.

| Quyết định / bài học | Ở đâu |
|---|---|
| Chia tiền: floor từng phần bằng `big.Int` + largest remainder → **tổng đúng** là bất biến có test (kể cả số âm, trọng số 0) | `shared/allocate.go` |
| `FreightAllocator` là **interface** dù một implementation: nhiều luật chia, mỗi luật thiệt một nhóm khách (§17 Domain Service) | `logistics/allocator.go` |
| Phần chia **đóng băng trong `BatchShippedEvent.Allocations`**; ordering và pricing dùng số đó, không tính lại | `batch.go`, `Reconciler.OnBatchShipped` |
| Lane được **thông báo** (`lane_defined`) như category; logistics chỉ giữ divisor + step, không giữ giá | `pricing/events.go`, `DefineLaneHandler` |
| Quote vs Actual là **projection ba nguồn** của pricing (chủ nửa Quote); chịu mọi thứ tự; nguồn của chính mình sai (quote lạ) → lỗi to | `pricing/reconciliation.go`, `reconciler.go` |
| Hai aggregate một transaction **có lý do ghi rõ** (parcel status là sổ sách của batch) | `logisticsapp/batches.go` |
| Từ chối sớm: lane không rule → chặn lúc **mở batch**, không lúc ship | `OpenBatchHandler` |
| `ship` trả **200 [allocation]** thay 204: người đóng gói cần thấy phần chia ngay — ngoại lệ có lý của "ghi trả id" | `http/logistics.go` |
| Reactor nuốt thêm `ErrOrderCancelled` cho `batch_shipped`: đơn huỷ sau mua vẫn có thể nằm trong thùng | `orderingapp/reactor.go` |
| Đặt tên `BatchClosedEvent` chủ động (lần hai gặp clash hằng/kiểu) | `logistics/events.go` |
| **Bug của tôi:** đếm sai event relay (5 vs 6) → test đỏ, code đúng — đếm lại từ worker log | `wire_test.go` |
| **Bug của tôi (lần 3):** script vá `assert` dừng giữa chừng, `wput` vẫn đẩy → "undefined". Từ nay `print SKIP` từng anchor, không `assert` | đợt 13 |

Smoke thật: 20 event #42–#61 qua 5 context; `order at the end: delivered; quote vs actual: quoted 163.22+25.00,
actual 163.22+27.50 → variance -2.50 USD`. Transcript: WALKTHROUGH.md §17i.

Tổng **201 test** (46 shared + 47 catalog + 14 pricing + 4 ordering + 4 procurement + 2 logistics + 7 guard
+ 14 catalogapp + 5 pricingapp + 3 orderingapp + 2 procurementapp + 1 logisticsapp + 5 memory + 5 codec + 16 http
+ 17 postgres + 3 wire + 6 worker).

---

### Đợt 14 (06/09) — P9/T1: auth thật, `X-Operator-ID` biến mất

| Quyết định / bug | Ở đâu |
|---|---|
| **Cửa đóng mặc định**: `authenticate` bọc **cả mux**, không bọc từng route — route thêm ngày mai đã ở sau cửa trước khi ai kịp nhớ. Danh sách route công khai thì chỉ cần quên một dòng là hở | `http/auth.go` |
| **401 ≠ 403**: "tôi không biết anh là ai" khác "tôi biết, nhưng cái này không phải của anh". Gộp hai cái là biến API thành máy đoán token hợp lệ | `errors.go` |
| **Mọi thất bại xác thực trả CÙNG một câu**: không token, token lạ, token đã thu hồi → `unauthenticated`. So sánh hash bằng `subtle.ConstantTimeCompare` để thời gian trả lời không lộ | `auth.Static.Verify` |
| **Danh tính là token, không phải header**: `X-Operator-ID` bị xoá hẳn. Header là thứ **người gọi tự viết**, mà "ai xác nhận listing này" thì không được | `auth.go`, `products.go` |
| Hệ quả: `operator_required` (400) **không kích hoạt được từ HTTP nữa** — sau `requireOperator` thì luôn có operator. Ba assertion cũ đổi thành khách → 403 | `operator_test.go`, `logistics_test.go`, `purchase_tasks_test.go` |
| `sourced_by` rời khỏi body: provenance suy từ **loại principal**. Khách không thể khai paste của mình là do operator gõ | `sourcingOf` |
| `customer_id` rời khỏi body: đơn thuộc về người cầm token. Không còn field nào để đặt hộ tên người khác | `orders.go` |
| Kiểm chủ đơn nằm ở **handler**, không ở adapter — `CancelOrder.OnBehalfOf` do adapter điền **từ token**, người gọi không gửi được. Trả `ErrNotOwner` → **404**, vì 403 sẽ xác nhận đơn có tồn tại | `cancel_order.go`, P9-PLAN §4 câu 1–2 |
| Token lưu **hash SHA-256**, không lưu thô; `revoked_at` là timestamp chứ không DELETE — "ngừng hoạt động lúc nào" là câu hỏi kiểm toán cần | `0006_auth.sql`, `token_repo.go` |
| Không salt: token là 256 bit ngẫu nhiên **của mình**, không phải mật khẩu người đặt — không có gì để brute-force, và salt sẽ làm không tra ngược được theo hash | `auth.HashToken` |
| `wire.Postgres` **không seed token nào**. Chìa đầu tiên qua `cmd/api -bootstrap-operator-token`, và **chỉ khi bảng rỗng** — credential cố định trong DB thật là cửa sau | `cmd/api/main.go` |
| `Verifier` và `Issuer` tách interface: một route phát, ba mươi route chỉ đọc — ba mươi cái đó không được có khả năng phát | `auth/issue.go` |
| **Bug của tôi:** thêm bảng `api_tokens` nhưng `pgtest` truncate theo **danh sách viết tay** → test đọc phải dòng của test trước. Sửa tận gốc: hỏi `pg_tables` thay vì đọc danh sách | `pgtest.truncateAll` |
| **Bug của tôi (lần 2):** script vá bằng `perl -0pi` báo "PATCH 2 chỗ" nhưng **không đổi gì** — đúng loại lỗi đợt 13 nói. Làm lại bằng `awk`, và **đếm lại sau khi chạy** rồi mới báo | — |

Smoke thật (bearer auth): 20 event #69–#88 qua 5 context; `order at the end: delivered; quote vs actual:
quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD` — số vàng **không đổi** sau khi thay toàn bộ
cơ chế xác thực.

Tổng **222 test** (+21 ròng, thêm 22 xoá 1): 6 `auth/auth_test.go` · 3 `auth/issue_test.go` ·
10 `http/auth_test.go` · 3 `postgres/token_repo_test.go` · 1 xoá (`TestPOSTProducts_operatorHeaderBecomesProvenance`
— test cái header đã không còn).
Với `PORTAGE_TEST_DSN`: **221 PASS + 1 SKIP** (skip là `TestRegisterMerchant_outboxFailureRollsBackTheSave`,
cố ý — bản Postgres của nó ở `postgres_test.go` và đang xanh).

**Còn nợ của T1:** `GET /tokens` để xem/thu hồi khoá (mới có `Revoke` ở repo, chưa có route).

### Đợt 15 (06/09) — P9/T3: quét quote hết hạn, use case đầu tiên không ai gọi

| Quyết định / bug | Chỗ |
|---|---|
| **Thời gian cũng là thứ gây ra việc.** Mười tám mục WALKTHROUGH trước, use case nào cũng có người gọi: request hoặc event. Quote quá hạn thì **không có gì xảy ra cả** — không ai phát được event cho việc đó, nên phải có người đi hỏi | `ExpireQuotesHandler` |
| Và đó chính là lý do `Clock` là **port**: "quá hạn chưa" phải kiểm chứng được mà không đợi 48 giờ. Trong test 48 giờ trôi trong một dòng `clk.Advance(49h)` | `wire_test.go` |
| **Mỗi quote một transaction.** Gói cả lô vào một tx thì ngắn hơn và tệ hơn: một dòng hỏng cuốn ngược việc đã làm cho quote khác. Lô 100 quote **không phải invariant** — ranh giới transaction = ranh giới aggregate (DDD §14) | `expire_quotes.go` |
| Danh sách việc đọc **ngoài** tx, quyết định đọc lại **trong** tx. Giữa hai dòng đó khách có thể vừa accept; lưu bản cũ xuống sẽ **xoá mất cái accept** | `ByID` lần hai |
| "Đã xong rồi" ≠ "không làm được": nuốt đúng `ErrQuoteNotIssued` + `ErrQuoteStillValid`, mọi lỗi khác vẫn nổ. Cùng bài học với `ordering.Reactor` | `errAlreadySettled` |
| `LIMIT` chặn **một lượt**, không bỏ việc — lượt sau làm tiếp; `ORDER BY expires_at` cho cái chờ lâu nhất đi trước | `IssuedBefore` |
| Index **partial** `quotes_open_idx (expires_at) WHERE status='issued'` — viết từ ngày 2, dùng đúng hôm nay. Bảng `quotes` lớn mãi, index chỉ lớn theo phần **còn mở** | `0002_pricing.sql` |
| `ByID` đọc một dòng, `IssuedBefore` đọc nhiều, nhưng 22 cột chỉ biến thành `Quote` **một lần** — copy khối scan là tạo chỗ thứ hai để một cột lệch | `scanQuote`, `type scanner interface{ Scan(...) }` |
| `worker.Sweeper` là anh em của `Relay`: relay có sẵn hàng đợi lúc khởi động nên chạy ngay; sweep chỉ có đồng hồ nên **chờ một nhịp** rồi mới quét. Pass lỗi = một dòng log, không phải chết process | `worker/sweep.go` |
| Sweep chạy **cạnh** relay trong `cmd/worker`, không tách binary: vài query một phút, thêm process là thêm thứ phải deploy. `cmd/api` memory cũng bật — bản dev nơi quote sống mãi thì khác production đúng ở cái luật quote sinh ra để giữ | `cmd/worker`, `cmd/api` |
| **Bug của tôi (T1, tìm ra hôm nay):** `smoke.ps1` chỉ chạy được **một lần mỗi database**. `-bootstrap-operator-token` chỉ nổ khi `api_tokens` rỗng, mà script sinh token mới mỗi lần → lần chạy thứ hai 401 sạch từ request đầu. Sửa: script `TRUNCATE api_tokens` trước khi bật api | `scripts/smoke.ps1` |

Bài học của bug cuối: **một guard đúng vẫn có thể làm hỏng thứ khác.** Guard
"chỉ bootstrap khi bảng rỗng" là đúng và phải giữ; cái sai là script dùng nó đã
không tự tạo lại điều kiện nó cần. Lần T1 chạy smoke thấy xanh chỉ vì database
lúc ấy tình cờ rỗng — **chạy một lần không chứng minh được gì về lần thứ hai**.

Smoke thật sau T3: 20 event #104–#123 qua 5 context; `quote … total 5393720 VND, deposit 2696860`,
`variance -2.50 USD` — **số vàng không đổi**.

Tổng **229 test** (+7): 2 `app/pricing` (sweep chọn đúng quote, `limit` trải backlog qua nhiều lượt) ·
1 `postgres` (`IssuedBefore` chỉ thấy quote còn mở và đã quá hạn) · 1 `wire` (49 h trôi → `GET /quotes/{id}` = `expired`) ·
3 `worker` (pass lỗi vẫn chạy tiếp, không quét trước nhịp đầu, panic khi thiếu `Pass`).
Với `PORTAGE_TEST_DSN`: **228 PASS + 1 SKIP**.

### Đợt 16 (06/09) — P9 xong: read model, AI, tỷ giá, router ACL, quản chìa

| Quyết định / bug | Chỗ |
|---|---|
| **Read model là package DUY NHẤT không có domain** — và đó là định nghĩa, không phải thiếu sót: nó không có invariant nào để giữ, vì nó không **quyết định** gì cả. Nên `OrderSummary` là struct field exported hết, không method, không validate | `internal/app/reporting` |
| Màn hình cần dữ liệu của bốn context → **đừng JOIN**. JOIN biến bốn bounded context thành một cục không tách nổi; gọi API nội bộ thì context nào chết là màn hình chết | `order_summaries` |
| Chịu sai thứ tự mua bằng **cái khung**: bất kỳ event nào cũng được TẠO dòng. Cách sai là "event của đơn tôi chưa biết → bỏ qua", nó im lặng mất dữ liệu | `update()` |
| Nhưng có đúng MỘT field chỉ `order_placed` được đặt: `status`. Khung do `batch_shipped` tạo chưa biết trạng thái, đoán một cái là **nói dối** với khách đang nhìn màn hình | `if s.Status == ""` |
| Bản gửi lại không được **kéo lùi** timestamp | `if at.After(s.UpdatedAt)` |
| `status` (ordering sở hữu) và `tracking` (logistics sở hữu) là **hai cột**. Gộp thì tiện và sai: "đã cân ở Denver" không phải trạng thái của đơn hàng | `Tracking` |
| `GET /me/orders` chứ không `GET /orders/{customer_id}`: route **không gọi tên được** khách khác thì không lộ được khách nào — mạnh hơn việc handler nhớ so sánh hai id | `summaries.go` |
| `?status=` gõ sai → **400**, không phải list rỗng. Rỗng trông y hệt "hết việc rồi" — câu trả lời tệ nhất cho một lỗi chính tả trong màn hình vận hành | `listOrders` |
| Read model **không FK, không NOT NULL**. Đổi lại: id/tiền/thời gian bằng zero phải xuống DB thành `NULL`, không phải UUID toàn số 0 — một id giả đọc lên sẽ thành id thật | `0007_reporting.sql`, `idOrNil` |
| **ACL cho AI: port trả bốn string thô.** Nếu port trả `shared.Money` thì port đang quyết cái gì hợp lệ — việc của domain. Model trả `"about $150"` chết ở `ParseMoney`, không phải ba bảng sau | `ListingDraft` |
| Máy được **tạo nháp**, không được **xác nhận**: `SourcedByFeed` + `Publish` từ chối listing chưa verified. Tính năng mới **không nới** luật cũ | `extract.go` |
| Mọi kiểu hỏng của dịch vụ ngoài → **một** sentinel: 429, 500, mạng đứt, JSON rác, JSON đúng cú pháp sai schema. Chia nhỏ thì mỗi người gọi phải tự quyết "thử lại được không" | `ErrExtractorUnavailable` → 503 |
| Ca tinh tế: `{"product":"…"}` decode **không lỗi**, mọi field rỗng. Bắt bằng tên rỗng — cho lọt thì nổi lên thành `ErrEmptyName`, một 400 đổ lỗi cho người gọi vì thứ họ không gửi | `openai.go` |
| **Tính năng tắt phải TRÔNG như đang tắt.** Postgres không có key → `Unavailable` → 503, tuyệt đối không rơi về `Fake`: API sẽ trả 201 với tên và giá bịa, dấu hiệu duy nhất là các con số toàn hư cấu | `wire.Postgres` |
| `POST /fx` là route **duy nhất không phát event** — không ai giữ bản sao tỷ giá, vì `Quote` đã đóng băng tỷ giá vào chính nó lúc phát hành | `set_rate.go` |
| **Bẫy: hai sentinel trùng tên** `ErrInvalidRate` (`pricing` = giá trong bảng cước, `shared` = tỷ giá). Cùng tên, khác package, khác cách sửa → phải là **hai mã lỗi** | `invalid_exchange_rate` |
| ACL Router chọn adapter **theo shop**, ở composition root. `if` trong use case sẽ mọc một nhánh mỗi shop mãi mãi; một ACL biết mọi shop thì shop nào chết cũng là mọi shop chết | `merchant.Router` |
| Projection `Items` là eventual: "không biết shop" ≠ "hỏng" → giao cho người mua. Fail cái task sẽ huỷ một đơn vì một cuộc đua vài trăm ms | `aclFor` |
| Thu hồi chìa bằng **hash**, không phải token: operator nhìn danh sách chưa bao giờ thấy token. Thu hồi hai lần / hash không tồn tại → 204 (404 biến route thành cách hỏi "hash này có thật không") | `DELETE /tokens/{hash}` |
| Chìa operator **hoạt động cuối cùng** → 409. API không ai quản được không phải API an toàn hơn — nó là API chỉ sửa được bằng cách vào DB gõ tay | `errLastOperator` |
| **Bug của tôi (2):** `smoke.ps1` treo ở `Get-Content bin\worker.log` vì worker vẫn giữ file. Sửa: dừng process **trước** rồi mới đọc log — cái `finally` giờ làm đúng thứ tự đó | `scripts/smoke.ps1` |
| **Bug của tôi (3):** test `TestClient_givesUpOnASlowModel` treo 60 s. `httptest.Server.Close` **đợi handler đang chạy**, mà server chưa ghi gì thì không phải lúc nào cũng nhận ra client đã ngắt. Sửa: test tự release handler qua một channel | `openai_test.go` |
| **Bug của tôi (4):** `newAPI()` trong test thiếu `Reporting` → panic nil ở request đầu. Sửa tận gốc: `NewHandler` **panic ngay lúc dựng** nếu thiếu `Summaries`/`Registry` — dây nối sai phải chết lúc khởi động, không phải lúc có khách (quy ước 1) | `server.go` |

**Smoke thật sau P9** — read model xác nhận trên binary thật:

```
order at the end: delivered; quote vs actual: quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD
my orders: 1 row(s), Air Trainer 90 - delivered, tracking shipped, shop ref on customer view: ''
```

Số vàng **không đổi một chữ** qua cả sáu task của P9 — đúng điều một loạt thay
đổi ở tầng adapter và một tầng đọc mới phải bảo đảm.

Tổng **268 test** (+39 so với T3): 5 `app/reporting` · 3 `postgres/reporting_repos_test.go` ·
3 `http/summaries_test.go` · 2 `http/fx_test.go` · 4 `http/from_url_test.go` · 3 `http/tokens_test.go` ·
5 `adapter/openai` · 3 `adapter/merchant` · 1 `postgres/token_repo_test.go` · 1 `auth/registry_test.go` ·
9 `adapter/memory` (coverage 17,5 % → 80,8 %).
Với `PORTAGE_TEST_DSN`: **267 PASS + 1 SKIP**, 37 route, 29 dòng `Subscribe`.

**Còn nợ sau P9:** `config/ratecard.yaml` (bảng giá thật, hiện là hằng trong
`wire.quotePolicy`) · một `adapter/merchant/<shop>.go` thật (chưa shop nào cho
API) · cổng thanh toán · `scripts/smoke.sh` mới chỉ **syntax-check** trên máy
Windows (không có `jq`/`psql`), lần chạy thật đầu tiên sẽ là CI.

### Đợt 17 (08/09) — variant/size: một chữ của shop đi qua ba context

Đợt này không sinh ra từ review code mà từ **ngồi bấm thật trên UI**. Câu hỏi
của anh: *"khách nhập `M 8 / W 9.5` thì input có đỡ được không?"* — trả lời được
phần dễ (`size` là string tự do, không enum), rồi lộ ra hai lỗ hở lớn hơn nhiều.

| Quyết định / bug | Chỗ nó nằm |
|---|---|
| `POST /orders` trước đợt này nhận **bất kỳ** uuid làm `variant_id`. Khách trả cọc xong, người đi mua nhận một task không thể thực hiện. Giờ ordering giữ bảng chiếu riêng và từ chối trước khi có đơn | `ordering/variant.go`, `place_order.go` |
| Không `JOIN` sang catalog: guard 7 cấm context import context, và dùng chung database chỉ là tiện, không phải được phép. Nên catalog **kể** — `catalog.variant_added`, event thứ 36 | `catalog/events.go`, `subscribe.go` |
| Hai người nghe, **hai mức chép**: procurement giữ đủ chữ (size, màu, mã shop) vì có người phải đọc; ordering giữ đúng hai id vì luật của nó chỉ cần thế. Projection giữ dữ liệu không ai đọc sẽ mốc mà không ai biết | `procurement.Variant` (5 field) vs `ordering.Variant` (2 field) |
| `PurchaseTask.Subject` — bốn chữ **chép** vào task lúc mở, không tra lại lúc đọc: việc đã giao thì không đổi nội dung, và bảng chiếu là eventual | `procurement/task.go`, `purchase_tasks.go` |
| Thiếu dòng variant (chưa relay tới) → task vẫn mở, chỉ là không có size. Fail cái task sẽ huỷ một đơn **đã trả cọc** vì một cuộc đua vài trăm ms — cùng lối nghĩ với `ErrItemNotFound` | `OpenTask` |
| `variantKey` giờ bỏ **mọi** khoảng trắng: `M8/W9.5` == `M 8 / W 9.5`. Nhưng `8 M / 9.5 W` vẫn khác — đảo thứ tự là chữ khác của shop, mình không có quyền đoán hộ | `catalog/variant.go` |
| **Tôi tự sửa đề xuất của mình:** đề xuất ban đầu là "chặn variant không size không màu". Đọc comment `VariantDetails` thì thấy nó có chủ đích (sản phẩm một size có đúng một variant rỗng, có test hẳn hoi). Luật đúng: không tên → chỉ được là variant **duy nhất** | `ErrUnnamedVariant` |
| `variant_unknown` mang **hai nghĩa** (id giả, hoặc chưa relay tới) như `listing_not_found` — nên FE để nó vào `retryOn` chứ không báo đỏ | `web/app/steps.js` |
| `TaskDetails` lên 5 field và **chỉ field thứ 5 không bắt buộc**; guard `reflect` phải nói được câu đó, cộng một ca "Subject rỗng vẫn mở được task" | `TestOpenTask_validatesEveryField` |
| Smoke đổi còn **một** variant và giữ id của nó: golden flow phải là đúng 20 event, bằng số `web/flow.html` đi qua. Trước đó script tạo variant thứ hai ngay trước khi đặt đơn → giờ là 409 vì chưa relay | `scripts/smoke.{ps1,sh}` |
| `smoke.ps1` thêm `-Port`: một `cmd/api -web` đang giữ 8080 cho console thì smoke vẫn chạy được cạnh nó, không phải giết process của người khác | `scripts/smoke.ps1` |
| Mã SKU trong test đổi từ một mã trông như hàng thật sang `EX-AT90-9-BLK` — repo giữ quy ước không có tên thương hiệu thật, kể cả trong fixture | 6 file test |

**Smoke thật sau đợt 17** — dòng đáng giá nhất là `buy:`, vì một người đọc nó là
đi mua được:

```
buy: Air Trainer 90 / US 9 · black / ref EX-AT90-9-BLK / https://…/t/air-trainer-90/abc
order at the end: delivered; quote vs actual: quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD
```

`smoke.sh` (CI) **assert** luôn cái label chứ không chỉ in ra. Số vàng không đổi
một chữ: 2500 g · 5.393.720 ₫ · 2.696.860 ₫ · variance −2.50 USD.

Tổng **291 test** (+23 so với P9): 1 `domain/procurement` (Subject rỗng) · 4 `domain/catalog`
(khoá bỏ khoảng trắng, đảo thứ tự vẫn khác, hai chiều của `ErrUnnamedVariant`, `VariantAdded` mang chữ đã trim) ·
1 `app/procurement` (subject chép vào task) · 1 `app/ordering` (hai kiểu variant sai) ·
2 `postgres` (`procurement_variants`, `ordering_variants`) · 2 `adapter/http` (taskView có chữ, `POST /orders` 409) ·
2 `eventcodec` (payload + decode của `variant_added`) · phần còn lại là ca thêm vào test cũ.
Với `PORTAGE_TEST_DSN`: **290 PASS + 1 SKIP**, 37 route, **31 dòng** `Subscribe`, 36 domain event.

Đọc chi tiết: `docs/WALKTHROUGH.md` §22 (tám mục, có transcript và bảng bài học).

**Còn nợ** vẫn y nguyên như sau P9: bảng giá thật trong `config/`, một
`adapter/merchant/<shop>.go` thật, cổng thanh toán, và `scripts/smoke.sh` chưa
chạy thật lần nào ngoài CI (máy Windows không có `jq`/`psql`).

### Đợt 18 (08/09) — clone sang Linux, và lần chạy thật đầu tiên của `smoke.sh`

Anh clone repo về máy Linux để build và học. Máy đó không có Go nên tôi cài
Go 1.27.1 vào `~/sdk/go` (tarball, không cần root), và **đúng lần chạy thật đầu
tiên của `scripts/smoke.sh`** thì nó tìm ra một bug thật.

| Quyết định / bug | Chỗ nó nằm |
|---|---|
| **Bug thật, tìm ra bởi smoke.sh:** `Migrate` tạo `schema_migrations` **ngoài** advisory lock. `CREATE TABLE IF NOT EXISTS` của PostgreSQL **không nguyên tử**: hai câu chạy song song đều thấy bảng chưa có, rồi một cái vỡ khi chèn row type — `duplicate key value violates unique constraint "pg_type_typname_nsp_index"`. `cmd/api` và `cmd/worker` khởi động cùng lúc, nên đây là đường đi thật lúc deploy đầu | `migrate.go` |
| Sửa: đưa `CREATE TABLE` **vào trong** transaction đã giữ lock. Bốn dòng dịch chỗ, và comment cũ nói lock bảo vệ "cùng một file không áp hai lần" — đúng phần vòng lặp nhưng bỏ sót chính bảng sổ sách | `migrate.go` |
| Vì sao Windows không bao giờ thấy: database ở đó đã migrate từ lâu, nên `CREATE TABLE IF NOT EXISTS` là no-op và không có gì để đua. Lỗi **chỉ** hiện trên database trống. Bài học: "chạy xanh nhiều lần" không chứng minh gì về lần chạy **đầu tiên** | — |
| Test hồi quy tự tạo một database **dùng-một-lần** rồi cho 4 goroutine cùng gọi `Migrate` sau một barrier. Không dùng `portage_test` vì "cold" phải nghĩa là cold thật. Đua thì có tính xác suất, nên test có thể xanh oan trên code lỗi, nhưng **không bao giờ đỏ oan** trên code đúng — đó là chiều quan trọng | `TestMigrate_survivesTwoProcessesOnAColdDatabase` |
| `smoke.sh` thêm `PORTAGE_PORT` và truyền `-addr` cho api, giống cờ `-Port` của `smoke.ps1`. Trên máy Linux này 8080 đã có người, và 5432 là Postgres dev của công ty nên Portage chạy ở 5433 bằng một `docker-compose.override.yml` chỉ tồn tại ở máy đó (`ports: !override`, vì Compose mặc định **cộng dồn** danh sách port chứ không thay) | `scripts/smoke.sh` |
| `smoke.sh` bỏ `-S` ở vòng chờ api: mấy nhịp dò đầu chắc chắn trượt lúc api còn đang bind, in "Failed to connect" ở đó làm một lần chạy lành trông như hỏng | `scripts/smoke.sh` |

**Kết quả trên Linux** (Go 1.27.1, Postgres 16 trong Docker, database trống):

```
smoke.sh EXIT=0 — 20 event, #9 → #28
buy Air Trainer 90 / US 9 · black / ref EX-AT90-9-BLK / https://…/t/air-trainer-90/abc
order at the end: delivered; quote vs actual: quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD
```

Và thứ máy Windows **không** làm được vì thiếu cgo: `go test ./... -race` xanh
trên cả 22 package, không một data race nào.

Tổng **292 test** (+1 so với đợt 17): `TestMigrate_survivesTwoProcessesOnAColdDatabase`.
Với `PORTAGE_TEST_DSN`: **291 PASS + 1 SKIP**; không có DSN: 262 PASS + 30 SKIP.
Con số giống nhau trên cả Windows và Linux.

**Nợ đã trả:** `scripts/smoke.sh` từ đợt 16 tới giờ mới chỉ syntax-check trên
Windows. Giờ nó đã chạy thật, trên Linux, trên database trống, và ngay lần đầu
đã bắt được một bug mà 292 test không bắt.
