# UI-REDESIGN-PLAN.md — làm lại giao diện bằng hai plugin thiết kế

> **Đã làm xong 23/09/2026** (SETUP.md §9 đợt 21, WALKTHROUGH.md §23j), theo đúng năm đề nghị ở §8
> vì chưa mục nào được trả lời. Phần dưới giữ nguyên như lúc viết.
>
> Plan viết 23/09/2026, **chưa làm gì**. Người thực hiện là một agent khác. File này phải
> đủ để agent đó bắt tay làm mà không cần đọc lại cuộc trò chuyện đã sinh ra nó.
> Anh Sỹ duyệt các mục ở **§8** trước khi agent bắt đầu. Mục nào chưa trả lời thì agent làm
> theo **đề nghị** ghi ở đó.

---

## 0. Đọc trước khi làm

| Đọc | Vì sao |
|---|---|
| `CLAUDE.md` | luật làm việc: giải thích bằng tiếng Việt, comment code bằng tiếng Anh, **không tự commit** |
| `docs/UI-GUIDE.md` | bốn màn hình dành cho ai, bấm gì, và bảng "hỏng thì xem" |
| `docs/WALKTHROUGH.md` §23 | vì sao có hai màn hình theo vai, thay cho kiểu "một nút cho mỗi endpoint" |
| `web/app/words.js` (đầu file) | **hai quy tắc của màn hình**, xem §2 |
| skill `frontend-design:frontend-design` và `ui-ux-pro-max:ui-ux-pro-max` | hai plugin mà việc này sinh ra để dùng |

Chạy lên để xem bản hiện tại:

```bash
go run ./cmd/api -web ./web     # in-memory, có sẵn chìa dev-operator / dev-customer
# http://localhost:8080/ui/            trang chọn
# http://localhost:8080/ui/customer.html
# http://localhost:8080/ui/staff.html
```

---

## 1. Hiện trạng (đếm ngày 23/09, HEAD `ea04696`)

| File | Dòng | Là gì |
|---|---:|---|
| `web/index.html` | 43 | trang chọn, HTML tĩnh dùng chung `screens.css` |
| `web/customer.html` + `web/app/customer.js` | 22 + 380 | màn hình khách: `PasteForm` · `MyProduct` · `QuoteBreakdown` · `MyOrders` |
| `web/staff.html` + `web/app/staff.js` | 22 + 382 | màn hình nhân viên: `ProductCard` (checklist bốn bước) · `BuyingQueue` · `MoneyQueue` |
| `web/app/ui.js` | 175 | component dùng chung: `Top` · `Card` · `Field` · `Select` · `Steps` · `Toasts` + ba hook `usePoll` · `useAction` · `useToasts` |
| `web/app/screens.css` | 186 | style của hai màn hình làm việc; màu xanh `#1f5fd0` trên nền `#f6f7f9`, font hệ thống, **chỉ có light** |
| `web/app/words.js` | 134 | mọi mã máy → chữ người đọc, kèm tone của pill và `why` |
| `web/app/portage.js` · `api.js` | 111 · 179 | `call()`, `money()`, chìa khoá, gọi `fetch` |
| `web/console.html` + `main.js` · `steps.js` · `console.css` | 45 + 611 · 321 · 188 | bảng kiểm API cho lập trình viên (26 bước) |
| `web/flow.html` | 1560 | mô phỏng luồng đơn, **không gọi API**, CSS nằm luôn trong file |
| `web/vendor/` | — | `react.production.min.js` · `react-dom.production.min.js` · `htm.module.js` |

Stack: React UMD và `htm` (JSX dạng template string), ES module, **không npm, không bước
build**. API binary phục vụ thư mục `web/` ở `/ui/`, cùng origin nên không có CORS
(`go run ./cmd/api -web ./web`).

---

## 2. Ràng buộc: làm lại cỡ nào cũng không được phá

1. **Không chữ nào của máy ra tới màn hình.** Mọi mã (`issued`, `branded`, `awaiting_deposit`…)
   đi qua `words.js`. Redesign **đổi cách trình bày** chữ, không đổi nội dung chữ. Muốn sửa
   câu chữ thì sửa trong `words.js`, không viết cứng vào component.
