# DDD — Sổ tay học, bám theo project Portage

> Tài liệu học cá nhân. Mỗi khái niệm DDD được giải thích bằng **chính code
> trong project này**, và đối chiếu với cách PHP/Symfony/Doctrine làm.
>
> Cập nhật: 2026-09-03 · Xem thêm [SETUP.md](SETUP.md)

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

## 6. Bounded Context — Ngữ cảnh giới hạn 🔜

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

Bounded context của Portage (thư mục đã tạo sẵn, phần lớn còn rỗng):

```
internal/domain/
├── shared/         ✅  Shared Kernel — Money, Weight, Rate
├── catalog/        🔜  Merchant, Category, CategoryPolicy
├── pricing/        🔜  RateCard, ShippingLane, Quote
├── ordering/       🔜  CustomerOrder  ← aggregate chính
├── procurement/    🔜  PurchaseTask
└── logistics/      🔜  Parcel, ConsolidationBatch
```

> **Mẹo nhận ra ranh giới context:** nếu hai bộ phận cãi nhau về định nghĩa một
> từ, và **cả hai đều đúng** — đó chính là ranh giới. Đừng bắt họ thống nhất,
> hãy tách context.

## 7. Context Map — Bản đồ quan hệ 📋

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

Context map của Portage:

```
   ┌──────────────┐  ProductPublished  ┌──────────────┐
   │   CATALOG    │───────event───────▶│   PRICING    │
   └──────────────┘                    └──────┬───────┘
                                              │ QuoteIssued
                                              ▼
                                       ┌──────────────┐
                                       │   ORDERING   │  ← aggregate chính
                                       │ CustomerOrder│
                                       └──────┬───────┘
                                              │ OrderDeposited
                                              ▼
                                       ┌──────────────┐
                                       │ PROCUREMENT  │  1 đơn khách
                                       │ PurchaseTask │  → N việc mua
                                       └──────┬───────┘
                     ┌────────────────────────┼────────────────┐
                     ▼                        ▼                ▼
            ┌────────────────┐      ┌────────────────┐  ┌─────────────┐
            │ ACL: Nike      │      │ ACL: TNF       │  │ ACL: Sony   │  ← Anti-Corruption
            └────────────────┘      └────────────────┘  └─────────────┘
                     │ PurchaseConfirmed / OutOfStock
                     ▼
            ┌──────────────────────┐
            │      LOGISTICS       │  kho Mỹ → gom lô → về VN
            │ Parcel, Consolidation│
            └──────────────────────┘
```

## 8. Shared Kernel ✅

**Là gì.** Phần code nhỏ mà **nhiều context dùng chung**.

**Vì sao phải giữ nhỏ.** Vì mọi context đều phụ thuộc vào nó — sửa một dòng là
ảnh hưởng tất cả. Shared kernel phình to = quay lại God Object.

**Trong Portage** — `internal/domain/shared/`, chỉ chứa thứ *thật sự* dùng chung:

```
money.go     Money, Currency, ExchangeRate
rate.go      Rate (tỷ lệ phần trăm)
weight.go    Weight, Dimensions, ChargeableWeight
```

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

### Value Object thứ ba: ExchangeRate — và bài học về thời gian

```go
type ExchangeRate struct {
    from, to         Currency
    minorPerMinorPPM int64
    Source           string    // lấy từ đâu
    AsOf             string    // lúc nào
}
```

**Vì sao tỷ giá phải là value object có `AsOf`?**

Vì báo giá phải **chụp ảnh** tỷ giá lúc báo. Nếu `Quote` chỉ trỏ tới "tỷ giá
hiện tại", thì báo giá hôm qua sẽ **tự đổi giá hôm nay** sau lưng khách.

Cùng nguyên tắc đó áp cho bảng giá cân ký: hôm nay bạn tăng giá, báo giá hôm qua
**không được đổi**. Nên bảng giá phải **có phiên bản**, và Quote giữ ảnh chụp.

## 12. Entity — Thực thể 🔜

**Là gì.** Object **có danh tính riêng**, phân biệt được kể cả khi mọi thuộc
tính giống hệt nhau.

**Khác Value Object thế nào:**

| | Value Object | Entity |
|---|---|---|
| Danh tính | không có | có ID |
| So sánh | theo giá trị | theo ID |
| Thay đổi | bất biến | thay đổi theo thời gian |
| Ví dụ | `Money`, `Weight` | `CustomerOrder`, `PurchaseTask` |

