# Catalog — nguồn dữ liệu sản phẩm & thiết kế v1

> Ghi chép quyết định thiết kế cho bounded context `catalog`, viết trước khi
> code. Xem thêm [DDD.md](DDD.md) §22 (Anti-Corruption Layer), §23 (đa thương
> hiệu), §28 (Quote vs Actual).
>
> Ngày: 2026-09-04 · Trạng thái: **v1 đã code xong** — `Merchant`, `CategoryPolicy`,
> `ParcelSpec`, `FreeShipping`, `Product`/`Variant`/`Provenance`, `SourceURL`,
> 3 repository interface ✅ · 14 domain event · 42 test, coverage 91%

---

## 0. Câu hỏi gốc

> *"Mình có nuôi danh mục sản phẩm theo Nike không, hay chỉ là khách dán link?"*

**Trả lời: cả hai — và đó là chủ đích.**

Nuôi bản sao catalog Nike là sai lầm: họ có hàng trăm nghìn SKU, mình bán vài
trăm, và bản sao sẽ **luôn lỗi thời** — giá đổi, hết size, ngừng bán, mà mình
không biết.

Catalog của mình **lớn dần từ những món khách đã đặt**. Nó là thứ *kiếm được*,
không phải thứ *chép về*.

---

## ✅ QUYẾT ĐỊNH 04/09/2026 — KHÔNG dùng affiliate Nike

Sau khi đọc nguyên văn Reseller Policy (xem §2): **bỏ hướng affiliate, tự làm.**

**Lý do:** ký Publisher Agreement là tự tạo ra vi phạm hợp đồng ngay từ ngày
đầu — hợp đồng cấm đúng việc mình làm hằng ngày. Không đáng đánh đổi để lấy
hoa hồng ở một nhánh dịch vụ phụ.

**Được gì khi bỏ:**

| | |
|---|---|
| Không phải qua 9 bước onboarding CJ | không cần W-8BEN, không cần khai tài khoản ngân hàng cho CJ |
| Không bị chặn bởi yêu cầu "website có lưu lượng" | bắt tay làm được ngay |
| Không phải cam kết refresh feed 7 ngày/tuần | bớt một hệ thống nền phải nuôi |
| **Không có ràng buộc hợp đồng về bán lại** | giữ nguyên vị thế người tiêu dùng bình thường |

**Mất gì:** dữ liệu sản phẩm có cấu trúc sẵn, và hoa hồng.

**Hệ quả thiết kế — cái này quan trọng:** operator trở thành **nguồn dữ liệu
chính**, không còn là phương án dự phòng. Nghĩa là **tốc độ nhập liệu của
operator giờ là tính năng cốt lõi**, không phải việc phụ. AI chuyển vai từ
"đi cào trang web" sang **"trợ lý cho operator"**: đọc đoạn text khách dán,
gợi ý mã HS, quy đổi size, ước lượng kích thước hộp theo ngành hàng.

Vai trò đó vừa hợp pháp, vừa vẫn là phần AI đáng giá để đưa vào CV.

---

## 1. Nguồn dữ liệu — `SourcingMode`

| Nguồn | Dùng khi | Ưu tiên xây |
|---|---|---|
| **Operator nhập tay** (có AI hỗ trợ) | nguồn **chính** | 🥇 **v1** |
| **Khách dán thông tin** → AI cấu trúc hoá | khách tự cung cấp, không phải hệ thống đi lấy | 🥈 v1 |
| ~~Feed / API của merchant~~ | ❌ bỏ với Nike; để ngỏ cho merchant khác có điều khoản thân thiện | sau |
| Dataset Kaggle (H&M, Zappos) | chỉ để bơm quy mô lúc test tải | khi cần |

> ⚠️ Phân biệt kỹ: **hệ thống tự tải trang** (scraping) ≠ **người dùng cung cấp
> dữ liệu**. Ranh giới này giờ là ràng buộc thiết kế, không phải sở thích —
> xem §2.

```go
type Merchant struct {
    ID       MerchantID
    Name     string        // "Nike", "The North Face", "Sony"  ← DỮ LIỆU
    Sourcing SourcingMode  // Link | Operator | Feed | API
}
```

> **Vì sao "khách dán link" là đường chính, không phải feed:** ngày đầu catalog
> **rỗng**. Và nó chạy được với **mọi** merchant, không riêng Nike — đúng yêu
> cầu đa thương hiệu. Feed sau này chỉ là *thêm một adapter*, domain không đổi
> một dòng. Đó chính là lời hứa của Ports & Adapters, và mình sẽ thấy nó xảy ra
> thật.

---

## 2. Nike Affiliate — thông tin đã xác minh

Nguồn: `nike.com/help/a/affiliate-program`, đọc ngày **04/09/2026**.

| | |
|---|---|
| Mạng affiliate | **CJ Affiliate** (Commission Junction) |
| Product feed | ✅ có — *"automated product feed"* cho affiliate đã duyệt |
| Hoa hồng | tới **15%** doanh số Mỹ (tuỳ đối tác; một số hàng như gift card không tính) |
| Cookie | **7 ngày** |
| Điều kiện duyệt | xét **lưu lượng website**, chất lượng nội dung, độ phù hợp thương hiệu |
| Chi phí | miễn phí |

### 🔴 ĐÃ XÁC MINH 04/09/2026 — Nike CẤM mua để bán lại

