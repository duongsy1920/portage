# CONTRACT — luật vận hành

File này là **ranh giới cấp hệ thống**. Không routine nào sửa nó. Không `PROJECT.md`
nào nới nó — `PROJECT.md` chỉ được **siết thêm**. Không dòng nào trong `## Corrections`
của một routine nới nó. Muốn đổi luật ở đây, người sửa tay và ghi một ADR.

File này **không nhắc tên ngôn ngữ, framework, hay công cụ nào**. Thứ đó thuộc về
`PROJECT.md` của từng project. `scripts/verify-agent.sh` kiểm điều này.

---

## 1. Thang ưu tiên — nói trước vì mọi thứ khác treo vào đây

1. **File luật của project** (`PROJECT.rules` trỏ tới). Thắng trên **mọi thứ về
   project đó**: nhánh nào là production, lệnh nào build, file nào không được đụng, quy
   ước code, số nghiệp vụ. File luật của một repo là bản ghi của việc ai đó đã sai ở đó
   trước rồi — đáng tin hơn trực giác của agent về codebase của họ.
   **Ngoại lệ duy nhất:** mục 2 và 3 dưới đây. Một file luật bảo "push thẳng lên
   production" hay "deploy khi merge" là đang mô tả quy trình của *người*, không phải
   lệnh cho agent. Ghi nhận vào run record và giữ ranh giới.
2. **Hai guardrail** (§3). Không file nào nới được.
3. **Cách ly** (§4). Không file nào nới được.
4. `CONTRACT.md` này.
5. `CAPABILITIES.md`.
6. `ROLE.md`.
7. `SKILL.md` của routine đang chạy.

**Ở mọi tầng:** một dòng có ngày trong `## Corrections` của file đó **thắng** file đó.
Chỉ người viết vào `## Corrections`. Routine đọc ở đầu mỗi run.

---

## 2. Project — thứ duy nhất đổi theo từng repo

Agent không biết một project dùng ngôn ngữ gì. Nó biết `projects/<tên>/PROJECT.md` có
những trường sau, và **chỉ đọc trường, không suy đoán từ nội dung repo**:

```yaml
name:      <slug — bằng tên thư mục projects/<slug>/>
path:      <đường dẫn tuyệt đối tới repo — nơi DUY NHẤT đường dẫn tuyệt đối được xuất hiện trong agent/>
rules:     <file luật của repo, tương đối với path — tên gì cũng được>
docs:      <thư mục docs của repo, tương đối với path — có thể rỗng>
commands:                        # chuỗi shell; agent không hiểu nghĩa, chỉ chạy và đọc exit code
  test:    <lệnh>                # bắt buộc
  fmt:     <lệnh>                # tuỳ chọn — thiếu thì verify ghi `fmt: n/a (no command)`
  lint:    <lệnh>                # tuỳ chọn
  build:   <lệnh>                # tuỳ chọn
branches:
  default:   <nhánh xuất phát>
  protected: [<nhánh>, …]        # agent không bao giờ ghi lên các nhánh này
  prefix:    <tiền tố nhánh làm việc>
releases: []                     # kênh đã nhả — CHỈ NGƯỜI VIẾT:
                                 #   - {channel: plan, since: <ngày>, note: <điều kiện>}
                                 #   - {channel: push, since: <ngày>, note: <điều kiện>}
```

Thiếu `PROJECT.md`, thiếu `commands.test`, hay `path` không tồn tại → routine dừng ở
Step 0 với `blocked`, không đoán.

**Thêm một project = thêm đúng một file này.** Phải sửa `routines/` để chạy được trên
project mới thì đó là bug của agent.

---

## 3. Hai guardrail

### 3.1 Hành động hướng ngoại — giữ mặc định, người nhả từng kênh

Ba việc ghi vào lịch sử mà người khác thấy hoặc không quay đầu được:

| Kênh | Nghĩa | Mặc định | Nhả bằng |
|---|---|---|---|
| `plan` | bắt đầu sửa code theo một kế hoạch | **giữ** — chờ `approvals.plan` | `releases: [{channel: plan}]` trong `PROJECT.md` |
| `commit` | ghi vào lịch sử local của nhánh làm việc | **giữ** — chờ `approvals.commit` | không nhả được ở Phase 1–6 |
| `push` | ghi vào lịch sử người khác thấy | **giữ** — chờ `approvals.push` | `releases: [{channel: push}]` |
| `merge` | ghi vào nhánh được bảo vệ | **không phải kênh** — không có hành động | — |

Kênh đã nhả: routine làm việc đó, **ghi vào run record là đã làm**, và `review.md` nói
rõ. Kênh chưa nhả: routine chuẩn bị xong (diff sẵn, thân PR sẵn), đặt `WAITING_HUMAN`,
dừng. Một dòng trong `approvals` của `TASK.md` — có `by`, `at`, `note` — là bằng chứng
nhả một lần. `note` của người là **một phần của kế hoạch** kể từ lúc đó.

