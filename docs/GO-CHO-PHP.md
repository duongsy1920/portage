# Cú pháp Go cho người viết PHP

> Bảng tra cứu, đối chiếu trực tiếp với PHP 8 / Symfony. Ví dụ lấy từ **code
> thật trong project này**, không phải tutorial chung chung.
>
> Trong code còn có comment `// [PHP]` giải thích tại chỗ. Xoá hết bằng:
> `grep -rn "// \[PHP\]" internal/` để xem, hoặc dùng lệnh ở cuối file này.

---

## 1. Ba khác biệt lớn nhất — đọc kỹ ba cái này trước

### (a) Kiểu viết SAU tên

```php
// PHP
private int $minor;
function add(Money $o): Money
```
```go
// Go — ngược lại hoàn toàn
minor int64
func Add(o Money) Money
```

Đọc từ trái sang: *"biến `minor`, kiểu `int64`"*.

### (b) Không có exception — hàm trả về hai giá trị

```php
// PHP
function parseMoney(string $s): Money   // throws MalformedAmount
try { $m = parseMoney($s); } catch (MalformedAmount $e) { ... }
```
```go
// Go
func ParseMoney(s string, c Currency) (Money, error)

m, err := ParseMoney(s, USD)
if err != nil {
    return err          // xử lý ngay tại chỗ, không nhảy đi đâu cả
}
```

`(Money, error)` = trả **hai** giá trị. Người gọi buộc phải nhận cả hai. Không thể "quên try/catch".

### (c) Chữ HOA / thường thay cho `public` / `private`

```php
public  function code(): string     // ai cũng gọi được
private int $minor;                 // chỉ trong class
```
```go
func (c Currency) Code() string     // C hoa  → public, package khác gọi được
minor int64                         // m thường → private trong package
```

**Không có từ khoá `public`/`private`/`protected` nào cả.** Ký tự đầu quyết định tất cả. Phạm vi là **package** (thư mục), không phải class.

---

## 2. Bảng tra nhanh

| PHP | Go | Ghi chú |
|---|---|---|
| `namespace App\Domain;` | `package shared` | một thư mục = một package |
| `use App\X;` | `import "…/x"` | |
| `$x = 5;` | `x := 5` | `:=` tạo biến mới, Go tự suy kiểu |
| `$x = 5;` (đã có `$x`) | `x = 5` | `=` khi biến đã tồn tại |
| `int $x;` | `var x int` | khai báo, chưa gán |
| `null` | `nil` | |
| `final class X {}` | `type X struct {}` | struct **không chứa** method |
| `new X(...)` | `X{...}` | không có từ khoá `new` |
| `$this` | `m` trong `func (m Money)` | tự đặt tên, gọi là *receiver* |
| `__construct` | `func NewX(...) X` | chỉ là quy ước đặt tên |
| `__toString()` | `func (x X) String() string` | `fmt` tự gọi |
| `interface X {}` | `type X interface {}` | **không có** `implements` |
| `implements X` | *(tự động)* | đủ method là thoả interface |
| `extends` | *(không có)* | dùng embedding, xem §5 |
| `array` (list) | `[]T` — slice | `append(s, v)` |
| `array` (assoc) | `map[K]V` | `m[k]` |
| `$a[] = $v;` | `s = append(s, v)` | **phải gán lại** |
| `count($a)` | `len(s)` | |
| `isset($m[$k])` | `v, ok := m[k]` | "comma ok" |
| `foreach ($a as $v)` | `for _, v := range a` | |
| `foreach ($a as $k => $v)` | `for k, v := range a` | |
| `for (...)` / `while` | `for` | Go chỉ có **một** từ khoá lặp |
| `throw new X()` | `return errors.New(...)` | lỗi là **giá trị trả về** |
| `throw new LogicException` | `panic(...)` | chỉ cho **bug**, không cho lỗi user |
| `catch (X $e)` | `errors.Is(err, ErrX)` | |
| `previous: $e` | `%w` trong `fmt.Errorf` | bọc lỗi gốc |
| `sprintf()` | `fmt.Sprintf()` | `%d %s %q %v %w` |
| `trim($s)` | `strings.TrimSpace(s)` | chuỗi là **tham số** |
| `strlen($s)` | `len(s)` | |
| `(int) $x` | `int(x)` | |
| `const X = 1;` | `const X = 1` | |
| `enum` | `type X string` + hằng | Go **không có** enum thật |
| `new X(name: 'a', site: $s)` (named args) | `NewX(XDetails{Name: "a", Site: s})` | struct tham số — xem SETUP.md quy ước 10 |
| `phpunit` | `go test ./...` | |

