# Lộ trình học bằng video — từ Symfony sang Go

Tài liệu này là **spec** của loạt video. Nó trả lời: tập nào dạy gì, theo thứ tự
nào, và vì sao thứ tự đó.

**Mục tiêu**: đi từ sáu năm PHP/Symfony sang Go nhanh nhất có thể, với nền đủ
vững để đi phỏng vấn, rồi mới tới DDD và kiến trúc thật của Portage.

Nguồn sự thật là **source code Go**, không phải file Markdown. Mọi con số và mọi
đoạn code lên hình đều nằm trong `src/data/portage.ts`, mỗi mảnh kèm đường dẫn và
số dòng. Cảnh không được phép tự viết ra một sự thật về codebase.

**Và phải đếm lại, không được tin bản kế hoạch cũ.** Ngày 17/09 đếm lại toàn bộ
`GO_USAGE` bằng grep trên repo: 8/10 con số đúng, 2 con số sai và đã sửa —
`assertions` 53 → **48** (bản cũ đếm nhầm), `defers` 96 → **95** (đếm cả một chữ
`defer` nằm trong comment). Cả hai đều chưa kịp lên hình. Lệnh đếm ghi ngay
trong `data/portage.ts` cạnh từng con số, để lần sau đếm lại được.

---

## 1. Ba mùa, và vì sao theo thứ tự đó

Bản kế hoạch đầu tiên sai ở đúng chỗ này, nên nói rõ để không lặp lại: nó dựng đồ
thị phụ thuộc **bên trong kiến trúc** rồi bắt đầu từ gốc của đồ thị đó. Nhưng mọi
tập đều chiếu code Go, mà người học chưa đọc được cú pháp Go. Dạy
`type CustomerOrder struct { shared.Events }` cho người chưa biết struct và
embedding là dạy trên nền trống.

Thứ tự đúng:

```
   MÙA 1 — GO           đọc được mọi file trong repo, và trả lời được phỏng vấn
        │                về ngôn ngữ. 10 tập.
        ▼
   MÙA 2 — DDD          hiểu từ vựng và lý lẽ, bằng chính ví dụ Symfony đã quen.
        │                6 tập.
        ▼
   MÙA 3 — PORTAGE      vì sao project THẬT này trông như vậy. 8 tập.
```

Mùa 1 là ưu tiên. Mùa 3 chỉ có nghĩa sau khi mùa 1 xong.

---

## 2. MÙA 1 — Go cho người viết PHP

Mục tiêu từng tập là **một** điều, và mỗi tập neo vào code thật trong repo.

| # | Tên | Học được gì | Neo vào code thật | |
|---|---|---|---|---|
| 1 | PHP chết sau mỗi request. Go thì không. | vòng đời tiến trình, goroutine, vì sao có pool, mutex, ctx | `cmd/api/main.go` · `adapter/memory` | **xong** |
| 2 | Đọc một dòng Go | kiểu sau tên, HOA/thường thay public/private, struct không chứa method | `shared/money.go` | **xong** |
| 3 | Lỗi là giá trị trả về | `(T, error)`, sentinel, `%w`, `errors.Is`, không có try/catch | `ParseMoney` · 456 chỗ `%w` | **xong** |
| 4 | Receiver: `(m Money)` và `(e *Events)` | copy hay sửa bản gốc, vì sao value object bất biến miễn phí | `money.go` vs `event.go` | **xong** |
| 5 | Interface ngầm: Go không có `implements` | khai báo ở nơi cần, `var _ X = (*Y)(nil)` | `app/ports.go` · 48 chỗ khẳng định | **xong** |
| 6 | Zero value: Go không có null | mỗi kiểu một giá trị mặc định, constructor phải tự vệ | `NewMoney` panic · `IsZero()` | **xong** |
| 7 | Embedding: giống kế thừa nhưng không phải | `struct{ shared.ID }`, nhúng `shared.Events` | `id.go` · 8 chỗ nhúng | **xong** |
| 8 | slice, map và ba cái bẫy | `append` phải gán lại, map duyệt ngẫu nhiên, so sánh struct | bug thật: thứ tự `CategoryRepo.All()` | **xong** |
| 9 | defer, panic, recover | `defer` chạy khi nào, panic giết cả process | 95 `defer` · 33 `panic` | **xong** |
| 10 | Đồng thời: goroutine, mutex, channel | repo có 3 goroutine và 25 mutex, **không có channel** — nói thẳng, rồi dạy channel riêng vì phỏng vấn sẽ hỏi | `cmd/api` · `adapter/memory` | **xong** |