Hai đơn hàng cùng sản phẩm, cùng giá, cùng ngày — vẫn là **hai đơn khác nhau**.
Vì chúng có ID riêng, và mỗi cái có số phận riêng.

**Trong Portage** — ID sinh trong domain, **không đợi database**:

```go
func New(...) (*CustomerOrder, error) {
    return &CustomerOrder{
        id: NewOrderID(),   // UUID, có danh tính NGAY
        ...
    }, nil
}
```

> **So với Doctrine:** entity Symfony của bạn có `private ?int $id = null` — ID
> là `null` cho tới khi `flush()`. Nghĩa là object *tồn tại mà chưa có danh
> tính*. Trong DDD đó là trạng thái không hợp lệ. Dùng UUID sinh ở domain thì
> object có danh tính từ giây đầu tiên, và không phụ thuộc DB.

## 13. Invariant — Bất biến thức 🔜

**Là gì.** Quy tắc **luôn phải đúng**, không có ngoại lệ, không có "trừ trường
hợp".

**Nguyên tắc vàng:** object **không được phép tồn tại** ở trạng thái vi phạm
invariant — *dù chỉ một phần nghìn giây*.

**Trong Portage:**

| Invariant | Ở aggregate nào |
|---|---|
| Số tiền cọc = 50% tổng đơn | `CustomerOrder` |
| Đơn đã giao thì không huỷ được | `CustomerOrder` |
| Trọng lượng tính phí ≥ cân thật | `Parcel` |
| Lô đã đóng thì không thêm đơn được | `ConsolidationBatch` |
| Đơn giá phải cùng tiền tệ với tổng | `Quote` |

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

## 14. Aggregate & Aggregate Root 🔜

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

## 15. Domain Event — Sự kiện nghiệp vụ 🔜

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
> và trong code có đúng `DepositPaid`.

## 16. Repository — Kho chứa 🔜

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

## 17. Domain Service 📋

**Là gì.** Nơi đặt logic nghiệp vụ **không thuộc về riêng một aggregate nào**.

**Khi nào cần.** Khi một quy tắc liên quan tới nhiều aggregate, hoặc là một phép
tính thuần không có trạng thái.

**Trong Portage — ví dụ hay nhất:**

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
| `FreightAllocator` | `PlaceOrderHandler` |

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

**Bằng chứng đã hoạt động trong Portage:** 12 test chạy trong **0,5 giây**,
không cần Docker, không cần mạng, không cần API key.

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

Kết quả hiện tại — **toàn bộ là thư viện chuẩn của Go**:

```
errors
fmt
math/big
strconv
strings
```

Cách kiểm tra nhanh hơn, dùng được trong CI sau này:

```bash
# Thư viện ngoài luôn có dấu chấm trong tên (github.com/…, gorm.io/…)
go list -f '{{range .Imports}}{{println .}}{{end}}' ./internal/domain/... | sort -u | grep "\."
# Không in ra gì  =  domain sạch
# In ra bất cứ dòng nào  =  có thứ bên ngoài đã rò vào domain
```

> ⚠️ Đừng dùng `go list -deps` để kiểm tra việc này — nó liệt kê cả dependency
> **gián tiếp**, nên sẽ hiện `runtime`, `os`, `syscall`… (những thứ `fmt` kéo
> theo) và làm bạn tưởng domain bẩn. `.Imports` mới là import trực tiếp.

## 22. Anti-Corruption Layer (ACL) 📋

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
├── nike.go       ACL cho nike.com
├── tnf.go        ACL cho thenorthface.com
└── generic.go    ACL cho web lạ — dùng OpenAI trích xuất từ link khách dán
```

**Đây là pattern DDD sinh ra đúng cho bài toán của bạn.** Và đó là lý do khi
adapter đổi từ "nhân viên nhập tay" sang "API thật của nhà phân phối", **domain
không đổi một dòng**.

## 23. Đa thương hiệu = dữ liệu, không phải code 🔜

Yêu cầu của bạn — *"mở rộng cho brand khác, không chỉ giày"* — có một hệ quả kỹ
thuật rất cụ thể:

> **Từ "Nike" không được xuất hiện một lần nào trong `internal/domain/`.**
> Nike là **một dòng trong bảng**, không phải một nhánh `if`.
> Thêm Adidas = thêm một dòng dữ liệu, **không phải deploy**.

```go
// ❌ SAI — Nike bị nhốt cứng trong nghiệp vụ
if order.Brand == "Nike" { ... }

