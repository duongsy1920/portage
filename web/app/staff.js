// staff.js — the screen a person doing the work actually uses.
//
// What it is NOT: one button per endpoint. That was the first console, and it
// failed for a reason worth writing down — it showed the shape of the API
// instead of the shape of the job. Nobody thinks "add variant, confirm
// listing, measure, publish". They think "a customer sent me a link, make it
// sellable".
//
// So the screen is a desk: three queues on the left, each with its count, and
// on the right the one piece of work that is open. Inside it exactly one step
// is actionable at a time, and a finished step shows what it recorded, not a
// greyed-out button: "US 9 · black" is proof the step happened. When a piece
// of work leaves its queue the desk opens the next one by itself, so nobody
// has to scroll to find what to do next.

import { React, html, Top, Sheet, Problem, Field, Steps, Journey, Keys, Icon, Toasts, PageHead,
  Skeleton, Empty, useToasts, usePoll, useAction } from "./ui.js";
import { call, money, retryOn, getTokens, setTokens } from "./portage.js";
import { TASK, PARCEL, BATCH, say, friendly, categoryWords, grams, dayMonth } from "./words.js";

const { useState, useEffect, useCallback, useMemo, useRef } = React;

const POLL_MS = 4000;

const STEP_TITLE = {
  variant: "Khách mua size nào",
  listing: "Xác nhận đúng sản phẩm",
  measure: "Cân và đo thật",
  publish: "Đăng bán",
};
const STEP_ORDER = ["variant", "listing", "measure", "publish"];

