# DDD — Sổ tay học, bám theo project Portage

> Tài liệu học cá nhân. Mỗi khái niệm DDD được giải thích bằng **chính code
> trong project này**, và đối chiếu với cách PHP/Symfony/Doctrine làm.
>
> Cập nhật: 2026-09-04 · Xem thêm [SETUP.md](SETUP.md) (mục 6 có bảng **quy ước code đã chốt**, mục 9 có **nhật ký review**)

## Cách đọc file này

Mỗi khái niệm có 4 phần cố định:

| Ký hiệu | Nghĩa |
|---|---|
| **Là gì** | định nghĩa ngắn, không vòng vo |
| **Vì sao cần** | vấn đề thật nó giải quyết |
| **Trong Portage** | code thật, đường dẫn file thật |
| **So với Symfony** | ánh xạ sang thứ bạn đã biết |

Trạng thái ở đầu mỗi mục:

- ✅ **đã code** — có file, có test chạy được
- 🔜 **sắp làm** — đã thiết kế, chưa viết
- 📋 **mới bàn** — mới thống nhất trên lý thuyết

> ⚠️ **Quan trọng cho CV/phỏng vấn:** chỉ nói về phần ✅. Phần 🔜 và 📋 khi được
> hỏi thì trả lời *"đây là hướng thiết kế tiếp theo"*, đừng nói như đã làm rồi.

---

# PHẦN I — DDD là gì

## 1. Định nghĩa

**Là gì.** Domain-Driven Design (Eric Evans, 2003) là cách xây phần mềm mà
**nghiệp vụ là trung tâm**, không phải database hay framework.

DDD **không phải** là:

- một design pattern đơn lẻ
- một thư viện, một framework
- một cấu trúc thư mục
- bắt buộc phải có microservices

DDD **là** một tập hợp gồm hai nhóm:

```
   ┌─────────────────────────────────────────────────────┐
   │  STRATEGIC DESIGN  —  thiết kế chiến lược           │
   │  Chia hệ thống lớn thành vùng. Phần KHÓ và QUAN     │
   │  TRỌNG nhất. Sai ở đây thì tactical giỏi mấy cũng   │
   │  không cứu được.                                     │
   │  → Ubiquitous Language, Bounded Context, Context Map │
   └─────────────────────────────────────────────────────┘
                            │
   ┌─────────────────────────────────────────────────────┐
   │  TACTICAL DESIGN  —  thiết kế chiến thuật            │
   │  Cách viết code BÊN TRONG một vùng. Phần dễ học,     │
   │  dễ tìm ví dụ, và cũng dễ bị lạm dụng.               │
   │  → Value Object, Entity, Aggregate, Repository...    │
   └─────────────────────────────────────────────────────┘
```

Đa số người "học DDD" chỉ học nhóm dưới, rồi nhét mọi thứ vào một context
khổng lồ. Đó là lý do họ thấy DDD rườm rà mà không được gì.

## 2. Khi nào KHÔNG nên dùng DDD

Nói trước để bạn không lạm dụng:

| Tình huống | Nên dùng gì |
|---|---|
| CRUD thuần — thêm/sửa/xoá, không có quy tắc | Doctrine + Controller là đủ |
| Script chạy một lần | Viết thẳng |
| Prototype vứt đi sau 2 tuần | Đừng phí công |
| Nghiệp vụ đơn giản, đội 1 người, không đổi | Overkill |

DDD **đáng tiền** khi: nghiệp vụ phức tạp, quy tắc nhiều và hay đổi, nhiều người
cùng làm, hệ thống sống lâu.

**Portage đáng dùng DDD vì:** đơn hàng đi qua 5 giai đoạn ở 2 quốc gia, có tiền,
có thuế, có bên thứ ba không kiểm soát được (web bán hàng Mỹ, hãng vận chuyển,
hải quan), và luật hoàn tiền phụ thuộc vào việc *đã mua hàng hay chưa*.

---

# PHẦN II — Strategic Design (chiến lược)

## 3. Ubiquitous Language — Ngôn ngữ chung ✅

**Là gì.** Một bộ từ vựng **duy nhất**, dùng y hệt nhau ở mọi nơi: lúc nói
chuyện với khách, trong tài liệu, trong tên class, trong tên cột database.

**Vì sao cần.** Vì mỗi lần "dịch" từ ngôn ngữ nghiệp vụ sang ngôn ngữ kỹ thuật
là một lần hiểu sai. Dịch qua dịch lại đủ nhiều là hệ thống không còn phản ánh
nghiệp vụ nữa.

**Trong Portage.** Bạn nói với mình: *"phí cân ký — mỗi seller tự định mức ký"*.
Đó là từ dân trong nghề dùng thật. Nên nó đi thẳng vào code:

```go
// ✅ ĐÚNG — giữ nguyên khái niệm nghiệp vụ
type RateCard struct { ... }          // bảng giá cân ký
func ChargeableWeight(...) Weight     // trọng lượng tính phí

// ❌ SAI — dịch sang "ngôn ngữ lập trình viên", mất nghĩa gốc
type ShippingFeeCalculationServiceImpl struct { ... }
func computeWeightValue(...) float64
```

Bảng từ vựng đã chốt của Portage:

| Tiếng Việt (bạn nói) | Trong code | Nghĩa |
|---|---|---|
| phí cân ký | `RateCard` | bảng giá vận chuyển theo kg |
| trọng lượng tính phí | `ChargeableWeight` | `max(cân thật, quy đổi thể tích)` |
| quy đổi thể tích | `VolumetricWeight` | `D×R×C ÷ 5000` |
| cọc | `Deposit` | 50% khách trả trước |
| gom lô | `ConsolidationBatch` | gom nhiều đơn vào một thùng |
| đơn của khách | `CustomerOrder` | thứ khách đặt trên web mình |
| việc đi mua | `PurchaseTask` | việc mua thật ở web Mỹ |
| nhà bán | `Merchant` | Nike, The North Face, Sony… |

> **Quy tắc:** nếu bạn phải giải thích một tên biến cho người làm nghiệp vụ,
> tên đó sai. Đổi tên, đừng viết comment.

## 4. Knowledge Crunching — Moi kiến thức nghiệp vụ ✅

**Là gì.** Hoạt động lập trình viên ngồi với **domain expert** (chuyên gia
nghiệp vụ) để đào ra quy tắc thật, rồi mô hình hoá lại.

**Trong Portage.** Bạn vừa là lập trình viên vừa là domain expert — bạn đã tự
order hàng Mỹ về VN. Cuộc trao đổi vừa rồi chính là knowledge crunching, và nó
đã sửa **hai** giả định sai của mình:

| Mình giả định | Thực tế bạn cung cấp | Hệ quả lên thiết kế |
|---|---|---|
| Ship lẻ từng đơn | Gom lô, chờ 4 tuần | Sinh ra aggregate `ConsolidationBatch` |
| Đóng thuế NK + VAT ở VN | Trả trọn gói đầu Mỹ, về tận nhà | `ShippingLane` thành khái niệm bậc nhất, mỗi tuyến một cách tính thuế |

> **Bài học:** mô hình sai không phải do code dở, mà do **thiếu kiến thức nghiệp
> vụ**. Không có cách nào "code cẩn thận hơn" để tránh. Chỉ có cách hỏi.

## 5. Domain, Subdomain 📋

**Là gì.** *Domain* là toàn bộ lĩnh vực kinh doanh. Chia nhỏ thành *subdomain*,
và **không phải subdomain nào cũng đáng đầu tư như nhau**:

| Loại | Nghĩa | Nên làm gì | Trong Portage |
|---|---|---|---|
| **Core** | lợi thế cạnh tranh, lý do khách chọn bạn | tự làm, làm kỹ, DDD đầy đủ | Pricing, Procurement, Consolidation |
| **Supporting** | cần có, nhưng không tạo khác biệt | tự làm, làm vừa đủ | Catalog, Customer |
| **Generic** | ai cũng cần, ai cũng giống nhau | **mua/dùng sẵn**, đừng tự viết | Auth, gửi email, thanh toán |

**Vì sao quan trọng.** Đây là bản đồ để phân bổ công sức. Tự viết hệ thống
authentication trong khi thuật toán chia cước còn chưa xong = đầu tư sai chỗ.

## 6. Bounded Context — Ngữ cảnh giới hạn ✅ (6/6 context — vòng đời §31 chạy hết bằng event)

**Là gì.** Một **ranh giới** mà bên trong đó, mỗi từ có đúng **một** nghĩa.

Đây là khái niệm **quan trọng nhất của cả DDD**.

**Vì sao cần.** Vì cùng một từ, hai bộ phận hiểu khác nhau — và ép chúng dùng
chung một class là nguồn gốc của "God Object" 40 cột.

**Trong Portage.** Từ "sản phẩm" mang nghĩa khác nhau ở mỗi vùng:

| Context | "Product" ở đây nghĩa là gì |
|---|---|
| **Catalog** | Air Max 90 — tên, ảnh, mô tả, các màu |
| **Pricing** | một món có giá + loại hàng + kích thước hộp |
| **Procurement** | một dòng cần đi mua: link, size US 9, màu đen |
| **Logistics** | một vật thể có cân nặng và thể tích |

```
   ❌ CÁCH SAI — một Entity cho tất cả

        ┌──────────────────────────────┐
        │  Product                     │
        │  id, name, image, desc,      │
        │  price, taxRate, hsCode,     │
        │  weight, length, width,      │
        │  merchantUrl, sizeUS, color, │
        │  stockStatus, batchId, ...   │  ← 40 cột, ai cũng sửa
        └──────────────────────────────┘

   ✅ CÁCH ĐÚNG — mỗi context có model riêng, nối bằng ID

    CATALOG              PRICING            LOGISTICS
    ┌─────────┐         ┌─────────┐        ┌─────────┐
    │ Product │         │ Item    │        │ Parcel  │
    │ name    │◀─ id ──▶│ price   │◀─ id ─▶│ weight  │
    │ images  │         │ category│        │ dims    │
    └─────────┘         └─────────┘        └─────────┘
```

Bounded context của Portage:

```
internal/domain/
├── shared/         ✅  Shared Kernel — Money, Weight, Rate, ID, Events, ParcelSpec, OperatorID
├── catalog/        ✅  Merchant, Product ⊃ Variant, CategoryPolicy, Provenance — 16 event
├── pricing/        ✅  ShippingLane, RateCard, DutyPolicy, QuotePolicy, Calculate, Quote — nghe catalog qua event
├── ordering/       ✅  CustomerOrder (cọc 50 %, Cancel/Refund = §26), AcceptedQuote, Variant — nghe pricing + catalog qua event
├── procurement/    ✅  PurchaseTask, PurchaseReceipt, PORT MerchantACL (ACL §22) — nghe ordering, nghe catalog; ordering nghe lại
├── logistics/      ✅  Parcel (cân thật), ConsolidationBatch, FreightAllocator (Domain Service §17), LaneRule — nghe procurement + pricing; ordering + pricing nghe lại
```

**Ranh giới này đã có test canh**, không chỉ là quy ước trên giấy:
`TestDecision_boundedContextsDoNotImportEachOther` đỏ ngay khi `catalog` import
`pricing`. Chỉ `shared` được dùng chung.

**Và ranh giới đó nhìn thấy được ở ba chỗ (05/09):**

1. `pricing.Listing.Product` là `shared.ID`, không phải `catalog.ProductID` — pricing
   **không gọi được tên** kiểu của catalog; id là danh tính mờ đi qua dây.
2. `listings.product` trong `0002_pricing.sql` là `uuid` trần, **không `REFERENCES`**
   sang `products` — ranh giới vẽ trong SQL. FK chéo context là cách phá ranh giới nhanh nhất.
3. `GoodsClass` (thường/hàng hiệu/điện tử/nhạy cảm) là **từ vựng của nhà gửi hàng**, chỉ
   pricing biết; catalog nói `footwear`, pricing tự dịch sang `branded` bằng `Classification`.
   Cùng đôi giày, hai context, hai từ — cả hai đều đúng.

> **Mẹo nhận ra ranh giới context:** nếu hai bộ phận cãi nhau về định nghĩa một
> từ, và **cả hai đều đúng** — đó chính là ranh giới. Đừng bắt họ thống nhất,
> hãy tách context.

## 7. Context Map — Bản đồ quan hệ ✅ (cạnh CATALOG → PRICING chạy thật)

**Là gì.** Sơ đồ mô tả các context **nói chuyện với nhau thế nào** và **ai phụ
thuộc ai**.

Các kiểu quan hệ thường gặp:

| Kiểu | Nghĩa |
|---|---|
| **Shared Kernel** | hai context dùng chung một phần code nhỏ (đổi phải bàn với nhau) |
| **Customer/Supplier** | context dưới phụ thuộc context trên, nhưng có tiếng nói |
| **Conformist** | phải theo model của bên kia, không cãi được |
| **Anti-Corruption Layer** | dựng lớp dịch để model bên kia không rò vào mình |
| **Published Language** | thống nhất một định dạng chung để trao đổi (event schema) |

**Cạnh đầu tiên đã chạy thật — CATALOG → PRICING, kiểu Customer/Supplier qua Published Language:**

```
 catalog.Product.Publish()  ──▶ outbox (jsonb)  ──▶ worker.Relay ──▶ wire.Subscribe ──▶ eventcodec.Decode
        ProductPublished           key viết tay        at-least-once     bảng định tuyến      → contracts.ProductPublishedV1
                                                                                                     │
                                                                              pricingapp.Projector ◀─┘  → listings
```

| Mảnh | File | Vai trong Context Map |
|---|---|---|
| `internal/contracts/catalog_v1.go` | `ProductPublishedV1`, `CategoryDefinedV1`, `MoneyV1`, `ParcelV1` — primitive-only, version trong tên | **Published Language** — package riêng vì không được là domain (guard 7) và không nên là adapter (app import nó) |
| `eventcodec.Encode` / `Decode` | event → JSON viết tay; JSON → V1 | người dịch hai chiều; decoder chỉ viết khi có người nghe |
| `wire.Subscribe(bus, g)` | event nào → consumer nào | bản đồ quan hệ **dưới dạng code** — đọc một hàm là thấy ai nghe ai |
| `pricingapp.Projector` | upsert `Listing`/`CategoryProfile` | **Customer** giữ bản sao theo từ vựng của mình; idempotent vì relay at-least-once |

Nghe **sai thứ tự** (measured tới trước published) không lỗi — bản sao tự khớp khi bản đầy
đủ tới. Nghe **hai lần** không đổi gì. Hai tính chất đó có test (`pricing_test.go`), không phải hy vọng.
Mổ xẻ từng file: WALKTHROUGH.md §14b.

Context map của Portage:

```
   ┌──────────────┐  ProductPublished  ┌──────────────┐
   │   CATALOG ✅ │───────event ✅────▶│   PRICING ✅ │   product_published/measured/repriced/retired, category_defined
   └──────────────┘                    └──────┬───────┘
                                              │ QuoteAccepted ✅ → ordering.AcceptedQuote (projection)
                                              ▼
                                       ┌──────────────┐
                                       │  ORDERING ✅ │  ← aggregate chính
                                       │ CustomerOrder│
                                       └──────┬───────┘
                                              │ DepositPaid ✅ → procurement.PurchaseTask
                                              ▼
                                       ┌──────────────┐
                                       │PROCUREMENT ✅│  1 đơn khách → 1 việc mua (v1)
                                       │ PurchaseTask │  ─ purchase_confirmed/failed ✅ → ORDERING.Reactor
                                       └──────┬───────┘
                     ┌────────────────────────┼────────────────┐
                     ▼                        ▼                ▼
            ┌────────────────┐      ┌────────────────┐  ┌─────────────┐
            │ ACL: Manual ✅ │      │ ACL: API 🔜    │  │ ACL: … 🔜   │  ← Anti-Corruption, cùng port procurement.MerchantACL
            └────────────────┘      └────────────────┘  └─────────────┘
                     │ PurchaseConfirmed / OutOfStock
                     ▼
            ┌──────────────────────┐
            │    LOGISTICS ✅      │  kho Mỹ → cân thật → gom lô → ship: batch_shipped{allocations}
            │ Parcel, Consolidation│  → ORDERING (in_transit) + PRICING.Reconciler (actual freight → Quote vs Actual)
            └──────────────────────┘
```

## 8. Shared Kernel ✅

**Là gì.** Phần code nhỏ mà **nhiều context dùng chung**.

**Vì sao phải giữ nhỏ.** Vì mọi context đều phụ thuộc vào nó — sửa một dòng là
ảnh hưởng tất cả. Shared kernel phình to = quay lại God Object.

**Trong Portage** — `internal/domain/shared/`, chỉ chứa thứ *thật sự* dùng chung:

```
money.go      Money, Currency, CurrencyFromCode
rate.go       Rate (tỷ lệ phần trăm), ExchangeRate (tỷ giá)
weight.go     Weight, Dimensions, ChargeableWeight
parcelspec.go ParcelSpec — cân nặng + hộp; DỜI từ catalog sang đây ngày 05/09 khi pricing cũng cần
id.go         ID — danh tính (UUID v7) mọi entity dùng chung
operator.go   OperatorID — "ai làm" — mọi context đều hỏi
event.go      Event, Events — cách mọi aggregate ghi nhận sự kiện
```

`ParcelSpec` là ví dụ về **cách một thứ vào shared kernel**: nó sống trong `catalog` cho tới
khi context thứ hai cần đúng kiểu đó (không phải bản sao mang nghĩa khác). Dời xong, guard 7
vẫn xanh — đó là bằng chứng nó thật sự dùng chung, không phải catalog rò sang pricing.