// ✅ ĐÚNG — merchant là dữ liệu
type Merchant struct {
    ID           MerchantID
    Name         string        // "Nike", "The North Face", "Sony"
    FreeShipOver shared.Money  // mỗi hãng một ngưỡng free ship
    Sourcing     SourcingMode  // Operator | Feed | API
}

// ✅ ĐÚNG — ngành hàng quyết định thuế, quy đổi, hạn chế
type CategoryPolicy struct {
    Code          CategoryCode  // footwear | apparel | electronics
    HSCode        string
    DutyRate      shared.Rate
    VolumetricDiv int           // hàng cồng kềnh có thể dùng số chia khác
    Restrictions  []Restriction // Battery, Liquid, Supplement…
}
```

`Restriction{Battery}` là thật: tai nghe Sony có pin lithium, không đi được mọi
tuyến bay. Nếu domain không biết khái niệm này, ngày đầu bán đồ điện tử là ngày
hàng kẹt ở kho.

## 24. CQRS 📋

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

## 25. Eventual Consistency — Nhất quán sau cùng 📋

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

## 26. Saga & Compensation 📋

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

## 27. Outbox Pattern 📋

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

## 28. Quote vs Actual — Ước tính và Thực tế 📋

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

---

# PHẦN VI — Trạng thái hiện tại của Portage

## Đã code ✅

```
internal/domain/shared/
├── money.go        Money, Currency, ExchangeRate, Rate      ← Value Object
├── rate.go         Rate, ExchangeRate                        ← Value Object
├── weight.go       Weight, Dimensions, ChargeableWeight      ← Value Object
├── money_test.go   7 test
└── weight_test.go  5 test
```

```
go test ./...   →  12/12 PASS   ·   coverage 71,2%   ·   0,5 giây
gofmt           →  sạch
go vet          →  sạch
```

**Khái niệm DDD đã thật sự hiện diện trong code:**

| Khái niệm | Bằng chứng |
|---|---|
| Value Object | `Money`, `Weight`, `Rate`, `ExchangeRate`, `Dimensions` |
| Bất biến | mọi method trả về giá trị mới, không sửa tại chỗ |
| Ubiquitous Language | `ChargeableWeight`, `VolumetricWeight` |
| Primitive obsession (tránh được) | không có `float64` cho tiền |
| Shared Kernel | `internal/domain/shared/` |
| Dependency Rule | domain chỉ import thư viện chuẩn Go |
| Test là tài liệu | test áo phao 0,9kg → 4,7kg |

## Chưa code 🔜

Aggregate · Entity · Domain Event · Repository · Bounded Context (5 cái còn
rỗng) · ACL · Saga · CQRS · Outbox · toàn bộ tầng adapter

> Nói cho đúng: hiện tại project mới có **Value Object và Shared Kernel**. Đó là
> nền móng thật, nhưng chưa phải "một hệ thống DDD". Khi phỏng vấn, nói đúng
> phần đã làm.

## Lộ trình tiếp theo

```
[ ] catalog/     Merchant, CategoryPolicy   → học Entity, đa thương hiệu
[ ] pricing/     RateCard, Quote            → học Aggregate, Quote vs Actual
[ ] ordering/    CustomerOrder              → học Invariant, Domain Event, luật cọc
[ ] procurement/ PurchaseTask               → học ranh giới Aggregate, ACL
[ ] logistics/   Parcel, Batch              → học Domain Service (chia cước)
[ ] adapter/postgres                        → học Repository, Outbox
[ ] adapter/http                            → học Application Service
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
9. Vì sao `ExchangeRate` cần trường `AsOf`?
10. Vì sao domain event luôn đặt tên ở thì quá khứ?

**Nâng cao**

11. Vì sao ranh giới giữa "hoàn 100% cọc" và "mất cọc" lại là sự kiện
    `PurchaseConfirmed`, không phải một mốc thời gian?
12. Nếu bỏ Outbox pattern thì hỏng chuyện gì? Kể một kịch bản cụ thể.
13. Vì sao AI (OpenAI) phải nằm sau interface, không gọi thẳng trong domain?
14. Nếu gộp Quote và Actual làm một cột `price` thì mất thông tin gì?
15. Muốn thêm brand Adidas, cần sửa những file nào? (Đáp án đúng: **không sửa
    file nào trong `internal/domain/`**.)

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
