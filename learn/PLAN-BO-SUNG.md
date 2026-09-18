# Plan bổ sung Mùa 1 — đối chiếu với roadmap.sh Go

Đầu vào: `roadmap.sh — Learn to become a Go developer` (bản PDF 06/09/2025, 84 mục).
Đầu ra mong muốn: biết **còn thiếu gì** sau 10 tập Mùa 1, và quyết định thiếu chỗ
nào thì vá, chỗ nào cố ý bỏ.

Tài liệu này là **plan**, chưa phải việc đã làm. Chốt xong mới dựng.

---

## 1. Cách đối chiếu, và chỗ nó không đáng tin

Không đọc bằng trí nhớ. Ba bước máy làm:

1. Rút 84 mục của roadmap từ text layer của PDF.
2. Grep từng mục vào **toàn bộ nội dung lên hình** của Mùa 1 (mọi chuỗi trong
   `src/videos/go/**`, cộng tên cảnh trong `spec.ts`).
3. Tách hai mức: xuất hiện trong `Headline`/`TermCard`/tên cảnh = **dạy hẳn**;
   chỉ xuất hiện trong bullet hoặc code = **nhắc qua**.

**Kết quả thô: 84 mục · chạm tới 38 · chưa chạm 46.**

**Chỗ phép đo này sai, nói trước:** heuristic "dạy hẳn" có âm tính giả. Ví dụ
`sync.Mutex` bị xếp "nhắc qua" vì nó được dạy bằng bullet và bảng so sánh chứ
không bằng `TermCard` — nhưng gần nửa tập 10 nói về nó. Nên con số 38 là sàn,
không phải số chính xác. Mọi quyết định dưới đây tôi đã mở file kiểm lại bằng
mắt, không dựa vào con số.

Lệnh dựng lại phép đo: `/tmp/gap.py` và `/tmp/depth.py` (tạm; nếu chốt plan thì
chuyển thành `scripts/roadmap-gap.py` để chạy lại được sau này).

---

## 2. Một lỗi phải sửa trước, không đợi tập mới

**Tập 10 đang nói sai trên hình.** Nó tuyên bố repo có **0 channel**, và bản in
1.10 in con số 0 to đùng. Đếm lại:

| | Số | Ở đâu |
|---|---|---|
| Khai báo kiểu `chan`, code chạy thật | **0** | — |
| **Đọc** từ channel (`<-`), code chạy thật | **4** | `worker/worker.go:135,137` · `worker/sweep.go:61,63` |
| Khai báo `chan` kể cả test | **5** | `postgres_test.go:106` — test đua migrate, đợt 18 |

Cả bốn chỗ đọc đều nằm trong khối `select`:

```go
select {
case <-ctx.Done():
    return ctx.Err()
case <-ticker.C:
}
```

"0 channel" đúng theo nghĩa *không tự tạo channel nào*, nhưng người học mở
`worker.go` ngày đầu tiên sẽ thấy `case <-ctx.Done():` và kết luận là video nói
dối. Giá của việc nói chính xác là một câu.

**Sửa:** đổi `GO_USAGE.channels` thành hai con số (`channelsDeclared: 0`,
`channelsRead: 4`), sửa lời tập 10 và bản in 1.10, thêm `channelsRead` vào
`verify-numbers.sh`. Rồi render lại tập 10.

Đây cũng là bằng chứng cho thấy cửa `verify-numbers.sh` chưa đủ: nó đếm đúng thứ
tôi bảo nó đếm, mà thứ tôi bảo nó đếm lại là thứ sai.

---

## 3. Tự phản biện trước khi đề xuất

Đề xuất hiển nhiên là "làm thêm ~15 tập phủ hết 46 lỗ hổng". Tôi phản biện chính
mình trước, vì ba lý do dưới đây đủ mạnh để bác nó:

**Phản biện 1 — phủ hết nghĩa là phải bịa code.** Khoảng 30 trong 46 mục không
có trong repo: gRPC, GORM, Cobra, bubbletea, Melody, CGO, plugin, compiler flag.
Muốn dựng tập cho chúng thì phải viết ví dụ tự nghĩ ra. Mà thứ khiến Mùa 1 khác
một video YouTube chính là **mọi dòng code lên hình đều mở được trong repo của
anh** — có cả `verify-snippets.py` canh điều đó. Đánh đổi cái đó lấy độ phủ là
đánh đổi sai.

