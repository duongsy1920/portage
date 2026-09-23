# Coding Agent — Khám phá và thiết kế kiến trúc

> Trả lời cho `prompt.md` ("Go + DDD Coding Agent"). Đây là **Architecture
> Discovery + Design**: chỉ có file này được tạo. Không dòng Go, không schema,
> không Docker, không scheduler, không repo hay module riêng — cho tới khi bản này
> được duyệt.
>
> Hai đầu vào đã audit trước khi thiết kế, đúng thứ tự anh yêu cầu:
> **Portage** (hệ thật, case study) và **`markfulton/ai-employees`** (cảm hứng,
> không chép). Mọi con số và `file:dòng` dưới đây đều mở file kiểm, không viết từ
> trí nhớ.

Tài liệu đi theo chuỗi câu hỏi anh đưa:

```
Portage hiện có gì?                                    → §1
ai-employees có gì đáng học?                           → §2
Coding Agent thực sự cần gì?                        → §3
Giữ gì · bỏ gì · hoãn gì — kể cả sửa lại prompt        → §4
Go + DDD model nào thực sự có nghĩa?                    → §5
Phase 1 chính xác build cái gì?                        → §6
```

Rồi lộ trình (§7), rủi ro (§8), danh sách quyết định (§9), và bảng ánh xạ về 27
mục Part 3 của prompt (§10) để anh dò không thiếu mục nào.

---

## 0. Cách audit, và chỗ nó không đáng tin

- **Portage:** đọc `CLAUDE.md`, `docs/SETUP.md` §9, `docs/HOC.md`, `docs/P9-PLAN.md`,
  `docs/P10-PLAN.md`, `.github/workflows/ci.yml`, `internal/domain/decisions_test.go`,
  đếm test suite, `scripts/`, tìm `.claude/` (không có).
- **ai-employees:** clone bản mới nhất (195 file `.md`). Đọc `docs/STANDARD.md`
  (bản chuẩn 1.3), `HOW-EMPLOYEES-WORK.md`, `GUARDRAILS.md`, `HARNESSES.md`, và
  **`web-dev-employee`** trọn bộ — employee gần "coding" nhất — gồm `ROLE`,
  `CONTRACT`, `SCHEDULE`, routine `web-fix-runner` (routine sửa code trên nhánh),
  `web-guardrail-review` (routine nới/siết quyền), và một change brief thật.
- **Không đáng tin ở đâu:** tôi đọc một employee làm việc *trên website*, không phải
  trên codebase thuần. Chỗ nào tôi "dịch" khái niệm của họ sang coding (ví dụ
  *outbound action* → *commit/push*) là phán đoán của tôi, có ghi rõ.

---

## 1. Portage hiện có gì?

Đây là phát hiện quan trọng nhất của toàn bộ bài audit: **Portage đã là một Coding
Agent thô, chạy bằng kỷ luật của người thay vì bằng file cấu trúc.** Thiết kế
từ trang trắng là bỏ qua hai mươi đợt kinh nghiệm đã trả giá.

| Khái niệm đang có trong Portage | Tương đương ở Coding Agent | Trạng thái | Bằng chứng |
|---|---|---|---|
| `CLAUDE.md` — luật làm việc: *không tự commit*, *git identity nào*, *panic vs error*, *luật vàng import* | **CONTRACT.md** + **file luật của project** | **Đã có** (nhưng trộn hai thứ — xem §3) | `CLAUDE.md` mục "Cách làm việc — bắt buộc", "Quy ước code", "Luật vàng" |
| `CLAUDE.md` "Ba bài học đã trả giá" — mỗi bài có ngày, có chuyện thật, có luật rút ra | **Corrections có ngày** — đúng LAW 5 của ai-employees (*hard-won rule carries its date and lives in a file*) | **Đã có** | `CLAUDE.md` mục "Ba bài học"; SETUP §9 đợt 13, 14, 15 |
| `CLAUDE.md` "Trước khi bảo xong: ba kiểm" (lộ thông tin · nhất quán · luồng thật) | **Invariant cuối run** (ROLE §11 của ai-employees: bốn điều phải đúng trước khi ghi record) | **Một phần** — có luật, chưa có bằng chứng ghi lại từng lần | `CLAUDE.md` mục "Trước khi bảo xong" |
| `docs/SETUP.md` §9 — ≈20 đợt review, mỗi đợt một bảng **"Quyết định / bug → Chỗ nó nằm"** | **Run log + Decision register** | **Đã có** — dạng Markdown, người đọc trước máy | `SETUP.md:1381` (đợt 18), `:1460` (đợt 20); CLAUDE.md nói 20 đợt, grep đếm 19 tiêu đề |
| `docs/HOC.md` — 10 buổi học, 17 câu tự kiểm, "nói gì khi phỏng vấn", "ba thứ chỉ học được khi CHẠY THẬT" | **Learning system** | **Một phần** — là *giáo trình*, không phải *rút bài học theo từng task*; không có danh sách nợ học | `HOC.md` 522 dòng, buổi 0–9 |
| `docs/P10-PLAN.md` — §1 "Hai điều đã kiểm trong code", §2 Thiết kế, §3 File phải sửa, §4 Ba cái bẫy, §5 Test phải có, §6 Ghi lại sau khi xong; mở đầu bằng *"Quyết định đã chốt 06/09/2026"* | **Task + Plan + Approval** — plan có acceptance, có "không làm", có duyệt | **Một phần** — có nội dung, **không có trạng thái**; duyệt là một dòng văn xuôi có ngày, không phải dữ liệu | `P10-PLAN.md:3`, cấu trúc 6 mục, 191 dòng |
| `docs/P9-PLAN.md` §4 "Những chỗ còn mở (❓) — đã có đề xuất, **chốt rồi làm**" | **Challenge mode** thô: liệt kê câu hỏi mở, người chốt, rồi mới làm | **Một phần** | `P9-PLAN.md` §4 |
| `learn/PLAN-BO-SUNG.md` §3 "Tự phản biện trước khi đề xuất" | **Plan phải tự phản biện** (bài học 1 của CLAUDE.md) | **Đã có** — là pattern, chưa là bước bắt buộc | `PLAN-BO-SUNG.md` §3 |
| Test suite: **66** file `_test.go`, **9** adapter bộ nhớ, **3** file pgtest, **3** file httptest; CI chạy `gofmt` · `go vet` · `go test -race -cover` | **Validation capability** (`test.run`, `lint.run`, `fmt.check`) | **Đã có** | `ci.yml:34-45` |
| `internal/domain/decisions_test.go` — **7 test canh quyết định** bằng `go/ast`; header nói thẳng: *"A rule that lives only in a document is a rule someone forgets. A rule that lives in a test is a rule that fails the build."* | **Guardrail ép bằng code** — chính là cơ chế Phase 4 sẽ dùng cho invariant của Task | **Đã có, và đã có triết lý** | `decisions_test.go:1-16, 92-260` |
| `scripts/smoke.sh` in đúng số vàng; `learn/scripts/verify-{numbers,snippets}.{sh,py}` | **Deterministic helpers có self-test** (ai-employees §4: *scripts/ with a --selftest*) | **Một phần** | `scripts/`, `learn/scripts/` |
| `CLAUDE.md` "Quyết định 08/09 — HOÃN, đừng đề xuất lại" | **ADR** (có Context, Decision, và **"Xem lại khi"**: *khi `reconciliations` cho thấy variance âm liên tục*) | **Một phần** — là ADR tốt, nằm trong file không phải chỗ của nó | `CLAUDE.md` mục "Quyết định 08/09" |
| Người gọi từng bước trong chat, review diff, bảo "commit" | **Operator session** (ai-employees §3a: *the member directing an interactive agent in chat*) | **Đã có** — và ở Phase 1 đây là **chế độ duy nhất** | toàn bộ lịch sử SETUP §9 |
| — | **Task status / lifecycle** dạng dữ liệu | **Thiếu** | không có |
| — | **Approval** dạng dữ liệu (ai · lúc nào · cổng nào · điều kiện) | **Thiếu** | P10 chỉ có "chốt 06/09" |
| — | **Learning debt** (nợ học tích luỹ) | **Thiếu** | HOC.md không có |
| — | **Rút bài học theo từng task** | **Thiếu** | SETUP §9 ghi *quyết định*, không ghi *điều học được* |
| — | **PROJECT manifest** — repo ở đâu, lệnh gì, nhánh nào | **Thiếu, và không nên nằm trong Portage** | CLAUDE.md là luật *của Portage*, đúng chỗ rồi |
| — | Cách gọi routine từ harness | **Thiếu** | không có `.claude/` |