---

## 3. Ký hiệu hay gặp mà không tra được trên Google

| Ký hiệu | Nghĩa | Ví dụ trong project |
|---|---|---|
| `:=` | tạo biến mới + gán | `minor, err := parseDecimal(...)` |
| `_` | biến vứt đi (bắt buộc, vì Go cấm biến thừa) | `for _, m := range rest` |
| `*T` | con trỏ tới T | `func (e *Events) Record(...)` |
| `&x` | lấy địa chỉ của x | `&CustomerOrder{...}` |
| `...T` | tham số biến thiên | `func Sum(first Money, rest ...Money)` |
| `s...` | bung slice thành nhiều tham số | `Sum(a, list...)` |
| `%w` | bọc lỗi (chỉ dùng trong `fmt.Errorf`) | `fmt.Errorf("…: %w", ErrX)` |
| `%q` | in chuỗi kèm nháy | `fmt.Errorf("parse %q", s)` |
| `%v` | in "giá trị mặc định" của bất cứ thứ gì | |
| `'a'` | **ký tự** (rune, là số) | `if r < '0'` |
| `"a"` | **chuỗi** | PHP không phân biệt hai cái này |
| `` `a` `` | chuỗi thô, không escape | |

---

## 4. Receiver — `(m Money)` và `(e *Events)` khác nhau thế nào

Đây là chỗ dễ nhầm nhất với người từ PHP sang.

```go
func (m Money) Add(o Money) (Money, error)   // KHÔNG có *  → Go COPY struct
func (e *Events) Record(ev Event)            // CÓ *        → sửa được bản gốc
```

| | `(m Money)` — giá trị | `(e *Events)` — con trỏ |
|---|---|---|
| Go làm gì | **copy** cả struct | truyền địa chỉ |
| Sửa được bản gốc? | ❌ không | ✅ có |
| Giống PHP thế nào | như `clone $this` mỗi lần gọi | như object PHP bình thường |
| Dùng cho | value object (`Money`, `Weight`) | thứ có trạng thái (`Events`, aggregate) |

> **PHP không có phân biệt này** — object luôn truyền theo tham chiếu. Đây là lý do value object của Go bất biến "miễn phí": chỉ cần bỏ dấu `*`.

---

## 5. Embedding — trông như kế thừa nhưng không phải

```go
type OrderID struct{ shared.ID }        // viết KIỂU, không đặt tên field

orderID.String()                        // dùng luôn method của shared.ID
```

```php
// Gần nhất bên PHP — nhưng KHÔNG tương đương
class OrderID extends ID {}      // ← không phải cái này (không có đa hình)
class OrderID { use IDTrait; }   // ← gần hơn
```

Khác kế thừa ở chỗ: **không đa hình, không override, không `parent::`**. Chỉ là "mượn method cho tiện". Go cố ý không có kế thừa.

---

## 6. Đọc thử một hàm thật, từng dòng

```go
func ParseMoney(s string, c Currency) (Money, error) {
	minor, err := parseDecimal(s, int(c.exponent))
	if err != nil {
		return Money{}, fmt.Errorf("parse %q as %s: %w", s, c.code, err)
	}
	return NewMoney(minor, c), nil
}
```

