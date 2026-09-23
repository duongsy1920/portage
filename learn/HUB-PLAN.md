# HUB-PLAN.md — một trang web để đi hết lộ trình học Go

> Plan viết 23/09/2026, **chưa làm gì**. Người thực hiện có thể là một agent khác: file này phải đủ
> để bắt tay làm mà không đọc lại cuộc trò chuyện. Anh Sỹ chốt các mục ở **§8**. Mục nào chưa trả
> lời thì làm theo **đề nghị** ghi ở đó.

---

## 0. Đọc trước

| Đọc | Vì sao |
|---|---|
| `learn/README.md`, `learn/CURRICULUM.md` | 24 tập, ba mùa, bốn cửa kiểm, luật "mọi dòng code lên hình phải có thật" |
| `learn/PLAN-BO-SUNG.md` | đối chiếu Mùa 1 với 84 mục roadmap.sh; sáu tập T11–T16 **chưa dựng** |
| `docs/HOC.md` | lộ trình buổi 0–9 và 17 câu tự kiểm tra |
| `learn/design/afterimage.md`, `learn/cheatsheet/plate.css` | ngôn ngữ thị giác của video và bản in |
| memory `ui-needs-polish-and-motion` | bài học đợt 21 của Portage: kiểm đo được không thay được việc **nhìn** |

---

## 1. Hiện trạng (đếm ngày 23/09, HEAD `5177a93`)

| Có | Ở đâu | Ghi chú |
|---|---|---|
| 24 tập video: Mùa 1 Go (10), Mùa 2 DDD (6), Mùa 3 Portage (8) | `learn/src/videos/{go,ddd,portage}/ep*/spec.ts` | mỗi tập là một composition Remotion; danh sách theo thứ tự dạy ở `SEASONS` trong `src/videos/index.ts` |
| mp4 | `learn/out/…` | **không vào git**, và máy Windows này chưa render. Muốn xem phải render trước |
| 24 bản in A3 | `learn/cheatsheet/*.html` + `plate.css` + `fonts/` | HTML tĩnh, font tự host, khổ cố định 1587×1123 px |
| Màn "Nhớ lại" cuối mỗi tập: 4 ý, 3 câu phỏng vấn, tập sau | `src/videos/*/ep*/scenes/*Recap.tsx` | dữ liệu **viết thẳng trong TSX**, không nằm ở chỗ nào đọc được |
| Neo vào code thật | `learn/src/data/portage.ts` | mỗi mảnh kèm đường dẫn và số dòng; `verify-snippets.py` canh |
| Lộ trình buổi 0–9, 17 câu tự kiểm tra | `docs/HOC.md` | Markdown, chưa gắn với tập nào bằng dữ liệu |
| 84 mục roadmap.sh | **không có trong repo** | script đối chiếu cũ nằm ở `/tmp` và đã mất (PLAN-BO-SUNG §1) |

Không có chỗ nào trả lời được câu **"hôm nay mình học gì tiếp, và mình đã tới đâu"**. Muốn xem một tập
phải mở Remotion Studio, là công cụ để *dựng* chứ không phải để *học*.

---

## 2. Trang này cho ai, làm việc gì

**Người dùng chính: anh Sỹ, đang học.** Năm việc, theo thứ tự hay gặp:

1. Mở ra là biết **tập tiếp theo** và mình đã đi được bao nhiêu.
2. Xem một tập, nhảy tới đúng cảnh cần xem lại.
3. Ôn nhanh 20–30 giây bằng bản in của tập đó.
4. Tự kiểm tra: câu phỏng vấn của từng tập và 17 câu của HOC.md, tự chấm "nhớ / chưa".
5. Biết còn thiếu gì so với roadmap.sh, và chỗ thiếu đó được vá bằng gì (tập, bản in, bài tập, hay cố ý bỏ).

**Người dùng phụ: người đọc CV.** Chỉ cần một đường link mở ra thấy ngay loạt bài và xem được một tập.
Chỉ có ý nghĩa nếu trang được đăng công khai (D3).