**Kết luận §1:** cái thiếu không phải *nội dung* — Portage có luật, có log, có plan,
có test canh kiến trúc, có bài học có ngày. Cái thiếu là **hình dạng dữ liệu**:
trạng thái, duyệt, nợ học là văn xuôi rải trong bốn file, nên không có gì đọc lại
được bằng máy, và không mang sang project khác được.

---

## 2. ai-employees có gì đáng học?

### 2.1 Nó là gì, nói ngắn

Một bộ **tám employee**, mỗi cái là **một thư mục file Markdown** — không có runtime,
không có database — chạy trên **scheduler của máy người dùng** (Task Scheduler,
launchd, cron), mỗi routine một `SKILL.md`, đọc và ghi file thường, gộp về một
"brief buổi sáng". Được xây để **chạy không người trông 90 ngày** cho việc
marketing/vận hành: quét web, soạn thư, điền form, đọc bảng quảng cáo. Kiến trúc
được đúc từ routine thật đang chạy, và mỗi luật đều ghi *ngày nó được học ra từ
sự cố nào*.

Năm luật cốt lõi, và ba luật thêm sau tuần đầu chạy thật:

| Luật | Nội dung một dòng |
|---|---|
| LAW 1 | Tự lực tối đa: *quyết thay vì đề xuất, sửa thay vì báo*. Chỉ dừng cho **hành động hướng ngoại đang bị giữ** hoặc **credential** |
| LAW 2 | Hai guardrail: (a) hành động hướng ngoại **giữ mặc định**, người **nhả từng kênh** trong `RELEASES.md` — file chỉ người viết; (b) credential — **luôn bật**, không có nhả |
| LAW 3 | Routine gọi **capability**, không gọi tool. `CAPABILITIES.md` ánh xạ sang route theo từng harness |
| LAW 4 | Không rải skill vào thư mục global |
| LAW 5 | **Mọi luật khó nhọc mới học được phải có ngày và nằm trong file** |
| LAW 6 | *Verify before you block*: trước khi báo bị chặn, bỏ ba phút kiểm xem thật sự còn chặn không |
| LAW 7 | Mang việc tới trước mặt người: form đã điền, tab mở sẵn, một click cuối là của người |
| LAW 8 | *Tick ghi nhận đồng ý; routine làm phần cơ học* |

### 2.2 Mâu thuẫn thật giữa ai-employees và prompt — và cách tôi phân xử

**ai-employees nói** (`CONTRACT §8.3`): *"Không viết cổng duyệt vào file hướng dẫn.
Harness đã quyết định agent được ghi file hay không, đó là kiểm soát bằng phần mềm.
Một cổng thứ hai bịa ra trong Markdown không thêm an toàn, chỉ thêm ma sát."*

**Prompt nói** (§12): *"Duyệt của người phải là khái niệm hạng nhất: Plan Approval,
Commit Approval…"*

Cả hai đều đúng — **về hai loại hành động khác nhau.** ai-employees đúng về *ghi
file*: harness đã hỏi, hỏi lần hai là thừa. Nhưng `git commit` không phải "ghi
file" — nó ghi vào **lịch sử dùng chung**; `git push` ghi vào **lịch sử người khác
thấy**; `merge` ghi vào **nhánh production**. Harness không biết sự khác nhau đó.
Và chính ai-employees cũng biết: LAW 2 **giữ mọi hành động hướng ngoại**, và
`web-dev-employee` ROLE §1.1 nói thẳng *"đầu ra của việc rủi ro là một thay đổi
review được trên nhánh, không bao giờ là push thẳng vào production."*

**Phân xử:** với một employee làm việc trên code, **commit / push / merge chính là
"outbound action"**. Chúng không phải cổng bịa ra — chúng là guardrail 1 của
ai-employees, dịch sang Git. Và cơ chế nhả cũng dịch theo: người **nhả từng kênh**
trong một file chỉ người viết. Còn **plan approval** — thứ ai-employees không có
vì họ không cần học — ở đây là cổng của **mục tiêu học**, và nó cũng nhả được khi
lòng tin đủ. Vậy prompt đúng ở *có cổng*, ai-employees đúng ở *cổng nào và nhả
bằng gì*.

### 2.3 Keep · Modify · Defer · Remove

Mọi khái niệm prompt §4 liệt kê, cộng những cái tôi tìm thấy khi đọc:

| Khái niệm | Quyết định | Vì |
|---|---|---|
| **Thang ưu tiên** (ROLE §0): *file luật của project* thắng kit trên mọi thứ về project, trừ guardrail | **Keep** nguyên | Giải quyết luôn vấn đề "CLAUDE.md trộn hai thứ" (§3): luật Portage **ở lại** Portage và **thắng**; employee chỉ giữ guardrail |
| **Hai guardrail + `RELEASES.md`** (LAW 2) | **Keep**, dịch sang Git | commit/push là outbound; người nhả từng kênh trong file chỉ người viết; **merge không bao giờ là kênh** |
| **`## Corrections`** ở đuôi mọi file, chỉ người viết, thắng file nó nằm trong | **Keep** nguyên | Rẻ nhất trong toàn bộ repo. Portage đã làm dạng thô ở "Ba bài học" |
| **`PAUSED`** — file rỗng dừng mọi routine, routine không được đụng | **Keep** nguyên | Không tốn gì; bắt buộc khi có scheduler (Phase 6) |
| **LAW 5** — luật khó nhọc có ngày, nằm trong file | **Keep** — Portage đã có | `SETUP.md §9` chính là nó |
| **LAW 6** — verify before you block | **Keep**, dịch: trước khi đặt `BLOCKED`, chạy lại test / đọc lại file / xem người đã trả lời chưa | Chặn loại lỗi "báo bị chặn trong khi đã hết chặn" |
| **Cách ly** (ROLE §1.1): làm trên nhánh · *không đụng working tree bẩn, không stash* · *đang ở nhánh không phải của mình → người đang làm dở, đổi gì cũng không* | **Keep** nguyên văn | Là guardrail riêng của coding mà prompt không có. Xoá việc dở của người là tai nạn thật |
| **"Gate hỏng là run thành công"** (fix-runner §6b): để nhánh y nguyên, ghi `gate-failed`, **không thử cách khác**, không push | **Keep** | Sửa lại chính thiết kế trước của tôi: test đỏ là *kết quả*, không phải *task thất bại* |
| **Đọc project trước khi đọc việc** (fix-runner Step 3): đọc file luật, đọc `docs/`, ghi `docs_read` có ngày | **Keep** | Là thứ làm employee dùng lại được trên repo khác |
| **Scan secret trước commit**; commit message không chứa token, stack trace, log | **Keep** | — |
| **Run record**: một schema, **từ vựng trạng thái đóng** ("không có giá trị thứ chín"), và danh sách **"không bao giờ xuất hiện"** (secret, diff, stack trace, tên người) | **Keep** kỷ luật; **Modify** định dạng | Từ vựng đóng chặn việc máy bịa trạng thái. Định dạng: Markdown+frontmatter Phase 1 (§5.5) |
| **Từ vựng cho việc không biết** (ROLE §9): `n/a (lý do)`, `not measured`; *"zero và null khác nhau; im lặng không phải pass"* | **Keep** | Khớp đúng luật "nói rõ CHƯA test gì" của Portage |
| **Invariant cuối run** (ROLE §11): bốn điều phải đúng, sai một là run thất bại dù đã làm gì | **Keep**, dịch (§5.3) | Portage "Trước khi bảo xong: ba kiểm" là bản thô |
| **Audit cơ học trước khi ship** (STANDARD §6): tên thư mục = tên YAML; không tham chiếu routine không tồn tại; không tên tool trong thân routine; không chuỗi giống credential | **Keep** → `scripts/verify-agent.sh` | Cùng triết lý `verify-numbers.sh` / `verify-snippets.py` của `learn/` |
| **Một người ghi cho file ghi đè; appender có tên cho ledger; file không ai đọc thì cắt** | **Keep** | — |
| **Change brief** (ví dụ thật `changes/2026-03-05-fix-C-005.md`): *"Mọi dòng dưới đây mô tả một thay đổi trên nhánh. Chưa merge. Merge là của bạn."* + Files · Gate · Rollback · Left for you | **Keep** làm khuôn `review.md` + thân PR | Mẫu tốt hơn cái tôi định viết |
| **Operator session** là actor thứ ba (§3a), phải để lại vết như routine | **Modify**: ở Phase 1 đây là actor **duy nhất** | Routine theo lịch tới Phase 6 |
| **LAW 1** tự lực tối đa, *quyết thay vì đề xuất* | **Modify** | Đúng cho ops 90 ngày không người trông. Sai cho project **học**: người *muốn* được hỏi (Challenge). → Tự lực **bên trong** plan đã duyệt; plan thì **đề xuất**, và cổng plan **nhả được** theo project khi đủ tin |
| **LAW 3** capability không tool + bảng route theo harness | **Modify** | Giữ kỷ luật đặt tên trong thân routine (`test.run`, không `go test`). **Không** dựng bảng route đa harness ở Phase 1 — một harness. `PROJECT.md commands.*` là ánh xạ duy nhất cần: thứ đổi theo *project*, không theo *harness* |
| **Bậc quyền (rung) + nới bằng bằng chứng** (`web-guardrail-review`): lớp thay đổi nào được làm không hỏi; **nới một bậc sau N tháng merge sạch liên tiếp, siết ngay khi một lần bị sửa**; ranh giới ngoài không bao giờ nới | **Modify → Phase 7** | Đúng là cách "nới cổng bằng số liệu" tôi mơ hồ đề xuất — họ đã làm cụ thể. Nguyên lý **bất đối xứng** (nới chậm, siết ngay) áp dụng từ ngày một |
| **Step 0 cố định** cho mọi routine | **Modify**: bỏ window/period/browser; thay bằng *pause → đọc luật project + Corrections → xác nhận đúng repo, tree sạch, đúng nhánh* | — |
| **Self-edit `SKILL.md`** + changelog chứa **nguyên văn đã thay** làm undo (§8.3) | **Defer → Phase 7**; **giữ ngay** một câu: *"tự sửa có thể làm việc được phép tốt hơn, không bao giờ nới cái được phép"* | Vòng lặp compounding, nhưng chưa có dữ liệu để tin |
| **SCHEDULE.md**, window, period key, once-per-period guard | **Defer → Phase 6** | Chạy LLM theo lịch khi không ai ngồi đó **mâu thuẫn** với *human-controlled* Phase 1 |
| **Chief of Staff** đọc-only | **Defer → Phase 8** | Chính họ nói: có giá trị khi có *đội*. Một employee + `grep WAITING` là đủ |
| **Recipes** — "học trên máy, không ship" | **Defer** tới khi một trình tự lặp 3 lần trong run log | — |
| **Push notification** (§2.3) | **Defer**, có lẽ Remove | Một người, ngồi tại bàn |
| **Nâng cấp kit / contribution draft** (§8.5) | **Remove** | Không ship sản phẩm |
| **Browser lane, mutex, save test, bảy nút bị cấm, luật LinkedIn** | **Remove** cơ chế; **giữ nguyên lý** | Không có browser. Nguyên lý *"nhãn không phải câu hỏi — thứ nút đó cam kết mới là"* dịch thành: `commit` cam kết vào lịch sử local (đi tiếp sau gate) · `push` cam kết vào lịch sử chung (giữ) · `merge` (không bao giờ) |
| **Dashboard / LAW 7 mang việc tới trước mặt** | **Remove** cơ chế; **giữ ý** | Deliverable là **diff review được + thân PR sẵn dán**, không phải file phải đi tìm |
| **Nhiều employee, fleet, prefix cho scheduler chung** | **Remove** | Một employee |
| **Trường chi phí trong run record** | **Remove** Phase 1 | — |
| **Luật văn phong (không em dash), đóng gói zip, `INSTALL-PROMPT`, `npx` installer** | **Remove** | Của một sản phẩm bán; không liên quan |

---

## 3. Coding Agent thực sự cần gì?

Lấy §1 trừ đi cái đã có, cộng §2 những cái đáng học, ra danh sách **cần mà chưa có**:

1. **Hình dạng dữ liệu cho Task** — trạng thái, bước, duyệt — thay cho văn xuôi
   trong P9/P10-PLAN. Đây là chỗ DDD có nghĩa thật (§5).
2. **Tách CLAUDE.md thành hai thứ nó đang trộn:** *luật của Portage* (panic vs
   error, luật vàng import, số nghiệp vụ) **ở lại Portage** — và **thắng**, theo
   thang ưu tiên ROLE §0. *Cách employee làm việc* (không tự commit, kiểm ba thứ
   trước khi bảo xong, tự phản biện trước khi đề xuất) **ra** `CONTRACT.md` dùng
   chung. Không chép — **trỏ**.
3. **`PROJECT.md`** — thứ duy nhất đổi theo project: đường dẫn, lệnh test/lint/fmt,
   nhánh được bảo vệ, file luật nào, kênh nào đã nhả. Đây là ranh giới **Agent
   (chung) / Project (riêng)** mà yêu cầu dùng lại bắt buộc phải có, và là bằng
   chứng cho ROLE §0.
4. **Rút bài học theo từng task + nợ học** — cái Portage thiếu và ai-employees không
   có (họ không cần học). Phải thiết kế để **không** thành gánh nặng tài liệu (§5.6).
5. **Guardrail cách ly Git** nguyên văn từ ai-employees: nhánh riêng, không đụng
   tree bẩn, gate hỏng thì để yên, scan secret trước commit.
6. **Bốn điều phải đúng cuối mỗi run**, ghi thành record — nâng "ba kiểm" của
   CLAUDE.md từ lời nhắc thành bằng chứng.
7. **Một cách gọi từ harness** — hiện không có `.claude/`; một skill trỏ vào
   `routines/<id>/SKILL.md` là đủ.
8. **`scripts/verify-agent.sh`** — audit cơ học: routine nào được tham chiếu
   phải tồn tại, thân routine không có tên tool, không chuỗi giống credential.

Và ba thứ **không cần**, dù prompt gợi: runtime Go (Phase 4), database, scheduler.

### 3.1 Dùng lại được — ba lời hứa, và cách kiểm từng lời

Anh nhấn mạnh hai lần: kiến trúc phải dùng cho **mọi** project sau, không riêng Go.
Một câu như vậy chỉ có giá khi **kiểm được**. Ba lời hứa, mỗi lời một phép kiểm chạy
bằng máy hoặc bằng một phase:

| Lời hứa | Nghĩa là | Kiểm bằng |
|---|---|---|
| **P1. `agent/` không chứa một chữ nào của riêng project nào** | không `go test`, không `gofmt`, không `phpunit`, không `CLAUDE.md`, không đường dẫn tuyệt đối, không số nghiệp vụ — ở mọi file ngoài `projects/` | `scripts/verify-agent.sh` mục (2) và (5): grep danh sách cấm, **đỏ là chưa xong** |
| **P2. Thêm một project = thêm đúng một file** | `projects/<tên>/PROJECT.md`; không sửa `routines/`, `CONTRACT`, `ROLE`, `CAPABILITIES` | Phase 3: một task PHP tới `DONE` với `git diff --stat employee/ -- ':!projects'` **rỗng** |
| **P3. Mọi thứ đổi theo project đi qua một schema duy nhất** | `PROJECT.md` là bề mặt biến thiên **duy nhất**; routine chỉ đọc trường, không đọc nội dung repo để suy đoán | schema dưới đây đóng; muốn thêm trường phải qua ADR |

**Schema `PROJECT.md` — thiết kế cho trường hợp tổng quát ngay từ đầu**, không phải
cho Portage rồi vá:

```yaml
name:      <slug>                # tên thư mục projects/<slug>/
path:      <đường dẫn tuyệt đối tới repo — nơi DUY NHẤT đường dẫn tuyệt đối được xuất hiện>
rules:     <file luật của repo, tên gì cũng được: CLAUDE.md · AGENTS.md · CONTRIBUTING.md · .cursorrules>
docs:      <thư mục docs của repo, có thể rỗng>
commands:                        # mọi giá trị là chuỗi shell; employee không hiểu nghĩa, chỉ chạy
  test:    <lệnh>                # bắt buộc
  fmt:     <lệnh>                # tuỳ chọn — thiếu thì verify ghi `fmt: n/a (no command)`
  lint:    <lệnh>                # tuỳ chọn
  build:   <lệnh>                # tuỳ chọn
branches:
  default:   <nhánh>
  protected: [<nhánh>, …]
  prefix:    agent/
releases:  []                    # kênh đã nhả, chỉ người viết: - {channel: plan|push, since: <ngày>, note: <điều kiện>}
```

Hai ví dụ đặt cạnh nhau để thấy **employee không phân biệt được chúng**:

```yaml
# projects/portage/PROJECT.md              # projects/fastboy-api/PROJECT.md
name: portage                              name: fastboy-api
path: /home/…/portage                      path: /home/…/fastboy-api
rules: CLAUDE.md                           rules: AGENTS.md
docs: docs/                                docs: docs/
commands:                                  commands:
  test: go test ./...                        test: vendor/bin/phpunit
  fmt:  gofmt -l .                           fmt:  vendor/bin/php-cs-fixer fix --dry-run
  lint: go vet ./...                         lint: vendor/bin/phpstan analyse
branches: {default: main,                  branches: {default: develop,
  protected: [main], prefix: agent/}           protected: [develop, main], prefix: agent/}
releases: []                               releases: []
```

Routine `verify` chạy `commands.test` và ghi exit code. Nó **không biết** một bên là
Go, một bên là PHP — và đó chính là bằng chứng P1.

**Cái gì sẽ phá dùng-lại, phải canh từ Phase 1:**

- Một `SKILL.md` viết *"chạy `go test ./...`"* thay vì *"chạy `PROJECT.commands.test`"*.
  → `verify-agent.sh` (2) bắt.
- `CONTRACT.md` có câu *"domain không được import yaml"* — đó là luật **của Portage**,
  phải ở `CLAUDE.md` của Portage. → `verify-agent.sh` (5) bắt tên ngôn ngữ; phần
  còn lại do review của người.
- `knowledge/go/` bị nhét bài về Portage. → bài về một repo cụ thể đi vào
  `projects/<tên>/knowledge/`, không vào `knowledge/` chung.
- Phase 4 runtime Go đọc `commands.test` rồi *parse output theo định dạng `go test`*.
  → runtime chỉ được đọc **exit code và stdout thô**; hiểu định dạng là việc của
  `PROJECT.md` nếu cần (`commands.test_report: <lệnh ra JSON>`), qua ADR.

**Điều gì thì KHÔNG hứa:** employee không dùng lại được cho project mà *người* chưa
viết `PROJECT.md`. Nó không tự đoán lệnh test. Đoán là chỗ ai-employees LAW 1 muốn
tự lực, và là chỗ tôi cố ý không theo — đoán sai lệnh test trên repo lạ là chạy
thứ không ai bảo chạy.

---

## 4. Giữ · bỏ · hoãn — kể cả sửa lại prompt

Anh bảo *"đừng cố chứng minh mọi thứ trong prompt đúng"*. Audit cho thấy tám chỗ
prompt sai hoặc over-engineered, và bằng chứng cho từng chỗ:

| Prompt nói | Audit cho thấy | Sửa thành |
|---|---|---|
| Approval là aggregate (§27, bảng khái niệm) | ai-employees: approval là **một dòng người viết** trong `RELEASES.md` + hold trong CONTRACT. Portage P10: một dòng *"chốt 06/09"*. Cả hai nơi không ai cần đối tượng riêng | Approval = **value object trong Task** (ai · lúc nào · cổng · điều kiện) + **Release** = dòng trong `PROJECT.md` chỉ người viết |
| Employee là domain entity | ai-employees: employee **là thư mục file**. Không trạng thái, không hành vi | `ROLE.md` + `CONTRACT.md`. Cấu hình |
| Vòng đời 8 trạng thái tuần tự (`TODO → ANALYZING → … → DONE`) | ai-employees dùng **từ vựng đóng** cho *run*, không cho task. Portage P10 **không có trạng thái nào** và vẫn chạy được 20 đợt — vì người biết bóng ở đâu | Tách **`status`** (bóng ở đâu, 7 giá trị đóng) khỏi **`step`** (routine nào). §5.4 |
| Sáu cổng duyệt | ai-employees: **hai** guardrail, kênh nhả được. Implementation ≡ Plan; Merge/Production không phải kênh vì **không bao giờ làm** | **Ba** hold: plan · commit · push. Plan và push **nhả được theo project**. Merge không có cổng vì không có hành động |
| `runlog.jsonl` | Portage: `SETUP.md §9` — 20 đợt Markdown, người đọc lại được. ai-employees: jsonl vì **có scheduler và có script tổng hợp** | Phase 1 **Markdown + frontmatter**, một file mỗi run. `jsonl` là chỉ mục sinh sau nếu Phase 6 cần |
| `agent.execute → Claude Code Adapter / OpenClaw Adapter` | ai-employees chạy trên **11 harness** mà **không có adapter nào**: routine là Markdown, `HARNESSES.md` là *bảng* cách gọi | Độc lập harness bằng **file thường**. Phần harness-specific gom vào **một** skill gọi. Adapter chỉ khi runtime Go (Phase 4) cần *gọi* harness |
| Cây Go `internal/{domain,application,infrastructure,interfaces}` | Portage chạy thật với `internal/{domain,app,adapter,platform}` và không có tầng `application/` tách biệt `interfaces/`. ai-employees **không có runtime** | Phase 4: `task/ · project/ · store/ · routine/ · git/ · cli/`. Không `application/`, không `interfaces/` (§5.7) |
| Tám routine: task-discovery · domain-analysis · planning · implementation · testing · review · learning · daily-review | `web-fix-runner` làm *analyze → sửa → gate → brief → push* trong **một** routine, dừng đúng ở hold. Nhưng đó là cho ops; **học** cần điểm dừng giữa các bước | **Sáu bước** Phase 1 (intake · analyze · plan · implement · verify · review), mỗi bước một `SKILL.md`, **người gọi từng bước**. Phase 6/7 gộp thành một run dừng ở hold. `learning` nằm **trong** review; `daily-review`, `task-discovery` hoãn |
| Learning debt là danh sách `[ ]` | Không có ở đâu cả — ai-employees không học, Portage HOC.md là giáo trình | Thiết kế **mới**, nhưng có ba luật chặn gánh nặng (§5.6) |

Chỗ **prompt đúng và audit xác nhận**: modular monolith · filesystem · một
employee · không Kafka/Redis/K8s · Chief of Staff hoãn · "mọi feature dạy ba thứ".

---

## 5. Go + DDD model nào thực sự có nghĩa?

### 5.1 Bảng khái niệm — sau phản biện

| Khái niệm | Trách nhiệm | Loại | Bằng chứng nó là loại đó |
|---|---|---|---|
| **Task** | một đơn vị việc, có trạng thái, có duyệt | **Domain — aggregate root duy nhất** | có danh tính, sống qua đổi trạng thái, mọi invariant đều nói về nó |
| **Approval** | người cho phép một cổng | **Value object trong Task** | ai-employees + P10: một dòng có ngày là đủ |
| **Release** | người nhả một kênh cho một project | **Dòng trong `PROJECT.md`**, chỉ người viết | LAW 2 nguyên văn |
| **Run** | một lần chạy một bước | **Record chỉ ghi thêm** | là sự thật quá khứ |
| **Project** | repo được làm việc trên | **Cấu hình + ranh giới** | ROLE §0: luật của nó thắng |
| **Employee** | ai, luật, khả năng | **Cấu hình** | ai-employees: thư mục file |
| **Routine** | cách làm một bước | **Vận hành** (`SKILL.md`) | domain không biết nó tồn tại |
| **Capability** | thứ routine cần | **Từ vựng đặt tên**, chưa là port | một harness → không có gì để ánh xạ |
| **Learning item** | điều học được từ một task | **Record**, gắn vào Task | không có invariant |
| **Change class / Rung** | lớp thay đổi và bậc quyền | **Phase 7** — value object của policy | `web-guardrail-review` |
| **Knowledge · ADR · Recipe** | tài liệu bền | **Hỗ trợ** | — |

