# Portage — Sổ tay dựng môi trường & khởi tạo source

> Sổ tay **tra cứu**: máy chạy nó thế nào, code nằm ở đâu, quy ước gì.
> Cập nhật: 2026-09-08 · Số liệu theo `HEAD` (P10 đang dở, chưa tính vào đây).
>
> | Muốn gì | Đọc file nào |
> |---|---|
> | học DDD theo khái niệm | [DDD.md](DDD.md) |
> | tra cú pháp Go ↔ PHP | [GO-CHO-PHP.md](GO-CHO-PHP.md) |
> | đọc code theo thứ tự | [WALKTHROUGH.md](WALKTHROUGH.md) |
> | lộ trình học + câu tự kiểm | [HOC.md](HOC.md) |
> | **vì sao code trông như vậy** | **§9 của file này** — 18 đợt review |

---

## 0. Project này là gì

**Portage** — đặt hàng hộ Mỹ → Việt Nam. Viết bằng **Go**, kiến trúc **DDD**.
Vừa là sản phẩm thật, vừa là project học Go + DDD.

Bối cảnh nghiệp vụ đầy đủ: `CLAUDE.md` và [CATALOG.md](CATALOG.md).

---

## 1. Máy cần có gì

| Công cụ | Bản đang dùng | Ghi chú |
|---|---|---|
| Go | 1.27.0 | bắt buộc |
| Git | 2.51 | bắt buộc |
| Docker | 29.7.2 | chỉ cần khi chạy Postgres (`docker compose up -d`) |

---

## 2. Cài Go

```powershell
winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements
```

Không có winget thì tải `.msi` ở <https://go.dev/dl/>.

> ⚠️ **Bẫy: cài xong gõ `go` báo "not found".** Bộ cài *đã* thêm PATH, nhưng
> **terminal đang mở không tự cập nhật**. Cách đúng: **đóng hẳn rồi mở lại**
> terminal. Cần dùng ngay thì set tạm:
>
> ```bash
> export PATH="$PATH:/c/Program Files/Go/bin"     # Git Bash
> $env:Path += ";C:\Program Files\Go\bin"         # PowerShell
> ```

Kiểm tra: `go version` → `go version go1.27.0 windows/amd64`

**Khác Composer:** Go **không** có `vendor/` trong project. Thư viện nằm chung ở
`GOMODCACHE` (`go env GOMODCACHE`), mọi project dùng chung → repo nhẹ.

---

## 3. Khởi tạo source (đã chạy một lần, ghi lại để dựng lại được)

```bash
go mod init github.com/duongsy/portage   # ~ composer init; tham số là namespace gốc
git init -b main
go get github.com/google/uuid@v1.6.0     # thư viện ngoài DUY NHẤT của domain
```

| Go | Symfony/PHP |
|---|---|
| `go.mod` | `composer.json` |
| `go.sum` | `composer.lock` |
| `go get` | `composer require` |
| `go mod download` | `composer install` |
| `go mod tidy` | `composer update` |

Sau `go get` nhớ commit **cả `go.mod` và `go.sum`**.

### 3b. VS Code — Ctrl+Click nhảy vào hàm

Cần **đúng hai thứ**, thiếu cái thứ hai là Ctrl+Click không chạy:

| Thứ | Vai trò | Bên PHP |
|---|---|---|
| extension `golang.go` | vỏ nối VS Code với Go | extension Intelephense |
| **`gopls`** | language server — thứ THẬT SỰ hiểu code | engine của Intelephense |

