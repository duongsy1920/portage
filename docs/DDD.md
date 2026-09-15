# DDD — Sổ tay khái niệm, bám theo project Portage

> Mỗi khái niệm DDD được giải thích bằng **chính code trong repo**, và đối chiếu với cách
> PHP/Symfony/Doctrine làm. Mọi thứ mô tả ở đây **đều đã có trong code**, trừ khi ghi rõ
> *"chưa có"*. Lịch sử vì sao code thành ra như vậy nằm ở [SETUP.md §9](SETUP.md), không lặp
> lại ở đây.

## Cách đọc file này

Mỗi khái niệm có bốn phần cố định, đọc theo thứ tự:

| Phần | Trả lời |
|---|---|
| **Là gì** | định nghĩa ngắn, bằng tiếng thường |
| **Vì sao cần** | vấn đề thật nó giải quyết |
| **Trong Portage** | code thật, đường dẫn file thật |
| **So với Symfony** | ánh xạ sang thứ bạn đã biết |

Một số mục có thêm **Bẫy đã gặp**: chuyện code từng sai ở đúng chỗ đó. Lần đọc đầu có thể bỏ
qua khối này. Lần đọc thứ hai thì đừng, vì đó là chỗ khái niệm trở thành thói quen.

Nếu bạn chưa quen từ vựng, đọc bảng 12 từ ở đầu [HOC.md](HOC.md) trước.

---

# PHẦN I — DDD là gì

## 1. Định nghĩa

**Là gì.** Domain-Driven Design (Eric Evans, 2003) là cách xây phần mềm mà **nghiệp vụ là
trung tâm**, không phải database hay framework.

DDD **không phải** một design pattern, một thư viện, một cấu trúc thư mục, và không bắt buộc
microservices.

DDD **là** hai nhóm việc:

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
   │  Cách viết code BÊN TRONG một vùng. Dễ học, dễ tìm  │
   │  ví dụ, và cũng dễ bị lạm dụng.                      │
   │  → Value Object, Entity, Aggregate, Repository...    │
   └─────────────────────────────────────────────────────┘
```

Đa số người "học DDD" chỉ học nhóm dưới, rồi nhét mọi thứ vào một context khổng lồ. Đó là lý
do họ thấy DDD rườm rà mà không được gì.

## 2. Khi nào KHÔNG nên dùng DDD

| Tình huống | Nên dùng gì |
|---|---|
| CRUD thuần: thêm, sửa, xoá, không có quy tắc | Doctrine + Controller là đủ |
| Script chạy một lần | viết thẳng |
| Prototype vứt đi sau 2 tuần | đừng phí công |
| Nghiệp vụ đơn giản, đội 1 người, không đổi | quá tay |

DDD **đáng tiền** khi nghiệp vụ phức tạp, quy tắc nhiều và hay đổi, nhiều người cùng làm, hệ
thống sống lâu.

**Portage đáng dùng DDD vì:** một đơn đi qua 5 giai đoạn ở 2 quốc gia, có tiền, có thuế, có
bên thứ ba không kiểm soát được (web bán hàng Mỹ, hãng vận chuyển), và luật hoàn tiền phụ
thuộc vào việc *đã mua hàng hay chưa*.

---

# PHẦN II — Strategic Design (chiến lược)

## 3. Ubiquitous Language — Ngôn ngữ chung

**Là gì.** Một bộ từ vựng **duy nhất**, dùng y hệt nhau ở mọi nơi: lúc nói chuyện với khách,
trong tài liệu, trong tên kiểu, trong tên cột database.

**Vì sao cần.** Mỗi lần "dịch" từ ngôn ngữ nghiệp vụ sang ngôn ngữ kỹ thuật là một lần hiểu
sai. Dịch qua dịch lại đủ nhiều thì hệ thống không còn phản ánh nghiệp vụ nữa.

**Trong Portage.** Từ dân trong nghề dùng đi thẳng vào tên kiểu:

| Người trong nghề nói | Trong code | Nghĩa |
|---|---|---|
| bảng giá cân ký | `pricing.RateCard` | giá mỗi kg cho từng loại hàng |
| cân tính phí | `shared.ChargeableWeight` | max(cân thật, quy đổi thể tích), làm tròn lên |
| quy đổi thể tích | `Dimensions.VolumetricWeight` | dài × rộng × cao (mm³) / 5000 |
| cọc | `CustomerOrder.Deposit` | 50 % khách trả trước |
| gom lô | `logistics.ConsolidationBatch` | nhiều thùng đi chung một hoá đơn cước |
| đơn của khách | `ordering.CustomerOrder` | thứ khách đặt trên web mình |
| việc đi mua | `procurement.PurchaseTask` | việc mua thật ở web Mỹ |
| shop | `catalog.Merchant` | Nike, The North Face, Sony… là **dữ liệu** |
| ai bảo đảm dữ liệu này | `catalog.Provenance` | nguồn + lúc nào + ai xác nhận |
| điểm không thể quay đầu | `CustomerOrder.ConfirmPurchase` | sau đó huỷ là mất cọc |

Cách đặt tên sai trông như thế này, và repo không có dòng nào như vậy:

```go
type ShippingFeeCalculationServiceImpl struct{}   // dịch sang tiếng lập trình viên, mất nghĩa
func computeWeightValue(...) float64
```

**So với Symfony.** Không khác. Chỉ là ở Symfony, tên class hay bị Doctrine và Bundle kéo
thành `ProductEntity`, `OrderManager`, `PriceHelper`.

> **Quy tắc:** nếu bạn phải giải thích một tên biến cho người làm nghiệp vụ, tên đó sai. Đổi
> tên, đừng viết comment.

## 4. Knowledge Crunching — Moi kiến thức nghiệp vụ

**Là gì.** Lập trình viên ngồi với **chuyên gia nghiệp vụ** để đào ra quy tắc thật, rồi mô
hình hoá lại. Trong project này, chủ repo vừa là lập trình viên vừa là chuyên gia: anh ấy đã
tự đặt hàng Mỹ về Việt Nam nhiều lần.

**Trong Portage.** Hai giả định ban đầu của thiết kế đã bị nghiệp vụ thật sửa lại:

| Giả định ban đầu | Thực tế | Hệ quả lên thiết kế |
|---|---|---|
| ship lẻ từng đơn | gom lô, chờ vài tuần | sinh ra aggregate `ConsolidationBatch` và bài toán chia cước |
| đóng thuế nhập khẩu và VAT ở Việt Nam | trả trọn gói ở đầu Mỹ, về tận nhà | `ShippingLane` thành khái niệm bậc nhất; thuế thuộc **tuyến**, không thuộc **món hàng** |

> **Bài học:** mô hình sai không phải do code dở, mà do **thiếu kiến thức nghiệp vụ**. Không
> có cách nào "code cẩn thận hơn" để tránh. Chỉ có cách hỏi.

## 5. Domain, Subdomain

**Là gì.** *Domain* là toàn bộ lĩnh vực kinh doanh. Chia nhỏ thành *subdomain*, và **không
phải subdomain nào cũng đáng đầu tư như nhau**:

| Loại | Nghĩa | Nên làm gì | Trong Portage |
|---|---|---|---|
| **Core** | lợi thế cạnh tranh, lý do khách chọn bạn | tự làm, làm kỹ | pricing, procurement, logistics (chia cước) |
| **Supporting** | cần có nhưng không tạo khác biệt | tự làm, làm vừa đủ | catalog, reporting |
| **Generic** | ai cũng cần, ai cũng giống nhau | dùng sẵn, đừng tự viết | auth (bearer token tối giản), thanh toán (chưa có) |

**Vì sao quan trọng.** Đây là bản đồ phân bổ công sức. Tự viết hệ thống đăng nhập đầy đủ
trong khi thuật toán chia cước còn sai là đầu tư sai chỗ. Repo vì thế có auth ở mức "chìa
khoá và loại chìa", không có user, role, hay đăng nhập.

## 6. Bounded Context — Ngữ cảnh giới hạn

**Là gì.** Một **ranh giới** mà bên trong đó, mỗi từ có đúng **một** nghĩa. Đây là khái niệm
quan trọng nhất của cả DDD.

**Vì sao cần.** Cùng một từ, hai bộ phận hiểu khác nhau. Ép chúng dùng chung một class là
nguồn gốc của "God Object" 40 cột.

**Trong Portage.** Từ "sản phẩm" mang nghĩa khác nhau ở mỗi vùng:

| Context | "Product" ở đây nghĩa là gì | Kiểu thật |
|---|---|---|
| **catalog** | Air Max 90: tên, shop, ngành hàng, các size | `catalog.Product` |
| **pricing** | một món có giá, một loại hàng, một hộp để tính cước | `pricing.Listing` |
| **procurement** | một dòng cần đi mua: tên, size, mã shop, link | `procurement.Item` + `Subject` |
| **logistics** | một vật thể có cân nặng và thể tích | `logistics.Parcel` |

```
   ❌ CÁCH SAI — một struct cho tất cả          ✅ CÁCH ĐÚNG — mỗi context một model, nối bằng ID

        ┌──────────────────────────┐              CATALOG          PRICING         LOGISTICS
        │  Product                 │              ┌─────────┐      ┌─────────┐     ┌─────────┐
        │  id, name, price, tax,   │              │ Product │      │ Listing │     │ Parcel  │
        │  weight, dims, url,      │              │ name    │◀─id─▶│ price   │◀─id▶│ actual  │
        │  size, color, stock,     │              │ variants│      │ parcel  │     │ batch   │
        │  batchId, ... 40 cột     │              └─────────┘      └─────────┘     └─────────┘
        └──────────────────────────┘
```

Năm context là năm thư mục:

```
internal/domain/
├── shared/         Shared Kernel: Money, Weight, Rate, ExchangeRate, ID, OperatorID, Events, ParcelSpec, Allocate
├── catalog/        Merchant, Product ⊃ Variant, CategoryPolicy, Provenance — 17 event
├── pricing/        ShippingLane, QuotePolicy, Calculate, Quote, Reconciliation, Listing — 4 event
├── ordering/       CustomerOrder, Refund, AcceptedQuote, Variant — 8 event
├── procurement/    PurchaseTask, PurchaseReceipt, Subject, port MerchantACL — 3 event
└── logistics/      Parcel, ConsolidationBatch, FreightAllocator, LaneRule — 5 event
```

Ranh giới này **nhìn thấy được ở ba chỗ** trong code, không chỉ là quy ước trên giấy:

1. **Kiểu.** `pricing.Listing.Product` là `shared.ID`, không phải `catalog.ProductID`. Pricing
   không gọi được tên kiểu của catalog; id là một danh tính mờ đi qua dây.
2. **SQL.** Cột `listings.product` trong `0002_pricing.sql` là `uuid` trần, không `REFERENCES`
   sang `products`. Khoá ngoại chéo context là cách phá ranh giới nhanh nhất.
3. **Từ vựng.** Catalog nói `footwear`; pricing tự dịch sang `branded` bằng `Classification`
   vì đó là từ của nhà gửi hàng. Cùng đôi giày, hai context, hai từ, cả hai đều đúng.

Và có **test canh**: `TestDecision_boundedContextsDoNotImportEachOther` trong
`internal/domain/decisions_test.go` đỏ ngay khi `catalog` import `pricing`. Chỉ `shared` được
dùng chung.

**So với Symfony.** Gần nhất là mỗi context một Bundle với schema Doctrine riêng, không
`@ORM\ManyToOne` chéo bundle. Trong thực tế Symfony hiếm khi làm được vì Doctrine muốn một
EntityManager biết mọi entity.

> **Mẹo nhận ra ranh giới:** nếu hai bộ phận cãi nhau về định nghĩa một từ, và **cả hai đều
> đúng**, đó chính là ranh giới. Đừng bắt họ thống nhất; hãy tách context.

## 7. Context Map — Bản đồ quan hệ

**Là gì.** Sơ đồ mô tả các context **nói chuyện với nhau thế nào** và **ai phụ thuộc ai**.

Các kiểu quan hệ thường gặp:

| Kiểu | Nghĩa | Trong Portage |
|---|---|---|
| **Shared Kernel** | dùng chung một phần code nhỏ, đổi phải bàn với nhau | `internal/domain/shared` |
| **Customer/Supplier** | context dưới phụ thuộc context trên, nhưng có tiếng nói | pricing nghe catalog |
| **Conformist** | phải theo model của bên kia, không cãi được | *(không có)* |
| **Anti-Corruption Layer** | lớp dịch để model bên kia không rò vào mình | `MerchantACL`, `ListingExtractor` (§22) |
| **Published Language** | một định dạng chung để trao đổi | `internal/contracts` (*V1) |

**Trong Portage.** Mọi cạnh của bản đồ là một event. Đường đi của một event qua ranh giới,
sáu file theo thứ tự:

```
 1. domain/catalog/events.go          ProductPublished{ID, Merchant, Category, Name, Source, Price, Parcel, At}
                                      event mang ĐỦ dữ liệu, vì bên nghe không load được aggregate của catalog
 2. adapter/eventcodec/codec.go       Encode: event → JSON, key viết tay = hợp đồng
 3. contracts/catalog_v1.go           ProductPublishedV1 — chỉ primitive, version trong tên, package riêng
 4. adapter/eventcodec/decode.go      Decode: JSON → V1 struct; chỉ viết khi có người nghe
 5. platform/wire/subscribe.go        bus.Subscribe("catalog.product_published", on(projector.OnProductPublished))
 6. app/pricing/projector.go          upsert pricing.Listing — idempotent, chịu sai thứ tự