2. **Mọi con số nói được nó từ đâu ra.** Giữ `why` và phần `QuoteBreakdown`: tiền hàng, thuế
   8,81 %, cước, phí dịch vụ, cọc 50 %.
3. **Không bước build, không CDN.** `index.html` giải thích lý do: *"the screen works with no
   network and no npm install"*. Hệ quả:
   - **không** dùng Tailwind (bản play-CDN cũng không);
   - **không** GSAP;
   - **không** Google Fonts qua mạng. Font phải tự host trong `web/vendor/fonts/` (xem §4.2).
4. **Một hành động sáng tại một thời điểm.** Bước đã xong **nói nó đã ghi gì**, không chỉ xám
   đi. Không để uuid lên trước mắt người đọc. Cả ba điều này đã ghi ở đầu `screens.css`.
5. **Không đổi API, không đổi Go.** Việc này chỉ đụng `web/` và docs. `go test ./...` phải
   giữ nguyên số test.
6. **Số vàng phải hiện đúng trên màn hình** khi đi hết luồng: báo giá `5.393.720 ₫`, cọc
   `2.696.860 ₫`. Nếu con số nào đổi thì đó là bug của redesign, không phải của nghiệp vụ.
7. **Bốn trang, bốn vai** (UI-GUIDE): không gộp trang khách với trang nhân viên, không biến
   `console.html` thành màn hình làm việc.

---

## 3. Hai plugin nói gì, và lấy gì từ đó

### 3.1 UI/UX Pro Max: kết quả thật của công cụ

Lệnh đã chạy ngày 23/09 (Python 3.12 vừa cài; trên PATH mới sau khi mở terminal mới):

```bash
python "<plugin>/scripts/search.py" "logistics delivery tracking cross-border purchasing" \
  --design-system -p "Portage" --density 7 --variance 4 --motion 3 -f markdown
```

Nó khớp dòng **49 · Logistics/Delivery** trong `products.csv` và `ui-reasoning.csv`. Tóm tắt
kết quả và quyết định cho từng mục:

| Công cụ đề nghị | Lấy? | Vì sao |
|---|---|---|
| Style **Minimalism & Swiss**, dạng lưới, tương phản cao | **lấy** | đúng loại việc: màn hình thao tác, đọc lướt, nhiều số |
| Màu trạng thái xanh lá / hổ phách / đỏ, tách khỏi màu nhấn | **lấy** | khớp tone `ok · warn · bad · info · flat` mà `words.js` đã có |
| Xanh `#2563EB` + cam `#EA580C` cho theo dõi đơn | **lấy ý, đổi màu** | ý hay: một màu cho "đang chạy", một màu cho "việc của bạn". Nhưng `#2563EB` là màu mặc định của mọi SaaS; plugin `frontend-design` coi đó là dấu hiệu của trang máy sinh (xem §4.1) |
| Font **Inter** | **bỏ** | lựa chọn mặc định nhất có thể có. Chính dữ liệu plugin có cặp **"Vietnamese Friendly"** (`typography.csv` mục số 21: Be Vietnam Pro + Noto Sans), hợp hơn cho một sản phẩm mà mọi chữ là tiếng Việt |
| Mẫu **"Real-Time / Operations Landing"**: hero, CTA "Start trial" | **bỏ** | đó là trang bán hàng. `customer.html` và `staff.html` là màn hình làm việc; không có ai cần thuyết phục |
| Hiệu ứng GSAP khi cuộn trang | **bỏ** | vi phạm ràng buộc 3; và một hàng đợi việc không cần hiện dần khi cuộn |
| Anti-pattern *"Static tracking · No map integration"* | **bỏ phần bản đồ** | Portage không có GPS. Hành trình là 4 trạng thái (`TRACKING` trong `words.js`: chưa có gì · chờ về kho · đã về kho, đã cân · đã bay). Vẽ bản đồ là bịa dữ liệu. Phần "đừng để tracking tĩnh" thì **lấy**: xem §4.3 |
| Checklist trước khi giao: tương phản 4.5:1, focus thấy được, `prefers-reduced-motion`, thử ở 375/768/1024/1440 px, không dùng emoji làm icon | **lấy nguyên** | thành cửa kiểm ở §6 |