`id.go` và `event.go` là **hạ tầng tactical** dùng chung — cách tạo danh tính,
cách ghi event — không phải khái niệm nghiệp vụ. Chúng ở shared kernel vì
**mọi** aggregate ở **mọi** context sẽ dùng y hệt nhau.

Không có gì khác. Không `Product`, không `Customer`, không `Order` — những thứ
đó mỗi context tự định nghĩa theo nghĩa riêng của mình.

---

# PHẦN III — Tactical Design (chiến thuật)

## 9. Anemic Domain Model — Mô hình thiếu máu ✅ (đã tránh)

**Là gì.** Class *trông như* object nhưng thực chất chỉ là **cái bao đựng dữ
liệu** với getter/setter. Martin Fowler gọi đây là **anti-pattern**.

**Trông thế nào.** Đây là code Symfony thật của bạn ở project `logifreight` cũ:

```php
class Booking
{
    private ?string $status = null;
    private ?string $origin = null;
    private ?string $destination = null;
    private ?string $weight = null;

    public function setStatus(?string $status): static { ... }
    public function setOrigin(?string $origin): static { ... }
    // ... setter cho TẤT CẢ
}
```

Nó cho phép:

```php
$b = new Booking();
$b->setStatus('bananas');       // ✅ chạy được. "bananas" là trạng thái hợp lệ?
$b->setOrigin('HCM');
$b->setDestination('HCM');      // ✅ chạy được. Ship từ HCM tới HCM?
$em->persist(new Booking());    // ✅ Booking rỗng vẫn lưu được
```

**Vấn đề.** Object không tự bảo vệ được mình. Mọi quy tắc nghiệp vụ phải nằm ở
chỗ khác — Controller, Service — và **rải rác khắp nơi**. Sáu tháng sau không ai
biết luật đầy đủ là gì.

**Đối lập của nó** là **Rich Domain Model**: object chứa cả dữ liệu **và** hành
vi, tự từ chối trạng thái sai.

## 10. Primitive Obsession — Nghiện kiểu nguyên thuỷ ✅ (đã tránh)

**Là gì.** Dùng `string`, `int`, `float` cho những khái niệm nghiệp vụ có quy
tắc riêng.

**Vì sao nguy hiểm.**

```go
// Chuyện gì cũng có thể xảy ra
func Ship(weight float64, price float64) { ... }

Ship(150.00, 1.2)   // ✅ compile được. Nhưng đảo ngược tham số rồi.
                    //    150 kg giá 1.2 đô. Không ai phát hiện.
```

Tàu thăm dò *Mars Climate Orbiter* của NASA nổ tung vì đúng loại lỗi này: một
bên dùng pound-force, một bên dùng newton, cùng là `float`.

**Trong Portage** — mọi khái niệm nghiệp vụ đều có kiểu riêng:

```go
func Ship(w shared.Weight, p shared.Money) { ... }

Ship(price, weight)   // ❌ KHÔNG COMPILE. Compiler chặn ngay.
```

## 11. Value Object — Đối tượng giá trị ✅

**Là gì.** Object **không có danh tính**, chỉ có giá trị. Hai value object bằng
nhau khi mọi thuộc tính bằng nhau. **Bất biến** (immutable).

Tiền 100.000₫ trong ví bạn và 100.000₫ trong ví tôi là **như nhau** — không cần
phân biệt "tờ tiền nào". Đó là value object.

**Ba đặc điểm bắt buộc:**

1. **Bất biến** — không sửa, chỉ tạo cái mới
2. **So sánh theo giá trị** — không theo ID
3. **Tự kiểm tra tính hợp lệ** — không tồn tại được ở trạng thái sai

### Value Object đã code trong Portage

Xem `internal/domain/shared/money.go`:

```go
type Money struct {
    minor    int64      // đơn vị nhỏ nhất: USD→cent, VND→đồng
    currency Currency
}
```

**Ba quyết định thiết kế, mỗi cái giải một vấn đề thật:**

#### (a) Tiền không bao giờ dùng float

```go
minor int64    // KHÔNG PHẢI float64
```

Vì trong dấu chấm động nhị phân, `0.1 + 0.2 != 0.3`. Sai một xu mỗi dòng, nhân
với triệu dòng, là mất tiền thật.

> Ở công ty bạn hẳn đã thấy Doctrine trả `DECIMAL` về dạng **string** — cùng lý
> do. PHP không có kiểu decimal gốc nên nó tránh float bằng cách trả string.

#### (b) Số thập phân đi kèm currency

```go
USD = Currency{code: "USD", exponent: 2}   // 2 số lẻ: cent
VND = Currency{code: "VND", exponent: 0}   // 0 số lẻ: không có đơn vị nhỏ hơn
```

Một `Money` hard-code 2 số lẻ sẽ **âm thầm chấp nhận** `3900000.50 VND` — số
tiền không tồn tại. Test chứng minh:

```go
func TestParseMoney_rejectsDecimalsVND(t *testing.T) {
	if _, err := shared.ParseMoney("3900000.50", shared.VND); err == nil {
		t.Fatal("want error for fractional VND, got nil")
	}
}
```

#### (c) Cộng nhầm tiền tệ là LỖI, không phải chuyện hên xui

```go
func (m Money) Add(o Money) (Money, error) {
    if m.currency != o.currency {
        return Money{}, fmt.Errorf("add %s to %s: %w",
            o.currency, m.currency, ErrCurrencyMismatch)
    }
    return Money{minor: m.minor + o.minor, currency: m.currency}, nil
}
```

Cộng 150 USD với 3.900.000 VND ra một con số vô nghĩa. Domain **từ chối**.

#### (d) Domain không đoán — parse số nghiêm ngặt

Bản đầu của `ParseMoney` "tiện tay" bỏ dấu phẩy để nhận `"1,000"`. Hậu quả
tìm ra hôm 04/09:

```go
shared.ParseMoney("150,50", shared.USD)   // người Việt nghĩ: 150,50 USD
// → 15050.00 USD                          // domain hiểu: mười lăm nghìn
```

Sai **100 lần**, không lỗi, không cảnh báo. Bây giờ chỉ nhận đúng một dạng
`-?digits[.digits]`; `"150,50"`, `"+150"`, `"150."`, `".5"` đều trả
`ErrMalformedAmount`. Chuẩn hoá theo locale (dấu phẩy hay dấu chấm) là việc
của **adapter** — form, API — *trước khi* gọi vào domain.

> **Nguyên tắc:** value object nhận **một** định dạng chuẩn. Càng "thông minh"
> càng nhiều cách hiểu sai. Symfony: giống việc dùng `NumberFormatter` ở
> `FormType`, rồi entity chỉ nhận số đã chuẩn.

#### (e) Zero value của Go là "chưa set", không phải giá trị

Đây là bẫy **riêng của Go** mà PHP không có. Mọi struct đều có zero value và
**không cấm được** người khác viết `shared.Money{}` hay `shared.Currency{}`.
Trước 04/09, `NewMoney(500, shared.Currency{})` tạo ra tiền với currency `""`,
in ra `"500 "`, cộng được với nhau. Cách xử lý:

```go
func (c Currency) IsZero() bool  { return c == Currency{} }
func (m Money) IsValid() bool    { return !m.currency.IsZero() }

func NewMoney(minor int64, c Currency) Money {
	if c.IsZero() {
		panic("shared: NewMoney called with the zero Currency")   // bug của mình
	}
	...
}
```

Coi zero value như **cột nullable chưa gán**: có `IsZero()`/`IsValid()` để hỏi,
và constructor từ chối xây lên trên nó.

### Value Object thứ hai: Weight

`internal/domain/shared/weight.go` — và đây là chỗ nghiệp vụ thật của bạn nằm
trong code:

```go
// ChargeableWeight is the number the invoice is built on.
//
//	chargeable = roundUp( max(actual, volumetric) )
func ChargeableWeight(actual Weight, d Dimensions, divisor int64, step Weight) Weight {
	return MaxWeight(actual, d.VolumetricWeight(divisor)).RoundUpTo(step)
}
```

Ba dòng này là **luật kinh doanh sống còn**: hãng bay tính tiền theo *chỗ chiếm*,
không theo cân nặng. Áo phao 0,9 kg bị tính **4,7 kg**.

```go
func TestChargeableWeight_downJacketBillsOnVolume(t *testing.T) {
	actual := shared.Grams(900)                    // áo phao 0,9 kg
	carton := shared.NewDimensionsCM(40, 32, 18)

	chargeable := shared.ChargeableWeight(actual, carton, 5000, shared.Grams(100))
	// → 4.700 kg
}
```

> **Test là tài liệu nghiệp vụ.** Sáu tháng sau mở file test ra vẫn hiểu tại sao
> lại tính như vậy. Đây chính là thứ DDD gọi là *"code kể được câu chuyện nghiệp
> vụ"*.

**Bug tìm được hôm 04/09 — và vì sao nó là bài học DDD, không chỉ là bug số học:**

```go
// Bản cũ
return Weight{grams: d.VolumeCM3() * 1000 / divisor}   // phép chia số nguyên CẮT XUỐNG
```

Thùng 32×27×11 cm = 9504 cm³ → quy đổi thật là **1900,8 g**. Cắt xuống thành
1900 g — rơi **đúng mốc** 100 g — nên `RoundUpTo(100g)` giữ nguyên `1.900 kg`.
Hãng bay tính **2.000 kg**. Mình báo giá thiếu một bậc cân, lỗ 100 g mỗi thùng
kiểu này, và **không có test nào cũ bắt được** vì các số ví dụ tình cờ không
rơi vào biên.

```go
// Bản mới — tính thẳng từ mm³ (bớt một lần cắt) và làm tròn LÊN
mm3 := d.lengthMM * d.widthMM * d.heightMM
return Weight{grams: (mm3 + divisor - 1) / divisor}
```

> **Bài học:** quy tắc nghiệp vụ *"hãng luôn làm tròn lên"* phải áp vào **mọi
> bước** của phép tính, không chỉ bước cuối. Một bước trung gian dùng phép chia
> mặc định của ngôn ngữ = một chỗ nghiệp vụ bị ngôn ngữ ghi đè âm thầm.
> Test tái hiện: `TestVolumetricWeight_roundsUpNotDown`.

Cùng đợt đó, `Weight` được thêm invariant **không âm**:

```go
func NewWeight(grams int64) (Weight, error)   // input ngoài → trả error
func Grams(g int64) Weight                     // literal trong code → panic nếu âm
```

Hai constructor cho hai loại người gọi — xem mục 13.

### Value Object thứ ba: ExchangeRate — và bài học về thời gian

```go
type ExchangeRate struct {
    from, to Currency
    digits   int64   // tỷ giá bỏ dấu chấm: "26000.5" → 260005
    scale    int32   // bao nhiêu chữ số là phần lẻ: 1
}
```

**Vì sao tỷ giá phải là value object?** Vì báo giá phải **chụp ảnh** tỷ giá lúc
báo. Nếu `Quote` chỉ trỏ tới "tỷ giá hiện tại", thì báo giá hôm qua sẽ **tự
đổi giá hôm nay** sau lưng khách. Cùng nguyên tắc đó áp cho bảng giá cân ký:
hôm nay bạn tăng giá, báo giá hôm qua **không được đổi**.

**Ba quyết định đã sửa hôm 04/09 — mỗi cái là một bài học:**

#### (a) `Source`, `AsOf` đi ra khỏi shared kernel

Bản cũ có `Source string` và `AsOf string` (exported, trong khi cả package đóng
kín field). Vấn đề không phải kiểu dữ liệu (`AsOf` đáng ra là `time.Time`) mà
là **trách nhiệm**: *"lấy ở ngân hàng nào, lúc nào"* là kiến thức của
**Pricing** — chỉ Pricing cần nó để đối soát Quote vs Actual. Logistics,
Ordering không quan tâm. Để nó trong shared kernel là bắt **mọi** context
gánh một mối quan tâm của **một** context.

```
   shared.ExchangeRate       = CÁCH đổi tiền        (mọi context dùng)
   pricing.RateSnapshot 🔜   = ExchangeRate + Source + AsOf   (chỉ Pricing dùng)
```

> **Bài học:** shared kernel càng nhỏ càng tốt (mục 8). Câu hỏi để quyết định
> một field có được vào shared kernel không: *"context nào KHÔNG cần nó?"* Có
> một context trả lời "không cần" → nó không thuộc shared kernel.

#### (b) Chính xác đến 12 số lẻ — vì phải đổi được cả chiều ngược

Bản cũ tái dùng `ParsePercent` (4 số lẻ) để đọc tỷ giá. `26000` thì được. Nhưng
hoàn tiền là **VND → USD**: `1/26000 = 0.0000384615…` cần 10 số lẻ → bản cũ
**không biểu diễn được**, và `"0.0000"` thì được nhận → mọi số tiền đổi ra
`0.00 USD` **không báo lỗi**.

Bây giờ `NewExchangeRate` dùng `parseDecimal` với 12 số lẻ, từ chối `≤ 0` và
`from == to` (`ErrInvalidRate`). `Money.Convert` làm một phép nhân và một phép
chia làm tròn qua `big.Int`, không float.

> **Bài học:** *"cùng là số"* không có nghĩa *"cùng một khái niệm"*. `Rate` (thuế
> suất, 4 số lẻ là thừa) và `ExchangeRate` (tỷ giá, cần nhiều hơn) là hai value
> object khác nhau với ràng buộc khác nhau. Tái dùng kiểu để tiết kiệm vài dòng
> = đem ràng buộc của khái niệm này áp lên khái niệm kia.

#### (c) Chuẩn hoá để `==` đúng nghĩa value object

Value object **so sánh theo giá trị**. Nhưng nếu lưu `"26000.50"` là
`digits=2600050, scale=2` và `"26000.5"` là `digits=260005, scale=1`, thì hai
tỷ giá **bằng nhau về giá trị** lại `!=` trong Go. `NewExchangeRate` cắt số 0
thừa ở cuối để một giá trị chỉ có **một** dạng biểu diễn:

```go
a := shared.MustExchangeRate(shared.USD, shared.VND, "26000.50")
b := shared.MustExchangeRate(shared.USD, shared.VND, "26000.5")
a == b   // true — và a.Rate() == b.Rate() == "26000.5"
```

`Rate()` trả dạng chuẩn đó — chính là chuỗi repository **lưu xuống DB** và
`NewExchangeRate` **đọc lên lại**. Không có nó, `adapter/postgres` không có cách
nào persist tỷ giá (field đóng kín, không getter).

> **Bài học:** value object cần **(1)** một dạng chuẩn để `==` đúng, và **(2)**
> một cách xuất ra để persist — mà không mở toang field. `Rate() string` làm cả
> hai. Cùng lý do, `Dimensions` có `LengthMM()/WidthMM()/HeightMM()`.

> **So với Doctrine:** một `#[ORM\Embeddable]` với 3 cột `rate_from`, `rate_to`,
> `rate_value` (text). Bên Symfony bạn quen để Doctrine ghi thẳng vào private
> field qua Reflection; Go không có Reflection kiểu đó cho field private, nên
> getter có mục đích rõ là bắt buộc.

## 12. Entity — Thực thể ✅

**Là gì.** Object **có danh tính riêng**, phân biệt được kể cả khi mọi thuộc
tính giống hệt nhau.

**Khác Value Object thế nào:**

| | Value Object | Entity |
|---|---|---|
| Danh tính | không có | có ID |
| So sánh | theo giá trị | theo ID |
| Thay đổi | bất biến | thay đổi theo thời gian |
| Ví dụ | `Money`, `Weight`, `FreeShipping` | `Merchant` ✅, `CustomerOrder`, `PurchaseTask` |

Hai đơn hàng cùng sản phẩm, cùng giá, cùng ngày — vẫn là **hai đơn khác nhau**.
Vì chúng có ID riêng, và mỗi cái có số phận riêng.

**Trong Portage** — ID sinh trong domain, **không đợi database**. Hạ tầng đã
code ở `internal/domain/shared/id.go`:

```go
type ID struct{ u uuid.UUID }          // UUID phiên bản 7

func NewID() ID                        // sinh mới, có timestamp ở 48 bit đầu
func ParseID(s string) (ID, error)     // đọc từ URL / DB — CHỈ nhận v7
func (id ID) String() string
func (id ID) IsZero() bool             // ID{} = "chưa có danh tính"
```

**Vì sao UUID v7, không phải v4?** v7 bắt đầu bằng thời điểm sinh (mili-giây),
nên **sắp xếp theo thời gian tạo** và index B-tree của Postgres ghi nối đuôi
thay vì rải ngẫu nhiên khắp các page như v4. Không phải auto-increment vì
auto-increment (1) phải hỏi DB mới có, (2) lộ số lượng đơn cho người ngoài.

**Mỗi aggregate bọc `ID` thành kiểu riêng** — một dòng, và compiler phân biệt:

```go
type OrderID struct{ shared.ID }             // embedding: kế thừa String(), IsZero()
func NewOrderID() OrderID { return OrderID{shared.NewID()} }

type MerchantID struct{ shared.ID }

func Load(id OrderID) { ... }
Load(merchantID)     // ❌ KHÔNG COMPILE — đúng thứ mình muốn
```

Rồi aggregate dùng nó ngay trong constructor:

```go
func New(...) (*CustomerOrder, error) {
    return &CustomerOrder{
        id: NewOrderID(),   // có danh tính NGAY, chưa chạm DB
        ...
    }, nil
}
```