Quy tắc trung thực của mùa 1: thứ gì **không có** trong repo thì nói rõ là không
có, rồi dạy riêng. Channel và `sync.WaitGroup` rơi vào nhóm đó.

---

## 3. MÙA 2 — DDD cho người viết Symfony

| # | Tên | Học được gì | |
|---|---|---|---|
| 1 | Mô hình thiếu máu | entity toàn setter, và vì sao luật nghiệp vụ rơi mất | **xong** |
| 2 | Value Object và Entity | cái gì có ID, cái gì không | **xong** |
| 3 | Aggregate và invariant | một transaction một aggregate | **xong** |
| 4 | Repository là cổng | đảo chiều phụ thuộc, domain không biết Postgres tồn tại | **xong** |
| 5 | Bounded Context | một từ, một nghĩa, trong một vùng | **xong** |
| 6 | Domain Event | aggregate chỉ ghi, tầng app mới phát | **xong** |

---

## 4. MÙA 3 — Portage: kiến trúc thật

Đây là chỗ đồ thị phụ thuộc kiến trúc mới có nghĩa.

```
   NGHIỆP VỤ: một đơn sống 30 ngày, 2 quốc gia, 3 bên ngoài không kiểm soát
        │  ⇒ không transaction nào giữ nổi
   AGGREGATE TÁCH RA  ⇒  BOUNDED CONTEXT  ⇒  DOMAIN EVENT  ⇒  OUTBOX
        │                                                        │
        │                          ├──▶ at-least-once ⇒ CONSUMER IDEMPOTENT
        │                          └──▶ bất đồng bộ   ⇒ NHẤT QUÁN SAU CÙNG
        │                                                        │
        │                                     ⇒ màn hình cần 4 context, đừng JOIN
        ▼                                            READ MODEL (CQRS)
   Hai nhánh song song:
   tiền qua 2 tiền tệ và 4 tuần    ⇒ QUOTE LÀ ẢNH CHỤP ⇒ QUOTE vs ACTUAL
   phụ thuộc bên thứ ba            ⇒ ANTI-CORRUPTION LAYER
```

| # | Tên | Mục tiêu | |
|---|---|---|---|
| 1 | Vì sao project này bị chia làm năm | năm vùng, và vì sao không gộp lại được | **xong** |
| 2 | Outbox: hai việc, một transaction | vì sao không "lưu xong rồi publish" | **xong** |
| 3 | Nhất quán sau cùng có mã trạng thái | 404 ngay sau publish là thiết kế | **xong** |
| 4 | Port & Adapter, và luật chiều phụ thuộc | vì sao 278 test chạy 2 giây | **xong** |
| 5 | Điểm không thể quay đầu | luật nghiệp vụ có tiền thật đằng sau | **xong** |
| 6 | Báo giá là một bức ảnh, và đối soát | `variance` trả lời câu hỏi kinh doanh | **xong** |
| 7 | Màn hình cần bốn context: đừng JOIN | read model dựng bằng event | **xong** |
| 8 | Máy được tạo nháp, không được xác nhận | hai anti-corruption layer | **xong** |

---

## 5. Mô hình trí nhớ dùng chung

Xuất hiện ở cuối **mọi** tập mùa 3, mỗi lần sáng lên đúng phần vừa học.