```

Bản đồ đầy đủ chính là 35 dòng của `wire.Subscribe`, đọc từ trên xuống:

```
   catalog.*                       → pricing (listings, profiles) · procurement (shops, items, variants)
                                   → ordering (variants: id có thật không) · reporting (worklist)
   pricing.quote_accepted          → ordering (accepted quotes)
   pricing.lane_defined            → logistics (cách đếm cân của tuyến)
   ordering.deposit_paid           → procurement mở việc mua, hỏi ACL của shop
   ordering.order_placed           → pricing (đơn này dựa trên quote nào — Quote vs Actual)
   procurement.purchase_confirmed  → ordering (qua điểm không thể quay đầu) · logistics (chờ thùng)
                                   → pricing (giá THẬT đã trả) · reporting
   procurement.purchase_failed     → ordering (bù trừ: huỷ sẽ hoàn đủ)
   logistics.batch_shipped         → ordering (in_transit) · pricing (cước THẬT) · reporting
   mọi event ở trên                → reporting: order_summaries, product_worklist
```

**So với Symfony.** `messenger.yaml routing:` cộng mọi `#[AsMessageHandler]` gom lại thành
một hàm đọc được. Ai nghe cái gì không còn phải grep.

## 8. Shared Kernel

**Là gì.** Phần code nhỏ mà **nhiều context dùng chung**.

**Vì sao phải giữ nhỏ.** Mọi context đều phụ thuộc vào nó: sửa một dòng là ảnh hưởng tất cả.
Shared kernel phình to là quay lại God Object.

**Trong Portage.** `internal/domain/shared/`, và chỉ có thế này:

```
money.go        Money, NewMoney, ParseMoney, Add/Sub/Mul/Convert
currency.go     Currency, CurrencyFromCode
rate.go         Rate (tỷ lệ phần trăm), ExchangeRate (tỷ giá, 12 số lẻ)
decimal.go      parseDecimal, addExact, divRoundHalfUp — máy móc dưới các value object (private)
weight.go       Weight, Dimensions, VolumetricWeight, ChargeableWeight
parcelspec.go   ParcelSpec — cân nặng + hộp; dùng cho cả ước lượng (category) và số đo thật (product)
allocate.go     Allocate — chia một số tiền theo trọng số, tổng đúng từng cent
id.go           ID (UUID v7)
operator.go     OperatorID — "ai làm"; catalog, procurement, logistics, ordering đều hỏi
event.go        Event, Events — cách mọi aggregate ghi event
```

Không có `Product`, `Customer`, `Order` nào ở đây. Những thứ đó mỗi context tự định nghĩa
theo nghĩa riêng.

Câu hỏi để quyết định một thứ có được vào shared kernel không: *"context nào KHÔNG cần nó?"*
Có một context trả lời "không cần" thì nó không thuộc shared kernel. `ParcelSpec` vào đây
đúng theo câu hỏi đó: nó sống trong `catalog` cho tới khi `pricing` cần **đúng kiểu đó**, cùng
nghĩa, không phải một bản sao mang nghĩa khác.

**Bẫy đã gặp.** Bản đầu của `ExchangeRate` có hai field `Source` và `AsOf` (lấy ở ngân hàng
nào, lúc nào). Chỉ pricing cần chúng để đối soát; ordering và logistics thì không. Để trong
shared kernel là bắt **mọi** context gánh một mối quan tâm của **một** context. Đã gỡ.
(SETUP.md §9 đợt 1.)

**So với Symfony.** `src/Shared/Domain/ValueObject/`, với một khác biệt: ở đây không có
`#[ORM\Embeddable]`. Struct không biết Doctrine tồn tại.

---

# PHẦN III — Tactical Design (chiến thuật)

## 9. Anemic Domain Model — Mô hình thiếu máu

**Là gì.** Class *trông như* object nhưng thực chất chỉ là **cái bao đựng dữ liệu** với
getter và setter. Martin Fowler gọi đây là anti-pattern.

**Trông thế nào.** Một entity Symfony điển hình:

```php
class Booking
{
    private ?string $status = null;
    private ?string $origin = null;
    private ?string $destination = null;

    public function setStatus(?string $status): static { ... }
    public function setOrigin(?string $origin): static { ... }
    // ... setter cho TẤT CẢ
}
```

Nó cho phép những điều này, và không dòng nào báo lỗi:

```php
$b->setStatus('bananas');       // "bananas" là trạng thái hợp lệ?
$b->setOrigin('HCM');
$b->setDestination('HCM');      // ship từ HCM tới HCM?
$em->persist(new Booking());    // Booking rỗng vẫn lưu được
```

**Vấn đề.** Object không tự bảo vệ được mình. Mọi quy tắc nghiệp vụ phải nằm ở chỗ khác
(Controller, Service) và rải rác khắp nơi. Sáu tháng sau không ai biết luật đầy đủ là gì.

**Trong Portage.** Không aggregate nào có setter. `TestDecision_noSettersOnDomainTypes` canh
điều đó bằng `go/ast`: một method tên `SetXxx` trong `internal/domain/` là build đỏ. Thay
setter là **hành vi có tên nghiệp vụ**: `Merchant.Suspend(reason, now)`,
`CustomerOrder.PayDeposit(amount, now)`, `Product.Publish(now)`. Mỗi hành vi tự kiểm luật
rồi mới đổi trạng thái.

**Đối lập của nó** là **Rich Domain Model**: object chứa cả dữ liệu và hành vi, tự từ chối
trạng thái sai.

## 10. Primitive Obsession — Nghiện kiểu nguyên thuỷ

**Là gì.** Dùng `string`, `int`, `float` cho những khái niệm nghiệp vụ có quy tắc riêng.

**Vì sao nguy hiểm.**

```go
func Ship(weight float64, price float64) { ... }

Ship(150.00, 1.2)   // compile được. Nhưng tham số đã đảo: 150 kg giá 1.2 đô. Không ai thấy.
```

Tàu thăm dò Mars Climate Orbiter của NASA nổ vì đúng loại lỗi này: một bên dùng pound-force,
một bên dùng newton, cùng là `float`.

**Trong Portage.** Mọi khái niệm nghiệp vụ có kiểu riêng, và compiler chặn thay người:

```go
func Ship(w shared.Weight, p shared.Money) { ... }

Ship(price, weight)   // KHÔNG COMPILE
```

Cả những thứ trông như "chỉ là chuỗi" cũng có kiểu: `catalog.Hostname`, `catalog.CategoryCode`,
`pricing.LaneCode`, `catalog.SourceURL`. Mỗi cái chỉ tạo được qua một hàm `Parse*` có kiểm tra,
nên không có đường tắt kiểu `CategoryCode("Giày Dép!")`.

## 11. Value Object — Đối tượng giá trị

**Là gì.** Object **không có danh tính**, chỉ có giá trị. Hai value object bằng nhau khi mọi
thuộc tính bằng nhau. Và **bất biến**: muốn "đổi" thì tạo cái mới.

Tiền 100 000 ₫ trong ví bạn và 100 000 ₫ trong ví tôi là **như nhau**. Không cần phân biệt
"tờ tiền nào". Đó là value object.

**Ba đặc điểm bắt buộc:**

1. **Bất biến**: không sửa, chỉ tạo cái mới.
2. **So sánh theo giá trị**, không theo ID.
3. **Tự kiểm tra tính hợp lệ**: không tồn tại được ở trạng thái sai.

### Value Object thứ nhất: `Money` — `internal/domain/shared/money.go`

```go
type Money struct {
	minor    int64      // đơn vị nhỏ nhất: USD → cent, VND → đồng
	currency Currency
}

func NewMoney(minor int64, c Currency) Money {
	if c.IsZero() {
		panic("shared: NewMoney called with the zero Currency")
	}
	return Money{minor: minor, currency: c}
}

func (m Money) Add(o Money) (Money, error) {
	if err := m.combinable(o, "add"); err != nil {   // khác tiền tệ → ErrCurrencyMismatch
		return Money{}, err
	}
	return Money{minor: addExact(m.minor, o.minor), currency: m.currency}, nil
}
```

Ba quyết định thiết kế, mỗi cái giải một vấn đề thật:

| Quyết định | Vì sao |
|---|---|
| `minor int64`, không bao giờ `float64` | trong dấu chấm động nhị phân, `0.1 + 0.2 != 0.3`. Sai một xu mỗi dòng, nhân với triệu dòng, là mất tiền thật |
| `Currency` mang theo số chữ số lẻ (`USD` 2, `VND` 0) | một `Money` cố định 2 số lẻ sẽ **âm thầm chấp nhận** `3900000.50 VND`, số tiền không tồn tại |
| cộng khác tiền tệ là **lỗi**, không phải chuyện hên xui | 150 USD + 3 900 000 VND ra một con số vô nghĩa; domain từ chối |

Field viết chữ thường nên **private trong package**: bên ngoài không có cách nào tạo một
`Money` sai ngoài việc gọi constructor. Method `Add` trả `Money` **mới**, không đụng `m`.

### Value Object thứ hai: `Weight` — `internal/domain/shared/weight.go`

```go
// chargeable = roundUp( max(actual, volumetric) )
func ChargeableWeight(actual Weight, d Dimensions, divisor int64, step Weight) Weight {
	return MaxWeight(actual, d.VolumetricWeight(divisor)).RoundUpTo(step)
}
```

Ba dòng này là **luật kinh doanh sống còn**: hãng bay tính tiền theo *chỗ chiếm*, không theo
cân nặng. Một áo phao 0,9 kg trong hộp 40 × 32 × 18 cm bị tính **4,7 kg**. Test
`TestChargeableWeight_downJacketBillsOnVolume` ghi đúng con số đó, và sáu tháng sau mở test
ra vẫn hiểu vì sao.

### Value Object thứ ba: `ExchangeRate` — và bài học về thời gian

```go
type ExchangeRate struct {
	from, to Currency
	digits   int64   // tỷ giá bỏ dấu chấm: "26000.5" → 260005
	scale    int32   // bao nhiêu chữ số là phần lẻ: 1
}
```

**Vì sao tỷ giá phải là value object?** Vì báo giá phải **chụp ảnh** tỷ giá lúc báo. Nếu
`Quote` chỉ trỏ tới "tỷ giá hiện tại", báo giá hôm qua sẽ tự đổi giá hôm nay sau lưng khách.
Một giá trị bất biến, được chép vào quote, là cách rẻ nhất để lời hứa "giá này giữ 48 giờ"
thành sự thật.

**So với Symfony.** `#[ORM\Embeddable]` với private field, hoặc `Brick\Money`. Hai khác biệt
thật: (1) struct ở đây không biết Doctrine tồn tại, phần map xuống cột nằm ở
`adapter/postgres`; (2) Go không có Reflection ghi vào private field, nên value object phải có
getter đọc ra được (`Minor()`, `Currency()`, `ExchangeRate.Rate()`) để repository lưu được.

**Bẫy đã gặp.** Năm bug thật ở đúng ba kiểu trên, mỗi cái một bài học:

| Bug | Biểu hiện | Bài học |
|---|---|---|
| `ParseMoney` "tiện tay" bỏ dấu phẩy | `"150,50"` → **15 050.00 USD**, không lỗi | domain nhận **một** định dạng `-?digits[.digits]`; dấu phẩy hay chấm là việc của adapter, theo `Accept-Language` |
| `VolumetricWeight` chia số nguyên, **cắt xuống** | 1900,8 g → 1900 g → làm tròn bước 100 g vẫn 1900; hãng tính 2000 | *"hãng luôn làm tròn lên"* phải áp vào **mọi bước**, không chỉ bước cuối |
| tỷ giá tái dùng `ParsePercent` (4 số lẻ) | `1/26000 = 0.0000384…` không biểu diễn được → mọi hoàn tiền VND→USD ra `0.00` | *"cùng là số"* không có nghĩa *"cùng một khái niệm"*; `Rate` và `ExchangeRate` là hai kiểu |
| `"26000.50"` và `"26000.5"` khác nhau khi `==` | hai tỷ giá bằng giá trị nhưng khác bit | value object cần **một dạng chuẩn** để `==` đúng; `NewExchangeRate` cắt số 0 thừa |
| `shared.Money{}` tạo được từ ngoài | tiền có currency `""`, `Add` vẫn chạy | zero value của Go là "chưa set"; constructor từ chối, và có `IsValid()` để hỏi |

Chi tiết từng bug và test tái hiện: SETUP.md §9 đợt 1 và 2.

## 12. Entity — Thực thể

**Là gì.** Object **có danh tính riêng**, phân biệt được kể cả khi mọi thuộc tính giống hệt.