> **Về Dependency Rule:** `github.com/google/uuid` là thư viện **ngoài** đầu
> tiên trong domain. Nó là *tiện ích thuần* (như `symfony/uid`) — không biết DB,
> không biết HTTP — nên được vào allowlist (xem mục 21). Framework và SDK thì
> không.

> **So với Doctrine:** entity Symfony của bạn có `private ?int $id = null` — ID
> là `null` cho tới khi `flush()`. Nghĩa là object *tồn tại mà chưa có danh
> tính*. Trong DDD đó là trạng thái không hợp lệ. Tương đương đúng bên Symfony
> là `symfony/uid`: `#[ORM\Column(type: 'uuid')]` **không** `#[ORM\GeneratedValue]`,
> và constructor tự gán `$this->id = Uuid::v7()`. `final class OrderId` bọc
> `Uuid` bên PHP ~ `type OrderID struct{ shared.ID }` bên Go.

## 13. Invariant — Bất biến thức ✅

**Là gì.** Quy tắc **luôn phải đúng**, không có ngoại lệ, không có "trừ trường
hợp".

**Nguyên tắc vàng:** object **không được phép tồn tại** ở trạng thái vi phạm
invariant — *dù chỉ một phần nghìn giây*.

**Trong Portage:**

| Invariant | Ở đâu | Trạng thái |
|---|---|---|
| Cân nặng không âm | `shared.Weight` | ✅ |
| Tỷ giá > 0, hai tiền tệ khác nhau | `shared.ExchangeRate` | ✅ |
| Tiền luôn có currency (không xây trên `Currency{}`) | `shared.Money` | ✅ |
| Số tiền cọc = 50% tổng đơn | `CustomerOrder` | 🔜 |
| Đơn đã giao thì không huỷ được | `CustomerOrder` | 🔜 |
| Trọng lượng tính phí ≥ cân thật | `Parcel` | 🔜 |
| Lô đã đóng thì không thêm đơn được | `ConsolidationBatch` | 🔜 |
| Đơn giá phải cùng tiền tệ với tổng | `Quote` | 🔜 |

**Cách bảo vệ trong Go** — constructor là cửa duy nhất:

```go
// Trường chữ thường = private trong package. KHÔNG CÓ setter.
type CustomerOrder struct {
    id      OrderID
    status  Status
    deposit shared.Money
}

// Cách DUY NHẤT để tạo. Không thể tạo ra thứ vô nghĩa.
func New(...) (*CustomerOrder, error) {
    if ... { return nil, ErrXxx }
    return &CustomerOrder{ status: StatusDraft, ... }, nil
}
```

Ngoài package, `o.status = "bananas"` **không biên dịch được**.

> **Mạnh hơn PHP ở chỗ nào:** `private` của PHP chỉ chặn lúc chạy, và Doctrine
> dùng Reflection chọc thủng thoải mái. Go chặn ở **compiler** — sai là không
> build được, không đợi tới lúc chạy.

### Hai loại constructor — `error` hay `panic`? ✅

Go không có exception. Khi input vi phạm invariant, có hai cách phản ứng, và
**chọn cách nào phụ thuộc vào AI gọi**:

```go
// Input KHÔNG TIN ĐƯỢC (form, CSV, dòng DB) → trả error, caller xử lý
func NewWeight(grams int64) (Weight, error) {
	if grams < 0 {
		return Weight{}, fmt.Errorf("%d g: %w", grams, ErrNegativeWeight)
	}
	return Weight{grams: grams}, nil
}

// Literal do LẬP TRÌNH VIÊN viết (code, test) → panic: đây là bug, sửa code
func Grams(g int64) Weight {
	w, err := NewWeight(g)
	if err != nil {
		panic("shared: " + err.Error())
	}
	return w
}
```

| | Validating (`Parse*`, `New*`) | Literal (`Grams`, `NewMoney`, `Must*`) |
|---|---|---|
| Ai gọi | adapter, với dữ liệu từ ngoài | code nghiệp vụ, test, hằng số |
| Sai thì | `error` — có thể bắt, báo lại người dùng | `panic` — dừng, đây là bug |
| Symfony | `\DomainException`, `\InvalidArgumentException` | `\LogicException` |

`panic` **không bao giờ** dùng cho lỗi của người dùng. Nếu thấy mình muốn
`recover()` một panic để hiển thị thông báo lỗi → đã dùng nhầm loại constructor.

Cùng logic: `errors.Is(err, shared.ErrNegativeWeight)` bên Go = `catch
(NegativeWeight $e)` bên PHP. Sentinel error (`var ErrX = errors.New(...)`) +
`fmt.Errorf("...: %w", ErrX)` để bọc thêm ngữ cảnh mà vẫn `errors.Is` được.

## 14. Aggregate & Aggregate Root ✅ (Merchant, Product)

**Là gì.** Một **cụm** object luôn phải nhất quán với nhau, và **một** object
làm cửa duy nhất ra vào cụm đó — gọi là **Aggregate Root**.

```
        ┌─ RANH GIỚI AGGREGATE ──────────────┐
        │                                    │
        │   CustomerOrder  ← Aggregate Root  │
        │        │                           │
        │        ├── OrderLine  (entity con) │
        │        ├── OrderLine                │
        │        └── Deposit    (value obj)  │
        │                                    │
        └────────────────────────────────────┘
              ▲
              │ Thế giới bên ngoài CHỈ được chạm vào root.
              │ Không ai với tay vào sửa thẳng OrderLine.
```

**Bốn luật của aggregate:**

1. Bên ngoài chỉ tham chiếu tới **root**, không tham chiếu vào ruột
2. Mọi thay đổi đi qua **method của root**
3. Root chịu trách nhiệm giữ **invariant** cho cả cụm
4. **Một transaction = một aggregate.** Nhiều aggregate thì dùng event

**Luật thứ 4 là luật hay bị vi phạm nhất**, và trong Portage nó cực kỳ quan
trọng:

```
   ❌ SAI — gộp làm một, đúng phản xạ Doctrine

        CustomerOrder
          └── PurchaseTask   ← nhét vào cùng aggregate

   Hệ thống sụp ngay lần đầu Nike hết hàng: bạn buộc phải sửa
   trạng thái đơn khách và việc mua trong cùng một transaction,
   trong khi hai thứ đó có nhịp sống hoàn toàn khác nhau.

   ✅ ĐÚNG — hai aggregate, hai context, nối bằng event

        ORDERING                      PROCUREMENT
        CustomerOrder  ──event──▶     PurchaseTask
        (khách thấy)                  (nhân viên mua thấy)
```

**Vì sao?** Một đơn khách có thể cần **3 việc mua ở 2 nhà bán**. Việc mua thứ 2
hết size. Khách đã trả cọc rồi. Hai thứ này **không thể** chung một transaction.

**Quy tắc chọn kích thước aggregate: càng nhỏ càng tốt.** Aggregate to = khoá
nhiều dòng = tắc nghẽn khi nhiều người dùng cùng lúc.

**Aggregate có entity con — đã code:** `Product` ⊃ `Variant`
(`internal/domain/catalog/product.go`, `variant.go`).

```go
type Product struct {
    shared.Events
    id       ProductID
    variants []Variant              // ← entity con, sống trong root
    ...
}

func (p *Product) AddVariant(d VariantDetails, now time.Time) (VariantID, error)  // cửa DUY NHẤT
func (p *Product) Variants() []Variant                                            // trả BẢN SAO
```

Bốn luật hiện ra trong code thế nào:

| Luật | Trong `Product` |
|---|---|
| 1. Ngoài chỉ chạm root | không có `*Variant` nào rò ra; `Variants()` trả copy; **không có `VariantRepository`** |
| 2. Đổi qua method của root | `AddVariant`; không có `variant.SetSize()` |
| 3. Root giữ invariant cho cả cụm | trùng (size, color) bị `AddVariant` từ chối — `Variant` không tự biết anh em của nó |
| 4. Một transaction = một aggregate | gộp hai `Product` trùng nhau là **workflow tầng app**, không phải method — vì đụng hai root |

`Variant` có `VariantID` (khách đặt *đúng* variant này, procurement mua *đúng*
variant này) nhưng không có vòng đời riêng — đó là định nghĩa "entity con".

## 15. Domain Event — Sự kiện nghiệp vụ ✅

**Là gì.** Ghi nhận **một việc đã xảy ra** trong nghiệp vụ. Tên luôn ở **thì quá
khứ**.

```go
type DepositPaid struct {
    OrderID OrderID
    Amount  shared.Money
    At      time.Time
}
```

**Vì sao cần.**

1. Để aggregate này báo cho aggregate khác mà **không phụ thuộc** vào nhau
2. Để lưu lại **lịch sử** — thứ mà cột `status` không bao giờ cho biết
3. Để thêm tính năng mới **không phải sửa code cũ** (thêm người nghe event)

**Cách aggregate phát event:**

```go
type CustomerOrder struct {
    ...
    events []DomainEvent      // gom lại, chưa gửi
}

func (o *CustomerOrder) PayDeposit(amount shared.Money, now time.Time) error {
    if o.status != StatusAwaitingDeposit {
        return ErrInvalidTransition
    }
    o.status = StatusDeposited
    o.record(DepositPaid{OrderID: o.id, Amount: amount, At: now})
    return nil
}
```

Chú ý: aggregate **không tự gửi** event đi. Nó chỉ **ghi lại**. Tầng application
lấy ra và phát sau khi lưu thành công. (Xem mục Outbox.)

**Hạ tầng đã code** — `internal/domain/shared/event.go`:

```go
// Event: một việc ĐÃ xảy ra. Tên dây (wire name) ổn định để ghi outbox.
type Event interface {
	EventName() string        // "ordering.deposit_paid" — không đổi khi đổi tên type Go
	OccurredAt() time.Time
}

// Events: bộ ghi mà aggregate root NHÚNG vào. Không phải value object —
// là kiểu duy nhất trong shared có pointer receiver, vì nó là sổ ghi chép.
type Events struct{ pending []Event }

func (e *Events) Record(ev Event)        // aggregate gọi trong method đổi trạng thái
func (e *Events) PullEvents() []Event    // app layer gọi SAU KHI save; trả ra và xoá
```

Cách aggregate sẽ dùng — chỉ là **nhúng** (embedding), không kế thừa:

```go
type CustomerOrder struct {
    shared.Events          // ← một dòng, CustomerOrder có sẵn Record/PullEvents
    id     OrderID
    status Status
}
```

Vòng đời một event, đúng thứ tự:

```
   1. o.PayDeposit(...)      → bên trong gọi o.Record(DepositPaid{...})   [domain]
   2. repo.Save(ctx, o)      → ghi aggregate                               [adapter]
   3. evs := o.PullEvents()  → lấy ra, sổ ghi trống lại                    [app]
   4. outbox.Append(evs)     → CÙNG transaction với bước 2                 [adapter]
   5. worker đọc outbox → publish                                          [cmd/worker]
```

Bước 3 **phải sau** bước 2: publish trước khi lưu là cách worker đi mua giày
cho một đơn chưa từng tồn tại. Test: `TestEvents_recordThenPullInOrderAndClear`.

**Event của Portage:**

```
QuoteIssued  →  OrderPlaced  →  DepositPaid  →  PurchaseRequested
                                                      │
                        ┌─────────────────────────────┤
                        ▼                             ▼
                PurchaseConfirmed              PurchaseFailed
                        │                     (hết size / hết hàng)
                        ▼                             │
                 ParcelReceived                       ▼
                        │                     RefundIssued
                        ▼
                  BatchClosed  →  BatchShipped  →  OrderDelivered
```

> **So với Symfony:** giống `EventDispatcher` + `EventSubscriber`, nhưng khác ở
> chỗ **domain event là một phần của model nghiệp vụ**, không phải cơ chế kỹ
> thuật. Nó xuất hiện trong ngôn ngữ chung: bạn nói *"khi khách trả cọc thì…"*,
> và trong code có đúng `DepositPaid`. Cặp `Record`/`PullEvents` chính là
> `private array $domainEvents` + `pullDomainEvents()` bạn thấy trong các entity
> Doctrine làm DDD, được một `postFlush` listener rút ra đẩy vào Messenger.

## 16. Repository — Kho chứa ✅ interface + in-memory + Postgres

**Là gì.** Một **interface** cho phép domain lấy và lưu aggregate, mà **không
biết** dữ liệu nằm ở đâu.

**Điểm mấu chốt — chiều phụ thuộc bị đảo ngược:**

```go
// internal/domain/ordering/repository.go
// Domain KHAI BÁO cái nó cần. Không biết Postgres tồn tại.
type Repository interface {
    Load(ctx context.Context, id OrderID) (*CustomerOrder, error)
    Save(ctx context.Context, o *CustomerOrder) error
}
```

```go
// internal/adapter/postgres/order_repo.go
// Adapter CÀI ĐẶT nó. Đây mới là nơi biết SQL.
type OrderRepo struct { db *pgxpool.Pool }

func (r *OrderRepo) Load(ctx context.Context, id ordering.OrderID) (*ordering.CustomerOrder, error) {
    // SQL ở đây
}
```

**Ba luật:**

1. Repository làm việc với **aggregate root**, không phải từng bảng
2. Interface đặt trong **domain**, cài đặt đặt trong **adapter**
3. Không có `findByStatusAndDateAndCustomer(...)` — truy vấn phức tạp thuộc về
   **read model** (xem CQRS)

> **So với Doctrine:** `BookingRepository extends ServiceEntityRepository` của
> bạn **kế thừa từ Doctrine** — tức domain biết Doctrine tồn tại. Ở đây thì
> ngược lại: domain chỉ khai báo interface, Postgres cắm vào từ ngoài.

**Đã code** — `internal/domain/catalog/repository.go`, hai PORT đầu tiên:

```go
type MerchantRepository interface {
	ByID(ctx context.Context, id MerchantID) (*Merchant, error)      // ErrMerchantNotFound
	BySite(ctx context.Context, site Hostname) (*Merchant, error)    // link khách đưa → mua ở đâu
	Save(ctx context.Context, m *Merchant) error                     // KHÔNG publish event
}

type CategoryRepository interface {
	ByCode(ctx context.Context, code CategoryCode) (CategoryPolicy, error)
	All(ctx context.Context) ([]CategoryPolicy, error)
	Save(ctx context.Context, p CategoryPolicy) error                // thay cả cục — VO không update từng phần
}
```

Ba điểm đáng để ý:

- **`Save` không phát event.** Tầng app gọi `m.PullEvents()` *sau khi* `Save`
  trả về, rồi ghi outbox **cùng transaction** (mục 15, mục 27). Repository mà
  tự publish là trộn hai trách nhiệm và mất khả năng làm outbox.
- **`context.Context` đứng đầu mọi method có thể chặn.** Nó mang tín hiệu huỷ
  và deadline từ request HTTP xuống tới câu SQL. PHP-FPM không có khái niệm này
  vì mỗi request là một process sống độc lập rồi chết.
- **Không có `ByStatusAndCurrencyAddedAfter(...)`.** Màn hình danh sách/lọc là
  read model (mục 24), không phải việc của repository.

**Cài đặt đầu tiên: in-memory** — `internal/adapter/memory/` (`MerchantRepo`,
`CategoryRepo`, `ProductRepo`). Đây là adapter thật, không phải fake trong
`tests/`: `cmd/api` chạy dev dùng được. Mỗi repo có một dòng khẳng định
compile-time rằng nó thoả port:

```go
var _ catalog.MerchantRepository = (*MerchantRepo)(nil)
```

**Cài đặt thứ hai: Postgres** — `internal/adapter/postgres/` (pgx v5). Ba điểm DDD:

- Aggregate toàn field private → repository đi qua **memento**
  `Snapshot()`/`FromSnapshot()` (`catalog/snapshot.go`), không qua export field, không
  qua Reflection. `FromSnapshot` không phát event và từ chối hình dạng không thể có.
- `Save` là **upsert** (`ON CONFLICT (id) DO UPDATE`): repository không hỏi "mới hay cũ",
  aggregate không có cờ `isNew`. Variants xoá + chèn lại cùng cha — entity con đi theo
  root (mục 14).
- Repository **không mở transaction**: `UnitOfWork.InTx` mở, bỏ `tx` vào `ctx`, repo
  lấy ra bằng `db(ctx)`. Cùng handler, cùng `ctx` → cùng transaction với outbox.

Chi tiết từng dòng: [WALKTHROUGH.md](WALKTHROUGH.md) §6a–§6d.

## 17. Domain Service ✅ (`pricing.Calculate` — hàm thuần; `logistics.FreightAllocator` — interface)

**Là gì.** Nơi đặt logic nghiệp vụ **không thuộc về riêng một aggregate nào**.

**Khi nào cần.** Khi một quy tắc liên quan tới nhiều aggregate, hoặc là một phép
tính thuần không có trạng thái.

**Đã code — `pricing.Calculate` (`internal/domain/pricing/calc.go`):** một hàm thuần từ
`QuoteInputs` (listing, profile, lane, tỷ giá, policy) ra `Breakdown`. Không clock, không
repository, không trạng thái — nên kiểm được từng dòng bằng bảng tính:

```
chargeable = lane.ChargeableWeight(parcel)          parcel = đã đo, không thì hộp mặc định của category → Estimated = true
subtotal   = item + item×tax + freight(class, chargeable) + surcharge + duty
total      = subtotal×fx + max(subtotal×fx × margin%, floor)
deposit    = total × deposit%
```