```
 ┌─CATALOG─┐ ┌─PRICING─┐ ┌ORDERING─┐ ┌PROCUREM.┐ ┌LOGISTICS┐
 └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
 ═════╧═══════════╧══════ OUTBOX ═════════╧═══════════╧══════
        mọi việc đã xảy ra rơi xuống đây, ai cần thì nhặt lên
 ────────────────────── REPORTING ───────────────────────────
        màn hình cho khách và nhân viên, dựng chỉ bằng event
```

**Không mũi tên nào nối thẳng hai context.** Đó là bài học, không phải thiếu sót
của hình vẽ.

---

## 6. Kiến trúc hệ thống video

```
learn/src/
├── data/portage.ts      SỰ THẬT rút từ repo Go. Sửa số ở ĐÂY, không sửa trong cảnh.
├── design/tokens.ts     màu, thang chữ, nhịp. Đổi một dòng là đổi cả loạt video.
├── design/fonts.ts      Be Vietnam Pro (dấu tiếng Việt) + JetBrains Mono (code)
├── design/motion.ts     một đường cong easing cho cả loạt
├── components/          từ vựng hình ảnh dùng chung:
│   ├── Stage.tsx          khung, nền, vùng an toàn, Split hai cột chỉnh được
│   ├── Text.tsx           Headline · Body · Callout
│   ├── Diagram.tsx        BỘ VẼ LUỒNG: hộp, hộp-ranh-giới, mũi tên tự vẽ, packet bay
│   ├── TermCard.tsx       màn hình "từ mới": nghĩa đen · bên Symfony · ở đây
│   ├── CompareTable.tsx   bảng PHP ↔ Go — xương sống của mùa 1
│   ├── Bullets.tsx        cột giải thích, từng câu hiện theo nhịp
│   ├── CodeCard.tsx       code THẬT, kèm path:line
│   ├── SketchCard.tsx     thiết kế project ĐÃ TỪ CHỐI, viền đứt, có cảnh báo
│   ├── ContextMap.tsx     mô hình trí nhớ dùng chung
│   ├── TimeAxis.tsx       trục thời gian, mốc đặt đúng tỷ lệ ngày thật
│   ├── Timeline.tsx       khung BEGIN…COMMIT gãy được
│   └── Sfx.tsx            ÂM THANH: tám tín hiệu, mỗi cái một nghĩa cố định
├── videos/registry.ts   KIỂU của một tập: id, scenePrefix, outDir, danh sách cảnh
├── videos/index.ts      DANH SÁCH cả loạt. Root.tsx và các script đều đọc từ đây.
├── videos/go/ep01…ep02/ MÙA 1
└── videos/portage/ep01/ MÙA 3
```

**Thêm một tập = thêm một `spec.ts`, vài cảnh, và một dòng trong `videos/index.ts`.**
Không viết lại animation, không sửa `Root.tsx`, không sửa script nào — `Root.tsx`
duyệt `SEASONS`, còn `scripts/episodes.py` tự đọc mọi `spec.ts` để lấy `scenePrefix`,
`outDir` và độ dài từng cảnh.

Thành phần dùng chung thêm trong đợt này:

| | Dùng để |
|---|---|
| `components/Anatomy.tsx` | mổ một dòng code thành từng mảnh, mỗi mảnh một đường dẫn xuống chú thích. Vị trí tính bằng đếm ký tự — font đều nét nên chính xác tuyệt đối, không cần đo DOM |
| `components/Recap.tsx` | màn "nhớ lại" cuối mỗi tập: bốn câu, hộp phỏng vấn, hộp tập sau |
| `cheatsheet/plate.css` | ngôn ngữ thị giác chung của mọi bản in. Mỗi bản chỉ giữ nhịp riêng và đặt `--accent` cho từng thẻ |

Ba quy tắc giữ cho loạt video không nói dối:

1. Code lên hình phải có **đường dẫn và số dòng**. Không có địa chỉ thì người học
   sẽ không đi tìm, và video dạy đoạn code thay vì dạy codebase.