| Dòng | Đọc là |
|---|---|
| `func ParseMoney(` | hàm public tên ParseMoney |
| `s string, c Currency)` | nhận `$s` kiểu string, `$c` kiểu Currency |
| `(Money, error) {` | trả về **hai** thứ: một Money và một error |
| `minor, err := …` | tạo hai biến mới, nhận hai giá trị trả về |
| `int(c.exponent)` | ép `$c->exponent` sang int |
| `if err != nil {` | nếu có lỗi (`err` khác null) |
| `return Money{}, fmt.…` | trả Money **rỗng** + lỗi đã bọc thêm ngữ cảnh |
| `return NewMoney(…), nil` | trả kết quả + **nil** = "không có lỗi" |

Tương đương PHP:

```php
public static function parseMoney(string $s, Currency $c): Money
{
    try {
        $minor = self::parseDecimal($s, $c->exponent);
    } catch (MalformedAmount $e) {
        throw new MalformedAmount(sprintf('parse "%s" as %s', $s, $c->code), previous: $e);
    }
    return self::newMoney($minor, $c);
}
```

---

## 7. Những thứ Go **không có** (đừng tìm)

| PHP có | Go |
|---|---|
| `try / catch / finally` | không — dùng `if err != nil`; `defer` thay `finally` |
| Kế thừa class | không — dùng embedding + interface |
| `abstract` / `final` | không |
| `public` / `private` / `protected` | không — dùng chữ HOA/thường |
| Enum | không — dùng `type X string` + hằng |
| `null` cho mọi kiểu | không — mỗi kiểu có "zero value" riêng |
| Nạp chồng hàm (overload) | không — đặt tên khác nhau |
| Tham số mặc định | không — viết hàm khác, hoặc truyền struct |
| Constructor trong class | không — quy ước `NewXxx()` |
| `$this` ngầm định | không — receiver phải đặt tên |

---

## 8. Zero value — bẫy riêng của Go

PHP có `null` cho mọi thứ. Go thì **mỗi kiểu có giá trị mặc định riêng** và
**không cấm được** người ta tạo ra nó:

| Kiểu | Zero value |
|---|---|
| `int64`, `int32` | `0` |
| `string` | `""` |
| `bool` | `false` |
| slice `[]T`, map, con trỏ, interface | `nil` |
| struct `Money` | `Money{}` — mọi field = zero |

Hệ quả thực tế trong project: `shared.Money{}` là code hợp lệ, tạo ra "tiền
không có tiền tệ". Không chặn được lúc biên dịch → phải chặn bằng logic:

```go
func (m Money) IsValid() bool { return !m.currency.IsZero() }

func NewMoney(minor int64, c Currency) Money {
	if c.IsZero() {
		panic("shared: NewMoney called with the zero Currency")
	}
	...
}
```

> Coi zero value như **cột nullable chưa gán** bên Doctrine: có hàm để hỏi, và
> constructor từ chối xây lên trên nó.

---

## 8b. Bẫy `==` với struct chứa `time.Time` hoặc slice

Ở §4 có nói value object của Go so sánh bằng `==` rất tự nhiên. **Có hai trường
hợp làm hỏng điều đó**, và cả hai đã gặp thật trong project.

### (a) `time.Time` — hai mốc bằng nhau vẫn có thể `!=`

```go
type Provenance struct {
    source SourcingMode
    at     time.Time        // ← chứa con trỏ *Location bên trong
    by     shared.OperatorID
}

a == b   // ❌ KHÔNG dùng được: cùng một thời điểm nhưng khác múi giờ → false
```

`time.Time` mang theo một con trỏ `*Location`. Hai giá trị chỉ **cùng một thời
điểm** vẫn có thể khác nhau về bit — ví dụ một cái đọc từ DB (UTC), một cái tạo
trong code (giờ máy).

**Cách đúng:**

```go
a.At().Equal(b.At())          // so THỜI ĐIỂM, bỏ qua múi giờ
```

Và vì `==` không dùng được, `IsZero()` phải **viết tay** thay vì `p == Provenance{}`:

```go
func (p Provenance) IsZero() bool { return p.source == "" && p.at.IsZero() }
```