Nó không thuộc `Quote` (Quote chỉ *giữ* kết quả), không thuộc `ShippingLane` (lane không
biết policy), không thuộc `Listing` (listing không biết lane). Ba value object góp mặt,
không ai sở hữu phép tính → Domain Service. Là **hàm**, không interface: chỉ có một cách
tính; interface đến khi có cách thứ hai (đúng tinh thần `FreightAllocator` bên dưới).

**Đã code (06/09) — ví dụ hay nhất để thấy vì sao cần interface:**

```go
// internal/domain/logistics/allocator.go
type FreightAllocator interface {
	Allocate(freight shared.Money, items []BatchItem) ([]Allocation, error)
}
type ByChargeableWeight struct{ Divisor int64; Step shared.Weight }   // implementation #1: theo cân tính phí
```

`ConsolidationBatch.Ship(freight, alloc, now)` nhận allocator làm **tham số** — batch không
biết luật chia, use case chọn luật từ `LaneRule` của tuyến. Bên dưới là `shared.Allocate`
(floor + largest remainder, `big.Int`) để 87.50 chia cho 2500 g và 5000 g ra **29.17 + 58.33 =
87.50** đúng từng cent. Bản phác cũ:

```go
// Chia cước của cả thùng cho từng đơn trong lô.
// Không thuộc về CustomerOrder (nó chỉ biết mình).
// Không thuộc về Batch (nó không biết luật chia).
type FreightAllocator interface {
    Allocate(batch Batch, actualFreight shared.Money) map[OrderID]shared.Money
}
```

Chia theo trọng lượng tính phí? theo thể tích? theo trị giá? **Ba cách cho ba
kết quả khác nhau**, và mỗi cách thiệt cho một nhóm khách khác nhau. Đây là
quyết định nghiệp vụ thật, xứng đáng có chỗ riêng trong code.

**Phân biệt với Application Service:**

| Domain Service | Application Service |
|---|---|
| Chứa **quy tắc nghiệp vụ** | Chỉ **điều phối** |
| Nằm trong `internal/domain/` | Nằm trong `internal/app/` |
| Không biết DB, không biết transaction | Mở transaction, gọi repository, phát event |
| `pricing.Calculate`, `FreightAllocator` | `pricingapp.IssueQuoteHandler` — gom listing/lane/profile/fx rồi **gọi** `Calculate` qua `IssueQuote` |

## 18. Factory 📋

**Là gì.** Nơi tập trung logic **tạo** aggregate khi việc tạo phức tạp.

Trong Go thường chỉ cần hàm `New...()` trong package là đủ — không cần class
Factory như Java/PHP.

```go
func New(c CustomerRef, lines []Line, q Quote, now time.Time) (*CustomerOrder, error)
```

## 19. Policy / Specification 📋

**Là gì.** Đóng gói một **quy tắc quyết định** thành object riêng, để nó có tên
và test được độc lập.

**Trong Portage:**

```go
// Luật hoàn cọc — quy tắc nghiệp vụ quan trọng nhất của bạn
type RefundPolicy interface {
    RefundFor(o *CustomerOrder, at time.Time) (shared.Money, error)
}

// Luật cảnh báo rủi ro tuyến vận chuyển
type LaneRiskPolicy interface {
    Assess(lane LaneCode, shipmentsThisMonth int) RiskLevel
}
```

**Vì sao tách ra.** Vì luật hoàn tiền **sẽ đổi** khi bạn kinh doanh thật. Tách
riêng thì đổi luật không phải đụng vào aggregate.

---

# PHẦN IV — Kiến trúc

## 20. Ports & Adapters (Hexagonal Architecture) ✅

**Là gì.** Kiến trúc đặt domain ở **giữa**, mọi thứ bên ngoài (DB, HTTP, OpenAI,
web bán hàng) cắm vào qua **interface**.

- **Port** = interface do domain khai báo — *"tôi CẦN gì"*
- **Adapter** = cài đặt cụ thể — *"làm BẰNG gì"*

```
        ┌─────────────────────────────────────────┐
        │   DOMAIN  (Go thuần, không import gì)   │
        │                                          │
        │   type PriceScraper interface {          │  ← PORT
        │       Fetch(ctx, url) (Product, error)   │    domain khai báo
        │   }                                      │    cái nó cần
        └───────────────┬──────────────────────────┘
                        │ implements
       ┌────────────────┼────────────────┐
       ▼                ▼                ▼
  ┌─────────┐    ┌────────────┐   ┌────────────┐
  │ OpenAI  │    │  Operator  │   │ Fake (test)│   ← ADAPTER
  │ adapter │    │  nhập tay  │   │            │     cách làm cụ thể
  └─────────┘    └────────────┘   └────────────┘
```

**Vì sao đây là ý tưởng quan trọng nhất về AI trong project này.**

Chỗ 90% project "AI + DDD" làm sai: nhét lời gọi OpenAI vào giữa business logic.
Đúng phải là ngược lại — **AI chỉ là một loại "database" khác**: một thứ ở ngoài,
chậm, hay lỗi, tốn tiền, phải cắm vào qua interface.

Nhờ vậy:

- Test domain **không cần API key, không tốn tiền**, chạy trong 5ms
- Đổi từ OpenAI sang model khác = viết adapter mới, domain **không đổi một dòng**
- Chạy song song nhiều adapter để so sánh chất lượng

**Bằng chứng đã hoạt động trong Portage:** 265/296 test chạy **không cần gì cả**,
không Docker, không mạng, không API key (31 test còn lại là tích hợp Postgres) — kể cả 32 test tầng app đi
qua thật handler → domain → repository → outbox, vì repository và outbox là
adapter in-memory cắm vào port.

**Port và adapter đã có thật:**

| Port (interface) | Khai báo ở | Adapter đã cài |
|---|---|---|
| `MerchantRepository`, `CategoryRepository`, `ProductRepository` | `internal/domain/catalog/repository.go` | `internal/adapter/memory/` |
| `Clock` | `internal/app/ports.go` | `internal/platform/clock/` — `System`, `Fixed` |
| `UnitOfWork`, `Outbox` | `internal/app/ports.go` | `internal/adapter/memory/` |
| `procurement.MerchantACL` | `internal/domain/procurement/repository.go` | `merchant.Manual` (một con người) · `merchant.Router` (chọn theo shop) — §22 |
| `catalogapp.ListingExtractor` — *"trang này bán cái gì"* | `internal/app/catalog/extract.go` | `openai.Client` · `openai.Fake` (dev) · `openai.Unavailable` (không key → 503) |
| `reportingapp.OrderSummaryRepository` | `internal/app/reporting/summary.go` | `memory` · `postgres` — bảng `order_summaries` |
| `logistics.FreightAllocator` | `internal/domain/logistics/allocator.go` | `ByChargeableWeight` — Domain Service, adapter ở ngay trong domain (§17) |
| `auth.Verifier` — *"token này là ai"* | `internal/platform/auth/auth.go` | `auth.Static` (RAM, dev + test) · `postgres.TokenRepo` (thật) |
| `auth.Issuer` — *"cắt cho tôi một chìa"* | `internal/platform/auth/issue.go` | cùng hai adapter trên |
| `auth.Registry` — *"những chìa nào đang có"* | `internal/platform/auth/registry.go` | cùng hai adapter trên |

Quy tắc chỗ đặt: **interface nằm ở nơi CẦN nó** (domain cần repository → ở
domain; app cần clock → ở app), không nằm ở nơi cài.

Hai dòng cuối là ví dụ rõ nhất của quy tắc đó, và cũng là câu hỏi khó nhất của
P9/T1: **`Verifier` để ở đâu?** Phản xạ là `internal/domain/identity/`. Sai —
vì *"ai đang gọi"* **không phải quy tắc nghiệp vụ**, mà là thứ **biên** xác lập
*trước khi* use case chạy. `Product.Publish` cần một `shared.OperatorID`; nó
không cần biết id ấy đến từ token, từ session, hay từ người gõ tay. Nên `auth`
nằm cạnh `clock` trong `platform/` — cả hai trả lời một câu hỏi mà domain
**nhận kết quả** chứ không **đi hỏi**:

```
   clock  →  "bây giờ là mấy giờ"   →  domain nhận `now time.Time`
   auth   →  "ai đang gọi"          →  domain nhận `shared.OperatorID`
```

Bằng chứng không phải lời hứa: guard 6 (*domain chỉ import stdlib + allowlist*)
vẫn xanh sau T1, vì `internal/domain/` **không import `auth`** một lần nào.

Và `Verifier` tách khỏi `Issuer` không phải cho đẹp: **một route phát chìa, ba
mươi route chỉ đọc**. Interface nhỏ là cách duy nhất nói điều đó với compiler —
cầm `Verifier` thì không có cách nào phát ra token mới. Chi tiết: WALKTHROUGH §18.

## 21. Dependency Rule — Luật chiều phụ thuộc ✅

**Là gì.** Mũi tên phụ thuộc **chỉ được hướng vào trong**. Domain không bao giờ
trỏ ra ngoài.

```
   cmd  ──▶  app  ──▶  domain
                ▲
   adapter ─────┘         domain KHÔNG trỏ ra ngoài, BAO GIỜ
```

**Trong Portage** — luật này ghi ngay đầu file `money.go`:

```go
// RULES FOR THIS PACKAGE (and every package under internal/domain):
//   - It may NOT import a database driver, an HTTP framework, or any SDK.
//   - It may NOT import another bounded context.
//   - Everything here is immutable: methods return new values, never mutate.
```

**Cách Go ép buộc:** thư mục tên `internal/` được **chính trình biên dịch** bảo
vệ — package ngoài module không import được. Không cần reviewer nhắc.

**Kiểm tra bằng lệnh** — chạy được ngay bây giờ, và nên chạy lại mỗi khi thêm
code vào domain:

```bash
# Liệt kê những gì domain import TRỰC TIẾP
go list -f '{{range .Imports}}{{println .}}{{end}}' ./internal/domain/... | sort -u
```

Kết quả hiện tại (04/09):

```
errors
fmt
github.com/google/uuid      ← thư viện ngoài DUY NHẤT, trong allowlist
math/big
strconv
strings
time
```

**Allowlist là gì và vì sao có.** Luật "domain không import gì ngoài" nói cho
đúng là "không import **framework, driver, SDK**" — những thứ có side effect,
biết mạng, biết đĩa. Một thư viện *tiện ích thuần* như `google/uuid` (sinh số
ngẫu nhiên + format) không phá luật, giống `symfony/uid` không làm entity
Symfony "bẩn". Nhưng phải **ghi rõ**, không để lọt âm thầm: allowlist nằm trong
`.github/workflows/ci.yml`, thêm gì phải sửa file đó và giải thích trong commit.

Kiểm tra bằng máy — **CI chạy bước này mỗi lần push**, build đỏ nếu vi phạm:

```bash
# Thư viện ngoài luôn có dấu chấm trong tên (github.com/…, gorm.io/…)
allow='^github\.com/google/uuid$'
go list -f '{{range .Imports}}{{println .}}{{end}}' ./internal/domain/... \
  | sort -u | grep "\." | grep -Ev "$allow"
# Không in ra gì  =  domain sạch
# In ra bất cứ dòng nào  =  có thứ bên ngoài đã rò vào domain
```

> **So với Symfony:** đây là việc `deptrac` làm — khai báo layer và luật ai được
> import ai, chạy trong CI. Ở đây không cần tool riêng vì `go list` có sẵn.

> ⚠️ Đừng dùng `go list -deps` để kiểm tra việc này — nó liệt kê cả dependency
> **gián tiếp**, nên sẽ hiện `runtime`, `os`, `syscall`… (những thứ `fmt` kéo
> theo) và làm bạn tưởng domain bẩn. `.Imports` mới là import trực tiếp.

## 22. Anti-Corruption Layer (ACL) ✅ HAI cái: `procurement.MerchantACL` (shop) và `catalogapp.ListingExtractor` (web lạ + AI)

**Là gì.** Lớp **dịch** giữa domain của bạn và một hệ thống bên ngoài mà bạn
không kiểm soát.

**Vì sao cần.** Vì model của Nike (style / colorway / size US) mà rò vào domain
của bạn thì domain sẽ dần biến dạng theo họ — và khi thêm Adidas với model khác
hẳn, bạn kẹt.

```
   ┌────────────────┐        ┌─────────────┐       ┌──────────────┐
   │ DOMAIN CỦA BẠN │◀──────▶│     ACL     │◀─────▶│  Nike, TNF,  │
   │                │  ngôn  │  lớp dịch   │  ngôn │  Sony, ...   │
   │ Product,       │  ngữ   │             │  ngữ  │              │
   │ Category       │  CỦA   │             │  CỦA  │  style,      │
   │                │  BẠN   │             │  HỌ   │  colorway    │
   └────────────────┘        └─────────────┘       └──────────────┘
```

**Trong Portage** — mỗi nhà bán một adapter, cùng một interface:

```
internal/adapter/merchant/
├── manual.go     ✅ adapter #1: một con người (ErrManualPurchase) — mọi shop hôm nay
├── router.go     ✅ chọn adapter theo shop; không có trong map → manual.go
└── nike.go       🔜 ngày Nike cho API: thêm file này + một dòng trong wire
```

**Đây là pattern DDD sinh ra đúng cho bài toán của bạn.** Và đó là lý do khi
adapter đổi từ "nhân viên nhập tay" sang "API thật của nhà phân phối", **domain
không đổi một dòng**.

**Đã code (05/09):** port ở domain, adapter đầu tiên là **một con người**.

```go
// internal/domain/procurement/repository.go — PORT, bằng từ vựng CỦA MÌNH
type MerchantACL interface {
	Purchase(ctx context.Context, task *PurchaseTask) (receipt PurchaseReceipt, ok bool, reason string, err error)
}

// internal/adapter/merchant/manual.go — adapter #1: không mua được, và nói thẳng
func (Manual) Purchase(…) (procurement.PurchaseReceipt, bool, string, error) {
	return procurement.PurchaseReceipt{}, false, "", procurement.ErrManualPurchase
}
```

`procurementapp.OpenTaskHandler` gọi port **ngay khi task mở** và có bốn nhánh cho bốn câu trả
lời: `ErrManualPurchase` → task ở `open` cho người mua (đường hôm nay); `ok` → `Confirm` ngay;
`!ok` + reason → `Fail` ngay; lỗi khác → trả lỗi, relay thử lại. Một fake `apiShop` 10 dòng trong
test chứng minh cả bốn — **trước khi** shop nào có API. Ngày Nike có API: thêm `adapter/merchant/nike.go`,
đổi `ACL: merchant.Manual{}` thành `merchant.Nike{…}` trong `wire` — use case không thêm `if`.
Mổ xẻ: WALKTHROUGH.md §16c.

**Đã code (06/09) — ACL thứ HAI, và một Router (P9/T5, T6).**

Hai ACL, hai nỗi lo khác nhau, cùng một hình dạng:

```
   procurement.MerchantACL       "đặt hàng ở shop này"    ← rủi ro: API người ta đổi/chết
   catalogapp.ListingExtractor   "trang này bán cái gì"   ← rủi ro: mô hình ĐOÁN, rất tự tin
```

```go
// internal/app/catalog/extract.go — PORT
type ListingDraft struct{ Name, Price, Currency, CategoryHint string }   // BỐN string
type ListingExtractor interface {
	Extract(ctx context.Context, url catalog.SourceURL) (ListingDraft, error)
}
```

**Bốn string, không phải `shared.Money`** — nếu port trả `Money` thì port đang
quyết cái gì hợp lệ, mà đó là việc của domain. Chuỗi của máy đi qua đúng cái
cửa mà chuỗi người gõ tay đi qua (`shared.ParseMoney`), nên model trả
`"about $150"` chết ngay ở biên chứ không ba bảng sau.

Và luật quan trọng nhất: **máy được TẠO nháp, không được XÁC NHẬN.** Provenance
là `SourcedByFeed`, còn `Product.Publish()` từ chối listing chưa `Verified()` —
nên không thứ gì mô hình bịa ra thành được giá bán trước khi một con người bấm
confirm-listing. Tính năng mới **không nới** luật cũ.

Ba adapter cho một port, và cái thứ ba mới là bài học:

```
   openai.Client       gọi API thật (net/http thuần, json_schema strict, timeout)
   openai.Fake         dev + test: `go run ./cmd/api` chạy không cần key, không tốn tiền
   openai.Unavailable  Postgres mà không có OPENAI_API_KEY → 503
```

> Postgres **tuyệt đối không** rơi về `Fake`. Rơi về thì API trả 201, một sản
> phẩm nháp tồn tại với tên và giá **bịa**, và dấu hiệu duy nhất cho thấy có gì
> sai là các con số toàn hư cấu. **Tính năng tắt phải trông như đang tắt.**

**`merchant.Router` (T6)** trả lời câu "shop nào có API": nó *chính nó* là một
`MerchantACL`, nên `OpenTaskHandler` không đổi một dòng.

```go
merchant.Router{Items: items, ByMerchant: map[shared.ID]procurement.MerchantACL{
	nikeID: merchant.Nike{…},   // ngày Nike cho API: THÊM MỘT DÒNG, ở wire
}}                              // shop không có trong map → Manual → người mua
```

Hai lựa chọn còn lại đều tệ hơn: một `if` trong use case sẽ mọc thêm một nhánh
mỗi shop **mãi mãi**; một ACL biết mọi shop thì shop nào chết cũng thành mọi
shop chết. Chi tiết: WALKTHROUGH.md §21.

## 23. Đa thương hiệu = dữ liệu, không phải code ✅

Yêu cầu của bạn — *"mở rộng cho brand khác, không chỉ giày"* — có một hệ quả kỹ
thuật rất cụ thể:

> **Từ "Nike" không được xuất hiện một lần nào trong `internal/domain/`.**
> Nike là **một dòng trong bảng**, không phải một nhánh `if`.
> Thêm Adidas = thêm một dòng dữ liệu, **không phải deploy**.

```go
// ❌ SAI — Nike bị nhốt cứng trong nghiệp vụ
if order.Brand == "Nike" { ... }

// ✅ ĐÚNG — merchant là dữ liệu  (catalog/merchant.go, bản thật)
type Merchant struct {
    shared.Events
    id       MerchantID
    name     string          // "Nike", "The North Face", "Sony" — chỉ là DỮ LIỆU
    site     Hostname
    currency shared.Currency
    freeShip FreeShipping    // never / over $50 / always — VO 3 trạng thái
    sourcing []SourcingMode  // operator | customer | feed
    status   MerchantStatus  // active | suspended
}

// ✅ ĐÚNG — ngành hàng nói món hàng LÀ gì  (catalog/category.go, bản thật)
type CategoryPolicy struct {
    code         CategoryCode   // footwear | apparel | electronics — khoá tự nhiên
    estimate     Estimate       // cân nặng + hộp ước lượng → quote khi chưa đo thật
    restrictions []Restriction  // battery, liquid, supplement…
}
// KHÔNG có thuế/HS code, KHÔNG có số chia thể tích: đó là dữ liệu của TUYẾN
// vận chuyển → pricing.ShippingLane. Xem CATALOG.md §6 và test canh số 4.
```

`Restriction{Battery}` là thật: tai nghe Sony có pin lithium, không đi được mọi
tuyến bay. Nếu domain không biết khái niệm này, ngày đầu bán đồ điện tử là ngày
hàng kẹt ở kho.

## 24. CQRS ✅ đủ hai nửa — view từ aggregate (`GET /quotes/{id}`) VÀ bảng đọc riêng (`order_summaries`)

**Là gì.** Tách **đường ghi** và **đường đọc** thành hai model khác nhau.

```
   GHI (Command)                    ĐỌC (Query)
   ────────────                     ───────────
   Aggregate                        Read model / view
   Bảo vệ invariant                 Chỉ để hiển thị
   Chuẩn hoá                        Bẹt, phi chuẩn hoá
   Một aggregate/transaction        Join thoải mái
```

**Vì sao Portage cần.** Vì hai nhu cầu này xung đột nhau:

| Nhu cầu | Model phù hợp |
|---|---|
| *"Trả cọc cho đơn X"* | Aggregate — cần khoá, cần bảo vệ luật |
| *"Hôm nay tôi đang ứng bao nhiêu tiền ngoài kia?"* | Read model — join 5 bảng, không cần khoá |

Câu hỏi thứ hai là **báo cáo vốn lưu động** — bạn thu cọc 50% nhưng phải trả
merchant 100%, nên luôn có tiền của bạn nằm ngoài. Nhồi câu truy vấn đó vào
aggregate sẽ làm hỏng aggregate.

**Đã code — bước đầu (05/09):** mọi endpoint ghi trả về **chỉ id** (`POST /quotes → 201 {"id"}`);
số liệu nằm ở `GET /quotes/{id}` trả `quoteView` — breakdown bẻ từng dòng, tiền là text
`{"amount":"150.00","currency":"USD"}` để client hiển thị, không cộng. Chưa có bảng đọc riêng:
adapter đọc `Quotes.ByID` rồi chiếu `Breakdown` ra view (`internal/adapter/http/quotes.go`).
Đó là CQRS *về hình dạng* (một đường ghi, một đường đọc, hai kiểu dữ liệu) chưa phải *về lưu trữ*.
Và một **projection thật** đã có ở chỗ khác: `pricing.Listing` là read model của pricing về
sản phẩm, do `Projector` dựng từ event — đúng nghĩa "model đọc được cập nhật bởi event".

**Đã code — read model THẬT (06/09, P9/T2):** `internal/app/reporting` +
bảng `order_summaries`. Đây mới là CQRS *về lưu trữ*:

```
   GHI                                          ĐỌC
   ───                                          ───
   catalog     products, merchants
   pricing     quotes, listings                 order_summaries
   ordering    orders            ──event──▶     một dòng phẳng mỗi đơn
   procurement purchase_tasks                   ghi CHỈ bằng event
   logistics   parcels, batches                 GET /me/orders · GET /orders?status=
```

Ba điều đáng nhớ, mỗi cái là một quyết định:

1. **Package `reportingapp` không có `internal/domain/reporting`.** Không phải
   quên: read model **không có invariant** nào để giữ, vì nó không quyết định
   gì. Nên `OrderSummary` là struct field exported hết, không method, không
   validate — giả vờ ngược lại là mời luật nghiệp vụ vào đúng chỗ không được
   có luật nào.

2. **Chỉ event được ghi vào nó.** Nó không đọc bảng của context nào. Đó là thứ
   ngăn "một màn hình" biến thành một câu JOIN qua năm bounded context và một
   database dùng chung không ai dám sửa.

3. **Idempotent VÀ chịu sai thứ tự.** Relay là at-least-once và không bảo đảm
   thứ tự giữa các luồng, nên `batch_shipped` có thể tới trước `order_placed`.
   Cách giải: **bất kỳ event nào cũng được TẠO dòng** (cái "khung"), và chỉ
   `order_placed` được đặt `status` đầu tiên — đoán một trạng thái là nói dối
   với khách đang nhìn màn hình. Chi tiết: WALKTHROUGH.md §20.

Câu hỏi "hôm nay tôi đang ứng bao nhiêu" giờ có hai bảng để trả lời:
`order_summaries` (đơn nào đang ở đâu) và `reconciliations` (§28, lời/lỗ từng
đơn). Cả hai đều là read model, cả hai đều dựng bằng event.


## 25. Eventual Consistency — Nhất quán sau cùng ✅ (có mã trạng thái, có test)

**Là gì.** Chấp nhận rằng hai aggregate **không đồng bộ tức thì**, mà sẽ khớp
nhau sau một lúc (qua event).

**Vì sao buộc phải chấp nhận.** Vì không có transaction nào bao được cả *"tiền
khách đã thu"* và *"đơn đã đặt ở Nike"*. Nike không tham gia transaction của bạn.

```
   t=0   Khách trả cọc          → CustomerOrder: DEPOSITED
   t=1   Event DepositPaid bay sang Procurement
   t=2   PurchaseTask được tạo  → PENDING
   t=3   Nhân viên đi mua       → CONFIRMED  hoặc  FAILED

   Trong khoảng t=0 → t=3, hệ thống KHÔNG nhất quán. Và điều đó CHẤP NHẬN ĐƯỢC.
   Cái không chấp nhận được là giả vờ nó nhất quán.
```

**Đã thấy bằng mắt (05/09):** `POST /products/{id}/publish → 204` rồi **ngay lập tức**
`POST /quotes → 404 listing_not_found`. Sản phẩm đã publish, nhưng pricing chưa nghe — worker
chưa relay. Chạy `relay.RunOnce` → `POST /quotes → 201`. Test vàng `wire_test.go wholeFlow`
bắn đúng ba lần để chứng minh khoảng không nhất quán này **tồn tại và kết thúc**. Mã lỗi
`listing_not_found` vì thế có hai nghĩa thật (chưa publish / chưa relay) — và tài liệu API nói
thẳng, thay vì hứa một thứ hệ thống không giữ được. In-memory, `cmd/api` chạy relay mỗi 200 ms
nên "khoảng" đó ngắn tới mức khó thấy — nhưng vẫn có.

## 26. Saga & Compensation ✅ nửa "bù trừ" (`CustomerOrder.Cancel`) — 📋 nửa điều phối

**Là gì.** Chuỗi các bước ở nhiều aggregate/context, và khi một bước hỏng thì
**bù trừ** các bước trước bằng **hành động nghiệp vụ**, không phải rollback DB.

**Vì sao không rollback được.** Bạn không thể "rollback" việc đã quẹt thẻ, cũng
không thể "rollback" đơn đã đặt ở Nike.

**Saga chính của Portage:**

```
  ┌──────────────┐
  │ Khách cọc 50%│
  └──────┬───────┘
         ▼
  ┌──────────────┐   thất bại   ┌────────────────────┐
  │ Đi mua ở Nike├─────────────▶│ BÙ TRỪ:            │
  └──────┬───────┘  (hết size)  │ hoàn 100% cọc      │
         │ thành công           │ hoặc đổi màu khác  │
         ▼                      └────────────────────┘
  ┌──────────────┐
  │ Về kho Mỹ    │
  └──────┬───────┘
         ▼
  ┌──────────────┐
  │ Gom lô, ship │
  └──────┬───────┘
         ▼
  ┌──────────────┐
  │ Khách trả nốt│
  └──────────────┘
```

**Điểm không thể quay đầu (point of no return)** — quy tắc quan trọng nhất của
mô hình cọc 50%:

```go
func (o *CustomerOrder) Cancel(reason CancelReason, now time.Time) (Refund, error) {
    switch o.status {
    case StatusAwaitingDeposit, StatusDeposited:
        return o.fullRefund(), nil       // CHƯA mua → hoàn 100%
    case StatusPurchased, StatusInTransit:
        return o.forfeitDeposit(), nil   // ĐÃ mua → mất cọc
    case StatusDelivered:
        return Refund{}, ErrAlreadyDelivered
    }
}
```

Sự kiện `PurchaseConfirmed` **chính là ranh giới đó**. Trước nó, rủi ro là của
bạn. Sau nó, rủi ro chuyển sang khách. Toàn bộ mô hình cọc 50% tồn tại vì ranh
giới này.

**Đã code (05/09, `internal/domain/ordering/order.go`)** — đoạn giả mã trên giờ là code thật,
thêm một trạng thái sơ đồ chưa có: `purchase_failed` (hết size) hoàn **đủ**, vì đó là lỗi
của mình. Bảng test `TestOrder_cancelRefundDependsOnThePointOfNoReturn`:

| Trạng thái khi huỷ | Hoàn | `Forfeited` | Vì |
|---|---|---|---|
| `awaiting_deposit` | 0 | false | chưa trả gì |
| `deposited` | **2 696 860** | false | chưa mua — tiền còn trong tay mình |
| `purchase_failed` | **2 696 860** | false | mình không mua được — lỗi của mình |
| `purchased` | 0 | **true** | đã ứng 100 % mua hàng cho khách |
| `in_transit` | 0 | **true** | hàng đang bay |
| `delivered` | — | — | `ErrAlreadyDelivered` |

`ConfirmPurchase` là method đổi `deposited → purchased` — điểm không thể quay đầu **là một
dòng code**.

**Vòng saga đã khép (05/09, chiều muộn):** `ordering.deposit_paid` → `procurementapp.OpenTaskHandler`
mở `PurchaseTask` và hỏi shop qua ACL (§22) → người mua đóng task → `procurement.purchase_confirmed`
→ `orderingapp.Reactor.OnPurchaseConfirmed` → `ConfirmPurchase` (đơn `purchased`), hoặc
`purchase_failed` → `FailPurchase` (đơn `purchase_failed`, huỷ sẽ hoàn **đủ**). Đọc `wire.Subscribe`
từ trên xuống chính là sơ đồ trên. Smoke thật 12 event: WALKTHROUGH.md §16h. Còn lại 📋: *ai* quyết
định hoàn hay đổi màu sau `purchase_failed` (hôm nay: khách/operator gọi `POST /orders/{id}/cancel`).

## 27. Outbox Pattern ✅ (bảng outbox, cùng transaction, worker at-least-once)

**Là gì.** Cách đảm bảo *"lưu dữ liệu"* và *"gửi event"* không bị lệch nhau.

**Vấn đề nó giải quyết:**

```
   ❌ Lưu DB thành công → chưa kịp gửi event → server chết
      → Đơn đã trả cọc nhưng KHÔNG AI đi mua. Khách mất tiền.
```

**Cách làm:** ghi event vào **cùng một bảng, cùng một transaction** với dữ liệu.
Một worker riêng đọc bảng đó và gửi đi.

```
   BEGIN TRANSACTION
     UPDATE customer_orders SET status = 'DEPOSITED' ...
     INSERT INTO outbox (event_type, payload) VALUES ('DepositPaid', ...)
   COMMIT                    ← hai việc, một transaction, không thể lệch

   [worker riêng]  đọc outbox → gửi → đánh dấu đã gửi
```

**Đã code, đủ hai nửa:**

| Nửa | Code | Test |
|---|---|---|
| Ghi cùng transaction | `postgres.Outbox.Append` dùng `db(ctx)` → cùng `tx` với `Save` | `TestOutbox_appendedInTheSameTransactionAndReadBack` |
| Save hỏng → không event | thứ tự trong handler | `…saveFailurePublishesNothing` |
| Save xong, outbox hỏng → **rollback** | `postgres.UnitOfWork.InTx` | `TestUnitOfWork_rollsBackTheSaveWhenTheOutboxFails` (chạy thật; skip với memory) |
| Worker đọc → publish → đánh dấu | `worker.Relay.RunOnce`, `cmd/worker` | 6 test relay: thứ tự, retry, batch |

**At-least-once** là lựa chọn: publish trước, `sent_at` sau. Sập giữa hai bước → gửi
lại. Không bao giờ mất; có thể thấy hai lần → **mọi subscriber phải idempotent**.
Ngược lại (đánh dấu trước) là mất event — một `DepositPaid` mất là một đơn không ai mua.
Payload là JSON viết tay theo hợp đồng (`eventcodec`, mục 7 Published Language).

## 28. Quote vs Actual — Ước tính và Thực tế ✅ cả hai nửa (`pricing.Quote` + `pricing.Reconciliation`)

**Đây là insight kinh tế lớn nhất của Portage**, và nó phải nằm trong model.

```
    Quote (báo giá)              Actual (thực tế)
    ═══════════════              ════════════════
    Ước tính                     Sự thật
    Trước khi mua                Sau khi cân, sau khi trả tiền
    CÓ HẠN dùng                  Vĩnh viễn
    Khách nhìn thấy              Kế toán nhìn thấy
              │                        │
              └────────┬───────────────┘
                       ▼
              CHÊNH LỆCH (Variance)
        = LỜI hoặc LỖ của bạn trên đơn đó
```

**Toàn bộ kinh tế học của nghề mua hộ nằm ở khoảng chênh này.** Ước tính nhẹ tay
→ lỗ. Ước tính nặng tay → mất khách.

> **Lỗi 90% project mắc phải:** lưu **một** con số "giá" rồi ghi đè khi biết số
> thật → **mất sạch dữ liệu để biết mình lời hay lỗ**.

Nên hệ thống phải ghi **cả hai** và đối soát. Trong DDD: `Quote` là aggregate
riêng, có `ExpiresAt`, và **chụp ảnh** mọi dữ kiện đã dùng — tỷ giá lúc đó,
phiên bản bảng giá lúc đó.

**Phần đã code — tầng dữ liệu đầu vào của Quote vs Actual** (`catalog`):

```
   CategoryPolicy.DefaultParcelSpec()      ParcelSpec ước lượng theo ngành    ← Pricing quote khi chưa đo
   Product.ParcelSpec()               ParcelSpec đo THẬT, kèm            ← Logistics ghi khi cân
   Product.ParcelProvenance()     ...ai đo, lúc nào, tin được không
```

Cùng một kiểu `ParcelSpec`, hai vai. Cái phân biệt ước tính với thực tế **không
phải kiểu dữ liệu** mà là `Provenance` đi kèm — và `Product.Publish()` chỉ cho
qua khi provenance của parcel là `Verified()`. Quote dựng trên draft là quote
biết mình đang đoán.

**Phần đã code — nửa Quote (05/09, `internal/domain/pricing/quote.go`):**

```go
type Quote struct {
	shared.Events
	id        QuoteID
	product   shared.ID
	lane      LaneCode
	breakdown Breakdown      // BỨC ẢNH: class, chargeable, Estimated, từng dòng USD, FX, từng dòng VND
	issuedAt  time.Time
	expiresAt time.Time      // policy.TTL — 48 h
	status    QuoteStatus    // issued → accepted | expired
}
```

Ba điều §28 đòi, giờ có bằng chứng:

| §28 đòi | Code |
|---|---|
| "chụp ảnh mọi dữ kiện đã dùng" | `Breakdown` giữ `FX`, `Chargeable`, `Class`, từng dòng tiền — đổi lane/policy/tỷ giá ngày mai, quote hôm nay **không đổi**; `quotes` trong Postgres lưu **từng cột** |
| "có hạn dùng" | `Accept(now)` sau `expiresAt` → chuyển `expired`, ghi event, **từ chối** — không lặng lẽ tính lại |
| "có hạn dùng" **cả khi không ai đụng tới** | `Expire(now)` do một **sweep theo giờ** gọi — `pricingapp.ExpireQuotesHandler`, `cmd/worker -sweep`. Không có sweep thì một quote không ai trả lời nằm `issued` **mãi mãi**: đúng luật khi đọc, sai khi đếm |
| "quote biết mình đang đoán" | `Breakdown.Estimated = true` khi cân bằng hộp mặc định của category thay vì hộp đã đo |

**Nửa Actual (06/09):** `pricing.Reconciliation` — một dòng mỗi đơn, tiền tuyến (USD), dựng từ
**ba** context: `ordering.order_placed` (đơn nào → quote nào; copy `QuotedGoods` = item + tax + duty,
`QuotedFreight` = freight + surcharge từ quote của chính pricing), `procurement.purchase_confirmed`
(`ActualGoods` = biên nhận shop), `logistics.batch_shipped` (`ActualFreight` = phần chia cước của
đơn này). `Variance() = quoted − actual`, chỉ khi `Complete()`.