2. Thiết kế mà project **đã từ chối** phải dùng `SketchCard`, viền đứt, kèm dòng
   chữ *"không phải code của Portage"*.
3. Thứ gì **không có** trong repo thì nói rõ là không có trước khi dạy nó.

---

## 7. Cặp học: mỗi tập là một video VÀ một bản tĩnh

Mỗi tập xuất ra **hai** thứ, và chúng là một cặp, không phải bản chính và bản
phụ:

| | Tối ưu cho | Dùng khi |
|---|---|---|
| `video.mp4` | giải thích tuần tự, animation dẫn mắt | lần đầu học một khái niệm |
| `cheatsheet.png` + `.pdf` | nhớ lại nhanh, quét trong 20–30 giây | ôn lại, trước phỏng vấn, dán lên tường |

Bản tĩnh **không phải** một khung cắt từ video, và cũng **không phải** kịch bản
đổ thành tài liệu. Nó là một bố cục riêng, dựng theo triết lý
[`design/afterimage.md`](design/afterimage.md): mặt phẳng được đối xử như một
bản in khoa học — chia ô bằng đường kẻ, mỗi ô một tiêu bản, nhãn đánh số, chú
thích lâm sàng ở lề.

Ràng buộc để cặp này thật sự là một cặp:

1. **Cùng từ vựng.** Từ nào xuất hiện trên bản tĩnh phải đã được định nghĩa
   trong video, bằng đúng câu đó.
2. **Cùng ngôn ngữ thị giác.** Cùng bảng màu (`design/tokens.ts`), cùng màu cho
   mỗi bounded context, cùng cặp font: JetBrains Mono cho tiếng của hệ thống,
   một sans humanist cho tiếng của người.
3. **Cùng ẩn dụ.** Dòng sông outbox, hộp ranh giới, packet mang nhãn — bản tĩnh
   dùng lại đúng những hình đó, không phát minh hình mới.
4. **Cùng con số.** Cả hai đọc từ `src/data/portage.ts`.

Bản tĩnh đánh số theo **mùa · tập** (`Plate 1.02`), không đánh số La Mã tuần tự —
24 bản mà đánh tuần tự thì chèn một tập vào giữa là phải đánh số lại cả loạt.

Nguồn bản tĩnh nằm ở `cheatsheet/*.html`, render bằng Chrome headless:

```bash
cd learn
./scripts/render-plates.sh         # mọi bản tĩnh → PNG @2x + PDF A3 ngang
```

Diagram vẽ bằng SVG toạ độ tuyệt đối, không phải layout tự động — vì yêu cầu
cao nhất của bản in là **không chữ nào đè chữ nào**, và toạ độ tuyệt đối là cách
duy nhất bảo đảm điều đó.

---

## 8. Âm thanh

Không có lời thoại. Chỉ tám tín hiệu, mỗi cái mang đúng một nghĩa ở mọi tập:

| Tín hiệu | Nghĩa | Âm |
|---|---|---|
| `appear` | một hộp, một dòng, một bước hiện ra | mouse-click, rất khẽ |
| `reveal` | một ý mới được giới thiệu | switch |
| `travel` | có thứ gì đó đang đi: packet trên dây | whoosh |
| `snap` | cắt nhanh, đổi chủ đề | whip |
| `land` | một điểm chốt lại, test xanh | ding |
| `crack` | có cái gì đó vỡ — tối đa một lần mỗi tập | bone-crack |
| `waiting` | thời gian trôi, phải chờ | loading-lag |
| `turn` | sang cảnh mới | page-turn |

Âm lượng đặt thấp có chủ đích. Video học được xem bằng mắt; tiếng chỉ để đánh dấu
**khi nào** có gì xảy ra, để mắt biết nhìn vào đâu. Bộ gói còn có các âm meme
(vine-boom, bruh…) và chúng bị loại hết: một câu đùa giữa lời giải thích lấy mất
mạch người học đang giữ.