/* ── one product, one checklist ───────────────────────────────────────────── */
function ProductSheet({ item, shops, categories, onDone }) {
  const shop = shops[item.merchant_id];
  const box = categories[item.category] || {};
  const act = useAction();

  // Prefilled with the customer's own words. They are the only person who
  // knows which size they want, and typing it again is both work and a chance
  // to get it wrong.
  const [size, setSize] = useState(item.requested_variant || "");
  const [colour, setColour] = useState("");
  const [ref, setRef] = useState("");
  // The weight is PREFILLED from the category's default box and labelled as a
  // guess, because that is exactly what it is: the shop's page does not
  // publish a packed weight, so somebody has to put the thing on a scale.
  const [g, setG] = useState("");
  const [l, setL] = useState("");
  const [w, setW] = useState("");
  const [h, setH] = useState("");
  useEffect(() => {
    if (box.weight_g && !g) {
      setG(String(box.weight_g));
      setL(String(box.length_mm));
      setW(String(box.width_mm));
      setH(String(box.height_mm));
    }
  }, [box.weight_g]);

  const post = (path, body) => call("POST", `/products/${item.product_id}${path}`, { as: "operator", body });

  // What THIS sheet just got the server to accept. The work list is a read
  // model, filled a relay tick later, so without this the step stayed "now"
  // for a second or more after a 201 and the cleared field looked like lost
  // data. The server's own answer replaces it as soon as the list catches up.
  const [local, setLocal] = useState({});
  const accepted = (step, recorded) => setLocal(x => ({ ...x, [step]: recorded || true }));

  // onDone is NOT awaited: once the server has accepted a step the next one
  // is open at once, and the list refreshes behind it.
  const addSize = () => act.go(async () => {
    await post("/variants", { size: size.trim(), color: colour.trim(), merchant_ref: ref.trim() });
    accepted("variant", [size.trim(), colour.trim()].filter(Boolean).join(" · "));
    onDone();
  });
  const confirm = () => act.go(async () => { await post("/confirm-listing"); accepted("listing"); onDone(); });
  const measure = () => act.go(async () => {
    await post("/measure", {
      weight_g: Number(g), length_mm: Number(l), width_mm: Number(w), height_mm: Number(h),
    });
    accepted("measure", `${grams(g)}, hộp ${l}×${w}×${h} mm`);
    onDone();
  });
  // publish can race the relay for nothing else, but a fresh variant may not
  // be visible to this handler yet on a cold start, so one retryable code.
  const publish = () => act.go(async () => {
    await retryOn(["no_variants"], () => post("/publish"));
    accepted("publish");
    onDone();
  });

  const isDone = step => stepIsDone(item, step) || !!local[step];
  const nowStep = STEP_ORDER.find(step => !isDone(step));
  const stateOf = mine => (mine === nowStep ? "now" : isDone(mine) ? "done" : "todo");

  const steps = [
    {
      key: "variant", title: STEP_TITLE.variant, state: stateOf("variant"),
      recorded: html`Đã có <b>${item.variant_label || (typeof local.variant === "string" && local.variant) || "một phiên bản"}</b>`,
      why: item.requested_variant
        ? html`Ô Size đã điền sẵn <b>đúng chữ khách viết</b>. Việc của bạn là mở trang shop xem có loại đó
            không, rồi sửa lại theo cách shop ghi nếu khác. Chữ cuối cùng ở đây in nguyên văn lên phiếu đi mua.
            Coi kỹ <b>chữ cái sau số</b>: cùng số 1 mà <b>1Y</b> (thiếu niên) và <b>1C</b> (trẻ nhỏ) là hai
            đôi khác nhau, mua nhầm là đổi cả đơn.`
        : html`Khách mua <b>một hình thức cụ thể</b>, không mua "một đôi giày". Chữ bạn gõ ở đây là chữ của
            shop, và nó sẽ hiện nguyên văn trên phiếu đi mua. Giữ nguyên cả chữ cái của hệ size: <b>M</b> nam, <b>W</b> nữ, <b>Y</b> thiếu niên, <b>C</b> trẻ nhỏ, còn áo thì <b>S/M/L</b>.`,
      children: html`
        <div class="grid3">
          <${Field} ...${{ label: "Size", value: size, placeholder: "US 9, M 8 / W 9.5, 1Y, L",
            onChange: e => setSize(e.target.value),
            hint: item.requested_variant ? `khách viết: ${item.requested_variant}` : "" }} />
          <${Field} ...${{ label: "Màu", value: colour, placeholder: "black",
            onChange: e => setColour(e.target.value) }} />
          <${Field} ...${{ label: "Mã của shop", value: ref, placeholder: "không có thì để trống",
            onChange: e => setRef(e.target.value) }} />
        </div>
        <div class="actions">
          <button class="primary" aria-busy=${act.busy} disabled=${act.busy || (!size.trim() && !colour.trim())}
            onClick=${addSize}>Lưu size này</button>
          <span class="dim">Trùng size và màu sẽ bị từ chối, không phân biệt hoa thường hay khoảng trắng.</span>
        </div>`,
      after: "Cần ít nhất một size trước khi đăng bán được.",
    },
    {
      key: "listing", title: STEP_TITLE.listing, state: stateOf("listing"),
      recorded: "Đã có người xác nhận",
      why: html`Đây là <b>chữ ký của bạn</b>: tôi đã mở trang đó và đúng là sản phẩm này, của shop này, ngành hàng này. Khách dán link thì chưa ai kiểm, nên thiếu bước này hệ thống sẽ không cho đăng bán.`,
      children: html`
        <div class="actions">
          <button class="primary" aria-busy=${act.busy} disabled=${act.busy} onClick=${confirm}>Tôi đã xem trang, đúng sản phẩm</button>
          ${item.source && html`<a href=${item.source} target="_blank" rel="noreferrer noopener">
            <${Icon} name="link" size=${16} />Mở trang shop</a>`}
        </div>`,
      after: "Ai xác nhận được lấy từ chìa khoá đang dùng, không phải từ một ô nhập.",
    },
    {
      key: "measure", title: STEP_TITLE.measure, state: stateOf("measure"),
      recorded: typeof local.measure === "string" ? html`Đã ghi <b>${local.measure}</b>` : "Đã ghi số cân",
      why: html`Trang shop <b>không công bố</b> cân nặng đóng gói, nên đây là số mình tự tích luỹ. Số điền sẵn là <b>hộp mẫu của nhóm ${categoryWords(item.category)}</b>, chỉ để đỡ gõ. Hãy sửa thành số cân thật của hộp trước mặt, vì cước tính theo nó.`,
      children: html`
        <div class="grid4">
          <${Field} ...${{ label: "Cân nặng (g)", value: g, inputMode: "numeric", onChange: e => setG(e.target.value) }} />
          <${Field} ...${{ label: "Dài (mm)", value: l, inputMode: "numeric", onChange: e => setL(e.target.value) }} />
          <${Field} ...${{ label: "Rộng (mm)", value: w, inputMode: "numeric", onChange: e => setW(e.target.value) }} />
          <${Field} ...${{ label: "Cao (mm)", value: h, inputMode: "numeric", onChange: e => setH(e.target.value) }} />
        </div>
        <div class="note">Cước lấy số lớn hơn giữa cân thật và cân quy đổi, rồi làm tròn lên bước 500 g.
          Với hộp ${l}×${w}×${h} mm thì cân quy đổi là ${volumetric(l, w, h)} g.</div>
        <div class="actions">
          <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !g || !l || !w || !h} onClick=${measure}>Lưu số cân</button>
        </div>`,
      after: "Thiếu một cạnh cũng không tính được cước, nên hệ thống đòi đủ bốn số.",
    },
    {
      key: "publish", title: STEP_TITLE.publish, state: stateOf("publish"),
      recorded: "Đã đăng bán",
      why: html`Từ giây này khách mới xin được báo giá. Trước đó mọi thứ chỉ là nháp, và báo giá trên dữ liệu chưa ai kiểm là báo giá sai tiền thật.`,
      children: html`
        <div class="actions">
          <button class="primary" aria-busy=${act.busy} disabled=${act.busy} onClick=${publish}>Đăng bán</button>
          <span class="dim">Sau bước này khách sẽ thấy nút xin báo giá.</span>
        </div>`,
      after: "Mở được khi ba bước trên xong.",
    },
  ];

  return html`
    <${Sheet} label=${item.name}>
      <div class="sheet-head">
        <h2>${item.name}</h2>
        <span class="pill">${categoryWords(item.category)}</span>
        ${item.sourced_by === "customer"
          ? html`<span class="pill info">khách gửi</span>`
          : html`<span class="pill">tự thêm</span>`}
      </div>
      <div class="facts">
        <span>Shop <b>${shop ? shop.name : "chưa rõ"}</b></span>
        <span>Giá trên web <b>${money(item.price)}</b></span>
        ${item.source && html`<a href=${item.source} target="_blank" rel="noreferrer noopener">${trim(item.source)}</a>`}
      </div>
      ${item.requested_variant
        ? html`<div class="note why">Khách yêu cầu: <b>${item.requested_variant}</b>. Mở trang shop kiểm
            xem có đúng loại đó không. Shop ghi khác thì sửa lại theo chữ của shop, vì phiếu đi mua sẽ
            in đúng chữ bạn nhập.</div>`
        : html`<div class="note">Khách <b>không ghi</b> muốn loại nào. Nếu sản phẩm có nhiều size thì hỏi
            lại khách trước khi nhập, đừng đoán.</div>`}
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}
      <${Steps} steps=${steps} />
    <//>`;
}