### 5.2 Hai bounded context — vì yêu cầu dùng lại

```
┌──────────────── AGENT (dùng chung) ────────────────┐
│ ROLE · CONTRACT · CAPABILITIES · routines/            │
│ knowledge/{go,ddd,agent}/ · learning/debt.md          │
│ decisions/ · logs/runs/ · PAUSED                      │
│                                                        │
│ Không biết "go test" tồn tại. Không nhắc tên ngôn ngữ. │
└───────────────────────┬────────────────────────────────┘
                        │ đọc PROJECT.md, đọc PROJECT.rules
┌───────────────────────▼────────────────────────────────┐
│ PROJECT (riêng từng repo) — projects/<tên>/            │
│ PROJECT.md: path · rules · commands · branches         │
│             · releases (kênh đã nhả)                   │
│ tasks/<id>/ · knowledge/ (riêng)                       │
│                                                        │
│ Luật của repo (CLAUDE.md của Portage) Ở LẠI trong repo │
│ và THẮNG employee trên mọi thứ về repo — trừ guardrail │
└────────────────────────────────────────────────────────┘
```

Chữ "task" cùng nghĩa ở cả hai — nên đây là hai **phạm vi trách nhiệm**, chưa phải
hai *ngôn ngữ*. Nó thành bounded context thật khi Phase 3 chạy trên repo Symfony
và chữ "test" bắt đầu nghĩa `phpunit`.

### 5.3 Aggregate Task và mười một invariant

Bảy invariant của Task (kiểm khi **đổi trạng thái**) và bốn invariant của Run
(kiểm **cuối mỗi run**, theo ROLE §11). Phase 1 routine + người giữ; Phase 4 code ép.

**Của Task:**

| # | Invariant | Vì |
|---|---|---|
| T1 | không vào `step: implement` khi `approvals` chưa có `plan` — **trừ khi** `PROJECT.releases` có `plan` | sửa code chưa ai đồng ý, trừ project đã đủ tin |
| T2 | không `DONE` khi thiếu **cả ba**: `review.md` · `VerifyResult` cuối exit 0 · `approvals.commit` | "xong" phải có bằng chứng |
| T3 | không push khi `approvals` chưa có `push` — **trừ khi** `PROJECT.releases` có `push` | push là điểm không quay đầu xã hội |
| T4 | `REJECTED` chỉ từ `WAITING_HUMAN`, và không đi tiếp | từ chối là quyết định của người; làm lại thì mở task mới |
| T5 | Task thuộc đúng một Project; mọi đường dẫn nó chạm nằm trong `PROJECT.path` | employee đa project nhầm repo là tai nạn |
| T6 | nhánh làm việc = `PROJECT.branches.prefix` + id; **không bao giờ** trong `protected` | ROLE §1.1 |
| T7 | `plan.md` **không đổi** sau `approvals.plan`; muốn đổi → về `WAITING_HUMAN` | duyệt plan này rồi làm plan khác là duyệt hư |

**Của Run** (bốn điều, sai một là run thất bại dù đã làm gì):

| # | Cuối mỗi run phải đúng |
|---|---|
| R1 | không merge, không push khi chưa nhả, không `--force`, không xoá nhánh, không sửa `CONTRACT.md`/`PROJECT.md`/`PAUSED` |
| R2 | mọi con số ghi ra run này đều đếm được từ file hoặc output *của run này*, kèm nguồn; không có thì `not measured` — không bao giờ `0` cho thứ chưa đo |
| R3 | đúng **một** run record được ghi thêm |
| R4 | không credential, token, connection string, hay giá trị biến môi trường nào được ghi, in, hay log ở bất kỳ đâu |

### 5.4 Vòng đời — `status` tách khỏi `step`

```yaml
status:      TODO | IN_PROGRESS | WAITING_HUMAN | DONE | BLOCKED | REJECTED | FAILED
step:        analyze | plan | implement | verify | review        # chỉ khi IN_PROGRESS
waiting_for: plan-approval | commit-approval | push-approval | question
verify:      passed | gate-failed | not-run                       # kết quả gate gần nhất
```

**Từ vựng đóng — không có giá trị thứ tám.** Bốn tình huống tưởng cần trạng thái
riêng ánh xạ như sau, không thương lượng (học từ `CONTRACT §4.1`):

- Test đỏ → `status` giữ nguyên, `verify: gate-failed`, **nhánh để y nguyên, không
  thử cách khác**. Gate hỏng là routine làm đúng việc của nó.
- Đang chờ người trả lời câu hỏi (Challenge) → `WAITING_HUMAN / question`.
- Gặp thứ ngoài luật → `BLOCKED` kèm lý do; **trước đó phải verify** (LAW 6).
- Không làm được gì cả → `FAILED`. Hiếm.

```
TODO ─intake─▶ IN_PROGRESS(analyze) ─▶ IN_PROGRESS(plan) ─▶ WAITING_HUMAN(plan-approval)
                    │ challenge: true                              │ duyệt (hoặc đã nhả)
                    ▼                                              ▼
           WAITING_HUMAN(question)                    IN_PROGRESS(implement) ─▶ (verify) ─▶ (review)
                                                                                              │
                                                                     WAITING_HUMAN(commit-approval)
                                                                                              │ duyệt
                                                                                        commit trên agent/…
                                                                                              │
                                                                  WAITING_HUMAN(push-approval) hoặc đã nhả
                                                                                              │
                                                                                            DONE
```

**Ai kéo cần:** employee kéo mọi chuyển *trừ* `WAITING_HUMAN → IN_PROGRESS` và
`→ REJECTED` — hai cái đó **chỉ người**, và người ghi một dòng vào `approvals`.

### 5.5 Lưu trữ

| | Filesystem (MD + frontmatter) | SQLite | PostgreSQL |
|---|---|---|---|
| Người đọc bằng mắt · Git diff · review trong PR | **có** | không | không |
| Sống qua đổi harness | có | có | cần server |
| Nhiều run ghi cùng lúc | **không** | có | có |
| Bằng chứng trong repo này | `SETUP §9` 20 đợt vẫn đọc được | — | cổng 5433 từng bị chiếm (SETUP §9 đợt 20) |

**Phase 1: filesystem.** Đúng cả prompt (§33 *"file là nguồn sự thật"*), cả
ai-employees (không runtime), cả kinh nghiệm Portage. Giới hạn ghi thẳng: không
khoá → Phase 6 phải xử (file lock hoặc một run tại một thời điểm). Đường ra:
SQLite làm **chỉ mục dựng lại được**, file vẫn là gốc.

### 5.6 Hệ học — thiết kế mới, có ba luật chặn gánh nặng

Đây là phần **không có ở đâu để chép**: ai-employees không học (họ *sửa quy trình*,
không học *khái niệm*), HOC.md là giáo trình. Nên thiết kế cẩn thận nhất:

```
review.md (nhìn diff còn nóng) ──▶ learning.md (≤ 12 dòng, mẫu cố định)
        ──▶ learning/debt.md (+ dòng nợ) ──▶ knowledge/ (CHỈ khi một nợ được gạch)
```

Mẫu `learning.md`: **Go** ≤ 3 dòng, mỗi dòng trỏ `file:dòng` vừa viết · **DDD** ≤ 3 ·
**Kỹ thuật** ≤ 2 · **Agent** ≤ 2 (điều gì về chính employee vừa lộ ra) · **Sai ở đâu**
· **Nợ mới**.

