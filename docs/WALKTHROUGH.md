# Walkthrough — đi hết một request, đọc code theo thứ tự

> Mục đích: học **toàn bộ flow** từ test → use case → domain → adapter bằng đúng
> code trong repo, không phải sơ đồ chung chung. Mở file bên cạnh, đọc theo số
> thứ tự. Mỗi bước ghi: *đọc file nào*, *đọc gì*, *học được gì*, *bên Symfony là gì*.
>
> Request được chọn: **khách dán link một đôi giày → hệ thống ghi nhận sản phẩm
> nháp** (`AddProduct`). Đây là use case giàu nhất: đụng 3 repository, 2 luật
> liên aggregate, phát hiện trùng, và 2 domain event.
>
> Ngày: 2026-09-04 (§0–§13) và 2026-09-05 (§14) · Trạng thái code: **cả 5 tầng chạy thật,
> HAI bounded context** — HTTP → use case → domain → Postgres (cùng transaction với outbox)
> → worker publish → `pricing` nghe event của `catalog`, dựng projection, phát báo giá.
> `go run ./cmd/api -dsn …` + `go run ./cmd/worker -dsn …`; không có `-dsn` thì api chạy
> in-memory (kèm relay trong process). Đọc §0–§13 trước (một context, hết 5 tầng), rồi §14
> (context thứ hai nói chuyện với context thứ nhất bằng gì), rồi §15 (`ordering`: luật cọc
> 50 % và điểm không thể quay đầu thành code), rồi §16 (`procurement`: ACL đầu tiên, saga khép kín),
> rồi §17 (`logistics`: cân thật, chia cước, Quote vs Actual — vòng đời §31 chạy hết).
> §18 là cửa auth, §19 việc không ai gọi, §20 read model, §21 ACL cho AI, và **§22 là
> đợt gần nhất**: một chữ của shop (`size`) đi qua ba context — đọc nó nếu muốn thấy
> cách thêm một event vào hệ thống đang chạy mà không phá số vàng.

---

## 0. Bản đồ — một request đi qua cửa auth rồi 5 tầng

```
   CỬA (auth)            internal/adapter/http/auth.go      ✅  Bearer token → auth.Principal; không token → 401, sai loại → 403   → §18
        │
        ▼
   HTTP adapter          internal/adapter/http/            ✅  giải mã JSON, chuẩn hoá "150,50" → "150.50", map lỗi → status
        │
        ▼
   USE CASE              internal/app/catalog/             ✅  AddProductHandler.Handle
        │   đọc Clock · mở UnitOfWork · gọi repository · để DOMAIN quyết · Save → PullEvents → Outbox
        ▼
   DOMAIN                internal/domain/catalog/          ✅  catalog.AddProduct, Product.FlagDuplicateOf
        │   toàn bộ quy tắc nghiệp vụ · không biết HTTP/SQL/giờ
        ▼
   REPOSITORY + OUTBOX   internal/adapter/postgres/        ✅  pgx, tx trong ctx, upsert qua Snapshot, outbox jsonb
        │                internal/adapter/memory/          ✅  cùng port, cho test và dev run
        ▼
   WORKER                internal/worker/ + cmd/worker/    ✅  Relay: Pending → publish → MarkSent (at-least-once)
        │                platform/wire/subscribe.go        ✅  bảng định tuyến: event nào → consumer nào (on[T] decode)
        ▼
   CONTEXT THỨ HAI       internal/app/pricing/ Projector   ✅  nghe catalog.* → listings, category_profiles (idempotent)
                         internal/domain/pricing/          ✅  ShippingLane, Calculate, Quote — báo giá là ảnh chụp   → §14
   CONTEXT THỨ BA        internal/app/ordering/ Projector  ✅  nghe pricing.quote_accepted → accepted_quotes
                         internal/domain/ordering/         ✅  CustomerOrder: cọc 50 %, Cancel = điểm không thể quay đầu → §15
   CONTEXT THỨ TƯ        internal/app/procurement/         ✅  nghe ordering.deposit_paid → PurchaseTask, hỏi MerchantACL
                         internal/adapter/merchant/Manual  ✅  ACL đầu tiên (một con người); purchase_* → ordering.Reactor → §16
   CONTEXT THỨ NĂM       internal/app/logistics/           ✅  nghe purchase_confirmed → Parcel; cân thật; batch; FreightAllocator
                         pricing.Reconciler                ✅  nghe 3 context → Quote vs Actual per order → §17
```

Điểm vào thật: `cmd/api/main.go` (`-dsn` rỗng → `wire.Memory`, có → `wire.Postgres`)
và `cmd/worker/main.go` (bắt buộc `-dsn`). Cả hai gọi cùng `internal/platform/wire`.
Chạy: `docker compose up -d` rồi hai lệnh ở SETUP.md §5d.

Ba PORT tầng app cần (`internal/app/ports.go`): `Clock`, `UnitOfWork`, `Outbox`.
Ba PORT domain `catalog` cần (`internal/domain/catalog/repository.go`): `MerchantRepository`,
`CategoryRepository`, `ProductRepository`; năm PORT của `pricing` (`pricing/repository.go`):
`LaneRepository`, `QuoteRepository`, `ListingRepository`, `CategoryProfileRepository`,
`ExchangeRates`. `memory`, `postgres` và `platform/clock` cài tất cả — `wire.Graph` cắm.
Hai PORT của biên (`internal/platform/auth`): `Verifier` ("token này là ai") và `Issuer`
("cắt cho tôi một chìa") — `auth.Static` cài cho dev/test, `postgres.TokenRepo` cho thật (§18b).

---

## 1. Ngoài cùng — từ `curl` tới `handler.Handle` (`internal/adapter/http/`)

**Đọc:** `server.go` → `decode.go` → `products.go` → `errors.go`. Package tên
`httpapi` để không trùng `net/http`.

### 1a. Đây là thứ thật sự chạy — transcript `curl` vào `api.exe` (04/09, 16:32)

```
### 1. POST /merchants   Accept-Language: vi   threshold "50,00"
HTTP/1.1 201 Created
{"id":"01a06bc2-fefb-7f38-a4dc-279e796ba321"}

### 2. POST /products    merchant_id = 01a06bc2-fefb-7f38-…
HTTP/1.1 201 Created
{"id":"01a06bc2-ff2c-7114-ab73-0e546829afea"}

### 3. POST /products/01a06bc2-ff2c-7114-…/publish      (draft, chưa có variant)
HTTP/1.1 409 Conflict
{"error":{"code":"no_variants","message":"publish \"Air Trainer 90\": product has no variants"}}

### 4. POST /products    cùng URL lần 2                    (nghi trùng — vẫn 201, nhưng bị cắm cờ)
HTTP/1.1 201 Created
{"id":"01a06bc2-ff4d-7310-93e7-bcdb77f1b022"}

### 5. POST /merchants   body '{"name": '
HTTP/1.1 400 Bad Request
{"error":{"code":"bad_json","message":"malformed json: unexpected EOF"}}

### 6. POST /products    merchant_id = "MERCHANT_ID"
HTTP/1.1 400 Bad Request
{"error":{"code":"invalid_id","message":"merchant id: parse id \"MERCHANT_ID\": invalid id"}}
```

Nhìn ba id: `01a06bc2-fefb-7f38`, `01a06bc2-ff2c-7114`, `01a06bc2-ff4d-7310` —
cùng tiền tố (mili-giây sinh), chữ `7` ở vị trí 14 (phiên bản), **tăng dần**.
Đó là UUID v7 (DDD.md §12) nhìn bằng mắt thường.

### 1b. `server.go` — bảng route là code, không phải annotation

```go
type Deps struct {              // một server, nhiều bounded context — route của context nào chỉ gọi use case của context đó
	Catalog catalogapp.Deps
	Pricing pricingapp.Deps
	…
	Auth   auth.Verifier        // "token này là ai" — §18
	Tokens auth.Issuer          // "cắt cho tôi một chìa" — chỉ POST /tokens dùng
}

func NewHandler(d Deps) http.Handler {
	if d.Auth == nil {
		panic("httpapi: NewHandler needs an auth.Verifier — an API with no door is a bug, not a configuration")
	}
	s := &server{ register: catalogapp.NewRegisterMerchantHandler(d.Catalog), … , issue: pricingapp.NewIssueQuoteHandler(d.Pricing), … }
	mux := http.NewServeMux()
	// Quyền ghi CÙNG DÒNG với đường dẫn → bảng route và bảng quyền không thể lệch nhau (§18c)
	mux.HandleFunc("POST /tokens", requireOperator(s.issueToken))                                  // §18g
	// catalog — tham chiếu + operator
	mux.HandleFunc("POST /categories", requireOperator(s.defineCategory))
	mux.HandleFunc("POST /merchants", requireOperator(s.registerMerchant))
	mux.HandleFunc("POST /products", requireAny(s.addProduct))                                     // §18e — token quyết định provenance
	mux.HandleFunc("POST /products/{id}/variants", requireOperator(s.addVariant))                  // §14h
	mux.HandleFunc("POST /products/{id}/confirm-listing", requireOperator(s.confirmListing))       // §14h
	mux.HandleFunc("POST /products/{id}/measure", requireOperator(s.measureProduct))               // §14h
	mux.HandleFunc("POST /products/{id}/publish", requireOperator(s.publishProduct))
	// pricing — hỏi giá thì cả hai, chấp giá thì KHÁCH
	mux.HandleFunc("POST /quotes", requireAny(s.issueQuote))                                       // §14e
	mux.HandleFunc("GET /quotes/{id}", requireAny(s.getQuote))                                     // read model đầu tiên
	mux.HandleFunc("POST /quotes/{id}/accept", requireCustomer(s.acceptQuote))
	// ordering — khách đặt đơn CỦA MÌNH; không còn field customer_id để đặt hộ tên người khác
	mux.HandleFunc("POST /orders", requireCustomer(s.placeOrder))                                  // §15d
	mux.HandleFunc("GET /orders/{id}", requireAny(s.getOrder))                                     // chủ đơn hoặc operator; người lạ → 404 (§18d)
	mux.HandleFunc("POST /orders/{id}/cancel", requireAny(s.cancelOrder))
	mux.HandleFunc("POST /orders/{id}/deposit", requireOperator(s.payDeposit))                     // tới khi có cổng thanh toán
	mux.HandleFunc("POST /orders/{id}/balance", requireOperator(s.payBalance))
	mux.HandleFunc("POST /orders/{id}/deliver", requireOperator(s.deliverOrder))
	// procurement — người mua
	mux.HandleFunc("GET /purchase-tasks", requireOperator(s.listOpenTasks))                        // §16e — danh sách việc
	mux.HandleFunc("GET /purchase-tasks/{id}", requireOperator(s.getTask))
	mux.HandleFunc("POST /purchase-tasks/{id}/confirm", requireOperator(s.confirmTask))
	mux.HandleFunc("POST /purchase-tasks/{id}/fail", requireOperator(s.failTask))
	// pricing (operator) + logistics — kho
	mux.HandleFunc("POST /lanes", requireOperator(s.defineLane))                                   // §17d
	mux.HandleFunc("GET /reconciliations/{order}", requireOperator(s.getReconciliation))            // §17d — Quote vs Actual
	mux.HandleFunc("GET /parcels", requireOperator(s.listPendingParcels))                           // §17f
	mux.HandleFunc("GET /parcels/{id}", requireOperator(s.getParcel))
	mux.HandleFunc("POST /parcels/{id}/receive", requireOperator(s.receiveParcel))
	mux.HandleFunc("POST /batches", requireOperator(s.openBatch))
	mux.HandleFunc("GET /batches/{id}", requireOperator(s.getBatch))
	mux.HandleFunc("POST /batches/{id}/parcels", requireOperator(s.addParcelToBatch))
	mux.HandleFunc("POST /batches/{id}/close", requireOperator(s.closeBatch))
	mux.HandleFunc("POST /batches/{id}/ship", requireOperator(s.shipBatch))                        // 200 [allocation] — ngoại lệ có lý

	// authenticate bọc CẢ mux, không bọc từng route: route thêm ngày mai ĐÃ ở
	// sau cửa trước khi có ai kịp nhớ ra phải bọc nó (§18c).
	return authenticate(d.Auth)(mux)
}
```

Ngày 1 có ba route và **không có GET**: đọc là việc của read model (DDD.md §24),
không nhồi vào endpoint ghi. Ngày 2 có `GET /quotes/{id}` — read model đầu tiên, dạng
đơn giản nhất (§14e); phần còn lại vẫn không có GET. `net/http` từ Go 1.22 tự route theo method + pattern
và trả `r.PathValue("id")` — không cần framework.

**Symfony:** `#[Route('/products', methods: ['POST'])]` trên Controller — ở đây
bảng route đọc một chỗ là thấy hết.

### 1c. `products.go` — ba việc của adapter, và chỉ ba

```go
func (s *server) addProduct(w http.ResponseWriter, r *http.Request) {
	var req addProductRequest                                     // ① DTO toàn string
	if err := decodeJSON(r, &req); err != nil { writeError(w, err); return }

	merchant, err := catalog.ParseMerchantID(req.MerchantID)      // ② string → value object,
	category, err := catalog.ParseCategoryCode(req.Category)      //    bằng Parse* CỦA DOMAIN —
	source, err := catalog.ParseSourceURL(req.SourceURL)          //    adapter không tự validate
	currency, err := shared.CurrencyFromCode(req.Currency)
	price, err := shared.ParseMoney(normalizeAmount(req.Price, language(r)), currency)  // ③ LOCALE
	operator, _ := operatorOf(r)                                  // ④ danh tính từ TOKEN,
	sourcedBy := sourcingOf(r)                                    //    không từ header người gọi tự khai (§18e)

	id, err := s.add.Handle(r.Context(), catalogapp.AddProduct{...})   // ⑤ gọi use case
	if err != nil { writeError(w, err); return }                        // ⑥ map lỗi
	writeJSON(w, http.StatusCreated, idResponse{ID: id.String()})
}
```

| # | Học được | Symfony |
|---|---|---|
| ① | Request là **string hết**: dây không biết `Money` là gì. `DisallowUnknownFields()` trong `decodeJSON` → gõ nhầm `"nmae"` là 400, không phải tên rỗng âm thầm | DTO + Serializer (`ALLOW_EXTRA_ATTRIBUTES=false`) |
| ② | Adapter **không viết validate riêng** — gọi `Parse*` của domain. Một luật, một chỗ | FormType gọi vào constructor VO |
| ③ | Chỗ doc hứa từ sáng: `normalizeAmount("50,00", "vi") = "50.00"`; với `en` = `"5000"`. **Adapter đoán theo locale, domain không đoán.** Test `…localeDecidesTheDecimalSeparator` ghi cả hai kết quả | `NumberFormatter` trong `MoneyType` |
| ④ | Danh tính người gọi vào từ **mép** hệ thống — và từ 06/09 là từ **token**, không phải header `X-Operator-ID` mà chính người gọi viết ra. `sourcingOf(r)` suy luôn provenance: operator → verified, khách → draft (§18e) | `$this->getUser()` |
| ⑤ | Adapter **không biết** repository, clock, outbox — chỉ biết `Handle` | Controller → `$bus->dispatch()` |
| ⑥ | Mọi lỗi qua **một** hàm | ExceptionListener |

### 1d. `errors.go` — một bảng, ba cuộc hội thoại

```go
var errorTable = []mapping{
	// 400 — request không hiểu được → sửa request
	{errBadJSON, 400, "bad_json"}, {shared.ErrMalformedAmount, 400, "malformed_amount"},
	{shared.ErrInvalidID, 400, "invalid_id"}, {catalog.ErrInvalidHostname, 400, "invalid_hostname"}, ...
	// 404 — trỏ vào thứ không tồn tại
	{catalog.ErrMerchantNotFound, 404, "merchant_not_found"}, ...
	// 409 — request đúng, NGHIỆP VỤ nói không → sửa trạng thái
	{catalog.ErrNoVariants, 409, "no_variants"}, {catalog.ErrUnverified, 409, "unverified"},
	{catalogapp.ErrMerchantInactive, 409, "merchant_inactive"}, {catalogapp.ErrPriceCurrency, 409, "price_currency"}, ...
}

func writeError(w http.ResponseWriter, err error) {
	for _, m := range errorTable {
		if errors.Is(err, m.target) { ... m.status, m.code, err.Error() ... return }
	}
	log.Printf("httpapi: unmapped error: %v", err)      // không có trong bảng = bug hoặc sự cố
	... 500, "internal", "internal error"                // không lộ nội bộ
}
```

`errors.Is` đi hết chuỗi `%w`: lỗi domain bọc hai lớp (`"publish \"x\": product has
no variants"`) vẫn tìm đúng dòng. `code` là slug **ổn định** cho client `switch`;
`message` là chuỗi của domain — nên body lỗi ở transcript 1a chính là câu domain
đã nói.

**Câu hỏi hay gặp:** vì sao `ErrPriceCurrency` là 409 chứ không phải 400? Vì
request **đúng hoàn toàn** — `"3900000" VND` parse được — chỉ là merchant này bán
USD. Sửa request không giúp; phải đổi merchant hoặc đổi tiền. Đó là 409.

### 1e. `cmd/api/main.go` — `main()` chỉ cờ + serve, `wire.Graph` là toàn bộ dây nối

```go
func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dsn := flag.String("dsn", "", "Postgres DSN; empty runs in memory")
	flag.Parse()

	var g wire.Graph
	if *dsn == "" {
		g = wire.Memory(clock.System{})            // chỗ DUY NHẤT chọn đồng hồ thật
		bus := worker.NewBus()                     // in-memory: relay chạy TRONG process này (§14g)
		wire.Subscribe(bus, g)
		go worker.New(worker.Deps{Source: g.Source, Publisher: bus, UoW: g.Catalog.UoW, Clock: g.Catalog.Clock, Interval: 200 * time.Millisecond}).Run(ctx)
	} else {
		g, closeFn, err = wire.Postgres(ctx, clock.System{}, *dsn)   // sai DSN → chết ở đây, không phải ở request đầu
		defer closeFn()                                              // relay là việc của cmd/worker
	}
	srv := &http.Server{Addr: *addr, Handler: httpapi.NewHandler(httpapi.Deps{Catalog: g.Catalog, Pricing: g.Pricing}), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.ListenAndServe())
}
```

`wire.Graph` (`internal/platform/wire/wire.go`, §14g) là object graph: một `Deps` cho mỗi
bounded context, dùng chung `Clock`/`UoW`/`Outbox`, cộng `Source` (outbox nhìn từ relay).
`wire.Memory` và `wire.Postgres` trả **cùng một `Graph`** — handler và use case không đổi
một dòng khi đổi adapter. `cmd/worker` gọi lại đúng `wire.Postgres` + `wire.Subscribe` —
không chép `main`, không lệch cấu hình.

So với `world` trong test (§2): **cùng struct `Deps`, khác giá trị** — `clock.System`
thay `clock.Fixed`, còn lại y hệt.

**Vì sao `cmd/worker` không có chế độ in-memory, còn `cmd/api` in-memory lại tự chạy
relay** (SETUP.md §9 đợt 8, đợt 10): outbox in-memory nằm trong process của `api` — process
khác không đọc được. Stub "chạy lên, in thông báo, thoát 0" tệ hơn không có gì: outbox đầy
dần mà không ai biết. *Im lặng là chế độ hỏng tệ nhất.* Còn relay-trong-api ở chế độ memory
là relay **thật**, chạy đúng `Subscribe` — không có nó thì `POST /quotes` in-memory mãi 404.

**Symfony:** `public/index.php` + `services.yaml` + `.env` — viết tay, 60 dòng,
không container.

### 1f. Test của adapter — `server_test.go`

`httptest.NewRequest` → `handler.ServeHTTP(rec, req)` → soi `rec.Code`, body, và
**trạng thái repo** sau đó. Không mở port. 8 test, 85,7% coverage. Hai test đáng
đọc: `…localeDecidesTheDecimalSeparator` (cùng bytes `"50,00"`, hai số tiền) và
`…400vs404vs409` (ba mã cho ba loại "không").

---

## 2. Bắt đầu từ test — `internal/app/catalog/add_product_test.go`

**Đọc:** struct `world`, hàm `newWorld()`, rồi `TestAddProduct_assemblesProvenanceAndSaves`.

```go
type world struct {
	clock      *clock.Fixed
	merchants  *memory.MerchantRepo
	categories *memory.CategoryRepo
	products   *memory.ProductRepo
	outbox     *memory.Outbox
	deps       catalogapp.Deps
}
```

`world` chính là **dây nối** (wiring) — mọi thứ một handler cần, cắm bằng tay.
`main()` sau này làm **y hệt**, chỉ thay `memory.*` bằng `postgres.*` và
`clock.Fixed` bằng `clock.System`. Không có container, không có autowiring.

**Học được:** test của tầng app là **integration test trong bộ nhớ** — đi qua
thật: handler → domain → repository → outbox, chỉ không có I/O. Chạy trong vài
mili-giây, không cần Docker.

**Symfony:** `KernelTestCase` + `self::getContainer()`; ở đây "container" là một
struct 6 field điền tay.

Đọc tiếp 3 assert cuối của test — đó là **spec** của use case:

| Assert | Nói lên điều gì |
|---|---|
| `p.Status() == ProductStatusDraft` | ghi nhận = nháp, chưa quote chắc chắn được |
| `prov.Source() == SourcedByCustomer`, `prov.At().Equal(now)`, `!prov.Verified()` | use case đã **ghép** đồng hồ + nguồn dữ liệu thành `Provenance` |
| outbox có đúng 1 `catalog.product_added` | event ra khỏi aggregate, vào outbox — và **chỉ** ra qua đường đó |

---

## 3. Use case — `internal/app/catalog/add_product.go`

**Đọc:** struct `AddProduct`, rồi `Handle` từ trên xuống.

### 2a. Command khác Details ở đúng hai field

```go
type AddProduct struct {
	Name, Merchant, Category, Source, Price   // ← giống catalog.ProductDetails
	SourcedBy catalog.SourcingMode            // ← dữ liệu của REQUEST: dữ liệu tới bằng đường nào
	Operator  shared.OperatorID               // ← dữ liệu của REQUEST: ai (từ auth)
}
```

Domain muốn `Provenance` (nguồn + lúc nào + ai). Request chỉ biết *nguồn* và
*ai*; *lúc nào* là của đồng hồ. **Ghép ba thứ đó lại là việc của tầng app** —
đây là lý do tồn tại của tầng này, không phải "cho đủ tầng".

**Symfony:** Command của Messenger là DTO thuần; `ProductDetails` thì gần với
tham số constructor của Entity. Hai thứ khác nhau vì phục vụ hai người gọi khác nhau.

### 2b. `Handle` — đọc theo số