| | Value Object | Entity |
|---|---|---|
| Danh tính | không có | có ID |
| So sánh | theo giá trị | theo ID |
| Thay đổi | bất biến | đổi theo thời gian, qua method |
| Ví dụ | `Money`, `Weight`, `FreeShipping`, `Provenance` | `Merchant`, `Product`, `CustomerOrder`, `PurchaseTask` |

Hai đơn hàng cùng sản phẩm, cùng giá, cùng ngày vẫn là **hai đơn khác nhau**, vì mỗi cái có
số phận riêng.

**Trong Portage.** ID sinh **trong domain**, không đợi database. `internal/domain/shared/id.go`:

```go
type ID struct{ u uuid.UUID }          // UUID phiên bản 7

func NewID() ID                        // sinh mới, 48 bit đầu là thời điểm sinh
func ParseID(s string) (ID, error)     // đọc từ URL / DB — chỉ nhận v7, đúng dạng chuẩn
func (id ID) String() string
func (id ID) IsZero() bool             // ID{} = "chưa có danh tính"
```

**Vì sao UUID v7?** Nó bắt đầu bằng thời điểm sinh (mili giây), nên sắp xếp được theo thứ tự
tạo và index B-tree của Postgres ghi nối đuôi thay vì rải ngẫu nhiên. Không dùng
auto-increment vì (1) phải hỏi DB mới có, (2) lộ số lượng đơn cho người ngoài.

**Mỗi aggregate bọc `ID` thành kiểu riêng**, một dòng, và compiler phân biệt:

```go
type OrderID struct{ shared.ID }        // embedding: dùng luôn String(), IsZero()
type MerchantID struct{ shared.ID }

func load(id OrderID) { ... }
load(merchantID)                        // KHÔNG COMPILE — đúng thứ mình muốn
```

Rồi aggregate dùng nó ngay trong constructor: `id: NewOrderID()`. Object có danh tính từ dòng
đầu, chưa chạm DB.

**So với Symfony.** Entity quen thuộc có `private ?int $id = null`: ID là `null` cho tới khi
`flush()`. Nghĩa là object *tồn tại mà chưa có danh tính*. Trong DDD đó là trạng thái không
hợp lệ. Tương đương đúng là `symfony/uid`: cột `uuid` **không** `#[ORM\GeneratedValue]`, và
constructor tự gán `$this->id = Uuid::v7()`. `final class OrderId` bọc `Uuid` bên PHP tương
ứng `type OrderID struct{ shared.ID }` bên Go.

## 13. Invariant — Bất biến thức

**Là gì.** Quy tắc **luôn phải đúng**, không có ngoại lệ, không có "trừ trường hợp".

**Nguyên tắc vàng:** object **không được phép tồn tại** ở trạng thái vi phạm invariant, dù
chỉ một phần nghìn giây.

**Trong Portage.** Mọi dòng dưới đây đều là code có test:

| Invariant | Ở đâu | Vi phạm thì |
|---|---|---|
| cân nặng không âm | `shared.NewWeight` | `ErrNegativeWeight` |
| tỷ giá > 0, hai tiền tệ khác nhau | `shared.NewExchangeRate` | `ErrInvalidRate` |
| tiền luôn có tiền tệ | `shared.NewMoney` | panic (input của lập trình viên) |
| hộp phải có cân nặng **và** đủ ba cạnh | `shared.NewParcelSpec` | `ErrIncompleteParcelSpec` |
| cọc phải **đúng** số, không thiếu không thừa | `CustomerOrder.PayDeposit` | `ErrWrongAmount` |
| đơn đã giao thì không huỷ được | `CustomerOrder.Cancel` | `ErrAlreadyDelivered` |
| đăng bán cần 5 điều kiện | `Product.Publish` | 5 sentinel khác nhau (§14) |
| lô đã đóng thì không thêm thùng | `ConsolidationBatch.AddParcel` | `ErrBatchNotOpen` |
| một quote chỉ đặt được **một** đơn | `PlaceOrderHandler` + `orders.quote UNIQUE` | `ErrQuoteAlreadyUsed` |

**Cách bảo vệ trong Go**: constructor là cửa duy nhất, và field private đóng mọi cửa khác.

```go
type CustomerOrder struct {          // field chữ thường = private trong package. KHÔNG có setter.
	shared.Events
	id      OrderID
	status  OrderStatus
	deposit shared.Money
	...
}

func PlaceOrder(d OrderDetails, now time.Time) (*CustomerOrder, error)   // cách DUY NHẤT để tạo
```

Ngoài package, `o.status = "bananas"` **không biên dịch được**.

**Mạnh hơn PHP ở chỗ:** `private` của PHP chỉ chặn lúc chạy, và Doctrine dùng Reflection chọc
thủng thoải mái. Go chặn ở **compiler**.

### Hai loại constructor: `error` hay `panic`?

Go không có exception. Khi input vi phạm invariant có hai cách phản ứng, và chọn cách nào
phụ thuộc vào **ai gọi**:

```go
// Input KHÔNG TIN ĐƯỢC (form, CSV, dòng DB) → trả error, người gọi xử lý
func NewWeight(grams int64) (Weight, error)

// Literal do LẬP TRÌNH VIÊN viết (code, test, hằng số) → panic: đây là bug, sửa code
func Grams(g int64) Weight
```

| | `Parse*`, `New*` (validating) | `Grams`, `NewMoney`, `Must*` (literal) |
|---|---|---|
| Ai gọi | adapter, với dữ liệu từ ngoài | code nghiệp vụ, test, hằng số |
| Sai thì | `error`, bắt được, báo lại người dùng | `panic`, dừng, đây là bug |
| Symfony | `\DomainException` | `\LogicException` |

`panic` **không bao giờ** dùng cho lỗi của người dùng. Nếu thấy mình muốn `recover()` một
panic để hiển thị thông báo lỗi, bạn đã dùng nhầm loại constructor.

`errors.Is(err, shared.ErrNegativeWeight)` bên Go đóng vai `catch (NegativeWeight $e)` bên
PHP. Sentinel error (`var ErrX = errors.New(...)`) cộng `fmt.Errorf("...: %w", ErrX)` để bọc
thêm ngữ cảnh mà vẫn `errors.Is` được.

## 14. Aggregate & Aggregate Root

**Là gì.** Một **cụm** object luôn phải nhất quán với nhau, và **một** object làm cửa duy
nhất ra vào cụm đó, gọi là **Aggregate Root**.

```
        ┌─ RANH GIỚI AGGREGATE ──────────────┐
        │                                    │
        │   Product  ← Aggregate Root        │
        │      │                             │
        │      ├── Variant  (entity con)     │
        │      ├── Variant                   │
        │      └── Provenance ×3 (value obj) │
        │                                    │
        └────────────────────────────────────┘
              ▲
              │ Thế giới bên ngoài CHỈ được chạm vào root.
              │ Không ai với tay vào sửa thẳng một Variant.
```

**Bốn luật của aggregate:**

1. Bên ngoài chỉ tham chiếu tới **root**, không tham chiếu vào ruột.
2. Mọi thay đổi đi qua **method của root**.
3. Root chịu trách nhiệm giữ **invariant** cho cả cụm.
4. **Một transaction = một aggregate.** Nhiều aggregate thì nối bằng event.

**Trong Portage.** `Product` ⊃ `Variant`, ở `internal/domain/catalog/product.go` và
`variant.go`:

```go
type Product struct {
	shared.Events
	id       ProductID
	variants []Variant              // entity con, sống trong root
	...
}

func (p *Product) AddVariant(d VariantDetails, now time.Time) (VariantID, error)   // cửa DUY NHẤT
func (p *Product) Variants() []Variant                                            // trả BẢN SAO
```

| Luật | Trong `Product` |
|---|---|
| 1. Ngoài chỉ chạm root | không có `*Variant` nào rò ra; `Variants()` trả copy; **không có `VariantRepository`** |
| 2. Đổi qua method của root | `AddVariant`; không có `variant.SetSize()` |
| 3. Root giữ invariant cho cả cụm | trùng (size, color) bị `AddVariant` từ chối; `Variant` không tự biết anh em của nó |
| 4. Một transaction = một aggregate | gộp hai `Product` trùng nhau là **workflow tầng app**, không phải method, vì đụng hai root |

`Variant` có `VariantID` (khách đặt *đúng* variant này, người đi mua mua *đúng* variant này)
nhưng không có vòng đời riêng. Đó là định nghĩa "entity con".

**Luật thứ 4 là luật hay bị vi phạm nhất**, và ở đây nó quyết định cả cấu trúc hệ thống:

```
   ❌ SAI — gộp làm một, đúng phản xạ Doctrine

        CustomerOrder
          └── PurchaseTask         ← nhét vào cùng aggregate

   Khách đã trả cọc. Người đi mua báo hết size. Hai việc đó xảy ra cách nhau
   nhiều giờ, ở hai nơi, do hai người. Ép chung một transaction là ép hai
   nhịp sống khác nhau vào một cái khoá.

   ✅ ĐÚNG — hai aggregate, hai context, nối bằng event

        ORDERING                          PROCUREMENT
        CustomerOrder  ──deposit_paid──▶  PurchaseTask
                       ◀─purchase_*────
```

**Quy tắc chọn kích thước aggregate: càng nhỏ càng tốt.** Aggregate to là khoá nhiều dòng, là
tắc nghẽn khi nhiều người dùng cùng lúc.

**Ngoại lệ có chủ đích.** `logisticsapp.AddParcelHandler` và `ShipBatchHandler` đụng **hai**
aggregate (`ConsolidationBatch` và `Parcel`) trong một transaction. Lý do ghi ngay trong
comment: trạng thái của parcel (`batched`, `shipped`) là sổ sách suy ra từ batch, không phải
invariant bên nào phụ thuộc, và một vòng event chỉ để đồng bộ hai chữ đó chỉ mua thêm độ trễ.
Nới luật thì phải nói vì sao, ở đúng chỗ nới.

**So với Symfony.** Doctrine khuyến khích `OneToMany` hai chiều và `cascade: persist`, nên mọi
thứ dễ dính vào một đồ thị lớn. Ở đây `ProductRepository.Save` ghi cả variants cùng cha
(xoá rồi chèn lại), và không có cách nào lưu một `Variant` lẻ.

## 15. Domain Event — Sự kiện nghiệp vụ

**Là gì.** Ghi nhận **một việc đã xảy ra** trong nghiệp vụ. Tên luôn ở **thì quá khứ**.

**Vì sao cần.**

1. Để aggregate này báo cho aggregate khác mà **không phụ thuộc** vào nhau.
2. Để lưu lại **lịch sử**, thứ mà cột `status` không bao giờ cho biết.
3. Để thêm tính năng mới **không phải sửa code cũ**: chỉ thêm người nghe.

**Trong Portage.** Hạ tầng ở `internal/domain/shared/event.go`:

```go
type Event interface {
	EventName() string        // "ordering.deposit_paid" — tên trên dây, không đổi khi đổi tên kiểu Go
	OccurredAt() time.Time
}

type Events struct{ pending []Event }     // sổ ghi mà aggregate root NHÚNG vào

func (e *Events) Record(ev Event)         // aggregate gọi trong method đổi trạng thái
func (e *Events) PullEvents() []Event     // tầng app gọi SAU KHI save; trả ra và xoá sổ
```

Aggregate dùng nó bằng **embedding**, một dòng:

```go
type Product struct {
	shared.Events          // Product có sẵn Record / PullEvents
	...
}

func (p *Product) Publish(now time.Time) error {
	// ... năm điều kiện ...
	p.status = ProductStatusPublished
	p.Record(ProductPublished{ID: p.id, Merchant: p.merchant, Category: p.category,
		Name: p.name, Source: p.source, Price: p.price, Parcel: p.parcel, At: now})
	return nil
}
```

Chú ý: aggregate **không tự gửi** event đi. Nó chỉ **ghi lại**. Tầng app lấy ra và ghi vào
outbox sau khi lưu thành công (§27). Vòng đời một event, đúng thứ tự:

```
   1. p.Publish(now)          → bên trong gọi p.Record(ProductPublished{...})   [domain]
   2. repo.Save(ctx, p)       → ghi aggregate                                    [adapter]
   3. evs := p.PullEvents()   → lấy ra, sổ trống lại                             [app]
   4. outbox.Append(ctx, evs) → CÙNG transaction với bước 2                      [adapter]
   5. worker đọc outbox → publish → người nghe ở context khác                    [cmd/worker]
```

Bước 3 **phải sau** bước 2: phát event trước khi lưu là cách worker đi mua giày cho một đơn
chưa từng tồn tại. Test `TestRegisterMerchant_saveFailurePublishesNothing` canh thứ tự đó.

**Hai mươi event của một đơn**, theo đúng thứ tự log của `scripts/smoke.sh`:

```
catalog.merchant_registered → product_added → variant_added → product_measured → product_published
pricing.quote_issued → quote_accepted
ordering.order_placed → deposit_paid
procurement.purchase_task_opened → purchase_confirmed
ordering.order_purchased                                          ← điểm không thể quay đầu
logistics.parcel_expected → parcel_received → batch_opened → batch_closed → batch_shipped
ordering.order_shipped → balance_paid → order_delivered
```