> PHP không có bẫy này vì `DateTimeImmutable` so sánh bằng `==` theo giá trị,
> còn `===` thì so sánh danh tính object — khác cơ chế hoàn toàn.

### (b) Slice và map — `==` không biên dịch được

```go
type Product struct {
    variants []Variant       // ← có slice
}
p1 == p2                     // ❌ LỖI BIÊN DỊCH: struct chứa slice không so sánh được
```

Go **cấm hẳn** ở mức biên dịch, không âm thầm sai như `time.Time`. Đỡ nguy hiểm
hơn, nhưng vẫn phải biết.

### Bảng tra

| Struct chứa | `==` dùng được? | Làm sao |
|---|---|---|
| chỉ số, chuỗi, bool | ✅ | dùng bình thường — `Money`, `Weight`, `Hostname`, `CategoryCode` |
| `time.Time` | ⚠️ **biên dịch được nhưng SAI** | `.Equal()`, và `IsZero()` viết tay |
| slice, map, func | ❌ **lỗi biên dịch** | so sánh từng phần, hoặc `slices.Equal` |
| con trỏ | ⚠️ so **địa chỉ**, không so giá trị | tự viết `Equals()` |

> **Quy tắc thực dụng:** value object nên chỉ chứa số/chuỗi để `==` dùng được —
> đó là một phần lý do `Money` giữ `int64` chứ không giữ `big.Int` (là con trỏ).
> Khi buộc phải chứa `time.Time`, viết `IsZero()` và `Equals()` bằng tay và ghi
> chú ngay trên struct, như `Provenance` đang làm.

---

## 9. Vòng đời chương trình Go — từ `go run` tới khi trả response

Đây là chỗ khác PHP **nhiều nhất**, và hiểu nó thì mọi thứ khác tự sáng.

### 9.1 Khác biệt gốc: PHP chết sau mỗi request, Go thì sống mãi

```
 ══ PHP / Symfony + PHP-FPM ═══════════════════════════════════════
   request 1 ──▶ nginx ──▶ php-fpm worker
                   │  autoload → Kernel::boot() → build container
                   │  → route → Controller → response
                   ▼
              💀 TIẾN TRÌNH CHẾT — mọi biến biến mất

   request 2 ──▶ ... LÀM LẠI TỪ ĐẦU, không nhớ gì từ request 1


 ══ Go ════════════════════════════════════════════════════════════
   go build ──▶ api.exe                (một file, không cần runtime)

   chạy api.exe MỘT LẦN:
        │  init package → main() → dựng pool DB, logger, handler
        │  → http.ListenAndServe(...) → ĐỨNG CHỜ
        │
        ├── request 1 ──▶ goroutine #1 ──┐
        ├── request 2 ──▶ goroutine #2 ──┤ DÙNG CHUNG mọi thứ
        ├── request 3 ──▶ goroutine #3 ──┘ đã dựng ở main()
        │
        ▼
   🟢 TIẾN TRÌNH VẪN SỐNG — nhớ mọi thứ giữa các request
```

**Một câu tóm tắt:** PHP dựng lại thế giới mỗi request; Go dựng thế giới **một
lần** rồi phục vụ mọi request trên đó.

Hệ quả có thật, không phải lý thuyết:

| | PHP | Go |
|---|---|---|
| Kết nối DB | mở/đóng mỗi request (hoặc pool ngoài) | **một pool duy nhất**, dựng ở `main()` |
| Biến toàn cục | chết theo request | **sống mãi** → phải lo tranh chấp (race) |
| Rò rỉ bộ nhớ | gần như không gặp | **có thật**, tích luỹ theo ngày |
| Cache trong RAM | không giữ được | giữ được, và rất nhanh |
| Lỗi chí mạng | giết 1 request | `panic` không bắt **giết CẢ tiến trình** |
| Deploy | copy file `.php` | build lại `.exe`, khởi động lại |

### 9.2 Thứ tự khởi động — cái gì chạy trước cái gì

Trước khi dòng đầu tiên của `main()` chạy, Go đã làm ba việc:

```
  ①  Nạp package theo cây phụ thuộc (sâu nhất trước)
         shared  →  catalog  →  app  →  adapter  →  main

  ②  Khởi tạo BIẾN CẤP PACKAGE của từng package
         var USD = Currency{code: "USD", exponent: 2}
         var currencies = map[string]Currency{...}
         ↑ hai dòng này chạy XONG trước khi main() bắt đầu

  ③  Chạy func init() của từng package (nếu có)

  ④  main()  ← bây giờ mới tới lượt bạn
```

Trong project này, `shared.USD`, `shared.VND`, `shared.currencies` đều là biến
cấp package — chúng **đã tồn tại** trước khi `main()` chạy dòng nào.

> **So với Symfony:** giống lúc container được build và các service `public: true`
> được khởi tạo. Khác ở chỗ Go làm việc này **một lần cho cả đời tiến trình**,
> còn PHP làm lại mỗi request (dù có cache).

`func init()` là hàm đặc biệt: không tham số, không giá trị trả về, **không gọi
được bằng tay**, Go tự chạy. Dùng ít thôi — thứ tự khó lần, và nó chạy cả khi
chạy test.

### 9.3 `main()` là **Composition Root** — thay cho `services.yaml`

Go **không có** dependency injection tự động. Không autowire, không container.
Bạn ráp tay, ở đúng một chỗ:

```go
// cmd/api/main.go     (🔜 chưa viết, đây là hình dạng)
func main() {
    cfg := config.Load()                              // đọc env

    pool, err := postgres.Connect(ctx, cfg.DSN)       // ① MỘT pool cho cả đời
    if err != nil { log.Fatal(err) }
    defer pool.Close()                                // chạy khi main() kết thúc

    // ② Ráp từ TRONG ra NGOÀI: adapter → app → http
    merchantRepo := postgres.NewMerchantRepo(pool)
    outbox       := postgres.NewOutbox(pool)
    clock        := platform.SystemClock{}

    registerMerchant := app.NewRegisterMerchantHandler(merchantRepo, outbox, clock)

    router := httpadapter.NewRouter(registerMerchant)

    srv := &http.Server{Addr: cfg.Addr, Handler: router}

    // ③ Chạy server trong goroutine riêng để main() còn chờ tín hiệu tắt
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // ④ Chờ Ctrl+C hoặc lệnh dừng từ hệ thống
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop                                            // ĐỨNG YÊN ở đây

    // ⑤ Tắt êm: ngừng nhận request mới, chờ request đang chạy xong
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}
```

| Ký hiệu mới | Nghĩa |
|---|---|
| `defer pool.Close()` | *"chạy dòng này khi hàm kết thúc"* — như `finally` |
| `go func() { ... }()` | chạy hàm này **song song**, không chờ |
| `make(chan os.Signal, 1)` | tạo *channel* — đường ống để goroutine nói chuyện |
| `<-stop` | **đọc từ channel; đứng chờ ở đây** cho tới khi có gì đó gửi vào |

> **Đây chính là `services.yaml` + `autowire` của Symfony, viết bằng tay.**
> Nghe cực hơn, nhưng đổi lại: sai là **không compile được**, không phải lỗi
> lúc chạy; và đọc `main()` là thấy toàn bộ hệ thống ráp thế nào, không phải mò
> qua nhiều file YAML.

### 9.4 Vòng đời một request — mỗi request là một goroutine

```
   http.ListenAndServe đang chờ...
              │
   request tới ─────▶ Go tự sinh MỘT GOROUTINE MỚI cho request này
                            │
                            ├─ router chọn handler
                            ├─ handler(w, r)
                            │     r.Context()  ← huỷ khi client ngắt kết nối
                            │     dùng CHUNG pool, logger, handler đã dựng ở main()
                            │
                            └─ ghi response → goroutine kết thúc
                                             │
                                  mọi thứ dựng ở main() VẪN CÒN NGUYÊN
```