Tệp `.wav` tải sẵn về `public/sfx/`, nên render không cần mạng.

**Mọi tập phải có tín hiệu.** Mùa 3 tập 1 dựng trước khi có `Sfx.tsx` nên lỡ ra
lò không có track âm thanh nào — phát hiện bằng cách soi lại file mp4 (`ffmpeg -i`
đếm `Audio: aac`), không phải bằng cách nhớ. Đã bổ sung 35 tín hiệu cho 11 cảnh
của nó, đặt đúng lúc có việc xảy ra (hộp hiện, packet bay, chỗ vỡ), không rải đều.
Cách kiểm nhanh:

```bash
for f in out/*/*/video.mp4; do
  echo "$f  $(node_modules/@remotion/compositor-linux-x64-gnu/ffmpeg -i "$f" 2>&1 | grep -c 'Audio: aac')"
done   # mỗi dòng phải kết thúc bằng 1
```

---

## 9. Khổ hình

Ngang **1920×1080**, 30fps, chữ tiếng Việt trên hình.

Đây là **bài học xem trên laptop cạnh repo đang mở**. Khổ ngang mua được thứ quan
trọng nhất: diagram và lời giải thích về nó nằm cạnh nhau, cùng lúc.

| Thứ | Giá trị |
|---|---|
| Lề an toàn hai bên | 110px |
| Lề trên và dưới | 76px |
| Tiêu đề | 56–68px |
| Chữ giải thích | 36px |
| Code | 27px |

### Cách viết câu tiếng Việt cho người học

Đây là chỗ đã sai và bị chỉ ra, nên ghi lại thành luật.

**Ba tật phải tránh**, cả ba đều làm câu đọc như dịch máy:

| Tật | Ví dụ đã viết sai | Sửa thành |
|---|---|---|
| Mở câu bằng `Cái ...` rồi gán ẩn dụ | *"Cái còi báo 'việc này thôi đi'."* | *"ctx mang lệnh dừng và hạn chót đi xuống mọi hàm bên dưới nó."* |
| Câu thứ hai rụng chủ ngữ | *"Mang theo hạn chót và tín hiệu huỷ."* | gộp vào câu trước, hoặc nêu rõ chủ ngữ |
| Nhân hoá máy móc | *"database hứa"* | *"Database bảo đảm"* |

**Bốn điều nên làm:**

1. **Dùng chính tên định danh làm chủ ngữ.** `ctx`, `Database`, `Go` — người
   đọc code sẽ gặp lại đúng chữ đó.
2. **Ẩn dụ chỉ dùng khi nó chính xác.** *"một cửa duy nhất"* cho aggregate thì
   đúng, vì thật sự chỉ có một đường vào. *"cái còi"* cho ctx thì không, vì ctx
   không phát ra gì cả, nó được truyền xuống.
3. **Không viết HOA để nhấn.** Chữ HOA giữa câu đọc như slide hội thảo. Nhấn
   bằng vị trí và màu, việc đó thuộc phần thiết kế.
4. **Đọc to lên.** Câu nào không nói ra miệng được thì viết lại.

**Một lượt rà không đủ.** Sửa xong tám câu định nghĩa rồi render lại, soi khung
1:04 thì thấy ngay dưới câu vừa sửa còn một câu cùng tật: *"Muốn hứa được vậy,
database phải giữ được cả hai việc trong tay cùng một lúc."* Lượt thứ hai grep
theo **tật** chứ không theo câu (`hứa`, `trong tay`, `Cái `) ra thêm ba chỗ:

| Đã viết | Sửa thành | Tật |
|---|---|---|
| *"Muốn hứa được vậy, database phải giữ được cả hai việc trong tay."* | *"Hai lệnh ghi phải đi qua cùng một transaction. Portage truyền nó trong ctx, mỗi repository tự lấy ra."* | nhân hoá + mơ hồ |
| *"Cái map kia nằm trong bộ nhớ chung."* | *"byID là một map sống trong bộ nhớ của process."* | mở bằng `Cái` |
| *"Một transaction giữ được cả ba mươi ngày này trong tay không?"* | *"Ba mươi ngày đó có gói vào một transaction được không?"* | nhân hoá |