Cả hệ có 37 event. Guard `TestEncode_contractCoversEveryDomainEvent` đọc mọi file `events.go`
bằng `go/ast` và đỏ nếu một event chưa có mapper trong `eventcodec`.

**So với Symfony.** Giống `EventDispatcher` + `EventSubscriber`, nhưng khác ở chỗ **domain
event là một phần của model nghiệp vụ**, không phải cơ chế kỹ thuật. Nó xuất hiện trong ngôn
ngữ chung: bạn nói *"khi khách trả cọc thì…"*, và trong code có đúng `DepositPaid`. Cặp
`Record` / `PullEvents` chính là `private array $domainEvents` + `pullDomainEvents()` bạn
thấy trong các entity Doctrine làm DDD, được một `postFlush` listener rút ra đẩy vào Messenger.

## 16. Repository — Kho chứa

**Là gì.** Một **interface** cho phép domain lấy và lưu aggregate mà **không biết** dữ liệu
nằm ở đâu.

**Điểm mấu chốt: chiều phụ thuộc bị đảo ngược.** Domain khai báo cái nó cần; adapter cài đặt.

**Trong Portage.** `internal/domain/ordering/repository.go`:

```go
type OrderRepository interface {
	ByID(ctx context.Context, id OrderID) (*CustomerOrder, error)          // ErrOrderNotFound
	ByQuote(ctx context.Context, quote shared.ID) (*CustomerOrder, error)  // một quote, một đơn
	Save(ctx context.Context, o *CustomerOrder) error
}
```

Và hai adapter cài nó: `internal/adapter/memory/` (map trong RAM, cho test và `go run` không
cờ) và `internal/adapter/postgres/` (pgx, cho `-dsn`, CI, production). Mỗi adapter có một
dòng khẳng định compile-time rằng nó thoả port:

```go
var _ ordering.OrderRepository = (*OrderRepo)(nil)
```

**Ba luật:**

1. Repository làm việc với **aggregate root**, không phải từng bảng. Không có
   `VariantRepository` trong catalog.
2. Interface ở **domain**, cài đặt ở **adapter**.
3. Không có `findByStatusAndDateAndCustomer(...)`. Danh sách và bộ lọc cho màn hình là việc
   của **read model** (§24).

Ba điểm đáng để ý trong cách cài:

- **`Save` không phát event.** Tầng app gọi `PullEvents()` *sau khi* `Save` trả về, rồi ghi
  outbox **cùng transaction** (§27).
- **Aggregate toàn field private, vậy Postgres dựng lại thế nào?** Bằng **memento**:
  `Snapshot()` trả trạng thái đi ra, `FromSnapshot()` dựng lại đi vào, không phát event, và từ
  chối hình dạng không thể có (published mà không có variant). `internal/domain/catalog/snapshot.go`.
- **Repository không mở transaction.** `UnitOfWork.InTx` mở, bỏ `tx` vào `ctx`; repo lấy ra
  bằng `db(ctx)`. Cùng `ctx` là cùng transaction với outbox, không ai phải truyền `tx` qua tham số.

**So với Symfony.** `BookingRepository extends ServiceEntityRepository` **kế thừa từ
Doctrine**: domain biết Doctrine tồn tại. Ở đây ngược lại. `Snapshot()` đóng vai
`toArray()`, `FromSnapshot()` đóng vai `fromArray()`, và **chỉ repository** gọi chúng, thay cho
việc Doctrine ghi thẳng vào private field bằng Reflection.

## 17. Domain Service — Dịch vụ nghiệp vụ

**Là gì.** Nơi đặt logic nghiệp vụ **không thuộc về riêng một aggregate nào**.

**Khi nào cần.** Khi một quy tắc liên quan tới nhiều value object hoặc aggregate, hoặc là một
phép tính thuần không có trạng thái.

**Trong Portage.** Hai domain service, hai hình dạng, và hình dạng khác nhau là có lý.

**`pricing.Calculate` là một hàm thuần** (`internal/domain/pricing/calc.go`): từ
`QuoteInputs` (listing, profile, lane, tỷ giá, policy) ra `Breakdown`. Không clock, không
repository, không trạng thái, nên kiểm được từng dòng bằng bảng tính:

```
chargeable = lane.ChargeableWeight(parcel)     parcel = đã đo; không thì hộp mặc định của ngành hàng → Estimated = true
subtotal   = item + item×tax + freight(class, chargeable) + surcharge + duty        (USD)
total      = subtotal×fx + max(subtotal×fx × margin%, floor)                        (VND)
deposit    = total × deposit%
```

Nó không thuộc `Quote` (Quote chỉ *giữ* kết quả), không thuộc `ShippingLane` (lane không biết
policy), không thuộc `Listing` (listing không biết lane). Ba value object góp mặt, không ai sở
hữu phép tính. Là **hàm** chứ không phải interface vì chỉ có một cách tính.

**`logistics.FreightAllocator` là một interface** (`internal/domain/logistics/allocator.go`):

```go
type FreightAllocator interface {
	Allocate(freight shared.Money, items []BatchItem) ([]Allocation, error)
}

type ByChargeableWeight struct{ Divisor int64; Step shared.Weight }   // cách chia số 1
```

Chia cước một thùng cho N đơn: theo cân tính phí? theo thể tích? theo giá trị? **Mỗi cách
thiệt cho một nhóm khách khác nhau.** Đó là quyết định nghiệp vụ thật, có thể đổi, nên là
interface. `ConsolidationBatch.Ship(freight, alloc, now)` nhận allocator làm **tham số**; batch
không biết luật chia. Bên dưới, `shared.Allocate` chia sàn rồi phát cent thừa theo largest
remainder, để 87.50 USD chia cho 2500 g và 5000 g ra **29.17 + 58.33 = 87.50** đúng từng cent.

**Phân biệt với Application Service:**

| Domain Service | Application Service |
|---|---|
| chứa **quy tắc nghiệp vụ** | chỉ **điều phối** |
| nằm trong `internal/domain/` | nằm trong `internal/app/` |
| không biết DB, không biết transaction | mở transaction, gọi repository, ghi outbox |
| `pricing.Calculate`, `FreightAllocator` | `pricingapp.IssueQuoteHandler` gom đầu vào rồi **gọi** `Calculate` |

**So với Symfony.** "Service" bên Symfony thường là cả hai loại trộn vào một class. Ở đây tách
bằng thư mục, và guard 6 (§21) bảo đảm loại thứ nhất không import được thứ gì của loại thứ hai.

## 18. Factory

**Là gì.** Nơi tập trung logic **tạo** aggregate khi việc tạo phức tạp.

**Trong Portage.** Không có class `Factory`. Trong Go, một hàm ở cấp package là đủ, và mỗi
aggregate root có đúng một hàm như vậy, là **cửa duy nhất** để tạo ra nó:

```go
func RegisterMerchant(d MerchantDetails, now time.Time) (*Merchant, error)
func AddProduct(d ProductDetails, now time.Time) (*Product, error)
func IssueQuote(in QuoteInputs, now time.Time) (*Quote, error)
func PlaceOrder(d OrderDetails, now time.Time) (*CustomerOrder, error)
func OpenTask(d TaskDetails, now time.Time) (*PurchaseTask, error)
func ExpectParcel(d ParcelDetails, now time.Time) (*Parcel, error)
func OpenBatch(lane string, now time.Time) (*ConsolidationBatch, error)
```

Tên hàm là **động từ nghiệp vụ**, không phải `NewProduct`. Mỗi hàm validate từng field bắt
buộc, sinh ID, đặt trạng thái đầu, và ghi event đầu tiên. Struct `XxxDetails` thay cho named
arguments của PHP (quy ước 10, SETUP.md §6).

## 19. Policy / Specification — *chưa tách thành kiểu riêng*

**Là gì.** Đóng gói một **quy tắc quyết định** thành object riêng, để nó có tên và test được
độc lập, và để đổi luật không phải sửa aggregate.

**Trong Portage: chưa có, và đây là lý do.** Luật hoàn cọc hôm nay là một `switch` năm nhánh
trong `CustomerOrder.Cancel` (§26). Nó có tên, có bảng test riêng, và chỉ có **một** luật.
Tách ra một `RefundPolicy` interface bây giờ là thêm một tầng gián tiếp để chuẩn bị cho một
luật thứ hai chưa tồn tại.

Khi nào nên tách: ngày có luật hoàn thứ hai thật (ví dụ khách quen hoàn khác khách mới, hay
tuyến vận chuyển khác luật khác). Lúc đó hình dạng sẽ là:

```go
type RefundPolicy interface {
	RefundFor(status OrderStatus, deposit shared.Money) Refund
}
```

và `Cancel` nhận policy làm tham số, đúng như `ConsolidationBatch.Ship` nhận
`FreightAllocator` hôm nay. `FreightAllocator` (§17) chính là ví dụ của pattern này đã được
tách, vì ở đó nhiều cách chia là chuyện có thật ngay từ đầu.

---

# PHẦN IV — Kiến trúc

## 20. Ports & Adapters (Hexagonal Architecture)

**Là gì.** Kiến trúc đặt domain ở **giữa**; mọi thứ bên ngoài (DB, HTTP, mô hình ngôn ngữ, web
bán hàng) cắm vào qua **interface**.

- **Port** là interface do bên cần khai báo: *"tôi CẦN gì"*.
- **Adapter** là cài đặt cụ thể: *"làm BẰNG gì"*.

```
        ┌─────────────────────────────────────────────┐
        │  APP + DOMAIN  (Go thuần)                    │
        │                                              │
        │  type ListingExtractor interface {           │   ← PORT: tầng app khai báo
        │      Extract(ctx, url) (ListingDraft, error) │     cái nó cần
        │  }                                           │
        └───────────────────┬──────────────────────────┘
                            │ thoả interface (ngầm, không có "implements")
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
   ┌──────────────┐  ┌──────────────┐  ┌───────────────────┐
   │ openai.Client│  │ openai.Fake  │  │ openai.Unavailable│   ← ADAPTER
   │ gọi API thật │  │ dev + test   │  │ không key → 503   │
   └──────────────┘  └──────────────┘  └───────────────────┘
```

**Trong Portage.** Mọi port và adapter đang có:

| Port (interface) | Khai báo ở | Adapter đã cài |
|---|---|---|
| `MerchantRepository`, `CategoryRepository`, `ProductRepository` | `domain/catalog/repository.go` | `adapter/memory`, `adapter/postgres` |
| `OrderRepository`, `QuoteRepository`, `TaskRepository`, `ParcelRepository`, … | `domain/<context>/repository.go` | cùng hai adapter trên |
| `Clock` | `app/ports.go` | `platform/clock.System` (thật), `clock.Fixed` (test) |
| `UnitOfWork`, `Outbox` | `app/ports.go` | `memory`, `postgres` |
| `procurement.MerchantACL` — *"đặt hàng ở shop này"* | `domain/procurement/repository.go` | `merchant.Manual` (một con người), `merchant.Router` (chọn theo shop) |
| `catalogapp.ListingExtractor` — *"trang này bán cái gì"* | `app/catalog/extract.go` | `openai.Client`, `openai.Fake`, `openai.Unavailable` |
| `logistics.FreightAllocator` — *"chia cước thế nào"* | `domain/logistics/allocator.go` | `ByChargeableWeight` |
| `reportingapp.OrderSummaryRepository`, `ProductWorklistRepository` | `app/reporting/` | `memory`, `postgres` |
| `auth.Verifier` — *"token này là ai"* | `platform/auth/auth.go` | `auth.Static` (RAM), `postgres.TokenRepo` |
| `auth.Issuer`, `auth.Registry` — *"cắt chìa"*, *"chìa nào đang có"* | `platform/auth/` | cùng hai adapter trên |

**Quy tắc chỗ đặt: interface nằm ở nơi CẦN nó**, không nằm ở nơi cài. Domain cần repository
thì interface ở domain. Tầng app cần đồng hồ thì `Clock` ở app. Adapter chỉ "tình cờ có đủ
method"; Go không có từ khoá `implements`, nên mỗi adapter có một dòng khẳng định:

```go
var _ catalog.MerchantRepository = (*MerchantRepo)(nil)   // adapter/memory/merchant_repo.go — interface thêm method → dòng này không compile
```

**Vì sao đây là ý tưởng quan trọng nhất về AI trong project này.** Chỗ nhiều project "AI +
DDD" làm sai là nhét lời gọi OpenAI vào giữa business logic. Đúng phải là ngược lại: **mô hình
ngôn ngữ chỉ là một loại "database" khác**, một thứ ở ngoài, chậm, hay lỗi, tốn tiền, phải
cắm qua interface. Port `ListingExtractor` trả **bốn string thô**, không trả `shared.Money`,
để domain vẫn là bên quyết định thế nào là hợp lệ (§22).

**Bằng chứng đã hoạt động:** 278 trên 312 test chạy **không cần gì cả**, không Docker, không
mạng, không API key, kể cả những test tầng app đi qua thật handler → domain → repository →
outbox. Vì repository và outbox là adapter in-memory cắm vào port.