Goroutine **không phải** thread hệ điều hành — nó nhẹ hơn hàng nghìn lần
(khoảng 2 KB lúc đầu). Chạy 100.000 goroutine cùng lúc là chuyện thường; 100.000
thread thì máy sập.

> Đây là lý do Go hợp với việc gọi OpenAI: một lời gọi mất 5–30 giây, nhưng
> goroutine nằm chờ gần như không tốn gì. PHP thì mỗi lời gọi giữ chặt một
> worker của PHP-FPM.

**Điều phải nhớ:** vì các goroutine **dùng chung** mọi thứ dựng ở `main()`, nếu
handler sửa một biến chung mà không khoá thì hai request có thể giẫm lên nhau —
gọi là **data race**. PHP không có vấn đề này vì các request không chia sẻ gì.

```bash
go test -race ./...     # trình dò race — trên Windows cần gcc, CI chạy hộ
```

### 9.5 `context.Context` — thứ PHP không có

`ctx` xuất hiện ở tham số đầu của gần như mọi hàm chạm I/O. Nó mang hai thứ:
**hạn chót** và **tín hiệu huỷ**.

```go
func (r *MerchantRepo) Save(ctx context.Context, m *catalog.Merchant) error
```

- Khách đóng tab → `ctx` bị huỷ → query SQL **dừng ngay**, không chạy phí
- Đặt timeout 5 giây → quá hạn thì mọi thứ bên dưới tự dừng

PHP không có khái niệm tương đương (`max_execution_time` chỉ giết cả script).

> **Quy tắc:** `ctx` luôn là tham số **đầu tiên**, tên luôn là `ctx`, **không**
> lưu vào struct. Và `internal/domain/` **không nhận `ctx`** — domain thuần
> tính toán, không chạm I/O. Chỉ repository và adapter mới cần.

### 9.6 Bản đồ: file nào chạy lúc nào

```
 LÚC BUILD          go build ./cmd/api    → api.exe
 ─────────          mọi file .go được biên dịch vào một file duy nhất

 LÚC KHỞI ĐỘNG      ① biến cấp package    shared.USD, shared.currencies
 (một lần)          ② func init()
                    ③ main()              dựng pool, handler, router
                                          → ListenAndServe → chờ

 MỖI REQUEST        adapter/http/         giải mã JSON, chuẩn hoá locale
 (goroutine riêng)  app/catalog/          mở transaction, lấy `now`
                    domain/catalog/  ✅   quy tắc nghiệp vụ (KHÔNG chạm I/O)
                    adapter/postgres/     SQL + ghi outbox, COMMIT

 CHẠY NỀN LIÊN TỤC  cmd/worker/           đọc outbox → publish
 (tiến trình khác)

 LÚC TẮT            srv.Shutdown()        ngừng nhận mới, chờ việc đang chạy
                    defer pool.Close()    đóng kết nối
```

### 9.7 Năm thứ người từ PHP hay vấp

| Vấp | Vì sao | Cách tránh |
|---|---|---|
| Sửa biến chung mà không khoá | goroutine dùng chung bộ nhớ | `sync.Mutex`, hoặc đừng dùng biến chung |
| Mở kết nối DB trong handler | phản xạ PHP | dựng pool **một lần** ở `main()` |
| Quên `defer` đóng tài nguyên | PHP tự dọn khi tiến trình chết | `defer` ngay sau khi mở |
| `panic` không bắt | giết **cả** tiến trình, không phải một request | middleware `recover()` ở tầng HTTP |
| Goroutine chạy mãi không thoát | rò rỉ tích luỹ theo ngày | luôn có đường thoát qua `ctx` |

### 9.8 So sánh nhanh vòng đời

| | Symfony | Portage (Go) |
|---|---|---|
| Điểm vào | `public/index.php` | `cmd/api/main.go` → `func main()` |
| Ráp phụ thuộc | `services.yaml`, autowire | viết tay trong `main()` |
| Vòng đời | mỗi request một lần | một lần cho cả tiến trình |
| Đồng thời | nhiều tiến trình PHP-FPM | nhiều goroutine trong **một** tiến trình |
| Huỷ giữa chừng | không có | `context.Context` |
| Chạy nền | Messenger + supervisor | `cmd/worker/` — cùng codebase, binary khác |
| Tắt êm | php-fpm reload | `srv.Shutdown(ctx)` |