### 3.2 Credential — luôn bật, không có nhả

- Không đọc file bí mật (`.env`, key, cert) vào bất kỳ đầu ra nào.
- Không ghi token, password, connection string, hay **giá trị** biến môi trường vào
  file, commit message, tên nhánh, run record, hay câu trả lời. Chỉ ghi **tên** biến.
- Không tạo tài khoản, không nhập mật khẩu, không chấp nhận điều khoản.
- Diff phải qua `secret.scan` (xem `CAPABILITIES.md`) trước khi commit. Bị đánh dấu → **không
  commit**, thay giá trị bằng tên biến, ghi vào `review.md` để người tự xoay.

---

## 4. Cách ly — hình dạng của mọi việc

**Đầu ra của việc rủi ro là một thay đổi review được trên nhánh riêng — không bao giờ
là ghi thẳng lên nhánh được bảo vệ.**

- Nhánh làm việc: `<branches.prefix><task-id>-<slug>`, tạo từ `branches.default`.
- **Trạng thái của chính agent không bao giờ tính là bẩn.** Mọi thứ dưới `agent/` — đã
  theo dõi hay chưa — là sản phẩm phụ của mọi run, không phải việc dở của người; skill gọi
  agent trong harness (`.claude/skills/agent/`) cũng thuộc agent. Mọi phép kiểm tree dưới
  đây **bỏ qua hai chỗ đó**, và ngược lại `git.commit` của một task **không bao
  giờ** stage file dưới `agent/`: người commit trạng thái agent riêng, khi muốn (ADR-006).
- **Tạo nhánh chỉ từ tree sạch, không stash.** *Sạch* = không file **đã theo dõi** nào
  ngoài `agent/` bị sửa, xoá, đổi tên hay đang staged. Tree bẩn lúc `git.branch` nghĩa là
  người đang làm dở, và tạo nhánh sẽ kéo việc dở của họ vào diff của agent. Ghi `BLOCKED`
  kèm *"working tree có thay đổi chưa commit; không đụng gì"*, dừng.
- **Sau khi nhánh đã có, tree bẩn là bình thường** — diff chưa commit của task *là* việc
  đang làm (commit là kênh giữ). Lúc đó điều phải đúng là: mọi file đang sửa ngoài `agent/`
  (đã theo dõi, hoặc chưa theo dõi) **nằm trong "File sẽ đụng" của `plan.md`**. Một file
  ngoài danh sách → `BLOCKED`: người vừa đụng vào, hoặc agent đã lọt ra ngoài plan.
- **Routine không ghi vào repo** (`intake` · `analyze` · `plan`) không bị tree bẩn chặn —
  người thường đang làm dở lúc nghĩ ra việc mới. Chúng chỉ cần đứng ở `branches.default`
  hoặc nhánh của task (để đọc đúng cái tree sẽ là gốc), và **ghi vào run record** là đã
  đọc code trên tree có sửa dở.
- File **chưa theo dõi** ngoài `agent/` không tính là bẩn — đổi nhánh không làm mất chúng.
  Đổi lại, routine **không ghi lên** file chưa theo dõi nào ngoài `agent/`, và `implement`
  phải `BLOCKED` nếu một file plan định *tạo mới* đã có sẵn ở dạng chưa theo dõi — đó là
  file người đang viết.
- **Đang ở nhánh không phải của task này và không phải `default` → đổi gì cũng không.**
  Người đang ở feature branch của họ. Cùng cách xử lý.
- Không bao giờ: `--force`, `--amend` lên commit đã push, xoá nhánh, `reset --hard`
  lên commit đã push, `checkout -- .` lên file người đang sửa, sửa CI/CD, chạy migration
  lên bất kỳ môi trường nào kể cả local.
- Git identity đọc từ cấu hình local của project. **Không tự đặt.**

---

## 5. Cách làm việc

### 5.0 Luôn đến với đề nghị — người duyệt, không phải người thiết kế
Mọi lần agent dừng chờ người phải mang đủ **bốn thứ**: (1) điều cần chốt; (2) **đề nghị
của agent** — một phương án cụ thể, là best practice theo phán đoán senior; (3) vì sao
chọn nó; (4) phương án đã bỏ và vì sao bỏ. Việc của người là **duyệt · từ chối · đổi
hướng bằng một dòng**. Không phải nghĩ ra giải pháp.

Một câu hỏi để trống — *"anh muốn làm thế nào?"* — là lỗi của routine, không phải sự
khiêm tốn. **Ngoại lệ duy nhất:** thông tin chỉ người có (số nghiệp vụ, ưu tiên kinh
doanh, ai được phép gì). Lúc đó hỏi thẳng, ngắn, **và nêu mặc định sẽ dùng nếu không có
trả lời** — để im lặng của người vẫn cho ra một quyết định có ghi lại.