**So với Symfony.** `services.yaml` với `App\Port\ClockInterface: '@App\Adapter\SystemClock'`.
Khác ở chỗ Go không có container: `internal/platform/wire/wire.go` cắm tay từng cái, và cắm
sai là không compile.

## 21. Dependency Rule — Luật chiều phụ thuộc

**Là gì.** Mũi tên phụ thuộc **chỉ được hướng vào trong**. Domain không bao giờ trỏ ra ngoài.

```
   cmd  ──▶  app  ──▶  domain
                ▲
   adapter ─────┘         domain KHÔNG trỏ ra ngoài, BAO GIỜ
```

**Trong Portage.** Luật ghi ngay đầu `internal/domain/shared/money.go`:

```go
// RULES FOR THIS PACKAGE (and every package under internal/domain):
//   - It may NOT import a database driver, an HTTP framework, or any SDK.
//   - It may NOT import another bounded context.
//   - Everything here is immutable: methods return new values, never mutate.
```

**Cách Go ép buộc.** Thư mục tên `internal/` được **chính compiler** bảo vệ: package ngoài
module không import được. Không cần reviewer nhắc.

**Kiểm tra bằng lệnh**, chạy được ngay:

```bash
go list -f '{{range .Imports}}{{println .}}{{end}}' ./internal/domain/... | sort -u
```

Kết quả phải chỉ có thư viện chuẩn (`errors`, `fmt`, `math/big`, `strconv`, `strings`,
`time`, `context`, `slices`…) và đúng **một** thư viện ngoài: `github.com/google/uuid`.

**Allowlist là gì.** Luật "domain không import gì ngoài" nói cho đúng là "không import
**framework, driver, SDK**", những thứ có side effect, biết mạng, biết đĩa. Một thư viện tiện
ích thuần như `google/uuid` (sinh số ngẫu nhiên và format) không phá luật, giống `symfony/uid`
không làm entity Symfony "bẩn". Nhưng phải **ghi rõ**: allowlist nằm trong
`.github/workflows/ci.yml` và trong guard 6 của `decisions_test.go`. Thêm gì phải sửa cả hai.

Hai lớp canh, không lớp nào là lời hứa:

| Lớp | Ở đâu | Đỏ khi |
|---|---|---|
| CI | `.github/workflows/ci.yml`, bước "luật import" | `go list` in ra một dòng ngoài allowlist |
| Test | `TestDecision_domainImportsOnlyStdlibAndAllowlist` | cùng điều kiện, nhưng chạy ngay trên máy bạn với `go test ./internal/domain/` |

**So với Symfony.** Đây là việc `deptrac` làm. Ở đây không cần tool riêng vì `go list` có sẵn.

> Đừng dùng `go list -deps` để kiểm tra việc này. Nó liệt kê cả dependency **gián tiếp**
> (`runtime`, `os`, `syscall`… do `fmt` kéo theo) và làm bạn tưởng domain bẩn. `.Imports` mới
> là import trực tiếp.

## 22. Anti-Corruption Layer (ACL) — Lớp chống ăn mòn

**Là gì.** Lớp **dịch** giữa domain của mình và một hệ thống bên ngoài mà mình không kiểm
soát.

**Vì sao cần.** Model của Nike (style, colorway, size US) mà rò vào domain thì domain sẽ dần
biến dạng theo họ. Khi thêm Adidas với model khác hẳn, mình kẹt.

```
   ┌────────────────┐        ┌─────────────┐       ┌──────────────┐
   │ DOMAIN CỦA MÌNH│◀──────▶│     ACL     │◀─────▶│  Nike, TNF,  │
   │ PurchaseTask,  │  ngôn  │  lớp dịch   │  ngôn │  Sony, ...   │
   │ PurchaseReceipt│  ngữ   │             │  ngữ  │  style,      │
   │                │  MÌNH  │             │  HỌ   │  colorway    │
   └────────────────┘        └─────────────┘       └──────────────┘
```

**Trong Portage.** Hai ACL, hai nỗi lo khác nhau, cùng một hình dạng:

```
   procurement.MerchantACL       "đặt hàng ở shop này"    rủi ro: API người ta đổi, người ta chết
   catalogapp.ListingExtractor   "trang này bán cái gì"   rủi ro: mô hình ngôn ngữ ĐOÁN, và đoán rất tự tin
```

### ACL thứ nhất: `MerchantACL` — `internal/domain/procurement/repository.go`

```go
type MerchantACL interface {
	// ok=false với reason nghĩa là shop từ chối (hết hàng); err nghĩa là không hỏi được.
	Purchase(ctx context.Context, task *PurchaseTask) (receipt PurchaseReceipt, ok bool, reason string, err error)
}
```

Adapter đầu tiên là **một con người**. `adapter/merchant/manual.go` không mua được và nói
thẳng bằng `ErrManualPurchase`: người đi mua tự vào shop rồi báo lại qua
`POST /purchase-tasks/{id}/confirm`. Nhờ adapter này, use case có port để gọi **ngay bây
giờ**, và có bốn nhánh cho bốn câu trả lời:

| ACL trả | `OpenTaskHandler` làm |
|---|---|
| `ErrManualPurchase` | task ở `open` cho người mua (đường hôm nay) |
| `ok = true` + biên nhận | `Confirm` ngay, cùng transaction |
| `ok = false` + lý do | `Fail` ngay |
| lỗi khác (shop sập) | trả lỗi, không lưu gì, relay thử lại sau |

Một fake ACL mười dòng trong test chứng minh cả bốn nhánh **trước khi** shop nào có API. Ngày
Nike cho API: thêm `adapter/merchant/nike.go`, thêm một dòng vào `merchant.Router` ở `wire`.
Use case không thêm một `if`.

`merchant.Router` chọn adapter theo shop và **chính nó** cũng là một `MerchantACL`, nên không
ai bên trên biết nó tồn tại. Shop không có trong map thì rơi về `Manual`. Hai lựa chọn còn lại
đều tệ hơn: một `if` trong use case sẽ mọc thêm một nhánh mỗi shop mãi mãi; một ACL biết mọi
shop thì shop nào chết cũng thành mọi shop chết.

### ACL thứ hai: `ListingExtractor` — `internal/app/catalog/extract.go`

```go
type ListingDraft struct{ Name, Price, Currency, CategoryHint string }   // BỐN string

type ListingExtractor interface {
	Extract(ctx context.Context, url catalog.SourceURL) (ListingDraft, error)
}
```

**Bốn string, không phải `shared.Money`.** Nếu port trả `Money` thì port đang quyết cái gì
hợp lệ, mà đó là việc của domain. Chuỗi của máy đi qua đúng cái cửa mà chuỗi người gõ tay đi
qua (`shared.ParseMoney`), nên model trả `"about $150"` chết ngay ở biên, không phải ba bảng
sau.

Và luật quan trọng nhất: **máy được TẠO nháp, không được XÁC NHẬN.** Provenance của nháp là
`SourcedByFeed`, và `Product.Publish()` từ chối listing chưa `Verified()`. Nên không thứ gì mô
hình bịa ra thành được giá bán trước khi một con người bấm confirm-listing. Tính năng mới
**không nới** luật cũ.

Ba adapter cho một port, và cái thứ ba mới là bài học:

```
   openai.Client       gọi API thật: net/http thuần, json_schema strict, timeout
   openai.Fake         dev và test: go run ./cmd/api chạy không cần key, không tốn tiền
   openai.Unavailable  Postgres mà không có OPENAI_API_KEY → 503
```

> Chế độ Postgres **tuyệt đối không** rơi về `Fake`. Rơi về thì API trả 201, một sản phẩm nháp
> tồn tại với tên và giá **bịa**, và dấu hiệu duy nhất có gì sai là các con số toàn hư cấu.
> **Tính năng tắt phải trông như đang tắt.**

Hai điều adapter này cố tình không làm: không dùng SDK (thêm dependency, thêm kiểu riêng
muốn rò lên tầng trên), và **không cào trang** (điều khoản của shop cấm; chỉ URL được gửi
cho model).

**So với Symfony.** `MerchantGatewayInterface` với `ManualGateway` là implementation "null
object" bạn đăng ký trong `services.yaml` trước khi có API thật.

## 23. Đa thương hiệu = dữ liệu, không phải code

**Là gì.** Yêu cầu *"mở rộng cho brand khác, không chỉ giày"* có một hệ quả kỹ thuật rất cụ
thể:

> **Từ "Nike" không được xuất hiện một lần nào trong `internal/domain/`.** Nike là **một dòng
> trong bảng**, không phải một nhánh `if`. Thêm Adidas là thêm một dòng dữ liệu, không phải
> deploy.

**Trong Portage.** `internal/domain/catalog/merchant.go`:

```go
type Merchant struct {
	shared.Events
	id       MerchantID
	name     string          // "Nike", "The North Face", "Sony" — chỉ là DỮ LIỆU
	site     Hostname
	currency shared.Currency
	freeShip FreeShipping    // never / over $50 / always — value object 3 trạng thái
	sourcing []SourcingMode  // operator | customer | feed — ai được gửi link cho shop này
	status   MerchantStatus  // active | suspended
}
```

Và ngành hàng nói món hàng **là** gì, `catalog/category.go`:

```go
type CategoryPolicy struct {
	code         CategoryCode   // footwear | apparel | electronics — khoá tự nhiên
	estimate     shared.ParcelSpec  // cân nặng + hộp ước lượng → báo giá khi chưa đo thật
	restrictions []Restriction  // battery, liquid, supplement…
}
```

`CategoryPolicy` **không có** thuế hay mã HS, và không có số chia thể tích. Đó là dữ liệu của
**tuyến vận chuyển** (`pricing.ShippingLane`): cùng đôi giày, đi hai tuyến, dòng thuế khác
hẳn. Guard 4 `TestDecision_catalogKnowsNothingAboutTax` canh việc chúng không quay lại.

`Restriction{Battery}` là thật: tai nghe Sony có pin lithium, không đi được mọi tuyến bay.
Nếu domain không biết khái niệm này, ngày đầu bán đồ điện tử là ngày hàng kẹt ở kho.

Guard 3 `TestDecision_noBrandNamesInCode` đỏ nếu "Nike", "Adidas"… xuất hiện trong code (comment
thì được).

## 24. CQRS — Tách đường ghi và đường đọc

**Là gì.** Tách **đường ghi** và **đường đọc** thành hai model khác nhau.

```
   GHI (Command)                    ĐỌC (Query)
   ────────────                     ───────────
   Aggregate                        Read model / view
   Bảo vệ invariant                 Chỉ để hiển thị
   Chuẩn hoá                        Bẹt, phi chuẩn hoá
   Một aggregate mỗi transaction    Ráp từ nhiều nguồn
```

**Vì sao Portage cần.** Màn hình "đơn của tôi" cần bảy thứ, và chúng thuộc **bốn** context:
tên sản phẩm (catalog), trạng thái đơn và đã cọc chưa (ordering), mã đơn ở shop (procurement),
thùng tới đâu (logistics). Không aggregate nào trả lời được, và không aggregate nào **được
phép** đọc bảng của context khác.

Ba lối thoát, và vì sao hai cái đầu sai:

| Cách | Vì sao không |
|---|---|
| JOIN bốn bảng lúc đọc | biến bốn bounded context thành **một cục** không tách nổi: đổi schema của catalog làm hỏng màn hình của ordering |
| Gọi API nội bộ của nhau | một context chết là màn hình chết, và bốn cuộc gọi cho một cái list |
| **Một bảng phẳng, ghi bằng event** | đúng |

**Trong Portage.** `internal/app/reporting/` là **package duy nhất không có
`internal/domain/…` tương ứng**. Đó không phải thiếu sót, đó là định nghĩa:

> **Read model không có invariant.** Không có gì trong nó có thể "sai" theo nghĩa nghiệp vụ,
> vì nó không **quyết định** cái gì. Nó chỉ nhớ lại điều mà năm context đã tuyên bố.

Nên `OrderSummary` là struct field exported hết, không method, không validate. Hai read model:

| Bảng | Cho màn hình | Nghe từ |
|---|---|---|
| `order_summaries` | đơn của tôi (khách), hàng đợi việc (nhân viên) | cả 5 context, 12 dòng `Subscribe` |
| `product_worklist` | còn việc gì phải làm với món khách gửi, bước tiếp theo là gì | catalog, 5 event |

Ba quyết định, mỗi cái là một bài học:

1. **Chỉ event được ghi vào nó.** Nó không đọc bảng ghi của context nào.
2. **Idempotent và chịu sai thứ tự.** Relay là at-least-once và không bảo đảm thứ tự giữa các
   luồng, nên `batch_shipped` có thể tới trước `order_placed`. Cách giải: **bất kỳ event nào
   cũng được TẠO dòng** (cái "khung"), và chỉ `order_placed` được đặt `status` đầu tiên.
   Đoán một trạng thái là nói dối với khách đang nhìn màn hình.
3. **Hai trục, không gộp.** `status` do ordering sở hữu (tiền và việc mua); `tracking` do
   logistics sở hữu (thùng ở đâu). `received` không phải một trạng thái của đơn.