Sau khi anh duyệt §8, agent **lưu design system** bằng công cụ, để các lần làm sau đọc lại
được thay vì sinh lại:

```bash
python "<plugin>/scripts/search.py" "logistics delivery tracking cross-border purchasing" \
  --design-system --persist -p "Portage" --output-dir "D:/portage" --density 7 --variance 4 --motion 3
```

Lệnh này tạo `design-system/portage/MASTER.md`. Agent **sửa tay** file đó cho khớp §4 (font,
màu, các mục đã bỏ ở bảng trên), vì kết quả thô của công cụ vẫn còn Inter và GSAP.
`<plugin>` = `C:/Users/tony_/.claude/plugins/cache/ui-ux-pro-max-skill/ui-ux-pro-max/2.13.0/.claude/skills/ui-ux-pro-max`.

### 3.2 Frontend Design: điều nó đòi

- Lấy chất liệu từ **chính nghề này**. Với Portage đó là nhãn vận đơn hàng không, cân kho,
  số ký quy đổi, "lô gom", phong bì *par avion* sọc xanh–đỏ. Không lấy chất liệu từ "một
  dashboard SaaS bất kỳ".
- Chỉ **một** chỗ táo bạo, phần còn lại im lặng và kỷ luật.
- Tránh các dấu hiệu máy sinh:
  - nhãn IN HOA giãn chữ trên mọi tiêu đề;
  - chuỗi meta nối bằng dấu `·`;
  - mọi thứ nhét vào card bo góc giống nhau, đổ cùng một bóng;
  - font mono cho nhãn nhỏ;
  - thêm `→` vào mọi nút.
- Chữ viết từ phía người dùng: *"Chuyển cọc"*, không phải *"Submit"*. Một hành động giữ
  **một** tên suốt luồng: nút *"Xác nhận đã mua"* thì thông báo là *"Đã xác nhận mua"*.
- Làm hai lượt: plan design → soi lại xem chỗ nào là mặc định → sửa → mới code.

---

## 4. Hướng thiết kế đề nghị

### 4.1 Màu: "phong bì par avion"

Ý tưởng: sọc xanh–đỏ của phong bì thư máy bay là thứ ai gửi hàng quốc tế cũng nhận ra, và
chỉ Portage có lý do dùng nó. Hai màu đó dùng **rất ít**. Phần lớn màn hình là nền giấy
xanh-xám và mực xanh đen.

| Token | Light | Dark | Dùng cho |
|---|---|---|---|
| `--paper` | `#F2F4F7` | `#0E141B` | nền trang |
| `--sheet` | `#FFFFFF` | `#16202A` | mặt phiếu (vùng đang làm việc) |
| `--ink` | `#14202E` | `#E6ECF2` | chữ chính |
| `--ink-2` | `#4A5868` | `#A9B6C4` | chữ phụ |
| `--rule` | `#D3DAE2` | `#2A3746` | đường kẻ |
| `--airmail-blue` | `#1B3F8B` | `#7FA3F0` | hành động chính, link, trạng thái "đang chạy" |
| `--airmail-red` | `#C22A2A` | `#F08080` | **chỉ** sọc nhận diện và "việc đang chờ bạn". **Không** dùng cho lỗi |
| `--ok` · `--warn` · `--bad` | `#1F7A4A` · `#8A5A00` · `#A3261F` | `#6CCB94` · `#E3B25A` · `#F08A7E` | tone của pill trong `words.js` |

Chuyện đỏ nhận diện trùng với đỏ lỗi: **cố ý tách**. Đỏ lỗi ngả gạch (`#A3261F`) và luôn đi
kèm chữ và icon, không bao giờ chỉ là màu (ux-guidelines: không dựa riêng vào màu).
Agent kiểm hai màu đỏ đặt cạnh nhau trong cả hai theme. Nếu vẫn lẫn thì đổi "việc đang chờ
bạn" sang `--airmail-blue` đậm và giữ đỏ **chỉ** cho sọc.