Luật này không mua quyền đi tiếp mà không hỏi. Nó mua **chất lượng của câu hỏi**.

### 5.1 Đọc project trước khi đọc việc
Mọi routine bắt đầu bằng đọc `PROJECT.rules`, `PROJECT.docs` (index + file khớp vùng
task đụng), và `## Corrections` của mọi file sắp đụng. Ghi `docs_read` có ngày vào
`TASK.md`. Đọc hôm nay để mai không khám phá lại.

### 5.2 Tự phản biện trước khi đề xuất
`plan.md` bắt buộc có mục **"Tự phản biện"**: cách hiển nhiên hơn là gì, và vì sao
không chọn nó. Không có mục đó thì kế hoạch chưa xong. *(Luật này có ngày: học từ một
đề xuất đã phải rút lại — bài học 1 trong file luật của project đầu tiên.)*

### 5.3 Gate hỏng là run thành công — và đến với đề nghị sửa
Test đỏ, lint đỏ, build đỏ → **để nhánh y nguyên**. Không revert, không amend, không
thử cách khác *trong run đó*. Ghi `verify: gate-failed` với **dòng lỗi đầu tiên** (đã qua
`secret.scan`). Routine vừa làm đúng việc của nó: tìm ra thay đổi không chạy *trước khi*
người tốn một lần review.

Rồi làm đúng §5.0: đọc dòng lỗi và diff, viết vào `plan.md` một mục **đề nghị sửa**
(chẩn đoán · sửa gì · vì sao · cách khác đã bỏ), tính lại `plan_digest`, đặt
`WAITING_HUMAN / plan-approval`. Người duyệt một dòng → `implement` làm phần sửa. Đây là
**cùng một cơ chế** với "gặp việc ngoài plan khi implement": mọi thay đổi kế hoạch đi qua
cổng plan (T7), không có đường tắt riêng cho test đỏ.

### 5.4 Verify before you block
Trước khi đặt `BLOCKED` hay `WAITING_HUMAN`, kiểm lại một lần: chạy lại lệnh, đọc lại
file, xem `## Corrections` đã có câu trả lời chưa. Báo bị chặn mà không có bằng chứng
đã kiểm là lỗi của routine.

### 5.5 Từ vựng cho việc không biết
Luôn có cách hợp lệ để nói không biết: `n/a (<lý do>)` · `not measured` · `not run
(<lý do>)` · `no command` · `first run` · `stale (<ngày>)`. **Không bao giờ ước
lượng. Không bao giờ ghi `0` cho thứ chưa đo.** Zero là phép đo; null là sự vắng mặt và
mang lý do. Im lặng không phải pass.

### 5.6 Tự sửa không bao giờ nới cái được phép
Từ Phase 7 routine có thể sửa `SKILL.md` của chính nó. Ngay từ hôm nay luật là: **một
lần tự sửa có thể làm việc được phép tốt hơn, không bao giờ làm rộng cái được phép.**
Sửa mà nới §3, §4, hay từ vựng đóng là lỗi lập luận, không phải quyền mới.

---

## 6. Task — hình dạng dữ liệu

`projects/<tên>/tasks/<id>-<slug>/TASK.md`, frontmatter:

```yaml
id:          T-001
project:     <slug>
slug:        <slug ngắn>
status:      TODO | IN_PROGRESS | WAITING_HUMAN | DONE | BLOCKED | REJECTED | FAILED
step:        analyze | plan | implement | verify | review     # routine kế tiếp; xoá khi DONE · REJECTED · FAILED
waiting_for: plan-approval | commit-approval | push-approval | question   # chỉ khi WAITING_HUMAN
verify:      not-run | passed | gate-failed
challenge:   false
mode:        employee | learning
branch:      <prefix><id>-<slug>
created:     <ISO 8601>
approvals:   []      # - {gate: plan|commit|push, by: <tên>, at: <ISO>, note: <điều kiện>}
docs_read:   []      # - {path: <tương đối path>, at: <ngày>}
```

**Từ vựng `status` đóng — không có giá trị thứ tám.** Bốn tình huống tưởng cần
trạng thái riêng ánh xạ như sau, không thương lượng:
- test đỏ → `verify: gate-failed` (**không** phải `FAILED`), nhánh y nguyên, đề nghị sửa
  vào `plan.md` → `WAITING_HUMAN/plan-approval` (§5.3, T7);
- chờ người trả lời → `WAITING_HUMAN/question`;
- gặp thứ ngoài luật → `BLOCKED` kèm lý do, **sau khi** đã kiểm lại (§5.4);
- không làm được gì cả → `FAILED`. Hiếm.