function stepIsDone(item, step) {
  if (step === "variant") return item.has_variant;
  if (step === "listing") return item.listing_confirmed;
  if (step === "measure") return item.measured;
  if (step === "publish") return item.published;
  return false;
}

function volumetric(l, w, h) {
  const v = Math.ceil((Number(l) * Number(w) * Number(h)) / 5000);
  return Number.isFinite(v) ? v : 0;
}

const trim = u => (u && u.length > 58 ? u.slice(0, 58) + "…" : u);

/* ── one thing to go and buy ──────────────────────────────────────────────── */
function TaskSheet({ task: t, order, facts, onDone, next }) {
  const act = useAction();
  const [ref, setRef] = useState("");
  const [paid, setPaid] = useState("");
  const st = say(TASK, t.status);

  return html`
    <${Sheet} label=${t.product_name || "việc đi mua"}>
      <div class="sheet-head">
        <h2>${t.product_name || "Sản phẩm chưa có tên"}</h2>
        ${t.variant_label && html`<span class="pill info">${t.variant_label}</span>`}
        <span class="spacer"></span>
        <span class="pill ${st.tone}">${st.words}</span>
      </div>
      <div class="facts">
        ${t.variant_ref && html`<span>Mã shop <b>${t.variant_ref}</b></span>`}
        <span>Trả bằng <b>${t.currency}</b></span>
        ${t.status !== "open" && t.reference && html`<span>Mã đơn ở shop <b>${t.reference}</b></span>`}
        ${t.paid && html`<span>Đã trả thật <b>${money(t.paid)}</b></span>`}
      </div>
      ${order && html`<${Journey} order=${order} viewer="staff" facts=${facts} />`}
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}
      ${t.status === "open" && html`
        <${React.Fragment}>
          ${t.source && html`
            <div class="actions">
              <a href=${t.source} target="_blank" rel="noreferrer noopener"><${Icon} name="link" size=${16} />Mở trang để mua</a>
            </div>`}
          <div class="grid2">
            <${Field} ...${{ label: "Mã đơn ở shop", value: ref, placeholder: "NK-20260905-001",
              onChange: e => setRef(e.target.value),
              hint: "Số đơn shop gửi trong mail xác nhận. Cần nó để đối chiếu khi kiện hàng về kho, và để khiếu nại nếu shop giao sai." }} />
            <${Field} ...${{ label: `Đã trả thật (${t.currency})`, value: paid, placeholder: "163.22",
              onChange: e => setPaid(e.target.value),
              hint: "Số tiền thật đã trả, không phải số đã báo khách. Hai số này lệch nhau là chuyện thường, và hệ thống cần cả hai để biết đơn lời hay lỗ." }} />
          </div>
          <div class="actions">
            <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !ref.trim() || !paid.trim()}
              onClick=${() => act.go(async () => {
                await call("POST", `/purchase-tasks/${t.id}/confirm`, { as: "operator",
                  body: { reference: ref.trim(), paid: paid.trim(), currency: t.currency } });
                await onDone(t);
              })}>Đã mua xong</button>
            <button class="danger" disabled=${act.busy}
              onClick=${() => act.go(async () => {
                await call("POST", `/purchase-tasks/${t.id}/fail`, { as: "operator",
                  body: { reason: "shop hết hàng" } });
                await onDone(t);
              })}>Không mua được</button>
          </div>
          <p class="dim">Nhập bằng đúng loại tiền shop tính (${t.currency}), và mã đơn không được để trống.</p>
        <//>`}
      ${t.status !== "open" && next}
    <//>`;
}

/* ── one order waiting for money ──────────────────────────────────────────────
 * Taking a payment is an operator action today: there is no payment gateway,
 * so a person confirms the transfer they can see in a bank app. The customer's
 * screen shows the amount and says who to expect it from.
 */