Mọi cặp chữ/nền phải đạt ≥ 4.5:1. Agent đo lại, không tin bảng này.

### 4.2 Chữ

- **Be Vietnam Pro** cho mọi thứ: tiêu đề 600–700, thân 400–500. Một họ chữ, thiết kế cho
  dấu tiếng Việt: dấu chồng như "ướ", "ặ" không đè lên dòng trên.
- Số tiền và số cân dùng `font-variant-numeric: tabular-nums` để cột số thẳng hàng. Agent
  **kiểm** Be Vietnam Pro có bảng `tnum` không. Nếu không có thì chỉ các ô số dùng
  **IBM Plex Mono**; không dùng mono cho nhãn.
- **Tự host**: tải `woff2` (subset `latin` + `vietnamese`, các weight 400/500/600/700) vào
  `web/vendor/fonts/`, khai báo `@font-face` trong CSS, kèm `OFL.txt`. Fallback:
  `system-ui, "Segoe UI", Roboto, sans-serif`.
- Cỡ chữ: thân 16 px (hiện là 15 px, dưới mức plugin khuyên). Thang 13 · 16 · 19 · 24 · 32.
  Dòng không quá khoảng 70 ký tự.

### 4.3 Chỗ táo bạo duy nhất: **dải hành trình**

Hiện tại mỗi đơn có một pill trạng thái. Đề nghị thay bằng một **dải ngang kiểu nhãn vận
đơn**, đặt trên mỗi đơn ở trang khách và mỗi món ở trang nhân viên. Trạng thái đọc từ
`ORDER`/`TRACKING` của `words.js`:

```
 ┌───────────────┬──────────────┬───────────────┬─────────────┬──────────┐
 │ Báo giá       │ Đã cọc       │ Đã mua        │ Kho Denver  │ Đã bay   │
 │ 5.393.720 ₫   │ 2.696.860 ₫  │ 20/09         │ 1,5 kg cân  │ lô #3    │
 └───────────────┴──────────────┴───────────────┴─────────────┴──────────┘
   ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■░░░░░░░░░░░░░░░░░░░░░░░░░░░
                                     ▲ bạn đang ở đây — "kho đang chờ hàng về"
```

- Mỗi ô **ghi con số hoặc ngày đã xảy ra ở bước đó**, lấy từ dữ liệu API có sẵn. Không có số
  thì để trống, **không bịa**. Đây là quy tắc 2 của `words.js` biến thành hình.
- Ô hiện tại dùng `--airmail-blue`. Ô "chờ bạn làm" (ví dụ chờ chuyển cọc) có sọc
  `--airmail-red`.
- Đơn huỷ hoặc `purchase_failed`: dải dừng ở ô cuối cùng đã đạt, và một dòng `why` nói tiền
  được hoàn thế nào.
- Chuyển ô có một hiệu ứng duy nhất (≤ 250 ms). Tắt hẳn khi `prefers-reduced-motion`.
- Màn hình hẹp (≤ 480 px): dải dựng dọc.

Phần còn lại của màn hình im lặng: danh sách thẳng hàng, **không** card bo góc cho từng thứ.
Chỉ "phiếu đang làm" (món đang mở ở trang nhân viên, món đang xin báo giá ở trang khách)
được nổi lên thành mặt `--sheet`.

### 4.4 Bố cục

- **Trang khách** (điện thoại trước): một cột. Trên cùng là "việc đang chờ bạn" (nếu có).
  Kế tiếp là ô gửi link. Rồi tới danh sách đơn, mỗi đơn có dải hành trình.
- **Trang nhân viên** (màn rộng trước): hai cột từ 1024 px. Trái là ba hàng đợi (món chờ xử
  lý · việc đi mua · tiền chờ thu), mỗi hàng đợi hiện **số đếm**. Phải là phiếu đang mở.
  Dưới 1024 px xếp chồng một cột.
- Chìa khoá: giữ ở cuối trang như hiện tại (UI-GUIDE §1), nhưng thu gọn thành một dòng.
- Icon: SVG vẽ tay trong một file `web/app/icons.js` (theo phong cách Lucide, stroke 1.5).
  Không emoji. Icon đứng một mình phải có `aria-label`.

---