Với `product_worklist`, câu "còn thiếu bước nào" được quyết **đúng một lần** trong
`WorklistItem.NextStep()`, không phải mỗi màn hình một bản.

**So với Symfony.** Bên Symfony hay làm bằng một Doctrine query JOIN 5 bảng trong một
`Repository::findForDashboard()`. Chạy được, cho tới ngày muốn tách bundle.

## 25. Eventual Consistency — Nhất quán sau cùng

**Là gì.** Chấp nhận rằng hai aggregate **không đồng bộ tức thì**, mà sẽ khớp nhau sau một lúc
qua event.

**Vì sao buộc phải chấp nhận.** Không có transaction nào bao được cả *"tiền khách đã thu"* và
*"đơn đã đặt ở Nike"*. Nike không tham gia transaction của mình.

```
   t=0   khách trả cọc          → CustomerOrder: deposited
   t=1   event deposit_paid nằm trong outbox
   t=2   worker chuyển, procurement mở PurchaseTask: open
   t=3   người đi mua           → confirmed hoặc failed

   Trong khoảng t=0 → t=3, hệ thống KHÔNG nhất quán. Và điều đó CHẤP NHẬN ĐƯỢC.
   Cái không chấp nhận được là giả vờ nó nhất quán.
```

**Trong Portage.** Thấy được bằng mắt: `POST /products/{id}/publish` → 204, rồi **ngay lập
tức** `POST /quotes` → 404 `listing_not_found`. Sản phẩm đã đăng bán, nhưng pricing chưa nghe
vì worker chưa chạy. Chạy relay một lượt → `POST /quotes` → 201.

Mã lỗi `listing_not_found` vì thế có **hai nghĩa thật**: chưa publish, hoặc đã publish mà relay
chưa chạy. Tài liệu API nói thẳng điều đó thay vì hứa một thứ hệ thống không giữ được. Cùng
hình dạng: `quote_not_accepted` khi đặt đơn, `variant_unknown` khi variant chưa relay tới.
Trang khách xử lý bằng cách **tự thử lại** vài nhịp, không báo đỏ.

Test vàng `wholeFlow` trong `wire_test.go` bắn `POST /quotes` ba lần để chứng minh khoảng
không nhất quán này **tồn tại và kết thúc**. Ở chế độ in-memory `cmd/api` chạy relay mỗi
200 ms nên khoảng đó ngắn tới mức khó thấy, nhưng vẫn có.

## 26. Saga & Compensation — Chuỗi bù trừ

**Là gì.** Chuỗi các bước ở nhiều aggregate hoặc context. Khi một bước hỏng, **bù trừ** các
bước trước bằng **hành động nghiệp vụ**, không phải rollback DB.

**Vì sao không rollback được.** Không thể "rollback" việc đã quẹt thẻ, cũng không "rollback"
đơn đã đặt ở Nike.

**Saga chính của Portage:**

```
  khách cọc 50 %  ─▶  đi mua ở shop  ─┬─ thành công ─▶  về kho Mỹ  ─▶  gom lô, ship  ─▶  khách trả nốt
                                      │
                                      └─ hết size ───▶  BÙ TRỪ: huỷ đơn, hoàn 100 % cọc
```

**Điểm không thể quay đầu** là quy tắc quan trọng nhất của mô hình cọc 50 %, và nó là một
method có test, `internal/domain/ordering/order.go`:

```go
func (o *CustomerOrder) Cancel(reason string, now time.Time) (Refund, error) {
	// ... reason không được rỗng ...
	var refund Refund
	switch o.status {
	case StatusAwaitingDeposit:
		refund = Refund{amount: shared.Zero(o.total.Currency())}          // chưa trả gì
	case StatusDeposited, StatusPurchaseFailed:
		refund = Refund{amount: o.deposit}                                // chưa mua → hoàn ĐỦ
	case StatusPurchased, StatusInTransit:
		refund = Refund{amount: shared.Zero(o.total.Currency()), forfeited: true}   // đã mua → MẤT cọc
	case StatusDelivered:
		return Refund{}, fmt.Errorf("cancel %s: %w", o.id, ErrAlreadyDelivered)
	default:
		return Refund{}, fmt.Errorf("cancel %s: %w", o.id, ErrOrderCancelled)
	}
	o.status = StatusCancelled
	o.refund = refund
	o.Record(OrderCancelled{ID: o.id, Reason: reason, Refund: refund.amount, Forfeited: refund.forfeited, At: now})
	return refund, nil
}
```

| Trạng thái khi huỷ | Hoàn | `Forfeited` | Vì |
|---|---|---|---|
| `awaiting_deposit` | 0 | false | chưa trả gì |
| `deposited` | **2 696 860 ₫** | false | chưa mua, tiền còn trong tay mình |
| `purchase_failed` | **2 696 860 ₫** | false | mình không mua được, lỗi của mình |
| `purchased` | 0 | **true** | đã ứng 100 % mua hàng cho khách |
| `in_transit` | 0 | **true** | hàng đang bay |
| `delivered` | không huỷ được | | `ErrAlreadyDelivered` |

Sự kiện `procurement.purchase_confirmed` **chính là ranh giới**. `ordering.Reactor` nghe nó và
gọi `ConfirmPurchase`, chuyển `deposited → purchased`. Trước dòng đó, rủi ro là của mình. Sau
dòng đó, rủi ro chuyển sang khách. Toàn bộ mô hình cọc 50 % tồn tại vì ranh giới này.

**Nửa nào đã có, nửa nào cố tình không.** Saga có hai nửa: **bù trừ** (đã có: `Cancel` với
bảng trên) và **điều phối tự động** (cố tình không). Sau `purchase_failed`, hệ thống **không
tự huỷ** đơn: một con người quyết định hoàn tiền hay đổi màu khác, qua
`POST /orders/{id}/cancel`. Tự động hoá một quyết định có tiền thật đằng sau là việc cần dữ
liệu vận hành thật trước.

Một chi tiết đáng đọc hai lần: `Reactor.alreadyDone` nuốt `ErrNotDeposited` khi
`purchase_confirmed` tới lần thứ hai. Đơn **đã** `purchased`, aggregate từ chối đúng, nhưng
với relay đó không phải lỗi cần thử lại, mà là *"đã xong rồi"*. Reactor là chỗ duy nhất biết
phân biệt "không làm được" với "đã làm rồi".

**So với Symfony.** Messenger với nhiều handler nối nhau bằng message. Khác ở chỗ luật bù trừ
nằm trong aggregate (`Cancel`), không nằm trong handler.

## 27. Outbox Pattern

**Là gì.** Cách đảm bảo *"lưu dữ liệu"* và *"gửi event"* không bị lệch nhau.

**Vấn đề nó giải quyết:**

```
   ❌ Lưu DB thành công → chưa kịp gửi event → server chết
      → đơn đã trả cọc nhưng KHÔNG AI đi mua. Khách mất tiền.
```

**Cách làm.** Ghi event vào **một bảng, cùng transaction** với dữ liệu. Một worker riêng đọc
bảng đó và gửi đi.

```
   BEGIN
     INSERT INTO orders ... status = 'deposited'
     INSERT INTO outbox (event_name, payload) VALUES ('ordering.deposit_paid', ...)
   COMMIT                    ← hai việc, một transaction, không thể lệch

   [worker riêng]  SELECT … WHERE sent_at IS NULL → publish → UPDATE sent_at
```

**Trong Portage.** Đủ hai nửa, mỗi nửa có test:

| Nửa | Code | Test |
|---|---|---|
| ghi cùng transaction | `postgres.Outbox.Append` dùng `db(ctx)` → cùng `tx` với `Save` | `TestOutbox_appendedInTheSameTransactionAndReadBack` |
| Save hỏng → không event | thứ tự trong handler (§15 bước 2 và 3) | `TestRegisterMerchant_saveFailurePublishesNothing` |
| Save xong, outbox hỏng → **rollback** cả hai | `postgres.UnitOfWork.InTx` | `TestUnitOfWork_rollsBackTheSaveWhenTheOutboxFails` |
| worker đọc → publish → đánh dấu | `worker.Relay.RunOnce`, `cmd/worker` | 6 test relay: thứ tự, retry, batch |

**At-least-once** là lựa chọn có chủ đích: publish trước, ghi `sent_at` sau. Sập giữa hai bước
thì lượt sau gửi **lại**. Không bao giờ mất; có thể thấy hai lần. Hệ quả bắt buộc: **mọi
subscriber phải idempotent** (upsert theo khoá, hoặc nuốt đúng sentinel "đã làm rồi"). Ngược
lại (đánh dấu trước, publish sau) là mất event, và một `deposit_paid` mất là một đơn không ai
mua.

Payload trong outbox là **JSON viết tay theo hợp đồng** (`adapter/eventcodec`), không phải
`json.Marshal(event)`. Vì bên nghe ở context khác không load được aggregate (guard 7), payload
là **tất cả** nó có. Để `json.Marshal` suy shape từ struct thì đổi tên field là đổi wire
format mà vẫn compile.

**So với Symfony.** `postFlush` listener rút domain event ra đẩy vào Messenger với transport
`doctrine://` chính là outbox. Ở đây viết tay khoảng 120 dòng để thấy hết vòng đời một
message.

## 28. Quote vs Actual — Ước tính và Thực tế

**Là gì.** Ghi **cả hai** con số, giá đã báo và giá thật đã trả, rồi đối soát. Không ghi đè.

```
    Quote (báo giá)              Actual (thực tế)
    ═══════════════              ════════════════
    ước tính                     sự thật
    trước khi mua                sau khi cân, sau khi trả tiền
    CÓ HẠN dùng                  vĩnh viễn
    khách nhìn thấy              kế toán nhìn thấy
              │                        │
              └────────┬───────────────┘
                       ▼
              CHÊNH LỆCH (variance) = LỜI hoặc LỖ của mình trên đơn đó
```

**Vì sao đây là insight kinh tế lớn nhất của nghề mua hộ.** Ước tính nhẹ tay thì lỗ. Ước tính
nặng tay thì mất khách. Toàn bộ kinh tế học nằm ở khoảng chênh này, và lỗi phổ biến nhất là
lưu **một** con số "giá" rồi ghi đè khi biết số thật, mất sạch dữ liệu để biết mình lời hay lỗ.

**Trong Portage: nửa Quote.** `internal/domain/pricing/quote.go`:

```go
type Quote struct {
	shared.Events
	id        QuoteID
	product   shared.ID
	lane      LaneCode
	breakdown Breakdown      // BỨC ẢNH: class, cân tính phí, Estimated, từng dòng USD, tỷ giá, từng dòng VND
	issuedAt  time.Time
	expiresAt time.Time      // issuedAt + policy.TTL (48 giờ)
	status    QuoteStatus    // issued → accepted | expired
}
```

Ba điều bức ảnh này bảo đảm:

| Bảo đảm | Code |
|---|---|
| chụp **mọi** dữ kiện đã dùng | `Breakdown` giữ tỷ giá, cân tính phí, loại hàng, từng dòng tiền. Đổi tỷ giá ngày mai (`POST /fx`), quote hôm nay **không đổi một chữ**. Postgres lưu breakdown **từng cột**, không `jsonb`, để đối soát là câu `SELECT` |
| có hạn dùng | `Accept(now)` sau `expiresAt` chuyển sang `expired`, ghi event, **rồi từ chối**. Tầng app commit trạng thái mới rồi mới trả lỗi, vì rollback thì khách bấm lại được vào giá cũ |
| có hạn dùng **cả khi không ai đụng tới** | `Expire(now)` do một **sweep theo giờ** gọi (`cmd/worker -sweep`). Không có sweep thì quote không ai trả lời nằm `issued` mãi mãi: đúng luật khi đọc, sai khi đếm |
| biết mình đang đoán | `Breakdown.Estimated = true` khi cân bằng hộp mặc định của ngành hàng thay vì hộp đã đo |

**Nửa Actual.** `internal/domain/pricing/reconciliation.go`, một dòng mỗi đơn, tiền tuyến (USD),
dựng từ **ba** context:

```go
type Reconciliation struct {
	Order, Quote shared.ID

	QuotedGoods   shared.Money   // item + thuế + duty, như đã báo          ← ordering.order_placed (chép từ quote)
	QuotedFreight shared.Money   // cước + phụ thu, như đã báo

	ActualGoods   shared.Money   // shop thật sự tính bao nhiêu             ← procurement.purchase_confirmed
	ActualFreight shared.Money   // phần chia hoá đơn hãng bay cho đơn này  ← logistics.batch_shipped
}

func (r Reconciliation) Complete() bool            // đủ cả hai actual
func (r Reconciliation) Variance() (shared.Money, error)   // quoted − actual, chỉ khi Complete
```

```
quoted  163.22 + 25.00 = 188.22 USD    (giày 150 + thuế 13.22; cước 2,5 kg × 10)
actual  163.22 + 27.50 = 190.72 USD    (shop tính đúng; hãng bay tính 27.50 thay 25.00)
variance                 −2.50 USD     ← ÂM = mình chịu. Phí dịch vụ 500 000 ₫ đã che khoản này.
```

