# Portage

Hệ thống mua hộ xuyên biên giới, Mỹ → Việt Nam. Khách dán link một sản phẩm ở web Mỹ
(Nike, The North Face, Sony…), mình mua hộ, gom lô tại kho Denver, Colorado, rồi ship về
tận nhà khách ở Việt Nam.

Viết bằng **Go**, kiến trúc **Domain-Driven Design (DDD)**. Vừa là sản phẩm thật, vừa là
project để học Go và DDD từ nền PHP/Symfony.

---

## Một đơn hàng đi qua đâu

```
  khách dán link          nhân viên kiểm, cân, đăng bán          khách xin giá, đồng ý
        │                              │                                  │
        ▼                              ▼                                  ▼
   ┌──────────┐  product_published  ┌──────────┐   quote_accepted   ┌──────────┐
   │ CATALOG  │ ──────────────────▶ │ PRICING  │ ─────────────────▶ │ ORDERING │
   └──────────┘                     └──────────┘                    └────┬─────┘
                                          ▲                              │ deposit_paid (cọc 50 %)
                                          │                              ▼
                                          │  purchase_confirmed    ┌─────────────┐
                                          │  (giá THẬT đã trả)     │ PROCUREMENT │  nhân viên đi mua
                                          │◀───────────────────────└──────┬──────┘
                                          │                               │ purchase_confirmed
                                          │  batch_shipped                ▼
                                          │  (cước THẬT)            ┌───────────┐
                                          └─────────────────────────│ LOGISTICS │  kho cân thật, gom lô, ship
                                                                    └───────────┘
   …và mọi event ở trên ──▶ REPORTING: hai bảng phẳng cho màn hình khách và nhân viên
```

Năm vùng nghiệp vụ (bounded context) không gọi thẳng nhau. Mỗi vùng ghi lại việc mình vừa
làm dưới dạng một **event** vào bảng `outbox`, cùng transaction với dữ liệu. Một tiến trình
riêng (**worker**) đọc bảng đó và báo cho các vùng đang lắng nghe. Cuối vòng đời, `pricing`
so **giá đã báo** với **giá thật đã trả** và ghi ra `variance`: đơn này lời hay lỗ.

---

## Chạy trong ba lệnh

```bash
go test ./...                    # 278 test, dưới 2 giây, không cần Docker
go run ./cmd/api -web ./web      # API + giao diện, dữ liệu trong RAM, mất khi tắt
# mở http://localhost:8080/ui/  — chìa khoá dev: dev-operator (nhân viên) / dev-customer (khách)
```

Muốn dữ liệu còn sau khi tắt thì thêm Postgres:

```bash
docker compose up -d
DSN="postgres://portage:portage@localhost:5432/portage?sslmode=disable"
go run ./cmd/api    -dsn "$DSN" -web ./web -bootstrap-operator-token "$(openssl rand -hex 32)"
go run ./cmd/worker -dsn "$DSN"           # relay outbox nằm ở đây; thiếu nó thì màn hình trống
bash scripts/smoke.sh                     # cả vòng đời một đơn trên binary thật, assert số vàng
```

Cách bấm từng bước trên giao diện: [docs/UI-GUIDE.md](docs/UI-GUIDE.md).

---

## Đọc gì, theo thứ tự nào

| Bạn muốn | Mở file |
|---|---|
| **Học Go và DDD từ đầu bằng repo này** | [docs/HOC.md](docs/HOC.md) — lộ trình 10 buổi, bắt đầu bằng 12 từ và một hình |
| Bấm thử trước khi đọc dòng code nào | [docs/UI-GUIDE.md](docs/UI-GUIDE.md) |
| Hiểu một đơn hàng từ cái link tới đồng bạc cuối | [docs/FLOW-ORDER.md](docs/FLOW-ORDER.md) |
| Tra một khái niệm DDD, đối chiếu Symfony | [docs/DDD.md](docs/DDD.md) |
| Đọc code theo thứ tự, từng file | [docs/WALKTHROUGH.md](docs/WALKTHROUGH.md) |
| Tra cú pháp Go khi đang quen PHP | [docs/GO-CHO-PHP.md](docs/GO-CHO-PHP.md) |
| Dựng môi trường, lệnh hàng ngày, 10 quy ước code | [docs/SETUP.md](docs/SETUP.md) |
| Vì sao code trông như vậy: nhật ký 20 đợt review | [docs/SETUP.md §9](docs/SETUP.md) |
| Vì sao không dùng affiliate feed, provenance là gì | [docs/CATALOG.md](docs/CATALOG.md) |
| Hai plan đã xong, giữ để thấy cách chốt quyết định | [docs/P9-PLAN.md](docs/P9-PLAN.md) · [docs/P10-PLAN.md](docs/P10-PLAN.md) |

---

## Số liệu

```
5 bounded context + 1 tầng đọc · 7 aggregate root · 37 domain event
41 route HTTP · 35 dòng định tuyến event (wire.Subscribe) · 13 migration
312 test (311 PASS + 1 SKIP cố ý) · 278 test chạy < 2 giây không cần Docker
7 test canh kiến trúc bằng go/ast · 2 dependency ngoài (uuid, pgx)
```