Đọc trực tiếp **NIKE Publisher Agreement (March 2020)** trong tài khoản CJ
(`members.cj.com/member/publisher/onboarding.cj` → *Accept Program Terms*),
mục **10.7 Reseller Policy**, nguyên văn:

> *"You agree not to buy products or services from the Nike.com site for
> subsequent resale, **or to aid or abet any third party in doing so**. […]
> You will not engage in **spidering or scraping** our site for content,
> images, etc."*

**Hai câu này phủ định gần như toàn bộ ý tưởng dùng affiliate Nike cho Portage:**

| Việc mình định làm | Điều khoản nói gì |
|---|---|
| Mua ở Nike.com rồi bán lại cho khách VN (**mua hộ**) | ❌ *"not to buy … for subsequent resale"* |
| Hỗ trợ khách khác làm việc đó | ❌ *"or to aid or abet any third party"* |
| Hệ thống tự tải trang Nike để AI trích xuất | ❌ *"not engage in spidering or scraping"* |
| Dùng product feed | ✅ được, nhưng phải **refresh hằng ngày, cả 7 ngày/tuần** (10.9) |

Và mục 3.1: hoa hồng chỉ trả khi **khách truy cập nike.com qua Qualifying Link
rồi mua**; không tính đơn qua điện thoại; không tính phí ship, thuế, gift card.

> ⚠️ **Mình không phải luật sư.** Đây là đọc hiểu văn bản, không phải tư vấn
> pháp lý. Nhưng câu chữ khá thẳng, không mơ hồ.

#### Kết luận thực tế

**Không nên ký thoả thuận này khi đang chạy mô hình mua hộ.**

Lý do không phải "mất hoa hồng" — mà là **ký vào là tự tạo ra một vi phạm hợp
đồng ngay từ ngày đầu**. Hiện tại việc mua hàng ở Nike với tư cách người tiêu
dùng rồi gửi về VN chịu điều khoản bán lẻ thông thường. Ký Publisher Agreement
là **thêm một cam kết bằng văn bản** rằng mình không mua để bán lại — trong khi
đó chính là việc mình làm.

Hậu quả nếu bị phát hiện: chấm dứt tài khoản, **thu hồi hoa hồng đã trả**, và
tên mình nằm trong danh sách của Nike.

#### Ảnh hưởng lên thiết kế Catalog

Điều khoản chống scraping làm rõ một ranh giới **quan trọng cho kiến trúc**:

```
   ❌ Hệ thống tự tải trang nike.com → AI đọc      = scraping (tự động)
   ✅ Khách tự copy thông tin sản phẩm → dán vào   = khách cung cấp dữ liệu
   ✅ Operator tự xem trang → nhập tay             = người dùng bình thường
```

Nghĩa là luồng *"khách dán link → hệ thống fetch → AI trích xuất"* **phải đổi**
thành *"khách dán **thông tin**, hoặc operator nhập tay"* — ít nhất với Nike.

Điều này **không phá kiến trúc**: `SourcingMode` vẫn vậy, chỉ là adapter `link/`
đổi cách lấy dữ liệu đầu vào. Đúng tinh thần Ports & Adapters — đổi adapter,
domain không đổi một dòng.

#### Còn đường nào dùng được affiliate không?

Chỉ còn **ship hộ thuần tuý**, khi khách **tự mua cho nhu cầu cá nhân** (không
bán lại) và tự bấm qua link của mình. Đó là referral đúng nghĩa, không vi phạm
Reseller Policy.

Nhưng phải cân nhắc: đáng để ký cả bản hợp đồng chỉ vì một nhánh dịch vụ nhỏ,
trong khi nhánh chính (mua hộ) lại là thứ hợp đồng cấm?

**Đề xuất: tạm dừng ở bước Accept Terms. Chưa ký.** Quyết định lại sau khi biết
rõ mình chạy mô hình nào là chính.

### ⚠️ Mâu thuẫn phải xử lý trước khi tính tiền

**Affiliate và mua hộ đi ngược chiều nhau:**

```
   AFFILIATE                          MUA HỘ
   ─────────                          ──────
   Khách → nike.com → Nike bán        Khách → web mình → MÌNH mua ở Nike
   Mình ăn hoa hồng                   Mình ăn chênh lệch
   Khách trả tiền cho Nike            Khách trả tiền cho mình
```

Trong mua hộ, **khách không bao giờ vào nike.com**. Người bấm mua là mình. Tự
bấm link của chính mình để lấy hoa hồng gọi là **self-referral**, và hầu hết
chương trình affiliate **cấm điều này**.

> 🔴 **VIỆC CẦN LÀM:** đọc mục *self-referral / self-purchase* trong Nike
> Affiliate Program Terms & Conditions **trước khi** tính hoa hồng vào doanh
> thu. Tính nhầm rồi bị thu hồi thì lệch cả mô hình tài chính.
> *(Chưa xác minh — cần đọc bản T&C đầy đủ.)*

### ✅ Nhưng affiliate ăn được ở luồng khác

Nhớ lại hai dịch vụ đã tách từ đầu:

| Dịch vụ | Ai bấm mua | Affiliate |
|---|---|---|
| **Mua hộ** (buy-for-me) | mình | ❌ self-referral |
| **Ship hộ** (forwarding) | **khách** | ✅ hợp lệ, ăn hoa hồng bình thường |