---

## 3. Ràng buộc

1. **`learn/` vẫn tách hẳn khỏi Go.** Không import gì từ repo Go. Không sửa gì ngoài `learn/`, trừ một
   dòng trong bảng docs của `CLAUDE.md`.
2. **Bốn cửa kiểm vẫn xanh**: `verify-numbers.sh`, `verify-snippets.py`, `check-overflow.py`,
   `render-plates.sh`. Trang học không được là lý do một tập đổi hình.
3. **Một nguồn cho mỗi sự thật.** Tên tập, cảnh, thời lượng, ý nhớ lại, câu phỏng vấn: trang học đọc
   đúng chỗ video đọc. Không chép sang file thứ hai.
4. **Thêm một tập vẫn là một `spec.ts` + một dòng trong `index.ts`.** Trang học tự thấy tập mới.
5. **Tiến độ học là của một người**, lưu trong trình duyệt. Có nút xuất/nhập file JSON để không mất
   khi đổi máy. Không server, không tài khoản.
6. Chất lượng tối thiểu: chạy được ở 375 · 768 · 1440 px, đi được bằng bàn phím, tôn trọng
   `prefers-reduced-motion`, tương phản ≥ 4.5:1.

---

## 4. Thiết kế đề nghị

### 4.1 Cách phát video: dùng composition sống, không cần mp4

`@remotion/player` phát thẳng một composition React trong trình duyệt. Trang học import `SEASONS`,
dựng mỗi tập bằng đúng `buildEpisode()` mà Studio dùng, và đưa vào `<Player>`. Được ba thứ:

- không phải render 24 mp4 trước, và không phải lưu mp4 ở đâu cả;
- sửa một cảnh là trang học thấy ngay;
- **mục lục cảnh** lấy từ `spec.scenes` (tên + `durationInFrames`): bấm một cảnh là tua tới đúng khung.

Stack: **Vite + React 19** ở `learn/hub/`, dùng chung `node_modules` và `tsconfig` của `learn/`.
Thêm `@remotion/player@4.0.525` (cùng bản với `remotion`, Remotion đòi mọi gói cùng số). Không có
Tailwind cho hub: CSS tay, theo token của `plate.css`.

### 4.2 Nhận diện: cùng một bộ với video và bản in

Video và bản in đã có ngôn ngữ riêng, **Afterimage**: nền `#080B10`, lưới ô mờ, JetBrains Mono +
Work Sans + Jura, màu nhấn `--go #00ADD8`, `--php #B39DFF`, `--money #F2B33D`. Trang học dùng đúng
bộ đó, để người học đi từ trang sang video sang bản in mà không thấy đổi thế giới.

**Không** dùng hệ "phong bì par avion" của Portage: đó là nhận diện của một sản phẩm mua hộ, còn
đây là một khoá học.

### 4.3 Bốn màn hình

**(a) Bản đồ lộ trình**: trang mở ra.

```
 HỌC GO — từ Symfony sang Go                           đã xem 7/24 · tự tin 4/24
 ┌────────────────────────────────────────────────────────────────────────────┐
 │ TIẾP THEO   Mùa 1 · Tập 8   slice, map và ba cái bẫy        1:16  [ Xem ]  │
 └────────────────────────────────────────────────────────────────────────────┘
  MÙA 1 · GO        ●━━●━━●━━●━━●━━●━━●━━○━━○━━○ · · ◌ ◌ ◌ ◌ ◌ ◌   (T11–T16 sắp có)
  MÙA 2 · DDD       ○━━○━━○━━○━━○━━○
  MÙA 3 · PORTAGE   ○━━○━━○━━○━━○━━○━━○━━○
  ● đã xem   ◉ tự tin (đã tự kiểm tra)   ○ chưa   ◌ đã lên kế hoạch, chưa dựng
```

