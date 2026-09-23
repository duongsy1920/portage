# Design System Master File — Portage

> Sinh bằng `ui-ux-pro-max` (`search.py --design-system --persist`, 23/09/2026, dòng 49
> *Logistics/Delivery*), rồi **sửa tay** theo `docs/UI-REDESIGN-PLAN.md` §4. Kết quả thô của
> công cụ có Inter, xanh `#2563EB`, cam `#EA580C`, GSAP và mẫu trang bán hàng; các mục đó đã bỏ,
> lý do ở §7 dưới cùng.
>
> **LOGIC:** khi làm một trang, xem `design-system/portage/pages/<trang>.md` trước. Có file đó thì
> luật trong đó thắng file này. Hiện chưa có trang nào cần luật riêng.

**Project:** Portage · **Loại:** màn hình làm việc (không phải trang bán hàng) ·
**Dials:** variance 4 · motion 3 · density 7

Code thật: `web/app/screens.css` (token, component), `web/app/fonts.css` (font),
`web/app/icons.js` (icon), `web/app/ui.js` (`Journey` và các component dùng chung).

---

## 1. Ý tưởng: phong bì *par avion*

Sọc xanh–đỏ của phong bì thư máy bay là thứ ai gửi hàng quốc tế cũng nhận ra, và chỉ Portage có
lý do dùng. Phần lớn màn hình là nền giấy xanh-xám và mực xanh đen; hai màu sọc dùng **rất ít**.

Chỗ táo bạo **duy nhất**: dải hành trình (§5). Mọi thứ khác gọn và kỷ luật, nhưng **không trơ**:
lượt đầu (23/09) hiểu "im lặng" thành "không có gì" và ra một cột phẳng với control mặc định, anh
Sỹ chê thẳng. Lượt hai sửa bằng bố cục có cột, panel có thứ bậc, control được thiết kế, và chuyển
động có lý do (§6).

## 2. Màu

| Token | Light | Dark | Dùng cho |
|---|---|---|---|
| `--paper` | `#F2F4F7` | `#0E141B` | nền trang |
| `--sheet` | `#FFFFFF` | `#16202A` | mặt panel, row và phiếu |
| `--ink` | `#14202E` | `#E6ECF2` | chữ chính |
| `--ink-2` | `#4A5868` | `#A9B6C4` | chữ phụ |
| `--rule` · `--rule-strong` | `#D3DAE2` · `#8D9AA8` | `#2A3746` · `#5C6B7C` | đường kẻ; kẻ đậm dưới tiêu đề mục |
| `--airmail-blue` | `#1B3F8B` | `#7FA3F0` | hành động chính, link, "đang ở bước này", **"việc đang chờ bạn"** |
| `--on-blue` | `#FFFFFF` | `#0E141B` | chữ trên nút xanh |
| `--airmail-red` | `#C22A2A` | `#F08080` | **chỉ** trong sọc: chân thanh trên cùng, ô "chờ bạn" của dải, lề danh sách "việc đang chờ bạn" |
| `--ok` · `--warn` · `--bad` | `#1F7A4A` · `#8A5A00` · `#A3261F` | `#6CCB94` · `#E3B25A` · `#F08A7E` | tone của pill trong `words.js` |
| `--*-wash` | `#E7ECF5` `#E4F1E9` `#F6EEDC` `#F7E5E3` | `#1B2A40` `#15291F` `#2B2414` `#321C1C` | nền pill, nền ô đang ở, nền lỗi |

**Đỏ nhận diện và đỏ lỗi — đã đo, và đã tách hẳn.** Tỷ lệ giữa hai màu đỏ là 1.28:1 ở light và
**1.07:1 ở dark**: mắt không phân biệt được. Nên theo phương án dự phòng của plan §4.1:
"việc đang chờ bạn" nói bằng **chữ đậm màu xanh** cộng với sọc; đỏ không bao giờ đứng một mình.
Lỗi luôn có chữ + nền `--bad-wash` + icon + `role="alert"` (component `Problem`).

**Tương phản (đo bằng script, 23/09).** 17 cặp × 2 theme đều ≥ 4.5:1. Thấp nhất:
`--ok` trên `--ok-wash` light = 4.58. Cao nhất cần để ý: nút xanh light 9.86, dark 7.40.

## 3. Chữ

- **Be Vietnam Pro** cho mọi thứ (400 · 500 · 600 · 700). Một họ chữ, vẽ cho dấu tiếng Việt.
  Tiêu đề 600–700, thân 400–500. Không IN HOA giãn chữ ở đâu cả.