## 5. Phạm vi và thứ tự làm

Mỗi giai đoạn kết thúc với `go test ./...` vẫn xanh và màn hình vẫn chạy được, để dừng ở đâu
cũng không để lại một nửa hỏng.

| # | Việc | File đụng | Xong khi |
|---|---|---|---|
| G0 | Lưu design system (§3.1), sửa tay cho khớp §4 | `design-system/portage/MASTER.md` (mới) | file có đủ màu, chữ, các mục đã bỏ, checklist |
| G1 | Nền: token hai theme, font tự host, reset, `icons.js`; component trong `ui.js` đổi theo | `web/app/screens.css` (viết lại) · `web/app/ui.js` · `web/app/icons.js` (mới) · `web/vendor/fonts/` (mới) | hai trang cũ vẫn chạy đúng luồng, chỉ khác giao diện |
| G2 | Component `Journey` (dải hành trình §4.3) | `web/app/ui.js` · `web/app/words.js` (chỉ **thêm** thứ tự các bước; không sửa chữ đang có) | hiển thị đúng với mọi `ORDER` × `TRACKING`, kể cả huỷ và `purchase_failed` |
| G3 | Trang nhân viên theo §4.4 | `web/staff.html` · `web/app/staff.js` | đi hết checklist bốn bước, mua, thu tiền; không cần cuộn tìm hành động kế tiếp |
| G4 | Trang khách theo §4.4 | `web/customer.html` · `web/app/customer.js` | gửi link → báo giá `5.393.720 ₫` → đồng ý → đặt đơn → cọc `2.696.860 ₫` |
| G5 | Trang chọn | `web/index.html` | bốn lối vào, nói rõ trang nào cho ai |
| G6 | `console.html`: **chỉ** đổi token/font cho đồng bộ, không đổi bố cục | `web/app/console.css` | 26 bước vẫn xanh, kết ở `variance -2.50 USD` |
| G7 | Docs | `docs/UI-GUIDE.md` (ảnh chụp/mô tả mới) · `docs/WALKTHROUGH.md` §23 · `CLAUDE.md` mục Giao diện · `docs/SETUP.md` §9 thêm một đợt | số dòng và tên file trong docs khớp code |

**Ngoài phạm vi** (chờ quyết ở §8): `flow.html`; màn hình bảng giá cước.

---

## 6. Kiểm chứng: không báo "xong" chỉ vì trang nhìn đẹp

1. `gofmt -l . && go vet ./... && go test ./...`: sạch, số test **bằng** trước khi làm.
2. **Đi thật hết luồng trên màn hình** với `go run ./cmd/api -web ./web`, trong trình duyệt
   (Claude in Chrome nếu có), từ đầu tới cuối:
   - nhân viên thêm shop (dùng `curl` như UI-GUIDE §0, vì chưa có màn hình cho việc này);
   - khách gửi link;
   - nhân viên đi hết bốn bước checklist;
   - khách xin báo giá → thấy **`5.393.720 ₫`** → đặt đơn → cọc **`2.696.860 ₫`**;
   - nhân viên nhận cọc, mua, và thấy dải hành trình đi tới đúng ô.
3. **Chạy lần thứ hai** trên cùng process (bài học 3 trong `CLAUDE.md`). Chạy một lần xanh
   không chứng minh gì cho lần hai.
4. Thử ở 375 · 768 · 1024 · 1440 px. Không có thanh cuộn ngang ở trang nào.
5. Light **và** dark: đo tương phản các cặp màu ở §4.1; hai màu đỏ không lẫn nhau.
6. Bàn phím: đi hết luồng khách chỉ bằng Tab/Enter; focus luôn thấy được.
7. Bật `prefers-reduced-motion` thì dải hành trình không có hiệu ứng.
8. Rút mạng rồi tải lại trang: font vẫn đúng, vì ràng buộc 3 cấm tải từ ngoài.
9. `console.html`: 26 bước vẫn xanh, kết ở `variance -2.50 USD`.
10. Tìm chữ máy lọt ra màn hình: tìm trong DOM các chuỗi `issued`, `awaiting_`, `branded`,
    `us_forwarder`, và bất kỳ chuỗi dạng uuid nào. Không được còn cái nào.