Mỗi chấm là một tập: rê chuột hiện tên, thời lượng, "neo vào" file nào. Thứ tự là `SEASONS`, không
phải thứ tự tự đặt. Sáu tập T11–T16 lấy từ PLAN-BO-SUNG §5, hiện mờ, không bấm được.

**(b) Trang một tập**: chỗ làm việc chính.

```
 ┌───────────────────────────────── Player 16:9 ─────────────┐ ┌ Cảnh ───────────────────┐
 │                                                            │ │ 1 · Lời hứa        0:10 │
 │                                                            │ │ 2 · PHP-FPM…       0:18 │
 │                                                            │ │▶3 · Go: dựng một lần    │
 └────────────────────────────────────────────────────────────┘ │ …                       │
 [ Video ]  [ Bản in ]  [ Neo vào code ]  [ Tự kiểm tra ]        └─────────────────────────┘
```

- *Bản in*: nhúng `cheatsheet/<tập>.html` trong `iframe`, co cho vừa bề rộng; có nút mở toàn màn và tải PDF nếu đã render.
- *Neo vào code*: các mảnh của `data/portage.ts` mà tập đó dùng, mỗi mảnh ghi `file:dòng`, link tới
  GitHub ở **đúng commit** mà video đã được kiểm (không phải `main`, vì dòng sẽ trôi).
- *Tự kiểm tra*: 4 ý nhớ lại (ẩn, bấm mới hiện) và 3 câu phỏng vấn. Tự chấm *nhớ / mơ hồ / chưa*.
  Chấm đủ thì tập thành "tự tin".
- Xem hết cảnh cuối thì tập tự thành "đã xem". Không bắt bấm nút.

**(c) Ôn tập**: thẻ câu hỏi từ mọi tập đã xem + 17 câu HOC.md. Lịch nhắc đơn giản: câu chấm "chưa"
quay lại sau 1 ngày, "mơ hồ" sau 3, "nhớ" sau 7. Một buổi ôn tối đa 10 thẻ.

**(d) Đối chiếu roadmap.sh**: 84 mục, mỗi mục một dòng, gắn một trong năm nhãn *đã dạy (tập N)*,
*sẽ có (T1x)*, *bản in*, *bài tập*, *cố ý bỏ (lý do)*. Lọc theo nhãn. Đây là PLAN-BO-SUNG §4–§9 thành
một bảng bấm được.

### 4.4 Dữ liệu: đưa những gì đang nằm trong TSX về `spec.ts`

| Dữ liệu | Hiện ở | Đề nghị |
|---|---|---|
| tên, cảnh, thời lượng | `spec.ts` | giữ nguyên |
| 4 ý, 3 câu phỏng vấn, tập sau | `*Recap.tsx` | chuyển vào `spec.recap`; `*Recap.tsx` đọc `spec.recap` (D4) |
| tập nào neo vào mảnh code nào | rải trong các cảnh | thêm `spec.anchors: (keyof typeof portage)[]` |
| buổi HOC.md ↔ tập | chỉ có trong chữ | `src/data/path.ts`: buổi 0–9, mỗi buổi liệt kê tập và mục WALKTHROUGH |
| 84 mục roadmap.sh | không có | `src/data/roadmap.ts`, gõ lại từ PDF (D5), kèm nhãn của §4.3(d) |
| tiến độ của người học | không có | `localStorage` khoá `learn.hub.progress.v1`; xuất/nhập JSON |

Thêm một cửa kiểm thứ năm, `scripts/verify-hub.py`: mọi tập trong `SEASONS` có `recap` đủ 4 ý và 3 câu,
mọi `anchors` tồn tại trong `portage.ts`, mọi mục roadmap có đúng một nhãn, tập được trỏ tới có thật.

### 4.5 Chuyển động