```
quoted  163.22 + 25.00 = 188.22    (giày 150 + tax 13.22; freight 2.5 kg × 10)
actual  163.22 + 27.50 = 190.72    (shop tính đúng; carrier tính 27.50 thay 25.00)
variance                 −2.50 USD  ← mình chịu 2.50; hoặc phí dịch vụ 500 000 ₫ đã che
```

Không ghi đè: `quotes` giữ từng cột ước tính, `reconciliations` giữ cả hai bên. Đọc:
`GET /reconciliations/{order}`. Đây là câu hỏi §24 đặt ra ("hôm nay tôi đang ứng bao nhiêu",
"đơn này lời hay lỗ") có bảng để trả lời — WALKTHROUGH.md §17d, §17i.

**Sweep — loại use case thứ ba (06/09).** Cho tới T3, mọi use case trong hệ đều
có người đẩy: một request HTTP, hoặc một event của context khác. `Expire` thì
không:

| Cái gì đẩy | Ví dụ |
|---|---|
| một con người | `POST /orders` |
| một context khác | `deposit_paid` → mở purchase task |
| **thời gian trôi** | quote quá 48 h |

Không ai phát được event `quote_got_old`, vì **không có gì xảy ra cả** — chỉ là
không có gì xảy ra. Nên phải có người đi hỏi: `worker.Sweeper` gọi
`ExpireQuotesHandler.Handle` theo nhịp, và đó cũng là lý do rõ nhất để `Clock`
là **port** (§20): "quá hạn chưa" phải kiểm chứng được mà không đợi 48 giờ —
trong test, `clk.Advance(49 * time.Hour)`.

Sweep xử lý **mỗi quote một transaction**, không gói cả lô: lô 100 quote không
phải một invariant, nên nó không được là một đơn vị nguyên tử (§14). Chi tiết
từng dòng: WALKTHROUGH.md §19.

---

## 29. Vòng đời một AGGREGATE ✅

Đây là vòng đời **đã có code thật** trong `catalog/merchant.go`. Năm bước, và
thứ tự giữa bước 3 và 4 là chỗ dễ sai nhất.

```
 ①  SINH RA                    ②  THAY ĐỔI              ③  GHI SỰ KIỆN
 ─────────────                 ────────────             ──────────────
 RegisterMerchant(...)         m.Rename("X", now)       m.Record(ev)
 · kiểm mọi tham số            · kiểm invariant         · CHỈ ghi vào bộ nhớ
 · sinh ID ngay (UUID v7)      · từ chối được           · KHÔNG gửi đi đâu
 · KHÔNG chạm DB               · không có setter
        │                             │                        │
        └─────────────┬───────────────┴────────────────────────┘
                      ▼
 ④  LƯU                              ⑤  PHÁT SỰ KIỆN
 ──────                              ───────────────
 repo.Save(ctx, m)                   evs := m.PullEvents()
 · tầng adapter                      · tầng app, SAU khi ④ xong
 · SQL ở đây                         · ghi vào outbox CÙNG transaction với ④
```

> **④ phải xong trước ⑤.** Phát sự kiện trước khi lưu là cách worker đi mua giày
> cho một đơn **chưa từng tồn tại trong DB**. Xem mục 27 (Outbox).

### Nhìn bằng code thật

```go
// ① SINH RA — constructor là cửa duy nhất
m, err := catalog.RegisterMerchant(
    "Example Sports",
    catalog.MustParseHostname("example.com"),
    shared.USD,
    catalog.MustFreeShippingOver(shared.MustParseMoney("50.00", shared.USD)),
    []catalog.SourcingMode{catalog.SourcedByOperator},
    now,                       // ← đồng hồ truyền từ ngoài (quy ước 7)
)
// m.ID() đã có giá trị NGAY, chưa chạm database
```

```go
// ② + ③ — hành vi, không phải setter. Bên trong tự Record.
func (m *Merchant) Rename(name string, now time.Time) error {
    name = strings.TrimSpace(name)
    if name == "" {
        return ErrEmptyName              // ← invariant từ chối
    }
    if name == m.name {
        return nil                       // ← không có gì xảy ra → KHÔNG có event
    }
    old := m.name
    m.name = name
    m.Record(MerchantRenamed{ID: m.id, From: old, To: name, At: now})
    return nil
}
```

Chú ý dòng `if name == m.name { return nil }`. **Sự kiện chỉ sinh ra khi có
việc thật sự xảy ra.** Đổi tên thành chính nó không phải một sự kiện. Test
`TestMerchant_renameRaisesEventOnlyWhenItChanges` canh đúng chỗ này.

```go
// ⑤ — tầng app, sau khi lưu
evs := m.PullEvents()      // trả ra rồi XOÁ; pull lần hai được mảng rỗng
```

### Ba luật của vòng đời này

| Luật | Vì sao | Test canh |
|---|---|---|
| ID sinh trong domain, không đợi DB | object phải có danh tính từ giây đầu | `TestRegisterMerchant_hasIdentityImmediately` |
| Aggregate chỉ **ghi**, app mới **phát** | không publish thứ chưa lưu | `TestRegisterMerchant_recordsEventButDoesNotPublish` |
| Không có việc gì xảy ra → không có event | event là *sự thật đã xảy ra*, không phải *lệnh* | `TestMerchant_renameRaisesEventOnlyWhenItChanges` |

---

## 30. Vòng đời một REQUEST — đi hết các tầng ✅ 5/5 (auth → HTTP → app → domain → Postgres → worker)

**Cả 5 tầng (+ cửa auth) đã code và chạy thật**: `go run ./cmd/api -dsn …` + `go run ./cmd/worker
-dsn …` + `curl` → dòng outbox `sent_at` và log của worker. Transcript ở
WALKTHROUGH.md §1a (HTTP) và §7 (outbox + worker).
Sơ đồ dưới là hình dạng đã chốt; bản dẫn đọc code **từng file, từng dòng** nằm ở
[WALKTHROUGH.md](WALKTHROUGH.md) — đọc file đó để học flow, đọc mục này để nhớ
mỗi tầng làm gì và **không** làm gì.

```
 ┌────────────────────────────────────────────────────────────────────┐
 │  0. CỬA (auth)          internal/adapter/http/auth.go         ✅   │
 │     authenticate(Verifier) bọc CẢ mux — fail-closed                │
 │     · Authorization: Bearer <token> → auth.Principal{kind, id}     │
 │     · không token / token lạ → 401, KHÔNG vào tới tầng 1           │
 │     · requireOperator / requireCustomer / requireAny → 403         │
 │     · KHÔNG đọc header do người gọi tự khai (X-Operator-ID đã xoá) │
 └───────────────────────────────┬────────────────────────────────────┘
                                 ▼
 ┌────────────────────────────────────────────────────────────────────┐
 │  1. HTTP                        internal/adapter/http/        ✅   │
 │     POST /merchants · /products · /products/{id}/publish           │
 │     · giải mã JSON (DisallowUnknownFields)                         │
 │     · chuẩn hoá theo locale ("150,50" → "150.50")   ← QUAN TRỌNG   │
 │     · map lỗi → 400 / 404 / 409 bằng MỘT bảng                     │
 │     · KHÔNG có quy tắc nghiệp vụ nào ở đây                         │
 └───────────────────────────────┬────────────────────────────────────┘
                                 ▼
 ┌────────────────────────────────────────────────────────────────────┐
 │  2. USE CASE                    internal/app/catalog/              │
 │     RegisterMerchantHandler                                        │
 │     · lấy `now` từ Clock                                           │
 │     · MỞ transaction                                               │
 │     · gọi domain, gọi repository                                   │
 │     · điều phối — KHÔNG chứa quy tắc nghiệp vụ                     │
 └───────────────────────────────┬────────────────────────────────────┘
                                 ▼
 ┌────────────────────────────────────────────────────────────────────┐
 │  3. DOMAIN                      internal/domain/catalog/     ✅     │
 │     catalog.RegisterMerchant(...)                                  │
 │     · TOÀN BỘ quy tắc nghiệp vụ ở đây                              │
 │     · không biết HTTP, không biết SQL, không biết giờ              │
 └───────────────────────────────┬────────────────────────────────────┘
                                 ▼
 ┌────────────────────────────────────────────────────────────────────┐
 │  4. REPOSITORY                  internal/adapter/postgres/     ✅   │
 │     INSERT INTO merchants ... ON CONFLICT DO UPDATE  (qua Snapshot) │
 │     INSERT INTO outbox ...        ← CÙNG transaction (tx trong ctx) │
 │     COMMIT                                                          │
 └───────────────────────────────┬────────────────────────────────────┘
                                 ▼
 ┌────────────────────────────────────────────────────────────────────┐
 │  5. WORKER                      internal/worker + cmd/worker  ✅   │
 │     SELECT … WHERE sent_at IS NULL FOR UPDATE SKIP LOCKED           │
 │     publish theo thứ tự → UPDATE sent_at   (at-least-once)          │
 │     …và cạnh nó một Sweeper: việc không event nào đẩy, chỉ GIỜ đẩy  │
 │     (quote quá 48 h → Expire) — cmd/worker -sweep, §28              │
 └───────────────────────────────┬────────────────────────────────────┘
                                 ▼
              context khác nghe event và làm việc của nó
```

Tầng use case — **code thật**, chú ý **thứ tự** các dòng:

```go
// internal/app/catalog/register_merchant.go   ✅
func (h *RegisterMerchantHandler) Handle(ctx context.Context, d catalog.MerchantDetails) (catalog.MerchantID, error) {
	now := h.deps.Clock.Now()                          // đồng hồ ở TẦNG APP, đọc một lần

	var id catalog.MerchantID                          // Go không có method generic →
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {   // fn trả error, kết quả qua closure
		m, err := catalog.RegisterMerchant(d, now)                   // ③ domain quyết định
		if err != nil {
			return err                                               //   lỗi nghiệp vụ, trả nguyên vẹn
		}
		if err := h.deps.Merchants.Save(ctx, m); err != nil {        // ④ LƯU TRƯỚC
			return fmt.Errorf("save merchant: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, m.PullEvents()); err != nil {  // ⑤ RỒI MỚI pull
			return fmt.Errorf("outbox: %w", err)
		}
		id = m.ID()
		return nil
	})
	return id, err
}
```

Ba khác biệt so với bản phác cũ, đều học được từ việc viết thật:

- **Command của `RegisterMerchant` chính là `catalog.MerchantDetails`** — bọc
  thêm một struct trùng 100% field chỉ là copy đổi tên.
- **`InTx` không generic.** Go không cho method generic trên interface, nên fn
  trả `error` và kết quả (`id`) được gán qua biến closure.
- **`Deps` + `mustHave`**: mọi dependency vào một struct (quy ước 10); thiếu cái
  nào → panic lúc dựng handler (quy ước 1 — input của lập trình viên).

Use case giàu nhất là `AddProductHandler` (ghép clock + auth thành `Provenance`,
hai luật liên aggregate, phát hiện trùng) — mổ xẻ từng dòng ở WALKTHROUGH.md §3.

### Mỗi tầng làm gì — và tuyệt đối không làm gì

| Tầng | Làm | **Không** làm |
|---|---|---|
| `adapter/http/auth.go` ✅ | biết **ai** đang gọi, chặn ở cửa (401/403), dịch token → `Principal` | quyết định *cái gì* được làm với dữ liệu — đó là việc của domain |
| `adapter/http` ✅ | giải mã, chuẩn hoá locale, mã lỗi HTTP — gọi `Parse*` của domain, không validate riêng | quy tắc nghiệp vụ; biết repository/clock |
| `app` | mở transaction, lấy `now`, điều phối | quy tắc nghiệp vụ |
| `domain` | **toàn bộ** quy tắc nghiệp vụ | biết DB / HTTP / giờ |
| `adapter/postgres` ✅ | SQL, map snapshot ↔ cột, transaction | quyết định nghiệp vụ; phát event |
| `cmd/worker` ✅ | đọc outbox, gửi đi, đánh dấu; chạy sweep theo giờ | sửa dữ liệu nghiệp vụ; **quyết định** — sweep chỉ gọi use case, luật ở `Quote.Expire` |

> **Chỗ chuẩn hoá locale ở tầng 1 là thật, không phải hình thức.** `ParseMoney`
> từ chối `"150,50"` (mục 11d). Người Việt gõ dấu phẩy. Nếu adapter không đổi
> dấu phẩy thành dấu chấm, khách không đặt được hàng. Nếu domain tự đoán, nó sẽ
> đoán sai với một trong hai nhóm người dùng.

### So với Symfony

| Portage | Symfony |
|---|---|
| `adapter/http/` | Controller + `FormType` (chỗ `NumberFormatter` chạy) |
| `app/` | Service điều phối, `MessageHandler` |
| `domain/` | *(Symfony không có tầng này — logic nằm rải trong Service)* |
| `adapter/postgres/` | Doctrine Repository + mapping |
| outbox + worker | `postFlush` listener + Messenger |

---

## 31. Vòng đời một ĐƠN HÀNG — đi hết các context 📋

Vòng đời này **đã chạy hết bằng code** (06/09): một test vàng đi từ `POST /merchants` tới
`delivered` và `variance` trên cả memory và Postgres; smoke thật 20 event qua 5 context
(WALKTHROUGH.md §17h–§17i). Mỗi mũi tên là một dòng trong `wire.Subscribe`. Sơ đồ vẫn đúng
để thấy vì sao phải tách context.

```
  ┌─ CATALOG ✅ ────────────────────────────────────────────────┐
  │  Merchant, CategoryPolicy, Product                          │
  │  Khách chọn được món gì                                     │
  └──────────────────────────┬──────────────────────────────────┘
                             │ ProductPublished
  ┌─ PRICING ✅ ─────────────▼──────────────────────────────────┐
  │  Quote = giá hàng + phí cân ký + phí dịch vụ + tỷ giá       │
  │  CHỤP ẢNH tỷ giá & bảng giá  ·  CÓ HẠN dùng                 │
  └──────────────────────────┬──────────────────────────────────┘
                             │ QuoteIssued
  ┌─ ORDERING ✅ ────────────▼──────────────────────────────────┐
  │  CustomerOrder  ·  khách trả CỌC 50%  (POST /orders, /deposit)│
  └──────────────────────────┬──────────────────────────────────┘
                             │ DepositPaid
  ┌─ PROCUREMENT ✅ ─────────▼──────────────────────────────────┐
  │  PurchaseTask — nhân viên đi mua thật ở web Mỹ (ACL Manual) │
  │  v1: 1 đơn → 1 việc mua; GET /purchase-tasks = việc cần làm │
  └──────────────────────────┬──────────────────────────────────┘
              ┌──────────────┴──────────────┐
              ▼                             ▼
      PurchaseConfirmed              PurchaseFailed (hết size)
              │                             │
              │                             ▼
              │                    ┌─ SAGA BÙ TRỪ ──────────┐
              │                    │ hoàn 100% cọc          │
              │                    │ hoặc đổi màu/size khác │
              │                    └────────────────────────┘
              ▼
  ╔═══════════════════════════════════════════════════════════╗
  ║  ⚠️  ĐIỂM KHÔNG THỂ QUAY ĐẦU                              ║
  ║  Trước đây: khách huỷ → hoàn 100% cọc                     ║
  ║  Từ đây:    khách huỷ → MẤT cọc (mình đang giữ hàng)      ║
  ╚═══════════════════════════════════════════════════════════╝
              │
  ┌─ LOGISTICS ✅ ───────────▼──────────────────────────────────┐
  │  Parcel về kho Mỹ → CÂN THẬT  (POST /parcels/{id}/receive)  │
  │  ConsolidationBatch: gom lô  (POST /batches …/parcels …/close)│
  │  Chia cước cả thùng cho từng đơn  ← FreightAllocator ✅      │
  └──────────────────────────┬──────────────────────────────────┘
                             │ BatchShipped
  ┌─ ORDERING ✅ ────────────▼──────────────────────────────────┐
  │  Khách trả nốt 50%  ·  giao tận nhà  ·  OrderDelivered ✅   │
  └──────────────────────────┬──────────────────────────────────┘
                             ▼
  ┌─ ĐỐI SOÁT ✅ (pricing.Reconciliation) ────────────────────────────────────────────────┐
  │  Quote (ước tính)  ─  Actual (thực tế)  =  LỜI hay LỖ       │
  │  GET /reconciliations/{order} → variance −2.50 USD          │
  └────────────────────────────────────────────────────────────┘
```

### Ba chỗ vòng đời này dạy điều mà sơ đồ tầng không dạy

**1. Vì sao `CustomerOrder` và `PurchaseTask` phải tách** — nhìn nhánh
`PurchaseFailed`. Một đơn khách sinh N việc mua; việc mua thứ 2 hỏng trong khi
tiền khách đã thu. Hai thứ đó **không thể** chung một transaction, vì Nike không
tham gia transaction của mình (mục 25, 26).

**2. Vì sao `Quote` phải chụp ảnh** — giữa lúc báo giá và lúc cân thật có **4
tuần**. Tỷ giá trôi, bảng giá đổi. Nếu `Quote` trỏ tới "tỷ giá hiện tại", giá
của khách tự nhảy sau lưng họ (mục 11, 28).