```go
func (h *AddProductHandler) Handle(ctx context.Context, cmd AddProduct) (catalog.ProductID, error) {
	now := h.deps.Clock.Now()                                    // ⓪ đọc giờ MỘT lần, ở đây

	var id catalog.ProductID
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {  // ① một transaction bao hết
		m, err := h.deps.Merchants.ByID(ctx, cmd.Merchant)         // ② tải aggregate KHÁC
		if err != nil { return err }                                //    → ErrMerchantNotFound đi thẳng ra
		if !m.IsActive() {                                          // ③ LUẬT LIÊN AGGREGATE #1
			return fmt.Errorf("...: %w", ErrMerchantInactive)
		}
		if cmd.Price.IsValid() && cmd.Price.Currency() != m.Currency() {   // ④ LUẬT LIÊN AGGREGATE #2
			return fmt.Errorf("...: %w", ErrPriceCurrency)
		}
		if _, err := h.deps.Categories.ByCode(ctx, cmd.Category); err != nil {  // ⑤ category phải tồn tại
			return err
		}

		prov, err := catalog.NewProvenance(cmd.SourcedBy, now, cmd.Operator)  // ⑥ GHÉP: nguồn + giờ + ai
		if err != nil { return err }
		p, err := catalog.AddProduct(catalog.ProductDetails{...}, now)         // ⑦ DOMAIN quyết định
		if err != nil { return err }

		dups, err := h.deps.Products.BySource(ctx, cmd.Source)     // ⑧ phát hiện trùng (phương án C)
		if len(dups) > 0 {
			p.FlagDuplicateOf(dups[0].ID(), "same page url", now)
		}

		if err := h.deps.Products.Save(ctx, p); err != nil {...}                 // ⑨ LƯU TRƯỚC
		if err := h.deps.Outbox.Append(ctx, p.PullEvents()); err != nil {...}   // ⑩ RỒI MỚI PULL
		id = p.ID()
		return nil
	})
	return id, err
}
```

| # | Học được | Bên Symfony |
|---|---|---|
| ⓪ | `now` đọc **một lần** rồi truyền xuống. Domain không có `time.Now()` — test canh số 2 đảm bảo | `ClockInterface` inject vào handler |
| ① | Go không có method generic → `InTx` trả `error`, kết quả lấy qua biến closure `id` | `$em->wrapInTransaction(fn)` |
| ③④ | Hai luật này **không thuộc `Product`** (nó không biết merchant đang bán tiền gì) và **không thuộc `Merchant`** (nó không biết product). Luật đụng 2 aggregate → tầng app. Mỗi luật một sentinel error riêng | logic này bên Symfony hay rơi vào Service hoặc Validator |
| ⑥ | Use case **dịch** ngôn ngữ request → ngôn ngữ domain | Handler tạo Value Object từ Command |
| ⑦ | Domain **có quyền nói không** (`ErrEmptyName`, `ErrPriceRequired`…) và lỗi đó trả **nguyên vẹn** — không bọc, không đổi | `throw new DomainException` trong Entity |
| ⑧ | Detector ở tầng app (hỏi repository), **quyết định** ở domain (`FlagDuplicateOf` từ chối trỏ chính mình, nhớ cặp đã bác) | — |
| ⑨⑩ | **Thứ tự sống còn.** Save lỗi → `PullEvents` không bao giờ chạy → outbox rỗng → không worker nào đi mua giày cho sản phẩm không tồn tại | `postFlush` listener — nhưng ở đây tường minh, không phải hook ngầm |

### 2c. Lỗi tầng app vs lỗi domain — `deps.go`

```go
var (
	ErrMerchantInactive = errors.New("merchant is not active")
	ErrPriceCurrency    = errors.New("price is not in the merchant's currency")
)
```

Chỉ **hai** lỗi thuộc tầng app, và cả hai đều **liên aggregate**. Nếu bạn thấy
mình muốn thêm lỗi thứ ba vào đây, hỏi: *nó có cần hai aggregate không?* Không →
nó thuộc domain.

### 2d. `Deps` và `mustHave` — dây nối kiểm tra lúc khởi động

```go
type Deps struct {
	Clock app.Clock; UoW app.UnitOfWork
	Merchants catalog.MerchantRepository; Categories catalog.CategoryRepository; Products catalog.ProductRepository
	Outbox app.Outbox
}

func NewAddProductHandler(d Deps) *AddProductHandler {
	mustHave("AddProductHandler", map[string]any{"Clock": d.Clock, ... })  // nil → panic ngay
	...
}
```

Quy ước 10 (struct tham số > 3 field) + quy ước 1 (input của lập trình viên sai
→ panic). Thiếu dependency là **bug lúc khởi động**, không phải lỗi lúc request.
Test `TestNewRegisterMerchantHandler_panicsOnMissingDeps` chứng minh.

**Symfony:** container compile lỗi khi service thiếu argument — cùng thời điểm
(khởi động), cùng mức (chết ngay).

---

## 4. Ports của tầng app — `internal/app/ports.go`

**Đọc:** cả file, 3 interface.

```go
type Clock interface      { Now() time.Time }
type Outbox interface     { Append(ctx context.Context, events []shared.Event) error }
type UnitOfWork interface { InTx(ctx context.Context, fn func(ctx context.Context) error) error }
```

**Học được:** interface **khai báo ở nơi cần**, không phải nơi cài. Tầng app cần
đồng hồ → tầng app khai báo `Clock`. Ai cài là chuyện của `platform/clock`. Đây
là Dependency Inversion bằng chính cú pháp Go: `memory.Outbox` không biết
`app.Outbox` tồn tại, nó chỉ tình cờ có đủ method.

Dòng này trong mỗi adapter là "bằng chứng" compile-time:

```go
var _ catalog.MerchantRepository = (*MerchantRepo)(nil)   // memory/merchant_repo.go
```

Interface thêm method → dòng này **không compile** → biết ngay, trước cả khi chạy test.

**Symfony:** `implements MerchantRepositoryInterface` — Go không có từ khoá đó,
nên dùng dòng khẳng định trên thay thế.

---

## 5. Domain — hai hàm được gọi, đọc đúng hai hàm

### 4a. `catalog.NewProvenance` — `internal/domain/catalog/provenance.go`

```go
func NewProvenance(source SourcingMode, at time.Time, by shared.OperatorID) (Provenance, error)
func (p Provenance) Verified() bool { return p.source == SourcedByOperator }
```

Một khẳng định "dữ liệu này từ đâu" — và **chỉ operator** mới là `Verified()`.
Trong test, `SourcedByCustomer` → `prov.Verified() == false` → sản phẩm là nháp
đúng nghĩa.

### 4b. `catalog.AddProduct` — `internal/domain/catalog/product.go`

**Đọc:** `AddProduct`, rồi struct `Product`, rồi `Publish`.

```go
func AddProduct(d ProductDetails, now time.Time) (*Product, error) {
	// validate TỪNG field bắt buộc — quy ước 10; test ...everyFieldIsValidated zero từng field
	...
	p := &Product{ ..., status: ProductStatusDraft, addedAt: now }
	p.Record(ProductAdded{ID: p.id, ...})      // ← GHI event, không phát
	return p, nil
}
```

`p.Record(...)` đến từ `shared.Events` được **nhúng** vào `Product`
(`internal/domain/shared/event.go`). Aggregate chỉ **ghi** vào sổ; ai lấy sổ ra
(`PullEvents`) và lúc nào là việc của tầng app — bước ⑩ ở trên.

**Học được:** vì sao `Product` không tự publish: nó không biết đã được lưu chưa.
Chỉ tầng app biết `Save` đã thành công.

### 4c. `Product.FlagDuplicateOf` — cùng file

```go
func (p *Product) FlagDuplicateOf(other ProductID, reason string, now time.Time) error {
	if other.IsZero() || other == p.id { return ...ErrInvalidDuplicate }
	if reason == "" { return ...ErrEmptyReason }
	if p.suspectedDuplicateOf == other || slices.Contains(p.dismissedDuplicates, other) {
		return nil     // ← đã gắn, hoặc operator ĐÃ BÁC cặp này → không làm gì, không event
	}
	...
	p.Record(ProductFlaggedDuplicate{..., Reason: reason})
}
```

Detector (tầng app) **đề nghị**, aggregate **quyết**. Cặp operator đã bác thì
detector chạy lại cũng không gắn được — vòng lặp "gắn → xoá → gắn" bị chặn
**trong domain**, không trông vào ai xây read model. (CATALOG.md §7)

---

## 6. Adapter — hai cài đặt cho cùng một port

Cùng `catalog.ProductRepository`, `app.UnitOfWork`, `app.Outbox` — hai package cài:

| | `internal/adapter/memory/` | `internal/adapter/postgres/` |
|---|---|---|
| Dùng khi | test, `go run ./cmd/api` không cờ | `-dsn`, CI, production |
| Lưu | con trỏ `*Product` trong map | cột trong bảng, qua `Snapshot()`/`FromSnapshot()` |
| Transaction | không có (`InTx` gọi thẳng `fn`) | `BEGIN` → tx bỏ vào `ctx` → `COMMIT`/`ROLLBACK` |
| Outbox | slice, `Drain()` cho test | bảng `outbox`, `payload jsonb` từ `eventcodec` |
| Test | 5 unit | 11 integration (6 catalog + 5 pricing), skip nếu không có `PORTAGE_TEST_DSN` |

Handler và use case **không biết** đang chạy cái nào — `wire.Memory` và `wire.Postgres`
trả về cùng một `wire.Graph`, và `TestMemory_/TestPostgres_runsTheWholeFlow` (§14j) chạy
**cả flow** — paste → publish → relay → quote → accept — trên cả hai.

### 6a. Cửa sau của aggregate — `internal/domain/catalog/snapshot.go`

Aggregate toàn field private (đúng, đó là cách giữ invariant). Vậy Postgres dựng lại
`*Product` từ 25 cột thế nào? **Memento**:

```go
type ProductSnapshot struct { ID; Merchant; Category; Name; Source; ListingProvenance; Price; ... ; Variants []VariantSnapshot; Status; AddedAt }
func (p *Product) Snapshot() ProductSnapshot                  // trạng thái đi RA — copy, không chia sẻ memory
func ProductFromSnapshot(s ProductSnapshot) (*Product, error) // trạng thái đi VÀO — không phát event
```

Hai quy tắc của `FromSnapshot`: **tin** giá trị nghiệp vụ (đã validate lúc đi vào),
nhưng **từ chối** hình dạng `Snapshot()` không thể tạo ra — published mà không có
variant, parcel mà không có provenance → `ErrInvalidSnapshot`. Đó là dữ liệu hỏng, và
dựng lên một aggregate nói dối thì tệ hơn báo lỗi.

**Symfony:** Doctrine ghi thẳng vào private field bằng Reflection — ngầm. Ở đây tường
minh: `Snapshot()` ~ `toArray()`, `FromSnapshot` ~ `fromArray()`, và **chỉ repository**
gọi. Test `TestProduct_snapshotRoundTrip`: đi qua mọi method nghiệp vụ → snapshot →
dựng lại → snapshot phải `DeepEqual`.

### 6b. `postgres.UnitOfWork` — transaction đi theo `ctx`

```go
func (u *UnitOfWork) InTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := u.pool.Begin(ctx)
	defer func() { if err != nil { tx.Rollback(ctx) } else { err = tx.Commit(ctx) } }()   // (rút gọn)
	return fn(context.WithValue(ctx, txKey{}, tx))                                        // tx nằm TRONG ctx
}

func db(ctx context.Context, pool *pgxpool.Pool) querier {   // mọi repo hỏi hàm này
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok { return tx }
	return pool
}
```

Handler viết `Save(ctx, m)` rồi `Outbox.Append(ctx, evs)` — cả hai lấy `tx` từ `ctx`
→ **một transaction**, không ai phải truyền `tx` qua tham số. Đây là cách Go thay cho
"EntityManager giữ transaction trong trạng thái toàn cục của request".

**Test từng bị skip giờ chạy thật:** `TestUnitOfWork_rollsBackTheSaveWhenTheOutboxFails`
— outbox giả hỏng → `count(*) FROM merchants == 0`. Lời hứa nguyên tử của §27 có bằng chứng.

### 6c. `postgres.Outbox` + `eventcodec` — payload là hợp đồng

```go
// eventcodec/codec.go — MỘT case cho MỘT event, key viết tay
case catalog.ProductRepriced:
	return m{"id": e.ID.String(), "from": money(e.From), "to": money(e.To), "at": ts(e.At)}, nil
```

Vì sao không `json.Marshal(event)`? Vì consumer ở context khác **không thể** load
aggregate của catalog (guard 7 cấm import) — payload là **tất cả** nó biết. Để
`json.Marshal` suy shape từ struct thì đổi tên field là đổi wire format mà vẫn compile.
Viết tay: đổi struct → mapper không compile; thêm event → guard
`TestEncode_contractCoversEveryDomainEvent` (đọc `events.go` bằng `go/ast`) đỏ.

### 6d. Migration — `postgres/migrate.go` + `migrations/0001_catalog.sql`

SQL thuần nhúng bằng `embed.FS`, áp theo tên file, ghi vào `schema_migrations`, giữ
`pg_advisory_xact_lock` để hai process khởi động cùng lúc không áp trùng. 40 dòng thay
cho một tool. Cột phản chiếu snapshot 1-1: tiền = minor + mã ISO, cân = gram, kích thước
= mm, id = uuid v7, giờ = timestamptz.

### 6e. Bản in-memory — `memory/`

**Đọc:** `product_repo.go` (`Save`, `BySource`), `memory.go` (`Outbox`, `UnitOfWork`).

```go
func (r *ProductRepo) Save(ctx context.Context, p *catalog.Product) error {
	if r.FailSaveWith != nil { return r.FailSaveWith }   // ← nút cho test bóp lỗi
	r.byID[p.ID()] = p                                    // ← giữ CON TRỎ
	return nil
}
```

Repo memory giữ **con trỏ**, không đi qua snapshot — vì Postgres đã làm và test việc
đó; ở đây copy chỉ làm test chậm hơn.

`UnitOfWork.InTx` chỉ gọi `fn(ctx)` — bộ nhớ không có transaction. Nó tồn tại để
handler **được nối y hệt** trong test và trong production.

Memory **không có transaction**: Save xong rồi Outbox hỏng → merchant vẫn nằm lại.
`TestRegisterMerchant_outboxFailureRollsBackTheSave` vì thế `t.Skip` với memory và
**chạy thật với Postgres** (§6b). Skip, không xoá: test bị skip hiện trong `go test -v`,
comment thì không.

`Outbox.Drain()` giả vai worker: lấy hết event ra theo thứ tự, làm rỗng.

**Symfony:** đây là `InMemoryRepository` bạn hay viết trong `tests/` — nhưng nằm
ở `adapter/` như một cài đặt bình đẳng, vì `cmd/api` chạy dev có thể dùng nó.

---

## 7. Kết quả — outbox thật, worker thật (transcript 04/09, 17:28)

Chạy `api.exe -dsn …` và `worker.exe -dsn … -every 300ms`, bắn 4 request:

```
### 1. POST /merchants                         → 201 {"id":"01a06bf6-27e5-7502-…"}
### 2. POST /products                          → 201 {"id":"01a06bf6-28a1-72f2-…"}
### 3. POST /products  cùng URL lần 2          → 201 {"id":"01a06bf6-28e8-71c9-…"}   (bị cắm cờ)
### 4. POST /products/…28a1…/publish           → 409 no_variants

### bảng outbox
 id |            event_name             | sent
----+-----------------------------------+------
  1 | catalog.merchant_registered       | t
  2 | catalog.product_added             | t
  3 | catalog.product_added             | t
  4 | catalog.product_flagged_duplicate | t

### worker log
event #1 catalog.merchant_registered       {"at": "…", "id": "01a06bf6-27e5-…", "name": "Example Sports", "site": "www.example.com"}
event #2 catalog.product_added             {"at": "…", "id": "01a06bf6-28a1-…", "name": "Air Trainer 90", "category": "footwear", "merchant": "01a06bf6-27e5-…"}
event #3 catalog.product_added             {"at": "…", "id": "01a06bf6-28e8-…", …}
event #4 catalog.product_flagged_duplicate {"at": "…", "id": "01a06bf6-28e8-…", "of": "01a06bf6-28a1-…", "reason": "same page url"}
```

Đọc theo tầng: request 3 ghi **hai** dòng outbox (3 = `product_added` của bản trùng,
4 = cờ nghi trùng) trong **cùng transaction** với dòng `products`; worker đọc theo `id`,
publish, `sent_at` được ghi. Payload đúng từng key của hợp đồng ở `codec_test.go`.

### 7a. Tầng 5 — `internal/worker/worker.go`

```go
func (r *Relay) RunOnce(ctx context.Context) (sent int, err error) {
	var pubErr error
	txErr := r.deps.UoW.InTx(ctx, func(ctx context.Context) error {   // MỘT transaction cho cả pass
		entries, _ := r.deps.Source.Pending(ctx, r.deps.Batch)         // SELECT … WHERE sent_at IS NULL ORDER BY id FOR UPDATE SKIP LOCKED
		var done []int64
		for _, e := range entries {
			if err := r.deps.Publisher.Publish(ctx, e); err != nil {   // publish THEO THỨ TỰ
				pubErr = …; break                                       // hỏng ở đâu, dừng ở đó — không nhảy cóc
			}
			done = append(done, e.ID)
		}
		return r.deps.Source.MarkSent(ctx, done, r.deps.Clock.Now())    // chỉ đánh dấu phần ĐÃ publish
	})
	…
}
```

| Học được | |
|---|---|
| **At-least-once** | publish trước, `MarkSent` sau → sập giữa chừng thì tick sau gửi **lại**. Không mất; có thể thấy hai lần → subscriber phải **idempotent**. Test `TestRelay_failureKeepsTheRestPendingAndRetries` |
| Thứ tự | hỏng ở event k → k và sau đó vẫn pending, tick sau đọc lại từ k. Không bỏ qua để giữ thứ tự trong outbox |
| Nhiều worker | `FOR UPDATE SKIP LOCKED` trong tx → hai worker không lấy trùng dòng |
| `Bus` | in-process, đăng ký theo tên event, `"*"` nghe tất; lỗi handler = lỗi publish → dòng bị giữ lại |
| Subscriber hôm nay | một dòng log trong `cmd/worker`. Cái thật đầu tiên: Pricing nghe `product_published` |

`memory.Outbox` cũng cài `worker.Source` → 6 test của relay chạy không cần DB.

**Symfony:** `messenger:consume` — nhưng đọc từ bảng outbox, và ~120 dòng để thấy hết
vòng đời một message.

### 7b. `cmd/worker/main.go` — bắt buộc `-dsn`

Không có chế độ in-memory: outbox memory nằm trong process của api, worker khác process
không đọc được. Binary từ chối khởi động thay vì "chạy lên, không làm gì, exit 0"
(SETUP.md §9 đợt 8). `g.Source` là outbox Postgres nhìn từ phía đọc (`worker.Source`) —
cùng object với `g.Catalog.Outbox` (`app.Outbox`), `wire` cắm cả hai vai. `wire.Subscribe(bus, g)`
gắn consumer thật (projector của pricing, §14b); subscriber `"*"` in log chỉ để mắt thấy.

---

## 8. Cùng flow, hai use case còn lại — khác ở đâu

| | `RegisterMerchantHandler` | `PublishProductHandler` |
|---|---|---|
| File | `register_merchant.go` | `publish_product.go` |
| Command | **chính là** `catalog.MerchantDetails` — bọc thêm struct trùng 100% field là copy đổi tên | một `ProductID` |
| Hình | tạo mới → Save → Pull | **load → hỏi aggregate → Save → Pull** |
| Luật tầng app | không có | không có — mọi "không" là của `Product.Publish`, trả nguyên |
| Học | dạng đơn giản nhất của flow | dạng "load-edit-flush" quen thuộc bên Symfony, nhưng thứ tự tường minh |

Ba handler, một khuôn: `now` → `InTx` → (load) → domain → Save → Pull → Outbox.
Khi viết use case thứ tư, chép khuôn này.

---

## 9. Ba test chứng minh thứ tự — không phải comment

| Test | Chứng minh |
|---|---|
| `TestRegisterMerchant_saveFailurePublishesNothing` | `FailSaveWith = boom` → `Handle` trả lỗi bọc `boom`, **outbox rỗng**. Save-trước-Pull là hành vi, không phải lời hứa |
| `TestRegisterMerchant_domainErrorLeavesNoTrace` | domain từ chối → repo `Len() == 0`, outbox rỗng, lỗi `errors.Is` được `ErrEmptyName` |
| `TestNewRegisterMerchantHandler_panicsOnMissingDeps` | thiếu dependency → panic lúc dựng handler, không đợi request |
| `TestRegisterMerchant_outboxFailureRollsBackTheSave` **(SKIP với memory)** | chiều ngược: Save xong, outbox hỏng → phải rollback. Memory không làm được — test nằm sẵn chờ Postgres |

> Nếu bạn muốn hiểu một quy tắc, tìm test **cố tình vi phạm** nó. Không có test
> như vậy = quy tắc chưa được chứng minh (bài học đợt 5, SETUP.md §9).

---

## 10. Thứ tự đọc toàn repo — cho người bắt đầu từ số 0

Mỗi dòng: file → khái niệm → đọc thêm ở đâu.