Mở trang thì các chấm của bản đồ sáng dần theo thứ tự đã xem. Chấm vừa thành "tự tin" thì nảy một
nhịp. Bấm một cảnh thì thanh tiến độ trượt tới. Thẻ ôn lật. Tất cả tắt dưới `prefers-reduced-motion`.
Player giữ nguyên chuyển động của video: đó là nội dung, không phải trang trí.

---

## 5. Thứ tự làm

| # | Việc | Xong khi |
|---|---|---|
| H0 | **Spike, làm trước mọi thứ**: Vite + `@remotion/player` phát được **một** tập (Go-Ep01) có âm thanh, đúng font, Tailwind của cảnh chạy | tập 1 phát hết, không lỗi console, ở 1440 và 390 px. Không được thì dừng, báo lại, chọn phương án B ở §7 |
| H1 | Dữ liệu: `spec.recap`, `spec.anchors`, `path.ts`; 24 file `*Recap.tsx` đọc từ spec | bốn cửa kiểm cũ xanh, ảnh cảnh "Nhớ lại" của 24 tập **giống hệt** trước (so bằng `stills.py`) |
| H2 | Bản đồ lộ trình + tiến độ | tiến độ còn sau khi tải lại; xuất rồi nhập lại ra đúng như cũ |
| H3 | Trang một tập: player, mục lục cảnh, bản in, neo code, tự kiểm tra | cả 24 tập mở được, bấm cảnh nào tua tới đúng cảnh đó |
| H4 | Ôn tập | lịch 1/3/7 ngày đúng khi giả lập đồng hồ |
| H5 | Đối chiếu roadmap.sh (chờ D5 có dữ liệu) | 84 dòng, mỗi dòng một nhãn, `verify-hub.py` xanh |
| H6 | Build tĩnh, (D3) đăng GitHub Pages; docs: `learn/README.md`, `CURRICULUM.md` §10, dòng trong `CLAUDE.md` | mở link trên máy khác, không cần `npm i` |

---

## 6. Kiểm chứng

1. Bốn cửa kiểm cũ + `verify-hub.py` + `npm run lint` xanh.
2. Script trình duyệt (Chrome có sẵn, `puppeteer-core`): mở từng tập trong 24 tập, đợi Player sẵn sàng,
   tua tới cảnh cuối, **không** `pageerror`, không request font nào hỏng.
3. Số trên trang khớp số trong `CURRICULUM.md` §11 (số cảnh, thời lượng từng tập).
4. Tiến độ: đánh dấu 3 tập, tải lại, còn; xoá `localStorage`, nhập file JSON, về đúng.
5. 375 · 768 · 1440 px không cuộn ngang; bàn phím đi được từ bản đồ tới tự kiểm tra; reduced-motion tắt
   hết chuyển động của trang (không phải của video).
6. **Nhìn**: chụp bản đồ và trang tập ở hai bề rộng, soi xem có ra một bộ với bản in không.

---

## 7. Tự phản biện

**Phương án đơn giản hơn: render 24 mp4, rồi một trang HTML tĩnh trỏ tới mp4 và bản in.** Không cần
Vite, không cần Player. Không chọn làm chính vì: mp4 không vào git, phải render lại mỗi lần sửa một
cảnh, phải có chỗ chứa 24 file video để đăng lên, và mất mục lục cảnh bấm được. **Đây là phương án B**
nếu H0 thất bại, và giữ nguyên được §4.2–§4.4.

**"Remotion Studio đã là một giao diện rồi."** Studio để dựng: hiện timeline, props, frame. Nó không
biết người học đã xem tập nào, và không có tự kiểm tra.

**Rủi ro lớn nhất: composition không chạy ngoài Studio.** Đã kiểm 23/09:
- `remotion.config.ts` có bật `@remotion/tailwind-v4`, nhưng **không cảnh nào dùng `className`**
  (đếm được 0 file), nên hub không cần Tailwind. Nếu H0 thấy khác thì đó là tin quan trọng, báo lại.
