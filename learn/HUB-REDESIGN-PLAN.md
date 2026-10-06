# Plan: làm lại trang học theo shadcn — danh sách bên trái, video bên phải

Ngày 06/10/2026. Yêu cầu của anh, nguyên văn: *"Refactor lại UI theo shadcn, với lại theo dạng
list bên trái và màn hình xem video bên phải, ngoài ra check lại toàn bộ câu từ, ngữ nghĩa phải
dễ đọc dễ hiểu, chứ hiện tại có vài từ hơi AI và theo dạng dịch 1 vài từ kỹ thuật bằng english
sang tiếng việt nên nó gây khó hiểu."*

Plan này viết **trước khi dựng**, theo đúng cách của `docs/UI-REDESIGN-PLAN.md` (đợt 21): mỗi
quyết định có đề nghị, lý do, phương án đã bỏ và chỗ có thể sai. Anh xem demo rồi chỉnh; không
cần đọc hết file này trước.

---

## 1. Hiện trạng, đo được

| Thứ | Số | Ở đâu |
|---|---|---|
| Trang của hub | 4: bản đồ · tập · ôn tập · roadmap | `hub/src/pages/` |
| Dòng code hub | 1 242 (TS/TSX/CSS) | `hub/src/` |
| CSS tự viết | 179 dòng, hệ "Afterimage" tối, chép từ `plate.css` | `hub/src/hub.css` |
| Tailwind | **đã có** trong `learn/` (4.0.0): Remotion dùng qua `src/index.css` | `src/Root.tsx` |
| Component thư viện | 0 — mọi nút, tab, thẻ đều tự vẽ | — |
| Từ dịch máy trong **khung** trang | "bản in" ×7 · "neo vào code" · "cảnh" ×3 · "tự tin" · "thẻ" · "bản đồ" | `hub/src/` |
| Từ dịch máy trong **nội dung** video | "mô hình thiếu máu" (3 cảnh + tên tập) · "Repository là cổng" (tên tập) · "bảng đọc" (2 cảnh) · "hợp đồng" (3 cảnh, tập 15) · "nhất quán sau cùng" (tên tập) | `src/videos/` |
| Số cũ còn sót | "278 test" viết chết trong 3 recap + 3 cảnh; số đo hiện tại là 266 | `ddd/ep04`, `go/ep05`, `portage/ep03` |

Hai phát hiện ngoài yêu cầu nhưng liên quan:

- **Video render ra đã mang preflight của Tailwind** (`Root.tsx` import `index.css`), còn hub
  hiện **không** nạp Tailwind — nên Player trong hub đang vẽ cảnh *khác* mp4 một chút. Đưa
  Tailwind vào hub làm hai bên **giống nhau hơn**, không phải xa hơn. Đây là lý do D1 không
  sợ preflight.
- "278" là con số chết. `verify-numbers.sh` không bắt được vì nó không nằm trong `GO_USAGE`.

## 2. Quyết định

### D1 — shadcn thật (Tailwind v4 + radix-ui qua CLI), không phải "vẽ cho giống shadcn"

**Đề nghị:** `components.json` ở `learn/`, alias `@/` → `hub/src/`, Tailwind qua
`@tailwindcss/vite`, component sinh bằng `npx shadcn add` vào `hub/src/components/ui/`:
button · badge · tabs · scroll-area · progress · separator · sheet · tooltip · dropdown-menu · card.

**Vì sao:** anh nói "theo shadcn"; viết CSS tay bắt chước shadcn là *dịch* shadcn, đúng cái lỗi
anh vừa chê ở câu chữ. Và trên CV, "shadcn/ui + Tailwind v4 + Radix" là thứ người đọc nhận ra.

**Đã bỏ:** (a) CSS tay theo token của shadcn — không có Radix thì tab, sheet, menu không có
bàn phím và ARIA đúng, phải tự viết lại; (b) giữ Afterimage tối như cũ và chỉ đổi bố cục —
không phải điều anh yêu cầu.

**Có thể sai ở:** thêm 6 dependency vào `learn/package.json` (chỉ hub dùng, Remotion không
đụng). Preflight của Tailwind áp lên cả Player — xem §1: mp4 vốn đã có preflight.

### D2 — Bố cục: một khung hai cột, danh sách bài bên trái, bài đang xem bên phải

```
┌──────────────────────────────────────────────────────────────────────┐
│ Học Go · Portage                    [Ôn tập 3] [roadmap.sh] [☾] [⋯]  │
├──────────────┬───────────────────────────────────────────────────────┤
│ Tiến độ 7/30 │  Mùa 1 · Tập 3                                        │
│ ▰▰▰▱▱▱▱▱▱▱   │  Lỗi là giá trị trả về                    1:49 · ✓    │
│              │  ┌─────────────────────────────────────────────────┐  │
│ MÙA 1 · Go   │  │                                                 │  │
│ ✓ 1 PHP chết…│  │                 PLAYER                          │  │
│ ✓ 2 Đọc một… │  │                                                 │  │
│ ▶ 3 Lỗi là…  │  └─────────────────────────────────────────────────┘  │
│   4 Receiver │  [Mục lục] [Cheatsheet] [Code trong repo] [Tự kiểm tra]│
│   …          │  1 · Lời hứa ............................. 0:00       │
│ MÙA 2 · DDD  │  2 · PHP-FPM: một request một đời ......... 0:10       │
│   1 Anemic…  │  …                                                    │
│ MÙA 3        │                                      [← Tập 2] [Tập 4 →]│
│ ───────────  │                                                       │
│ 2 cheatsheet │                                                       │
│ 3 bài tập    │                                                       │
└──────────────┴───────────────────────────────────────────────────────┘
```