```
 1. internal/domain/shared/money.go          Value Object bất biến, không float, error thay exception   DDD.md §11
 2. internal/domain/shared/decimal.go        parse nghiêm ngặt, half-up, big.Int — máy móc dưới VO      SETUP.md quy ước 2, 3
 3. internal/domain/shared/weight.go         invariant không âm, làm tròn LÊN, hai loại constructor      DDD.md §13, quy ước 1, 4
 4. internal/domain/shared/id.go             UUID v7, typed ID bằng embedding                            DDD.md §12, quy ước 5
 5. internal/domain/shared/event.go          Event + Events: ghi, không phát                            DDD.md §15, quy ước 6
 6. internal/domain/catalog/merchant.go      ENTITY / AGGREGATE ROOT đầu tiên, vòng đời, MerchantDetails DDD.md §14, quy ước 10
 7. internal/domain/catalog/freeshipping.go  VO 3 trạng thái, zero value an toàn                         quy ước 9
 8. internal/domain/catalog/category.go      VO cấu hình khoá tự nhiên; CategoryCode là struct           CATALOG.md §6
 9. internal/domain/catalog/parcelspec.go    một VO hai vai — độ tin nằm ở Provenance                    CATALOG.md §3, §4
10. internal/domain/catalog/provenance.go    nguồn dữ liệu theo NHÓM thuộc tính                          CATALOG.md §4
11. internal/domain/catalog/product.go       aggregate có entity con, Publish 5 điều kiện                 DDD.md §14, §28
12. internal/domain/catalog/events.go        14 event — thì quá khứ, wire name                            DDD.md §15
13. internal/domain/catalog/repository.go    PORT: domain khai báo cái nó cần                            DDD.md §16
14. internal/domain/decisions_test.go        7 guard kiến trúc bằng go/ast                               SETUP.md §6
15. internal/app/ports.go                    PORT của tầng app: Clock, UnitOfWork, Outbox                DDD.md §20, §30
16. internal/app/catalog/deps.go             Deps + mustHave — dây nối kiểm tra lúc khởi động            quy ước 10
17. internal/app/catalog/add_product.go      USE CASE — §3 ở trên                              DDD.md §30
18. internal/adapter/memory/*.go             ADAPTER in-memory, khoảng trống snapshot                    §6 ở trên
19. internal/platform/clock/clock.go         System vs Fixed                                              quy ước 7
20. internal/app/catalog/*_test.go           integration test trong bộ nhớ                               §2, §9 ở trên
21. internal/adapter/http/server.go          bảng route = code; NewHandler(Deps)                          §1b
22. internal/adapter/http/decode.go          DisallowUnknownFields, language(), normalizeAmount()         §1c
23. internal/adapter/http/products.go        ba việc của adapter, đúng thứ tự                             §1c
24. internal/adapter/http/errors.go          một bảng lỗi → 400/404/409, errors.Is                        §1d
25. cmd/api/main.go                          main() = cờ + serve; -dsn chọn wire.Memory/Postgres            §1e
26. internal/domain/catalog/snapshot.go      memento — cửa sau cho persistence, FromSnapshot từ chối hỏng    §6a
27. internal/adapter/eventcodec/codec.go     payload = hợp đồng, viết tay; guard go/ast phủ mọi event        §6c
28. internal/adapter/postgres/postgres.go    pool, querier, tx trong ctx, UnitOfWork thật                    §6b
29. internal/adapter/postgres/migrate.go     embed.FS + schema_migrations + advisory xact lock               §6d
30. internal/adapter/postgres/product_repo.go upsert 25 cột, variants delete+insert, scan → snapshot         §6
31. internal/adapter/postgres/outbox.go      Append trong tx; Pending FOR UPDATE SKIP LOCKED; MarkSent      §7a
32. internal/adapter/postgres/pgtest/        DB test có advisory lock — package song song không giẫm nhau   SETUP.md §5e
33. internal/platform/wire/wire.go           Memory / Postgres — hai object graph, một Deps                  §1e
34. internal/worker/worker.go                Relay + Bus, at-least-once                                      §7a
35. cmd/worker/main.go                       -dsn bắt buộc, wire.Subscribe + subscriber log                  §7b
── ngày 2: context thứ hai ─────────────────────────────────────────────────────────────────────────────
36. internal/domain/shared/parcelspec.go     VO hai context cùng dùng → shared kernel thật                   §14b
37. internal/domain/catalog/events.go        ProductPublished mang Price+Parcel; CategoryDefined              §14b
38. internal/contracts/catalog_v1.go         PUBLISHED LANGUAGE — V1 struct, primitive, package riêng         §14b, DDD.md §7
39. internal/adapter/eventcodec/decode.go    Decode(name, payload) → V1; into[T] generic; ErrNoDecoder        §14b
40. internal/domain/pricing/lane.go          LaneCode, GoodsClass, RateCard, DutyPolicy, ShippingLane, Classification  §14c
41. internal/domain/pricing/policy.go        MarginPolicy max(%, sàn); QuotePolicy — cấu hình chụp vào quote  §14c
42. internal/domain/pricing/listing.go       Listing, CategoryProfile — PROJECTION, field exported, shared.ID §14c
43. internal/domain/pricing/calc.go          Calculate — DOMAIN SERVICE thuần; Estimated                      §14c, DDD.md §17
44. internal/domain/pricing/quote.go         Quote — aggregate "ảnh chụp"; Accept trễ = expire + từ chối      §14c, DDD.md §28
45. internal/domain/pricing/repository.go    5 PORT của pricing                                               DDD.md §16
46. internal/app/pricing/projector.go        bên nghe: idempotent, chịu sai thứ tự, từ vựng riêng            §14b
47. internal/app/pricing/issue_quote.go      gom đầu vào → domain đóng băng → Save → outbox                   §14d
48. internal/app/pricing/accept_quote.go     commit phần đã xảy ra, rồi báo phần bị từ chối                   §14d
49. internal/app/catalog/{add_variant,confirm_listing,measure_product}.go  ba bước operator                  §14h
50. internal/adapter/http/quotes.go          POST /quotes, GET /quotes/{id} (read model), accept              §14e
51. internal/adapter/postgres/migrations/0002_pricing.sql  5 bảng, KHÔNG FK sang catalog                     §14f
52. internal/adapter/postgres/pricing_repos.go  laneFrom qua constructor; quotes từng cột; fx text            §14f
53. internal/platform/wire/wire.go           Graph{Catalog, Pricing, Source}; quotePolicy(); seed lane+fx     §14g
54. internal/platform/wire/subscribe.go      BẢNG ĐỊNH TUYẾN event → consumer; on[T]                          §14b
55. internal/platform/wire/wire_test.go      wholeFlow — test vàng, chạy trên cả hai graph                    §14j
56. scripts/smoke.ps1                        cả flow trên binary thật + Postgres thật                         §14i
── context thứ ba: ordering ────────────────────────────────────────────────────────────────────────────────
57. internal/domain/ordering/order.go        CustomerOrder — state machine viết tay; Refund; Cancel = §26     §15b, DDD.md §26
58. internal/domain/ordering/acceptedquote.go  projection từ pricing.quote_accepted                           §15a
59. internal/domain/ordering/events.go       8 event; DepositPaid mang product + variant cho procurement      §15a
60. internal/contracts/pricing_v1.go         QuoteAcceptedV1 — cạnh thứ hai của Context Map                   §15a
61. internal/app/ordering/deps.go            Deps.mutate — khuôn §8 viết một lần                              §15c
62. internal/app/ordering/place_order.go     hai luật liên aggregate: quote đã accept, một quote một đơn      §15c
63. internal/app/ordering/{payments,cancel_order,lifecycle}.go  bảy handler ~6 dòng                           §15c
64. internal/adapter/http/orders.go          6 route; refund chỉ ở GET                                        §15d
65. internal/adapter/postgres/migrations/0003_ordering.sql  orders.quote UNIQUE — DB canh nửa sau của luật    §15e
66. internal/platform/wire/wire_test.go      wholeFlow kéo tới cancel + refund                                §15f
── context thứ tư: procurement ─────────────────────────────────────────────────────────────────────────────
67. internal/domain/procurement/task.go      PurchaseTask open → confirmed | failed; PurchaseReceipt = Actual đầu tiên  §16b
68. internal/domain/procurement/repository.go  Shop/Item projection; TaskRepository; PORT MerchantACL          §16c, DDD.md §22
69. internal/adapter/merchant/manual.go      ACL adapter đầu tiên: một con người (ErrManualPurchase)             §16c
70. internal/contracts/{ordering,procurement}_v1.go  DepositPaidV1, PurchaseConfirmedV1, PurchaseFailedV1        §16a
71. internal/app/procurement/open_task.go    idempotent theo đơn; tra projection; ba nhánh của ACL               §16d
72. internal/app/procurement/projector.go    OnMerchantRegistered (Currency), OnProductPublished (listener thứ 2) §16f
73. internal/app/ordering/reactor.go         OnPurchaseConfirmed/Failed; "đã làm rồi" ≠ "không làm được"        §16d
74. internal/adapter/http/purchase_tasks.go  GET danh sách việc; confirm/fail; KHÔNG có POST tạo task            §16e
75. internal/adapter/postgres/migrations/0004_procurement.sql  purchase_tasks UNIQUE("order"); 2 bảng projection §16d
76. internal/platform/wire/subscribe.go      bảng định tuyến = saga §26 đọc từ trên xuống                        §16f
── context thứ năm: logistics + Quote vs Actual ────────────────────────────────────────────────────────────
77. internal/domain/shared/allocate.go       chia tiền floor + largest remainder, tổng đúng                     §17b
78. internal/domain/logistics/parcel.go      Parcel expected → received (CÂN THẬT) → batched → shipped           §17c
79. internal/domain/logistics/batch.go       ConsolidationBatch; Ship đóng băng allocations vào event          §17c
80. internal/domain/logistics/allocator.go   FreightAllocator (Domain Service §17) + ByChargeableWeight; LaneRule §17c
81. internal/domain/pricing/reconciliation.go  Quote vs Actual per order; Variance()                            §17d, DDD.md §28
82. internal/app/pricing/{define_lane,reconciler}.go  lane được thông báo; projection ba nguồn                  §17d
83. internal/app/logistics/{parcels,batches}.go  ExpectParcel (listener #3), Receive, Open/Add/Close/Ship        §17e
84. internal/adapter/http/{logistics,lanes}.go  màn hình kho; POST /lanes; GET /reconciliations/{order}          §17f
85. internal/adapter/postgres/migrations/0005_logistics.sql  bảng con của aggregate; reconciliations nullable   §17g
86. internal/platform/wire/wire_test.go      wholeFlow tới delivered + variance −2.50                             §17h
── cửa: ai đang gọi (P9/T1) ────────────────────────────────────────────────────────────────────────────────
87. internal/platform/auth/auth.go            Principal, PORT Verifier, Static, HashToken — vì sao ở platform/ §18a, §18f
88. internal/platform/auth/issue.go           PORT Issuer tách khỏi Verifier; NewToken (crypto/rand)            §18b
89. internal/adapter/http/auth.go             middleware bọc CẢ mux; requireOperator/Customer/Any; sourcingOf   §18c, §18e
90. internal/adapter/http/tokens.go           POST /tokens — token trả về đúng MỘT lần                          §18g
91. internal/adapter/postgres/migrations/0006_auth.sql  api_tokens: lưu hash, revoked_at chứ không DELETE       §18f
92. internal/adapter/postgres/token_repo.go   Verify bằng subtle.ConstantTimeCompare; IsEmpty cho bootstrap     §18g
93. cmd/api/main.go                           -bootstrap-operator-token: chỉ chạy khi bảng rỗng                 §18g
── một chữ của shop đi qua ba context (đợt variant/size) ───────────────────────────────────────────────────
94. internal/domain/catalog/variant.go        khoá chống trùng bỏ MỌI khoảng trắng; variant không tên phải là duy nhất §22a
95. internal/domain/ordering/variant.go       bản sao mỏng nhất có thể: có thật không, của sản phẩm nào            §22c
96. internal/domain/procurement/task.go       Subject — bốn chữ CHÉP vào task lúc mở, không tra lại sau           §22d
97. internal/adapter/postgres/migrations/0008_variant_subject.sql  hai bảng chiếu + 4 cột subject + cột source    §22d
```

Đọc xong 97 file (≈ 12.400 dòng kể cả comment) là đọc hết code hiện có — từ `curl` tới
`variance: -2.50 USD`.

Đọc xong 35 file đầu (≈ 4.500 dòng kể cả comment) là đọc hết ngày 1 — từ `curl` tới dòng
`sent_at` trong Postgres. Thêm 21 file ngày 2 (≈ 2.500 dòng) là đọc hết code hiện có — tới
dòng `total_minor` của một báo giá.

---

## 11. Hành trình một ngày — thứ tự đã xây và bài học mỗi bước

| Bước | Làm gì | Bài học then chốt | Ghi ở |
|---|---|---|---|
| 1 | Review `shared/`: 10 bug (làm tròn xuống, `"150,50"`, tỷ giá 0, tràn int64, zero value…) | mỗi bug một test đỏ trước khi sửa; **nhất quán** quan trọng hơn đúng cục bộ | SETUP.md §9 đợt 1–2 |
| 2 | Chốt 9 quy ước code, viết `ID`, `Events` | quyết định phải nằm trong code/test, không chỉ trong doc | SETUP.md §6 |
| 3 | 7 test canh kiến trúc (`decisions_test.go`) | "quy tắc chỉ nằm trong tài liệu là quy tắc sẽ bị quên" — và guard phải **cố tình phá** để biết nó đỏ được | SETUP.md §6 |
| 4 | `catalog`: Merchant, CategoryPolicy — review chéo lần 1 | gỡ thuế **và** số chia khỏi category: hỏi *"ai sở hữu dữ liệu này?"*; entity phải có **cả hai chiều** (Suspend/Reinstate) | CATALOG.md §6, SETUP.md §9 đợt 3 |
| 5 | Struct tham số `MerchantDetails` | struct chỉ khi (1) validate mọi field, (2) test zero **từng field** — zero cả struct là vô dụng | quy ước 10, đợt 5 |
| 6 | `Product` ⊃ `Variant`, `Provenance` 3 nhóm, `ParcelSpec` hai vai, phương án C | invariant nhiều điều kiện → mỗi điều kiện một lỗi; `now` chỉ đi kèm khi có gì được ghi | CATALOG.md §4, §7; đợt 4 |
| 7 | Review chéo lần 2: reason + event cho `ClearDuplicateFlag`, nhớ cặp đã bác | thứ chặn vòng lặp phải nằm **trong aggregate**; test chưa từng đỏ = chưa chứng minh gì | đợt 5 |
| 8 | Tầng app + adapter in-memory | use case tồn tại để **ghép** (giờ + auth → Provenance) và giữ **luật liên aggregate**; Save trước Pull là test, không phải comment | DDD.md §30 |
| 9 | `adapter/http` + `cmd/api` — chạy thật bằng `curl` (§1) | adapter làm **đúng ba việc**: decode, locale, map lỗi; `"50,00"` là 50 hay 5000 tuỳ `Accept-Language` — đúng chỗ đoán; 400/404/409 là ba cuộc hội thoại khác nhau | SETUP.md §5d, §9 đợt 7 |
| 10 | Snapshot/FromSnapshot + eventcodec (P1) | aggregate mở **cửa sau** cho persistence, không mở field; payload outbox là **hợp đồng** nên viết tay (phương án B), có guard go/ast | §6a, §6c |
| 11 | Postgres adapter + wire + docker compose (P2) | transaction đi theo `ctx`; test rollback từng skip **chạy thật**; test tích hợp song song trên một DB cần **advisory lock** (`pgtest`) | §6b, SETUP.md §5e, §9 đợt 9 |
| 12 | worker + `cmd/worker` (P3) | at-least-once là lựa chọn có chủ đích; subscriber idempotent; `-dsn` bắt buộc | §7a, §7b |
| **Ngày 2** | | | |
| 13 | Dời `ParcelSpec` → `shared`; `CategoryDefined`; `ProductPublished` mang price+parcel | event mang **đủ** trạng thái consumer cần; kiểu hai context dùng → shared kernel thật | §14b |
| 14 | `internal/contracts` + `eventcodec.Decode` | Published Language là package **riêng** — không domain, không adapter; decoder chỉ viết khi có người nghe | §14b, DDD.md §7 |
| 15 | Domain `pricing`: lane/policy/calc/quote (RED → GREEN, 14 test) | `Calculate` là hàm thuần; quyết định duy nhất (cân cái gì) lộ ra `Estimated`; `Accept` trễ vừa đổi vừa từ chối | §14c |
| 16 | `pricingapp`: IssueQuote, AcceptQuote, Projector | cấu hình nghiệp vụ là VO trong `Deps`; projection idempotent + chịu sai thứ tự; commit rồi mới báo từ chối | §14d |
| 17 | Codec mở rộng + guard quét mọi context | guard phải **mở rộng theo hệ**; patch theo anchor cũ sau khi dời kiểu hỏng im lặng → fetch lại trước khi sửa | §14k |
| 18 | Ba use case operator + route; `httpapi.Deps{Catalog, Pricing}`; `/quotes` | flow curl đi tới publish **không chọc tay** vào aggregate; `listing_not_found` có hai nghĩa thật | §14e, §14h |
| 19 | `0002_pricing.sql` + 5 repo; `wire.Graph` + `Subscribe`; relay trong api (memory); smoke thật | ranh giới vẽ trong SQL (không FK chéo); một test vàng chạy cả hệ trên hai graph | §14f, §14g, §14i, §14j |
| 20 | Domain `ordering`: `CustomerOrder`, `Refund`, `Cancel` matrix, snapshot (RED → GREEN) | §26 là một method có bảng test 5 dòng; hằng trạng thái đụng tên event → `Status*`; cọc phải đúng số | §15b |
| 21 | `orderingapp` + `QuoteAcceptedV1` + codec + HTTP `/orders` + `0003` + wire | `mutate` viết khuôn một lần; luật "một quote một đơn" có hai nửa (app + `UNIQUE`); `quote_not_accepted` trước relay | §15c–§15f |
| 22 | Domain `procurement`: `PurchaseTask`, `PurchaseReceipt`, port `MerchantACL`; `shared.ErrOperatorRequired` | biên nhận gắt vì nó đẩy đơn qua điểm không thể quay đầu; sentinel dùng chung lên shared kernel giữ cùng giá trị | §16b |
| 23 | `merchant.Manual` + `procurementapp` (3 nhánh ACL) + `Reactor` bên ordering + 4 contract + codec | ACL là port bằng từ vựng của mình, adapter đầu tiên là một người; fake ACL chứng minh 3 nhánh; Reactor phân biệt "đã xong" | §16c, §16d |
| 24 | `0004` + `/purchase-tasks` + `Subscribe` 5 dòng + test vàng tới `forfeited` + smoke 12 event | một event hai listener; không có endpoint tạo task; bảng định tuyến đọc từ trên xuống = saga | §16e–§16h |
| **Ngày 3** | | | |
| 25 | `shared.Allocate`; domain `logistics` (Parcel, Batch, `FreightAllocator`, `LaneRule`) | chia tiền tổng đúng bằng `big.Int`; luật chia là interface; phần chia đóng băng vào event | §17b, §17c |
| 26 | Pricing: `LaneDefined` + `DefineLane`; `Reconciliation` + `Reconciler` ba nguồn | lane được thông báo như category; §28 thành bảng, `Variance()` | §17d |
| 27 | `logisticsapp`, `Reactor.OnBatchShipped`, `0005`, 8 route, `Subscribe` +6, test vàng tới `delivered`, smoke 20 event | hai aggregate một tx có lý do ghi rõ; từ chối sớm; vòng đời §31 chạy hết bằng code | §17e–§17i |

---

## 12. Symfony ↔ Portage cho riêng flow này

| Bạn quen bên Symfony | Ở đây | Khác gì |
|---|---|---|
| Controller nhận Request, gọi `$bus->dispatch(new AddProduct(...))` | `httpapi.addProduct` gọi thẳng `s.add.Handle(ctx, cmd)` | không có bus tự route; gọi tay, thấy rõ |
| `#[Route]` annotation | `mux.HandleFunc("POST /products", ...)` trong `NewHandler` | bảng route là code, một chỗ |
| Serializer + `ALLOW_EXTRA_ATTRIBUTES=false` | `json.Decoder.DisallowUnknownFields()` | gõ nhầm tên field = 400 |
| `MoneyType` + `NumberFormatter` theo locale | `normalizeAmount(s, language(r))` | adapter đoán dấu thập phân, domain không |
| ExceptionListener → status | `errorTable` + `writeError` | một bảng, `errors.Is` |
| `public/index.php` + `services.yaml` | `cmd/api/main.go` + `wire.Memory/Postgres` | dây nối viết tay, hai graph |
| Doctrine hydrate private field bằng Reflection | `Snapshot()` / `FromSnapshot()` | tường minh, chỉ repository gọi |
| `doctrine:migrations:migrate` | `postgres.Migrate` + `migrations/*.sql` nhúng | 40 dòng, không tool |
| `$em->wrapInTransaction()` (thật) | `postgres.UnitOfWork.InTx` — tx trong `ctx` | không trạng thái toàn cục |
| Serializer normalizer viết tay | `eventcodec.Encode` | payload là hợp đồng |
| `messenger:consume` | `worker.Relay.Run` + `cmd/worker` | đọc outbox, at-least-once |
| DAMADoctrineTestBundle | `pgtest.Pool(t)` + advisory lock | test tích hợp không giẫm nhau |
| `AddProductHandler::__invoke(AddProduct $cmd)` | `AddProductHandler.Handle(ctx, cmd)` | thêm `ctx` mang huỷ/deadline |
| `ClockInterface $clock` autowire | `Deps.Clock` điền tay trong `main()` / test | không container; `mustHave` thay lỗi compile container |
| `$em->wrapInTransaction(fn)` | `UoW.InTx(ctx, fn)` | fn chỉ trả `error`; kết quả qua closure |
| `$repo->find($id)` trả Entity | `Products.ByID(ctx, id)` trả `*Product` | lỗi là giá trị trả về, `errors.Is(err, ErrProductNotFound)` |
| Entity `throw new DomainException` | `return ..., ErrX` — trả nguyên vẹn lên | không try/catch; `if err != nil { return err }` |
| `$em->persist($p); $em->flush()` | `Products.Save(ctx, p)` | tường minh, không unit-of-work ngầm theo dõi thay đổi |
| `postFlush` listener → Messenger | `Outbox.Append(ctx, p.PullEvents())` **sau** Save | thứ tự viết ra, không phải hook |
| `InMemoryRepository` trong `tests/` | `adapter/memory` — cài đặt bình đẳng | dùng được cho dev run |
| Message class đánh version trong package `Contracts` dùng chung | `internal/contracts` `*V1` | primitive-only; đổi nghĩa = `V2` |
| `messenger.yaml routing:` + `#[AsMessageHandler]` | `wire.Subscribe(bus, g)` | một bảng, đọc là thấy ai nghe gì |
| Argument resolver phản chiếu kiểu tham số handler | `on[T]` generic + `eventcodec.Decode` | decode ở đúng một chỗ, type-assert tường minh |
| Handler ghi bảng projection của bundle khác (CQRS) | `pricingapp.Projector` | upsert → idempotent; sai thứ tự → bỏ qua |
| Service tính toán thuần (`PriceCalculator`) | `pricing.Calculate` — Domain Service | không interface khi chỉ có một cách tính |
| Read model / DTO cho `GET` | `httpapi.quoteView` | đọc thẳng repo, chưa có bảng đọc riêng |
| `#[ORM\ManyToOne]` sang entity bundle khác | **không có** — `listings.product uuid` trần | ranh giới context vẽ trong SQL |
| Symfony Workflow component (`workflow.yaml`, `$workflow->apply()`) | `CustomerOrder` — method có tên + `switch o.status` | state machine là code domain, đọc được, test được từng cạnh |
| AbstractHandler / trait "find → method → flush" | `orderingapp.Deps.mutate(ctx, id, fn)` | closure là phần khác duy nhất |
| `UniqueEntity(fields: ["quote"])` | `Orders.ByQuote` ở app **và** `orders.quote UNIQUE` ở DB | hai nửa của một luật |
| `MerchantGatewayInterface` + `ManualGateway` (null object) trong `services.yaml` | `procurement.MerchantACL` + `merchant.Manual{}` cắm ở `wire` | Anti-Corruption Layer: port bằng từ vựng của mình |
| Mock gateway bằng Prophecy/PHPUnit | `apiShop` — struct 10 dòng cài interface | fake tự viết, đọc được, không framework |
| Handler nhận event từ bundle khác rồi gọi service nội bộ | `orderingapp.Reactor` | chỗ duy nhất phân biệt "đã làm rồi" với "không làm được" |
| `moneyphp Money::allocate()` | `shared.Allocate` | 30 dòng, `big.Int`, test tổng đúng |
| Strategy pattern (`FreightAllocatorInterface` + N implementation) | `logistics.FreightAllocator` + `ByChargeableWeight` | Domain Service là interface vì có nhiều luật |
| Bảng báo cáo tổng hợp từ nhiều bundle | `pricing.Reconciliation` + `Reconciler` | projection ba nguồn, chịu mọi thứ tự |

---

## 13. Tự kiểm tra — trả lời không nhìn code

1. `now` được đọc ở đâu, và vì sao không đọc trong `catalog.AddProduct`?
2. Kể hai luật trong `AddProductHandler` không thể nằm trong `Product`. Vì sao?
3. Nếu đảo hai dòng `Save` và `Outbox.Append(PullEvents())`, kịch bản nào hỏng? Test nào bắt?
4. `AddProduct` (command) và `catalog.ProductDetails` khác nhau đúng hai field — hai field đó là gì, thuộc về ai?
5. Vì sao `memory.UnitOfWork.InTx` chỉ gọi `fn(ctx)` mà vẫn cần tồn tại?
6. `memory.ProductRepo.Save` giữ con trỏ. Postgres không làm được vậy — thiếu API gì, và vì sao chưa thêm?
7. Detector trùng nằm ở tầng nào? Quyết định gắn cờ nằm ở tầng nào? Ai nhớ cặp đã bác?
8. `var _ catalog.MerchantRepository = (*MerchantRepo)(nil)` làm gì, và nó thay cho từ khoá nào bên PHP?
9. Handler thứ tư (ví dụ `MeasureProduct`) sẽ có hình gì? Viết khuôn 6 dòng.
10. Test nào chứng minh "thiếu dependency là bug khởi động, không phải lỗi request"?