**3. Vì sao cần đối soát ở cuối** — cước **ước tính** lúc báo giá và cước
**thật** sau khi cân thùng là hai con số khác nhau. Chênh lệch đó chính là lời
hoặc lỗ. Ghi đè số ước tính bằng số thật = **mất sạch dữ liệu để biết mình có
lãi hay không** (mục 28).

---

# PHẦN V — Những cái bẫy

| Bẫy | Dấu hiệu | Cách tránh |
|---|---|---|
| **Anemic model** | class chỉ có getter/setter | đặt hành vi vào trong object |
| **God aggregate** | một aggregate 30 trường, ai cũng sửa | tách theo invariant, càng nhỏ càng tốt |
| **Primitive obsession** | `float64` cho tiền, `string` cho trạng thái | value object có kiểu riêng |
| **Framework rò vào domain** | `import "gorm.io/gorm"` trong domain | interface ở domain, cài đặt ở adapter |
| **Một context cho tất cả** | `Product` có 40 cột | tách bounded context |
| **Repository quá thông minh** | `findByAAndBAndC(...)` | truy vấn phức tạp → read model |
| **Event thì hiện tại** | `CreateOrder` | luôn quá khứ: `OrderCreated` |
| **DDD cho CRUD** | viết 200 dòng để insert một dòng | dùng DDD đúng chỗ đáng dùng |
| **Ghi đè ước tính bằng số thật** | chỉ có một cột `price` | lưu cả Quote và Actual |
| **Domain đoán locale** | `ParseMoney` bỏ dấu phẩy → `"150,50"` = 15.050 | nhận một định dạng chuẩn; adapter chuẩn hoá |
| **Phép chia mặc định trong chuỗi tính phí** | `mm3 / divisor` cắt xuống giữa chừng | luật làm tròn áp vào **mọi** bước |
| **Zero value coi là hợp lệ** (riêng Go) | `Money{}` cộng được, `Currency{}` in ra `""` | `IsZero()`/`IsValid()`; constructor từ chối |
| **Tái dùng kiểu vì "cùng là số"** | `ExchangeRate` dùng parser của `Rate` (4 số lẻ) | mỗi khái niệm một VO, ràng buộc riêng |
| **Shared kernel gánh việc của một context** | `Source`/`AsOf` của Pricing nằm trong `shared` | hỏi: *context nào KHÔNG cần field này?* |
| **Hai nguồn sự thật cho một con số** | `volumetricDiv` vừa ở `CategoryPolicy` vừa ở `ShippingLane` | hỏi: *dữ liệu này AI sở hữu?* — một chủ, một chỗ |
| **Identity là `type X string`** | `CategoryCode("Giày Dép!")` compile được | khoá tự nhiên = struct field private + `Parse*`; `type X string` chỉ cho enum đóng |
| **Entity chỉ có chiều bật** | `EnableSourcing` mà không `DisableSourcing`; không `Suspend` | mỗi trạng thái có đường vào **và** đường ra |

---

# PHẦN VI — Trạng thái hiện tại của Portage

## Đã code ✅

```
internal/
├── contracts/                   catalog_v1.go — *V1 DTO                          ← PUBLISHED LANGUAGE (05/09)
├── app/
│   ├── ports.go / deps.go       Clock, UnitOfWork, Outbox; MustHave              ← PORT tầng app
│   ├── catalog/  (catalogapp)   7 USE CASE: RegisterMerchant, AddProduct, PublishProduct, DefineCategory,
│   │                            AddVariant, ConfirmListing, MeasureProduct
│   ├── pricing/  (pricingapp)   IssueQuote, AcceptQuote, ExpireQuotes (SWEEP — không ai gọi, chỉ thời gian), Projector, Reconciler ← context thứ hai (05/09)
│   ├── ordering/ (orderingapp)  PlaceOrder, PayDeposit/PayBalance, Cancel, ConfirmPurchase/FailPurchase/Ship/Deliver,
│   │                            Projector (nghe pricing), Reactor (nghe procurement); Deps.mutate ← context thứ ba (05/09)
│   ├── procurement/ (procurementapp)  OpenTask (hỏi ACL, 4 nhánh), ConfirmTask, FailTask, Projector (nghe catalog) ← context thứ tư (05/09)
│   ├── logistics/ (logisticsapp)  ExpectParcel (nghe procurement), ReceiveParcel, Open/Add/Close/ShipBatch, Projector (nghe pricing) ← context thứ năm (06/09)
│   └── reporting/ (reportingapp)  READ MODEL: Projector nghe CẢ NĂM context → order_summaries; KHÔNG có domain (§24) ← 06/09
├── adapter/merchant/            Manual (một con người) + Router (chọn theo shop)  ← ANTI-CORRUPTION LAYER #1 (§22)
├── adapter/openai/              Client / Fake / Unavailable → ListingDraft         ← ANTI-CORRUPTION LAYER #2 (§22)
├── adapter/http/ (httpapi)      NewHandler(Deps{5 context + Reporting, Auth/Tokens/Registry, Extractor}): 37 route SAU middleware auth, decode+locale, errorTable, quote/order/task/parcel/batch/reconciliation view ← ADAPTER HTTP
│   └── auth.go / tokens.go      authenticate() bọc cả mux (401), requireOperator/Customer/Any (403), POST /tokens ← CỬA (06/09)
├── adapter/eventcodec/          Encode(event) → JSON hợp đồng; Decode → V1; guard go/ast mọi context
├── adapter/postgres/            repo qua Snapshot, UnitOfWork tx-trong-ctx, Outbox, 0001 … 0007, repo 5 context + reconciliations + api_tokens + order_summaries ← ADAPTER Postgres
├── adapter/memory/              repo catalog + pricing, Outbox, UnitOfWork        ← ADAPTER in-memory
├── worker/                      Relay (at-least-once) + Bus + Sweeper (việc theo GIỜ)  ← tầng 5
├── platform/auth/               Principal, port Verifier/Issuer, Static; HashToken (sha256) ← "ai đang gọi" là việc của BIÊN (§20)
├── platform/clock/              System (time.Now), Fixed (test)         ← ADAPTER cho Clock
├── platform/wire/               Memory() / Postgres() → Graph{6 context, Source, Auth/Tokens/Registry, Extractor}; Subscribe() = 31 dòng định tuyến event
│
cmd/api/main.go                  -dsn → wire.Memory (+ relay trong process) | wire.Postgres → :8080
cmd/worker/main.go               -dsn (bắt buộc) → wire.Postgres → wire.Subscribe → Relay.Run
scripts/smoke.ps1                cả flow trên binary thật + Postgres thật
│
└── domain/
├── decisions_test.go            ← 7 TEST CANH QUYẾT ĐỊNH (xem SETUP.md)
│
├── shared/                      Shared Kernel
│   ├── operator.go              OperatorID                          ← danh tính liên context
│   ├── parcelspec.go            ParcelSpec — cân + hộp (dời từ catalog 05/09) ← Value Object hai context dùng
│   ├── money.go                 Money                               ← Value Object
│   ├── currency.go              Currency, CurrencyFromCode          ← Value Object
│   ├── rate.go                  Rate, ExchangeRate                  ← Value Object
│   ├── weight.go                Weight, Dimensions, ChargeableWeight← Value Object
│   ├── decimal.go               parseDecimal, addExact, divRoundHalfUp (private)
│   ├── id.go                    ID (UUID v7)                        ← hạ tầng Entity
│   └── event.go                 Event, Events                       ← hạ tầng Event
│
└── catalog/                     Bounded Context đầu tiên
    ├── merchant.go              Merchant ← AGGREGATE ROOT đầu tiên, có vòng đời
    │                            MerchantID, MerchantStatus, SourcingMode, Hostname
    ├── freeshipping.go          FreeShipping                        ← Value Object
    ├── category.go              CategoryPolicy ← VO cấu hình KHOÁ TỰ NHIÊN
    │                            CategoryCode (struct), Restriction
    ├── provenance.go            Provenance — nguồn, lúc nào, ai         ← Value Object
    ├── sourceurl.go             SourceURL — link tham chiếu, Host()     ← Value Object
    ├── product.go               Product ← AGGREGATE ROOT thứ hai, có entity con
    │                            ProductID, ProductStatus, ProductDetails
    ├── variant.go               Variant ← ENTITY CON; VariantID, VariantDetails; khoá chống trùng
    │                            bỏ MỌI khoảng trắng; unnamed() → "không tên phải là duy nhất"
    ├── events.go                16 event: 7 Merchant* + Product Added/Measured/Repriced/
    │                            FlaggedDuplicate/DuplicateCleared/Published(+Price,Parcel,Source)/Retired
    │                            + CategoryDefined + VariantAdded (entity con cũng kể được)
    ├── repository.go            Merchant/Category/ProductRepository   ← PORT
    └── snapshot.go              Snapshot()/FromSnapshot() — cửa sau cho persistence

└── pricing/                     Bounded Context thứ hai (05/09)
    ├── lane.go                  LaneCode (khoá tự nhiên), GoodsClass, RateCard, DutyPolicy,
    │                            LaneDetails, ShippingLane ← VO bảng giá, Classification
    ├── policy.go                MarginPolicy, QuotePolicyDetails, QuotePolicy ← cấu hình nghiệp vụ chụp vào quote
    ├── listing.go               Listing, CategoryProfile          ← PROJECTION từ event của catalog
    ├── calc.go                  QuoteInputs, Breakdown, Calculate ← DOMAIN SERVICE (hàm thuần)
    ├── quote.go                 Quote ← AGGREGATE ROOT thứ ba: ảnh chụp có hạn; Accept/Expire; QuoteSnapshot
    ├── events.go                QuoteIssued/Accepted/Expired
    └── repository.go            Lane/Quote/Listing/CategoryProfileRepository, ExchangeRates ← 5 PORT

└── ordering/                    Bounded Context thứ ba (05/09)
    ├── order.go                 CustomerOrder ← AGGREGATE ROOT thứ tư: state machine 7 trạng thái, Refund VO,
    │                            Cancel = điểm không thể quay đầu (§26), OrderSnapshot
    ├── acceptedquote.go         AcceptedQuote                    ← PROJECTION từ pricing.quote_accepted
    ├── variant.go               Variant{Variant, Product}        ← PROJECTION từ catalog.variant_added,
    │                            mỏng nhất có thể: KHÔNG giữ size, vì luật của ordering không cần
    ├── events.go                8 event; DepositPaid mang product + variant cho procurement
    └── repository.go            OrderRepository (ByID, ByQuote), AcceptedQuoteRepository, VariantRepository ← PORT

└── procurement/                 Bounded Context thứ tư (05/09)
    ├── task.go                  PurchaseTask ← AGGREGATE ROOT thứ năm: open → confirmed | failed;
    │                            PurchaseReceipt = ACTUAL đầu tiên (Quote vs Actual §28)
    │                            Subject = mua gì, bằng CHỮ, chép vào task lúc mở
    ├── events.go                PurchaseTaskOpened, PurchaseConfirmed, PurchaseFailed
    └── repository.go            Shop/Item/Variant (projection của catalog, Variant.Label());
                                 TaskRepository, VariantRepository; MerchantACL ← PORT + ACL

└── logistics/                   Bounded Context thứ năm (06/09)
    ├── parcel.go                Parcel ← AGGREGATE ROOT: expected → received (CÂN THẬT) → batched → shipped
    ├── batch.go                 ConsolidationBatch ← AGGREGATE ROOT: open → closed → shipped; Ship(freight, allocator)
    ├── allocator.go             FreightAllocator ← DOMAIN SERVICE (interface); ByChargeableWeight; LaneRule (projection)
    ├── events.go                ParcelExpected/Received, BatchOpened/Closed/Shipped{Allocations}
    └── repository.go            ParcelRepository (Pending), BatchRepository, LaneRuleRepository ← PORT
    (pricing/reconciliation.go   Reconciliation — Quote vs Actual per order, Variance() — projection 3 nguồn)
```

```
go test ./...   →  296 test (08/09, + hai route tham chiếu cho UI) — 31 tích hợp chạy khi có PORTAGE_TEST_DSN; 265 còn lại < 2 giây không cần Docker
coverage        →  shared 91% · catalog 89% · pricing 74% · ordering 81% · procurement 81% · logistics 76% · app 71–82% · http 81% · codec 99% · postgres 80% · memory 81% · openai 81% · merchant 89% · auth 89% · wire 95% · worker 87%
go run          →  cmd/api (in-memory hoặc -dsn) · cmd/worker -dsn · docker compose up -d · scripts/smoke.ps1
gofmt / vet     →  sạch
CI              →  gofmt, vet, test -race, dependency rule
```

**Khái niệm DDD đã thật sự hiện diện trong code:**