`variance` **luôn âm** qua nhiều đơn nghĩa là bảng giá cước đang sai, không phải xui. Bảng
`reconciliations` sinh ra để trả lời đúng câu hỏi đó, và đó cũng là lý do quyết định "một đơn
= một cái" được **hoãn** thay vì đổi: chưa có dữ liệu thật để biết đơn giản hoá đó có sai không.

**Sweep là loại use case thứ ba.** Mọi use case khác đều có người đẩy: một request HTTP, hoặc
một event của context khác. `Expire` thì không:

| Cái gì đẩy | Ví dụ |
|---|---|
| một con người | `POST /orders` |
| một context khác | `deposit_paid` → mở việc mua |
| **thời gian trôi** | quote quá 48 giờ |

Không ai phát được event `quote_got_old`, vì **không có gì xảy ra cả**. Nên phải có người đi
hỏi: `worker.Sweeper` gọi `ExpireQuotesHandler` theo nhịp. Đó cũng là lý do rõ nhất để `Clock`
là **port**: "quá hạn chưa" phải kiểm chứng được mà không đợi 48 giờ. Trong test,
`clk.Advance(49 * time.Hour)`. Sweep xử lý **mỗi quote một transaction**: lô 100 quote không
phải một invariant, nên không được là một đơn vị nguyên tử (§14).

**So với Symfony.** Không có pattern tương ứng trong framework. Đây là quyết định về **mô hình
dữ liệu**: hai bảng (`quotes`, `reconciliations`) thay cho một cột `price` bị ghi đè.

---

# PHẦN V — Ba vòng đời

## 29. Vòng đời một AGGREGATE

Năm bước, và thứ tự giữa bước ③ và ④ là chỗ dễ sai nhất.

```
 ①  SINH RA                    ②  THAY ĐỔI              ③  GHI SỰ KIỆN
 ─────────────                 ────────────             ──────────────
 RegisterMerchant(d, now)      m.Rename("X", now)       m.Record(ev)
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
 · SQL ở đây                         · ghi outbox CÙNG transaction với ④
```

**Code thật**, `internal/domain/catalog/merchant.go`:

```go
// ② + ③ — hành vi, không phải setter. Bên trong tự Record.
func (m *Merchant) Rename(name string, now time.Time) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName              // invariant từ chối
	}
	if name == m.name {
		return nil                       // không có gì xảy ra → KHÔNG có event
	}
	old := m.name
	m.name = name
	m.Record(MerchantRenamed{ID: m.id, From: old, To: name, At: now})
	return nil
}
```

Dòng `if name == m.name { return nil }` đáng nhớ: **sự kiện chỉ sinh ra khi có việc thật sự
xảy ra.** Đổi tên thành chính nó không phải một sự kiện.

| Luật | Vì sao | Test canh |
|---|---|---|
| ID sinh trong domain, không đợi DB | object phải có danh tính từ giây đầu | `TestRegisterMerchant_hasIdentityImmediately` |
| aggregate chỉ **ghi**, app mới **phát** | không publish thứ chưa lưu | `TestRegisterMerchant_recordsEventButDoesNotPublish` |
| không có việc gì xảy ra → không có event | event là *sự thật đã xảy ra*, không phải *lệnh* | `TestMerchant_renameRaisesEventOnlyWhenItChanges` |

## 30. Vòng đời một REQUEST — đi hết các tầng

```
   0. CỬA        adapter/http/auth.go       Bearer token → auth.Principal. Không token → 401, sai loại → 403.
        │                                   Bọc CẢ mux: route thêm ngày mai đã ở sau cửa.
        ▼
   1. HTTP        adapter/http/              giải mã JSON (DisallowUnknownFields) · chuẩn hoá "150,50" → "150.50"
        │                                   theo Accept-Language · map lỗi → 400/404/409 bằng MỘT bảng.
        │                                   KHÔNG có quy tắc nghiệp vụ.
        ▼
   2. USE CASE    app/<context>/             lấy `now` từ Clock · MỞ transaction · gọi domain · gọi repository ·
        │                                   Save → PullEvents → Outbox. Điều phối, KHÔNG chứa quy tắc.
        ▼
   3. DOMAIN      domain/<context>/          TOÀN BỘ quy tắc nghiệp vụ. Không biết HTTP, SQL, giờ.
        ▼
   4. REPOSITORY  adapter/postgres/          INSERT … ON CONFLICT DO UPDATE (qua Snapshot) · INSERT INTO outbox
        │                                   CÙNG transaction (tx nằm trong ctx) · COMMIT
        ▼
   5. WORKER      worker/ + cmd/worker       SELECT … WHERE sent_at IS NULL FOR UPDATE SKIP LOCKED
                                            → publish theo thứ tự → UPDATE sent_at (at-least-once)
                                            → context khác nghe và làm việc của nó
```

**Tầng use case, code thật**, `internal/app/catalog/register_merchant.go`. Chú ý **thứ tự**:

```go
func (h *RegisterMerchantHandler) Handle(ctx context.Context, d catalog.MerchantDetails) (catalog.MerchantID, error) {
	now := h.deps.Clock.Now() // đồng hồ đọc Ở ĐÂY, một lần, rồi truyền xuống

	var id catalog.MerchantID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		m, err := catalog.RegisterMerchant(d, now) // ① domain quyết định
		if err != nil {
			return err // lỗi nghiệp vụ, trả nguyên vẹn
		}
		if err := h.deps.Merchants.Save(ctx, m); err != nil { // ② LƯU TRƯỚC
			return fmt.Errorf("save merchant: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, m.PullEvents()); err != nil { // ③ RỒI MỚI pull
			return fmt.Errorf("outbox: %w", err)
		}
		id = m.ID()
		return nil
	})
	return id, err
}
```

Ba điều Go buộc phải viết khác Symfony, và cả ba đều học được khi viết thật:

- **`InTx` không generic.** Go không cho method generic trên interface, nên `fn` trả `error` và
  kết quả (`id`) được gán qua biến closure.
- **Command của `RegisterMerchant` chính là `catalog.MerchantDetails`.** Bọc thêm một struct
  trùng 100 % field chỉ là copy đổi tên. `AddProduct` thì khác: command có thêm `SourcedBy` và
  `Operator` (dữ liệu của *request*), use case ghép chúng với `now` thành `Provenance`. Đó là
  lý do tầng app tồn tại.
- **`Deps` + `mustHave`.** Mọi dependency vào một struct; thiếu cái nào thì panic lúc dựng
  handler, không đợi tới request đầu.

### Mỗi tầng làm gì, và tuyệt đối không làm gì

| Tầng | Làm | **Không** làm |
|---|---|---|
| `adapter/http/auth.go` | biết **ai** đang gọi, chặn ở cửa (401/403), dịch token → `Principal` | quyết định *cái gì* được làm với dữ liệu |
| `adapter/http` | giải mã, chuẩn hoá locale, mã lỗi HTTP; validate là gọi `Parse*` của domain | quy tắc nghiệp vụ; biết repository hay clock |
| `app` | mở transaction, lấy `now`, điều phối, luật **liên aggregate** (shop còn hoạt động? giá đúng tiền tệ của shop?) | quy tắc của một aggregate |
| `domain` | **toàn bộ** quy tắc nghiệp vụ | biết DB, HTTP, giờ |
| `adapter/postgres` | SQL, map snapshot ↔ cột, transaction | quyết định nghiệp vụ; phát event |
| `worker` | đọc outbox, gửi đi, đánh dấu; chạy sweep theo giờ | sửa dữ liệu nghiệp vụ; sweep chỉ **gọi** use case |

**Vì sao chuẩn hoá locale ở tầng 1 là thật, không phải hình thức.** `ParseMoney` từ chối
`"150,50"` (§11). Người Việt gõ dấu phẩy. Nếu adapter không đổi dấu phẩy thành dấu chấm theo
`Accept-Language`, khách không đặt được hàng. Nếu domain tự đoán, nó đoán sai với một trong
hai nhóm người dùng.

### So với Symfony

| Portage | Symfony |
|---|---|
| `adapter/http/auth.go` | Security firewall + voter |
| `adapter/http/` | Controller + `FormType` (chỗ `NumberFormatter` chạy) |
| `app/` | Service điều phối, `MessageHandler` |
| `domain/` | *(Symfony không có tầng này; logic nằm rải trong Service)* |
| `adapter/postgres/` | Doctrine Repository + mapping |
| outbox + worker | `postFlush` listener + Messenger |

## 31. Vòng đời một ĐƠN HÀNG — đi hết các context

Vòng đời này chạy hết bằng code: một test vàng đi từ `POST /merchants` tới `delivered` và
`variance` trên cả in-memory và Postgres, và `scripts/smoke.sh` chạy đúng flow đó trên binary
thật. Mỗi mũi tên dưới đây là một dòng trong `wire.Subscribe`.

```
  ┌─ CATALOG ──────────────────────────────────────────────────┐
  │  Merchant · CategoryPolicy · Product ⊃ Variant               │  khách chọn được món gì
  └──────────────────────────┬─────────────────────────────────┘
                             │ product_published (mang giá + hộp đã đo + link)
  ┌─ PRICING ────────────────▼─────────────────────────────────┐
  │  Quote = giá hàng + thuế + cước + phí dịch vụ, × tỷ giá      │  CHỤP ẢNH tỷ giá và bảng giá · hạn 48 giờ
  └──────────────────────────┬─────────────────────────────────┘
                             │ quote_accepted
  ┌─ ORDERING ───────────────▼─────────────────────────────────┐
  │  CustomerOrder · khách trả CỌC 50 %                          │
  └──────────────────────────┬─────────────────────────────────┘
                             │ deposit_paid
  ┌─ PROCUREMENT ────────────▼─────────────────────────────────┐
  │  PurchaseTask · hỏi MerchantACL · người đi mua thật          │
  └──────────────────────────┬─────────────────────────────────┘
              ┌──────────────┴──────────────┐
              ▼                             ▼
      purchase_confirmed              purchase_failed (hết size)
              │                             │
              │                    ┌─ BÙ TRỪ ────────────────┐
              │                    │ huỷ → hoàn 100 % cọc    │
              │                    └─────────────────────────┘
              ▼
  ╔═══════════════════════════════════════════════════════════╗
  ║  ĐIỂM KHÔNG THỂ QUAY ĐẦU (ordering: deposited → purchased) ║
  ║  Trước: khách huỷ → hoàn 100 % cọc                         ║
  ║  Sau:   khách huỷ → MẤT cọc (mình đang giữ hàng)           ║
  ╚═══════════════════════════════════════════════════════════╝
              │
  ┌─ LOGISTICS ──────────────▼─────────────────────────────────┐
  │  Parcel về kho Mỹ → CÂN THẬT · ConsolidationBatch gom lô     │
  │  Ship: chia hoá đơn cước cho từng đơn (FreightAllocator)     │
  └──────────────────────────┬─────────────────────────────────┘
                             │ batch_shipped (mang phần chia, đóng băng)
  ┌─ ORDERING ───────────────▼─────────────────────────────────┐
  │  in_transit · khách trả nốt 50 % · giao tận nhà · delivered  │
  └──────────────────────────┬─────────────────────────────────┘
                             ▼
  ┌─ ĐỐI SOÁT (pricing.Reconciliation) ────────────────────────┐
  │  Quote (ước tính) − Actual (thực tế) = variance −2.50 USD    │  GET /reconciliations/{order}
  └────────────────────────────────────────────────────────────┘
```

Ba chỗ vòng đời này dạy điều mà sơ đồ tầng (§30) không dạy:

1. **Vì sao `CustomerOrder` và `PurchaseTask` phải tách.** Nhìn nhánh `purchase_failed`. Tiền
   khách đã thu; việc mua hỏng nhiều giờ sau, ở một nơi khác. Hai thứ đó không thể chung một
   transaction, vì shop không tham gia transaction của mình (§25, §26).
2. **Vì sao `Quote` phải chụp ảnh.** Giữa lúc báo giá và lúc cân thật có nhiều tuần. Tỷ giá
   trôi, bảng giá đổi. Nếu `Quote` trỏ tới "tỷ giá hiện tại", giá của khách tự nhảy sau lưng họ
   (§11, §28).
3. **Vì sao cần đối soát ở cuối.** Cước ước tính lúc báo giá và cước thật sau khi cân là hai
   con số khác nhau. Chênh lệch đó chính là lời hay lỗ, và chỉ có nếu giữ **cả hai** (§28).

Từng bước, từng request, từng mã lỗi: [FLOW-ORDER.md](FLOW-ORDER.md).

---

# PHẦN VI — Khái niệm → mở file nào

Muốn thấy một khái niệm bằng mắt, mở đúng file này. Mỗi dòng một file, đọc được trong vài phút.