**Tầng HTTP**

11. Ba việc của adapter là gì? Kể một việc adapter **không được** làm và nó thuộc tầng nào.
12. `"50,00"` với `Accept-Language: vi` ra bao nhiêu? Với `en`? Vì sao domain không tự xử?
13. `ErrPriceCurrency` là 409, `ErrMalformedAmount` là 400. Tiêu chí phân biệt là gì?
14. `cmd/api/main.go` và `world` trong test giống nhau ở đâu, khác nhau ở đâu? Khi có Postgres, file nào đổi?
15. Vì sao không có `GET /products/{id}`?

**Hạ tầng**

16. Vì sao `Product` cần `Snapshot()` thay vì export field? `FromSnapshot` tin gì và từ chối gì?
17. Transaction đi từ `InTx` tới `Outbox.Append` bằng đường nào? Vì sao không truyền `tx` qua tham số?
18. Test rollback từng skip với memory — vì sao skip, và giờ nó chạy ở đâu?
19. Vì sao payload outbox viết tay từng key mà không `json.Marshal(event)`? Test nào bắt event thiếu mapper?
20. At-least-once nghĩa là gì trong `Relay.RunOnce`? Đảo `Publish` và `MarkSent` thì mất gì?
21. Hai package test tích hợp chạy song song trên một DB thì hỏng thế nào, và `pgtest` sửa bằng gì?
22. Vì sao `cmd/worker` không có chế độ in-memory? Và vì sao `cmd/api` in-memory lại **có** relay?

**Context thứ hai (§14)**

23. `pricing` biết về một sản phẩm bằng đường nào? Kể đủ 6 file bytes đi qua từ `Product.Publish` tới `listings`.
24. Vì sao `ProductPublished` phải mang `Price` và `Parcel` thay vì chỉ `ID`?
25. `internal/contracts` không nằm trong domain, không nằm trong adapter. Hai lý do?
26. Projector cần ba tính chất nào? Với mỗi tính chất, chuyện gì hỏng nếu thiếu, và test nào bắt?
27. `Calculate` quyết định đúng một điều. Điều gì, và kết quả lộ ra ở field nào?
28. `Quote.Accept` sau hạn làm gì với trạng thái, và tầng app phải xử lý khác thường thế nào? So với `Relay.RunOnce`.
29. `listing_not_found` có hai nghĩa. Kể cả hai và test nào chứng minh nghĩa thứ hai.
30. Vì sao `listings.product` là `uuid` trần, không `REFERENCES products(id)`?
31. `Policy` và `Classification` nằm trong `Deps` nhưng không phải interface. Chúng là gì, dựng ở đâu, vì sao panic khi zero?
32. Giày 150 USD, hộp 34×23×13, 1250 g, tuyến 5000/500 g, class branded 10 USD/kg, tax 8,81 %, tỷ giá 26 000, sàn 500k. Tính tay total. (Đáp số ở §14i.)

**Context thứ ba (§15)**

33. Ordering biết số tiền của một quote bằng đường nào? Vì sao không đọc bảng `quotes`?
34. Kể 5 trạng thái huỷ được và số hoàn ở mỗi trạng thái. Trạng thái nào là điểm không thể quay đầu, và sự kiện nào của context khác đánh dấu nó?
35. Vì sao `PayDeposit` từ chối cả cọc thiếu **và** cọc thừa?
36. Luật "một quote một đơn" nằm ở đâu? Vì sao cần cả hai chỗ?
37. `Deps.mutate` bỏ đi bao nhiêu dòng lặp? Closure `fn` nhận gì, trả gì? Khi nào **không** nên có helper này?
38. Đặt tên hằng `OrderPurchased` cho trạng thái thì compile lỗi gì? Vì sao PHP không gặp chuyện này?

**Context thứ tư (§16)**

39. `MerchantACL.Purchase` trả bốn kiểu kết quả. Kể cả bốn và use case làm gì với từng cái. Cái nào khiến relay thử lại?
40. Vì sao viết một adapter `Manual` **không mua được**? Nó mua lại được gì cho thiết kế?
41. Procurement biết shop tính tiền gì bằng đường nào? Phải đổi những file nào để `merchant_registered` mang `Currency`, và guard nào bắt nếu quên một chỗ?
42. `Reactor.alreadyDone` nuốt sentinel nào, vì sao, và vì sao aggregate hay relay không tự làm được việc đó?
43. `product_published` có hai listener. Nếu listener thứ hai lỗi, chuyện gì xảy ra với dòng outbox và với listener thứ nhất?
44. Vì sao không có `POST /purchase-tasks`? Luật nào được giữ nhờ endpoint **không tồn tại**?
45. Trong smoke 16h, giữa event #30 và #31 code nào chạy, trong transaction nào?

**Context thứ năm (§17)**

46. Chia 100.00 USD cho 3 phần bằng nhau: float ra gì, `Allocate` ra gì, và cent thừa đi về đâu theo luật nào?
47. Vì sao `FreightAllocator` là interface dù chỉ có một implementation? Kể hai luật chia khác và ai thiệt với mỗi luật.
48. Phần chia cước sống ở đâu sau khi ship? Vì sao ordering và pricing **không** tính lại từ hoá đơn?
49. `AddParcelHandler` sửa hai aggregate trong một transaction. Điều kiện nào cho phép, và câu nào trong comment nói điều đó?
50. `Reconciler` nhận `batch_shipped` trước `order_placed` thì làm gì? Nhận `order_placed` trên quote lạ thì làm gì? Vì sao khác nhau?
51. `purchase_confirmed` có ba listener. Kể cả ba và bảng mỗi cái ghi.
52. Quoted 163.22 + 25.00, actual 163.22 + 27.50: variance là bao nhiêu, dấu nghĩa là gì, và số nào trong quote đã đúng tuyệt đối?

---

## 14. Ngày 2 — context thứ hai: `pricing`, từ event tới báo giá

> Ngày 1 dựng **một** context chạy hết 5 tầng. Ngày 2 trả lời câu hỏi DDD.md §6–§7
> đặt ra mà chưa có code chứng minh: *hai bounded context nói chuyện với nhau bằng
> gì, và ranh giới nằm ở đâu trong code?* Câu trả lời ngắn: **bằng event đi qua
> outbox, dịch qua một Published Language, vào một projection của bên nghe** — và
> ranh giới là guard 7 (`pricing` không import `catalog`) cộng với việc `pricing`
> **không có foreign key** nào sang bảng của `catalog`.
>
> Chọn `pricing` làm context thứ hai vì nó là **insight kinh tế lớn nhất** của nghề
> (DDD.md §28): báo giá phải là *bức ảnh* chụp mọi dữ kiện lúc đó — tỷ giá, bảng
> giá, hộp đã đo hay hộp ước lượng — để sau này đối soát với thực tế.

### 14a. Bản đồ mới — hai context, một outbox, một relay

```
 curl ─▶ httpapi ─▶ catalogapp ─▶ catalog (aggregate) ─▶ outbox ──┐   cùng transaction
                                                                    │
                    ┌───────────────────── worker.Relay ◀───────────┘   Pending → publish → MarkSent
                    ▼
              worker.Bus ── wire.Subscribe ── on[T] ── eventcodec.Decode ─▶ contracts.ProductPublishedV1
                                                                                     │
                                                              pricingapp.Projector ◀─┘   upsert listings / category_profiles
                                                                                     
 curl ─▶ httpapi ─▶ pricingapp.IssueQuote ─▶ pricing.Calculate ─▶ Quote (ảnh chụp) ─▶ quotes + outbox ─▶ relay ─▶ (ordering 🔜)
```

Điểm cần nhìn thấy: **catalog không gọi pricing, pricing không đọc bảng của catalog.**
Thứ duy nhất đi qua ranh giới là một dòng `jsonb` trong `outbox`, và người dịch nó là
`eventcodec` + `contracts`. Đây là **Customer/Supplier qua Published Language** trong
ngôn ngữ của Context Map (DDD.md §7).

**Symfony:** hai bundle, mỗi bundle một schema Doctrine riêng, không `@ORM\ManyToOne`
chéo bundle; nói chuyện qua Messenger với message class đánh version.

### 14b. Đường đi của một event qua ranh giới — đọc 6 file theo thứ tự

**1. `internal/domain/catalog/events.go` — event mang đủ trạng thái.** `ProductPublished`
ngày 1 chỉ có `ID, Merchant, Category, Name`. Ngày 2 mang thêm `Price` và `Parcel`:

```go
type ProductPublished struct {
	ID       ProductID
	Merchant MerchantID
	Category CategoryCode
	Name     string
	Price    shared.Money       // ← mới
	Parcel   shared.ParcelSpec  // ← mới: publish đòi parcel đã đo, nên luôn có
	At       time.Time
}
```

Vì sao? Vì bên nghe **không thể** load aggregate của catalog để hỏi thêm (guard 7). Event
là *tất cả* nó có. Một event chỉ mang `ID` bắt consumer gọi ngược lại — và "gọi ngược" là
coupling trá hình. Thêm `CategoryDefined` (mới) vì pricing cần hộp ước lượng và
restriction của category, và cách duy nhất để nó có là catalog **thông báo**.

`ParcelSpec` vì thế **dời** từ `catalog` sang `shared` (`shared/parcelspec.go`): hai
context cùng dùng một kiểu → đúng nghĩa shared kernel. Kiểm chứng: `catalog` giảm một
file, `pricing` không import `catalog`, guard 7 vẫn xanh.

**2. `internal/adapter/eventcodec/codec.go` — Encode, key viết tay** (§6c, không đổi).
Thêm 3 case cho event của pricing. Guard `TestEncode_contractCoversEveryDomainEvent` giờ
quét `../../domain/*/events.go` — **mọi** context, không chỉ catalog. Thêm context thứ ba
mà quên mapper vẫn đỏ, không cần ai sửa test.

**3. `internal/contracts/catalog_v1.go` — Published Language.** Package **mới, đứng
riêng**, không thuộc domain cũng không thuộc adapter:

```go
type ProductPublishedV1 struct {
	ID       string    `json:"id"`
	Merchant string    `json:"merchant"`
	Category string    `json:"category"`
	Name     string    `json:"name"`
	Price    MoneyV1   `json:"price"`
	Parcel   ParcelV1  `json:"parcel"`
	At       time.Time `json:"at"`
}
```

Vì sao không để trong `catalog`? Vì `pricing` import nó — mà `pricing` không được import
`catalog`. Vì sao không để trong `eventcodec`? Vì tầng app (`pricingapp.Projector`) nhận
struct này làm tham số, và app không nên import adapter. Nó là **ngôn ngữ chung**, nên ở
chỗ chung. Luật của package: chỉ primitive và struct V1 khác; version trong tên; **không
bao giờ** một kiểu domain; đổi nghĩa = kiểu mới (`V2`), không sửa dưới tên cũ.

**4. `internal/adapter/eventcodec/decode.go` — Decode, chỉ khi có người nghe.**

```go
var decoders = map[string]func([]byte) (any, error){
	"catalog.product_published": into[contracts.ProductPublishedV1],
	…
}
func Decode(name string, payload []byte) (any, error)   // ErrNoDecoder nếu chưa ai nghe
```

`into[T]` là **generic** (GO-CHO-PHP.md §11): một hàm sinh một closure cho từng kiểu.
Không có decoder cho `merchant_registered` — chưa ai nghe → *decoder nobody calls is a
contract nobody tests*. Viết khi cần.

**5. `internal/platform/wire/subscribe.go` — bảng định tuyến event.**

```go
func Subscribe(bus *worker.Bus, g Graph) {
	projector := pricingapp.NewProjector(g.Pricing)
	bus.Subscribe("catalog.product_published", on(projector.OnProductPublished))
	bus.Subscribe("catalog.product_measured",  on(projector.OnProductMeasured))
	bus.Subscribe("catalog.product_repriced",  on(projector.OnProductRepriced))
	bus.Subscribe("catalog.product_retired",   on(projector.OnProductRetired))
	bus.Subscribe("catalog.category_defined",  on(projector.OnCategoryDefined))
}

func on[T any](handle func(context.Context, T) error) worker.Handler {
	return func(ctx context.Context, e worker.Entry) error {
		msg, err := eventcodec.Decode(e.Name, e.Payload)   // bytes → V1 struct, ở ĐÚNG MỘT chỗ
		if err != nil { return err }
		m, ok := msg.(T)
		if !ok { return fmt.Errorf("%s decodes to %T, handler wants %T", e.Name, msg, *new(T)) }
		return handle(ctx, m)
	}
}
```

Ai nghe cái gì đọc một chỗ là hết — đây là `messenger.yaml routing:` + mọi
`#[AsMessageHandler]` gom lại. Nó nằm ở `wire` (composition root) vì chỉ chỗ đó được phép
biết **cả** wire format **và** mọi consumer. `cmd/worker` và `cmd/api` (chế độ memory) gọi
cùng hàm này.

**6. `internal/app/pricing/projector.go` — bên nghe.**

```go
func (p *Projector) OnProductPublished(ctx context.Context, m contracts.ProductPublishedV1) error {
	product, err := shared.ParseID(m.ID)          // wire → value object qua constructor VALIDATE
	price, err  := moneyFrom(m.Price)
	parcel, err := parcelFrom(m.Parcel)
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return p.deps.Listings.Save(ctx, pricing.Listing{     // UPSERT → idempotent
			Product: product, Name: m.Name, Category: m.Category, Price: price,
			Parcel: parcel, Measured: true, Active: true,
		})
	})
}
```

Ba tính chất, **mỗi tính chất một test** (`pricing_test.go`):

| Tính chất | Nghĩa | Test chứng minh |
|---|---|---|
| **Idempotent** | nhận cùng message 2 lần → 1 listing | `for range 2 { OnProductPublished }` rồi `Len() == 1` |
| **Chịu được sai thứ tự** | `product_measured` cho sản phẩm chưa nghe → **bỏ qua**, không lỗi | measured trước publish → `ErrListingNotFound` vẫn còn |
| **Từ vựng riêng** | `Class` do `Classification` của pricing gán; catalog không biết "sensitive" là gì | profile `electronics` → `ClassElectronics` |

Idempotent là **bắt buộc**, không phải đẹp: relay là at-least-once (§7a), message *sẽ* tới hai
lần lúc nào đó.

### 14c. Domain `pricing` — đọc theo thứ tự file

**`lane.go`.** Năm kiểu, đọc từ trên xuống:

| Kiểu | Là gì | Điểm học |
|---|---|---|
| `LaneCode` | khoá tự nhiên `"us_forwarder"` | struct, không `type string` — chỉ `ParseLaneCode` tạo được (giống `CategoryCode`) |
| `GoodsClass` | thường / hàng hiệu / điện tử / nhạy cảm | **từ vựng của nhà gửi hàng**, không phải của catalog — cùng đôi giày, catalog gọi `footwear`, forwarder gọi `branded` |
| `RateCard` | giá/kg cho 4 class | **4 field thay map**: so sánh được bằng `==`, không thể thiếu class |
| `DutyPolicy` | thuế nhập khẩu hiện ra sao trên tuyến | `DutyBundled()` (zero value, tuyến đang dùng) hoặc `DutyItemised(rate)`. Đây chính là field catalog **đã từ chối sở hữu** (CATALOG.md §6): cùng đôi giày, hai tuyến, hai câu trả lời |
| `ShippingLane` | toàn bộ bảng giá của một tuyến | VO khoá tự nhiên, bất biến; `ChargeableWeight`, `Freight`, `Surcharge`, `Duty` là *hành vi* của bảng giá |
| `Classification` | category của mình → class của họ | không biết → `standard` (rẻ nhất): thiếu dòng hiện ra là báo giá **thấp** lúc đối soát, không phải khách bị tính thừa |

```go
func (l ShippingLane) Freight(class GoodsClass, w shared.Weight) shared.Money {
	return l.rates.PerKg(class).Mul(shared.RatePPM(w.Grams() * 1000))   // 2500 g = 2 500 000 ppm của 1 kg — không nhân float
}
```

**`policy.go`.** `MarginPolicy.Fee = max(percent × subtotal, floor)` — "lãi 500k/đơn" là
sàn. `QuotePolicy{SalesTax, Margin, Deposit, TTL}` là bộ hằng số nghiệp vụ, `NewQuotePolicy`
từ chối deposit ngoài (0, 100%], TTL ≤ 0. Cả hai đều được **chụp vào từng quote**.

**`listing.go`.** `Listing` và `CategoryProfile` là **projection** — field exported, không
constructor. Vì sao phá lệ "không export field"? Vì đây là *bản ghi pricing giữ*, không phải
aggregate có invariant; `Calculate` validate cái nó cần khi cần. Và `Product shared.ID`, không
`catalog.ProductID`: pricing không được gọi tên kiểu của catalog — id là **danh tính mờ** đi
qua dây.

**`calc.go` — Domain Service đầu tiên (DDD.md §17).** Hàm thuần: không clock, không repo,
không trạng thái.

```
chargeable = lane.ChargeableWeight(parcel)          parcel = đã đo, không thì hộp mặc định của category
subtotal   = item + item×tax + freight(class, chargeable) + surcharge + duty       (tiền tuyến, USD)
subtotalₕ  = subtotal × fx                                                          (tiền nhà, VND)
total      = subtotalₕ + max(subtotalₕ × margin%, floor)
deposit    = total × deposit%
```

Quyết định duy nhất trong hàm là *"cân cái gì?"*: hộp thật nếu đã lên cân, không thì hộp
ước lượng — và `Breakdown.Estimated` **nói rõ** đang đoán. Đây là §28 thành code: quote
dựng trên ước lượng là quote biết mình đang đoán.

**`quote.go` — aggregate root thứ ba.** `Quote` giữ `Breakdown` (bức ảnh) + `issuedAt`,
`expiresAt`, `status`. Hai hành vi, không setter:

```go
func (q *Quote) Accept(now time.Time) error {
	if q.status != QuoteIssued { return …ErrQuoteNotIssued }
	if q.IsExpired(now) {                        // TRỄ: đổi trạng thái + ghi event RỒI từ chối
		q.status = QuoteExpired
		q.Record(QuoteExpiredEvent{…})
		return …ErrQuoteExpired
	}
	q.status = QuoteAccepted
	q.Record(QuoteAcceptedEvent{…})
	return nil
}
```

Accept trễ **vừa** đổi trạng thái **vừa** trả lỗi — hình dạng hiếm, và nó buộc tầng app phải
xử lý khác thường (14d). `Expire(now)` là bản quét cho quote không ai trả lời.
`QuoteSnapshot`/`QuoteFromSnapshot` cùng khuôn §6a; `FromSnapshot` từ chối breakdown thiếu
hoặc `expires ≤ issued`.

**Số vàng** (`quote_test.go`, giày 150 USD, hộp 34×23×13 cm, tỷ giá 26.000, class standard 9 USD/kg):

| | Đã đo 1250 g | Chưa đo → category 1200 g |
|---|---|---|
| thể tích 10 166 cm³ / 5000 | 2034 g | (33×22×13) 9 438 cm³ → 1888 g |
| chargeable (bước 500 g) | **2500 g** | **2000 g** |
| freight 9 USD/kg | 22.50 | 18.00 |
| tax 8,81 % | 13.22 | 13.22 |
| subtotal USD | 185.72 | 181.22 |
| × 26 000 | 4 828 720 | 4 711 720 |
| fee max(10 %, 500k) | 500 000 | 500 000 |
| **total VND** | **5 328 720** | **5 211 720** |
| deposit 50 % | 2 664 360 | 2 605 860 |

Wire seed xếp `footwear → branded` (10 USD/kg) nên số chạy thật ở 14i là **5 393 720**
(freight 25.00). Chênh 65 000 ₫ đúng bằng 2,5 kg × 1 USD × 26 000 — tự kiểm tra được.

### 14d. Use case — `internal/app/pricing/`

`Deps` có hai thứ *không phải service*: `Policy` và `Classification` — **cấu hình nghiệp
vụ**, là value object dựng một lần ở `wire`, không phải interface. Không có `IsZero` check
lúc dựng handler thì quote sẽ tính với deposit 0 % — nên `NewIssueQuoteHandler` panic khi
`Policy.IsZero()`.

`IssueQuoteHandler.Handle`: parse lane code → trong `InTx`: `Listings.ByProduct` →
`Lanes.ByCode` → `Profiles.ByCode` (**tuỳ chọn**: `ErrProfileNotFound` được nuốt, vì sản
phẩm đã đo không cần hộp ước lượng) → `Rates.Current(lane.Currency(), home)` với `home` lấy
từ `Policy.Margin().Floor().Currency()` → `pricing.IssueQuote` → `Save` → `Outbox.Append`.
Khuôn §8, thêm bước "gom đầu vào".

`AcceptQuoteHandler.Handle` — **đọc hai lần**:

```go
var refused error
err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
	q, err := h.deps.Quotes.ByID(ctx, id)
	if err := q.Accept(now); err != nil {
		if !errors.Is(err, pricing.ErrQuoteExpired) {
			return err          // không đổi gì → rollback, báo
		}
		refused = err           // ĐÃ đổi (issued → expired): phải LƯU, rồi mới báo
	}
	Save; Outbox.Append(PullEvents)
	return nil
})
if err != nil { return err }
return refused
```

Cùng ý với `Relay.RunOnce` (§7a): *commit phần đã xảy ra, rồi báo phần bị từ chối*. Lỗi trả
về không có nghĩa là "không có gì thay đổi" — phải đọc aggregate mới biết.

### 14e. HTTP — `quotes.go`

| Route | Trả | Ghi chú |
|---|---|---|
| `POST /quotes {"product_id","lane"}` | 201 `{"id"}` | ghi trả id, số nằm ở GET |
| `GET /quotes/{id}` | 200 `quoteView` | **read model đầu tiên** của repo |
| `POST /quotes/{id}/accept` | 204 | trễ → 409 `quote_expired` và quote **đã** expired |

`quoteView` là CQRS **ở dạng đơn giản nhất** (DDD.md §24): chưa có bảng đọc riêng, adapter
đọc `Quotes.ByID` rồi bẻ `Breakdown` ra từng dòng. Tiền trong view là `{"amount":"150.00",
"currency":"USD"}` — **text**, không float: client hiển thị, không cộng.

Mã lỗi mới trong `errorTable`:

| Status | Code | Khi nào |
|---|---|---|
| 400 | `invalid_lane_code` | `"US Forwarder!"` |
| 404 | `lane_not_found`, `quote_not_found`, `listing_not_found` | |
| 409 | `no_exchange_rate` | chưa ai set tỷ giá hôm nay |
| 409 | `listing_inactive`, `nothing_to_weigh` | sản phẩm đã retire; chưa đo và category không có ước lượng |
| 409 | `quote_expired`, `quote_not_issued`, `quote_still_valid` | |

`listing_not_found` có **hai nghĩa thật**: sản phẩm chưa publish, *hoặc* đã publish mà relay
chưa chạy. **Eventual consistency có mã trạng thái** (DDD.md §25) — test vàng 14j bắn
`POST /quotes` ba lần để thấy nó chuyển từ 404 sang 201 đúng lúc relay chạy.

### 14f. Postgres — `migrations/0002_pricing.sql` + `pricing_repos.go`

Năm bảng: `lanes`, `fx_rates`, `listings`, `category_profiles`, `quotes`. Ba điều đáng nhìn:

1. **Không một `REFERENCES` nào sang bảng của catalog.** `listings.product` là `uuid` trần.
   Ranh giới bounded context **vẽ được trong SQL** — join chéo bằng FK là cách phá ranh giới
   nhanh nhất.