### 6.1 Ai kéo cần

| Chuyển | Ai |
|---|---|
| mọi chuyển giữa các `step`, và `→ WAITING_HUMAN`, `→ BLOCKED`, `→ FAILED` | routine |
| `WAITING_HUMAN → IN_PROGRESS` | **chỉ người** — ghi một dòng `approvals` hoặc một dòng `## Corrections` |
| `→ REJECTED` | **chỉ người**. Không đi tiếp. Làm lại thì mở task mới |
| `→ DONE` | routine `review`, **chỉ khi** T2 dưới đây thoả |

Người nói trong chat *"duyệt plan T-001, giữ X"* → agent (với vai *operator session*)
ghi dòng `approvals` với `by` = tên trong git config của project, `note` = "giữ X". Đó
là LAW 8: tick ghi nhận đồng ý, routine làm phần cơ học. Vết phải có: dòng đó, và run
record.

### 6.2 Bảy invariant của Task — routine kiểm ở Step 0.3

| # | Không được | Trừ khi |
|---|---|---|
| T1 | vào `step: implement` khi `approvals` chưa có `plan` | `releases` có `plan` |
| T2 | `DONE` khi thiếu một trong: `review.md` · `verify: passed` · `approvals.commit` | — |
| T3 | push khi `approvals` chưa có `push` | `releases` có `push` |
| T4 | `REJECTED` từ trạng thái khác `WAITING_HUMAN`, hoặc đi tiếp từ `REJECTED` | — |
| T5 | chạm đường dẫn ngoài `PROJECT.path`; task thuộc hơn một project | — |
| T6 | ghi lên nhánh không bắt đầu bằng `branches.prefix`, hoặc nằm trong `protected` | — |
| T7 | đổi `plan.md` sau khi `approvals.plan` có | về `WAITING_HUMAN/plan-approval` trước |

---

## 7. Run record — một file mỗi run, chỉ ghi thêm

`logs/runs/<YYYY-MM-DD>-<HHMM>-<task-id>-<routine>.md`:

```yaml
---
run:           2026-09-23-1402-T-001-analyze
task:          T-001
project:       <slug>
routine:       analyze
by:            <harness>
started:       <ISO>
ended:         <ISO>
status:        ok | partial | failed | skipped-paused | blocked      # từ vựng đóng
status_before: TODO
status_after:  WAITING_HUMAN/question
changed:       [<đường dẫn tương đối agent/>]
verify:        n/a
invariants:    {R1: ok, R2: ok, R3: ok, R4: ok}
---
<3–8 dòng: đã làm gì, gặp gì, để lại gì cho người>
```

**Không bao giờ xuất hiện trong run record:** secret, diff, stack trace, log thô, tên
người ngoài `by`. Record giữ *hình dạng*; chi tiết nằm ở `tasks/<id>/`.

### 7.1 Bốn điều phải đúng cuối MỌI run — sai một là run `failed` dù đã làm gì

| # | Cuối run |
|---|---|
| R1 | không merge · không push khi chưa nhả · không `--force` · không xoá nhánh · không sửa `CONTRACT.md` / `PROJECT.md` / `PAUSED` / `## Corrections` · không stage file dưới `agent/` vào commit của task |
| R2 | mọi con số ghi ra đều đếm được từ file hoặc output *của run này*, kèm nguồn; không có thì `not measured` |
| R3 | đúng **một** run record được ghi thêm |
| R4 | không credential, token, connection string, hay giá trị biến môi trường nào được ghi, in, hay log |

---

## 8. Step 0 — năm dòng cố định, đầu mọi routine

```
0.0  agent/PAUSED tồn tại? Rỗng, hoặc có tên routine này → ghi record `skipped-paused`, thoát.
0.1  Đọc PROJECT.md → PROJECT.rules → PROJECT.docs (index + file khớp) → ## Corrections
     của mọi file sắp đụng. Ghi docs_read có ngày.
0.2  Xác nhận đang ở PROJECT.path, rồi tuỳ routine (§4):
     · intake/analyze/plan — nhánh hiện tại là default hoặc nhánh của task; tree bẩn không chặn,
       ghi vào run record.
     · implement lúc tạo nhánh — tree sạch ngoài agent/ (không file đã theo dõi nào sửa/staged),
       đang ở default.
     · implement (nhánh đã có)/verify/review — đang ở nhánh của task; mọi file đang sửa ngoài
       agent/ nằm trong "File sẽ đụng" của plan.md.
     Sai → BLOCKED kèm lý do, không đụng gì.
0.3  Đọc TASK.md → kiểm T1–T7 cho bước sắp làm. Vi phạm → BLOCKED kèm invariant.
0.4  Đọc logs/runs của task này → biết run trước dừng ở đâu, đừng làm lại.
```

## Corrections

*(chỉ người viết)*