- **Be Vietnam Pro không có `tnum`** (đã kiểm bằng fontTools: GSUB chỉ có `ccmp dnom frac liga
  locl numr`, chữ số rộng từ 385 tới 710 đơn vị). Nên có họ phụ `"Portage Figures"`: **chỉ**
  chữ số và `, - .` lấy từ IBM Plex Mono qua `unicode-range`; chữ cái và `₫` vẫn là Be Vietnam Pro.
  Dùng cho cột tiền (`td.amount`, `.fig`), không bao giờ cho nhãn.
- Tự host trong `web/vendor/fonts/` (subset `latin` + `vietnamese`, ~148 KB), giấy phép
  `OFL.txt` và `OFL-IBM-Plex-Mono.txt`. Không request nào ra ngoài (đã kiểm: chặn mọi host khác
  localhost, 32 tổ hợp trang × theme × bề rộng vẫn nạp đúng font).
- Thang: 13 · 16 · 19 · 24 · 32 px. Thân 16 px, dòng 1.55. Ghi chú tối đa 70 ký tự.

## 4. Bố cục và bề mặt

- Mỗi trang mở bằng `PageHead`: tiêu đề 30 px, một câu nói trang để làm gì, và vài **số đếm sống**
  (việc chờ bạn, món đang xử lý, đơn đang chạy); số nào có việc thì viền xanh.
- Ba tầng bề mặt: **panel** (`Section`: viền, bo 16 px, bóng nhẹ) chứa một mục; **row** (bo 12 px)
  là một món trong mục, viền đậm khi rê chuột; **sheet** là việc đang mở: bóng sâu hơn và một vạch
  xanh ở mép trái.
- Bo góc có thứ bậc: panel/sheet 16 px, row/ghi chú 12 px, nút/ô nhập 8 px, pill tròn hẳn.
- **Trang khách**: từ 1024 px hai cột. Trái 340–400 px là ô gửi link (dính khi cuộn); phải là
  việc đang chờ bạn → hàng của tôi → đơn của tôi. Hẹp hơn thì một cột theo thứ tự: việc đang chờ
  bạn → gửi link → hàng → đơn (`display: contents` + `order`).
- **Trang nhân viên** (màn rộng trước): từ 1024 px hai cột, trái 300–380 px là ba hàng đợi có
  số đếm (món chờ xử lý · việc đi mua · kho Denver · tiền chờ thu · chờ giao), phải là phiếu đang mở.
  Việc đã xong trong phiên nằm dưới nhãn riêng, để số đếm luôn khớp số dòng đang chờ.
  Dưới 1024 px xếp chồng; bấm một dòng thì cuộn tới phiếu.
- Vùng chạm tối thiểu 46 px cho nút và ô nhập. Ô nhập có nền `--field`, viền đổi khi rê, vòng
  `--ring` 4 px khi focus; select có mũi tên riêng. Nút chính xanh đặc có bóng cùng màu, nhấc 1 px
  khi rê, lún khi bấm, và **thành spinner khi đang gửi request** (`aria-busy`).

## 5. Dải hành trình (`Journey` trong `ui.js`, `journeyOf` trong `words.js`)

Sáu ô kiểu nhãn vận đơn: Báo giá · Đã cọc · Đã mua · Kho Denver · Đã bay · Đã giao.

| Trạng thái ô | Trông thế nào |
|---|---|
| `done` | nhãn mực đậm, icon xanh |
| `now` | cả ô đổ xanh đặc, chữ trắng, icon nhún nhẹ |
| `yours` | ô viền bằng sọc đỏ–xanh của phong bì; câu dưới dải có nhãn "Việc của bạn: …" |
| `stopped` | icon ✕, nền `--bad-wash`, nhãn màu `--bad` (huỷ hoặc shop không bán được) |
| `off` | nhãn gạch ngang, nền kẻ chéo mờ: bước sẽ không tới nữa |
| `todo` | chữ phụ |

Dưới nhãn là **đường bay**: một đường nét đứt, phần đã đi tô xanh, và một chiếc máy bay đứng ở
giữa ô đang ở. Khi đơn chuyển bước, máy bay **bay** tới ô mới (900 ms); lúc trang mở nó cất cánh
từ đầu đường.