**Phản biện 2 — với những mục bịa được, Go by Example tốt hơn tôi.** Một tập
video ~4 giờ dựng để dạy `for range` thì thua hai mươi dòng trên
gobyexample.com, đọc trong hai phút và tra lại được. Video chỉ thắng khi cái cần
dạy là **vì sao repo này làm vậy** — thứ không có trên internet.

**Phản biện 3 — "còn lỗ hổng" không đồng nghĩa với "cần video".** Danh sách lệnh
toolchain là tài liệu tra cứu, không phải chuyện kể. Nhét nó vào video là chọn
sai phương tiện.

**Kết luận của phản biện:** không phủ hết. Chia bốn nhóm, mỗi nhóm một phương
tiện, và **ghi rõ nhóm bỏ hẳn kèm lý do** để sau này không ai tưởng là bỏ sót.

---

## 4. Bốn nhóm quyết định

| Nhóm | Phương tiện | Tiêu chí |
|---|---|---|
| A | **Tập video mới** | có vật liệu thật trong repo **và** phỏng vấn hay hỏi |
| B | **Bản in, không video** | là tài liệu tra cứu, không phải chuyện kể |
| C | **Bài tập trên repo thật** | phải tự tay chạy mới hiểu, xem không hiểu được |
| D | **Bỏ hẳn** | không có trong repo, và phỏng vấn backend ít hỏi |

---

## 5. Nhóm A — sáu tập mới

Mỗi tập vẫn giữ luật của Mùa 1: **mục tiêu đúng một điều**, neo vào code thật có
đường dẫn và số dòng, kết bằng màn nhớ lại + hộp phỏng vấn.

### T11 · Closure: hàm nhận hàm làm tham số

- **Neo:** `UnitOfWork.InTx(ctx, func(ctx) error { … })` — **37 chỗ gọi** trong
  `internal/app/`. Đây là hình dạng của *mọi* use case trong repo.
- **Dạy:** anonymous function · closure bắt biến ngoài · vì sao `InTx` nhận hàm
  thay vì trả transaction · named return `(err error)` cho phép `defer` ghi đè.
- **Roadmap phủ:** Closures · Anonymous Functions · Named Return Values.
- **Vì sao trước tiên:** không đọc được closure thì không đọc được use case nào,
  và hai tập sau đều dựng trên nó.

### T12 · interface rỗng và type switch

- **Neo:** `eventcodec/codec.go:71` — `switch e := ev.(type)` với 37 loại event.
  Cộng 17 chỗ dùng `any`/`interface{}`.
- **Dạy:** vì sao `shared.Event` là interface rỗng · type switch · type assertion
  với `ok` · cái giá: mất kiểm tra lúc biên dịch, đổi lấy một chỗ mã hoá duy nhất.
- **Roadmap phủ:** Empty Interfaces · Type Assertions · Type Switch.

### T13 · Generics: hai chỗ duy nhất trong repo

- **Neo:** `wire/subscribe.go:105` — `func on[T any](handle func(ctx, T) error)`,
  và `eventcodec/decode.go:66` — `func into[T any](payload []byte)`.
- **Dạy:** vì sao cần · type inference · type constraint · và câu trả lời phỏng
  vấn: `on[T]` biến 35 handler có kiểu thành 35 listener không kiểu; không có
  generics thì phải viết 35 adapter gần giống hệt nhau.
- **Roadmap phủ:** cả nhánh Generics (why · functions · types · constraints ·
  inference).
- **Ghi chú:** repo chỉ có **2** chỗ dùng generics. Tập này phải nói thẳng điều
  đó — generics là công cụ hiếm dùng, không phải thứ rắc vào mọi nơi.

### T14 · select, ticker, và dừng một việc đang chạy

- **Neo:** `worker/worker.go:134` và `worker/sweep.go:60` — bốn chỗ đọc channel
  thật. Cộng `postgres_test.go:105` dùng `sync.WaitGroup` + `make(chan struct{})`
  để bắn bốn goroutine cùng lúc (test đua migrate).