Khách ship hộ **tự vào nike.com mua**. Nếu họ đi qua link affiliate của mình thì
đó là referral đúng nghĩa — ăn **cả phí ship hộ lẫn hoa hồng 15%**.

> **Đây là lý do kỹ thuật để giữ hai luồng dịch vụ tách bạch trong domain.**
> Không chỉ là mô hình hoá cho đẹp — chúng là **hai dòng doanh thu khác nhau**.
> Ghi vào backlog: luồng ship hộ phải gắn link affiliate khi hiện sản phẩm.

### Rào cản trước mắt

Nike duyệt affiliate dựa trên **lưu lượng website**. Hiện chưa có site nào →
**feed là năng lực có sau, không phải ngày một**. Kiến trúc phải chạy tốt khi
chưa có feed.

---

## 3. Điều quan trọng nhất: feed KHÔNG có thứ mình cần nhất

Feed affiliate sinh ra để **bán hàng**, không phải để **vận chuyển hàng**.

```
   FEED CHO MÌNH              CÁI MÌNH CẦN ĐỂ BÁO GIÁ
   ─────────────              ────────────────────────
   ✅ tên, ảnh, mô tả         ✅ tên, ảnh
   ✅ giá $150                ✅ giá $150
   ✅ màu, size               ✅ màu, size
   ✅ còn hàng / hết          ✅ còn hàng
   ✅ link mua, mã sản phẩm   ⚠️ KÍCH THƯỚC HỘP   ← QUYẾT ĐỊNH TIỀN CƯỚC
                              ⚠️ cân nặng đóng gói
                              ⚠️ mã HS → thuế
                              ⚠️ có pin lithium không
```

Nhớ bài học đắt nhất của project: **áo phao TNF cân 0,9 kg nhưng bị tính cước
4,7 kg**. Con số đó đến từ **kích thước hộp** — thứ không feed nào đưa.

> **Hệ quả:** cân nặng và kích thước thùng là dữ liệu **mình phải tự đo và tự
> tích luỹ**. Đó là **tài sản riêng**, thứ đối thủ không copy được, và là lý do
> chính đáng để giữ một catalog của riêng mình thay vì chỉ hiển thị lại feed.

---

## 4. `Provenance` — khái niệm mới, và vì sao cần

Ba đường vào tạo ra sản phẩm **không cùng mức đáng tin**:

```
   Khách dán link → AI      Operator nhập tay     Feed CJ
   ──────────────────       ─────────────────     ───────
   giá:        có thể sai   người xác nhận        chính xác
   cân nặng:   AI ĐOÁN ⚠️    người ước lượng       không có ⚠️
   kích thước: KHÔNG CÓ ⚠️   người ước lượng       không có ⚠️
   mã HS:      AI đoán      người tra             không có
```

**Nên `Provenance` phải theo TỪNG NHÓM THUỘC TÍNH, không phải một nhãn cho cả
sản phẩm.** Feed cho giá đáng tin nhưng kích thước thì không có — một nhãn
"nguồn = Feed" gắn cho cả sản phẩm sẽ nói dối.

```go
// provenance.go — bản đã code
type Provenance struct {
    source SourcingMode        // operator | customer | feed
    at     time.Time
    by     shared.OperatorID   // ai xác nhận — bắt buộc khi source = operator
}
func (p Provenance) Verified() bool   // = source == operator: có NGƯỜI BÊN MÌNH đứng sau

// product.go — ba nhóm, ba provenance
type Product struct {
    listingProv Provenance   // tên + merchant + category + URL  → món này LÀ gì
    priceProv   Provenance   // giá                              → CÓ giá bao nhiêu
    parcelProv  Provenance   // cân nặng + hộp                   → BILL bao nhiêu
}
```

Ba nhóm, không phải hai như bản nháp: **category sai là restriction sai** (tai
nghe có pin bị xếp vào "apparel" → đi tuyến không cho pin → kẹt kho), nên "món
này là gì" cũng cần người xác nhận, không chỉ cân nặng. Nhóm *thuế* đã bỏ —
thuế thuộc `pricing.ShippingLane` (§6).

`Verified()` chỉ đúng với `operator`. Feed nói đúng điều **shop nói** (giá, tên)
nhưng không biết điều **mình cần** (hộp, cân nặng) — nên với mục đích của mình,
feed cũng là chưa xác thực.

> **Đã cân nhắc và KHÔNG chọn (04/09, review chéo):** thu `listing` thành
> `classification` — chỉ phủ category, với lý lẽ tên sai/merchant sai *tự lộ*
> (khách kêu, operator thấy khi mua), chỉ category sai mới âm thầm. Lý lẽ đúng
> về kỹ thuật. **Giữ `listing`** vì: (1) lúc xác nhận "món này là gì", operator
> nhìn **cả** tên, shop, link, ngành — tách riêng category là tách một hành động
> thật thành hai dấu; (2) "tự lộ" = lộ **sau khi** đã báo giá hoặc đã mua —
> muộn hơn một bước so với xác nhận trước khi publish. Nếu sau này thấy operator
> chỉ thật sự soi category, đổi tên là việc 15 phút, có test bao.

### Invariant rút ra được — đã thành code: `Product.Publish`