Để ý câu đầu: bản sửa **dài hơn** bản gốc. Viết rõ thường tốn chữ hơn viết ẩn dụ,
và nó nói thêm được một điều thật (`InTx` nhét `pgx.Tx` vào ctx —
`internal/adapter/postgres/postgres.go:104`) mà bản ẩn dụ không nói.

### Ba quy tắc của lời giải thích

1. **Từ mới phải được giới thiệu trước khi dùng**, bằng `TermCard`, và luôn kèm
   một dòng *"bên Symfony cái này là gì"*. Định nghĩa rơi vào chỗ trống; một phép
   tương đương rơi vào sáu năm kinh nghiệm đã có.
2. **Câu đầy đủ, không phải từ khoá rời.**
3. **Bullet hiện dần theo nhịp.** Cả khối chữ hiện cùng lúc thì được đọc theo thứ
   tự ngẫu nhiên, mà thứ tự chính là lập luận.
4. **Tắt ligature của JetBrains Mono.** Font này tự nối `!=` thành `≠` và `:=`
   thành một ký tự dính. Trong editor thì đẹp; trong bài học thì sai, vì người
   học phải thấy đúng những ký tự họ sẽ gõ. Tắt ở `src/index.css` cho video và ở
   `cheatsheet/plate.css` cho bản in.
5. **Code PHP để đối chiếu phải dán nhãn.** `PHP_SNIPPETS` trong `data/portage.ts`
   mang `lang: "php"`, và `CodeCard` in chữ PHP ở góc phải. Người xem không được
   phép nhầm một minh hoạ với code có thật trong repo.

---

## 10. Lệnh

Trước khi render bất cứ thứ gì, đếm lại số:

```bash
cd learn
./scripts/verify-numbers.sh        # 10 con số của GO_USAGE vs repo Go thật
```

Kết quả xuất ra `out/`, chia theo mùa và tập:

```
out/mua1-go/ep01-vong-doi/       video.mp4 · cheatsheet.png · cheatsheet.pdf · canh/
out/mua3-portage/ep01-nam-vung/  video.mp4 · cheatsheet.png · cheatsheet.pdf · canh/
```

```bash
cd learn
npx remotion studio --no-open                                    # xem và sửa trực tiếp
npx remotion still Go-Ep01-S3 out/s3.png --frame=480 --scale=0.5  # soi một khung
npx remotion render Go-Ep01-VongDoi out/go-ep01.mp4               # render mùa 1 tập 1
npx remotion render Ep01-NamVung    out/ep01.mp4                  # render mùa 3 tập 1
```

Mỗi cảnh là một composition riêng (`Go-Ep01-S1` …), nên sửa một cảnh không phải
tua cả video.

---

## 11. Việc còn lại

**Cả 24 tập đã dựng xong.** Mỗi tập một video và một bản in A3, cùng nằm trong
`out/<mùa>/<tập>/`. Tất cả đã qua bốn cửa kiểm ở §12.