**Ba luật chặn "hệ học thành gánh nặng":** (1) không `learning.md` nào không gắn
task thật; (2) `knowledge/` **chỉ nhận** bài khi một dòng `debt.md` được gạch —
không nhận bài viết trước; (3) nợ 30 ngày không ai đụng → `review` gợi xoá.

**Ba chế độ là hai cờ trên Task, không phải mode của hệ:** `challenge: true` →
`analyze` hỏi 3–5 câu rồi **dừng** (`WAITING_HUMAN/question`) — đúng P9-PLAN §4
"chốt rồi làm", thành dữ liệu. `mode: learning` → `plan` viết gợi ý, `implement`
**không chạy**, `review` review code *của người*.

### 5.7 Kiến trúc Go — Phase 4, mô tả trước, **không dựng**

```
cmd/agent/            một binary CLI
internal/
  task/             Task · TaskID · Approval · VerifyResult · 7 invariant · bảng chuyển trạng thái
  project/          PROJECT.md → struct; kiểm T5, T6; đọc releases
  store/            đọc/ghi frontmatter TASK.md, logs/runs — MỘT cài đặt: filesystem
  routine/          điều phối: gọi harness, ghi run, đổi trạng thái Task, kiểm R1–R4
  git/              bọc git: branch, diff, commit; từ chối push khi thiếu release/approval
  cli/              agent task new|approve|reject|status
```

Khác cây prompt: **không** `application/`, **không** `interfaces/` — đúng như Portage
đang chạy với `app/ · adapter/ · platform/`. Chiều phụ thuộc `cli → routine →
{task, project, store, git}`; `task` không import ai. Interface **chỉ khi có bản giả
cho test** (bài Mùa 1 tập 5). Invariant thành test đỏ khi vi phạm — **đúng triết lý
`decisions_test.go`** của Portage: *"a rule that lives in a test fails the build"*.
Dựng **chỉ khi** ≥ 20 task thật đã qua Phase 1–3.

---

## 6. Phase 1 chính xác build cái gì?

### 6.1 Ở đâu

Anh nói *"không tạo repo/module riêng ở phase này"* và *"sau này có thể tách"*.
Vậy Phase 1 đặt tại **`agent/`** ở gốc Portage — **một thư mục, không `go.mod`,
không code** — cây được thiết kế để `git mv` ra repo riêng **không sửa một dòng**:
mọi đường dẫn bên trong là tương đối, đường dẫn tới Portage nằm **duy nhất** trong
`projects/portage/PROJECT.md`. Chưa tạo thư mục đó cho tới khi bản này duyệt.

### 6.2 Cây thư mục — chỉ những gì cần

```
agent/
├── ROLE.md                    ai · làm gì · không làm gì · thành công là gì
├── CONTRACT.md                guardrail + cách làm việc — không nhắc tên ngôn ngữ nào
├── CAPABILITIES.md            từ vựng: file.read · file.write · code.search · shell.run
│                              · git.branch · git.diff · git.commit · test.run
├── PAUSED                     (chỉ khi người tạo) — rỗng: dừng hết; ghi id: dừng từng cái
├── routines/
│   ├── intake/SKILL.md        lời mô tả việc → TASK.md
│   ├── analyze/SKILL.md       đọc project trước · nếu challenge: hỏi rồi dừng
│   ├── plan/SKILL.md          kế hoạch + mục "Không làm" + tự phản biện → hold plan
│   ├── implement/SKILL.md     trên nhánh agent/ · không vượt plan · không đụng tree bẩn
│   ├── verify/SKILL.md        chạy PROJECT.commands.* · nói rõ CHƯA test gì
│   └── review/SKILL.md        Go · DDD · kỹ thuật + learning.md + change brief → hold commit
├── projects/
│   └── portage/
│       ├── PROJECT.md         path · rules: CLAUDE.md · commands · branches · releases
│       ├── tasks/T-001-ratecard-yaml/{TASK,analysis,plan,review,learning}.md
│       └── knowledge/         riêng Portage: invariant, quyết định (trỏ SETUP §9)
├── knowledge/{go,ddd,agent}/  chung — rỗng lúc đầu, chỉ đầy khi gạch nợ
├── learning/debt.md
├── decisions/ADR-001…003.md
├── logs/runs/                 một file mỗi run, chỉ ghi thêm
└── scripts/verify-agent.sh audit cơ học (§6.7)
```

**Bỏ, và vì sao:** `SCHEDULE.md` (Phase 6) · `state/` `queue/` (trạng thái nằm
trong `TASK.md`) · `reports/` (Phase 8) · `recipes/` (chưa lặp 3 lần) ·
`task-discovery` tự quét (lòng tin chưa có) · `daily-review` (cần lịch hoặc người nhớ).

### 6.3 `CONTRACT.md` — mở đầu bằng thang ưu tiên

```
1. File luật của project (PROJECT.rules) — thắng trên mọi thứ về project đó,
   trừ mục 2 và 3 dưới đây.
2. Hai guardrail — không file nào, kể cả PROJECT.md, nới được:
   (a) hành động hướng ngoại giữ mặc định: commit · push. Merge KHÔNG là kênh.
   (b) credential: không đọc .env vào log, không in token, không ghi giá trị
       biến môi trường — chỉ ghi TÊN biến.
3. Cách ly: nhánh agent/<id>-<slug> · không đụng working tree bẩn, không stash
   · đang ở nhánh không phải của mình → người đang làm dở, đổi gì cũng không
   · không --force · không xoá nhánh · không sửa CONTRACT/PROJECT/PAUSED.
4. CONTRACT.md này. 5. CAPABILITIES.md. 6. ROLE.md. 7. SKILL.md của routine.
Ở mọi tầng: một dòng trong ## Corrections của file đó thắng file đó.
```

Rồi: **verify before you block** · **gate hỏng là run thành công** · **tự phản biện
trước khi đề xuất** (bài học 1 của CLAUDE.md, nâng thành luật) · **bốn điều cuối
run R1–R4** · **từ vựng cho việc không biết** · **scan secret trước commit** ·
*"tự sửa có thể làm việc được phép tốt hơn, không bao giờ nới cái được phép"*.

`CONTRACT.md` **không nhắc** Go, PHP, `go test`, Symfony. Cái đó ở `PROJECT.md`.

### 6.4 `PROJECT.md` — ranh giới dùng lại

```yaml
name: portage
path: /home/duongvantiensy/portage
rules: CLAUDE.md                     # employee ĐỌC ở Step 0, không chép; thắng trên mọi thứ về Portage
docs: docs/                          # đọc index + file khớp vùng task đụng (fix-runner Step 3)
commands:
  test: go test ./...
  fmt:  gofmt -l .
  lint: go vet ./...
branches:
  default: main
  protected: [main]
  prefix: agent/
releases: []                         # rỗng = giữ hết. Người nhả: - {channel: push, since: 2026-10-01, note: "…"}
```

Cho một repo Symfony ở Fastboy, đổi đúng bốn dòng `commands`. Agent **không
đổi một dòng**. Đó là Phase 3, và là điều kiện để tin DEC-003.

### 6.5 Step 0 của mọi routine — cố định, năm dòng

```
0.0  PAUSED tồn tại? → ghi run record `skipped-paused`, thoát.
0.1  Đọc PROJECT.md → đọc file PROJECT.rules trỏ tới → đọc ## Corrections của mọi file
     sẽ đụng. Ghi docs_read có ngày vào TASK.md.
0.2  Xác nhận: đúng PROJECT.path · git status sạch (không sạch → BLOCKED "người đang làm
     dở", đổi gì cũng không) · nhánh hiện tại là agent/<task-id>-* hoặc default.
0.3  Đọc TASK.md: status, step, approvals, releases → kiểm T1–T7 cho bước sắp làm.
0.4  Đọc logs/runs của task này → biết run trước dừng ở đâu.
```

### 6.6 Sáu routine — mỗi cái một dòng