- Font đi qua `@remotion/google-fonts` (`src/design/fonts.ts`): tải qua mạng lúc chạy, nên hub **không
  chạy được khi mất mạng**. Chấp nhận được cho một trang đăng công khai. Nếu cần chạy offline thì
  chuyển sang font tự host như bản in đã làm (`cheatsheet/fonts/`), nhưng đó là việc riêng, đụng video.
- Âm thanh `@remotion/sfx` với file trong `public/sfx`: Vite phải phục vụ đúng thư mục `public` đó.

Vì thế H0 là một spike **trước** mọi việc khác, có điều kiện dừng rõ ràng.

**Player nặng trên điện thoại.** Phát video bằng React tốn CPU hơn mp4. Chấp nhận được với người học
chính (máy tính). Nếu D3 = đăng công khai, trang nên có nút "tải mp4" cho điện thoại khi đã render.

**Chuyển Recap vào `spec.ts` là đụng 24 tập đang chạy tốt.** Đúng, và PLAN-BO-SUNG §11 cố ý không dựng
lại tập cũ. Nhưng việc này là **di chuyển dữ liệu**, không đổi hình. H1 có điều kiện xong là ảnh cảnh
cuối của 24 tập giống hệt trước. Phương án khác là script regex đọc TSX (như `episodes.py` đọc
`spec.ts`), nhưng làm vậy thì trang học đọc một thứ vốn không phải dữ liệu, và vỡ âm thầm khi ai đó
viết lại một câu nhiều dòng.

**Lịch ôn 1/3/7 là thứ trang trí?** Có thể. Nó là phần nhỏ nhất, làm sau cùng trong nhóm màn hình,
và bỏ được mà không ảnh hưởng gì khác.

---

## 8. Cần anh chốt (im lặng = làm theo đề nghị)

| # | Câu hỏi | Đề nghị | Đã bỏ |
|---|---|---|---|
| D1 | Phạm vi: chỉ Mùa 1 (Go) hay cả ba mùa? | **cả ba mùa, Mùa 1 lên đầu**: `SEASONS` đã có đủ, lọc bớt chỉ là thêm việc | chỉ Mùa 1: DDD và Portage vẫn phải học, lại phải mở Studio |
| D2 | Phát video bằng `@remotion/player` (§4.1)? | **có**, nếu H0 qua | mp4 + HTML tĩnh: là phương án B |
| D3 | Đăng công khai (GitHub Pages) để lên CV? | **có, sau H6**; repo `duongsy1920/portage` đã public (kiểm 23/09) | chỉ chạy máy nhà: mất cái lợi cho CV |
| D4 | Chuyển nội dung "Nhớ lại" từ TSX vào `spec.ts`? | **có**, kèm điều kiện ảnh giống hệt (H1) | regex đọc TSX: vỡ âm thầm |
| D5 | Nguồn 84 mục roadmap.sh? | **anh gửi lại PDF 06/09/2025**; agent gõ vào `roadmap.ts` và chạy lại phép đối chiếu thành `scripts/roadmap-gap.py` (PLAN-BO-SUNG §10.2) | tự chép từ roadmap.sh bây giờ: bản web có thể đã khác bản PDF mà PLAN-BO-SUNG đã đếm |

---

## 9. Quan hệ với các plan khác

- **PLAN-BO-SUNG** (T11–T16, hai bản in, ba bài tập) độc lập với plan này. Trang học hiện các tập đó ở
  trạng thái "đã lên kế hoạch" từ dữ liệu, nên dựng tập nào xong là nó tự sáng lên.
- **Lỗi "0 channel" của tập 10** (PLAN-BO-SUNG §2) nên sửa **trước H3**, vì trang học sẽ đưa tập đó ra
  trước mắt nhiều hơn.
- **`docs/UI-NEXT-PLAN.md`** (giao diện Portage) không liên quan: khác thư mục, khác nhận diện, khác người dùng.

---

## Corrections

*(chỉ anh viết. Một dòng có ngày ở đây thắng mọi thứ ở trên.)*