function MoneySheet({ order: o, facts, onDone }) {
  const act = useAction();
  const which = o.deposit_paid ? "balance" : "deposit";
  const amount = o.deposit_paid ? o.balance : o.deposit;

  return html`
    <${Sheet} label=${o.product_name || "đơn chờ tiền"}>
      <div class="sheet-head">
        <h2>${o.product_name || "Sản phẩm chưa có tên"}</h2>
        ${o.placed_by_id && html`<span class="pill">đơn đặt hộ (nhân viên)</span>`}
      </div>
      <${Journey} order=${o} viewer="staff" facts=${facts} />
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}
      <table class="t">
        <tbody>
          <tr class="due">
            <th>Cần thu<div class="hint">${o.deposit_paid
              ? "Phần còn lại: tổng trừ cọc, thu khi hàng đã bay."
              : "Cọc: một nửa tổng, thu trước khi đi mua."}</div></th>
            <td class="amount">${money(amount)}</td>
          </tr>
        </tbody>
      </table>
      <div class="actions">
        <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !amount}
          onClick=${() => act.go(async () => {
            await call("POST", `/orders/${o.order_id}/${which}`, { as: "operator",
              body: { amount: amount.amount, currency: amount.currency } });
            await onDone();
          })}>Đã nhận đủ tiền</button>
      </div>
      <div class="note">Chỉ bấm khi tiền đã <b>thực sự</b> vào tài khoản. Phải đúng số, thiếu hay thừa một
        đồng hệ thống đều không nhận: cọc thiếu không phải một cam kết nhỏ hơn, và cọc thừa là một khoản
        phải hoàn mà không ai xin.</div>
    <//>`;
}

/* ── a box at the Denver warehouse ────────────────────────────────────────────
 * Two jobs, one sheet: weigh it when it arrives (the real weight, the one the
 * freight is finally charged on), then put it into the batch being packed.
 */
function ParcelSheet({ parcel: p, order, facts, openBatch, onDone }) {
  const act = useAction();
  const [g, setG] = useState("");
  const [l, setL] = useState("");
  const [w, setW] = useState("");
  const [h, setH] = useState("");
  const st = say(PARCEL, p.status);
  const name = (order && order.product_name) || "Kiện hàng";

  const receive = () => act.go(async () => {
    await call("POST", `/parcels/${p.id}/receive`, { as: "operator",
      body: { weight_g: Number(g), length_mm: Number(l), width_mm: Number(w), height_mm: Number(h) } });
    await onDone();
  });
  // No batch being packed yet: open one on the only lane there is, then add.
  const pack = () => act.go(async () => {
    let batch = openBatch;
    if (!batch) batch = (await call("POST", "/batches", { as: "operator", body: { lane: LANE } })).id;
    await call("POST", `/batches/${batch}/parcels`, { as: "operator", body: { parcel_id: p.id } });
    await onDone("b:" + batch);
  });

  return html`
    <${Sheet} label=${name}>
      <div class="sheet-head">
        <h2>${name}</h2>
        <span class="spacer"></span>
        <span class="pill ${st.tone}">${st.words}</span>
      </div>
      <div class="facts">
        ${p.reference && html`<span>Mã đơn ở shop <b>${p.reference}</b></span>`}
        ${p.actual && html`<span>Cân thật <b>${grams(p.actual.weight_g)}</b></span>`}
        ${p.actual && html`<span>Hộp <b>${p.actual.length_mm}×${p.actual.width_mm}×${p.actual.height_mm} mm</b></span>`}
      </div>
      ${order && html`<${Journey} order=${order} viewer="staff" facts=${facts} />`}
      ${st.why && html`<div class="note why">${st.why}</div>`}
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}
      ${p.status === "expected" && html`
        <${React.Fragment}>
          <div class="grid4">
            <${Field} ...${{ label: "Cân nặng (g)", value: g, inputMode: "numeric", placeholder: "1250", onChange: e => setG(e.target.value) }} />
            <${Field} ...${{ label: "Dài (mm)", value: l, inputMode: "numeric", placeholder: "340", onChange: e => setL(e.target.value) }} />
            <${Field} ...${{ label: "Rộng (mm)", value: w, inputMode: "numeric", placeholder: "230", onChange: e => setW(e.target.value) }} />
            <${Field} ...${{ label: "Cao (mm)", value: h, inputMode: "numeric", placeholder: "130", onChange: e => setH(e.target.value) }} />
          </div>
          ${l && w && h && html`<div class="note">Với hộp ${l}×${w}×${h} mm thì cân quy đổi là ${volumetric(l, w, h)} g.
            Cước lấy số lớn hơn giữa số này và cân thật.</div>`}
          <div class="actions">
            <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !g || !l || !w || !h} onClick=${receive}>Lưu số cân kiện</button>
            <span class="dim">Chỉ cân khi hộp đã thật sự nằm ở kho.</span>
          </div>
        <//>`}
      ${p.status === "received" && html`
        <div class="actions">
          <button class="primary" aria-busy=${act.busy} disabled=${act.busy} onClick=${pack}>
            ${openBatch ? "Xếp vào lô đang gom" : "Mở lô mới và xếp vào"}</button>
          <span class="dim">${openBatch ? "Lô đang gom còn nhận thêm kiện." : "Chưa có lô nào đang gom, nên mở lô mới."}</span>
        </div>`}
    <//>`;
}

/* ── a batch: close it, then put the airline's invoice on it ─────────────────
 * The batch is never measured as a whole: the airline bills the batch, and
 * that bill is split across its parcels by chargeable weight (FreightAllocator).
 */