- **Dạy:** `select` · `ctx.Done()` là một channel · `time.Ticker` ·
  `sync.WaitGroup` · buffered vs unbuffered.
- **Ngoài Portage, dán nhãn rõ:** fan-in, fan-out, pipeline — repo không có,
  nhưng phỏng vấn hỏi. Làm đúng cách tập 10 đã làm với channel.
- **Roadmap phủ:** Select Statement · WaitGroups · Buffered vs Unbuffered ·
  Concurrency Patterns.
- **Kèm theo:** tập này là chỗ sửa lời tuyên bố "0 channel" của tập 10.

### T15 · struct tag, JSON, và đường ra khỏi domain

- **Neo:** **475 struct tag `json:"…"`**, `encoding/json` trong 21 file,
  `internal/contracts/*V1`.
- **Dạy:** struct tag là gì · vì sao domain không có tag nào mà contract thì có ·
  Published Language · `encoding/json` dùng reflection nên tag mới có tác dụng ·
  vì sao tên trường JSON không được đổi tuỳ tiện (đánh version `V1`).
- **Roadmap phủ:** Struct Tags & JSON · Structs · chạm tới Reflection.

### T16 · Vì sao 278 test chạy dưới hai giây

- **Neo:** 278 test không cần Docker · 7 test canh kiến trúc bằng `go/ast` ·
  `httptest` ở 3 file · `t.Run` ở 4 chỗ · adapter bộ nhớ.
- **Dạy:** port/adapter làm test nhanh · subtest `t.Run` · `httptest` ·
  `-cover` · fake thật (adapter bộ nhớ) so với mock sinh máy.
- **Nói thẳng một sự thật khó chịu:** repo **không** có test table-driven nào
  (`for _, tc := range` = 0), và **không** có `Benchmark` nào. Cả hai đều là thứ
  phỏng vấn hỏi. Nên dạy chúng ở mức "ngoài Portage", và nếu anh muốn thì đây là
  gợi ý sửa repo chứ không phải sửa video.
- **Roadmap phủ:** testing package · Mocks and Stubs · httptest · Coverage ·
  (Benchmarks ở mức ngoài Portage).

**Thứ tự đề xuất:** T11 → T12 → T13 → T14 → T15 → T16.
Lý do: T13 cần T11 và T12 (vì `on[T]` là generic **nhận một closure** và trả về
một listener không kiểu). Bốn tập còn lại độc lập với nhau.

**Nếu chỉ làm được một tập:** làm **T16**. Nó là thứ phỏng vấn hỏi nhiều nhất
trong sáu tập, và nó độc lập hoàn toàn.

---

## 6. Nhóm B — hai bản in, không làm video

### Bản in "Cú pháp Go còn lại"

Những mục nhỏ, tra cứu là đủ, kể thành chuyện thì nhạt:
`const`/`iota` · scope & shadowing · rune và string literal (thô vs diễn giải) ·
ép kiểu · `for` / `for range` · `if` / `switch` · `break` / `continue` /
`goto` · variadic · chuyển đổi slice ↔ array.

Neo được vào repo ở vài chỗ (`iota` ×1, `rune` ×5), còn lại là cú pháp thuần.

### Bản in "Toolchain và standard library"

`go run/build/install/test/fmt/vet/doc/version` · `go mod init/tidy/vendor` ·
dùng thư viện bên thứ ba · `golangci-lint`, `staticcheck`, `revive` ·
`govulncheck` · `go generate` và build tag · biên dịch chéo · và bảng stdlib:
`io` · `flag` · `time` · `encoding/json` · `os` · `bufio` · `slog` · `regexp` ·
`go:embed` (repo có **1** chỗ dùng).

Hai bản in này dùng đúng khuôn `plate.css` và `scripts/render-plates.sh` sẵn có.

---

## 7. Nhóm C — ba bài tập trên repo thật

Xem không hiểu được, phải tự chạy mới thấy. Viết thành một mục trong
`docs/HOC.md` chứ không dựng video:

1. **Escape analysis trên code thật.** `go build -gcflags='-m' ./internal/domain/shared`
   rồi đọc xem `Money` nào nằm stack, cái nào thoát ra heap. Phủ: Memory
   Management · Escape Analysis — hai mục phỏng vấn hay hỏi mà không có cách nào
   dạy bằng hình.