| Khái niệm | Bằng chứng |
|---|---|
| Value Object | `Money`, `Weight`, `Rate`, `ExchangeRate`, `Dimensions`, `FreeShipping`, `Hostname` |
| Bất biến | method trả giá trị mới, không sửa tại chỗ (trừ `Events`, cố ý) |
| So sánh theo giá trị | `ExchangeRate` chuẩn hoá để `"26000.50" == "26000.5"` |
| **Entity** | `Merchant` (UUID v7, vòng đời `active ⇄ suspended`); `CategoryPolicy` dùng **khoá tự nhiên** `CategoryCode` — không phải danh tính nào cũng cần UUID |
| **Aggregate Root** | `Merchant` — field private, không setter, đổi qua `Rename` / `Enable`+`DisableSourcing` / `ChangeFreeShipping` / `Suspend`+`Reinstate` |
| **Repository là PORT** | `MerchantRepository`, `CategoryRepository`, `ProductRepository` — interface ở domain, `Save` không publish |
| **Invariant nhiều field** | `NewParcelSpec`: cân nặng **và** đủ 3 cạnh; `Product.Publish`: 5 điều kiện, mỗi cái một sentinel |
| **Aggregate có entity con** | `Product` ⊃ `Variant` — không `*Variant` rò ra, không `VariantRepository` |
| **Provenance theo nhóm thuộc tính** | 3 `Provenance` trên `Product`; chỉ `operator` là `Verified()` |
| **Vòng đời có trạng thái cuối** | `Product`: draft → published → retired; retired giữ lại cho đơn đã tham chiếu |
| Invariant trong constructor | `RegisterMerchant` từ chối tên rỗng, hostname sai, currency lệch; `NewWeight` từ chối âm |
| **Domain Event nghiệp vụ** | 14 event, thì quá khứ, wire name ổn định — `MerchantSuspended` (Procurement nghe), `ProductPublished`/`ProductMeasured` (Pricing nghe) |
| **Event chỉ ghi, không phát** | `Record` trong aggregate, `PullEvents` ở tầng app sau khi lưu |
| **Không có việc → không có event** | `Rename` về đúng tên cũ: không sinh event |
| Ubiquitous Language | `ChargeableWeight`, `VolumetricWeight`, `FreeShipping` |
| Primitive obsession (tránh được) | không `float64` cho tiền; `Hostname` thay `string`; `parseDecimal` nghiêm ngặt |
| Shared Kernel nhỏ | đã **thu nhỏ** (bỏ `Source`/`AsOf` khỏi `ExchangeRate`) |
| Dependency Rule | domain chỉ stdlib + allowlist — **test canh, không chỉ CI** |
| Bounded Context tách thật | `catalog` không import context khác — **test canh** |
| Đa thương hiệu = dữ liệu | không có tên hãng nào trong code — **test canh** |
| Bẫy riêng của Go được xử lý | slice trả ra là **bản sao** (`Sourcing()`, `Restrictions()`) |
| Test là tài liệu | áo phao 0,9kg → 4,7kg; `"150,50"` không phải 15.050 |
| **Quyết định là tài liệu chạy được** | `decisions_test.go` — 7 guard, đã kiểm chứng từng cái sập bẫy |
| **Application Service / Use case** | `catalogapp.AddProductHandler` — ghép clock + auth → `Provenance`, hai luật liên aggregate, không một quy tắc nghiệp vụ nào khác |
| **Port ở nơi cần, adapter ở ngoài** | `app.Clock`/`UnitOfWork`/`Outbox` khai báo ở app; `memory`, `platform/clock` cài; `var _ Port = (*Adapter)(nil)` kiểm compile-time |
| **Save → Pull → Outbox là hành vi, không phải comment** | `TestRegisterMerchant_saveFailurePublishesNothing` |
| **Wiring là input của lập trình viên** | `Deps` + `mustHave` panic lúc dựng handler; `cmd/api/main.go` điền cùng struct đó |
| **Adapter HTTP làm đúng ba việc** | decode (`DisallowUnknownFields`) · locale (`normalizeAmount`) · map lỗi (`errorTable`); validate là `Parse*` của domain |
| **400 / 404 / 409 là ba câu trả lời khác nhau** | không hiểu request · không có thứ đó · nghiệp vụ từ chối — một bảng, `errors.Is` |
| **Memento cho persistence** | `Snapshot()`/`FromSnapshot()` — field vẫn private, repository vẫn dựng lại được; round trip test |
| **Repository thật + transaction thật** | `postgres.*Repo` upsert, `UnitOfWork.InTx` tx-trong-ctx; test rollback chạy thật |
| **Outbox đủ hai nửa** | ghi cùng transaction · worker at-least-once · payload là hợp đồng viết tay |
| **Published Language** | `internal/contracts` (*V1 DTO, package riêng) + `eventcodec` Encode/Decode — 21 event decode được, key viết tay, guard `go/ast` quét mọi `domain/*/events.go` |
| **Hai context nói chuyện qua event** | `catalog.ProductPublished` → outbox → relay → `wire.Subscribe` → `pricingapp.Projector` → `listings`; pricing **không import, không FK** sang catalog |
| **Projection / Customer giữ bản sao** | `pricing.Listing`, `CategoryProfile` — idempotent, chịu sai thứ tự, từ vựng riêng (`Classification`) — mỗi tính chất một test |
| **Cùng một event, hai mức chép khác nhau** | `catalog.variant_added` → `procurement.Variant` giữ 5 field (có người phải đọc size) nhưng `ordering.Variant` giữ 2 (luật chỉ cần "có thật? của ai?"). Projection giữ dữ liệu **không ai đọc** sẽ mốc mà không test nào bắt |
| **Chép vs tra lại** | `PurchaseTask.Subject` chép bốn chữ lúc mở việc: việc đã giao cho người thì không đổi nội dung, và bảng chiếu là eventual nên tra lúc đọc có thể trống |
| **Domain Service** | `pricing.Calculate` — hàm thuần, kiểm được bằng bảng tính; quyết định duy nhất lộ ra `Estimated` |
| **Quote là ảnh chụp có hạn** | `Quote{Breakdown, expiresAt}`; `Accept` trễ → expired + event + từ chối; Postgres lưu breakdown từng cột |
| **Cấu hình nghiệp vụ là value object** | `QuotePolicy`, `Classification` dựng ở `wire`, nằm trong `Deps`, panic khi zero |
| **Eventual consistency có mã trạng thái** | `listing_not_found` sau publish, trước relay — test vàng `wholeFlow` chứng minh khoảng không nhất quán tồn tại và kết thúc |
| **Read model đầu tiên** | `GET /quotes/{id}` → `quoteView`, tiền là text; ghi vẫn chỉ trả id |
| **Aggregate vừa đổi vừa từ chối** | `AcceptQuoteHandler` commit expiry rồi trả `ErrQuoteExpired` |
| **Saga — nửa bù trừ** | `CustomerOrder.Cancel`: 5 trạng thái → 5 khoản hoàn; `ConfirmPurchase` là điểm không thể quay đầu — bảng test |
| **State machine là code domain** | `CustomerOrder` 8 method có tên nghiệp vụ; `refuse()` trả lý do thật; snapshot từ chối 8 hình dạng không thể có |
| **Luật liên aggregate hai nửa** | "một quote một đơn": `Orders.ByQuote` ở app + `orders.quote UNIQUE` ở DB |
| **Cạnh thứ hai của Context Map** | `pricing.quote_accepted` → `contracts.QuoteAcceptedV1` → `ordering.AcceptedQuote`; ordering không tính lại giá |
| **Anti-Corruption Layer** | `procurement.MerchantACL` (port, từ vựng của mình) + `merchant.Manual` (adapter #1: một con người); use case xử lý 4 câu trả lời, fake ACL chứng minh |
| **Saga khép kín** | `deposit_paid` → `PurchaseTask` → `purchase_confirmed/failed` → `ordering.Reactor` → `purchased` / `purchase_failed`; smoke 12 event qua 4 context |
| **"Đã làm rồi" ≠ "không làm được"** | `Reactor.alreadyDone` nuốt `ErrNotDeposited` cho event tới lần hai — chỗ duy nhất biết phân biệt |
| **Actual đầu tiên của Quote vs Actual** | `PurchaseReceipt.Paid` (163.22 USD thật) cạnh `Quote.Breakdown.ItemPrice + SalesTax` (163.22 ước tính) — hai số, hai bảng, không ghi đè |
| **Domain Service là interface** | `logistics.FreightAllocator` + `ByChargeableWeight`; `ConsolidationBatch.Ship` nhận allocator làm tham số |
| **Tiền chia đúng từng cent** | `shared.Allocate` — floor + largest remainder, `big.Int`; 87.50 → 29.17 + 58.33 |
| **Quote vs Actual khép** | `pricing.Reconciliation` từ 3 context; `Variance() = −2.50 USD` trên đơn thật; `GET /reconciliations/{order}` |
| **Vòng đời §31 chạy hết** | test vàng `wholeFlow` tới `delivered`; smoke 20 event qua 5 context; `Subscribe` 31 dòng = sơ đồ §31 |
| **Hai aggregate một transaction, có lý do** | `logisticsapp.AddParcelHandler`/`ShipBatchHandler` — parcel status là sổ sách của batch |

## Chưa code 🔜

**P9 XONG (06/09)** — cả tám task, xem [P9-PLAN.md](P9-PLAN.md) cho plan gốc và
[FLOW-ORDER.md](FLOW-ORDER.md) cho kết quả: T1 auth (§20, §30, WALKTHROUGH §18) · T2 read model
(§24, WT §20) · T3 quét quote (§28, WT §19) · T4 `POST /fx` (WT §21g) · T5 `adapter/openai`
(§22, WT §21) · T6 ACL Router (§22) · T7 `FLOW-ORDER.md` · T8 dọn dẹp.

**Chưa code, sau P9:** `config/ratecard.yaml` (bảng giá thật) · một `adapter/merchant/<shop>.go`
thật (chưa shop nào cho API) · "ai quyết định" sau `purchase_failed` (v1: con người) · cổng thanh
toán thật (webhook → `PayDeposit`/`PayBalance`) · thông báo khách · identity context thật thay `Static`.

> **Nói cho đúng khi phỏng vấn:** project có **Value Object, Shared Kernel, bảy Aggregate Root
> (một có entity con) với 36 Domain Event, hai Domain Service (hàm thuần `Calculate`, interface
> `FreightAllocator`), HAI Anti-Corruption Layer (API shop, và AI đọc trang web), hơn bốn mươi use case
> ở tầng app đi qua Port/Adapter thật (in-memory và Postgres), và một bộ test canh kiến trúc**. **Năm
> bounded context nối nhau bằng 36 event qua outbox và một Published Language**, mỗi bên nghe giữ projection
> idempotent; **vòng đời một đơn chạy hết**: dán link → báo giá → cọc → đi mua → điểm không thể quay đầu →
> cân thật → gom lô → chia cước đúng từng cent → giao → **đối soát Quote vs Actual** — 20 event qua 5
> context, trên cả memory và Postgres bằng cùng một test. **Auth thật**: bearer token, port `Verifier`/`Issuer`
> ở `platform/` chứ không ở domain, middleware bọc cả mux (fail-closed), 401 ≠ 403, và chủ đơn sai → 404 để
> không xác nhận đơn có thật. **Read model thật**: `order_summaries` dựng từ event của cả năm context,
> ghi CHỈ bằng event, idempotent và chịu sai thứ tự — CQRS về lưu trữ, không chỉ về hình dạng. **Hai
> Anti-Corruption Layer**: một cho API shop, một cho mô hình ngôn ngữ đọc trang web — và cái thứ hai
> chỉ được TẠO nháp, không được xác nhận, nên không gì AI bịa ra thành được giá bán. Chưa có ACL API
> thật của một shop, chưa có cổng thanh
> toán thật. Nói đúng phần đã làm — [FLOW-ORDER.md](FLOW-ORDER.md) đi hết một đơn, [WALKTHROUGH.md](WALKTHROUGH.md) §14–§21 dẫn từ `Product.Publish`
> tới `variance = −2.50 USD`.

## Lộ trình tiếp theo

```
[x] catalog/     Merchant, CategoryPolicy, ParcelSpec, repository interface → Entity, đa thương hiệu, Port
[x] catalog/     Product, Variant, Provenance, SourceURL                     → Aggregate có entity con, invariant nhiều điều kiện
[x] pricing/     ShippingLane, RateCard, Quote, Calculate → Aggregate ảnh chụp, Domain Service, Quote vs Actual (nửa Quote)
[x] contracts + Projector                   → Published Language, Customer/Supplier, projection idempotent, eventual consistency
[x] ordering/    CustomerOrder, Refund, Cancel → Saga (nửa bù trừ), state machine, luật cọc 50 %, nghe quote_accepted
[x] procurement/ PurchaseTask, MerchantACL  → ACL (port + Manual + fake), saga khép kín, "đã làm rồi" ≠ "không làm được"
[x] logistics/   Parcel, Batch, FreightAllocator, shared.Allocate → Domain Service là interface; chia tiền đúng cent; Quote vs Actual khép
[x] app/catalog + adapter/memory            → Application Service, Port/Adapter, Save→Pull→Outbox
[x] adapter/http + cmd/api                  → chuẩn hoá locale, map lỗi → 400/404/409, chạy thật
[x] adapter/postgres + snapshot + wire      → Repository thật, Outbox thật, transaction thật
[x] worker + cmd/worker                     → at-least-once, Bus; wire.Subscribe = bảng định tuyến
[x] 0002_pricing.sql + wire.Graph + /quotes → hai context trên một outbox, một test vàng cho cả hệ
[ ] adapter/openai                          → học Ports & Adapters với AI
```

---

# PHẦN VII — Bảng tra: Symfony ↔ Go ↔ DDD

| Symfony / Doctrine | Portage (Go) | Khái niệm DDD |
|---|---|---|
| `#[ORM\Entity] class X` | struct trong `internal/domain/` | Entity / Aggregate Root |
| `#[ORM\Embeddable]` | struct bất biến, không ID | Value Object |
| `XRepository extends ServiceEntityRepository` | interface ở domain + cài đặt ở adapter | Repository (đảo chiều phụ thuộc) |
| `EntityManager::flush()` | `repo.Save(ctx, aggregate)` | ranh giới transaction |
| `services.yaml`, autowire | truyền tham số vào constructor | Dependency Injection thủ công |
| `EventSubscriber` | `[]DomainEvent` + outbox | Domain Event |
| Validator constraints | kiểm tra trong constructor | Invariant |
| `Controller` | `internal/adapter/http/` | Adapter |
| Service điều phối | `internal/app/` | Application Service |
| Service chứa luật | `internal/domain/*/` | Domain Service |
| `DECIMAL` trả về string | `Money{minor int64}` | tránh float |
| `ClockInterface` | truyền `now time.Time` vào method | domain không đọc đồng hồ |
| Bundle | Bounded Context | Bounded Context |
| `phpunit` | `go test ./...` | — |
| `symfony/uid` `Uuid::v7()` | `shared.NewID()` | Entity identity |
| `final class OrderId { Uuid $value }` | `type OrderID struct{ shared.ID }` | typed ID |
| `$domainEvents[]` + `pullDomainEvents()` | `shared.Events` — `Record` / `PullEvents` | Domain Event |
| `\DomainException` (bắt được) | `return error` + `errors.Is` | vi phạm nghiệp vụ |
| `\LogicException` (bug, không bắt) | `panic` | lỗi lập trình |
| `Brick\Math\BigDecimal::of($s)` | `parseDecimal` (nghiêm ngặt) | tránh float |
| `deptrac` | `go list -f '{{.Imports}}'` trong CI | Dependency Rule |
| `composer.lock` | `go.sum` | — |
| `MessageHandler::__invoke(Command)` | `catalogapp.XHandler.Handle(ctx, cmd)` | Application Service |
| `ClockInterface` / `MockClock` | `app.Clock` / `clock.Fixed` | domain không đọc giờ |
| `$em->wrapInTransaction(fn)` | `app.UnitOfWork.InTx(ctx, fn)` | ranh giới transaction |
| autowiring / constructor injection | `catalogapp.Deps` điền tay + `mustHave` | Dependency Injection thủ công |
| `InMemoryRepository` trong `tests/` | `internal/adapter/memory` — adapter bình đẳng | Port/Adapter |

---

# PHẦN VIII — Từ điển Anh–Việt

| Tiếng Anh | Tiếng Việt | Nghĩa ngắn |
|---|---|---|
| Domain | miền nghiệp vụ | lĩnh vực kinh doanh đang mô hình hoá |
| Subdomain | miền con | một mảng của domain |
| Ubiquitous Language | ngôn ngữ chung | bộ từ vựng dùng thống nhất mọi nơi |
| Bounded Context | ngữ cảnh giới hạn | ranh giới mà mỗi từ có một nghĩa |
| Context Map | bản đồ ngữ cảnh | quan hệ giữa các context |
| Entity | thực thể | object có danh tính |
| Value Object | đối tượng giá trị | object chỉ có giá trị, bất biến |
| Aggregate | cụm nhất quán | nhóm object luôn nhất quán với nhau |
| Aggregate Root | gốc của cụm | cửa duy nhất ra vào aggregate |
| Invariant | bất biến thức | quy tắc luôn phải đúng |
| Domain Event | sự kiện nghiệp vụ | ghi nhận việc đã xảy ra |
| Repository | kho chứa | nơi lấy/lưu aggregate |
| Domain Service | dịch vụ nghiệp vụ | logic không thuộc aggregate nào |
| Application Service | dịch vụ ứng dụng | lớp điều phối use case |
| Anti-Corruption Layer | lớp chống ăn mòn | lớp dịch với hệ thống ngoài |
| Ports & Adapters | cổng và bộ chuyển | domain ở giữa, ngoại vi cắm vào |
| Saga | chuỗi bù trừ | quy trình dài, hỏng thì bù bằng nghiệp vụ |
| Compensation | hành động bù trừ | việc làm để "gỡ" bước trước |
| Eventual Consistency | nhất quán sau cùng | khớp nhau sau một lúc, không tức thì |
| CQRS | tách đọc–ghi | hai model cho hai mục đích |
| Read Model | mô hình đọc | bảng bẹt chỉ để hiển thị |
| Outbox | hộp gửi đi | bảng event ghi cùng transaction |
| Anemic Model | mô hình thiếu máu | class rỗng chỉ có getter/setter |
| Knowledge Crunching | moi kiến thức | ngồi với chuyên gia để đào quy tắc |

---

# PHẦN IX — Tự kiểm tra

Trả lời được không nhìn tài liệu thì đã hiểu:

**Cơ bản**

1. Value Object khác Entity ở điểm nào? Cho ví dụ trong Portage.
2. Vì sao `Money` không dùng `float64`?
3. Vì sao `Currency` phải mang theo số chữ số thập phân?
4. Invariant là gì? Kể 2 invariant của Portage.
5. Anemic Domain Model là gì, vì sao nó tệ?

**Trung bình**

6. Vì sao `CustomerOrder` và `PurchaseTask` phải là hai aggregate khác nhau?
7. "Một transaction = một aggregate" nghĩa là gì, vì sao?
8. Vì sao domain không được import driver database?
9. `Source`/`AsOf` của tỷ giá đã bị đưa ra khỏi `shared.ExchangeRate`. Chúng đi
    đâu, và câu hỏi nào giúp quyết định một field có thuộc shared kernel không?
10. Vì sao domain event luôn đặt tên ở thì quá khứ?

**Nâng cao**

11. Vì sao ranh giới giữa "hoàn 100% cọc" và "mất cọc" lại là sự kiện
    `PurchaseConfirmed`, không phải một mốc thời gian?
12. Nếu bỏ Outbox pattern thì hỏng chuyện gì? Kể một kịch bản cụ thể.
13. Vì sao AI (OpenAI) phải nằm sau interface, không gọi thẳng trong domain?
14. Nếu gộp Quote và Actual làm một cột `price` thì mất thông tin gì?
15. Muốn thêm brand Adidas, cần sửa những file nào? (Đáp án đúng: **không sửa
    file nào trong `internal/domain/`**.)

**Từ đợt review 04/09**

16. Vì sao `Grams(-1)` panic nhưng `NewWeight(-1)` trả `error`? Ai gọi cái nào?
17. `VolumetricWeight` cắt xuống 1900,8 g → 1900 g rồi `RoundUpTo(100 g)` vẫn ra
    1900. Vì sao bước cuối làm tròn lên mà vẫn sai?
18. Vì sao `"26000.50"` và `"26000.5"` phải là **cùng một** `ExchangeRate`
    (`==`)? Điều đó liên quan gì đến định nghĩa value object?
19. `shared.Money{}` là gì? Vì sao Go không cho cấm nó, và domain xử lý thế nào?
20. `github.com/google/uuid` nằm trong domain có vi phạm Dependency Rule không?
    Tiêu chí phân biệt "tiện ích" và "framework/SDK" là gì?
21. Tại sao `Events` là kiểu duy nhất trong `shared` có pointer receiver, và vì
    sao điều đó không phá nguyên tắc "shared toàn value object bất biến"?

**Từ tầng app (04/09, tối)** — thêm 15 câu ở [WALKTHROUGH.md](WALKTHROUGH.md) §13.

22. Hai lỗi duy nhất của tầng app là gì, và tiêu chí để một lỗi được ở đó?
23. Vì sao `AddProduct` (command) không phải là `catalog.ProductDetails`, còn
    `RegisterMerchant` thì nhận thẳng `catalog.MerchantDetails`?

---

# PHẦN X — Đọc thêm

| Sách / nguồn | Ghi chú |
|---|---|
| *Domain-Driven Design* — Eric Evans (2003) | sách gốc, dày và khó, nên đọc sau |
| *Implementing Domain-Driven Design* — Vaughn Vernon | thực dụng hơn, nhiều code |
| *Domain-Driven Design Distilled* — Vaughn Vernon | mỏng, đọc trước cuốn trên |
| *Learning Domain-Driven Design* — Vlad Khononov | dễ tiếp cận nhất cho người mới |
| martinfowler.com — bài *AnemicDomainModel* | ngắn, nên đọc ngay |
| *Domain-Driven Design with Golang* — Matthew Boyle | đúng ngôn ngữ đang dùng |

> **Lời khuyên:** đừng đọc hết rồi mới code. Code tới đâu, đọc lại phần đó tới
> đó. DDD là thứ chỉ hiểu được khi đã tự tay làm sai một lần.