function BatchSheet({ batch: b, orderById, parcelById, onDone }) {
  const act = useAction();
  const [freight, setFreight] = useState("");
  const st = say(BATCH, b.status);
  const nameOf = parcelId => {
    const p = parcelById[parcelId];
    const o = p && orderById[p.order_id];
    return (o && o.product_name) || "Kiện hàng";
  };

  const close = () => act.go(async () => {
    await call("POST", `/batches/${b.id}/close`, { as: "operator" });
    await onDone();
  });
  const ship = () => act.go(async () => {
    await call("POST", `/batches/${b.id}/ship`, { as: "operator", body: { freight: freight.trim(), currency: FREIGHT_CURRENCY } });
    await onDone();
  });

  return html`
    <${Sheet} label="Lô gom">
      <div class="sheet-head">
        <h2>Lô gom ${b.parcels.length} kiện</h2>
        <span class="spacer"></span>
        <span class="pill ${st.tone}">${st.words}</span>
      </div>
      <div class="facts">
        <span>Mở ngày <b>${dayMonth(b.opened_at)}</b></span>
        ${b.closed_at && html`<span>Dán kín <b>${dayMonth(b.closed_at)}</b></span>`}
        ${b.shipped_at && html`<span>Bay <b>${dayMonth(b.shipped_at)}</b></span>`}
        ${b.freight && html`<span>Cước cả lô <b>${money(b.freight)}</b></span>`}
      </div>
      ${st.why && html`<div class="note why">${st.why}</div>`}
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}
      <table class="t">
        <tbody>
          ${b.status === "shipped" && b.allocations
            ? b.allocations.map(a => html`
                <tr key=${a.parcel_id}>
                  <th>${nameOf(a.parcel_id)}<div class="hint">tính cước trên ${grams(a.chargeable_g)}</div></th>
                  <td class="amount">${money(a.freight)}</td>
                </tr>`)
            : b.parcels.map(id => {
                const p = parcelById[id];
                return html`
                  <tr key=${id}>
                    <th>${nameOf(id)}<div class="hint">${p && p.actual ? `cân thật ${grams(p.actual.weight_g)}` : ""}</div></th>
                    <td class="amount"></td>
                  </tr>`;
              })}
        </tbody>
      </table>
      ${b.status === "open" && html`
        <div class="actions">
          <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !b.parcels.length} onClick=${close}>Dán kín lô</button>
          <span class="dim">Dán kín rồi thì không thêm kiện được nữa.</span>
        </div>`}
      ${b.status === "closed" && html`
        <${React.Fragment}>
          <div class="grid2">
            <${Field} ...${{ label: `Cước hãng bay cho cả lô (${FREIGHT_CURRENCY})`, value: freight, placeholder: "27.50",
              onChange: e => setFreight(e.target.value),
              hint: "Đúng số trên hoá đơn của hãng bay. Hệ thống chia nó cho từng kiện theo cân tính cước." }} />
          </div>
          <div class="actions">
            <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !freight.trim()} onClick=${ship}>Xác nhận lô đã bay</button>
          </div>
        <//>`}
    <//>`;
}

/* ── one order on its way to the door ────────────────────────────────────── */
function DeliverSheet({ order: o, facts, onDone }) {
  const act = useAction();
  return html`
    <${Sheet} label=${o.product_name || "đơn chờ giao"}>
      <div class="sheet-head"><h2>${o.product_name || "Sản phẩm chưa có tên"}</h2></div>
      <${Journey} order=${o} viewer="staff" facts=${facts} />
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}
      <div class="note why">Đã thu đủ cả cọc lẫn phần còn lại. Bấm khi kiện đã tới tay khách: đơn khép lại ở đây.</div>
      <div class="actions">
        <button class="primary" aria-busy=${act.busy} disabled=${act.busy}
          onClick=${() => act.go(async () => {
            await call("POST", `/orders/${o.order_id}/deliver`, { as: "operator" });
            await onDone();
          })}>Đã giao tận tay khách</button>
      </div>
    <//>`;
}

/* ── a queue on the left ──────────────────────────────────────────────────── */
function Queue({ title, icon, count, empty, rows, selected, onPick, closedLabel }) {
  const open = rows ? rows.filter(r => !r.closed) : null;
  const closed = rows ? rows.filter(r => r.closed) : [];
  const item = r => html`
    <li key=${r.key}>
      <button class=${"pick" + (r.closed ? " closed" : "")}
        aria-current=${selected === r.key ? "true" : undefined} onClick=${() => onPick(r.key)}>
        ${r.name}
        <span class=${"sub" + (r.yours ? " yours" : "")}>${r.sub}</span>
        <span class="go"><${Icon} name="chevron" size=${16} /></span>
      </button>
    </li>`;
  return html`
    <div class="queue">
      <div class="section-head"><${Icon} name=${icon} /><h2>${title}</h2>
        <span class="spacer"></span><span class=${"count" + (count > 0 ? " hot" : "")} key=${count}>${count}</span></div>
      ${!rows && html`<${Skeleton} rows=${2} />`}
      ${open && !open.length && html`<div class="empty">${empty}</div>`}
      ${open && open.length > 0 && html`<ul>${open.map(item)}</ul>`}
      ${closed.length > 0 && html`
        <div class="queue-done">
          <div class="queue-done-label">${closedLabel || "Vừa xong trong phiên này"} <span>${closed.length}</span></div>
          <ul>${closed.map(item)}</ul>
        </div>`}
    </div>`;
}

/* ── the screen ──────────────────────────────────────────────────────────── */
const LANE = "us_forwarder";      // the one lane seeded today; POST /lanes adds more
const FREIGHT_CURRENCY = "USD";   // the airline bills the us_forwarder lane in dollars