> **Không thể phát hành `Quote` chắc chắn từ sản phẩm chưa có kích thước xác
> thực.**

Trong code, "chắc chắn" = `Product` ở trạng thái **published**, và `Publish()`
từ chối cho tới khi đủ **5 điều kiện**, mỗi điều kiện một sentinel error để màn
hình operator nói đúng thứ còn thiếu:

| # | Điều kiện | Sai thì |
|---|---|---|
| 1 | đang là draft | `ErrNotDraft` |
| 2 | có ≥ 1 variant — có thứ để đặt | `ErrNoVariants` |
| 3 | `listingProv.Verified()` — người xác nhận món này là gì (`ConfirmListing`) | `ErrUnverified` |
| 4 | đã `Measure` **và** `parcelProv.Verified()` — người đã cân | `ErrUnverified` |
| 5 | không còn cờ nghi trùng (§7) | `ErrSuspectedDuplicate` |

Sản phẩm **draft** vẫn quote được — nhưng Pricing biết mình đang đoán: lấy
`CategoryPolicy.DefaultParcelSpec()` thay cho `Product.ParcelSpec()` chưa có, và tự chọn một
trong ba cách bên dưới. Đó là việc của Pricing, không phải của Catalog:

- báo giá kèm **biên an toàn rộng hơn**, hoặc
- đánh dấu *"giá tạm tính, chốt lại khi cân thật"*, hoặc
- đẩy sang operator xác nhận trước khi báo khách

Đây chính là `Quote vs Actual` ([DDD.md §28](DDD.md)) nhưng ở **tầng dữ liệu đầu
vào** — cùng một nguyên tắc, gặp lại ở chỗ khác.

### `Variant` — `size` là **chữ của shop**, không phải enum của mình (08/09)

Khách mua một hình thức cụ thể, và cách gọi hình thức đó là của **shop**, không
phải của Portage. `size` và `color` là `string` tự do: `US 9`, `M 8 / W 9.5`,
`42 EU`, `XS`, `Sail/Gum` đều vào được. Không có bảng size chuẩn, vì mọi bảng
size chuẩn đều sai với một shop nào đó, và người đi mua thì đọc chữ của shop.

Nhưng chữ tự do vẫn cần **khoá chuẩn hoá** để không có hai variant cùng nghĩa:

| Hai cách viết | Cùng một variant? | Vì sao |
|---|---|---|
| `US 9` / `us 9` | **có** | hoa/thường không phải thông tin |
| `M8/W9.5` / `M 8 / W 9.5` | **có** | khoảng trắng không phải thông tin — `strings.Fields` bỏ hết |
| `8 M / 9.5 W` / `M 8 / W 9.5` | **không** | đảo thứ tự là chữ khác; đoán rằng hai cách đó cùng nghĩa là đổi nghĩa dữ liệu của shop |

Và một luật nữa, đến từ việc đọc lại comment cũ trước khi đổi: `VariantDetails`
cho phép **rỗng hết** vì một sản phẩm một size có đúng một variant không có gì để
nói. Nên luật không phải "chặn variant rỗng", mà là:

> variant không size không màu chỉ được là variant **duy nhất** của sản phẩm
> (`ErrUnnamedVariant`).

Sản phẩm một size vẫn chạy; còn "một cái `US 9` cộng một cái không tên" thì không
— lúc đi mua không ai biết cái không tên là cái gì.

`AddVariant` phát `catalog.variant_added` mang **cả ba chữ** (size, color,
merchant_ref). Entity con không có repository, nhưng vẫn kể được — và phải kể,
vì ordering và procurement cần size mà không được đọc bảng của catalog
(WALKTHROUGH §22).

### Khách xin size nào — `RequestedVariant` (08/09)

`Variant` là **sự thật về sản phẩm**: shop bán những hình thức này. Còn khách muốn hình thức
nào là một **yêu cầu**, và hai thứ đó không cùng loại.

Nên bản nháp mang thêm hai trường, cả hai nói về yêu cầu chứ không nói về sản phẩm:

| Trường | Trả lời | Rỗng khi nào |
|---|---|---|
| `RequestedBy` | khách nào đang đợi món này | operator tự thêm hàng, chưa ai xin |
| `RequestedVariant` | họ xin size hoặc màu nào, **bằng chữ của họ** | sản phẩm một loại, hoặc khách không ghi |

Cả hai **không** nằm trong `Provenance`. Provenance trả lời *"dữ liệu này từ đâu tới và ai
bảo đảm"* — nó là chuyện độ tin. Hai trường này trả lời *"ai đang đợi và họ xin gì"* — chuyện
đơn hàng. Gộp lại là làm mờ cả hai.

Và `RequestedVariant` **không bao giờ** được coi là tên của một `Variant` có thật. Nó là chữ
ai đó gõ trước khi có người kiểm shop bán gì. Operator vẫn phải mở trang, rồi tạo `Variant`
thật — với ô đã điền sẵn chữ đó, và yêu cầu gốc vẫn hiện bên cạnh sau đó, vì *"khách xin
US 9, mình tạo US 9.5"* là chuyện phải nói với khách.

Đây cũng là lý do thứ hai để size là chữ tự do. Giày người lớn có hệ `M` và `W` in cùng một
trang; giày trẻ dùng `Y` và `C`; áo dùng `S/M/L`; điện thoại dùng dung lượng. Một enum phủ
hết thì không tồn tại, và enum nào cố phủ sẽ từ chối một size thật vào ngày shop đặt tên
mới. Cùng số 1 mà `1Y` và `1C` là hai đôi khác nhau — nên UI **dạy** hệ size thay vì ép
danh sách chọn.