```bash
code --install-extension golang.go
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

| Phím | Việc | Intelephense gọi là |
|---|---|---|
| `Ctrl+Click` / `F12` | nhảy vào định nghĩa | Go to Definition |
| `Alt+←` | quay lại chỗ cũ | Go Back |
| `Ctrl+Shift+F12` | mọi nơi đang gọi hàm này | Find All References |
| `Ctrl+T` | tìm theo tên hàm/kiểu toàn project | Go to Symbol in Workspace |
| `F2` | đổi tên toàn project, an toàn | Rename Symbol |
| `Ctrl+.` | sửa nhanh (thêm import, tách biến) | Quick Fix |

> **Riêng Go, PHP không có:** `Ctrl+Click` vào một **interface** sẽ hỏi *"đi tới
> định nghĩa hay đi tới các cài đặt?"* — Go không có `implements`, đây là cách
> duy nhất tìm ra struct nào thoả interface đó.

---
## 3b. Git — tách danh tính cá nhân khỏi danh tính công ty

> Mục này từng bị bỏ khi file được viết lại gọn hơn, và được **lấy lại nguyên văn**: nó là
> thứ giữ cho commit cá nhân không mang email công ty, và cái cảnh báo "không sửa global"
> dưới đây là lý do duy nhất nó không xảy ra.

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

## 4. Cấu trúc thư mục — cây DUY NHẤT của repo

```
portage/
├── go.mod · go.sum             ~ composer.json / composer.lock
├── .editorconfig               LF, tab cho Go — mọi editor đọc file này
├── .github/workflows/ci.yml    gofmt · vet · test -race · luật import domain (§5b)
├── docker-compose.yml          postgres:16 — db `portage` (dev) + `portage_test` (test)
├── docker/initdb/              SQL chạy một lần khi volume trống: tạo portage_test
├── config/                     bảng giá, thuế — DỮ LIỆU (còn là hằng trong wire.go)
├── web/                        BỐN trang, không npm, không build
│   ├── index.html              trang chọn; `/` cũng nhảy về đây
│   ├── customer.html           trang khách  ─┐ React (thư viện trong web/vendor/)
│   ├── staff.html              trang nhân viên ┘ dùng chung web/app/*.js
│   ├── flow.html               mô phỏng 26 bước, MỘT file, không gọi API
│   ├── console.html            bảng kiểm API (vanilla, cho lập trình viên)
│   ├── vendor/                 React + ReactDOM + htm, kèm trong repo
│   └── app/                    portage.js (chỗ DUY NHẤT gọi API) · words.js
│                               (mã của máy → chữ người đọc) · ui.js · screens.css
├── scripts/
│   ├── smoke.ps1               cả flow trên binary thật + Postgres thật (Windows)
│   └── smoke.sh                bản Linux — CHẠY TRONG CI, assert số vàng
│
├── cmd/                        nơi để hàm main() — mỗi thư mục một chương trình
│   ├── api/main.go             -addr, -dsn, -web → wire.Memory | wire.Postgres → serve
│   └── worker/main.go          -dsn bắt buộc → Relay.Run + Sweeper (-sweep)
│
└── internal/                   Go CẤM module khác import thư mục tên `internal`
    │
    ├── domain/                 TRÁI TIM — nghiệp vụ thuần, KHÔNG import gì bên ngoài
    │   ├── decisions_test.go   7 test canh kiến trúc bằng go/ast (§6b)
    │   │
    │   ├── shared/             Shared Kernel — 46 test, 90,7 %
    │   │   ├── money.go        Money — cộng/trừ/nhân tỷ lệ/đổi tiền
    │   │   ├── decimal.go      parseDecimal, addExact/subExact/mulExact (private)
    │   │   ├── allocate.go     Allocate — chia tiền floor + largest remainder
    │   │   ├── currency.go     Currency, CurrencyFromCode
    │   │   ├── rate.go         Rate (tỷ lệ %), ExchangeRate (12 số lẻ)
    │   │   ├── weight.go       Weight, Dimensions, quy đổi thể tích
    │   │   ├── parcelspec.go   ParcelSpec — cân + hộp (DỜI từ catalog: pricing cũng dùng)
    │   │   ├── id.go           ID — UUID v7, sinh trong domain
    │   │   ├── event.go        Event, Events — aggregate GHI event
    │   │   └── operator.go     OperatorID — "ai làm", catalog và procurement cùng dùng
    │   │
    │   ├── catalog/            17 event — 47 test, 89,1 %
    │   │   ├── merchant.go     Merchant (root), MerchantStatus, SourcingMode, Hostname
    │   │   ├── product.go      Product (root) — draft → published → retired
    │   │   ├── variant.go      Variant — entity con; khoá chống trùng bỏ MỌI khoảng trắng
    │   │   ├── category.go     CategoryPolicy (VO khoá tự nhiên), Restriction
    │   │   ├── provenance.go   Provenance — nguồn + thời điểm + ai xác nhận
    │   │   ├── sourceurl.go    SourceURL — link tham chiếu, KHÔNG fetch (ràng buộc pháp lý)
    │   │   ├── freeshipping.go FreeShipping — VO ba trạng thái
    │   │   └── snapshot.go     MerchantSnapshot/ProductSnapshot ↔ FromSnapshot
    │   │
    │   ├── pricing/            4 event — 14 test, 73,6 %
    │   │   ├── lane.go         LaneCode, GoodsClass, RateCard, DutyPolicy, ShippingLane
    │   │   ├── policy.go       MarginPolicy (max %, sàn), QuotePolicy
    │   │   ├── calc.go         QuoteInputs, Breakdown, Calculate — DOMAIN SERVICE thuần
    │   │   ├── quote.go        Quote (root) — IssueQuote, Accept, Expire
    │   │   └── listing.go      Listing, CategoryProfile — projection từ event catalog
    │   │
    │   ├── ordering/           8 event — 4 test, 81,4 %
    │   │   ├── order.go        CustomerOrder (root) — cọc 50 %, Refund, 8 method
    │   │   ├── acceptedquote.go AcceptedQuote — projection từ pricing.quote_accepted
    │   │   └── variant.go      Variant{Variant, Product} — bản sao MỎNG nhất, không giữ size
    │   │
    │   ├── procurement/        3 event — 4 test, 80,6 %
    │   │   ├── task.go         PurchaseTask (root), PurchaseReceipt (Actual đầu tiên),
    │   │   │                   Subject (mua gì, bằng chữ — chép lúc mở việc)
    │   │   └── repository.go   + PORT MerchantACL  ← Anti-Corruption Layer #1
    │   │
    │   └── logistics/          5 event — 2 test, 76,1 %
    │       ├── parcel.go       Parcel — Receive/AssignToBatch/MarkShipped
    │       ├── batch.go        ConsolidationBatch — AddParcel/Close/Ship
    │       └── allocator.go    FreightAllocator (PORT + Domain Service), LaneRule
    │
    ├── contracts/              PUBLISHED LANGUAGE — DTO *V1 đi qua ranh giới context
    │
    ├── app/                    Use case — gọi domain, gọi repository
    │   ├── ports.go            Clock, UnitOfWork, Outbox — cái tầng app CẦN
    │   ├── deps.go             MustHave — kiểm dây nối lúc khởi động
    │   ├── catalog/            catalogapp — 14 test, 77 % · 7 use case
    │   ├── pricing/            pricingapp — 7 test, 72,9 % · IssueQuote, AcceptQuote,
    │   │                       ExpireQuotes (không ai gọi), Reconciler, Projector
    │   ├── ordering/           orderingapp — 4 test, 70,8 % · PlaceOrder, Pay*, Cancel,
    │   │                       Projector (nghe pricing), Reactor (nghe procurement)
    │   ├── procurement/        procurementapp — 4 test, 81,5 % · OpenTask (hỏi ACL)
    │   ├── logistics/          logisticsapp — 1 test, 78,3 % · cả flow kho
    │   └── reporting/          reportingapp — 8 test, 79,8 % · HAI read model, KHÔNG có domain
    │                           order_summaries (đơn) + product_worklist (việc còn phải làm)
    │                           12 handler nghe CẢ NĂM context
    │
    ├── adapter/                THẾ GIỚI BÊN NGOÀI cắm vào đây
    │   ├── http/               httpapi — 43 test, 80,7 % · 41 route, auth bọc CẢ mux
    │   │   ├── server.go       NewHandler(Deps{...}) → http.Handler
    │   │   ├── auth.go         authenticate() fail-closed, requireOperator/Customer/Any
    │   │   ├── errors.go       errorTable + writeError → 400/401/403/404/409/500/503
    │   │   ├── decode.go       decodeJSON (DisallowUnknownFields), language(), amount
    │   │   └── …               categories · merchants · products · quotes · orders ·
    │   │                       purchase_tasks · logistics · lanes · fx · tokens · summaries
    │   ├── postgres/           30 test tích hợp, 79,8 % · pgx, repo qua Snapshot
    │   │   ├── migrate.go      MỌI thứ trong MỘT tx sau pg_advisory_xact_lock —
    │   │   │                   kể cả CREATE TABLE schema_migrations (bug đợt 18)
    │   │   ├── migrations/     0001 catalog · 0002 pricing · 0003 ordering ·
    │   │   │                   0004 procurement · 0005 logistics · 0006 auth ·
    │   │   │                   0007 reporting · 0008 variant_subject ·
    │   │   │                   0009 product_requester · 0010 product_worklist ·
    │   │   │                   0011 requested_variant
    │   │   ├── outbox.go       Append (cùng tx) · Pending (FOR UPDATE SKIP LOCKED)
    │   │   ├── token_repo.go   Verify dùng subtle.ConstantTimeCompare; Revoke ghi cột
    │   │   └── pgtest/         Pool(t): skip nếu không DSN, advisory lock, truncate
    │   ├── memory/             15 test, 80,8 % · repo CẢ 6 context + Outbox + UnitOfWork
    │   ├── eventcodec/         5 test, 98,6 % · Encode 37 event viết tay = HỢP ĐỒNG
    │   ├── merchant/           3 test, 88,9 % · Manual + Router  ← ACL #1
    │   └── openai/             5 test, 80,9 % · Client · Fake · Unavailable  ← ACL #2
    │
    ├── worker/                 9 test, 87,3 % · Relay (at-least-once) + Bus + Sweeper
    │
    └── platform/
        ├── auth/               10 test, 88,8 % · Principal, HashToken, Static
        ├── clock/              System (time.Now) và Fixed (test)
        └── wire/               4 test, 94,9 %
            ├── wire.go         Memory() / Postgres() → Graph{...}; seed lane + fx
            └── subscribe.go    31 dòng định tuyến = vòng đời đơn hàng
```

### Luật vàng: chiều mũi tên phụ thuộc

```
   cmd  ──▶  app  ──▶  domain
                ▲
   adapter ─────┘         domain KHÔNG BAO GIỜ trỏ ra ngoài
```

| | `internal/domain/` |
|---|---|
| **KHÔNG được import** | driver DB · framework HTTP · SDK OpenAI · bounded context khác |
| **Được import** | stdlib + allowlist — hiện chỉ `github.com/google/uuid` |
| **Ai canh** | CI (§5b) **và** `decisions_test.go` guard 6, 7 — vi phạm là build đỏ |

**Cái giá nếu lỡ import:** test domain sẽ cần Docker, cần API key, chạy chậm.
Làm đúng → test domain chạy trong **vài mili-giây**, không cần gì cả.

> **So với Symfony:** ở Symfony `Entity` dính chặt Doctrine — `#[ORM\Column]` nằm
> ngay trong class nghiệp vụ. Ở đây ngược lại: struct nghiệp vụ hoàn toàn sạch,
> phần map xuống DB nằm riêng ở `adapter/postgres`.

**Vì sao tên là `internal/`?** Đây là quy ước **được chính compiler Go ép buộc**:
package dưới thư mục tên `internal` chỉ import được bởi code cùng module. Không
cần tài liệu nhắc, không cần reviewer canh — compiler chặn thẳng.

---
## 5. Lệnh dùng hàng ngày

| Việc | Go | Symfony / PHP |
|---|---|---|
| chạy toàn bộ test | `go test ./...` | `vendor/bin/phpunit` |
| test kèm chi tiết | `go test -v ./...` | `phpunit --testdox` |
| test một package | `go test ./internal/domain/shared` | `phpunit tests/Shared` |
| đúng MỘT test | `go test -run TestMoney_convert ./...` | `phpunit --filter` |
| bỏ qua cache, chạy thật | `go test -count=1 ./...` | *(PHP không cache)* |
| đo độ phủ | `go test -cover ./...` | `phpunit --coverage-text` |
| soi lỗi tĩnh | `go vet ./...` | `phpstan analyse` |
| format | `gofmt -w .` | `php-cs-fixer fix` |
| chạy thẳng | `go run ./cmd/api` | `php bin/console …` |
| biên dịch ra binary | `go build ./cmd/api` | *(PHP không có)* |

> `./...` = "thư mục này **và mọi thư mục con**". Nhớ dấu ba chấm.
>
> ⚠️ Go **cache kết quả test**. Thấy `(cached)` là nó không chạy lại. Ép chạy
> thật: `-count=1`.

**Cổng chất lượng — gõ trước mỗi lần commit:**

```bash
gofmt -l . && go vet ./... && go test -count=1 ./...
```

### 5b. CI — `.github/workflows/ci.yml`

Mỗi lần push, máy Linux của GitHub chạy 4 bước:

| Bước | Lệnh | Bắt lỗi gì | Symfony |
|---|---|---|---|
| gofmt | `gofmt -l .` phải rỗng | quên format | `php-cs-fixer --dry-run` |
| vet | `go vet ./...` | printf sai kiểu, copy lock | `phpstan` |
| test | `go test -race -count=1 -cover ./...` | test fail + **data race** | `phpunit` |
| luật import | `go list -f '{{range .Imports}}…'` | domain import ngoài allowlist | `deptrac` |

Bước cuối là **kiến trúc được kiểm bằng máy**: không ai phải nhớ luật "domain
không import driver DB" — lỡ vi phạm là build đỏ.

### 5c. Lệnh riêng của project này

Cheatsheet `go` đầy đủ đã có ở `go help`. Đây là những lệnh **chỉ project này cần**:

```bash
# Domain có đang import thứ ngoài allowlist không?  (rỗng = sạch)
go list -f '{{range .Imports}}{{println .}}{{end}}' ./internal/domain/... \
  | sort -u | grep '\.' | grep -Ev '^github\.com/google/uuid$'

# KIỂM CHỨNG email đã đi vào commit (§3b) — đừng tin config, hãy đọc commit
git log -1 --format='%an <%ae>'

# Chạy riêng 7 guard kiến trúc
go test ./internal/domain/ -v
```

> ⚠️ Đừng dùng `go list -deps` — nó in cả dependency **gián tiếp** (`runtime`,
> `os`, `syscall`… do `fmt` kéo theo) làm tưởng domain bẩn. `.Imports` mới là
> import trực tiếp.

| Cờ hay quên | Nghĩa |
|---|---|
| `-count=1` | bỏ qua cache, chạy thật |
| `-race` | dò data race (**cần gcc — Windows không chạy**, CI lo) |
| `-failfast` | dừng ở test fail đầu tiên |
| `-timeout 90s` | **luôn truyền** với test tích hợp (§5e bài học 2) |

### 5d. Chạy API thật và bấm thử

```bash
# In-memory — không cần gì, Ctrl+C là mất dữ liệu
go run ./cmd/api                 # :8080
go run ./cmd/api -web ./web      # + console thật ở /ui/console.html

# Postgres — cần Docker đang chạy
docker compose up -d
DSN="postgres://portage:portage@localhost:5432/portage?sslmode=disable"
go run ./cmd/api    -dsn "$DSN"
go run ./cmd/worker -dsn "$DSN" -every 500ms -sweep 10s
```

**Nhanh nhất — cả flow trên binary thật, tự dọn:**

```bash
bash scripts/smoke.sh                                          # Linux (CI chạy cái này)
powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1     # Windows
```

Bấm tay từng bước trên giao diện: [UI-GUIDE.md](UI-GUIDE.md).

#### Chìa khoá — mọi route đều cần `Authorization: Bearer <token>`

```bash
# Memory: hai token dev in sẵn trong wire.go
OPTOK=dev-operator ; CUSTOK=dev-customer

# Postgres: chìa đầu tiên CHỈ tạo được khi bảng api_tokens còn rỗng
OPTOK=$(openssl rand -hex 32)
go run ./cmd/api -dsn "$DSN" -bootstrap-operator-token "$OPTOK"
# rồi operator cắt chìa cho khách — token trả về MỘT LẦN, kho chỉ giữ hash
CUSTOK=$(curl -s -X POST localhost:8080/tokens -H "Authorization: Bearer $OPTOK" \
  -H 'Content-Type: application/json' -d '{"kind":"customer","label":"app"}' | jq -r .token)
```

#### Mười bước theo đúng thứ tự nghiệp vụ

`OP="Authorization: Bearer $OPTOK"` · `CU="Authorization: Bearer $CUSTOK"`

```bash
# 1. Đăng ký shop — "50,00" là dấu THẬP PHÂN vì Accept-Language: vi
curl -X POST localhost:8080/merchants -H "$OP" -H 'Accept-Language: vi' \
  -d '{"name":"Example Sports","site":"www.example.com","currency":"USD",
       "free_shipping":{"kind":"over","threshold":"50,00"},"sourcing":["operator"]}'   # 201 → M

# 2. Sản phẩm khách dán — KHÔNG có field sourced_by: provenance suy từ TOKEN
curl -X POST localhost:8080/products -H "$CU" \
  -d '{"name":"Air Trainer 90","merchant_id":"<M>","category":"footwear",
       "source_url":"https://www.example.com/t/x","price":"150.00","currency":"USD"}'  # 201 → P

# 3. Publish sớm — chưa có variant, nghiệp vụ từ chối
curl -X POST localhost:8080/products/<P>/publish -H "$OP"                    # 409 no_variants

# 4. Ba bước operator — KHÔNG còn X-Operator-ID: operator LÀ chủ token
curl -X POST localhost:8080/products/<P>/variants -H "$OP" -d '{"size":"US 9","color":"black"}'
curl -X POST localhost:8080/products/<P>/confirm-listing -H "$OP"                          # 204
curl -X POST localhost:8080/products/<P>/measure -H "$OP" \
  -d '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}'                    # 204
curl -X POST localhost:8080/products/<P>/publish -H "$OP"                                  # 204

# 5. Báo giá — NGAY sau publish có thể 404: worker chưa relay (memory: đợi 200 ms)
curl -X POST localhost:8080/quotes -H "$CU" -d '{"product_id":"<P>","lane":"us_forwarder"}' # 201 → Q
curl localhost:8080/quotes/<Q> -H "$CU"          # total 5393720 VND, deposit 2696860
curl -X POST localhost:8080/quotes/<Q>/accept -H "$CU"   # 204; operator gọi → 403

# 6. Đặt hàng — chủ đơn LÀ chủ token, không còn customer_id trong body
curl -X POST localhost:8080/orders -H "$CU" -d '{"quote_id":"<Q>","variant_id":"<V>"}'      # 201 → O
curl -X POST localhost:8080/orders/<O>/deposit -H "$OP" -H 'Accept-Language: vi' \
  -d '{"amount":"2.696.860","currency":"VND"}'                            # 204; sai số → 409

# 7. Đi mua — worker relay deposit_paid → procurement tự mở task
curl localhost:8080/purchase-tasks -H "$OP"                          # rỗng = relay chưa chạy → T
curl -X POST localhost:8080/purchase-tasks/<T>/confirm -H "$OP" \
  -d '{"reference":"NK-20260905-001","paid":"163.22","currency":"USD"}'   # 204
# → đơn về "purchased" = ĐIỂM KHÔNG THỂ QUAY ĐẦU; huỷ sau đây mất cọc

# 8. Kho — purchase_confirmed cũng bảo kho chờ một thùng
curl localhost:8080/parcels -H "$OP"                                                 # → PA
curl -X POST localhost:8080/parcels/<PA>/receive -H "$OP" \
  -d '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}'   # 204 — CÂN THẬT
curl -X POST localhost:8080/batches -H "$OP" -d '{"lane":"us_forwarder"}'            # 201 → B
curl -X POST localhost:8080/batches/<B>/parcels -H "$OP" -d '{"parcel_id":"<PA>"}'   # 204
curl -X POST localhost:8080/batches/<B>/close -H "$OP"                # 204; rỗng → 409
curl -X POST localhost:8080/batches/<B>/ship  -H "$OP" -d '{"freight":"27.50","currency":"USD"}'
curl -X POST localhost:8080/orders/<O>/balance -H "$OP" -d '{"amount":"2696860","currency":"VND"}'
curl -X POST localhost:8080/orders/<O>/deliver -H "$OP"                              # 204

# 9. Đối soát Quote vs Actual — SỐ VÀNG
curl localhost:8080/reconciliations/<O> -H "$OP"
# → quoted 163.22 + 25.00 · actual 163.22 + 27.50 · variance -2.50 USD

# 10. Màn hình + quản chìa
curl "localhost:8080/me/orders" -H "$CU"                 # khách: đơn của CHÍNH MÌNH
curl "localhost:8080/orders?status=deposited" -H "$OP"   # operator: hàng đợi việc
curl localhost:8080/tokens -H "$OP"                      # hash + kind + label (KHÔNG có token)
curl -X DELETE localhost:8080/tokens/<HASH> -H "$OP"     # chìa operator cuối → 409
```

**Thử phá:** bỏ `Accept-Language` ở bước 1 → threshold thành **5000.00 USD** (`en`
coi dấu phẩy là hàng nghìn). Gõ `"nmae"` thay `"name"` → 400 `bad_json`. Gửi lại
đúng request 2 → 201 nhưng sản phẩm mới bị **cắm cờ nghi trùng**.

#### Bảng mã lỗi — bốn cuộc hội thoại khác nhau

| Status | Nghĩa | Ví dụ `code` |
|---|---|---|
| **400** | không hiểu được → **sửa request** | `bad_json` · `malformed_amount` · `invalid_id` |
| **401 / 403** | **ai đang gọi** → 401 không biết anh; 403 biết, nhưng không phải của anh | `unauthenticated` · `forbidden` |
| **404** | trỏ vào thứ không tồn tại | `*_not_found` · `listing_not_found` = chưa publish **hoặc** relay chưa chạy |
| **409** | request đúng, **nghiệp vụ từ chối** → **sửa trạng thái** | `no_variants` · `quote_expired` · `wrong_amount` · `batch_empty` |
| **503** | request đúng, **thứ giúp mình không có** → thử lại sau | `extractor_unavailable` |
| **500** | bug hoặc sự cố; chi tiết CHỈ ở log | `internal` |

> **Vì sao khách hỏi đơn người khác trả 404 chứ không 403:** 403 sẽ **xác nhận
> đơn đó có thật**. 404 không tiết lộ gì. (WALKTHROUGH §18d)

Bảng đầy đủ: `internal/adapter/http/errors.go` — một `errorTable`, một `writeError`.

### 5e. Test tích hợp với Postgres

```bash
docker compose up -d
export PORTAGE_TEST_DSN="postgres://portage:portage@localhost:5432/portage_test?sslmode=disable"
go test -count=1 -timeout 90s ./...
```

| Không có biến | Có biến |
|---|---|
| `adapter/postgres` + `platform/wire` **tự `t.Skip`** | connect → migrate → truncate → chạy thật |
| unit test không bao giờ cần Docker | 32 test tích hợp |

**Ba điều đã học bằng cách vấp:**

1. **`go test ./...` chạy các package SONG SONG.** Hai package cùng `TRUNCATE`
   một DB thì đỏ 1/3 lần. Không dùng `-p 1` (che bệnh) — `pgtest.Pool(t)` giữ
   **advisory lock** của Postgres trên một connection riêng suốt test: package
   nào có khoá thì sở hữu DB, package khác chờ. Một DB, song song tuỳ ý.
2. **Luôn truyền `-timeout`.** Test treo mà không có timeout thì `go test` đợi
   10 phút.
3. **Giết terminal KHÔNG giết `go test`.** Process `*.test` sống tiếp và **giữ
   khoá** → mọi lần chạy sau đều treo. Khi test tích hợp treo bất thường:
   ```bash
   pkill -f '\.test$'
   docker compose exec -T postgres psql -U portage -d portage_test \
     -c "select pid, state, query from pg_stat_activity where datname='portage_test';"
   ```

CI có `services: postgres` và set biến này → test tích hợp luôn chạy trên GitHub.

---
## 6. Đã viết được gì

```
5 bounded context + 1 tầng đọc  ·  37 domain event  ·  41 route
35 dòng Subscribe  ·  305 test (304 PASS + 1 SKIP cố ý)
272 test chạy < 2 giây, KHÔNG cần Docker  ·  7 test canh kiến trúc bằng go/ast
```

| Context | Aggregate root |
|---|---|
| `catalog` | Merchant, Product ⊃ Variant |
| `pricing` | Quote, Reconciliation (Quote vs Actual) |
| `ordering` | CustomerOrder (cọc 50 %, điểm không thể quay đầu) |
| `procurement` | PurchaseTask + port MerchantACL |
| `logistics` | Parcel, ConsolidationBatch + FreightAllocator |
| `reporting` | *(không có domain — read model, không invariant nào)* |

Coverage từng package: xem cây ở §4. Chạy lại: `go test -cover ./...`

### 6b. Test canh quyết định — `internal/domain/decisions_test.go`

**Vì sao có nhóm test này:** ngày 04/09, `CategoryPolicy` bị thêm `hsCode` và
`dutyRate` — **vài ngày sau khi đã ghi hai lần** rằng tuyến vận chuyển đang gộp
thuế vào giá /kg và không tách dòng thuế. **Mọi unit test đều xanh.** Thiết kế
vẫn sai.

> Unit test hỏi *"hàm này trả đúng chưa?"*.
> Test canh quyết định hỏi *"code này còn tuân quy ước đã chốt không?"*.
> Hai câu hỏi khác nhau — và cái thứ hai mới là cái hay bị quên.

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

Mỗi guard ghi rõ ba thứ: **quyết định là gì**, **ghi ở tài liệu nào**, **làm gì
nếu nó đỏ** — kể cả trường hợp *"quyết định đã đổi, sửa guard đi"*. Đây là
tripwire, không phải điều răn.

---

### 6c. Mười quy ước code

> Mã số 1–10 dưới đây **được code trỏ vào** (`decisions_test.go` "convention 2",
> `merchant.go` "quy ước 5"…). Đổi số là hỏng 47 chỗ — thêm quy ước thì đánh
> số 11, đừng chèn giữa.

#### Quy ước 1 — Input sai của ai?

| | |
|---|---|
| **Luật** | input của **người dùng** sai → `error`. Input của **lập trình viên** sai → `panic` |
| **Symfony** | `error` ~ `\DomainException`; `panic` ~ `\LogicException` (không nên bắt) |
| **Code** | mọi `Parse*`/`New*` trả `error`; mọi `Must*` panic |

```go
// Validating — input KHÔNG tin được (form, CSV, dòng DB)
ParseMoney, ParsePercent, NewWeight, ParseID, ParseHostname, RegisterMerchant

// Literal — lập trình viên gõ tay, sai là bug
NewMoney, Grams, Kilos, RatePPM, MustParseHostname, MustProvenance
```

**Adapter nhận input từ ngoài → LUÔN dùng loại validating.** `panic` ở đây nghĩa
là "code của mình có bug", **không bao giờ** là lỗi của người dùng.

#### Quy ước 2 — Tiền là `int64` đơn vị nhỏ nhất

| | |
|---|---|
| **Luật** | không bao giờ `float`. Nhân/chia qua `math/big`. Half-up, làm tròn **xa số 0** |
| **Symfony** | `moneyphp/money` — cùng lý do. `Money::allocate()` = `shared.Allocate` |
| **Code** | `internal/domain/shared/money.go`, `decimal.go` |
| **Máy canh** | guard 1 — `float64` lọt vào domain là **build đỏ** |

```go
// SAI — 0.1 + 0.2 != 0.3, và không ai thấy cho tới lúc đối soát
var price float64 = 150.00

// ĐÚNG — 15000 xu, số nguyên, không có "gần đúng"
shared.NewMoney(15_000, shared.USD)
```

**Vì sao half-up *xa số 0*:** `13.215 → 13.22` **và** `-13.215 → -13.22`. Hoàn
tiền phải bằng đúng số thuế đã thu, đổi dấu. Làm tròn *về* 0 thì hoàn thiếu 1 xu.

**Tràn `int64` thì `panic`, không phải `error`** — 9,2 tỷ tỷ đơn vị nhỏ nhất là
92 triệu tỷ đô. Không nghiệp vụ nào chạm tới đó, nên chạm = bug (quy ước 1).
Bug thật đã gặp: §9 đợt 2 #6.

#### Quy ước 3 — Parse số nghiêm ngặt: chỉ nhận `-?digits[.digits]`

| | |
|---|---|
| **Luật** | không dấu phẩy, không `+`, không `150.`, không `.5` |
| **Symfony** | `NumberFormatter` theo locale — nhưng để ở tầng **controller**, không phải Entity |
| **Code** | `shared.parseDecimal` — dùng chung cho Money, Rate, ExchangeRate |

**Vì sao:** `"150,50"` là **150,50** với người Việt nhưng là **15.050** với người
Mỹ. Domain mà đoán thì sai với một trong hai. Chuẩn hoá theo locale là việc của
adapter, **trước khi** gọi vào domain (`Accept-Language` → `normalizeAmount`).

Bug thật đã gặp: §9 đợt 1 #2.

#### Quy ước 4 — Cân nặng: MỌI bước làm tròn đều làm tròn LÊN

| | |
|---|---|
| **Luật** | kể cả bước trung gian `VolumetricWeight` |
| **Code** | `shared/weight.go` |

```
SAI :  thùng 32×27×11 cm = 1900,8 g → cắt xuống 1900 g → bill 1.900 kg
ĐÚNG:                      1900,8 g → làm tròn lên 1901 g → bill 2.000 kg
```

**Vì sao:** 1900 rơi đúng dưới mốc bậc cân 100 g → **mất nguyên một bậc**, mình
chịu lỗ. Trong chuỗi tính phí, một bước cắt xuống là mất một bậc.
Bug thật đã gặp: §9 đợt 1 #1.

#### Quy ước 5 — ID là UUID v7, sinh trong domain, mỗi aggregate một kiểu riêng

| | |
|---|---|
| **Luật** | `type OrderID struct{ shared.ID }` — compiler phân biệt `OrderID` ≠ `MerchantID` |
| **Symfony** | `symfony/uid` — nhưng Doctrine sinh id lúc `flush()`, ở đây sinh ngay |
| **Code** | `shared/id.go` |

```go
type OrderID struct{ shared.ID }
func NewOrderID() OrderID { return OrderID{shared.NewID()} }
```

Không có `private ?int $id = null` chờ `flush()`. **Object có danh tính từ dòng
đầu constructor.** v7 có timestamp ở đầu nên index B-tree ghi nối đuôi, không
rải rác như v4.

#### Quy ước 6 — Aggregate chỉ GHI event, tầng app mới PHÁT

| | |
|---|---|
| **Luật** | `o.Record(ev)` trong domain · `o.PullEvents()` → Outbox trong app, **sau** khi Save |
| **Symfony** | `DomainEvent` gom trong Entity rồi dispatch ở Handler — cùng ý |
| **Code** | `shared/event.go`; khuôn ở mọi `app/*/`; `adapter/*/outbox.go` |

```go
// ĐÚNG — thứ tự này không được đảo
repo.Save(ctx, o)                       // 1. lưu trạng thái
outbox.Append(ctx, o.PullEvents())      // 2. ghi event, CÙNG transaction
```

**Không bao giờ publish trước khi lưu.** Đảo hai dòng này là event bay ra cho
một trạng thái chưa tồn tại. Test bắt: `TestRelay_*` và test rollback.

#### Quy ước 7 — Domain không đọc đồng hồ

| | |
|---|---|
| **Luật** | method nhận `now time.Time` qua tham số. KHÔNG `time.Now()` trong domain |
| **Symfony** | inject `ClockInterface` (PSR-20) thay vì `new \DateTime()` |
| **Code** | `app/ports.go` (`Clock`) · `platform/clock/` (`System`, `Fixed`) |
| **Máy canh** | guard 2 — `time.Now()` trong domain là **build đỏ** |

Tầng app lấy `now` **một lần** ở đầu handler rồi truyền xuống. Production cắm
`clock.System`, test cắm `clock.Fixed` → test về hạn 48 h không cần `sleep`.

#### Quy ước 8 — `internal/app/<context>/`, không để phẳng

| | |
|---|---|
| **Luật** | mỗi bounded context một thư mục con, package tên `<context>app` |
| **Vì sao tên khác** | `package catalog` sẽ **trùng tên** với domain → mọi file phải alias import |

Khuôn một use case — sáu bước, mọi handler đều giống nhau:

```
now → InTx → (load) → domain → Save → PullEvents → Outbox
```

Để phẳng thì `app/` sẽ phình thành một package 50 file.

#### Quy ước 9 — Zero value là "chưa set", KHÔNG phải giá trị hợp lệ

| | |
|---|---|
| **Luật** | mỗi VO có `IsZero()`/`IsValid()`; constructor **từ chối** xây trên zero value |
| **Symfony** | ~ cột nullable chưa gán. PHP có `?Type`, Go **không cấm được** `Money{}` |

```go
// Go LUÔN cho phép viết cái này — không cấm được
m := shared.Money{}          // 0, currency rỗng

// Nên constructor phải tự vệ
shared.NewMoney(5, shared.Currency{})   // → panic
```

**Vì sao gắt:** hai giá trị **cùng vô hiệu** thì "khớp" nhau — `Money{}.Add(Money{})`
từng trả `"0 "` không lỗi. Guard phải hỏi *"có hợp lệ không"* **trước** khi hỏi
*"có khớp không"*. Bug thật: §9 đợt 2 #14.

#### Quy ước 10 — Constructor > 3 tham số → struct `XxxDetails`, kèm HAI điều kiện

| | Tham số vị trí | Struct tham số |
|---|---|---|
| Thêm field | **lỗi biên dịch** ở mọi chỗ gọi | vẫn compile, field mới **âm thầm** mang zero value |
| Đảo 2 tham số cùng kiểu | compiler **không** bắt | không xảy ra — có tên |
| Đọc call site | phải nhớ thứ tự | tự giải thích |

Compile-time mạnh hơn runtime, nên struct **chỉ** được dùng khi đủ **hai** điều kiện:

1. **Constructor validate mọi field bắt buộc.** Field nào cho phép zero thì zero
   phải là **đáp án nghiệp vụ an toàn** (`FreeShipping{}` = shop luôn tính phí ship).
2. **Có test `Test…_everyFieldIsValidated`** — dùng `reflect` zero **từng field
   một** và đòi constructor từ chối.

> ⚠️ Một test zero **cả struct** là KHÔNG đủ — đã kiểm chứng: thêm field
> `Country` không validate, **91 test vẫn xanh**, vì `MerchantDetails{}` chết ở
> `Name` rỗng rồi return, không bao giờ chạy tới field mới. §9 đợt 5 #24.

`now time.Time` để **ngoài** struct: nó là "lúc nào", không phải chi tiết của
merchant. Symfony có named arguments; Go không.

---
## 7. Số liệu nghiệp vụ đã chốt

> Code trỏ vào mục này ở nhiều chỗ (`wire.go`, `pricing/lane.go`, `policy.go`,
> `quote_test.go`). **Đổi số ở đây là phải đổi cả code và test vàng.**

| Tham số | Giá trị | Nằm ở đâu trong code |
|---|---|---|
| Tỷ giá | 26.000 ₫/USD | `wire.seedPricing` |
| Thuế bán hàng Colorado | 8,81 % | kho Denver |
| Cọc khách trả trước | 50 % giá trị đơn | `ordering.CustomerOrder` |
| Phí dịch vụ | max(10 %, sàn 500.000 ₫) | `wire.quotePolicy()` |
| Quy đổi thể tích | mm³ / 5000, bước cân 500 g | `wire.seedPricing` |
| Hạn báo giá | 48 h | `wire.quotePolicy()` — chụp vào **từng** quote |
| Giá cân | 9/10/12/14 USD/kg + phụ thu pin 3 USD | `wire.seedPricing` |
| Phân loại hàng | footwear, apparel → branded · electronics → electronics · còn lại standard | `wire.goodsClasses()` |

> ⚠️ `goodsClasses()` **thiếu một dòng** → món đó rơi về `standard` → **báo giá
> THẤP**, và chỉ lộ ra ở bước đối soát cuối cùng.

**Số vàng — `scripts/smoke.sh` phải luôn in đúng những con số này:**

```
quote … total 5393720 VND, deposit 2696860
variance -2.50 USD          (quoted 163.22+25.00, actual 163.22+27.50)
```

> ⚠️ Script **in** cả bốn số nhưng chỉ **assert** hai: `deposit=2696860` và
> `variance=-2.50` (`smoke.sh` dòng 129–130). In ra mà không assert thì mắt
> người phải canh — và mắt người bỏ sót. Đổi nghiệp vụ mà `total` lệch thì
> script vẫn **xanh**.

### Số còn thiếu — cần hỏi nhà gửi hàng bên Mỹ

- [ ] Bảng giá theo **kg**, phân theo loại hàng (thường / hàng hiệu / điện tử / nhạy cảm)
- [ ] Bước làm tròn cân — 100 g hay 500 g?
- [ ] Số chia thể tích họ dùng — 5000 hay 6000?
- [ ] Có phụ thu hàng **pin lithium** không (tai nghe, đồng hồ…)?
- [ ] Trần giá trị mỗi kiện?

### Còn nợ trong code

- [ ] `config/ratecard.yaml` — bảng giá đang là **hằng** trong `wire.go`
- [ ] `adapter/merchant/<shop>.go` — ACL **thật** (chưa shop nào cho API)
- [ ] Cổng thanh toán

> Việc **đã xong** không liệt kê ở đây nữa — `git log` và §9 đã kể đủ, và một
> danh sách 30 dòng `[x]` chỉ làm loãng 3 dòng `[ ]` thật sự còn lại.

### Ghi chú rủi ro

Tuyến "gửi ở Mỹ, trả trọn gói, về tận nhà VN" hoạt động tốt ở quy mô **cá nhân**.
Ở quy mô kinh doanh (nhiều thùng/tháng, từ một địa chỉ gửi tới nhiều người nhận),
cách hải quan xử lý có thể khác. Hệ thống nên **đếm và cảnh báo** khi tần suất
vượt ngưỡng tự đặt, thay vì giả vờ rủi ro đó không tồn tại.

**Ràng buộc pháp lý:** điều khoản của Nike và phần lớn shop **cấm cào trang**.
Nên hệ thống **không tự fetch trang shop** — khách/operator nhập tay, hoặc
`adapter/openai` đọc từ URL. Đây là ràng buộc thiết kế, không phải chú thích.

---

## 8. Dựng lại trên máy mới

```bash
# 1. Cài Go (§2) — rồi ĐÓNG HẲN và MỞ LẠI terminal
winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements   # Windows
sudo apt install golang-go                                                                # Linux

# 2. Clone + danh tính cá nhân NGAY, trước commit đầu tiên (§3b)
git clone <repo> portage && cd portage
git config --local user.email "duongsy1920@gmail.com"

# 3. Thư viện + xác nhận chạy được
go mod download        # ~ composer install
go test ./...          # 272 test, < 2 giây, không cần Docker

# 4. Muốn chạy thật thì thêm Postgres
docker compose up -d
bash scripts/smoke.sh
```

---
## 9. Nhật ký review 04/09/2026 — bug đã tìm thấy và bài học

> **Đây là chỗ trả lời câu *"vì sao code lại trông như thế này"*.** 18 đợt, mỗi
> đợt một bảng `quyết định / bug → chỗ nó nằm`. Đọc 2–3 đợt cuối là nắm được
> tình trạng hiện tại. 24 bug được đánh số — câu tự kiểm rút từ chúng ở
> [HOC.md](HOC.md).

| Đợt | Ngày | Chủ đề | Bug đắt nhất của đợt |
|---|---|---|---|
| 1 | 04/09 | review `shared/` | làm tròn **xuống** ở bước trung gian → mất một bậc cân |
| 2 | 04/09 | tổ chức lại source | `Add`/`Sub` tràn `int64` âm thầm, `Mul` thì panic — **không nhất quán** |
| 3 | 04/09 | `catalog` | thuế và số chia nằm nhầm trong `CategoryPolicy` |
| 4 | 04/09 | `Product`/`Variant`/`Provenance` | — |
| 5 | 04/09 | review chéo từ session Windows | test zero **cả struct** không guard được field mới |
| 6 | 04/09 | tầng app + adapter in-memory | — |
| 7 | 04/09 | `adapter/http` + `cmd/api` | — |
| 8 | 04/09 | quyết định **không** làm `cmd/worker` vội | — |
| 9 | 04/09 | Postgres, outbox, worker | test tích hợp song song giẫm chân nhau trên một DB |
| 10 | 05/09 | context 2 — `pricing` | patch theo anchor cũ **hỏng im lặng** sau khi đổi tên |
| 11 | 05/09 | context 3 — `ordering` | hằng trạng thái trùng tên kiểu event (Go một namespace) |
| 12 | 05/09 | context 4 — `procurement`, ACL #1 | script vá dừng giữa chừng mà vẫn đẩy file cũ |
| 13 | 06/09 | context 5 — `logistics`, đối soát | — |
| 14 | 06/09 | P9/T1 auth thật | `pgtest` truncate theo danh sách bảng **viết tay** |
| 15 | 06/09 | P9/T3 quét quote hết hạn | smoke xanh chỉ vì DB tình cờ rỗng; lần 2 **401 sạch** |
| 16 | 06/09 | P9 xong — read model, AI, tỷ giá | — |
| 17 | 08/09 | variant/size qua ba context | `POST /orders` nhận **bất kỳ** uuid làm `variant_id` |
| 18 | 08/09 | clone sang Linux, `smoke.sh` chạy thật | `CREATE TABLE IF NOT EXISTS` **không nguyên tử** — chỉ vỡ khi DB cold |


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

| **Lỗ hổng test, tìm ra khi viết bài tập cho `HOC.md`:** hành vi "thiếu dòng variant thì task **vẫn** mở" hoàn toàn không có test. Đổi nhánh `ErrVariantNotFound` thành `return err` mà cả suite vẫn xanh — nghĩa là một quyết định có chủ đích đang không được canh. Thêm `TestOpenTask_opensEvenWhenTheVariantRowHasNotArrived`, và tách `seedShopAndItem` khỏi `seedCatalog` để dựng được đúng cảnh sai thứ tự | `procurement_test.go` |
| Bài học: viết tài liệu hướng dẫn *"sửa dòng này thì test đỏ"* là một cách kiểm thử độ phủ. Nếu phá mà không đỏ thì hoặc code không quan trọng, hoặc test đang thiếu | `HOC.md` buổi 9 |

Tổng **293 test** (+2 so với đợt 17): `TestMigrate_survivesTwoProcessesOnAColdDatabase`
và `TestOpenTask_opensEvenWhenTheVariantRowHasNotArrived`.
Với `PORTAGE_TEST_DSN`: **292 PASS + 1 SKIP**; không có DSN: 263 PASS + 30 SKIP.
Con số giống nhau trên cả Windows và Linux.

**Nợ đã trả:** `scripts/smoke.sh` từ đợt 16 tới giờ mới chỉ syntax-check trên
Windows. Giờ nó đã chạy thật, trên Linux, trên database trống, và ngay lần đầu
đã bắt được một bug mà 292 test không bắt.

### Đợt 19 (08/09) — hai màn hình theo vai, và một luật đã bị bỏ quên

Anh bấm thử bản console đầu rồi nói thẳng: *"UI gì mà khó xài dữ vậy ta?"*, kèm danh sách
rất cụ thể. Đợt này là câu trả lời, và nó sinh ra bảng đọc thứ hai cùng ba trường mới.

| Quyết định / bug | Chỗ nó nằm |
|---|---|
| Lỗi gốc: bản console bày ra **hình dạng của API** — một nút cho mỗi endpoint, bốn nút sáng cùng lúc, một ô dán uuid, và `1250` hardcode trong nhãn nút. Người dùng không nghĩ "thêm variant"; họ nghĩ *"làm cho cái link này bán được"* | `WALKTHROUGH.md` §23a |
| Nhân viên **không hề biết** khách muốn size nào — hệ thống chưa bao giờ ghi. Thêm `RequestedVariant` (chữ của khách) cạnh `RequestedBy` (ai đang đợi). Cả hai nói về **yêu cầu**, nên cố tình không nằm trong `Provenance` | `catalog/product.go`, `0009`, `0011` |
| Size phải là **chữ tự do**: giày người lớn có hệ M/W, giày trẻ có Y/C, áo có S/M/L, điện thoại có dung lượng. Enum nào cũng sẽ từ chối một size thật. UI **dạy** hệ size thay vì ép danh sách | `web/app/{customer,staff}.js` |
| `catalog.listing_confirmed` — event thứ 37. Không event nào khác cho biết bước xác nhận đã xong, nên bảng đọc không thể biết, và một hàng chờ bắt người ta làm lại việc đã làm là hàng chờ họ thôi tin | `catalog/events.go` |
| Bảng đọc `product_worklist`: bốn event, một dòng, `NextStep()` quyết **đúng một lần** trong model. Hai màn hình tự tính "còn thiếu gì" sẽ lệch nhau vào ngày có bước thứ năm, và lệch im lặng | `reporting/worklist.go` |
| `GET /categories` và `GET /merchants` — hai tập đóng đã có ở domain nhưng **không có route đọc**, nên form buộc phải hỏi uuid | `http/reference.go` |
| **Bug thật:** adapter in-memory của `CategoryRepo.All()` duyệt map Go, mà Go cố tình ngẫu nhiên hoá thứ tự duyệt map; Postgres thì `ORDER BY code`. Hai adapter không đồng ý → ngành hàng mặc định của form đổi mỗi lần tải trang. Có test canh cả hai giờ | `memory/category_repo.go` |
| **Lỗ hổng nghiệp vụ, tìm ra khi đang viết docs:** `Merchant.Supports()` tồn tại từ đầu và **chưa ai gọi**. `sourcing` được ghi, được thông báo bằng event riêng, rồi bỏ quên — khách dán được link cho shop chỉ dành nhân viên. Giờ kiểm trong `AddProduct`, 409 `sourcing_not_allowed` (không phải 403: chìa hợp lệ, shop không nhận nguồn đó) | `app/catalog/add_product.go` |
| Hai fixture test phải mở rộng vì chúng **đang chạy qua đường mà luật cấm**, suốt thời gian đó không ai biết, vì không có gì kiểm | `server_test.go`, `register_merchant_test.go` |
| **Bug của tôi:** sửa migration 0009 tại chỗ với lý do "nó chỉ chạy trên database tạm", quên là máy Windows cũng đã chạy → 31 test đỏ ở đó với `column requested_by already exists`. Sửa một migration là phải dựng lại DB ở **mọi** máy đã chạy nó | `docker compose down -v` |
| Mục **§3b** của file này từng bị bỏ khi file được viết lại gọn hơn. Đã lấy lại nguyên văn: nó giữ cho commit cá nhân không mang email công ty | `SETUP.md` §3b |

### Năm lỗi chỉ trình duyệt thật mới bắt được

`go test` xanh, `curl` đúng, và cả năm vẫn còn nguyên. Chỉ lộ ra khi cho Chrome bấm hết
luồng bằng Playwright:

```
style="margin-left:auto"        React đòi object, không phải chuỗi        → error #62
Field({...}) gọi như hàm        useId nhập vào hook list của cha          → error #310
<label> không nối for/id        hỏng trình đọc màn hình VÀ getByLabel
dòng tiền 0.00 vẫn hiện         so chuỗi với "0" mà API trả "0.00"
ngành hàng mặc định đổi         map của Go, xem dòng bảng trên
```

Bài học: **một UI không có test là một UI chưa ai chạy.** Với backend thì `go test ./...` là
bằng chứng; với màn hình, bằng chứng duy nhất là có cái gì đó bấm nó.

Tổng **305 test** (+14 so với đợt 18). Với `PORTAGE_TEST_DSN`: **304 PASS + 1 SKIP**; không
DSN: 272 PASS + 32 SKIP. 41 route, 35 dòng `Subscribe`, 37 domain event, 11 migration. Số
giống nhau trên cả Windows và Linux.

Đọc chi tiết: `WALKTHROUGH.md` §23 (chín mục) và `UI-GUIDE.md` (bốn màn hình).