function App() {
  const [toasts, pushToast, dismiss] = useToasts();
  const [shops, setShops] = useState({});
  const [categories, setCategories] = useState({});
  const [token, setToken] = useState(getTokens().operator);
  const [picked, setPicked] = useState(null);
  const openRef = useRef(null);

  const loadRef = useCallback(async () => {
    const [list, cats] = await Promise.all([
      call("GET", "/merchants", { as: "operator" }),
      call("GET", "/categories", { as: "operator" }),
    ]);
    setShops(Object.fromEntries(list.map(m => [m.id, m])));
    setCategories(Object.fromEntries(cats.map(c => [c.code, c])));
  }, []);
  useEffect(() => { loadRef().catch(() => {}); }, [loadRef]);

  const loadQueue = useCallback(() => call("GET", "/product-queue", { as: "operator" }), []);
  const loadTasks = useCallback(() => call("GET", "/purchase-tasks", { as: "operator" }), []);
  // Every order, once: the money and delivery queues are filters of it, and
  // every other sheet finds its own order here to draw the journey.
  const loadOrders = useCallback(() => call("GET", "/orders", { as: "operator" }), []);
  // There is no "list batches" route, but every parcel knows its batch, so the
  // batches being packed are found through the parcels that are in them.
  const loadWarehouse = useCallback(async () => {
    const parcels = await call("GET", "/parcels", { as: "operator" });
    const ids = [...new Set(parcels.map(p => p.batch_id).filter(Boolean))];
    const batches = await Promise.all(ids.map(id => call("GET", `/batches/${id}`, { as: "operator" })));
    return { parcels, batches };
  }, []);

  const [queue, queueErr, reloadQueue] = usePoll(loadQueue, POLL_MS, (next, prev) => {
    if (!prev) return;
    for (const item of next) {
      const before = prev.find(p => p.product_id === item.product_id);
      if (!before) pushToast("Khách vừa gửi link", `${item.name}: việc đầu tiên là nhập size`);
    }
  });
  // GET /purchase-tasks lists only OPEN work, and GET /parcels only what has
  // not flown, so finished work vanishes from the API the moment it is done.
  // The desk remembers what it saw this session, so a finished sheet can stay
  // open and show where the order went next.
  const seen = useRef({ tasks: {}, taskInfo: {}, parcels: {}, batches: {} });
  const [tasks, , reloadTasks] = usePoll(loadTasks, POLL_MS, (next, prev) => {
    for (const t of next) seen.current.tasks[t.id] = t;
    if (!prev) return;
    const open = next.filter(t => t.status === "open").length;
    const was = prev.filter(t => t.status === "open").length;
    if (open > was) pushToast("Có đơn đã trả cọc", "Vào mua hộ khách ở hàng Việc đi mua.");
  });
  const [orders, , reloadOrders] = usePoll(loadOrders, POLL_MS, (next, prev) => {
    if (!prev) return;
    if (owed(next).length > owed(prev).length) pushToast("Có đơn chờ tiền", "Xem hàng Tiền chờ thu.");
  });
  const [warehouse, , reloadWarehouse] = usePoll(loadWarehouse, POLL_MS, (next, prev) => {
    for (const p of next.parcels) seen.current.parcels[p.id] = p;
    for (const b of next.batches) seen.current.batches[b.id] = b;
    if (!prev) return;
    const was = new Set(prev.parcels.map(p => p.id));
    if (next.parcels.some(p => !was.has(p.id))) pushToast("Có kiện sắp về kho", "Xem hàng Kho Denver.");
  });

  const reload = useCallback(async () => {
    await Promise.all([reloadQueue(), reloadTasks(), reloadOrders(), reloadWarehouse()]);
  }, [reloadQueue, reloadTasks, reloadOrders, reloadWarehouse]);
  // After an action the read side is a relay tick behind: ask again a few
  // times in quick succession instead of leaving the screen stale until the
  // next 4-second poll.
  const settle = useCallback(async () => {
    await reload();
    for (const ms of [250, 700, 1500]) setTimeout(() => { reload(); }, ms);
  }, [reload]);

  const orderById = useMemo(() => Object.fromEntries((orders || []).map(o => [o.order_id, o])), [orders]);
  const owedOrders = useMemo(() => (orders ? owed(orders) : null), [orders]);
  const toDeliver = useMemo(() => (orders ? orders.filter(o => o.status === "in_transit" && o.balance_paid) : null), [orders]);
  const openTasks = tasks ? tasks.filter(t => t.status === "open") : null;
  const doneTasks = tasks ? Object.values(seen.current.tasks)
    .filter(t => !tasks.some(x => x.id === t.id))
    .map(t => ({ ...t, ...(seen.current.taskInfo[t.order_id] || {}), status: taskStatusOf(orderById[t.order_id]) }))
    .reverse().slice(0, 5) : [];
  const parcelById = { ...seen.current.parcels, ...Object.fromEntries((warehouse ? warehouse.parcels : []).map(p => [p.id, p])) };
  const liveBatchIds = new Set((warehouse ? warehouse.batches : []).map(b => b.id));
  const batches = warehouse ? warehouse.batches : [];
  const flownBatches = Object.values(seen.current.batches).filter(b => b.status === "shipped" || !liveBatchIds.has(b.id));
  const openBatch = batches.find(b => b.status === "open");

  // Everything the API told this desk about one order, for its journey strip.
  const factsFor = orderId => {
    const info = seen.current.taskInfo[orderId] || {};
    const parcel = Object.values(parcelById).find(p => p.order_id === orderId);
    const batch = Object.values(seen.current.batches).find(b => b.allocations && b.allocations.some(a => a.order_id === orderId));
    const share = batch && batch.allocations.find(a => a.order_id === orderId);
    return {
      boughtAt: info.closed_at, paid: info.paid,
      weighedG: parcel && parcel.actual ? parcel.actual.weight_g : 0, receivedAt: parcel && parcel.received_at,
      shippedAt: batch && batch.shipped_at, freight: share && share.freight,
    };
  };
  const nameOfOrder = id => (orderById[id] && orderById[id].product_name) || "Kiện hàng";

  const rows = {
    product: queue && queue.map(item => {
      const n = STEP_ORDER.indexOf(item.next_step);
      return { key: "p:" + item.product_id, name: item.name, yours: true,
        sub: n >= 0 ? `Bước ${n + 1} trên 4: ${STEP_TITLE[item.next_step].toLowerCase()}` : "chờ đăng bán" };
    }),
    task: openTasks && [...openTasks, ...doneTasks].map(t => ({
      key: "t:" + t.id, name: t.product_name || "Sản phẩm chưa có tên", closed: t.status !== "open",
      yours: t.status === "open", sub: [t.variant_label, say(TASK, t.status).words].filter(Boolean).join(", "),
    })),
    warehouse: warehouse && [
      ...warehouse.parcels.filter(p => p.status === "expected" || p.status === "received").map(p => ({
        key: "k:" + p.id, name: nameOfOrder(p.order_id), yours: p.status === "received",
        sub: say(PARCEL, p.status).words,
      })),
      ...batches.map(b => ({
        key: "b:" + b.id, name: `Lô gom ${b.parcels.length} kiện`, yours: true, sub: say(BATCH, b.status).words,
      })),
      ...flownBatches.map(b => ({
        key: "b:" + b.id, name: `Lô gom ${b.parcels.length} kiện`, closed: true, sub: say(BATCH, "shipped").words,
      })),
    ],
    money: owedOrders && owedOrders.map(o => ({
      key: "m:" + o.order_id, name: o.product_name || "Sản phẩm chưa có tên", yours: true,
      sub: `${money(o.deposit_paid ? o.balance : o.deposit)}, ${o.deposit_paid ? "phần còn lại" : "cọc"}`,
    })),
    deliver: toDeliver && toDeliver.map(o => ({
      key: "g:" + o.order_id, name: o.product_name || "Sản phẩm chưa có tên", yours: true, sub: "đã thu đủ, chờ giao",
    })),
  };
  const QUEUES = ["product", "task", "warehouse", "money", "deliver"];

  // The next piece of work, in the order the job runs: make it sellable, buy
  // it, weigh and fly it, take the money, hand it over. Finished work is
  // never "next".
  const all = QUEUES.flatMap(q => rows[q] || []);
  const firstWork = all.filter(r => !r.closed)[0];
  const current = all.find(r => r.key === picked) ? picked : firstWork ? firstWork.key : null;

  const pick = key => {
    setPicked(key);
    // stacked layout: the sheet is below the queues, so bring it into view
    if (window.matchMedia("(max-width: 1023px)").matches && openRef.current) {
      requestAnimationFrame(() => openRef.current.scrollIntoView({ block: "start" }));
    }
  };

  // After buying, the task leaves the open list but the sheet stays, pinned,
  // so the strip can be seen moving on; the next job is one button away.
  // The task itself is read back once: it now carries the date and the real
  // price, which the strip shows in the "Đã mua" box.
  const bought = key => async task => {
    setPicked(key);
    try { seen.current.taskInfo[task.order_id] = await call("GET", `/purchase-tasks/${task.id}`, { as: "operator" }); } catch { /* the strip just stays blank */ }
    await settle();
  };
  // Packing and shipping move the desk to the batch; shipping reads it back
  // once more, because after it flies it drops out of GET /parcels.
  const toBatch = async key => { if (key) setPicked(key); await settle(); };
  const shipped = id => async () => {
    setPicked("b:" + id);
    try { seen.current.batches[id] = await call("GET", `/batches/${id}`, { as: "operator" }); } catch { /* keep the old view */ }
    await settle();
  };
  const nextWork = firstWork && firstWork.key !== current ? firstWork : null;
  const nextButton = nextWork && html`
    <div class="actions"><button class="primary" onClick=${() => pick(nextWork.key)}>Mở việc kế tiếp: ${nextWork.name}</button></div>`;

  const kind = current ? current.slice(0, 1) : null;
  const id = current ? current.slice(2) : null;
  let sheet = null;
  if (kind === "p") {
    const item = queue.find(x => x.product_id === id);
    sheet = html`<${ProductSheet} key=${current} item=${item} shops=${shops} categories=${categories} onDone=${settle} />`;
  } else if (kind === "t") {
    const t = [...openTasks, ...doneTasks].find(x => x.id === id);
    sheet = html`<${TaskSheet} key=${current} task=${t} order=${orderById[t.order_id]} facts=${factsFor(t.order_id)}
      onDone=${bought(current)} next=${nextButton} />`;
  } else if (kind === "k") {
    const p = parcelById[id];
    sheet = html`<${ParcelSheet} key=${current} parcel=${p} order=${orderById[p.order_id]} facts=${factsFor(p.order_id)}
      openBatch=${openBatch && openBatch.id} onDone=${toBatch} />`;
  } else if (kind === "b") {
    const b = batches.find(x => x.id === id) || seen.current.batches[id];
    sheet = html`
      <div>
        <${BatchSheet} key=${current} batch=${b} orderById=${orderById} parcelById=${parcelById}
          onDone=${b.status === "closed" ? shipped(b.id) : settle} />
        ${b.status === "shipped" && nextButton && html`<div style=${{ marginTop: "16px" }}>${nextButton}</div>`}
      </div>`;
  } else if (kind === "m") {
    const o = owedOrders.find(x => x.order_id === id);
    sheet = html`<${MoneySheet} key=${current} order=${o} facts=${factsFor(o.order_id)} onDone=${settle} />`;
  } else if (kind === "g") {
    const o = toDeliver.find(x => x.order_id === id);
    sheet = html`<${DeliverSheet} key=${current} order=${o} facts=${factsFor(o.order_id)} onDone=${settle} />`;
  }

  const openCount = q => (rows[q] || []).filter(r => !r.closed).length;
  const waiting = QUEUES.reduce((n, q) => n + openCount(q), 0);

  return html`
    <${React.Fragment}>
      <${Top} role="Bạn đang xem với vai nhân viên" waiting=${waiting}>
        <a href="./customer.html">Màn hình khách</a>
        <a href="./console.html">Bảng kiểm API</a>
      <//>
      <div class="wrap wide">
        <${PageHead} title="Bàn làm việc"
          lead="Mỗi việc một phiếu. Làm xong một việc, phiếu kế tiếp tự mở theo đúng thứ tự: làm cho bán được, đi mua, cân và gom lô, thu tiền, giao."
          stats=${[
            { label: "việc đang chờ", value: waiting, hot: waiting > 0 },
            { label: "đơn đang bay về", value: orders ? orders.filter(o => o.status === "in_transit").length : 0 },
          ]} />
        ${queueErr && html`<${Problem}>${friendly(queueErr)}<//>`}
        <div class="desk">
          <nav class="queues enter" aria-label="Hàng đợi việc">
            <${Queue} title="Món chờ xử lý" icon="box" count=${openCount("product")} rows=${rows.product}
              empty="Không còn món nào. Khi khách dán link, nó xuất hiện ở đây trong vài giây."
              selected=${current} onPick=${pick} />
            <${Queue} title="Việc đi mua" icon="bag" count=${openCount("task")} rows=${rows.task}
              empty="Chưa có việc đi mua. Việc chỉ mở khi khách trả cọc." closedLabel="Vừa mua xong trong phiên này"
              selected=${current} onPick=${pick} />
            <${Queue} title="Kho Denver" icon="scale" count=${openCount("warehouse")} rows=${rows.warehouse}
              empty="Kho chưa chờ kiện nào. Kiện xuất hiện ở đây khi một món đã mua xong." closedLabel="Lô vừa bay trong phiên này"
              selected=${current} onPick=${pick} />
            <${Queue} title="Tiền chờ thu" icon="wallet" count=${openCount("money")} rows=${rows.money}
              empty="Không có đơn nào đang chờ tiền."
              selected=${current} onPick=${pick} />
            <${Queue} title="Chờ giao" icon="home" count=${openCount("deliver")} rows=${rows.deliver}
              empty="Chưa có đơn nào đã thu đủ và chờ giao."
              selected=${current} onPick=${pick} />
          </nav>
          <main class="open" ref=${openRef}>
            ${sheet || html`<${Empty} icon="coffee">Hết việc. Hàng đợi bên cạnh tự cập nhật, không cần tải lại.<//>`}
          </main>
        </div>

        <${Keys} token=${token} onSave=${t => { setTokens({ operator: t }); setToken(t); reload(); }}>
          Chìa khoá thay cho đăng nhập. Nó cũng là cách hệ thống ghi lại <b>ai</b> đã xác nhận sản phẩm và <b>ai</b> đã nhận tiền, nên đừng dùng chung chìa với người khác. Bản chạy thử để sẵn <b>dev-operator</b>; bản thật mỗi người một chìa.
        <//>
      </div>
      <${Toasts} items=${toasts} onDismiss=${dismiss} />
    <//>`;
}

/** taskStatusOf reads how a remembered task ended from its order. */
function taskStatusOf(o) {
  if (!o) return "confirmed";
  return o.status === "purchase_failed" || o.status === "cancelled" ? "failed" : "confirmed";
}

/** owed: orders a person has to take money for right now. */
function owed(orders) {
  return orders.filter(o => o.status === "awaiting_deposit" || (o.status === "in_transit" && !o.balance_paid));
}

window.ReactDOM.createRoot(document.getElementById("root")).render(html`<${App} />`);