---

## 5. Best practice khi tích hợp feed (áp dụng cho CJ và mọi feed sau này)

### (1) Feed là adapter, không phải domain

```
internal/domain/catalog/           ← KHÔNG có chữ "Nike", "CJ" nào
   type ProductSource interface {  ← domain chỉ nói CẦN GÌ
       Fetch(ctx, ref MerchantRef) (ProductData, error)
   }

internal/adapter/merchant/
   ├── cj/         ACL cho CJ Affiliate
   ├── link/       khách dán link + AI trích xuất
   └── manual/     operator nhập tay
```

Model của Nike/CJ (`styleColor`, `pid`, `gtin`…) **không được rò vào domain**.
Đó đúng là Anti-Corruption Layer ([DDD.md §22](DDD.md)).

### (2) Lưu nguyên bản feed, tách khỏi bản đã dịch

```
merchant_feed_raw     ← y nguyên JSON/CSV họ gửi, KHÔNG đụng vào
products              ← model của MÌNH, dịch ra từ bảng trên
```

Feed đổi định dạng → dịch lại từ bản gốc mà không mất dữ liệu. Bỏ qua bước này
là sau phải đi tải lại lịch sử, và thường là không tải lại được.

### (3) Feed cập nhật giá, KHÔNG cập nhật báo giá đã phát

Nike giảm giá lúc 3 giờ sáng → `Quote` của khách hôm qua **giữ nguyên giá hôm
qua**. Lại là nguyên tắc chụp ảnh của Quote ([DDD.md §28](DDD.md)).

### (4) Feed là nguồn *bổ sung*, không phải nguồn *chính*

Kiến trúc phải chạy được khi:

- feed chết / CJ đổi API
- Nike ngừng chương trình affiliate
- thêm merchant **không có** feed nào cả

Đường "khách dán link" là **xương sống**. Feed chỉ nâng chất lượng dữ liệu cho
những món bán nhiều.

---

## 6. Phạm vi Catalog v1

```
internal/domain/catalog/
├── merchant.go       ✅ Merchant (aggregate root) — Rename, Enable/DisableSourcing,
│                        ChangeFreeShipping, Suspend/Reinstate; MerchantID, MerchantStatus,
│                        SourcingMode, Hostname, MerchantDetails
├── freeshipping.go   ✅ FreeShipping — value object 3 trạng thái
├── category.go       ✅ CategoryPolicy — VO khoá tự nhiên: ParcelSpec mặc định + Restrictions
│                        CategoryCode (struct, validate ^[a-z][a-z0-9_]*$), Restriction
├── parcelspec.go     ✅ ParcelSpec — cân nặng + hộp; dùng chung cho ước lượng (category)
│                        và số đo thật (product) — độ tin nằm ở Provenance
├── provenance.go     ✅ Provenance — nguồn + thời điểm + ai xác nhận, theo NHÓM thuộc tính
├── sourceurl.go      ✅ SourceURL — link tham chiếu (không bao giờ fetch), tách Host()
├── product.go        ✅ Product (aggregate root) — AddVariant, ConfirmListing, Measure,
│                        Reprice, FlagDuplicateOf/ClearDuplicateFlag, Publish, Retire
├── variant.go        ✅ Variant — entity con; VariantID, VariantDetails; khoá chống
│                        trùng bỏ MỌI khoảng trắng; unnamed() cho luật "không tên = duy nhất"
├── events.go         ✅ 16 event: 7 Merchant* + Product Added/Measured/Repriced/
│                        FlaggedDuplicate/DuplicateCleared/Published/Retired + CategoryDefined
│                        + VariantAdded (08/09 — size ra khỏi catalog lần đầu)
└── repository.go     ✅ MerchantRepository, CategoryRepository, ProductRepository — PORT
```

`shared/operator.go` ✅ `OperatorID` — "ai xác nhận" là danh tính mọi context đều
hỏi (đo, mua, đóng gói), nên ở shared kernel; tên/role/login thuộc context
identity sau này.

### Tầng app của catalog — `internal/app/catalog/` (package `catalogapp`) ✅

| Use case | Làm gì | Luật riêng của tầng app |
|---|---|---|
| `RegisterMerchantHandler` | `MerchantDetails` → `Merchant` → Save → outbox | — |
| `AddProductHandler` | command (+ `SourcedBy`, `Operator`, `RequestedBy`, `RequestedVariant`) → `Provenance` → `Product` nháp → phát hiện trùng → Save → outbox | `ErrMerchantInactive` (shop bị treo thì không thêm hàng), `ErrSourcingNotAllowed` (shop không nhận nguồn này), `ErrPriceCurrency` (giá phải cùng tiền tệ với shop) |
| `PublishProductHandler` | load → `Publish(now)` → Save → outbox | — (mọi "không" là của `Product.Publish`) |
| `DefineCategoryHandler` (05/09) | `CategoryPolicy` → Save → **thông báo** `CategoryDefined` | — ; seed của `wire` cũng đi đường này nên mỗi lần khởi động thông báo lại (consumer idempotent) |
| `AddVariantHandler` (05/09) | load → `AddVariant` → Save → **outbox** (`variant_added`, 08/09) | — (`ErrDuplicateVariant`, `ErrUnnamedVariant` đều là của aggregate) |
| `ConfirmListingHandler` (05/09) | operator + now → `Provenance` verified → `ConfirmListing` → Save | — (`ErrOperatorRequired` là của `NewProvenance`) |
| `MeasureProductHandler` (05/09) | operator + now → `Provenance` → `Measure(spec)` → Save → `product_measured` | — |