2. `quotes` lưu breakdown **từng cột** (`item_minor`, `tax_minor`, `freight_minor`, … `fx_rate text`)
   thay một cột `jsonb`: đối soát Quote vs Actual sau này là câu `SELECT`, không phải parse JSON.
3. `fx_rate`/`fx_rates.rate` là **text** `"26000.5"` — `ExchangeRate.Rate()` là dạng chuẩn hoá,
   `NewExchangeRate` đọc lại y nguyên; `DeepEqual` round trip pass.

`laneFrom` dựng lại `ShippingLane` **qua `NewShippingLane`**, không gán field: một dòng bị sửa
tay thành vô nghĩa (divisor 0) bị từ chối lúc load, không lúc tính tiền. Test
(`pricing_repos_test.go`, 5 test): round trip từng repo bằng `reflect.DeepEqual` — kể cả struct
toàn field unexported như `ShippingLane`; lane có **cả** surcharge **và** duty itemised để cột
nullable và cột flag đều được thử.

### 14g. `wire` — từ `Deps` thành `Graph`

```go
type Graph struct {
	Catalog catalogapp.Deps
	Pricing pricingapp.Deps
	Source  worker.Source   // cùng outbox mà hai Deps ghi vào, nhìn từ phía relay
}
```

Hai context **chia sẻ** `Clock`, `UoW`, `Outbox` — cùng con trỏ, nên `Projector.InTx` lồng
trong `Relay.RunOnce.InTx` **join** transaction đang mở: projection và `sent_at` commit cùng
nhau. `quotePolicy()`/`goodsClasses()` là cấu hình nghiệp vụ SETUP.md §7 dưới dạng code;
`seedPricing` ghi lane + tỷ giá thẳng repo (chưa ai nghe lane → không có event để thông báo).

`cmd/api` chế độ memory chạy **relay trong process** (goroutine, 200 ms) với đúng
`wire.Subscribe` — không thì `POST /quotes` in-memory mãi 404. Chế độ Postgres **không**: đó
là việc của `cmd/worker`, chạy cả hai thì mỗi event tới hai lần (vô hại, vô ích). Quyết định
đợt 8 ("không stub worker") vẫn đứng: đây không phải binary giả, là relay thật bên trong api.

### 14h. Ba bước operator để publish qua HTTP — `catalogapp` + `operator.go`

Ngày 1 `TestPOSTPublish` phải **chọc tay** vào aggregate để có variant/xác nhận/cân. Giờ có
đường đi thật:

| Use case | Route | Điều kiện |
|---|---|---|
| `AddVariant` | `POST /products/{id}/variants` → 201 | trùng size+màu → 409 `duplicate_variant` |
| `ConfirmListing` | `POST /products/{id}/confirm-listing` → 204 | **operator-only** (`requireOperator`, §18c): danh tính lấy từ token, khách gọi → 403. Lỗi `operator_required` của `NewProvenance` vẫn còn trong domain nhưng **không kích hoạt được từ HTTP nữa** (§18e) |
| `MeasureProduct` | `POST /products/{id}/measure` `{weight_g, length_mm, …}` → 204 | cần operator; phát `product_measured` |

Cùng khuôn §8. `operatorOf(r)` đọc header, trả zero id nếu thiếu — **chấp nhận hay không là
việc của use case**, adapter không quyết.

### 14i. Kết quả — chạy thật (`scripts/smoke.ps1`, 05/09 15:28)

`api.exe -dsn` + `worker.exe -dsn -every 300ms`, 8 request:

```
merchant 01a070ae-e253-… / product 01a070ae-e2e5-…
published; waiting for the worker to relay product_published...
quote 01a070ae-e73e-…: issued branded chargeable 2500 g, total 5393720 VND, deposit 2696860

--- worker log ---
event #11 catalog.merchant_registered  {"id": "01a070ae-e253-…", "name": "Example Sports", …}
event #12 catalog.product_added        {"id": "01a070ae-e2e5-…", "category": "footwear", …}
event #13 catalog.product_measured     {"id": "01a070ae-e2e5-…", "parcel": {"weight_g": 1250, "length_mm": 340, …}, "verified": true}
event #14 catalog.product_published    {"id": "01a070ae-e2e5-…", "price": {"minor": 15000, "currency": "USD"}, "parcel": {…}, …}
event #15 pricing.quote_issued         {"id": "01a070ae-e73e-…", "lane": "us_forwarder", "total": {"minor": 5393720, "currency": "VND"},
                                        "deposit": {"minor": 2696860, …}, "estimated": false, "expires_at": "2026-09-07T08:28:25Z"}
event #16 pricing.quote_accepted       {"id": "01a070ae-e73e-…", "total": {…5393720…}, …}
```

Đọc theo tầng: #14 là catalog nói; giữa #14 và #15, worker gọi `Projector.OnProductPublished`
(không có log riêng — projection là bảng `listings`); #15 là pricing nói, **sau khi** nghe.
`expires_at` = `at` + 48 h đúng TTL của policy.

### 14j. Test vàng — `wire_test.go` `wholeFlow`, chạy trên **cả hai** graph

Một test, hai lần (`TestMemory_runsTheWholeFlow`, `TestPostgres_runsTheWholeFlow`), đúng thứ
tự curl ở trên, và thêm ba khẳng định chỉ test mới nói được:

1. `POST /quotes` **trước publish** → `listing_not_found`.
2. `POST /quotes` **sau publish, trước relay** → vẫn `listing_not_found`. *(Eventual consistency.)*
3. `relay.RunOnce` → ≥ 5 dòng; chạy lại → **0** dòng (không gửi lại); quote → 201, total
   `5393720`; accept → 204; `RunOnce` → đúng **2** (`quote_issued` + `quote_accepted`).

Đây là test thay cho `TestMemory_/TestPostgres_servesTheSeededCatalog` ngày 1: cùng lời hứa
Ports & Adapters, phạm vi giờ là **cả hệ**.

### 14k. Bài học ngày 2

| Bài học | Ở đâu |
|---|---|
| Event **mang đủ trạng thái** consumer cần; "chỉ gửi id" là coupling trá hình | `catalog/events.go` |
| Published Language là **package riêng** — không domain (guard 7), không adapter (app import nó) | `internal/contracts` |
| Kiểu hai context cùng dùng → shared kernel **thật**; dời `ParcelSpec` xong guard vẫn xanh là bằng chứng | `shared/parcelspec.go` |
| Guard phải **mở rộng theo hệ**: quét `domain/*/events.go`, không hard-code một file | `codec_test.go` |
| Projection: idempotent + chịu sai thứ tự + từ vựng riêng — **mỗi cái một test** | `pricingapp/pricing_test.go` |
| Hàm thuần là chỗ đặt phép tính nghiệp vụ; **quyết định duy nhất** trong hàm (cân cái gì) phải lộ ra kết quả (`Estimated`) | `pricing/calc.go` |
| Có aggregate **vừa đổi trạng thái vừa từ chối** → tầng app commit rồi mới báo | `accept_quote.go` |
| Ranh giới context vẽ trong SQL: **không FK chéo**; breakdown lưu từng cột cho đối soát | `0002_pricing.sql` |
| Cấu hình nghiệp vụ là value object dựng ở composition root, không phải env var rải rác | `wire.quotePolicy()` |
| Đổi tên/dời kiểu xong thì **mọi patch theo anchor cũ hỏng im lặng** — luôn fetch lại file trước khi sửa | đợt 10, SETUP.md §9 |
| Getter `Duty()` đụng method `Duty(item)`: Go không overload — đặt tên theo *vai* (`DutyPolicy()`) | `lane.go` |
| Test cũng có bug: struct ẩn danh trong test không có `json:"sales_tax"` → decode rỗng → "tax = """ | `quotes_test.go` |
| Coverage `memory` tụt 65 % → 38 %: repo pricing in-memory được **package khác** dùng nhưng không có unit test riêng — coverage đo theo package | `memory/pricing_repos.go` |

---

## 15. Context thứ ba: `ordering` — luật cọc 50 % thành code

> Pricing trả lời *"bao nhiêu tiền?"*. Ordering trả lời *"khách đã cam kết gì, đã trả gì,
> và nếu huỷ thì lấy lại gì?"*. Đây là context mà **mô hình kinh doanh** nằm trong code:
> khách cọc 50 %, mình ứng 100 % đi mua, và có một **điểm không thể quay đầu** (DDD.md
> §26) — trước nó huỷ là hoàn đủ, sau nó huỷ là mất cọc. Toàn bộ §26 giờ là một method
> có test: `CustomerOrder.Cancel`.

### 15a. Bản đồ — cạnh thứ hai của Context Map

```
 POST /quotes/{id}/accept ─▶ pricing.QuoteAccepted ─▶ outbox ─▶ relay ─▶ Subscribe ─▶ orderingapp.Projector ─▶ accepted_quotes
                                                                                                                       │
 POST /orders {quote_id, variant_id} + Bearer khách ─▶ orderingapp.PlaceOrder ── đọc AcceptedQuote ◀─────────────┘
                                                            │
                                                            ▼
                                                   ordering.CustomerOrder  awaiting_deposit ─▶ deposited ─▶ purchased ─▶ in_transit ─▶ delivered
                                                            │                                   └─▶ purchase_failed          (bất kỳ) ─▶ cancelled + Refund
                                                            ▼
                                                   ordering.deposit_paid ─▶ outbox ─▶ (procurement 🔜 P7: đi mua)
```

Cùng khuôn với cạnh catalog → pricing (§14a): ordering **không đọc bảng `quotes`**, nó
được **báo** một lần qua `pricing.quote_accepted` và giữ bản sao `AcceptedQuote` (chỉ 4
field: quote, product, total, deposit). Số tiền của đơn là số của quote — ordering **không
bao giờ tính lại giá**.

### 15b. Domain — `internal/domain/ordering/order.go`

**State machine viết tay**, mỗi chuyển trạng thái là một method có tên nghiệp vụ:

| Method | Từ → tới | Ai gọi | Từ chối bằng |
|---|---|---|---|
| `PlaceOrder(details, now)` | ∅ → `awaiting_deposit` | khách | `ErrInvalidOrder` (từng field), `ErrCurrencyMismatch` |
| `PayDeposit(amount, now)` | `awaiting_deposit` → `deposited` | thanh toán | `ErrWrongAmount` — phải **đúng** số cọc, không hơn không kém |
| `ConfirmPurchase(now)` | `deposited` → `purchased` | procurement (event) | `ErrNotDeposited` |
| `FailPurchase(reason, now)` | `deposited` → `purchase_failed` | procurement (event) | `ErrNotDeposited`, `ErrEmptyReason` |
| `Ship(now)` | `purchased` → `in_transit` | logistics (event) | `ErrNotPurchased` |
| `PayBalance(amount, now)` | `in_transit` (cờ `balancePaid`) | thanh toán | `ErrNotInTransit`, `ErrWrongAmount` |
| `Deliver(now)` | `in_transit` → `delivered` | operator | `ErrNotInTransit`, `ErrBalanceUnpaid` |
| `Cancel(reason, now)` | 5 trạng thái đầu → `cancelled` | khách / mình | `ErrAlreadyDelivered`, `ErrOrderCancelled`, `ErrEmptyReason` |

Vì sao cọc phải **đúng số**? Cọc 40 % không phải "cam kết nhỏ hơn" — nó là *không cam kết*
(mình vẫn phải ứng 100 %); cọc thừa là một khoản hoàn không ai xin. Luật này là của
aggregate, không phải của form.

**`Cancel` — §26 thành code, kèm bảng test:**

```go
switch o.status {
case StatusAwaitingDeposit:                 refund = Refund{amount: Zero}                    // chưa trả gì
case StatusDeposited, StatusPurchaseFailed: refund = Refund{amount: o.deposit}               // chưa mua → hoàn ĐỦ
case StatusPurchased, StatusInTransit:      refund = Refund{amount: Zero, forfeited: true}   // đã mua → MẤT cọc
case StatusDelivered:                       return ErrAlreadyDelivered
default:                                    return ErrOrderCancelled
}
```

`TestOrder_cancelRefundDependsOnThePointOfNoReturn` chạy 5 kịch bản (bảng `cases`) —
mỗi dòng là một trạng thái, một số tiền hoàn, một cờ `forfeited`. `purchase_failed` hoàn
đủ vì đó là **lỗi của mình** (hết size), không phải của khách.

`Refund` là value object hai field (`amount`, `forfeited`) — `Refund{}` nghĩa là "chưa huỷ".
`refuse(expected)` chọn đúng sentinel: gọi `PayDeposit` trên đơn đã huỷ trả
`ErrOrderCancelled`, không phải `ErrNotAwaitingDeposit` — người gọi biết **lý do thật**.

**Snapshot** từ chối 8 hình dạng không thể có (`TestOrder_snapshotRoundTripAndRejectsImpossibleShapes`):
refund trên đơn đang sống, huỷ mà không có lý do, trả đủ tiền trước khi hàng đi, giao mà
chưa thu đủ…

### 15c. Use case — `internal/app/ordering/`

**`Deps.mutate`** — khuôn §8 (*load → hỏi aggregate → Save → Pull → Outbox*) viết **một lần**:

```go
func (d Deps) mutate(ctx context.Context, id ordering.OrderID, fn func(o *ordering.CustomerOrder) error) error {
	return d.UoW.InTx(ctx, func(ctx context.Context) error {
		o, err := d.Orders.ByID(ctx, id)
		if err := fn(o); err != nil { return err }      // ← dòng DUY NHẤT khác nhau giữa các use case
		Save; Outbox.Append(o.PullEvents())
	})
}
```

Bảy handler (`PayDeposit`, `PayBalance`, `Cancel`, `ConfirmPurchase`, `FailPurchase`, `Ship`,
`Deliver`) mỗi cái còn ~6 dòng: lấy `now`, gọi `mutate` với một closure. Catalog chưa làm
vậy vì mới có 2 handler cùng hình — đến handler thứ bảy thì lặp đã thành bug chờ sẵn.

**`PlaceOrderHandler`** giữ hai luật **liên aggregate** (đúng việc của tầng app):

1. quote phải **đã được chấp nhận** — `Quotes.ByID` trên projection; không có →
   `ErrQuoteNotAccepted` (chưa accept, *hoặc* accept rồi mà relay chưa chạy — lại là §25);
2. **một quote, một đơn** — `Orders.ByQuote` có rồi → `ErrQuoteAlreadyUsed`. Double-click
   không phải hai cam kết. Postgres giữ nửa thứ hai của luật này: `orders.quote UNIQUE`.

`CancelOrderHandler.Handle` **trả `Refund`** cùng với lỗi — người gọi (sau này: tầng thanh
toán) biết phải chuyển bao nhiêu mà không đọc lại aggregate.

### 15d. HTTP — `orders.go`

| Route | Trả | Mã lỗi đáng nhớ |
|---|---|---|
| `POST /orders {quote_id, variant_id}` (khách, §18) | 201 `{"id"}` | 403 nếu operator gọi; 409 `quote_not_accepted`, 409 `quote_already_used` |
| `GET /orders/{id}` | 200 `orderView` (total, deposit, balance, `refund` chỉ khi đã huỷ) | 404 `order_not_found` — **kể cả khi đơn có thật nhưng của khách khác** (§18d) |
| `POST /orders/{id}/deposit {amount, currency}` | 204 | 409 `wrong_amount`, 409 `not_awaiting_deposit` |
| `POST /orders/{id}/balance {amount, currency}` | 204 | 409 `not_in_transit`, `wrong_amount` |
| `POST /orders/{id}/cancel {reason}` | 204 — số hoàn xem ở GET | 400 `empty_reason`, 409 `already_delivered`, `order_cancelled` |
| `POST /orders/{id}/deliver` | 204 | 409 `balance_unpaid` |

Số tiền thanh toán đi qua `normalizeAmount` như mọi số tiền ở biên: `"2.696.860"` với
`Accept-Language: vi` là 2 696 860 ₫ (test `TestOrders_placePayCancel`). Từ 06/09 **không còn
`customer_id` trong body**: chủ đơn là chủ token (§18e), nên cũng không còn field nào để đặt hộ tên người khác.

### 15e. Postgres — `0003_ordering.sql`

Hai bảng: `accepted_quotes` (projection) và `orders`. `orders.quote UNIQUE` là **luật
nghiệp vụ được DB bảo vệ** — nếu hai request đặt cùng quote lọt qua `ByQuote` cùng lúc, cái
thứ hai chết ở INSERT thay vì thành đơn đôi. `refund_minor` nullable: `NULL` = chưa huỷ (khớp
`Refund{}`); `forfeited` là cột riêng vì "hoàn 0 ₫ vì chưa trả" và "hoàn 0 ₫ vì mất cọc" là
hai sự thật khác nhau. Không FK sang `quotes` hay `products` — ranh giới (§14f).

### 15f. Wire + test vàng

`Graph.Ordering` dùng chung `Clock`/`UoW`/`Outbox`; `Subscribe` thêm **một dòng**:
`pricing.quote_accepted → orders.OnQuoteAccepted`. `wholeFlow` kéo dài thêm 8 khẳng định:
`POST /orders` **trước relay** → `quote_not_accepted`; sau relay → 201; đặt lần hai →
`quote_already_used`; cọc thiếu → `wrong_amount`; cọc đủ → `deposited`; huỷ → `cancelled`,
`refund = 2 696 860`; `RunOnce` → đúng 3 (`order_placed`, `deposit_paid`, `order_cancelled`).
Chạy trên cả memory và Postgres.

### 15g. Bài học

| Bài học | Ở đâu |
|---|---|
| Go có **một namespace cho cả package**: hằng `OrderPurchased` (trạng thái) đụng kiểu `OrderPurchased` (event) → compile lỗi. Đặt `Status*` cho trạng thái, để tên event thì quá khứ | `order.go`, `events.go` |
| Khuôn lặp đến lần thứ 7 → viết `mutate` một lần; closure `fn` là phần khác duy nhất | `orderingapp/deps.go` |
| Luật liên aggregate có **hai nửa**: tầng app kiểm (`ByQuote`) và DB canh (`UNIQUE`) — race điều kiện chỉ nửa đầu không đủ | `place_order.go`, `0003_ordering.sql` |
| Sentinel phải nói **lý do thật**: đơn đã huỷ trả `ErrOrderCancelled` dù bạn gọi gì | `refuse()` |
| Cùng một số 0 ₫ có hai nghĩa → hai cột/hai field (`amount`, `forfeited`) | `Refund` |
| Thiếu `Money.IsPositive` → thêm vào shared kernel **kèm test**, không viết `!IsZero && !IsNegative` rải rác | `shared/money.go` |

---

## 16. Context thứ tư: `procurement` — đi mua, và Anti-Corruption Layer đầu tiên

> Ba context trước đều xử lý **dữ liệu của mình**. Procurement là context đầu tiên phải
> chạm vào **thế giới bên ngoài**: web của shop, giỏ hàng của họ, số đơn của họ. Đây là
> chỗ DDD.md §22 (ACL) và §26 (saga) gặp nhau trong code: `PurchaseTask` mở ra khi khách
> cọc, hỏi shop qua một **port** (`MerchantACL`), đóng lại bằng biên nhận thật hoặc lý do
> thất bại — và mỗi kết cục là một event **ordering** nghe để đi tiếp hoặc bù trừ.

### 16a. Bản đồ — vòng saga khép kín

```
 ordering.deposit_paid ──▶ relay ──▶ procurementapp.OpenTaskHandler
                                          │  Items.ByProduct → Shops.ByID  (projection của catalog: shop nào, tiền gì)
                                          │  procurement.OpenTask(...)
                                          │  ACL.Purchase(task) ──▶ merchant.Manual → ErrManualPurchase → task ở OPEN
                                          ▼                         (API shop sau này: ok → Confirm ngay · !ok → Fail ngay)
                                   purchase_tasks (buyer's to-do)  ◀── GET /purchase-tasks
                                          │
      POST /purchase-tasks/{id}/confirm {reference, paid, currency} + Bearer operator   ── hoặc ──  POST …/fail {reason}
                                          │                                                              │
                              procurement.purchase_confirmed                               procurement.purchase_failed
                                          │  relay → orderingapp.Reactor                                 │
                                          ▼                                                              ▼
                        ordering: deposited → PURCHASED (điểm không thể quay đầu)         ordering: deposited → purchase_failed
                                                                                            (huỷ → hoàn đủ: lỗi của mình)
```

Bốn context, **một bảng định tuyến** (`wire.Subscribe`, 16f) — đọc từ trên xuống chính là
saga §26.

### 16b. Domain — `internal/domain/procurement/task.go`

| Kiểu | Là gì | Điểm học |
|---|---|---|
| `PurchaseTask` | aggregate root: `open → confirmed \| failed` | trạng thái cuối là **cuối** — task hỏng không thử lại trên cùng task; đơn quyết định (huỷ hoàn đủ, hay đơn mới) |
| `TaskDetails{Order, Product, Variant, Currency}` | đầu vào `OpenTask` | procurement được **báo** mua gì; `Currency` là của shop — biên nhận phải cùng tiền |
| `PurchaseReceipt{Reference, Paid, PaidBy}` | shop trả về gì | **Actual đầu tiên** của Quote vs Actual (§28): trả thật 163.22 USD trong khi quote tính 150 + tax 13.22 — lệch là chuyện có thật, ghi lại, không ghi đè |
| `Shop{Merchant, Name, Site, Currency}`, `Item{Product, Merchant, Name}` | projection từ catalog | procurement không hỏi catalog "shop này tính tiền gì" — nó **nghe** `merchant_registered` (giờ mang `Currency`) và `product_published` |

`Confirm(receipt, now)` từ chối bốn thứ, mỗi thứ một sentinel: không có reference
(`ErrEmptyReference`), tiền ≤ 0 (`ErrInvalidTask`), sai tiền tệ (`ErrPaidCurrency`), không
biết ai mua (`shared.ErrOperatorRequired`). Vì sao gắt vậy? Vì `purchase_confirmed` là thứ
**đẩy đơn qua điểm không thể quay đầu** — biên nhận giả là khách mất cọc oan.

`ErrOperatorRequired` **dời lên shared kernel**: catalog (Provenance) và procurement (biên nhận)
cùng hỏi "ai làm", cùng từ chối một kiểu. `catalog.ErrOperatorRequired = shared.ErrOperatorRequired`
— cùng *giá trị*, nên `errors.Is` và bảng lỗi HTTP không đổi.

### 16c. Port `MerchantACL` và adapter `merchant.Manual` — ACL thật đầu tiên

```go
// internal/domain/procurement/repository.go — PORT, bằng từ vựng CỦA MÌNH
type MerchantACL interface {
	Purchase(ctx context.Context, task *PurchaseTask) (receipt PurchaseReceipt, ok bool, reason string, err error)
}

// internal/adapter/merchant/manual.go — adapter đầu tiên: một con người
func (Manual) Purchase(ctx context.Context, task *procurement.PurchaseTask) (procurement.PurchaseReceipt, bool, string, error) {
	return procurement.PurchaseReceipt{}, false, "", procurement.ErrManualPurchase
}
```

Vì sao viết một adapter **không mua được**? Vì nhờ nó, use case có port để gọi **ngay bây
giờ**, và luồng *"mở task → hỏi shop → rơi về người"* là code thật có test thật. Ngày shop
có API: thêm một file trong `adapter/merchant/`, đổi một dòng trong `wire` — use case
không thêm nhánh `if`. Ba câu trả lời của port, ba nhánh trong `OpenTaskHandler`:

| ACL trả | Handler làm | Test |
|---|---|---|
| `ErrManualPurchase` | task ở `open` cho người mua | `TestOpenTask_manualShopLeavesTheTaskOpen` |
| `ok = true` + receipt | `Confirm` ngay → `purchase_task_opened` + `purchase_confirmed` cùng transaction | `TestOpenTask_anAPIShopClosesTheTaskAtOnce/bought` |
| `ok = false` + reason | `Fail` ngay | `…/sold out` |
| lỗi khác (shop sập) | **trả lỗi** → không lưu gì, relay giữ dòng lại thử sau | `…` phần cuối: 0 task, 0 event |

`apiShop` trong test là **fake ACL** ~10 dòng — đủ để chứng minh use case đúng với cả ba
câu trả lời trước khi bất kỳ shop nào có API. Đó là Ports & Adapters trả lãi.

**Symfony:** `MerchantGatewayInterface` với `ManualGateway` là implementation "null object"
bạn đăng ký trong `services.yaml` trước khi có API thật.

### 16d. Use case — `internal/app/procurement/`

`OpenTaskHandler.Handle` (16a) có bốn bước đáng đọc:

1. **Idempotent theo đơn**: `Tasks.ByOrder` có rồi → trả id cũ, không làm gì. `deposit_paid`
   tới hai lần (at-least-once) không mở hai task. Postgres canh nửa sau: `purchase_tasks."order" UNIQUE`.
2. **Tra projection**: `Items.ByProduct` → `Shops.ByID` → tiền tệ của shop. Thiếu (relay chưa
   tới `product_published`) → **trả lỗi để relay thử lại** — không đoán USD.
3. `procurement.OpenTask` → `ACL.Purchase` → ba nhánh (16c).
4. Save + outbox — **một transaction** cho cả mở lẫn đóng-ngay.

`OnDepositPaid(ctx, contracts.DepositPaidV1)` là điểm vào cho `wire.Subscribe`: parse 3 id →
`Handle`. `ConfirmTaskHandler`/`FailTaskHandler` là `mutate` (§15c) với một closure.

**`orderingapp.Reactor`** — chiều ngược: `OnPurchaseConfirmed` → `ConfirmPurchaseHandler`,
`OnPurchaseFailed` → `FailPurchaseHandler`. Một chi tiết đáng đọc hai lần: `alreadyDone`
nuốt `ErrNotDeposited`. Vì sao? Event tới lần hai, đơn **đã** `purchased`, aggregate từ chối
đúng — nhưng với relay đó không phải lỗi cần thử lại, mà là *"đã xong rồi"*. Reactor là **chỗ
duy nhất** biết phân biệt "không làm được" với "đã làm rồi"; aggregate không biết, relay không biết.

### 16e. HTTP — `purchase_tasks.go`

| Route | Trả | Ghi chú |
|---|---|---|
| `GET /purchase-tasks` | 200 `[taskView]` | **danh sách việc** của người mua: task `open`, cũ nhất trước |
| `GET /purchase-tasks/{id}` | 200 `taskView` | `paid`, `reference`, `closed_at` chỉ khi đã đóng |
| `POST /purchase-tasks/{id}/confirm {reference, paid, currency}` | 204 | `requireOperator` → khách gọi 403; sai tiền → 409 `paid_currency` |
| `POST /purchase-tasks/{id}/fail {reason}` | 204 | 400 `empty_reason`; đóng rồi → 409 `purchase_task_not_open` |

Không có `POST /purchase-tasks`: task **không tạo bằng tay** — chỉ `deposit_paid` qua relay
mới mở được. Endpoint không tồn tại là cách rẻ nhất để giữ luật "không mua khi chưa có cọc".

### 16f. Wire — bảng định tuyến giờ là saga

```go
func Subscribe(bus *worker.Bus, g Graph) {
	projector := pricingapp.NewProjector(g.Pricing)
	bus.Subscribe("catalog.product_published", on(projector.OnProductPublished))   // pricing nghe
	…
	orders := orderingapp.NewProjector(g.Ordering)
	bus.Subscribe("pricing.quote_accepted", on(orders.OnQuoteAccepted))            // ordering nghe pricing

	shops := procurementapp.NewProjector(g.Procurement)
	bus.Subscribe("catalog.merchant_registered", on(shops.OnMerchantRegistered))    // procurement nghe catalog
	bus.Subscribe("catalog.product_published", on(shops.OnProductPublished))       // ← listener THỨ HAI trên cùng event
	buyer := procurementapp.NewOpenTaskHandler(g.Procurement)
	bus.Subscribe("ordering.deposit_paid", on(buyer.OnDepositPaid))                // procurement nghe ordering

	reactor := orderingapp.NewReactor(g.Ordering)
	bus.Subscribe("procurement.purchase_confirmed", on(reactor.OnPurchaseConfirmed)) // ordering nghe procurement
	bus.Subscribe("procurement.purchase_failed", on(reactor.OnPurchaseFailed))
}
```

`product_published` có **hai** listener (pricing dựng listing, procurement dựng item) — `Bus`
gọi theo thứ tự, một cái hỏng thì dòng outbox giữ lại cho **cả hai** (at-least-once, cả hai
idempotent nên vô hại). `ACL: merchant.Manual{}` cắm ở đúng một chỗ.

### 16g. Test vàng kéo tới điểm không thể quay đầu

`wholeFlow` (`wire_test.go`) giờ đi: … → cọc → `GET /purchase-tasks` **rỗng** (chưa relay) →
`RunOnce` = 2 → **một** task, `currency: USD` (từ projection, không phải VND của cọc) →
confirm với biên nhận `163.22 USD` → đơn **vẫn** `deposited` (chưa relay) → `RunOnce` = 2
(`task_opened` + `purchase_confirmed`) → đơn `purchased` → huỷ → `cancelled`, `refund 0`,
`forfeited: true` → `RunOnce` = 2. Trên cả memory và Postgres. Khoảng không nhất quán
(§25) xuất hiện **ba lần** trong một test, mỗi lần có mã trạng thái hoặc trạng thái đọc được.

### 16h. Kết quả — chạy thật (`scripts/smoke.ps1`, 05/09 16:13)

```
quote 01a070d7-e09e-…: issued branded chargeable 2500 g, total 5393720 VND, deposit 2696860
order 01a070d7-e4e6-…: deposited; buyer's list: 1 open task(s), first for order 01a070d7-e4e6-… in USD
order after purchase_confirmed: purchased; cancel now → cancelled, refund 0 VND, forfeited=True

--- worker log ---
event #23 catalog.merchant_registered      {"currency": "USD", "site": "www.example-….com", …}
event #24 catalog.product_added
event #25 catalog.product_measured
event #26 catalog.product_published        ← pricing dựng listing, procurement dựng item (2 listener)
event #27 pricing.quote_issued
event #28 pricing.quote_accepted           ← ordering dựng accepted_quote
event #29 ordering.order_placed
event #30 ordering.deposit_paid            ← procurement mở task, hỏi Manual → ở open
event #31 procurement.purchase_task_opened
event #32 procurement.purchase_confirmed   {"by": "0192a1b2-…", "paid": {"minor": 16322, "currency": "USD"}, "reference": "NK-20260905-001", …}
event #33 ordering.order_purchased         ← Reactor: điểm không thể quay đầu
event #34 ordering.order_cancelled         {"reason": "changed mind after purchase", "refund": {"minor": 0, …}, "forfeited": true}
```

Mười hai event, bốn context, một outbox, một worker. Giữa #30 và #31 là `OpenTaskHandler`
(cùng transaction với `sent_at` của #30); giữa #32 và #33 là `Reactor`.

### 16i. Bài học

| Bài học | Ở đâu |
|---|---|
| ACL là **port bằng từ vựng của mình**; adapter đầu tiên có thể là một người — miễn là use case gọi port ngay từ đầu | `procurement.MerchantACL`, `merchant.Manual` |
| Fake ACL ~10 dòng trong test chứng minh cả ba câu trả lời của port trước khi có API thật | `procurement_test.go` `apiShop` |
| Event cần thêm dữ liệu cho consumer mới (`MerchantRegistered.Currency`) → đổi struct, codec, contract, bảng test cùng lúc — guard bắt được nếu quên | `catalog/events.go`, `codec_test.go` |
| Sentinel hai context cùng dùng → shared kernel, giữ **cùng giá trị** để `errors.Is` cũ vẫn đúng | `shared.ErrOperatorRequired` |
| "Đã làm rồi" ≠ "không làm được": chỉ **Reactor** phân biệt được, và nó nuốt đúng một sentinel | `orderingapp/reactor.go` |
| Một event, nhiều listener — Bus gọi theo thứ tự; idempotent làm cho at-least-once vô hại với cả hai | `wire/subscribe.go` |
| Không có endpoint tạo task = luật "không mua khi chưa cọc" rẻ nhất | `http/purchase_tasks.go` |
| **Bug của tôi:** patch theo anchor bị lệch → script Python dừng giữa chừng nhưng `wput` vẫn đẩy file cũ → lỗi compile "undefined". Script vá phải in `SKIP` từng anchor thay vì dừng, và **kiểm tra kết quả trước khi đẩy** | đợt 12, SETUP.md §9 |

---

## 17. Context thứ năm: `logistics` — cân thật, gom lô, chia cước, và đối soát Quote vs Actual

> Context cuối cùng của vòng đời đơn (DDD.md §31). Ba việc: **cân thật** (Actual của
> Quote vs Actual), **gom lô** (một thùng, nhiều khách), **chia cước** — hoá đơn một
> thùng thành N dòng cho N đơn, đúng đến từng cent. Và nhờ nó, pricing khép được §28:
> bảng `reconciliations` nói mỗi đơn **lời hay lỗ bao nhiêu**.

### 17a. Bản đồ — 5 context, 20 event của một đơn, từ dán link tới giao tận nhà

```
 procurement.purchase_confirmed ─▶ logisticsapp.ExpectParcel ─▶ Parcel(expected)   ← listener #3 của cùng event
 POST /parcels/{id}/receive {g, mm} + operator ─▶ Parcel.Receive ─▶ parcel_received  ← CÂN THẬT
 POST /batches {lane} ─▶ ConsolidationBatch(open)   (lane phải có LaneRule — projection của pricing.lane_defined)
 POST /batches/{id}/parcels ─▶ batch.AddParcel + parcel.AssignToBatch     (hai aggregate, một transaction)
 POST /batches/{id}/close  ─▶ closed
 POST /batches/{id}/ship {freight} ─▶ batch.Ship(freight, ByChargeableWeight{divisor, step}) ─▶ batch_shipped{allocations[]}
                                              │                                  │
                          ordering.Reactor.OnBatchShipped ─▶ in_transit    pricing.Reconciler.OnBatchShipped ─▶ ActualFreight
 POST /orders/{id}/balance → POST /orders/{id}/deliver ─▶ delivered
 GET /reconciliations/{order} ─▶ quoted 163.22 + 25.00 · actual 163.22 + 27.50 · variance −2.50 USD
```

### 17b. `shared.Allocate` — chia tiền đúng đến từng cent

```go
Allocate(100.00 USD, [1, 1, 1]) → 33.34, 33.33, 33.33      // tổng ĐÚNG 100.00
```

Chia rồi làm tròn từng phần → tổng lệch. `Allocate` làm **floor** từng phần bằng `big.Int`,
rồi trả từng cent thừa cho phần bị mất nhiều nhất khi floor (largest remainder). Đây là
`moneyphp->allocate()` viết ra 30 dòng để thấy vì sao float không dùng được cho việc này.
Test: 5 vector kể cả số âm (khoản tín dụng chia cùng cách) và trọng số 0.

### 17c. Domain `logistics`

| Kiểu | Là gì | Điểm học |
|---|---|---|
| `Parcel` | aggregate: `expected → received → batched → shipped` | `Receive(actual, by, now)` đòi operator và đủ 3 cạnh — số này là **Actual** của §28 |
| `ConsolidationBatch` | aggregate: `open → closed → shipped`; `items []BatchItem` | một parcel/đơn một lần; `Ship(freight, alloc)` **đóng băng** `allocations` vào event — tính lại bằng luật khác sau này là viết lại lịch sử |
| `FreightAllocator` | **Domain Service §17**, là interface | chia theo cân tính phí / thể tích / giá trị — mỗi luật thiệt một nhóm khách → xứng đáng có chỗ riêng |
| `ByChargeableWeight{Divisor, Step}` | implementation duy nhất hôm nay | `ChargeableWeight` của shared kernel cho từng parcel → `shared.Allocate` |
| `LaneRule{Code, Divisor, Step}` | projection của `pricing.lane_defined` | logistics chỉ cần *cách đếm cân*, không cần giá — pricing giữ bảng giá |

Số vàng (`logistics_test.go`): hoá đơn 87.50 USD, giày 2500 g + áo khoác 5000 g chargeable →
**29.17 / 58.33** — cent thừa về phần có phần dư lớn hơn, tổng đúng 87.50.

### 17d. Pricing lớn thêm: `LaneDefined`, `DefineLane`, `Reconciliation`

- **`ShippingLane.Defined(at)`** + `pricingapp.DefineLaneHandler`: lane được **thông báo** như
  category (§14b) — seed đi qua handler, `POST /lanes` cũng vậy. Event mang divisor + step +
  currency, **không mang giá**.
- **`pricing.Reconciliation`** — Quote vs Actual cho MỘT đơn, tiền tuyến (USD):
  `QuotedGoods` (item + tax + duty), `QuotedFreight` (freight + surcharge), `ActualGoods` (biên
  nhận procurement), `ActualFreight` (phần chia từ batch). `Complete()` khi đủ hai actual;
  `Variance() = quoted − actual`: dương = báo giá dư (lãi giữ được), âm = ước tính thấp (mình chịu).
- **`pricingapp.Reconciler`** nghe **ba** context: `order_placed` (đơn nào → quote nào; copy
  phần quoted từ **quote của chính pricing**), `purchase_confirmed` (actual goods),
  `batch_shipped` (actual freight). Chịu sai thứ tự: hoá đơn tới trước cả khi biết đơn → giữ
  dòng dở, `order_placed` điền sau. Quote lạ → **lỗi** (retry + báo), không lặng lẽ tạo dòng.

### 17e. Use case — `logisticsapp`

`ExpectParcelHandler.OnPurchaseConfirmed` idempotent theo đơn (`Parcels.ByOrder`).
`OpenBatchHandler` từ chối lane không có `LaneRule` **ngay lúc mở** — không để tới lúc ship mới
biết không chia được. `AddParcelHandler` và `ShipBatchHandler` đụng **hai aggregate trong một
transaction** (batch + parcel). Đây là nới lỏng "một aggregate một transaction" **có chủ đích**:
trạng thái của parcel là sổ sách suy ra từ batch, không phải invariant bên nào phụ thuộc; một
vòng event để đồng bộ chỉ mua thêm độ trễ. Ghi rõ trong comment để người sau không tưởng là quên.

`orderingapp.Reactor.OnBatchShipped` lặp qua allocations → `ShipOrderHandler`; nuốt
`ErrNotPurchased` **và** `ErrOrderCancelled`: đơn huỷ sau khi mua (mất cọc) vẫn có thể nằm
trong thùng — hàng của mình, bán lại — không phải lỗi.

### 17f. HTTP

| Route | Trả | Ghi chú |
|---|---|---|
| `GET /parcels` | 200 `[parcelView]` | màn hình kho: chưa ship, cũ nhất trước |
| `POST /parcels/{id}/receive {weight_g, …}` | 204 | `requireOperator`; 409 `parcel_not_expected` nếu cân lần hai |
| `POST /batches {lane}` | 201 `{"id"}` | 404 `lane_rule_not_found` nếu pricing chưa định nghĩa lane |
| `POST /batches/{id}/parcels {parcel_id}` | 204 | 409 `duplicate_parcel`, `parcel_not_received` |
| `POST /batches/{id}/close` | 204 | 409 `batch_empty` |
| `POST /batches/{id}/ship {freight, currency}` | **200 `[allocationView]`** | ngoại lệ "ghi trả id": người đóng gói cần thấy phần chia ngay; phần chia cũng đã đóng băng trong event |
| `POST /lanes {…}` | 204 | operator định nghĩa/định nghĩa lại lane; quote đã phát giữ số cũ |
| `GET /reconciliations/{order}` | 200 | `complete`, `quoted`, `actual`, `variance` (chỉ khi complete) |

### 17g. Postgres — `0005_logistics.sql`

`batch_items` và `batch_allocations` là **bảng con của aggregate** (như `product_variants`):
ghi lại cùng cha (delete + insert), không bao giờ sửa riêng. `reconciliations` toàn cột
nullable — dòng **lớn dần** theo event tới, đúng như `Reconciler` cần. `"order"` là từ khoá
SQL → phải quote, ở cả ba bảng.

### 17h. Test vàng — tới `delivered` và `variance`

`wholeFlow` giờ chạy hết vòng đời §31: … `purchased` → `GET /parcels` thấy 1 parcel `expected`
(listener thứ 3 của `purchase_confirmed`) → receive → batch → ship 27.50 → `RunOnce` = **6**
(`order_purchased`, `parcel_expected`, `parcel_received`, `batch_opened`, `batch_closed`,
`batch_shipped`) → đơn `in_transit` → balance → deliver → `delivered` →
`GET /reconciliations/{order}` → `complete: true`, `variance: -2.50 USD` → `RunOnce` = 3.
Trên cả memory và Postgres. Kịch bản "huỷ sau khi mua → mất cọc" chuyển về test đơn vị
(`order_test.go`, `orders_test.go`) vì một đơn không thể vừa giao vừa huỷ.

### 17i. Kết quả — chạy thật (`scripts/smoke.ps1`, 06/09 00:25)

```
quote …: issued branded chargeable 2500 g, total 5393720 VND, deposit 2696860
order …: deposited; buyer's list: 1 open task(s) … in USD
order after purchase_confirmed: purchased; warehouse expects 1 parcel(s), ref NK-20260905-001
batch … shipped: order … billed on 2500 g → freight 27.50 USD
order at the end: delivered; quote vs actual: quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD

--- worker log: 20 event, #42 → #61 ---
#42 pricing.lane_defined            ← seed đi qua DefineLane; logistics dựng lane_rules
#43–#46 catalog.*                   #47 quote_issued  #48 quote_accepted
#49 order_placed  #50 deposit_paid  #51 purchase_task_opened  #52 purchase_confirmed
#53 order_purchased  #54 parcel_expected      ← cùng #52, hai listener + reconciler
#55 parcel_received  #56 batch_opened  #57 batch_closed  #58 batch_shipped
#59 order_shipped  #60 balance_paid  #61 order_delivered
```

Một đơn, năm context, hai mươi event, một worker. Mỗi mũi tên trong sơ đồ §31 giờ là một
dòng trong log này.

> **Sau đợt variant/size (08/09):** cùng flow đó có thêm một dòng
> `catalog.variant_added` ngay sau `product_added`, nên khối `catalog.*` là
> **#43–#47** và dòng cuối là **#62**. Một đơn = **20 event** (không tính
> `lane_defined` của seed). `web/flow.html` đi đúng 20 event đó, và
> `scripts/smoke.ps1` in ra đúng chừng ấy — §22.

### 17j. Bài học

| Bài học | Ở đâu |
|---|---|
| Chia tiền = floor + largest remainder bằng `big.Int`, **tổng đúng** là bất biến có test | `shared/allocate.go` |
| Luật chia cước là **interface** vì có nhiều luật, mỗi luật thiệt một người khác nhau | `logistics.FreightAllocator` |
| Phần chia **đóng băng vào event** — consumer dùng số đó, không tính lại | `BatchShippedEvent.Allocations` |
| Hai aggregate một transaction: được, khi trạng thái bên kia là sổ sách và **ghi lý do** | `logisticsapp/batches.go` |
| Từ chối sớm: batch mở trên lane không có rule bị chặn lúc **mở**, không lúc **ship** | `OpenBatchHandler` |
| Projection ba nguồn phải **chịu mọi thứ tự** và giữ dòng dở; nguồn của chính mình sai → lỗi to | `pricingapp/reconciler.go` |
| Đặt tên event `*Event` khi trùng hằng trạng thái (`BatchClosed`) — lần thứ hai gặp, lần này chủ động | `logistics/events.go` |
| **Bug của tôi:** đếm sai số event relay (5 thay vì 6) trong test vàng — test đỏ, code đúng; đếm lại từ log | `wire_test.go` |
| **Bug của tôi:** script vá dừng giữa chừng → file cũ được đẩy → "undefined". Lần thứ ba. Quy tắc mới: **luôn `print` từng anchor, không `assert`** | đợt 13 |

---

## 18. Ai đang gọi — auth là adapter, danh tính là port

Năm context, ba mươi route, và cho tới hôm nay **bất kỳ ai** cũng gọi được tất
cả. Danh tính operator đi trong header `X-Operator-ID` mà **chính người gọi tự
điền**, còn `customer_id` là một field trong body. P9/T1 xoá cả hai.

### 18a. Vì sao auth KHÔNG nằm trong domain

Phản xạ đầu tiên là tạo `internal/domain/identity/`. Sai — và lý do đáng nhớ:

> **"Ai đang gọi" không phải quy tắc nghiệp vụ.** Nó là thứ **biên** xác lập
> *trước khi* bất kỳ use case nào chạy. `Product.Publish` cần một
> `shared.OperatorID`; nó không cần biết id đó đến từ token, từ session, hay từ
> một người gõ tay vào màn hình quản trị.

Nên `auth` sống ở `internal/platform/` — cùng chỗ với `clock`. Cả hai trả lời
một câu hỏi mà domain **nhận kết quả** chứ không **đi hỏi**:

```
   clock  →  "bây giờ là mấy giờ"   →  domain nhận `now time.Time`
   auth   →  "ai đang gọi"          →  domain nhận `shared.OperatorID`
```

Guard 6 (*domain chỉ import stdlib + allowlist*) vẫn xanh sau cả T1 — vì
`internal/domain/` **không import `auth`** một lần nào.

### 18b. Một port, hai nửa, hai adapter

```
                    platform/auth
        ┌───────────────────────────────────┐
        │  Verifier   "token này là ai?"    │  ← 30 route dùng
        │  Issuer     "cắt cho tôi chìa"    │  ← 1 route dùng
        └───────────────┬───────────────────┘
                        │
         ┌──────────────┴───────────────┐
         ▼                              ▼
   auth.Static                   postgres.TokenRepo
   map trong RAM                 bảng api_tokens
   dev + mọi test HTTP           production
```

Tách `Verifier` khỏi `Issuer` không phải cho đẹp: **một route phát chìa, ba
mươi route chỉ đọc.** Ba mươi cái đó không được có khả năng phát. Interface nhỏ
là cách duy nhất nói điều đó với compiler.

### 18c. Cửa đóng mặc định

```go
// server.go — authenticate bọc CẢ mux, không bọc từng route
return authenticate(d.Auth)(mux)
```

Hai cách làm, khác biệt không nhỏ:

| | Rủi ro |
|---|---|
| Bọc **từng route** | quên một dòng = một endpoint mở toang, im lặng |
| Bọc **cả mux** | route thêm ngày mai **đã** ở sau cửa; muốn mở phải viết ngoại lệ, ở một chỗ nhìn thấy |

`TestAuth_noTokenIsAlways401` là thứ giữ lời hứa đó.

Trên mỗi route, quyền ghi **cùng dòng với đường dẫn** — bảng route và bảng
quyền không thể lệch nhau, vì chúng là **một**:

```go
mux.HandleFunc("POST /categories", requireOperator(s.defineCategory))
mux.HandleFunc("POST /products",   requireAny(s.addProduct))
mux.HandleFunc("POST /orders",     requireCustomer(s.placeOrder))
```

### 18d. 401 và 403 là hai câu chuyện khác nhau

```
   401 unauthenticated   tôi không biết anh là ai
   403 forbidden         tôi biết anh là ai; cái này không phải của anh
```

Gộp làm một nghe "an toàn hơn" nhưng thực ra **tệ hơn**: người tấn công không
phân biệt được token sai với token đúng-nhưng-thiếu-quyền, nên họ đoán mãi.
Tách ra thì 403 nói *"token của anh thật"* — điều chỉ có ích cho người **đã có
token thật**.

Nhưng có một chỗ **phải** nói dối, và nói dối có chủ đích:

```go
// Khách hỏi đơn của người khác → 404, KHÔNG phải 403
{ordering.ErrNotOwner, http.StatusNotFound, "order_not_found"},
```

403 ở đây sẽ **xác nhận đơn đó tồn tại**. Với một khách tò mò gõ id ngẫu nhiên,
đó là rò rỉ. 404 nói đúng một điều: *"không có gì cho anh ở đây"*.

### 18e. Danh tính là token, không phải header — và nó đổi cả provenance

Đây là chỗ T1 chạm vào **nghiệp vụ**, không chỉ hạ tầng.

`Product.Publish` đòi `listingProv.Verified()` — nghĩa là **một con người đã
xác nhận món này là gì**. Trước T1, "con người" đó là giá trị trong header
`X-Operator-ID`, mà **chính người gọi viết ra**. Nói cách khác: ai cũng tự xưng
là operator được, và invariant quan trọng nhất của catalog dựa trên lời tự xưng.

```go
// TRƯỚC: người gọi tự khai
operator := r.Header.Get("X-Operator-ID")
sourcedBy := req.SourcedBy          // "operator" hay "customer", tuỳ họ

// SAU: cả hai suy từ token
operator, _ := operatorOf(r)        // rỗng nếu là khách — đúng
sourcedBy := sourcingOf(r)          // operator → verified; customer → không
```

`TestAuth_provenanceComesFromTheTokenNotAHeader` gửi kèm `X-Operator-ID` của
một người lạ và khẳng định nó **không ảnh hưởng gì**.

**Hệ quả thú vị:** lỗi `operator_required` (400) giờ **không kích hoạt được từ
HTTP nữa** — sau `requireOperator` thì luôn có operator. Ba assertion cũ test
lỗi đó đã đổi thành "khách gọi → 403". Không phải mất test; là một **trạng thái
sai đã trở thành không thể biểu diễn**.

### 18f. Lưu hash, và vì sao KHÔNG salt

```go
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
```

Phản xạ từ PHP là gọi `password_hash()` (bcrypt/argon2). **Sai chỗ.** Bcrypt cố
tình chậm và tự sinh salt ngẫu nhiên — đúng cho **mật khẩu người đặt**, vì
người ta chọn mật khẩu đoán được.

Token ở đây là **256 bit ngẫu nhiên của mình**:

- không có gì để brute-force — không gian tìm kiếm là 2²⁵⁶
- không rainbow table nào chứa nó
- và salt sẽ **phá chính thao tác cần làm**: tra token theo hash

Đổi lại, `Verify` so sánh bằng `subtle.ConstantTimeCompare` — để **thời gian**
trả lời không tiết lộ token đúng được bao nhiêu ký tự đầu.

### 18g. Chìa đầu tiên: `-bootstrap-operator-token`

`wire.Postgres` **không seed token nào**. Vậy operator đầu tiên ở đâu ra?

```bash
go run ./cmd/api -dsn postgres://… -bootstrap-operator-token "$(openssl rand -hex 32)"
```

Và cờ đó **chỉ chạy khi `api_tokens` rỗng**. Guard đó là toàn bộ ý nghĩa:

> Nhét credential cố định vào DB **đã có người dùng** là cửa sau — ai đọc script
> deploy, lịch sử shell, hay chính `wire.go` cũng gọi API với tư cách nhân viên
> được. Trên DB rỗng thì **chưa có ai để giả mạo**, nên cùng hành động đó chỉ là
> cắt chiếc chìa đầu tiên.

Chìa thứ hai trở đi qua `POST /tokens` (operator-only) — trả token **một lần
duy nhất**, vì kho chỉ giữ hash. Mất thì cắt lại, không lấy lại được.

### 18h. Bảng quyền

| Route | Ai |
|---|---|
| `POST /tokens`, `/categories`, `/merchants`, `/lanes` | operator |
| `POST /products` | operator **hoặc** khách (provenance khác nhau) |
| `POST /products/{id}/{variants,confirm-listing,measure,publish}` | operator |
| `POST /quotes`, `GET /quotes/{id}` | cả hai |
| `POST /quotes/{id}/accept`, `POST /orders` | **khách** |
| `GET /orders/{id}`, `POST /orders/{id}/cancel` | chủ đơn **hoặc** operator |
| `POST /orders/{id}/{deposit,balance,deliver}` | operator (tới khi có cổng thanh toán) |
| `/purchase-tasks/**`, `/parcels/**`, `/batches/**`, `/reconciliations/**` | operator |

### 18i. Smoke thật, sau khi thay toàn bộ cơ chế xác thực

```
merchant 01a074a4-… / product 01a074a4-…
quote 01a074a4-…: issued branded chargeable 2500 g, total 5393720 VND, deposit 2696860
order 01a074a4-…: deposited; buyer's list: 1 open task(s)
order after purchase_confirmed: purchased; warehouse expects 1 parcel(s), ref NK-20260905-001
batch 01a074a4-… shipped: order 01a074a4-… billed on 2500 g → freight 27.50 USD
order at the end: delivered; quote vs actual: quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD
```

20 event `#69`–`#88` qua 5 context. **Số vàng không đổi một chữ** so với §17i —
đúng điều một thay đổi ở tầng adapter phải đảm bảo. (Transcript này chạy **trước**
đợt variant/size; giờ cùng flow ra 21 dòng vì có thêm `catalog.variant_added` — §22.)

### 18j. Bài học

| Bài học | Chỗ nó hiện ra |
|---|---|
| **"Ai đang gọi" là việc của biên, không phải của domain** — cùng lý do với `Clock` | `platform/auth`; guard 6 vẫn xanh |
| **Cửa đóng mặc định**: bọc cả mux, không bọc từng route | `authenticate(d.Auth)(mux)` |
| **401 và 403 là hai câu hỏi khác nhau**; nhưng chủ đơn sai → **404**, vì 403 xác nhận đơn tồn tại | `errorTable` |
| **Đừng tin thứ người gọi tự khai** — header và field trong body đều là lời tự xưng | `sourcingOf`, `customerOf` |
| Khi cửa bảo đảm điều kiện, **trạng thái sai trở thành không biểu diễn được** — `operator_required` chết tự nhiên | 3 assertion đổi thành 403 |
| **Hash token, đừng bcrypt nó**: 256 bit ngẫu nhiên khác mật khẩu người đặt | `HashToken` |
| **Credential cố định trong DB thật = cửa sau**; bootstrap chỉ khi bảng rỗng | `bootstrapOperator` |
| Interface nhỏ nói lên quyền: `Verifier` cho 30 route, `Issuer` cho 1 | `auth/issue.go` |
| **Bug của tôi:** thêm bảng nhưng `pgtest` truncate theo danh sách viết tay → test đọc rác của test trước. Sửa: hỏi `pg_tables`, đừng đọc danh sách | `truncateAll` |
| **Bug của tôi:** script vá báo "PATCH 2 chỗ" mà không đổi gì — đúng cảnh báo đợt 13. Từ nay **đếm lại sau khi chạy rồi mới báo** | — |

---

## 19. Việc mà KHÔNG AI gọi — quét quote hết hạn

Mười tám mục trước, mọi use case đều có người gọi: một request HTTP, hoặc một
event do context khác phát. Mục này là use case đầu tiên **không có ai cả**.

### 19a. Vì sao quote phải hết hạn

`pricing.Quote` là **ảnh chụp** (§14c, DDD.md §28). Nó đóng băng ba thứ đều là
sự thật *của một khoảnh khắc*:

```
   giá trên web shop   →  shop đổi giá lúc nào cũng được
   tỷ giá 26 000       →  ngân hàng đổi mỗi ngày
   cân nặng ước tính   →  đo lại thì khác
```

Cho khách chấp nhận một ảnh chụp hai tuần tuổi nghĩa là **bán bằng cái giá mình
không còn có**. `QuotePolicy.TTL` = 48 h là câu trả lời, và `Quote.Expire` là
chỗ nó thành code.

Nhưng ai gọi `Expire`? Trước T3 chỉ có **một** chỗ: `Accept` muộn tự lật quote
sang `expired` rồi từ chối (§14d). Nghĩa là một quote không ai đụng tới thì
**mãi mãi ở trạng thái `issued`** — đúng luật khi đọc, sai khi đếm: báo cáo
"đang có bao nhiêu báo giá chờ khách" sẽ cộng cả những cái chết từ tháng trước.

### 19b. Cái gì gây ra việc này? — Thời gian

Đây là điểm học của cả mục:

| | Cái gì đẩy | Ví dụ |
|---|---|---|
| Request | một con người | `POST /orders` |
| Event | một context khác | `deposit_paid` → mở task mua |
| **Sweep** | **thời gian trôi** | quote quá 48 h |

Không có event `quote_got_old` nào để nghe, vì **không có gì xảy ra cả** — chỉ
là không có gì xảy ra. Không ai phát được một event như thế, nên phải có người
**đi hỏi**.

Và đó chính là lý do `Clock` là một **port** chứ không phải `time.Now()` rải
khắp nơi: câu hỏi "cái này quá hạn chưa" phải kiểm chứng được mà **không cần
đợi 48 giờ**. Trong test, 48 giờ trôi qua trong một dòng:

```go
clk.Advance(49 * time.Hour)
n, err := sweep.Handle(ctx)   // n == 1
```

### 19c. `worker.Sweeper` — anh em của `Relay`

```
   Relay    đọc bảng outbox  →  publish  →  MarkSent      ← có việc thì mới chạy
   Sweeper  chờ tick         →  Pass(ctx) (int, error)    ← đến giờ thì chạy
```

Hai khác biệt nhỏ, đều có lý do:

- **`Sweeper` chờ một nhịp rồi mới quét lần đầu**, `Relay` thì chạy ngay. Relay
  có sẵn hàng đợi lúc khởi động; sweep chỉ có cái đồng hồ, mà quét bảng ngay
  lúc boot thì một process restart liên tục sẽ quét đi quét lại vô ích.
- **Pass lỗi = một dòng log, không phải chết process.** Binary này sống để relay
  outbox; sweep không với tới DB một giây không được kéo theo cả cái đó.

```go
// cmd/worker/main.go
if *sweep > 0 {
	expire := pricingapp.NewExpireQuotesHandler(g.Pricing, *sweepBatch)
	go func() { _ = worker.NewSweeper("quote-expiry", expire.Handle, *sweep, nil).Run(ctx) }()
}
```

Chạy **cạnh** relay chứ không tách binary riêng: vài query một phút, thêm một
process là thêm một thứ phải deploy mà chẳng được gì. `cmd/api` chế độ memory
cũng bật sweep — vì ở đó không có `cmd/worker`, mà một bản dev nơi quote sống
mãi thì khác production **đúng ở cái luật mà quote sinh ra để giữ**.

### 19d. Mỗi quote một transaction — và vì sao

```go
for _, q := range due {
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error { … })   // MỘT quote, MỘT tx
}
```

Gói cả lô vào một transaction thì code ngắn hơn và hành vi **tệ hơn**: một dòng
hỏng, hoặc một quote vừa bị khách accept xen vào, sẽ **cuốn ngược toàn bộ** việc
đã làm cho những quote khác.

Lý do sâu hơn nằm ở DDD.md §14: **ranh giới transaction = ranh giới aggregate**.
Lô 100 quote này không phải một invariant — không có luật nghiệp vụ nào nói
chúng phải cùng hết hạn — nên nó **không được** là một đơn vị nguyên tử.

### 19e. Đọc hai lần: danh sách rồi mới đọc lại trong transaction

```go
due, _ := h.deps.Quotes.IssuedBefore(ctx, now, h.limit)   // NGOÀI transaction — chỉ là danh sách việc
for _, q := range due {
	_ = h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		fresh, err := h.deps.Quotes.ByID(ctx, q.ID())        // TRONG transaction — đọc lại
		…
	})
}
```

Giữa hai dòng đó, khách có thể vừa bấm accept. Bản trong danh sách khi ấy là
**cũ**, và lưu nó xuống sẽ **xoá mất cái accept của khách**. Đọc lại trong
transaction là cách rẻ nhất để không bao giờ ghi đè bằng dữ liệu cũ.

Và khi domain từ chối vì đã có người nhanh tay hơn:

```go
if errors.Is(err, pricing.ErrQuoteNotIssued) || errors.Is(err, pricing.ErrQuoteStillValid) {
	return errAlreadySettled     // rollback, nhưng KHÔNG tính là lỗi của lượt quét
}
```

"Đã xong rồi" **khác** "không làm được" — cùng bài học với `ordering.Reactor`
(§16d). Kết quả mình muốn đã là sự thật.

### 19f. Một câu SQL, một index

```sql
-- 0002_pricing.sql, viết từ ngày 2 cho đúng ngày hôm nay
CREATE INDEX quotes_open_idx ON quotes (expires_at) WHERE status = 'issued';
```

```sql
SELECT … FROM quotes WHERE status = 'issued' AND expires_at < $1 ORDER BY expires_at, id LIMIT $2
```

Index **partial**: nó chỉ chứa quote còn `issued`. Bảng `quotes` lớn dần mãi
mãi, nhưng index này **chỉ lớn theo số quote đang chờ** — số đó có trần, vì
sweep chính là thứ dọn nó.

`LIMIT` chặn một lượt chứ **không bỏ việc**: lượt sau làm tiếp, và `ORDER BY
expires_at` bảo đảm cái chờ lâu nhất được xử trước. Test
`TestExpireQuotes_limitSpreadsTheBacklogOverPasses` quét 5 quote bằng ba lượt
`limit = 2` rồi khẳng định lượt thứ tư tìm thấy 0.

### 19g. `scanQuote` — hai câu hỏi, một chỗ đọc cột

`ByID` đọc một dòng, `IssuedBefore` đọc nhiều dòng, nhưng 22 cột chỉ được biến
thành `Quote` **đúng một lần**:

```go
type scanner interface{ Scan(dest ...any) error }   // pgx.Row và pgx.Rows đều có
func scanQuote(s scanner) (*pricing.Quote, error)
```

Copy khối scan sang hàm thứ hai là tạo **chỗ thứ hai để một cột lệch đi** — và
kiểu lệch đó không làm test đỏ, nó làm số tiền sai.

### 19h. Bài học

| Bài học | Chỗ nó hiện ra |
|---|---|
| Có loại use case **không ai gọi** — thời gian là thứ đẩy nó | `ExpireQuotesHandler` |
| Vì thế `Clock` phải là port: 48 giờ trôi trong một dòng test | `clk.Advance(49h)` |
| **Ranh giới transaction = ranh giới aggregate**: lô không phải invariant nên không được nguyên tử | mỗi quote một `InTx` |
| Danh sách việc đọc ngoài tx, quyết định đọc lại **trong** tx — đừng ghi đè bằng dữ liệu cũ | `ByID` lần hai |
| "Đã xong rồi" ≠ "không làm được" — nuốt đúng hai sentinel, không nuốt hết | `errAlreadySettled` |
| `LIMIT` chặn **một lượt**, không bỏ việc; `ORDER BY` quyết cái nào đi trước | `IssuedBefore` |
| Index **partial** chỉ lớn theo phần còn mở, không theo cả bảng | `quotes_open_idx` |
| Hai câu hỏi, **một** chỗ đọc cột — chỗ thứ hai là chỗ để cột lệch | `scanQuote` |
| Sweep chờ một nhịp rồi mới chạy; pass lỗi là log, không phải chết | `worker.Sweeper` |
| **Bug của tôi (T1):** `smoke.ps1` chỉ chạy được **một lần** mỗi database — bootstrap chỉ nổ khi `api_tokens` rỗng. Lần chạy thứ hai 401 sạch. Sửa: script tự xoá bảng trước | `scripts/smoke.ps1` |

---

## 20. Read model thật — một bảng, năm nguồn

Cho tới §19, "đọc" trong project này là **đọc lại chính aggregate ghi ra nó**:
`GET /orders/{id}` load `CustomerOrder` rồi bày ra. Nó đúng, và nó có một trần
rất thấp — trần đó là chỗ mục này bắt đầu.

### 20a. Câu hỏi mà không aggregate nào trả lời được

Màn hình "đơn của tôi" cần bảy thứ, và chúng **không thuộc về ai cả**:

```
   tên sản phẩm     catalog     biết
   trạng thái đơn   ordering    biết
   đã cọc chưa      ordering    biết
   mã đơn ở shop    procurement biết
   thùng tới đâu    logistics   biết
   hoàn bao nhiêu   ordering    biết
   giá bao nhiêu    ordering    biết (đã chụp từ pricing)
```

Bốn context. `GET /orders/{id}` chỉ với ordering thì mãi mãi không hiển thị nổi
tên sản phẩm — nó **không được phép** đọc bảng của catalog.

Ba lối thoát, và vì sao hai cái đầu sai:

| Cách | Vì sao không |
|---|---|
| JOIN bốn bảng lúc đọc | biến bốn bounded context thành **một cục** không tách nổi: đổi schema của catalog làm hỏng màn hình của ordering |
| Gọi API nội bộ của nhau | một context chết là màn hình chết; và bây giờ có bốn cuộc gọi mạng cho một cái list |
| **Một bảng phẳng, ghi bằng event** | ✅ CQRS (DDD.md §24) |

### 20b. Package không có domain

```
internal/app/reporting/   ← package DUY NHẤT không có internal/domain/… tương ứng
```

Đó không phải thiếu sót, đó là định nghĩa:

> **Read model không có invariant.** Không có gì trong nó có thể "sai" theo
> nghĩa nghiệp vụ, vì không có gì trong nó **quyết định** cái gì. Nó chỉ nhớ
> lại điều mà năm context đã tuyên bố.

Nên `OrderSummary` là struct **field exported hết**, không method, không
constructor, không validate. Giả vờ ngược lại là mời luật nghiệp vụ vào đúng
chỗ không được có luật nào.

Hai luật thay thế, và chúng đủ:

```
   1. CHỈ event được ghi vào bảng này — không đọc bảng ghi của context nào
   2. Mọi handler idempotent VÀ chịu được sai thứ tự
```

### 20c. "Khung" — mẹo mua sự chịu-sai-thứ-tự

Relay là at-least-once và **không** bảo đảm thứ tự giữa các luồng. Nghĩa là
`batch_shipped` có thể tới trước `order_placed`.

Cách sai: *"event cho đơn tôi chưa biết → bỏ qua"*. Cách đó im lặng mất dữ
liệu đúng vào ngày relay gửi lại thứ gì đó lệch thứ tự.

Cách làm ở đây — **bất kỳ event nào cũng được TẠO dòng**:

```go
s, err := p.deps.Summaries.ByOrder(ctx, id)
if errors.Is(err, ErrSummaryNotFound) {
	s = OrderSummary{Order: id, Tracking: TrackingNone}   // cái KHUNG
}
```

Và đúng một field được bảo vệ:

```go
if s.Status == "" {                       // chỉ order_placed đặt status ĐẦU TIÊN
	s.Status = ordering.StatusAwaitingDeposit
}
```

Vì sao chỉ nó? Vì một cái khung do `batch_shipped` tạo ra **chưa biết** trạng
thái, và đoán một cái là **nói dối** với khách đang nhìn màn hình.

Cùng lý do với timestamp:

```go
if at.After(s.UpdatedAt) {   // event mới nhất thắng; bản gửi lại KHÔNG kéo lùi
	s.UpdatedAt = at
}
```

### 20d. Tên sản phẩm tới muộn

Sản phẩm được publish **trước** khi có ai đặt — thường là thế. "Thường" không
phải bảo đảm, nên `OnProductPublished` làm hai việc:

```go
p.deps.Names.Save(ctx, product, m.Name)              // nhớ tên
rows, _ := p.deps.Summaries.ByProduct(ctx, product)  // và VÁ những dòng đã có
for _, s := range rows { s.ProductName = m.Name; … }
```

Bảng `product_names` bé tí tồn tại đúng vì hai thứ này tới theo thứ tự nào
cũng được.

### 20e. Hai trục, không phải một

```
   status    ordering sở hữu   awaiting_deposit → deposited → purchased → in_transit → delivered
   tracking  logistics sở hữu  none → expected → received → shipped
```

Gộp lại thì tiện hơn, và sai: `received` (đã cân ở Denver) **không phải** một
trạng thái của đơn hàng — đơn vẫn đang là `purchased`. Trộn hai vòng đời vào
một cột là cách chắc chắn nhất để sau này không tách ra được nữa.

### 20f. "Me" là token, không phải tham số

```go
mux.HandleFunc("GET /me/orders", requireCustomer(s.myOrders))
```

`/orders/{customer_id}` sẽ là một route mà tính an toàn của nó phụ thuộc vào
việc handler **nhớ** so sánh hai id. Ngày ai đó quên, mọi khách đọc được đơn
của mọi khách.

> Một route **không gọi tên được** khách khác thì không làm lộ được khách nào.

Còn `GET /orders?status=` là hàng đợi của operator — và `?status=deposted` (gõ
sai) trả **400**, không phải danh sách rỗng: rỗng trông y hệt "hết việc rồi",
câu trả lời tệ nhất có thể cho một lỗi chính tả trong màn hình vận hành.

Và `shop_reference` chỉ operator thấy: mã đơn ở shop là số nội bộ, đưa cho
khách chỉ mời họ gọi thẳng cho shop về một đơn shop nghĩ là của người khác.

### 20g. Bảng không có FOREIGN KEY

```sql
CREATE TABLE order_summaries (
    "order" uuid PRIMARY KEY,
    customer uuid,          -- KHÔNG references
    product  uuid,          -- KHÔNG references
    ...
);
```

Hai thứ cố tình vắng mặt:

- **Khoá ngoại.** Read model tham chiếu sang bảng ghi thì hai bên không bao giờ
  deploy/migrate/tách riêng được nữa — và sẽ **lỗi** đúng vào lần relay gửi một
  event tới trước cái dòng nó "tham chiếu".
- **NOT NULL.** Dòng nửa vời là trạng thái **bình thường** ở đây (cái khung),
  không phải dữ liệu hỏng.

Đổi lại, adapter Postgres phải cẩn thận đúng một chỗ: id/tiền/thời gian bằng
zero phải xuống DB thành `NULL`, chứ không phải
`00000000-0000-0000-0000-000000000000` — một id giả đọc lên sẽ thành id thật.

### 20h. 12 dòng subscribe

```go
screens := reportingapp.NewProjector(g.Reporting)
bus.Subscribe("catalog.product_published", on(screens.OnProductPublished)) // người nghe THỨ BA
bus.Subscribe("ordering.order_placed",     on(screens.OnOrderPlaced))
… 10 dòng nữa …
```

Bảng định tuyến từ 16 lên **29 dòng** (sau đợt variant/size ở §22 là **31**). Nhìn
cột người-nghe là thấy hình dạng CQRS: bên ghi **chia nhỏ** theo invariant, bên đọc
**ráp lại** theo màn hình.