Luật: một ô chỉ ghi **con số hoặc ngày mà API thật sự gửi** cho bước đó, cho **đúng màn hình được
đọc nó**. Cả hai màn hình: tổng và ngày đặt (Báo giá), số cọc (Đã cọc), phần còn lại khi đang bay
mà chưa trả (Đã giao). Chỉ màn nhân viên, và chỉ trong phiên đã làm việc đó: số tiền mua thật và
ngày mua (`GET /purchase-tasks/{id}`), cân thật và ngày cân (`GET /parcels`), phần cước của đơn
và ngày bay (`GET /batches/{id}`). Ô đã qua mà không có số thì ghi "xong", để một bước đã xong
không trông như một bước thiếu. Không thêm endpoint để lấp (plan §2.5).

≤ 560 px: dải dựng dọc, đường bay ẩn. `prefers-reduced-motion`: mọi hiệu ứng rút về 0,001 ms
(đã kiểm bằng trình duyệt).

## 6. Icon và chuyển động

- Icon SVG vẽ tay trong `web/app/icons.js`, lưới 24, nét 1.5, kiểu Lucide. Không emoji. Icon
  đứng một mình phải có `label` (thành `role="img"` + `aria-label`); icon cạnh chữ bị ẩn khỏi
  trình đọc màn hình.
- Chuyển động đều trả lời một điều gì đó, không để trang trí:

  | Khi | Chuyển động |
  |---|---|
  | trang mở | sọc phong bì chạy ngang; các panel lên dần theo thứ tự (60 ms mỗi cái) |
  | một món, một việc mới xuất hiện | dòng đó trượt lên, mờ dần vào |
  | chọn một việc ở bàn nhân viên | phiếu mới trượt lên thay phiếu cũ |
  | bước checklist xong | số thành vòng xanh "nảy", dấu ✓ tự vẽ, sợi chỉ giữa các bước xanh lên |
  | bước đang làm | vòng sáng quanh số "thở" chậm |
  | đơn chuyển bước | máy bay bay tới ô mới |
  | báo giá hiện ra | tám dòng tiền lần lượt hiện, 40 ms mỗi dòng |
  | đang gửi request | nút thành spinner |
  | lỗi | ghi chú lỗi rung nhẹ một lần |
  | thông báo | trượt vào từ phải, thanh đếm ngược 9 giây ở chân |
  | bấm "Xem" | cuộn mượt tới món đó, món đó phát sáng một lần |
  | đang tải | thanh skeleton lấp lánh thay cho chữ "Đang tải…" |

## 7. Đã bỏ từ kết quả của công cụ, và vì sao

| Công cụ đề nghị | Quyết định |
|---|---|
| Inter | **bỏ** — lựa chọn mặc định nhất; chính dữ liệu công cụ có cặp "Vietnamese Friendly" (Be Vietnam Pro) |
| `#2563EB` + cam `#EA580C` | **bỏ màu, giữ ý** (một màu "đang chạy", một dấu "việc của bạn") — xanh đó là màu mặc định của mọi SaaS |
| Mẫu "Real-Time / Operations Landing" (hero, CTA "Start trial") | **bỏ** — đây là màn hình làm việc, không ai cần thuyết phục |
| GSAP scroll reveal | **bỏ** — không CDN, không build; hàng đợi việc không cần hiện dần |
| Card nổi khi hover (`translateY`, `shadow-lg`) | **chỉ giữ ở chỗ bấm được**: thẻ lối vào ở trang chọn (nhấc 3 px) và nút (1 px). Panel và row nội dung không nhấc, chỉ đổi viền |
| Anti-pattern "no map integration" | **bỏ phần bản đồ** — Portage không có GPS, vẽ bản đồ là bịa dữ liệu. Phần "đừng để tracking tĩnh" thì giữ: đó là dải hành trình |
| Google Fonts `@import` | **bỏ** — font tự host |

## 8. Checklist trước khi giao (giữ nguyên từ công cụ, thêm hai dòng của Portage)

- [ ] Không emoji làm icon; icon cùng một bộ (`icons.js`)
- [ ] `cursor: pointer` trên mọi thứ bấm được
- [ ] Tương phản ≥ 4.5:1 ở **cả hai** theme
- [ ] Focus thấy được khi đi bằng bàn phím
- [ ] `prefers-reduced-motion` được tôn trọng
- [ ] Thử 375 · 768 · 1024 · 1440 px, không thanh cuộn ngang
- [ ] Không nội dung nào nằm dưới thanh trên cùng dính
- [ ] **Không chữ máy nào lên màn hình** (tìm `issued`, `awaiting_`, `branded`, `us_forwarder`, uuid)
- [ ] **Số vàng đúng**: báo giá `5.393.720 ₫`, cọc `2.696.860 ₫`