Chạy với adapter in-memory hoặc **Postgres** (`internal/adapter/postgres/` — repo đi
qua `Snapshot()`/`FromSnapshot()` của aggregate, outbox cùng transaction), phơi ra HTTP
bởi `internal/adapter/http/`, event ra ngoài bằng `internal/worker` (at-least-once).
Payload event là hợp đồng viết tay ở `internal/adapter/eventcodec/`. Mổ xẻ từng dòng từ
`curl` tới `sent_at`: [WALKTHROUGH.md](WALKTHROUGH.md).

### ⚠️ Đã gỡ khỏi `CategoryPolicy`: mã HS và thuế suất nhập khẩu

Bản đầu có `hsCode` và `dutyRate`. **Gỡ ngày 04/09** sau khi nhận ra hai lỗi:

**Lỗi 1 — mâu thuẫn với tuyến vận chuyển đã chốt.** Tuyến đang dùng là *dịch vụ
gửi hàng ở Mỹ, trọn gói tận nhà VN*: thông quan và thuế **đã nằm trong giá /kg**,
mình không khai hải quan, không tính thuế. Xây sẵn phần tính thuế là xây cho một
tuyến đã quyết định không dùng.

**Lỗi 2 — và cái này nặng hơn: nó nằm SAI bounded context.**

```
   Thuế = f( ngành hàng , TUYẾN VẬN CHUYỂN )
                          ↑
   Tuyến trọn gói        → không có dòng thuế riêng, đã gộp vào giá /kg
   Tuyến nhập thương mại → tính riêng theo mã HS
```

Cùng đôi giày, đi hai tuyến khác nhau thì dòng thuế **khác hẳn**. Nghĩa là thuế
là thuộc tính của **cách vận chuyển**, không phải của **món hàng**.

`CategoryPolicy` trả lời *"món này **là** cái gì"*. Nó không được biết mình gửi
bằng đường nào. Khi nào cần tính thuế thật, nó thuộc về `pricing`, gắn vào
`ShippingLane`.

> **Bài học DDD:** để `dutyRate` trong Catalog sẽ tạo ra một con số **trông có
> thẩm quyền nhưng sai** — category khai 30% trong khi tuyến đang dùng lờ nó đi.
> Field đặt sai context nguy hiểm hơn field thừa: field thừa thì không ai gọi,
> field sai chỗ thì có người gọi và tin.

### ⚠️ Đã gỡ tiếp (04/09, chiều): số chia thể tích — cùng một lỗi, lệch một field

Sau khi gỡ thuế, `volumetricDiv` vẫn còn trong `CategoryPolicy`, với comment
tự nhận *"the real divisor is carrier data and travels with the shipping lane"*
— rồi vẫn lưu theo ngành hàng, và có test đóng đinh `luggage` dùng 6000.

Áp đúng câu hỏi vừa dùng cho thuế:

```
   Số chia thể tích = f( TUYẾN VẬN CHUYỂN )       ← không phụ thuộc món hàng
   Cùng đôi giày: FedEx 5000, tuyến khác 6000
   Cùng tuyến:    giày hay áo phao đều 5000
```

Để ở category là tạo **hai nguồn sự thật** — category nói 6000, lane nói 5000,
ai thắng? Gỡ `volumetricDiv` và `CategoryPolicy.ChargeableWeight`; `ShippingLane`
(pricing) giữ số chia **và** bước làm tròn, ngay cạnh thuế, cùng một lý do.

> **Bài học:** khi gỡ một field vì "nó thuộc context khác", hỏi tiếp *"còn field
> nào cùng nguồn gốc không?"*. Thuế và số chia đều là dữ liệu của hãng vận
> chuyển; gỡ một mà để một là sửa nửa lỗi.

### ✅ Đã thêm (04/09, chiều): `ParcelSpec` (tên đầu: `Estimate`) — thứ §3 gọi là "tài sản riêng"

> **Cập nhật 05/09:** `ParcelSpec` **dời sang `internal/domain/shared/parcelspec.go`**. Lý do:
> `pricing` cần đúng kiểu này (hộp ước lượng trong `CategoryProfile`, hộp đo thật trong
> `Listing`) mà guard 7 cấm `pricing` import `catalog`. Một kiểu hai context dùng chung, cùng
> nghĩa → shared kernel. Mọi thứ dưới đây vẫn đúng, chỉ đổi tiền tố `catalog.` → `shared.`.

§3 nói *"cân nặng và kích thước thùng là dữ liệu mình phải tự tích luỹ"*, §4 nói
quote trên sản phẩm chưa đo là *"quote dựa trên phỏng đoán"* — nhưng
`CategoryPolicy` không có gì để Pricing quote một món **lần đầu** (chưa có
`Product` đo thật). Thêm:

```go
type ParcelSpec struct {        // parcelspec.go
    weight shared.Weight        // giày ~1.2 kg
    dims   shared.Dimensions    // hộp 33×22×13 cm
}
NewParcelSpec(w, d) (ParcelSpec, error)   // cả hai bắt buộc — thiếu một nửa là quote sai một cách tự tin
```

Đặt tên lúc đầu là `Estimate`; đổi thành `ParcelSpec` khi viết `Product`, vì cùng
shape đó `Product` dùng cho **số đo thật**. Giá trị cùng kiểu, độ tin khác nhau
— và độ tin nằm ở `Provenance`, không nằm ở tên kiểu (§4).

Vai trò: **mặc định theo ngành hàng**. Khi `Product` đã lên cân thật (Provenance
§4), số đo thật ghi đè. Pricing ghép với lane:

```go
shared.ChargeableWeight(est.Weight(), est.Dimensions(), lane.Divisor(), lane.Step())
```

### `CategoryCode`: từ `type string` sang `struct{ code string }`

`type CategoryCode string` để ai cũng ép được `CategoryCode("Giày Dép!")` — compile
bình thường, bỏ qua mọi kiểm tra. Nó là khoá tự nhiên đi vào URL, config, cột
primary key, nên phải validate như `Hostname`: chỉ tạo được qua
`ParseCategoryCode` (`^[a-z][a-z0-9_]*$`, lower + trim). Struct field private thì
compiler chặn đường tắt.

### `Merchant`: thêm cách "tắt"

Bản đầu có `Rename`, `EnableSourcing` nhưng không có cách dừng mua ở một shop.
Thêm `Suspend(reason, now)` / `Reinstate(now)` + `MerchantStatus`, `DisableSourcing`,
`ChangeFreeShipping` (hãng đổi chính sách ship liên tục). `MerchantSuspended` là
event **liên context** đáng giá nhất của catalog: Procurement nghe nó để dừng tạo
việc mua và đưa người vào xem các việc đang mở.

Còn lại trong `CategoryPolicy`, đều là thuộc tính của **món hàng**:

| Field | Dùng để |
|---|---|
| `code` | danh tính; tra bảng giá của nhà gửi hàng (hàng thường / hiệu / điện tử / nhạy cảm) |
| `estimate` | cân nặng + hộp ước lượng → Pricing quote khi chưa có số đo thật |
| `restrictions` | pin lithium, chất lỏng… — nhà gửi hàng từ chối hoặc phụ thu |

### Học được gì ở v1

| Khái niệm DDD | Học qua cái gì | |
|---|---|---|
| **Entity** (khác Value Object) | `Merchant` có ID, có vòng đời (`Suspend`/`Reinstate`) | ✅ |
| **Entity khoá tự nhiên vs VO cấu hình** | `CategoryPolicy` — khoá là `CategoryCode`, bất biến, thay cả cục | ✅ |
| **Aggregate Root** | `Merchant` là cửa duy nhất — field private, đổi qua method có tên | ✅ |
| **Domain Event** đầu tiên | `MerchantRegistered`… `MerchantSuspended` — 7 event | ✅ |
| **Invariant nhiều field** | `NewParcelSpec`: cân nặng **và** đủ 3 cạnh, thiếu một là lỗi | ✅ |
| **Repository là PORT** | `MerchantRepository`, `CategoryRepository` — interface ở domain | ✅ |
| **Đa thương hiệu = dữ liệu** | không tên hãng nào trong code — có test canh | ✅ |
| **Aggregate có entity con** | `Product` là root; `Variant` chỉ tạo qua `AddVariant`, `Variants()` trả copy, không có `VariantRepository` | ✅ |
| **Provenance theo nhóm thuộc tính** | 3 `Provenance` trên `Product`; `Publish` đòi listing + parcel `Verified()` | ✅ |
| **Invariant 5 điều kiện, mỗi cái một lỗi** | `Publish()` — `ErrNotDraft`, `ErrNoVariants`, `ErrUnverified`, `ErrSuspectedDuplicate` | ✅ |
| **Trạng thái có vòng đời** | `draft → published → retired`; retired vẫn giữ cho đơn đã tham chiếu | ✅ |
| **Quyết định treo → code** | cờ nghi trùng (§7, phương án C) là field + 2 method, gộp là workflow tầng app | ✅ |
| **Event mang đủ trạng thái** (05/09) | `ProductPublished` mang `Price` + `Parcel` + `Source`; `CategoryDefined` mới — vì `pricing` không load được aggregate của catalog (guard 7), event là *tất cả* nó có | ✅ |
| **Entity con cũng có thể cần event** (08/09) | `VariantAdded` — hai context ngoài cần biết size mà không được đọc bảng của catalog: ordering để từ chối id lạ, procurement để in "US 9 · black" cho người đi mua. Entity con **không** có repository, nhưng vẫn **kể** được | ✅ |
| **Một trường không ai kiểm thì không phải luật** (08/09) | `Merchant.Supports()` có từ đầu và **không ai gọi**: `sourcing` được ghi, được thông báo bằng `merchant_sourcing_enabled/disabled`, rồi bỏ quên. Giờ `AddProduct` kiểm nó, và có test. Tìm ra khi đang viết docs cho UI: định mô tả một luật mà nó không tồn tại | ✅ |
| **Chuẩn hoá chữ tự do vừa đủ** (08/09) | khoá chống trùng bỏ mọi khoảng trắng (`M8/W9.5` == `M 8 / W 9.5`) nhưng **không** sắp lại thứ tự (`8 M / 9.5 W` vẫn khác) — đoán hộ shop là đổi nghĩa dữ liệu của họ | ✅ |
| **Catalog không biết ai nghe** (05/09) | catalog chỉ `Record`; `pricing` dựng `Listing`/`CategoryProfile` từ event qua `internal/contracts` — catalog không có dòng code nào về pricing. Xem WALKTHROUGH.md §14b | ✅ |
| **Event lớn lên theo consumer** (05/09, chiều muộn) | `procurement` cần biết shop tính tiền gì → `MerchantRegistered` mang thêm `Currency`; đổi struct + codec + `MerchantRegisteredV1` + bảng test cùng lúc, guard `go/ast` bắt nếu quên. Catalog vẫn không biết procurement là ai | ✅ |