| Routine | Vào | Ra | Hold |
|---|---|---|---|
| `intake` | mô tả việc | `TASK.md` (goal · context · acceptance · constraints · learning · challenge?) | — |
| `analyze` | `TODO` | `analysis.md`: aggregate nào, invariant ở đâu, chạm context nào. **Challenge:** 3–5 câu → `WAITING_HUMAN/question` | question |
| `plan` | có `analysis.md` | `plan.md`: bước · file sẽ đụng · test sẽ thêm · **Không làm** · **Tự phản biện** | plan (nhả được) |
| `implement` | `approvals.plan` hoặc release | code trên `agent/…`, đúng file plan liệt kê | — |
| `verify` | có diff | `VerifyResult`: chạy gì · exit · bao lâu · **CHƯA test gì** (từ vựng không biết). Đỏ → `gate-failed`, nhánh để yên | — |
| `review` | verify xong | `review.md` = **change brief** (What changed and why · Files · Gate · Rollback · Left for you) + ba chiều Go/DDD/kỹ thuật + `learning.md` + nợ | commit → push (nhả được) |

Người gọi từng routine bằng **một** skill Claude Code: *"chạy `<routine>` cho
`<task-id>`"* → skill mở `employee/routines/<routine>/SKILL.md` và làm theo. Đó là
phần harness-specific duy nhất.

### 6.7 `scripts/verify-agent.sh` — cửa kiểm cơ học

Học từ STANDARD §6 và từ chính `learn/scripts/`: (1) mọi routine được tham chiếu ở
`CONTRACT`/`SKILL` phải tồn tại; (2) thân `SKILL.md` **không** chứa `go test`,
`phpunit`, `gofmt` — chỉ tên capability; (3) không chuỗi giống credential trong
`agent/`; (4) mọi `TASK.md` parse được và `status` nằm trong 7 giá trị; (5)
`CONTRACT.md` không nhắc tên ngôn ngữ. Chạy trước khi bảo "xong" một task.

### 6.8 Ví dụ đầu-cuối: T-001 trên Portage

Không dùng "Order Cancellation" của prompt — Portage **đã có** (`ordering/order.go:123-124`,
`CancelReason()` dòng 214). Dùng việc **còn nợ thật** trong `CLAUDE.md`:

> **T-001 — Đưa bảng giá từ hằng trong `wire.go` ra `config/ratecard.yaml`.**
> `internal/platform/wire/wire.go:279` chốt phí 10 % sàn 500 000 ₫; `:284` thuế 8,81 %;
> `:286` cọc 50 %; `:354` bốn mức cước; `:378` tỷ giá 26 000. Chưa có `config/`.

**Run 1 — `intake`** → `TASK.md`: `status: TODO`, `challenge: true`, acceptance: *`go
test` xanh không giảm test · `smoke.sh` in đúng số vàng · thiếu file hoặc số sai → api
không khởi động, báo lỗi đọc được*; constraints: *không đổi số nghiệp vụ · domain không
import yaml (luật vàng)*.

**Run 2 — `analyze`.** Step 0.1 đọc `CLAUDE.md` → thấy luật 1 (panic vs error), luật
vàng, "Còn nợ". Challenge → **hỏi rồi dừng**: *(1) bảng giá là value object của
`pricing` hay cấu hình của `platform/wire`? (2) file sai số là lập trình viên sai
(panic) hay vận hành sai (error)? (3) parse yaml nằm đâu để domain vẫn chỉ stdlib +
uuid? (4) test nào đang dựa vào hằng trong `wire.go`?* → `WAITING_HUMAN/question`.

Người trả lời trong `## Corrections` của `TASK.md`. **Run 3 — `analyze`** viết
`analysis.md`: *VO giữ trong `pricing`; parse ở `internal/adapter/config`; sai số lúc
khởi động là vận hành → error + thoát; `wire.Memory` cho test giữ hằng cũ.*

**Run 4 — `plan`** → `plan.md` kết bằng **Không làm** (*không đụng `MustParse*` trong
test · không đổi số · không hot-reload*) và **Tự phản biện** (*"cách hiển nhiên hơn là
để hằng và thêm cờ — vì sao không: vì `CLAUDE.md` đã ghi nợ này từ đợt 9"*) →
`WAITING_HUMAN/plan-approval`.

Người duyệt: `approvals: [{gate: plan, by: sy, at: …, note: "giữ MustParse ở wire.go làm fallback"}]`.
Note **là một phần của plan** từ giờ (T7).

**Run 5 — `implement`** trên `agent/T-001-ratecard-yaml`. **Run 6 — `verify`:**

```
go test ./...   exit 0   1.7s   22 package
gofmt -l .      exit 0
go vet ./...    exit 0
CHƯA TEST: scripts/smoke.sh — cần Postgres, cổng 5433 bị chiếm (SETUP §9 đợt 20)
```

**Run 7 — `review`** → `review.md` mở bằng *"Mọi dòng dưới đây mô tả một thay đổi
trên nhánh. Chưa merge. Merge là của anh."*, rồi Files · Gate · Rollback (*revert một
commit*) · Left for you (*chạy smoke khi cổng 5433 rảnh*). `learning.md`: Go — *zero
value của RateCard không an toàn → constructor chặn*; DDD — *cấu hình ≠ value object*;
Agent — *challenge lần đầu: 4 câu, 1 câu đổi hướng (panic → error)*. Nợ mới → `debt.md`.
→ `WAITING_HUMAN/commit-approval`.

Duyệt → commit. `releases` rỗng → `WAITING_HUMAN/push-approval` → người tự push hoặc
duyệt → `DONE`. Cuối **mỗi** run: kiểm R1–R4, ghi **một** file vào `logs/runs/`.

**Sinh ra:** 5 file trong `tasks/T-001-…/`, 7 file trong `logs/runs/`, +1 dòng
`debt.md`, một nhánh `agent/T-001-ratecard-yaml` với một commit. `main` không đổi.

### 6.9 Ra khỏi Phase 1 khi

5 task thật trên Portage tới `DONE` · ≥ 1 task `challenge: true` · ≥ 1 task bị
**từ chối ở plan** (chứng minh cổng hoạt động) · ≥ 1 lần `gate-failed` được ghi
đúng (nhánh để yên) · `verify-agent.sh` xanh · ≤ 1 vi phạm R1–R4 trong run log.

### 6.10 Phase 1 **không** build

Runtime Go · SQLite/Postgres · scheduler · nhiều employee · Chief of Staff · vector DB ·
web UI · recipes · `SCHEDULE.md` · `state/` `queue/` `reports/` · tự tìm việc · tự nối
routine · tự thử lại · agent adapter · tự sửa `SKILL.md` · push notification.

---

## 7. Lộ trình

| Phase | Mục tiêu | Ra khỏi khi | **Cố ý không làm** |
|---|---|---|---|
| **0** | hiểu ai-employees, audit Portage, viết bản này | duyệt | code |
| **1** | employee Markdown chạy trọn task trên Portage (§6) | §6.9 | runtime, DB, scheduler |
| **2** | hệ học có giá: 1 nợ được gạch bằng bài thật; `daily` do người gọi | 1 file trong `knowledge/` sinh từ nợ | tự động |
| **3** | **dùng lại**: `projects/<symfony>/PROJECT.md`, một task PHP xong **không sửa `routines/`** | task PHP `DONE` | Go runtime |
| **4** | runtime Go ép 11 invariant thành test đỏ (§5.7) | mọi invariant có test đỏ khi vi phạm | HTTP, DB |
| **5** | gọt DDD sau ≥ 20 task: ADR "đã học gì" | ≥ 1 khái niệm bị **bỏ** | thêm pattern |
| **6** | scheduler: `verify` theo lịch trên nhánh có sẵn; `SCHEDULE.md`; khoá file; retry là invariant | 1 tuần không hỏng state | tự implement |
| **7** | nới cổng bằng bằng chứng (rung/class từ `web-guardrail-review`): **nới một bậc sau N task merge sạch, siết ngay khi một lần bị sửa**; self-edit `SKILL.md` có changelog undo | quyết định nới/không bằng số | tự nới guardrail |
| **8** | Chief of Staff: script đọc-only ra `reports/daily.md` | `grep WAITING` không đủ nữa | agent thứ hai |

---

## 8. Rủi ro và cách chặn