2. **Đọc một stack trace thật.** Cố tình gọi `NewMoney(5, shared.Currency{})` ở
   một test, đọc panic từ trên xuống. Phủ: Stack Traces & Debugging.
3. **Cắm pprof vào `cmd/api` rồi đo.** Repo hiện **không** có pprof (0 chỗ). Đây
   là bài tập thêm code, không phải bài đọc code. Phủ: pprof · trace.

---

## 8. Nhóm D — bỏ hẳn, và vì sao

Ghi ra để sau này là **quyết định**, không phải bỏ sót:

| Mục | Vì sao bỏ |
|---|---|
| History of Go · Setting up · Hello World | anh đang chạy repo này hằng ngày rồi |
| Building CLIs (Cobra, urfave/cli, bubbletea) | repo không có CLI framework nào; `cmd/api` dùng `flag` của stdlib |
| Web frameworks (gin, echo, fiber, beego) | repo dùng `net/http` thuần — và đó là một quyết định, không phải thiếu sót |
| ORMs (GORM) | repo dùng `pgx` trực tiếp và viết mapping tay, có lý do trong `postgres.go` |
| Logging (Zerolog, Zap) | repo dùng `log` của stdlib |
| gRPC & Protocol Buffers | không có, và một hệ mua hộ một người dùng chưa cần |
| Realtime (Melody, Centrifugo) | không có |
| CGO · Unsafe · Plugins · Compiler & Linker Flags | không có, và phỏng vấn backend gần như không hỏi |
| Reflection (chuyên sâu) | chỉ chạm tới ở T15 qua `encoding/json`; đi sâu là chủ đề của người viết thư viện |
| Publishing Modules | repo là ứng dụng, không phải thư viện |

Nếu sau này anh đi phỏng vấn chỗ nào hỏi mấy thứ này, mở lại bảng và làm thêm —
nhưng lúc đó là quyết định mới, có thông tin mới.

---

## 9. Độ phủ sau khi làm xong

| | Mục | Ghi chú |
|---|---|---|
| Mùa 1 hiện tại chạm tới | 38 | con số sàn, xem §1 |
| Nhóm A thêm (6 tập) | ~13 | closure, generics, type switch, select, JSON, testing |
| Nhóm B thêm (2 bản in) | ~16 | cú pháp còn lại + toolchain/stdlib |
| Nhóm C thêm (3 bài tập) | ~3 | escape analysis, stack trace, pprof |
| Nhóm D bỏ hẳn | ~14 | có lý do ở §8 |

Con số "~" là cố ý: vài mục nằm vắt giữa hai nhóm (ví dụ `go:embed` vừa là stdlib
vừa là build tooling). Không làm tròn cho đẹp.

---

## 10. Việc phải làm, theo thứ tự

1. **Sửa lời "0 channel" của tập 10** (§2) — làm trước, vì đó là lỗi đang nằm
   trên hình, không phải thiếu sót.
2. Chuyển `/tmp/gap.py` thành `scripts/roadmap-gap.py` để chạy lại được, và
   thêm `channelsRead` vào `verify-numbers.sh`.
3. Dựng T11 → T16, mỗi tập đủ bốn cửa kiểm ở `CURRICULUM.md` §12.
4. Hai bản in nhóm B.
5. Ba bài tập nhóm C, viết vào `docs/HOC.md`.
6. Cập nhật `CURRICULUM.md`: Mùa 1 từ 10 lên 16 tập, và thêm mục "đối chiếu với
   roadmap.sh" trỏ về file này.

**Ước lượng:** sáu tập video là phần nặng nhất. Hai bản in và ba bài tập nhẹ hơn
nhiều vì đã có sẵn khuôn.

---

## 11. Thứ plan này cố ý không làm

- **Không** đổi cấu trúc ba mùa. Mùa 2 (DDD) và Mùa 3 (Portage) giữ nguyên.
- **Không** dựng lại 10 tập cũ theo roadmap. Chúng đang đúng và đang chạy được;
  chỉ tập 10 có một câu sai và đã ghi ở §2.
- **Không** đề xuất bỏ ràng buộc "mọi dòng code lên hình phải có thật trong repo".
  Đúng ràng buộc đó mới là lý do loạt này đáng làm.