---

## 7. Quyết định còn treo

### 🔴 Khách dán link trùng sản phẩm catalog đã có — xử lý sao?

| | Cách | Đánh giá |
|---|---|---|
| **A** | Nhận diện và **dùng lại** sản phẩm đã có | Tốt nhất về dữ liệu (biết ngay cân nặng, kích thước, thuế → báo giá chắc chắn). Nhưng phải giải bài toán *"link này có phải món kia không"* |
| **B** | Luôn **tạo mới** | Đơn giản, nhưng catalog đầy bản trùng, mất lợi thế dữ liệu đã tích luỹ |
| **C** | Tạo mới + **đánh dấu nghi trùng**, operator gộp tay | ⭐ **Đề xuất cho v1** — đúng nghiệp vụ (operator vẫn phải xem đơn trước khi mua), không cần AI so khớp phức tạp ngay, vẫn để đường mở lên A sau |

**✅ Chốt 04/09: phương án C.** Trong code:

- `ProductRepository.BySource(url)` — câu hỏi đầu tiên: *đã có sản phẩm nào cho
  trang này chưa?* Tầng app hỏi, rồi quyết định có cắm cờ hay không. **Đã chạy
  thật** trong `AddProductHandler` bước ⑧ — test
  `TestAddProduct_flagsSuspectedDuplicate`.
- `Product.FlagDuplicateOf(other, reason)` — cắm cờ, phát `ProductFlaggedDuplicate`
  (kèm lý do detector thấy: "same page url") vào hàng đợi operator. Không tự trỏ
  chính mình, không trỏ id rỗng.
- `Product.Publish()` **từ chối** khi còn cờ (`ErrSuspectedDuplicate`).
- `Product.ClearDuplicateFlag(reason)` — operator xem xong, nói "đây là món khác"
  **và vì sao**. Phát `ProductDuplicateCleared` — bản ghi quyết định.
- **Cặp đã bác không bị gắn cờ lại.** Aggregate nhớ `dismissedDuplicates`;
  `FlagDuplicateOf(cùng other)` sau đó là no-op. Không có bước này thì đợt rà
  trùng sau lại gắn cờ, operator lại xoá — vòng lặp vô tận, và không ai biết
  lần trước đã xét. Event ghi lịch sử; nhưng thứ **chặn vòng lặp** phải nằm
  trong aggregate, không trông vào ai đó xây read model từ event.
- **Gộp** (chuyển variant, retire bản trùng) là **workflow tầng app**, không phải
  method của aggregate — vì nó đụng hai aggregate, mà một transaction chỉ một
  aggregate (DDD.md §14 luật 4).

### Việc cần làm ngoài code

- [ ] Đọc **self-referral clause** trong Nike Affiliate T&C — quyết định hoa hồng có tính vào doanh thu được không
- [ ] Xin **bảng giá cân ký** thật từ nhà gửi hàng bên Mỹ (xem [SETUP.md](SETUP.md) §7)
- [ ] Xác nhận số chia thể tích nhà vận chuyển dùng: **5000** hay 6000
- [ ] Ghi backlog: luồng **ship hộ** phải gắn link affiliate khi hiện sản phẩm

---

## 8. Nguồn & mức độ tin cậy của tài liệu này

| Nội dung | Trạng thái |
|---|---|
| Nike dùng CJ Affiliate, có product feed, 15%, cookie 7 ngày | ✅ đọc từ trang chính thức của Nike, 04/09/2026 |
| **Nike cấm mua để bán lại + cấm scraping** | ✅ **đọc nguyên văn NIKE Publisher Agreement (March 2020) trong tài khoản CJ, 04/09/2026 — mục 10.7 Reseller Policy** |
| Hoa hồng chỉ tính khi khách qua Qualifying Link mua trên nike.com | ✅ mục 3.1 cùng văn bản |
| Feed phải refresh hằng ngày nếu dùng | ✅ mục 10.9 cùng văn bản |
| Feed affiliate không chứa kích thước hộp | ⚠️ **suy luận**, chưa xem feed thật |
| Ship hộ (khách tự mua, dùng cá nhân) ăn được hoa hồng | ⚠️ đúng về logic đọc hiểu, **chưa có xác nhận từ Nike/CJ** |

> Toàn văn thoả thuận xem trực tiếp trong tài khoản CJ. Không lưu bản sao vào
> repo này (văn bản có bản quyền, và repo có thể công khai).