### 20i. Bài học

| Bài học | Chỗ nó hiện ra |
|---|---|
| Màn hình cần dữ liệu của nhiều context → **đừng JOIN**, hãy dựng một bảng bằng event | `order_summaries` |
| Read model **không có domain** vì nó không có invariant nào để giữ | `internal/app/reporting` |
| Chịu sai thứ tự mua bằng cái **khung**: event nào cũng được tạo dòng | `update()` |
| Nhưng vẫn có field chỉ MỘT event được đặt — đoán là nói dối với người dùng | `if s.Status == ""` |
| Bản gửi lại **không được kéo lùi** timestamp | `if at.After(s.UpdatedAt)` |
| Hai vòng đời khác chủ → **hai cột**, đừng gộp | `status` vs `tracking` |
| Route không gọi tên được người khác thì không lộ được người nào | `GET /me/orders` |
| Lọc sai chính tả phải **400**, đừng trả danh sách rỗng | `?status=deposted` |
| Read model **không FK, không NOT NULL** — và vì thế zero phải xuống DB là NULL | `0007_reporting.sql`, `idOrNil` |

---

## 21. ACL thứ hai — cho AI đọc trang web, mà không để nó quyết gì

`procurement.MerchantACL` bảo vệ mình khỏi **API của công ty khác**. Mục này
thêm cái thứ hai, bảo vệ mình khỏi thứ khó chịu hơn nhiều: **một trang web lạ
và một phỏng đoán của mô hình ngôn ngữ**.

### 21a. Cùng một hình dạng, hai nỗi lo khác nhau

```
   MerchantACL       → "đặt hàng ở shop này"     rủi ro: API người ta đổi, người ta chết
   ListingExtractor  → "trang này bán cái gì"    rủi ro: mô hình ĐOÁN, và đoán rất tự tin
```

Port ở tầng app, không phải domain — cùng lý do với `auth` ở §18a: *"ai đọc hộ
trang này"* không phải quy tắc nghiệp vụ.

```go
// internal/app/catalog/extract.go
type ListingDraft struct{ Name, Price, Currency, CategoryHint string }   // BỐN string
type ListingExtractor interface {
	Extract(ctx context.Context, url catalog.SourceURL) (ListingDraft, error)
}
```

**Bốn string, không phải `shared.Money`.** Đó là điểm quan trọng: nếu port trả
`Money` thì port đang quyết định thế nào là hợp lệ — mà đó là việc của domain.
Chuỗi đi qua đúng cái cửa mà chuỗi người gõ tay đi qua:

```go
cur, err := shared.CurrencyFromCode(draft.Currency)   // "XYZ" → 400 unknown_currency
price, err := shared.ParseMoney(draft.Price, cur)     // "about 150" → 400 malformed_amount
```

Model trả `"about $150"` thì chết **ở đây**, không phải ba cái bảng sau.

### 21b. Máy được TẠO nháp, không được XÁC NHẬN

```go
SourcedBy: catalog.SourcedByFeed,   // máy đưa vào, chưa ai vouch
```

Và `Product.Publish()` từ chối listing chưa `Verified()`. Ghép hai điều đó lại:

> **Không thứ gì mô hình bịa ra có thể thành giá bán cho tới khi một con người
> bấm confirm-listing.**

Đó không phải một `if` ở đâu đó trong route AI — nó là **luật cũ của catalog**
mà route mới không được miễn trừ. Tính năng mới không nới luật cũ.

Cùng ý: route đòi `category` **bắt buộc**, dù model có trả `category_hint`.
Category quyết thuế và cân nặng ước lượng — tiền thật. Một phỏng đoán không kèm
độ tin cậy không phải cơ sở cho việc đó, nên gợi ý trả về cho người xem và
không tự chọn gì.

### 21c. Không SDK, không cào trang

Adapter là `net/http` + `encoding/json`, khoảng trăm dòng.

- **Không SDK**: thêm một dependency, thêm một cách cấu hình timeout, và một bộ
  kiểu riêng muốn rò lên tầng trên. Port là bốn string; SDK không giúp gì.
- **Không cào trang**: Portage **không** tự fetch trang shop (CATALOG.md §2 —
  điều khoản của Nike cấm, và của phần lớn shop khác cũng vậy). Chỉ cái URL
  được gửi cho model.

Và cái ép model trả lời được là `json_schema` với `strict: true` — nhờ vậy
adapter **không có một cái regex nào**.

### 21d. Mọi kiểu hỏng đều là một câu trả lời

```
429, 500, mạng đứt, JSON rác, JSON đúng cú pháp sai schema, không có key
        → tất cả bọc thành catalogapp.ErrExtractorUnavailable → HTTP 503
```

Vì sao gộp: chia nhỏ ra thì mỗi người gọi phải tự quyết *"cái này thử lại được
không"*. Một mã, một câu trả lời: **thử lại sau, hoặc gõ tay**.

Ca tinh tế nhất: model trả `{"product":"Air Trainer 90"}` — JSON hợp lệ, decode
vào struct **không lỗi**, mọi field rỗng. Adapter bắt bằng cái tên rỗng:

```go
if strings.TrimSpace(draft.Name) == "" {
	return …, fmt.Errorf("openai: the model did not recognise %s: %w", url, ErrExtractorUnavailable)
}
```

Cho lọt thì nó nổi lên thành `ErrEmptyName` của domain — một 400 **đổ lỗi cho
người gọi** vì thứ họ không hề gửi.

### 21e. Tắt phải TRÔNG như đang tắt

```go
// wire.Postgres
var extractor catalogapp.ListingExtractor = openai.Unavailable{}
if key := os.Getenv("OPENAI_API_KEY"); key != "" {
	extractor = openai.New(key)
}
```

`wire.Memory` dùng `openai.Fake` để `go run ./cmd/api` chạy được không cần key,
không cần mạng, không tốn tiền. Postgres thì **tuyệt đối không** rơi về Fake:

> API sẽ trả **201**, một sản phẩm nháp tồn tại với tên và giá bịa, và dấu hiệu
> duy nhất cho thấy có gì sai là các con số toàn là hư cấu. 503 tệ hơn nhiều về
> mặt trải nghiệm và tốt hơn nhiều về mặt sự thật.

### 21f. `merchant.Router` — chỗ để cắm shop có API

```go
type Router struct {
	Items      procurement.ItemRepository            // product → shop
	ByMerchant map[shared.ID]procurement.MerchantACL
	Fallback   procurement.MerchantACL               // nil = Manual
}
```

Router **chính nó** là một `MerchantACL`, nên không ai bên trên biết nó tồn tại.
Thêm một shop có API = một dòng trong map, ở `wire` — nơi duy nhất được phép
biết những adapter nào có mặt.

Hai lựa chọn còn lại đều tệ hơn:

| | Vì sao không |
|---|---|
| `if` trong `OpenTaskHandler` | nhét chuyện triển khai vào use case, và mọc thêm một nhánh mỗi shop, mãi mãi |
| Một ACL biết mọi shop | shop nào chết cũng thành mọi shop chết; thêm shop = sửa code dùng chung |

Một chi tiết đáng nhớ: projection `Items` là **eventual** (§25). Task có thể tồn
tại trước khi procurement nghe được sản phẩm thuộc shop nào.

```go
if errors.Is(err, procurement.ErrItemNotFound) {
	return fallback, nil    // KHÔNG biết ≠ hỏng: giao cho người mua
}
```

Không biết shop mà lại **fail** cái task thì sẽ huỷ một đơn hàng vì một cuộc
đua vài trăm mili-giây.

### 21g. `POST /fx` — route duy nhất không phát event

```go
func (h *SetExchangeRateHandler) Handle(ctx context.Context, rate shared.ExchangeRate) error {
	return h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return h.deps.Rates.Set(ctx, rate)
	})
}
```

Mọi lệnh ghi khác trong hệ đều kết thúc bằng một event, vì có ai đó giữ bản sao.
Tỷ giá thì **không ai giữ bản sao**: `Quote` không *đọc* tỷ giá về sau, nó đã
**đóng băng** tỷ giá vào breakdown của chính nó lúc phát hành (§28).

Nên đổi tỷ giá hôm nay chỉ đổi cái **quote tiếp theo** — đúng bản chất của một
tỷ giá: câu trả lời hiện tại cho một câu hỏi, không phải một sự kiện quá khứ.

Và luật giữ nó an toàn nằm ở domain, không ở đây: `shared.ExchangeRate` từ chối
tỷ giá ≤ 0 và cặp `from == to`, nên không ai publish được "1 USD = 0 VND" rồi kéo
mọi báo giá sau đó về 0.

> **Bẫy đã gặp:** có **hai** sentinel tên `ErrInvalidRate` — `pricing` (một giá
> trong bảng cước) và `shared` (một tỷ giá). Cùng tên, khác package, khác cách
> sửa. Chúng phải là **hai mã lỗi**: `invalid_rate` và `invalid_exchange_rate`,
> nếu không client được chỉ đi sửa sai chỗ.

### 21h. Quản chìa: `GET /tokens`, `DELETE /tokens/{hash}`

Món nợ của §18 đã trả. Điều đáng học nằm ở **cái handle**:

```
   token   chỉ tồn tại MỘT lần, lúc phát        → không bao giờ liệt kê được
   hash    là khoá chính của kho                → an toàn để hiện, và là cái để thu hồi
```

Nên `DELETE /tokens/{hash}`, không phải `{token}`. Operator nhìn danh sách chưa
bao giờ thấy token và sẽ không bao giờ thấy.

Ba quyết định nhỏ, mỗi cái một lý do:

| Quyết định | Vì sao |
|---|---|
| Thu hồi chìa đã thu hồi, hoặc hash không tồn tại → **204** | ý muốn của người gọi đã thành sự thật; trả 404 biến route này thành cách hỏi "hash này có thật không" |
| Postgres giữ dòng với `revoked_at`, Static thì xoá hẳn | *"ai từng có quyền, tới khi nào"* là câu kiểm toán thật sự hỏi; adapter dev không có kiểm toán nào để trả lời |
| Chìa operator **hoạt động cuối cùng** → 409 `last_operator_key` | API không ai quản được không phải API an toàn hơn — nó là API chỉ sửa được bằng cách vào DB gõ tay |

### 21i. Bài học

| Bài học | Chỗ nó hiện ra |
|---|---|
| ACL cho AI: port trả **string thô**, để domain quyết cái gì hợp lệ | `ListingDraft` bốn string |
| Máy được **tạo nháp**, không được **xác nhận** — luật cũ không được nới cho tính năng mới | `SourcedByFeed` + `Publish` |
| Gợi ý của máy **không tự chọn** thứ dính tới tiền | `category` vẫn bắt buộc |
| Mọi kiểu hỏng của một dịch vụ ngoài → **một** sentinel, một mã | `ErrExtractorUnavailable` → 503 |
| JSON hợp lệ mà rỗng cũng là hỏng — bắt nó, đừng để domain đổ lỗi nhầm người | tên rỗng → unavailable |
| Tính năng tắt phải **trông như** đang tắt; đừng bao giờ rơi về Fake ở production | `openai.Unavailable` |
| Không cào trang người ta — điều khoản là ràng buộc thiết kế, không phải chú thích | chỉ gửi URL |
| Định tuyến ACL theo shop nằm ở **composition root**, không phải trong use case | `merchant.Router` |
| "Không biết" ≠ "hỏng": projection chưa kịp thì giao cho người, đừng fail | `ErrItemNotFound` → fallback |
| Ghi mà **không ai giữ bản sao** thì không cần event | `POST /fx` |
| Hai sentinel trùng tên khác package = **hai mã lỗi**, không phải một | `invalid_exchange_rate` |
| Thu hồi bằng **hash**, vì token chỉ tồn tại một lần | `DELETE /tokens/{hash}` |
| Đừng cho phép tự khoá mình ra ngoài | 409 `last_operator_key` |

---

## 22. Một chữ của shop đi qua ba context — `size`

> Đợt này không thêm context nào, không thêm aggregate nào. Nó sửa **hai chỗ hở**
> mà chỉ lộ ra khi ngồi bấm thật trên UI: `POST /orders` nhận **bất kỳ** uuid làm
> `variant_id`, và màn hình người đi mua chỉ có bốn cái uuid với một mã tiền tệ.
> Cả hai đều là cùng một nguyên nhân: **size chưa bao giờ ra khỏi catalog**.

### 22a. Bắt đầu từ câu hỏi của người dùng

Câu hỏi thật: *"Khách nhập size kiểu `M 8 / W 9.5` thì input có đỡ được không?"*

Đọc code trả lời được ngay phần dễ: `size` là **string tự do**, không enum, nên
`M 8 / W 9.5`, `42 EU`, `XS` đều vào được. Nhưng đọc tiếp thì lộ ba chuyện:

| Chỗ hở | Trước đợt này |
|---|---|
| Khoá chống trùng | `variantKey` chỉ hạ chữ thường + `TrimSpace` → `M8/W9.5` và `M 8 / W 9.5` là **hai** variant |
| `POST /orders` | không kiểm `variant_id` gì cả: uuid nào cũng thành đơn |
| Màn hình đi mua | `taskView` trả 4 uuid; không ai cầm được nó vào shop |

Chỗ hở thứ hai là chỗ đau nhất: khách gõ tay hay client gửi sai một ký tự thì
người đi mua nhận một task **không thể thực hiện**, sau khi khách **đã trả cọc**.

### 22b. Vì sao không phải là một câu `JOIN`

Câu hỏi tự nhiên: procurement cần tên sản phẩm và size — sao không `SELECT` sang
bảng của catalog?

Vì guard 7 trong `internal/domain/decisions_test.go` cấm context này import
context kia, và bảng cũng vậy: hai context dùng chung một database hôm nay chỉ vì
tiện, không phải vì được phép. Cùng lý do đã có từ §14: **catalog phải KỂ**, và
payload event là **tất cả** những gì bên nghe có.

Nên `catalog.variant_added` là event thứ 36 — và nó ra đời ở đúng chỗ đã có sẵn
một event trống chỗ:

```go
// catalog/product.go — AddVariant giờ ghi lại việc mình vừa làm
p.variants = append(p.variants, v)
p.Record(VariantAdded{
    ID: p.id, Variant: v.id, Size: v.size, Color: v.color, MerchantRef: v.merchantRef, At: now,
})
```

`ProductPublished` cũng được thêm một field: `Source` — link gốc. Người đi mua
cần **trang để mở**, và trước đó chỉ catalog biết nó.

### 22c. Hai người nghe, hai mức chép khác nhau

Đây là phần đáng đọc nhất của đợt này. Cùng một event, hai context nghe, và
**chép khác nhau vì luật của họ khác nhau**:

```go
// ordering/variant.go — mỏng nhất có thể
type Variant struct {
    Variant shared.ID
    Product shared.ID
}
```

```go
// procurement/repository.go — đủ chữ để in ra cho người đọc
type Variant struct {
    Variant     shared.ID
    Product     shared.ID
    Size        string
    Color       string
    MerchantRef string
}

func (v Variant) Label() string   // "M 8 / W 9.5 · black", hoặc size, hoặc màu, hoặc ""
```

Ordering không giữ `size` **cố ý**. Luật của nó là: *"variant này có thật không,
và có phải của sản phẩm mà quote đã báo giá không"* — hai câu đó trả lời được bằng
hai id. Một projection giữ dữ liệu **không ai đọc** sẽ mốc mà không ai biết, vì
không test nào chạy qua nó.

```go
// app/ordering/place_order.go — kiểm cái duy nhất khách tự chọn
v, err := h.deps.Variants.ByID(ctx, cmd.Variant)
if errors.Is(err, ordering.ErrVariantNotFound) {
    return fmt.Errorf("variant %s: %w", cmd.Variant, ordering.ErrVariantUnknown)
}
if v.Product != q.Product {
    return fmt.Errorf("variant %s is a form of product %s, quote %s prices %s: %w",
        cmd.Variant, v.Product, cmd.Quote, q.Product, ordering.ErrVariantNotForProduct)
}
```

Thứ tự kiểm **có ý nghĩa**: quote → "một quote một đơn" → variant. Đổi thứ tự là
đổi mã lỗi client nhận được ở trường hợp sai cả hai, và test vàng bắt đúng chuyện đó.

### 22d. `Subject` — chép, không tra lại

Procurement có bảng chiếu đủ chữ rồi, nhưng `taskView` **không** tra bảng đó lúc
đọc. Nó đọc từ chính task:

```go
// procurement/task.go
type Subject struct {
    ProductName  string
    VariantLabel string   // "M 8 / W 9.5 · black"; rỗng với sản phẩm một size
    VariantRef   string   // mã của shop, nếu shop có
    Source       string   // trang để mở ra mà mua
}
```

`TaskDetails` giờ có 5 field, và **chỉ field thứ 5 là không bắt buộc** — test
`TestOpenTask_validatesEveryField` canh đúng câu đó bằng `reflect`, cộng một ca
mới: `Subject{}` rỗng vẫn mở được task.

Vì sao chép mà không tra? Ba lý do, xếp theo mức quan trọng:

| Lý do | Nếu tra lại lúc đọc |
|---|---|
| Việc đã giao thì **không đổi nội dung** | shop rename sản phẩm ngày mai → task hôm qua nói khác đi |
| Bảng chiếu là **eventual** | `variant_added` chưa relay tới → màn hình đi mua trống, dù task có thật |
| Task là **bằng chứng** | biên nhận đối chiếu với thứ đã in ra lúc mở, không với thứ đọc lại lúc tranh chấp |

Và vì bảng chiếu là eventual, `OpenTask` **chịu được** việc thiếu dòng variant:

```go
subject := procurement.Subject{ProductName: item.Name, Source: item.Source}
if v, err := h.deps.Variants.ByID(ctx, cmd.Variant); err == nil {
    subject.VariantLabel, subject.VariantRef = v.Label(), v.MerchantRef
} else if !errors.Is(err, procurement.ErrVariantNotFound) {
    return err
}
```

Task vẫn mở, chỉ là không có size — đúng như một sản phẩm một size. Fail cái task
sẽ **huỷ một đơn đã trả cọc** vì một cuộc đua vài trăm mili giây. Cùng lối suy
nghĩ với `ErrItemNotFound` ở §16.

### 22e. Hai luật domain mới, cả hai đều là "đừng để hai thứ giống nhau tồn tại"

```go
// catalog/variant.go — khoá chống trùng bỏ MỌI khoảng trắng
func variantKey(s string) string {
    return strings.ToLower(strings.Join(strings.Fields(s), ""))
}
```

`M8/W9.5` và `M 8 / W 9.5` giờ là **một**. Nhưng `8 M / 9.5 W` **vẫn khác** — đảo
thứ tự là chữ khác của shop, và mình không có quyền đoán rằng hai cách viết đó
cùng nghĩa.

Luật thứ hai đến từ một lần tôi tự sửa mình. Đề xuất ban đầu của tôi là *"chặn
variant không size không màu"*. Đọc comment của `VariantDetails` thì thấy nó có
chủ đích: *một sản phẩm một size có đúng một variant không có gì để nói, và đó là
variant hợp lệ* — có cả test cho nó. Nên luật đúng là chặt hơn mà **không phá**:

```go
ErrUnnamedVariant = errors.New("a variant with no size or colour must be the product's only one")
```

Không size không màu → chỉ được là variant **duy nhất** của sản phẩm. Sản phẩm một
size vẫn chạy; còn "một cái US 9 và một cái không tên" thì không, vì lúc đi mua
không ai biết cái không tên là cái gì.

### 22f. Ba mã lỗi mới, và vì sao một trong ba là **thử lại được**

| Sentinel | HTTP | Ý nghĩa cho client |
|---|---|---|
| `ordering.ErrVariantUnknown` | 409 `variant_unknown` | id này mình chưa từng phát — **hoặc** `variant_added` chưa relay tới |
| `ordering.ErrVariantNotForProduct` | 409 `variant_not_for_product` | id có thật, nhưng của sản phẩm khác — từ chối thật, thử lại vô nghĩa |
| `catalog.ErrUnnamedVariant` | 409 `unnamed_variant` | variant không tên phải là variant duy nhất |

`variant_unknown` có **hai nghĩa** như `listing_not_found` ở §14 — nên FE xử đúng
kiểu đó: `web/app/steps.js` để nó vào danh sách `retryOn` của bước đặt hàng, cùng
với `quote_not_accepted`. Nhất quán sau cùng thì phải chờ được, không phải báo đỏ.

### 22g. Kết quả thật — `scripts/smoke.ps1`, 08/09

```
merchant … / product …
published; waiting for the worker to relay product_published...
quote …: issued branded chargeable 2500 g, total 5393720 VND, deposit 2696860
order …: deposited; buyer's list: 1 open task(s), first for order … in USD
buy: Air Trainer 90 / US 9 · black / ref EX-AT90-9-BLK / https://…/t/air-trainer-90/abc
order at the end: delivered; quote vs actual: quoted 163.22+25.00, actual 163.22+27.50 → variance -2.50 USD

--- worker log ---
#331 catalog.merchant_registered   #332 catalog.product_added
#333 catalog.variant_added         ← dòng mới của đợt này
#334 catalog.product_measured      #335 catalog.product_published
#336 quote_issued  #337 quote_accepted  #338 order_placed  #339 deposit_paid
#340 purchase_task_opened  #341 purchase_confirmed  #342 order_purchased
#343 parcel_expected  #344 parcel_received  #345 batch_opened  #346 batch_closed
#347 batch_shipped  #348 order_shipped  #349 balance_paid  #350 order_delivered
```

Dòng `buy:` là toàn bộ mục đích của đợt này: **một người đọc được nó và đi mua
được**. Số vàng không đổi một chữ — 2500 g, 5.393.720 ₫, 2.696.860 ₫, variance
−2.50 USD.

`smoke.sh` (CI) thì **assert** luôn cái label, chứ không chỉ in:

```bash
[ "$LABEL" = "US 9 · black" ] || { echo "FAIL: the buyer's task says '$LABEL'"; fail=1; }
```

### 22h. Bài học

| Bài học | Chỗ nó hiện ra |
|---|---|
| Lỗ hở lớn nhất lộ ra khi **bấm thật trên UI**, không phải khi đọc test | `taskView` toàn uuid |
| Hai context nghe cùng một event được phép chép **khác nhau** — theo luật của mình | `ordering.Variant` 2 field vs `procurement.Variant` 5 field |
| Projection giữ dữ liệu **không ai đọc** thì sẽ mốc mà không ai biết | vì sao ordering không giữ `size` |
| Việc đã giao cho người thì **chép** nội dung, đừng tra lại lúc đọc | `PurchaseTask.Subject` |
| Bảng chiếu eventual thiếu dòng → **giảm chất lượng**, không phải hỏng | `OpenTask` bỏ qua `ErrVariantNotFound` |
| Một mã lỗi mang hai nghĩa thì phải nói rõ, và FE phải **thử lại** | `variant_unknown` trong `retryOn` |
| Chữ tự do vẫn cần **khoá chuẩn hoá**; nhưng đừng chuẩn hoá tới mức đoán hộ shop | bỏ khoảng trắng, **không** đảo thứ tự |
| Đọc comment cũ trước khi đổi luật: nó có thể đang bảo vệ một ca có thật | `VariantDetails` → `ErrUnnamedVariant` |
| Thêm event = thêm dòng `Subscribe` + dòng `eventcodec` + hàng test hợp đồng | guard `TestEncode_contractCoversEveryDomainEvent` |