---

## 10. Xoá comment `// [PHP]` khi không cần nữa

Xem còn bao nhiêu:

```bash
grep -rn "// \[PHP\]" internal/ | wc -l
```

Xoá toàn bộ (Git Bash):

```bash
find internal -name "*.go" -exec sed -i '/^\s*\/\/ \[PHP\]/d' {} +
gofmt -w . && go test ./...
```

> Chạy `go test ./...` sau khi xoá để chắc không lỡ tay cắt vào code.
> Comment không ảnh hưởng chương trình, nhưng lệnh `sed` thì có.

---

## 11. Generics — `func on[T any](…)` (Go 1.18+)

Trong repo có hai hàm generic, cả hai đều ở chỗ "bytes → kiểu": `eventcodec.into[T]` và
`wire.on[T]`.

```go
func into[T any](payload []byte) (any, error) {
	var v T                                  // T là kiểu được điền khi GỌI
	if err := json.Unmarshal(payload, &v); err != nil {
		return nil, err
	}
	return v, nil
}

var decoders = map[string]func([]byte) (any, error){
	"catalog.product_published": into[contracts.ProductPublishedV1],   // ← "khởi tạo" T = ProductPublishedV1
	"catalog.product_measured":  into[contracts.ProductMeasuredV1],
}
```

`into[contracts.ProductPublishedV1]` **không gọi** hàm — nó tạo ra *một hàm mới* với `T` đã
điền, rồi bỏ vào map. Bên PHP không có tương đương trực tiếp; gần nhất là một closure sinh
từ template, hoặc `/** @template T */` của Psalm/PHPStan — nhưng đó chỉ là chú thích cho
tool, còn ở Go compiler **sinh code thật** và kiểm kiểu thật.

```go
func on[T any](handle func(context.Context, T) error) worker.Handler {
	return func(ctx context.Context, e worker.Entry) error {
		msg, err := eventcodec.Decode(e.Name, e.Payload)   // trả `any`
		m, ok := msg.(T)                                   // type assertion về T
		if !ok { return fmt.Errorf("%s decodes to %T, handler wants %T", e.Name, msg, *new(T)) }
		return handle(ctx, m)
	}
}
bus.Subscribe("catalog.product_published", on(projector.OnProductPublished))   // T suy ra từ kiểu tham số của OnProductPublished
```

Ở dòng cuối không cần viết `on[contracts.ProductPublishedV1]` — compiler **suy ra** `T` từ
chữ ký của `projector.OnProductPublished`. Đây là cách thay cho "argument resolver" của
Symfony Messenger (nó nhìn type-hint của `__invoke` để biết deserialize vào class nào):
ở đây type-hint là `T`, và việc "nhìn" xảy ra lúc compile.

| Ký hiệu | Đọc là |
|---|---|
| `[T any]` | tham số kiểu tên `T`, chấp nhận mọi kiểu (`any` = `interface{}`) |
| `into[Foo]` | bản của `into` với `T = Foo` — một *giá trị* kiểu hàm, chưa gọi |
| `var v T` | biến có kiểu `T`, zero value của kiểu đó |
| `*new(T)` | cách viết "một giá trị zero kiểu T" trong biểu thức (chỉ dùng để in `%T`) |
| `msg.(T)` | type assertion: `msg` (kiểu `any`) có phải `T` không |

**Khi nào dùng:** khi cùng một thuật toán áp cho nhiều kiểu **và** kiểu đó phải được kiểm
lúc compile. Khi chỉ cần "một hàm cho nhiều kiểu mà không quan tâm kiểu là gì", `any` +
type switch (như `eventcodec.Encode`) đủ và dễ đọc hơn. Repo cố tình dùng cả hai để bạn
thấy sự khác biệt.