| Context | Aggregate root | Trả lời câu gì |
|---|---|---|
| `catalog` | Merchant, Product ⊃ Variant | bán món gì, của shop nào, có size nào |
| `pricing` | Quote, Reconciliation | giá bao nhiêu, và cuối cùng lời hay lỗ |
| `ordering` | CustomerOrder | khách cam kết gì, đã trả gì, huỷ thì lấy lại gì |
| `procurement` | PurchaseTask | ai đi mua, mua xong chưa, trả thật bao nhiêu |
| `logistics` | Parcel, ConsolidationBatch | thùng cân bao nhiêu, đi lô nào, cước chia sao |
| `reporting` | *(không có domain)* | hai bảng phẳng cho màn hình, dựng chỉ bằng event |

**Số nghiệp vụ đã chốt:** thuế bán hàng Colorado 8,81 % · tỷ giá 26 000 ₫/USD · cọc 50 % ·
phí dịch vụ max(10 %, sàn 500 000 ₫) · quy đổi thể tích mm³ / 5000, bước cân 500 g · hạn báo
giá 48 giờ.

**Số vàng** mà `scripts/smoke.sh` phải in ra sau mỗi lần sửa:

```
quote … total 5393720 VND, deposit 2696860
variance -2.50 USD          (quoted 163.22+25.00, actual 163.22+27.50)
```

---

## API — 41 route

Mọi route đứng sau `Authorization: Bearer <token>`. Không token → **401**. Đúng token sai
loại → **403**. Khách hỏi đơn của người khác → **404**, không phải 403, để không xác nhận đơn
đó có thật.

| Route | Ai gọi | Ghi chú |
|---|---|---|
| `POST /tokens` · `GET /tokens` · `DELETE /tokens/{hash}` | nhân viên | cắt chìa (token trả về **một lần**), xem, thu hồi; chìa nhân viên cuối → 409 |
| `POST /categories` · `GET /categories` | nhân viên · cả hai | ngành hàng: hộp mẫu để ước lượng cước |
| `POST /merchants` · `GET /merchants` | nhân viên · cả hai | shop: tiền tệ, nguồn nào được gửi link |
| `POST /lanes` · `POST /fx` | nhân viên | tuyến vận chuyển và tỷ giá; `/fx` không phát event |
| `POST /products` | cả hai | nguồn dữ liệu suy từ token: nhân viên → đã kiểm, khách → nháp |
| `POST /products/from-url` | cả hai | AI đọc trang → nháp; không có API key → **503** |
| `POST /products/{id}/{variants,confirm-listing,measure,publish}` | nhân viên | bốn bước để một món bán được |
| `GET /product-queue` · `GET /me/products` | nhân viên · khách | việc còn phải làm · món tôi đã gửi |
| `POST /quotes` · `GET /quotes/{id}` · `POST /quotes/{id}/accept` | cả hai | báo giá là **ảnh chụp**, hạn 48 giờ |
| `POST /orders` | cả hai | khách đặt cho mình; nhân viên đặt hộ phải kèm `customer_id` (400 nếu thiếu, 403 nếu khách gửi kèm) |
| `GET /me/orders` · `GET /orders?status=` | khách · nhân viên | đơn của tôi · hàng đợi việc; status sai → 400 |
| `GET /orders/{id}` · `POST /orders/{id}/cancel` | chủ đơn hoặc nhân viên | người lạ → 404 |
| `POST /orders/{id}/{deposit,balance,deliver}` | nhân viên | xác nhận tay, tới khi có cổng thanh toán |
| `GET /purchase-tasks` · `GET …/{id}` · `POST …/{id}/{confirm,fail}` | nhân viên | việc đi mua; không có `POST /purchase-tasks`: việc chỉ mở khi có cọc |
| `GET /parcels` · `GET …/{id}` · `POST …/{id}/receive` | nhân viên | kho: cân **thật** |
| `POST /batches` · `GET …/{id}` · `POST …/{id}/{parcels,close,ship}` | nhân viên | gom lô, chia cước đúng từng cent |
| `GET /reconciliations/{order}` | nhân viên | giá đã báo so với giá thật → `variance` |

Mỗi route ở trên làm gì, ai nghe event nào, hỏng thì mã lỗi gì:
[docs/FLOW-ORDER.md](docs/FLOW-ORDER.md).

---

## Chưa có

- `config/ratecard.yaml` — bảng giá cước còn là hằng số trong `internal/platform/wire/wire.go`.
- Một adapter shop thật (`adapter/merchant/<shop>.go`) — chưa shop nào cho API; hôm nay một
  người đi mua rồi báo lại.
- Cổng thanh toán — cọc và phần còn lại đang do nhân viên xác nhận tay.

## Ràng buộc pháp lý

Điều khoản của Nike và phần lớn shop **cấm cào trang** và cấm mua để bán lại qua affiliate
feed. Hệ thống vì thế **không tự tải trang shop**: khách hoặc nhân viên nhập tay, hoặc
`adapter/openai` đọc từ URL người dùng đưa. Đây là ràng buộc thiết kế, không phải chú thích.
Chi tiết: [docs/CATALOG.md](docs/CATALOG.md) §2.
