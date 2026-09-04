# Portage — Sổ tay dựng môi trường & khởi tạo source

> File ghi chép cá nhân. Ghi lại **đúng những gì đã làm**, để sau này dựng lại
> trên máy khác không phải mò lại từ đầu.
>
> Cập nhật: 2026-09-03
>
> 📚 **Học DDD:** xem [DDD.md](DDD.md) — toàn bộ khái niệm, kiến trúc, bẫy
> thường gặp, giải thích bằng chính code của project này.

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
| Docker | 29.7.2 | sẽ dùng chạy Postgres |
| winget | 1.29.290 | trình cài đặt của Windows |
| Node.js | có sẵn ở `D:\NodeJs` | **không thuộc stack**; chỉ dùng chạy script tính giá nháp |

> ⚠️ **Node KHÔNG cần cho project này.** Stack chính là Go + Postgres.
> Node chỉ tình cờ được dùng để chạy mấy file tính chi phí lúc bàn nghiệp vụ.

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
├── cmd/                        # nơi để hàm main() — mỗi thư mục 1 chương trình
│   ├── api/                    #   web server
│   └── worker/                 #   chạy job nền (gọi OpenAI, đóng lô hàng…)
├── internal/                   # Go CẤM module khác import thư mục tên `internal`
│   ├── domain/                 # TRÁI TIM — nghiệp vụ thuần, KHÔNG import gì bên ngoài
│   │   ├── shared/             #   Shared Kernel: Money, Weight, Rate
│   │   ├── catalog/            #   Merchant, Category, CategoryPolicy
│   │   ├── pricing/            #   Quote, RateCard, ShippingLane
│   │   ├── ordering/           #   CustomerOrder   (aggregate chính)
│   │   ├── procurement/        #   PurchaseTask    (việc đi mua ở web Mỹ)
│   │   └── logistics/          #   Parcel, ConsolidationBatch
│   ├── app/                    # Use case / điều phối — gọi domain, gọi repository
│   ├── adapter/                # THẾ GIỚI BÊN NGOÀI cắm vào đây
│   │   ├── postgres/           #   cài đặt repository bằng Postgres
│   │   ├── http/               #   handler, router
│   │   ├── openai/             #   gọi OpenAI
│   │   └── merchant/           #   ACL cho từng web bán hàng (Nike, TNF…)
│   └── platform/               # config, logger, connection pool
├── config/                     # bảng giá, thuế suất — DỮ LIỆU, không phải code
└── docs/                       # file này
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

Lỡ import → test domain sẽ cần Docker, cần API key, chạy chậm.
Làm đúng → test domain chạy trong **vài mili-giây**, không cần gì cả.

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
| Xem cấu hình | `go env` | `php -i` |

> `./...` nghĩa là "thư mục hiện tại và **mọi** thư mục con". Nhớ dấu ba chấm.

---

## 6. Đã viết được gì (tính tới 03/09/2026)

```
internal/domain/shared/
├── money.go        Money, Currency — cộng/trừ/nhân tỷ lệ/đổi tiền
├── rate.go         Rate (tỷ lệ %), ExchangeRate (tỷ giá)
├── weight.go       Weight, Dimensions, trọng lượng quy đổi thể tích
├── money_test.go
└── weight_test.go
```

Chạy kiểm tra:

```bash
cd /d/portage
go test ./...
# ok  github.com/duongsy/portage/internal/domain/shared
```

### Ba nguyên tắc đã cài cứng vào code

1. **Tiền không bao giờ dùng float.** Lưu bằng `int64` theo đơn vị nhỏ nhất
   (USD → cent, VND → đồng). Vì `0.1 + 0.2 != 0.3` trong dấu chấm động.
2. **USD và VND không cộng được với nhau.** `Add()` trả lỗi `ErrCurrencyMismatch`.
   VND có **0 chữ số thập phân**, USD có 2 — số thập phân đi kèm currency,
   không hard-code trong `Money`.
3. **Trọng lượng tính phí = làm-tròn-lên( max(cân thật, quy đổi thể tích) ).**
   Hãng bay tính tiền theo **chỗ chiếm**, không theo cân nặng.
   Áo phao 0,9 kg bị tính 4,7 kg. Báo giá theo cân thật là lỗ chắc.

---

## 7. Việc tiếp theo

- [ ] `git init` + commit đầu tiên
- [ ] `internal/domain/catalog` — Merchant, Category, CategoryPolicy
- [ ] `internal/domain/pricing` — RateCard, ShippingLane, Quote
- [ ] `config/ratecard.yaml` — điền bảng giá **thật** xin từ nhà gửi hàng bên Mỹ
- [ ] `internal/domain/ordering` — aggregate `CustomerOrder` + luật cọc 50%
- [ ] `docker-compose.yml` cho Postgres
- [ ] Tầng adapter: postgres, http

### Số liệu nghiệp vụ đã chốt

| Tham số | Giá trị | Ghi chú |
|---|---|---|
| Tỷ giá | 26.000 ₫/USD | tỷ giá ngân hàng |
| Cọc khách trả trước | 50% giá trị đơn | |
| Lãi mục tiêu | 500.000 ₫/đơn | mức sàn tối thiểu |
| Kho Mỹ | Denver, Colorado | sales tax ~8,81% |
| Tuyến vận chuyển chính | dịch vụ gửi hàng ở Mỹ, trọn gói tận nhà VN | phí trả hết đầu Mỹ |
| Chia thể tích | 5000 (cm³/kg) | cần xác nhận lại |

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