- Cột trái (`Sidebar`, 300 px, cuộn riêng bằng `ScrollArea`): thanh tiến độ, ba mùa, mỗi tập một
  dòng có số, tên, thời lượng, dấu ✓ (đã xem) hoặc ✓✓ (đã thuộc); dưới cùng là hai cheatsheet
  tra cứu và ba bài tập. Chỗ "theo buổi của HOC.md" thành một công tắc nhóm **Theo mùa / Theo
  buổi học** ngay trên danh sách. Dưới 1024 px cột trái vào `Sheet` (ngăn kéo), mở bằng nút ☰.
- Cột phải: tiêu đề bài, Player, bốn tab bên dưới. **Mục lục** (các đoạn của video, bấm là tua)
  nằm trong tab thay vì cột thứ ba, vì ba cột không vừa 1280 px khi đã có danh sách bài.
- Trang chủ `#/` không còn "bản đồ": mở thẳng **bài kế tiếp** (như mọi trang khoá học). Tiến
  độ xuất / nhập / xoá vào menu ⋯ ở góc phải.
- `#/on` (Ôn tập) và `#/roadmap` thay chỗ cột phải, cột trái vẫn còn.

**Đã bỏ:** giữ trang bản đồ chấm tròn làm trang chủ — trùng với cột trái. Ba cột (danh sách ·
player · mục lục) — chật ở laptop 13".

### D3 — Sáng/tối theo hệ điều hành, có nút đổi; palette zinc của shadcn

Video thì tối, nhưng trang quanh video sáng là chuyện bình thường (YouTube). Nút ☾ lưu lựa
chọn vào `localStorage`. Màu nhấn duy nhất: `primary` của zinc; màu "đã thuộc" dùng xanh lá
của Tailwind (`emerald-600`), không đem bảng màu Afterimage sang.

### D4 — Chữ: Be Vietnam Pro tự host, JetBrains Mono cho code

Lý do nằm sẵn ở `design-system/portage/MASTER.md` §3: vẽ cho dấu tiếng Việt, đã có font
file trong `web/vendor/fonts/` (chép sang `learn/public/fonts/`, 392 KB, giấy phép OFL kèm theo).
Inter/Geist mặc định của shadcn đặt dấu tiếng Việt xấu hơn rõ. Mono giữ JetBrains Mono để
code trên trang giống code trong video.

### D5 — Câu chữ: một bảng từ, áp cho cả khung trang lẫn nội dung

Luật: **danh từ kỹ thuật giữ tiếng Anh** (ai đi phỏng vấn cũng nói bằng tiếng Anh); **động từ và
câu dẫn bằng tiếng Việt thường**, không có từ nào mình phải nghĩ "ý nó là gì".

| Đang có | Đổi thành | Vì sao |
|---|---|---|
| bản in | **cheatsheet** | "bản in" nghe như hoá đơn; cheatsheet là từ ai cũng dùng |
| neo vào code | **code trong repo** | "neo" là dịch *anchor*, không ai nói vậy |
| cảnh | **đoạn** (trong mục lục video) | "cảnh" là từ của phim; video ở đây là bài giảng |
| tự tin | **đã thuộc** | "tự tin" là cảm xúc, "đã thuộc" là trạng thái đo được |
| thẻ (ôn tập) | **câu hỏi** | không ai gọi câu hỏi ôn là "thẻ" ngoài người dùng Anki |
| ý nhớ lại | **ý chính** | |
| tới hạn | **đến hạn ôn** | |
| bản đồ | *(bỏ, thành danh sách bên trái)* | |
| đối chiếu roadmap.sh | **so với roadmap.sh** | |
| Xuất JSON / Nhập JSON | **Tải tiến độ về / Nạp tiến độ từ file** | |
| mô hình thiếu máu | **anemic model** | tên riêng của một anti-pattern; search "mô hình thiếu máu" không ra gì |
| Repository là cổng | **Repository là một port** | "Port & Adapter" đã để tiếng Anh ở mọi chỗ khác |
| bảng đọc | **read model** | cùng lý do; CQRS nói "read model" |
| hợp đồng (tập 15) | **contract** | là tên package `contracts` |
| nhất quán sau cùng | **eventual consistency** | người phỏng vấn hỏi bằng tiếng Anh |