---

## 7. Tự phản biện

**Cách hiển nhiên hơn: dùng nguyên design system của công cụ** (Inter, xanh `#2563EB`, cam
`#EA580C`). Nhanh hơn, và "đúng dữ liệu". Không chọn vì plugin thứ hai, `frontend-design`,
được cài chính để tránh loại kết quả đó: nó gọi thẳng tên font mặc định và màu SaaS mặc định
là dấu hiệu của trang máy sinh. Plan này vẫn giữ **cấu trúc** mà công cụ đề nghị (Swiss, màu
trạng thái riêng, checklist tiếp cận). Chỉ đổi phần **nhận diện**, và lấy phần đó từ nghề
gửi hàng máy bay.

**Cách hiển nhiên hơn nữa: chuyển sang Vite + Tailwind + shadcn** vì plugin có sẵn hướng dẫn
cho stack `shadcn`. Không chọn:
- phá ràng buộc 3 (chạy không cần mạng, không cần npm);
- thêm một toolchain Node vào một repo Go;
- đổi cách chạy mà UI-GUIDE và WALKTHROUGH §23 đang dạy.

Nếu sau này thật sự cần build thì đó là một quyết định riêng, có ADR riêng, không nên nhét
vào một đợt đổi giao diện.

**Rủi ro lớn nhất: dải hành trình hứa nhiều hơn dữ liệu có.** Ví dụ ô "Đã bay · lô #3" chỉ
hiện được nếu API trả số lô cho khách. Agent phải kiểm từng ô với response thật
(`GET /me/orders`, `/me/products`, bảng đọc `reporting`). Ô nào không có dữ liệu thì để
trống. **Không** thêm endpoint để lấp: ràng buộc 5 cấm đụng Go.

---

## 8. Cần anh chốt (mỗi câu kèm đề nghị; im lặng = làm theo đề nghị)

| # | Câu hỏi | Đề nghị | Đã bỏ |
|---|---|---|---|
| D1 | Hướng màu "phong bì par avion" (§4.1)? | **đồng ý** | xanh/cam của công cụ: an toàn nhưng là mặc định |
| D2 | Có làm dark mode không? | **có**, trong G1 (token đã khai sẵn hai theme) | chỉ light như bây giờ: plugin coi đó là thiếu |
| D3 | `flow.html` (1560 dòng, CSS nằm trong file)? | **để nguyên** đợt này | làm lại cùng lúc: tăng gấp đôi khối lượng cho một trang không phải để làm việc |
| D4 | Màn hình **bảng giá cước** (`POST /lanes`, `POST /fx`) cho nhân viên? | **để sau**, chờ Q5 của T-001 (`agent/projects/portage/tasks/T-001-ratecard-yaml/analysis.md`) | làm luôn: quyết định lane ở file hay ở database chưa chốt, màn hình sẽ phải làm lại |
| D5 | Font tự host (~200 KB `woff2` trong repo)? | **có** | Google Fonts qua mạng: phá ràng buộc 3 |

---

## 9. Trạng thái working tree lúc viết plan: agent làm UI **đừng đụng**

Có việc dở của T-001 (ratecard) chưa commit, **không thuộc** việc này:

```
 M agent/projects/portage/PROJECT.md
 M agent/projects/portage/tasks/T-001-ratecard-yaml/{TASK.md,analysis.md}
 M go.mod  go.sum                         ← thêm gopkg.in/yaml.v3
?? agent/logs/runs/2026-09-23-19{45,49}-T-001-analyze.md
?? config/ratecard.yaml
?? internal/adapter/config/               ← có MỘT test đang đỏ (kỳ vọng "150%", domain in "150.0000%")
```

Vì test đỏ đó, `go test ./...` ở §6.1 **sẽ đỏ** cho tới khi anh quyết giữ hay gỡ phần
ratecard. Agent làm UI ghi nhận điều này và so số test **không tính** package
`internal/adapter/config`. Không tự sửa, không tự xoá.

---

## Corrections

*(chỉ anh viết. Một dòng có ngày ở đây thắng mọi thứ ở trên.)*