| Khái niệm | Mở | Nhìn gì |
|---|---|---|
| Value Object | `domain/shared/money.go` | struct field private, `Add` trả giá trị mới, `NewMoney` panic khi currency rỗng |
| Value Object mang luật kinh doanh | `domain/shared/weight.go` | `ChargeableWeight` ba dòng |
| Entity, typed ID | `domain/shared/id.go`, `domain/catalog/merchant.go` | `type MerchantID struct{ shared.ID }` |
| Invariant nhiều điều kiện | `domain/catalog/product.go` → `Publish` | năm `if`, năm sentinel |
| Aggregate có entity con | `domain/catalog/product.go`, `variant.go` | `AddVariant` là cửa duy nhất, `Variants()` trả copy |
| Domain Event, ghi không phát | `domain/shared/event.go` | `Record` / `PullEvents` |
| Repository là port | `domain/ordering/repository.go` | ba method, không có `findByStatusAnd…` |
| Hai adapter một port | `adapter/memory/ordering_repos.go`, `adapter/postgres/ordering_repos.go` | cùng interface, khác cách lưu |
| Memento cho persistence | `domain/catalog/snapshot.go` | `Snapshot()` / `ProductFromSnapshot()` |
| Domain Service là hàm | `domain/pricing/calc.go` | `Calculate(QuoteInputs) (Breakdown, error)` |
| Domain Service là interface | `domain/logistics/allocator.go` | `FreightAllocator`, `ByChargeableWeight` |
| Chia tiền đúng từng cent | `domain/shared/allocate.go` | floor + largest remainder |
| Aggregate là state machine | `domain/ordering/order.go` | 8 method có tên nghiệp vụ, `refuse()` chọn sentinel thật |
| Điểm không thể quay đầu | `domain/ordering/order.go` → `Cancel` | `switch` năm nhánh |
| Quote là ảnh chụp có hạn | `domain/pricing/quote.go` | `Accept` trễ vừa đổi trạng thái vừa từ chối |
| Quote vs Actual | `domain/pricing/reconciliation.go` | `Variance() = quoted − actual` |
| Provenance theo nhóm thuộc tính | `domain/catalog/provenance.go`, `product.go` | ba `Provenance` trên một `Product` |
| Projection của context khác | `domain/pricing/listing.go`, `domain/ordering/variant.go` | field exported, `Product shared.ID` không phải `catalog.ProductID` |
| Cùng event, hai mức chép | `domain/ordering/variant.go` (2 field) vs `domain/procurement/repository.go` `Variant` (5 field) | mỗi bên chép đúng cái luật mình cần |
| Port của tầng app | `app/ports.go` | `Clock`, `UnitOfWork`, `Outbox` |
| Use case: ghép rồi gọi domain | `app/catalog/add_product.go` | `now` + `SourcedBy` + `Operator` → `Provenance`; hai luật liên aggregate |
| Khuôn use case viết một lần | `app/ordering/deps.go` → `mutate` | load → hỏi aggregate → Save → Pull → Outbox |
| "Đã làm rồi" khác "không làm được" | `app/ordering/reactor.go` → `alreadyDone` | nuốt đúng một sentinel |
| Anti-Corruption Layer #1 | `domain/procurement/repository.go` → `MerchantACL`, `adapter/merchant/manual.go`, `router.go` | adapter đầu tiên là một con người |
| Anti-Corruption Layer #2 | `app/catalog/extract.go`, `adapter/openai/` | bốn string; `Unavailable` → 503 |
| Published Language | `contracts/*_v1.go` | chỉ primitive, version trong tên |
| Payload là hợp đồng viết tay | `adapter/eventcodec/codec.go`, `codec_test.go` | một case một event; guard `go/ast` |
| Bản đồ event | `platform/wire/subscribe.go` | 35 dòng, đọc từ trên xuống là §31 |
| Composition root | `platform/wire/wire.go` | `Memory()` / `Postgres()` trả cùng một `Graph` |
| Outbox cùng transaction | `adapter/postgres/outbox.go`, `postgres.go` → `db(ctx)` | tx nằm trong `ctx` |
| At-least-once relay | `worker/worker.go` → `RunOnce` | publish trước, `MarkSent` sau |
| Việc không ai gọi | `app/pricing/expire_quotes.go`, `worker/sweep.go` | mỗi quote một transaction |
| Read model không có domain | `app/reporting/summary.go`, `projector.go` | cái "khung", `if s.Status == ""` |
| Read model thứ hai | `app/reporting/worklist.go` | `NextStep()` quyết một lần |
| Auth là việc của biên | `platform/auth/auth.go`, `adapter/http/auth.go` | `Verifier` tách `Issuer`; bọc cả mux |
| 400 / 404 / 409 là ba câu trả lời | `adapter/http/errors.go` | một `errorTable`, `errors.Is` |
| Adapter làm đúng ba việc | `adapter/http/products.go`, `decode.go` | decode · locale · map lỗi; validate là `Parse*` của domain |
| Test canh kiến trúc | `domain/decisions_test.go` | 7 guard bằng `go/ast` |
| Test vàng cả hệ | `platform/wire/wire_test.go` → `wholeFlow` | một test, hai graph, tới `variance` |

---

# PHẦN VII — Bảng tra: Symfony ↔ Go ↔ DDD

| Symfony / Doctrine | Portage (Go) | Khái niệm DDD |
|---|---|---|
| `#[ORM\Entity] class X` | struct trong `internal/domain/` | Entity / Aggregate Root |
| `#[ORM\Embeddable]` | struct bất biến, không ID | Value Object |
| `XRepository extends ServiceEntityRepository` | interface ở domain + cài đặt ở adapter | Repository (đảo chiều phụ thuộc) |
| `EntityManager::flush()` | `repo.Save(ctx, aggregate)` | ranh giới transaction |
| `$em->wrapInTransaction(fn)` | `app.UnitOfWork.InTx(ctx, fn)` | Unit of Work |
| `services.yaml`, autowire | `wire.Graph` điền tay, `Deps` + `mustHave` | Dependency Injection thủ công |
| `EventSubscriber` | `shared.Events` + outbox | Domain Event |
| `postFlush` listener + Messenger `doctrine://` | outbox + `worker.Relay` | Outbox |
| `messenger.yaml routing:` | `wire.Subscribe` | Context Map dưới dạng code |
| Validator constraints | kiểm tra trong constructor và method | Invariant |
| `Controller` + `FormType` | `internal/adapter/http/` | Adapter |
| Service điều phối, `MessageHandler::__invoke` | `internal/app/*/XHandler.Handle` | Application Service |
| Service chứa luật | `internal/domain/*/` | Domain Service |
| `DECIMAL` trả về string, `Brick\Money` | `Money{minor int64}` | tránh float |
| `ClockInterface` / `MockClock` | `app.Clock` / `clock.Fixed`; domain nhận `now time.Time` | domain không đọc đồng hồ |
| Bundle | Bounded Context | Bounded Context |
| `symfony/uid` `Uuid::v7()` | `shared.NewID()` | Entity identity |
| `final class OrderId { Uuid $value }` | `type OrderID struct{ shared.ID }` | typed ID |
| `\DomainException` (bắt được) | `return error` + `errors.Is` | vi phạm nghiệp vụ |
| `\LogicException` (bug, không bắt) | `panic` | lỗi lập trình |
| `NumberFormatter` trong `FormType` | `normalizeAmount` theo `Accept-Language`, ở adapter | domain không đoán locale |
| `deptrac` | `go list -f '{{.Imports}}'` trong CI + guard 6 | Dependency Rule |
| `InMemoryRepository` trong `tests/` | `internal/adapter/memory`, adapter bình đẳng | Port / Adapter |
| Security firewall + voter | `authenticate()` bọc mux + `requireOperator/Customer/Any` | auth là việc của biên |
| `phpunit` | `go test ./...` | |
| `composer.lock` | `go.sum` | |

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
| Repository | kho chứa | nơi lấy và lưu aggregate |
| Domain Service | dịch vụ nghiệp vụ | logic không thuộc aggregate nào |
| Application Service | dịch vụ ứng dụng | lớp điều phối use case |
| Anti-Corruption Layer | lớp chống ăn mòn | lớp dịch với hệ thống ngoài |
| Ports & Adapters | cổng và bộ chuyển | domain ở giữa, ngoại vi cắm vào |
| Published Language | ngôn ngữ công bố | định dạng chung để hai context trao đổi |
| Projection | bảng chiếu | bản sao mỏng dựng từ event của context khác |
| Read Model | mô hình đọc | bảng bẹt chỉ để hiển thị |
| CQRS | tách đọc–ghi | hai model cho hai mục đích |
| Saga | chuỗi bù trừ | quy trình dài, hỏng thì bù bằng nghiệp vụ |
| Compensation | hành động bù trừ | việc làm để "gỡ" bước trước |
| Point of no return | điểm không thể quay đầu | mốc mà sau đó bù trừ đổi nghĩa |
| Eventual Consistency | nhất quán sau cùng | khớp nhau sau một lúc, không tức thì |
| Idempotent | lũy đẳng | làm hai lần như một lần |
| At-least-once | ít nhất một lần | không mất, có thể trùng |
| Outbox | hộp gửi đi | bảng event ghi cùng transaction |
| Memento / Snapshot | bản chụp trạng thái | cách lưu aggregate mà không mở field |
| Anemic Model | mô hình thiếu máu | class rỗng chỉ có getter/setter |
| Knowledge Crunching | moi kiến thức | ngồi với chuyên gia để đào quy tắc |

---

# PHẦN IX — Tự kiểm tra

Trả lời được không nhìn tài liệu thì đã hiểu.

**Cơ bản**

1. Value Object khác Entity ở điểm nào? Cho ví dụ trong Portage.
2. Vì sao `Money` không dùng `float64`?
3. Vì sao `Currency` phải mang theo số chữ số thập phân?
4. Invariant là gì? Kể hai invariant của Portage.
5. Anemic Domain Model là gì, vì sao nó tệ?

**Trung bình**

6. Vì sao `CustomerOrder` và `PurchaseTask` phải là hai aggregate ở hai context?
7. "Một transaction = một aggregate" nghĩa là gì? Kể chỗ repo cố tình nới luật đó, và lý do.
8. Vì sao domain không được import driver database? Điều đó đổi lại được gì cụ thể?
9. `Source` và `AsOf` của tỷ giá đã bị đưa ra khỏi `shared.ExchangeRate`. Câu hỏi nào giúp
   quyết định một field có thuộc shared kernel không?
10. Vì sao domain event luôn đặt tên ở thì quá khứ?

**Nâng cao**

11. Vì sao ranh giới giữa "hoàn 100 % cọc" và "mất cọc" là sự kiện `purchase_confirmed`,
    không phải một mốc thời gian?
12. Nếu bỏ Outbox thì hỏng chuyện gì? Kể một kịch bản cụ thể.
13. Vì sao mô hình ngôn ngữ phải nằm sau interface và trả string thô, không trả `Money`?
14. Nếu gộp Quote và Actual làm một cột `price` thì mất thông tin gì?
15. Muốn thêm brand Adidas, cần sửa những file nào? (Đáp án đúng: **không sửa file nào trong
    `internal/domain/`**.)

**Từ những bug đã gặp**

16. Vì sao `Grams(-1)` panic nhưng `NewWeight(-1)` trả `error`? Ai gọi cái nào?
17. `VolumetricWeight` cắt xuống 1900,8 g → 1900 g rồi `RoundUpTo(100 g)` vẫn ra 1900. Vì sao
    bước cuối làm tròn lên mà vẫn sai?
18. Vì sao `"26000.50"` và `"26000.5"` phải là **cùng một** `ExchangeRate` khi so `==`?
19. `shared.Money{}` là gì? Vì sao Go không cho cấm nó, và domain xử lý thế nào?
20. `github.com/google/uuid` nằm trong domain có vi phạm Dependency Rule không? Tiêu chí phân
    biệt "tiện ích" và "framework/SDK" là gì?
21. Tại sao `Events` là kiểu duy nhất trong `shared` có pointer receiver, và vì sao điều đó
    không phá nguyên tắc "shared toàn value object bất biến"?

**Từ tầng app**

22. Tầng app của catalog chỉ có vài lỗi riêng (`ErrMerchantInactive`, `ErrPriceCurrency`,
    `ErrSourcingNotAllowed`). Tiêu chí để một lỗi được ở tầng app thay vì domain là gì?
23. Vì sao `AddProduct` (command) không phải là `catalog.ProductDetails`, còn
    `RegisterMerchant` thì nhận thẳng `catalog.MerchantDetails`?

Thêm 17 câu ở [HOC.md](HOC.md), phần *Tự kiểm tra*.

---

# PHẦN X — Đọc thêm

| Sách / nguồn | Ghi chú |
|---|---|
| *Domain-Driven Design* — Eric Evans (2003) | sách gốc, dày và khó, nên đọc sau |
| *Implementing Domain-Driven Design* — Vaughn Vernon | thực dụng hơn, nhiều code |
| *Domain-Driven Design Distilled* — Vaughn Vernon | mỏng, đọc trước cuốn trên |
| *Learning Domain-Driven Design* — Vlad Khononov | dễ tiếp cận nhất cho người mới |
| martinfowler.com, bài *AnemicDomainModel* | ngắn, nên đọc ngay |
| *Domain-Driven Design with Golang* — Matthew Boyle | đúng ngôn ngữ đang dùng |

> **Lời khuyên:** đừng đọc hết rồi mới code. Code tới đâu, đọc lại phần đó tới đó. DDD là thứ
> chỉ hiểu được khi đã tự tay làm sai một lần.