| Tập | Tên | Cảnh | Dài |
|---|---|---|---|
| MÙA 1 · 1 | PHP chết sau mỗi request. Go thì không. | 8 | 2:08 |
| MÙA 1 · 2 | Đọc một dòng Go | 8 | 2:06 |
| MÙA 1 · 3 | Lỗi là giá trị trả về | 7 | 1:49 |
| MÙA 1 · 4 | Receiver: (m Money) và (e *Events) | 6 | 1:32 |
| MÙA 1 · 5 | Interface ngầm: Go không có implements | 6 | 1:32 |
| MÙA 1 · 6 | Zero value: Go không có null | 5 | 1:14 |
| MÙA 1 · 7 | Embedding: giống kế thừa nhưng không phải | 5 | 1:15 |
| MÙA 1 · 8 | slice, map và ba cái bẫy | 5 | 1:16 |
| MÙA 1 · 9 | defer, panic, recover | 5 | 1:17 |
| MÙA 1 · 10 | Đồng thời: goroutine, mutex, channel | 5 | 1:17 |
| MÙA 2 · 1 | Mô hình thiếu máu | 4 | 0:58 |
| MÙA 2 · 2 | Value Object và Entity | 3 | 0:43 |
| MÙA 2 · 3 | Aggregate và invariant | 3 | 0:44 |
| MÙA 2 · 4 | Repository là cổng | 3 | 0:43 |
| MÙA 2 · 5 | Bounded Context | 3 | 0:44 |
| MÙA 2 · 6 | Domain Event | 3 | 0:44 |
| MÙA 3 · 1 | Vì sao project Go này bị chia làm năm? | 11 | 2:45 |
| MÙA 3 · 2 | Outbox: hai việc, một transaction | 3 | 0:43 |
| MÙA 3 · 3 | Nhất quán sau cùng có mã trạng thái | 3 | 0:43 |
| MÙA 3 · 4 | Port & Adapter, và luật chiều phụ thuộc | 3 | 0:43 |
| MÙA 3 · 5 | Điểm không thể quay đầu | 3 | 0:44 |
| MÙA 3 · 6 | Báo giá là một bức ảnh, và đối soát | 2 | 0:28 |
| MÙA 3 · 7 | Màn hình cần bốn context: đừng JOIN | 3 | 0:41 |
| MÙA 3 · 8 | Máy được tạo nháp, không được xác nhận | 3 | 0:46 |

Còn lại:

- [ ] Xem lại bằng mắt từng video (checker chỉ đo lề, không đo nhịp kể chuyện)
- [ ] Khi một con số trong repo Go đổi: `./scripts/verify-numbers.sh`, sửa
      `src/data/portage.ts`, rồi render lại tập nào dùng con số đó
- [ ] Khi code Go đổi: `python3 scripts/verify-snippets.py` để biết snippet nào lệch file

---

## 12. Bốn cửa kiểm trước khi báo xong

`tsc` xanh và preview trong Studio **không** bắt được lỗi thật — đã có hơn mười
lỗi hình lọt qua cả hai. Bốn lệnh này thì bắt được:

```bash
cd learn
./scripts/verify-numbers.sh        # 11 con số GO_USAGE vs repo Go thật
python3 scripts/verify-snippets.py # 30 snippet: mọi dòng code có thật trong file nó khai
python3 scripts/check-overflow.py  # 110 cảnh: không cảnh nào tràn khung
# và sau khi render: mọi mp4 phải có đúng 1 track âm thanh
for f in out/*/*/video.mp4; do
  echo "$f $(node_modules/@remotion/compositor-linux-x64-gnu/ffmpeg -i "$f" 2>&1 | grep -c 'Audio: aac')"
done
```

Mỗi cửa sinh ra từ một lỗi thật đã xảy ra:

| Cửa | Lỗi nó bắt được lần đầu |
|---|---|
| `verify-numbers.sh` | `assertions` khai 53 nhưng repo có 48; `defers` khai 96 vì đếm cả một chữ `defer` nằm trong comment |
| `verify-snippets.py` | snippet `oneDoor` viết `StatusPlaced`/`ErrWrongStatus` — hai cái tên không tồn tại trong repo; `repoPort` âm thầm bỏ mất method `All()`; `mainOnce` viết `go relay.Run(ctx)` trong khi file thật là `go func() { … }()` |
| `check-overflow.py` | chữ bị cắt dưới đáy khung ở 1:04 tập 1 — người học báo, không phải máy |
| đếm track âm thanh | Mùa 3 tập 1 ra lò câm vì dựng trước khi có `Sfx.tsx` |