**Giữ nguyên, cố ý:** "Mùa / Tập" — không phải từ dịch, và đổi là render lại 30 video + 32 bản
in; "Điểm không thể quay đầu" — tiếng Việt tự nhiên, không phải thuật ngữ; "Nhớ lại" trong tên
đoạn cuối mỗi video — là chữ của video, đổi phải render lại 30 tập. Nếu anh muốn "Phần / Bài"
thay cho "Mùa / Tập", đó là một đợt riêng, nói một câu là tôi làm.

Nội dung video đụng tới: `ddd/ep01` (tên tập + 3 cảnh + recap + cheatsheet), `ddd/ep04` (tên
tập), `go/ep15` (3 cảnh + recap), `portage/ep03` (tên tập), `portage/ep07` (1 cảnh + recap),
cộng các tập có câu "tập sau" nhắc tới chúng (`go/ep10`, `go/ep16`, `ddd/ep03`, `portage/ep02`,
`portage/ep06`). Render lại đúng những tập đó, chạy đủ năm cửa kiểm.

### D6 — "278" thành `REPO.testsNoDocker` ở mọi chỗ còn viết chết

Ba recap và ba cảnh. Thêm một dòng vào `verify-numbers.sh`: grep `278` trong `src/videos`
phải ra 0 — để số chết không quay lại.

### D7 — Kiểm thử bằng `data-testid`, không bằng class

`hub-check.mjs` hiện bám vào class CSS (`.dots .dot`, `.ep-head .ok`…). shadcn sinh class
tiện ích nên không còn tên để bám. Đặt `data-testid` ở đúng những chỗ script cần: danh sách
bài, dòng bài, nhãn đã xem, tab, dòng mục lục, nút chấm, bảng roadmap. Script viết lại theo đó,
vẫn đủ các kiểm cũ: 30 tập phát tới cuối, ba bề rộng, bàn phím, reduced-motion, tiến độ qua
reload, ôn tập, roadmap, hai link cheatsheet trả 200.

## 3. Tự phản biện

- **Cách hiển nhiên hơn:** chỉ đổi CSS sang token shadcn và đổi chữ, giữ nguyên bốn trang.
  Rẻ hơn một nửa. **Không chọn** vì anh yêu cầu rõ bố cục hai cột, và cột trái làm trang bản
  đồ thừa — giữ cả hai là hai nơi cho một sự thật.
- **Chỗ dễ sai nhất:** Tailwind preflight đổi cách Player vẽ cảnh. Đã kiểm: mp4 render qua
  `Root.tsx` *đã* có preflight. Vẫn sẽ chụp một cảnh trong Player trước/sau để so bằng mắt.
- **Chỗ dễ sai thứ hai:** đổi tên tập ("Mô hình thiếu máu" → "Anemic model") kéo theo
  CURRICULUM §2/§11, README, cheatsheet, `recap.next` của tập trước, `verify-hub` bảng §11.
  Có script kiểm, nên lệch là đỏ, không phải quên.
- **Không làm:** đổi vocab video "Mùa/Tập/Nhớ lại"; sửa `web/` của Portage; thêm backend hay
  tài khoản; dark mode cho cheatsheet (bản in vốn là tối, in ra giấy là việc khác).

## 4. Cửa kiểm trước khi báo xong

`tsc` · `eslint src hub/src` · `npm run hub:build` · `hub-check.mjs` trên bản build (Chrome
thật) · `verify-hub.py` · `verify-numbers.sh` (có dòng mới cho "278") · `verify-snippets.py` ·
`check-overflow.py` cho các tập render lại · đếm audio các mp4 render lại · chụp màn hình ba
bề rộng cho anh xem.

## Corrections

*(anh viết — một dòng có ngày ở đây thắng file này)*

---

## 5. Kết quả (06/10/2026, cùng ngày)

Làm đúng D1–D7. Chạy `cd learn && npm run hub` → `http://localhost:5173/`.

| Cửa kiểm | Kết quả |
|---|---|
| `tsc` · `eslint src hub/src` | sạch (còn đúng lỗi cũ `Sfx.tsx:68`, không đụng) |
| `npm run hub:build` + `hub-check.mjs` trên bản build | **sạch hai lần**: 30 bài phát tới cuối, 3 bề rộng, ngăn kéo, bàn phím, reduced-motion, dark mode qua reload, tiến độ qua reload, ôn tập, roadmap, 2 cheatsheet HTTP 200, không còn chữ cũ |
| `verify-hub.py` · `roadmap-gap.py` · `verify-snippets.py` | khớp hết |
| `verify-numbers.sh` | 32 con số khớp; dòng mới: không còn "278" viết chết |
| 12 tập render lại (bảng từ + số 278) | 12/12 exit 0, 12/12 có audio, `check-overflow` 0 cảnh tràn |
| 5 cheatsheet render lại | ok |
| `npm run hub` (dev) | trang, font, cheatsheet, CSS Tailwind đều 200 |

Hai chỗ anh có thể đổi hướng bằng một câu: (1) "Mùa / Tập" → "Phần / Bài" (render lại 30 video + 32 cheatsheet); (2) nền mặc định: hiện theo hệ điều hành, có nút đổi — nếu muốn luôn sáng hoặc luôn tối thì sửa một dòng ở `hub/src/theme.ts`.