| Rủi ro | Chặn |
|---|---|
| Over-engineering | không interface khi một cài đặt; §5.7 kiểm lúc Phase 4 |
| DDD cargo cult | bảng §5.1 bắt buộc cột "bằng chứng nó là loại đó" cho mọi khái niệm mới; Phase 5 phải **bỏ** ≥ 1 |
| Agent chép luật project | `PROJECT.rules` **trỏ**; `verify-agent.sh` (5) cấm tên ngôn ngữ trong `CONTRACT` |
| Agent vượt plan | T7 + `review` so diff với `plan.md` |
| Xoá việc dở của người | ROLE §1.1 nguyên văn: tree bẩn → đổi gì cũng không |
| Hỏng trạng thái do run song song | Phase 1 người gọi từng run; Phase 6 khoá |
| Leo thang quyền qua note/recipe | `CONTRACT` bất khả xâm phạm; `releases` chỉ người viết; nới **chậm**, siết **ngay** |
| Hệ học thành gánh nặng | ba luật §5.6 |
| Quá nhiều routine | routine thứ 7 phải thay một cái cũ hoặc có ≥ 3 run log chứng minh |
| Dính Claude Code | `SKILL.md` chỉ văn xuôi + capability; phần harness ở **một** skill gọi |
| Portage mãi là project duy nhất | Phase 3 là điều kiện tin DEC-003 — không qua là ý tưởng dùng lại chưa được kiểm |

---

## 9. Danh sách quyết định

| # | Quyết định | Lý do | Đánh đổi |
|---|---|---|---|
| DEC-001 | Phase 1 không runtime Go; employee là file Markdown harness đọc | biết routine thật trước khi mã hoá; ai-employees chạy 11 harness bằng đúng cách này | invariant giữ bằng kỷ luật tới Phase 4 |
| DEC-002 | Employee là cấu hình, không entity | ai-employees: thư mục file | xem lại khi có employee thứ hai |
| DEC-003 | Tách Agent (chung) / Project (riêng) bằng `PROJECT.md`; luật project **ở lại** project và **thắng** (ROLE §0) | yêu cầu dùng lại; CLAUDE.md của Portage đã đúng chỗ | thêm một file mỗi project; Phase 3 là bài kiểm |
| DEC-004 | Task là aggregate root duy nhất; Approval là VO trong Task; Release là dòng trong `PROJECT.md` chỉ người viết | mọi invariant nói về Task; LAW 2 | list `approvals` dài dần |
| DEC-005 | Run là record chỉ ghi thêm | sự thật quá khứ | lịch sử = grep |
| DEC-006 | Filesystem MD+frontmatter; SQLite chỉ mục nếu Phase 6 cần | SETUP §9 chứng minh 20 đợt | không ghi song song |
| DEC-007 | Không Capability→Adapter Phase 1–3; `PROJECT.commands.*` là ánh xạ duy nhất; kỷ luật đặt tên capability trong `SKILL.md` giữ | một harness; thứ đổi là project, không phải harness | Phase 4 thêm interface `Agent` một method |
| DEC-008 | Routine vận hành, sáu bước, người gọi từng bước Phase 1; Phase 6/7 gộp một run dừng ở hold | học cần điểm dừng; `fix-runner` chứng minh gộp được sau | Phase 1 chậm có chủ ý |
| DEC-009 | Ba hold: plan · commit · push. Plan và push **nhả được theo project**. Merge không là kênh | commit/push là outbound action (LAW 2 dịch sang Git); Implementation ≡ Plan | người bấm ba lần cho tới khi nhả |
| DEC-010 | Challenge/Learning là cờ trên Task | P9 §4 "chốt rồi làm" thành dữ liệu | — |
| DEC-011 | Rút bài học nằm trong `review`; `knowledge/` chỉ nhận bài đã gạch nợ; nợ 30 ngày gợi xoá | không có ở đâu để chép; phải chặn gánh nặng | ít bài — là ý đồ |
| DEC-012 | Chief of Staff Phase 8, script đọc-only | ai-employees: giá trị khi có đội | — |
| DEC-013 | Run log Markdown mỗi run, từ vựng trạng thái **đóng**, danh sách "không bao giờ xuất hiện" | SETUP §9 + CONTRACT §4 | máy đọc phải parse frontmatter |
| DEC-014 | Nhánh `agent/<id>-<slug>`; không đụng tree bẩn; gate hỏng để yên; **không bao giờ merge** | ROLE §1.1, fix-runner §6b | người mở PR |
| DEC-015 | Độc lập harness bằng file thường + một skill gọi; không adapter | HARNESSES.md: 11 harness, 0 adapter | phần harness gom một chỗ |
| DEC-016 | `status` (7 giá trị đóng) tách `step`; `verify` là trường riêng | 8 trạng thái của prompt trộn hai câu hỏi; gate-failed không phải task failed | hai trường thay một |
| DEC-017 | `CONTRACT.md` không nhắc tên ngôn ngữ; `verify-agent.sh` kiểm | dùng lại thật, không phải hứa | — |
| DEC-018 | Adopt nguyên văn: `PAUSED` · `## Corrections` · LAW 5 · LAW 6 · R1–R4 · từ vựng không biết · scan secret · change brief | rẻ, đã được trả giá ở nơi khác | — |
| DEC-019 | Nới quyền chỉ ở Phase 7, bằng bằng chứng, **bất đối xứng** (nới chậm, siết ngay); *tự sửa không bao giờ nới cái được phép* là luật từ ngày một | `web-guardrail-review` | Phase 1–6 người bấm nhiều |
| DEC-020 | Phase 1 ở `agent/` trong Portage, không `go.mod`, mọi đường dẫn tương đối, `git mv` ra được không sửa | anh: không tạo repo riêng phase này, nhưng tách sau | một thư mục lạ trong repo Go |

---

## 10. Ánh xạ về 27 mục Part 3 của prompt

| Part 3 | Ở đây | | Part 3 | Ở đây |
|---|---|---|---|---|
| 1 Executive Summary | §0 + §1 kết luận | | 14 Learning System | §5.6 |
| 2 Goals / Non-goals | §6.9, §6.10, §7 cột cuối | | 15 Knowledge & State | §6.2 |
| 3 Concepts | §5.1 | | 16 Git Workflow | §6.3 mục 3, §6.8 |
| 4 High-Level Architecture | §5.2 | | 17 Guardrails / ma trận | §6.3, R1–R4 |
| 5 Employee Design | §6.2–6.5 | | 18 Run Logging | §5.4, DEC-013 |
| 6 Routine Design | §6.6, DEC-008 | | 19 ADR Strategy | §1 hàng "Quyết định 08/09" + Phase 5 |
| 7 Task Lifecycle | §5.4 | | 20 Self-Improvement | Phase 7, DEC-019 |
| 8 Domain Model | §5.1, §5.3 | | 21 Chief of Staff | DEC-012, Phase 8 |
| 9 Aggregate & Invariant | §5.3 | | 22 End-to-End | §6.8 |
| 10 Go Architecture | §5.7 | | 23 Phase Roadmap | §7 |
| 11 Persistence | §5.5 | | 24 Phase 1 Definition | §6 |
| 12 Agent Integration | DEC-007, DEC-015 | | 25 Risks | §8 |
| 13 Human Approval | §2.2, DEC-009 | | 26 Challenges | §4 |
| | | | 27 Decision Summary | §9 |

---

**Ba câu tôi tự chốt** (anh bảo chọn best practice, không chờ anh):

1. **Huỷ task giữa chừng:** người ghi `status: REJECTED` kèm note; employee ở run
   kế tiếp xoá nhánh `agent/…` **chỉ khi** nhánh không có commit nào, còn có commit
   thì để lại và ghi vào `Left for you`. Không có routine `abandon` riêng — một
   dòng người viết là đủ, và giữ nguyên tinh thần T4.
2. **Tên routine:** `verify`, không `test` — vì nó chạy test + fmt + lint **và nói rõ
   cái chưa chạy**; "test" hứa ít hơn thứ nó làm.
3. **Cổng push:** giữ, nhưng **nhả được theo project** (`releases`). Mặc định giữ vì
   ai-employees đúng: nhả là việc của người, bằng bằng chứng, không phải mặc định
   của kiến trúc.
