// staff.js — the screen a person doing the work actually uses.
//
// What it is NOT: one button per endpoint. That was the first console, and it
// failed for a reason worth writing down — it showed the shape of the API
// instead of the shape of the job. Nobody thinks "add variant, confirm
// listing, measure, publish". They think "a customer sent me a link, make it
// sellable".
//
// So every pasted product is one card with a numbered checklist, and exactly
// one step is actionable at a time. A finished step shows what it recorded,
// not a greyed-out button: "US 9 · black" is proof the step happened.

import { React, html, Top, Card, Field, Steps, Toasts, useToasts, usePoll, useAction } from "./ui.js";
import { call, money, retryOn, getTokens, setTokens } from "./portage.js";
import { TASK, say, friendly, categoryWords } from "./words.js";

const { useState, useEffect, useCallback, useMemo } = React;

const POLL_MS = 4000;

/* ── one product, one checklist ───────────────────────────────────────────── */
function ProductCard({ item, shops, categories, onDone }) {
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

  const addSize = () => act.go(async () => {
    await post("/variants", { size: size.trim(), color: colour.trim(), merchant_ref: ref.trim() });
    setSize(""); setColour(""); setRef("");
    await onDone();
  });
  const confirm = () => act.go(async () => { await post("/confirm-listing"); await onDone(); });
  const measure = () => act.go(async () => {
    await post("/measure", {
      weight_g: Number(g), length_mm: Number(l), width_mm: Number(w), height_mm: Number(h),
    });
    await onDone();
  });
  // publish can race the relay for nothing else, but a fresh variant may not
  // be visible to this handler yet on a cold start, so one retryable code.
  const publish = () => act.go(async () => {
    await retryOn(["no_variants"], () => post("/publish"));
    await onDone();
  });

  const stateOf = mine => (item.next_step === mine ? "now" : stepIsDone(item, mine) ? "done" : "todo");

  const steps = [
    {
      key: "variant", title: "Khách mua size nào", state: stateOf("variant"),
      recorded: html`Đã có <b>${item.variant_label || "một phiên bản"}</b>`,
      why: item.requested_variant
        ? html`Ô Size đã điền sẵn <b>đúng chữ khách viết</b>. Việc của bạn là mở trang shop xem có loại đó
            không, rồi sửa lại theo cách shop ghi nếu khác. Chữ cuối cùng ở đây in nguyên văn lên phiếu đi mua.
            <br />Coi kỹ <b>chữ cái sau số</b>: cùng số 1 mà <span class="mono">1Y</span> (thiếu niên) và
            <span class="mono">1C</span> (trẻ nhỏ) là hai đôi khác nhau, mua nhầm là đổi cả đơn.`
        : html`Khách mua <b>một hình thức cụ thể</b>, không mua "một đôi giày". Chữ bạn gõ ở đây là chữ của
            shop, và nó sẽ hiện nguyên văn trên phiếu đi mua. Giữ nguyên cả chữ cái của hệ size:
            <span class="mono">M</span> nam, <span class="mono">W</span> nữ, <span class="mono">Y</span>
            thiếu niên, <span class="mono">C</span> trẻ nhỏ, còn áo thì <span class="mono">S/M/L</span>.`,
      children: html`
        <div class="grid3">
          <${Field} ...${{ label: "Size", value: size, placeholder: "US 9 · M 8 / W 9.5 · 1Y · 10C · L",
            onChange: e => setSize(e.target.value),
            hint: item.requested_variant ? `khách viết: ${item.requested_variant}` : "" }} />
          <${Field} ...${{ label: "Màu", value: colour, placeholder: "black",
            onChange: e => setColour(e.target.value) }} />
          <${Field} ...${{ label: "Mã của shop", value: ref, placeholder: "không có thì để trống",
            onChange: e => setRef(e.target.value) }} />
        </div>
        <div class="actions">
          <button class="primary" disabled=${act.busy || (!size.trim() && !colour.trim())}
            onClick=${addSize}>Lưu size này</button>
          <span class="dim">Trùng size và màu sẽ bị từ chối, không phân biệt hoa thường hay khoảng trắng.</span>
        </div>`,
      after: "Cần ít nhất một size trước khi đăng bán được.",
    },
    {
      key: "listing", title: "Xác nhận đúng sản phẩm", state: stateOf("listing"),
      recorded: "Đã có người xác nhận",
      why: html`Đây là <b>chữ ký của bạn</b>: tôi đã mở trang đó và đúng là sản phẩm này, của shop này, ngành hàng này. Khách dán link thì chưa ai kiểm, nên thiếu bước này hệ thống sẽ không cho đăng bán.`,
      children: html`
        <div class="actions">
          <button class="primary" disabled=${act.busy} onClick=${confirm}>Tôi đã xem trang, đúng sản phẩm</button>
          ${item.source && html`<a href=${item.source} target="_blank" rel="noreferrer noopener">Mở trang shop</a>`}
        </div>`,
      after: "Ai xác nhận được lấy từ chìa khoá đang dùng, không phải từ một ô nhập.",
    },
    {
      key: "measure", title: "Cân và đo thật", state: stateOf("measure"),
      recorded: "Đã ghi số cân",
      why: html`Trang shop <b>không công bố</b> cân nặng đóng gói, nên đây là số mình tự tích luỹ. Số điền sẵn là <b>hộp mẫu của nhóm ${categoryWords(item.category)}</b>, chỉ để đỡ gõ — hãy sửa thành số cân thật của hộp trước mặt, vì cước tính theo nó.`,
      children: html`
        <div class="grid2">
          <${Field} ...${{ label: "Cân nặng (g)", value: g, inputMode: "numeric", onChange: e => setG(e.target.value) }} />
          <${Field} ...${{ label: "Dài (mm)", value: l, inputMode: "numeric", onChange: e => setL(e.target.value) }} />
          <${Field} ...${{ label: "Rộng (mm)", value: w, inputMode: "numeric", onChange: e => setW(e.target.value) }} />
          <${Field} ...${{ label: "Cao (mm)", value: h, inputMode: "numeric", onChange: e => setH(e.target.value) }} />
        </div>
        <div class="note">Cước lấy số lớn hơn giữa cân thật và cân quy đổi, rồi làm tròn lên bước 500 g.
          Với hộp ${l}×${w}×${h} mm thì cân quy đổi là ${volumetric(l, w, h)} g.</div>
        <div class="actions">
          <button class="primary" disabled=${act.busy || !g || !l || !w || !h} onClick=${measure}>Lưu số cân</button>
        </div>`,
      after: "Thiếu một cạnh cũng không tính được cước, nên hệ thống đòi đủ bốn số.",
    },
    {
      key: "publish", title: "Đăng bán", state: stateOf("publish"),
      recorded: "Đã đăng bán",
      why: html`Từ giây này khách mới xin được báo giá. Trước đó mọi thứ chỉ là nháp, và báo giá trên dữ liệu chưa ai kiểm là báo giá sai tiền thật.`,
      children: html`
        <div class="actions">
          <button class="primary" disabled=${act.busy} onClick=${publish}>Đăng bán</button>
          <span class="dim">Sau bước này khách sẽ thấy nút xin báo giá.</span>
        </div>`,
      after: "Mở được khi ba bước trên xong.",
    },
  ];

  return html`
    <div class="item">
      <div class="item-head">
        <span class="name">${item.name}</span>
        <span class="pill flat">${categoryWords(item.category)}</span>
        ${item.sourced_by === "customer"
          ? html`<span class="pill info">khách gửi</span>`
          : html`<span class="pill flat">tự thêm</span>`}
        <span class="spacer"></span>
        <span class="dim">Giá trên web: ${money(item.price)}</span>
      </div>
      <div class="item-body">
        <div class="meta" style=${{ marginBottom: "10px" }}>
          <span>Shop: <b>${shop ? shop.name : "—"}</b></span>
          ${item.source && html`<a href=${item.source} target="_blank" rel="noreferrer noopener">${trim(item.source)}</a>`}
        </div>
        ${item.requested_variant
          ? html`<div class="note why">Khách yêu cầu: <b>${item.requested_variant}</b>. Mở trang shop kiểm
              xem có đúng loại đó không. Shop ghi khác thì sửa lại theo chữ của shop, vì phiếu đi mua sẽ
              in đúng chữ bạn nhập.</div>`
          : html`<div class="note">Khách <b>không ghi</b> muốn loại nào. Nếu sản phẩm có nhiều size thì hỏi
              lại khách trước khi nhập, đừng đoán.</div>`}
        ${act.problem && html`<div class="note bad">${friendly(act.problem)}</div>`}
        <${Steps} steps=${steps} />
      </div>
    </div>`;
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

/* ── things to go and buy ─────────────────────────────────────────────────── */
function BuyingQueue({ tasks, onDone }) {
  const act = useAction();
  const [ref, setRef] = useState({});
  const [paid, setPaid] = useState({});

  if (!tasks || !tasks.length) {
    return html`<div class="empty">Chưa có việc đi mua. Việc chỉ mở khi khách trả cọc.</div>`;
  }
  return html`
    <${React.Fragment}>
      ${act.problem && html`<div class="note bad">${friendly(act.problem)}</div>`}
      ${tasks.map(t => html`
        <div class="item" key=${t.id}>
          <div class="item-head">
            <span class="name">${t.product_name || "—"}</span>
            ${t.variant_label && html`<span class="pill info">${t.variant_label}</span>`}
            <span class="spacer"></span>
            <span class="pill ${say(TASK, t.status).tone}">${say(TASK, t.status).words}</span>
          </div>
          <div class="item-body">
            <div class="meta" style=${{ marginBottom: "10px" }}>
              ${t.variant_ref && html`<span>Mã shop: <b class="mono">${t.variant_ref}</b></span>`}
              <span>Trả bằng: <b>${t.currency}</b></span>
              ${t.source && html`<a href=${t.source} target="_blank" rel="noreferrer noopener">Mở trang để mua</a>`}
            </div>
            ${t.status === "open" && html`
              <div class="grid2">
                <${Field} ...${{ label: "Mã đơn ở shop", value: ref[t.id] || "", placeholder: "NK-20260905-001",
                  onChange: e => setRef({ ...ref, [t.id]: e.target.value }),
                  hint: "Số đơn shop gửi trong mail xác nhận. Cần nó để đối chiếu khi kiện hàng về kho, và để khiếu nại nếu shop giao sai." }} />
                <${Field} ...${{ label: `Đã trả thật (${t.currency})`, value: paid[t.id] || "", placeholder: "163.22",
                  onChange: e => setPaid({ ...paid, [t.id]: e.target.value }),
                  hint: "Số tiền thật đã trả, không phải số đã báo khách. Hai số này lệch nhau là chuyện thường, và hệ thống cần cả hai để biết đơn lời hay lỗ." }} />
              </div>
              <div class="actions">
                <button class="primary" disabled=${act.busy || !ref[t.id] || !paid[t.id]}
                  onClick=${() => act.go(async () => {
                    await call("POST", `/purchase-tasks/${t.id}/confirm`, { as: "operator",
                      body: { reference: ref[t.id].trim(), paid: paid[t.id].trim(), currency: t.currency } });
                    await onDone();
                  })}>Đã mua xong</button>
                <button class="danger" disabled=${act.busy}
                  onClick=${() => act.go(async () => {
                    await call("POST", `/purchase-tasks/${t.id}/fail`, { as: "operator",
                      body: { reason: "shop hết hàng" } });
                    await onDone();
                  })}>Không mua được</button>
                <span class="dim">Nhập bằng đúng loại tiền shop tính (${t.currency}), và mã đơn không được để trống.</span>
              </div>`}
          </div>
        </div>`)}
    <//>`;
}

/* ── orders waiting for money ─────────────────────────────────────────────────
 * Taking a payment is an operator action today: there is no payment gateway,
 * so a person confirms the transfer they can see in a bank app. The customer's
 * screen shows the amount and says who to expect it from.
 */
function MoneyQueue({ orders, onDone }) {
  const act = useAction();
  if (!orders || !orders.length) {
    return html`<div class="empty">Không có đơn nào đang chờ tiền.</div>`;
  }
  return html`
    <${React.Fragment}>
      ${act.problem && html`<div class="note bad">${friendly(act.problem)}</div>`}
      <div class="wrap-x">
        <table class="t">
          <thead><tr><th>Sản phẩm</th><th>Cần thu</th><th>Việc</th></tr></thead>
          <tbody>
            ${orders.map(o => html`
              <tr key=${o.order_id}>
                <td>${o.product_name || "—"}</td>
                <td style=${{ whiteSpace: "nowrap" }}><b>${money(o.deposit_paid ? o.balance : o.deposit)}</b>
                  <div class="hint">${o.deposit_paid
                    ? "phần còn lại, thu khi hàng đã bay"
                    : "một nửa tổng, thu trước khi đi mua"}</div></td>
                <td>
                  <button disabled=${act.busy}
                    onClick=${() => act.go(async () => {
                      const which = o.deposit_paid ? "balance" : "deposit";
                      const amount = o.deposit_paid ? o.balance : o.deposit;
                      await call("POST", `/orders/${o.order_id}/${which}`, { as: "operator",
                        body: { amount: amount.amount, currency: amount.currency } });
                      await onDone();
                    })}>Đã nhận đủ tiền</button>
                </td>
              </tr>`)}
          </tbody>
        </table>
      </div>
      <div class="note">Chỉ bấm khi tiền đã <b>thực sự</b> vào tài khoản. Phải đúng số, thiếu hay thừa một
        đồng hệ thống đều không nhận: cọc thiếu không phải một cam kết nhỏ hơn, và cọc thừa là một khoản
        phải hoàn mà không ai xin.</div>
    <//>`;
}

/* ── the screen ──────────────────────────────────────────────────────────── */
function App() {
  const [toasts, pushToast, dismiss] = useToasts();
  const [shops, setShops] = useState({});
  const [categories, setCategories] = useState({});
  const [token, setToken] = useState(getTokens().operator);

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
  const loadMoney = useCallback(async () => {
    const [awaiting, transit] = await Promise.all([
      call("GET", "/orders?status=awaiting_deposit", { as: "operator" }),
      call("GET", "/orders?status=in_transit", { as: "operator" }),
    ]);
    return [...awaiting, ...transit.filter(o => !o.balance_paid)];
  }, []);

  const [queue, queueErr, reloadQueue] = usePoll(loadQueue, POLL_MS, (next, prev) => {
    if (!prev) return;
    for (const item of next) {
      const before = prev.find(p => p.product_id === item.product_id);
      if (!before) pushToast("Khách vừa gửi link", `${item.name} — việc đầu tiên: nhập size`);
    }
  });
  const [tasks, , reloadTasks] = usePoll(loadTasks, POLL_MS, (next, prev) => {
    if (!prev) return;
    const open = next.filter(t => t.status === "open").length;
    const was = prev.filter(t => t.status === "open").length;
    if (open > was) pushToast("Có đơn đã trả cọc", "Vào mua hộ khách, xem phần Việc đi mua");
  });

  const [moneyRows, , reloadMoney] = usePoll(loadMoney, POLL_MS, (next, prev) => {
    if (!prev) return;
    if (next.length > prev.length) pushToast("Có đơn chờ tiền", "Xem phần Đơn chờ thu tiền");
  });

  const reload = useCallback(async () => {
    await Promise.all([reloadQueue(), reloadTasks(), reloadMoney()]);
  }, [reloadQueue, reloadTasks, reloadMoney]);

  const waiting = (queue ? queue.length : 0)
    + (tasks ? tasks.filter(t => t.status === "open").length : 0)
    + (moneyRows ? moneyRows.length : 0);

  return html`
    <${React.Fragment}>
      <${Top} title="Portage" role="Bạn đang xem với vai nhân viên" waiting=${waiting}>
        <a href="./customer.html">Màn hình khách →</a>
        <a href="./console.html" style=${{ marginLeft: "14px" }}>Bảng kiểm API</a>
      <//>
      <div class="wrap stack">
        ${queueErr && html`<div class="note bad">${friendly(queueErr)}</div>`}

        <${Card} title="Việc cần làm" right=${html`<span class="dim">${queue ? queue.length : 0} sản phẩm chưa đăng bán</span>`}>
          ${!queue && html`<div class="empty">Đang tải…</div>`}
          ${queue && !queue.length && html`<div class="empty">Không còn việc nào. Khi khách dán link, nó xuất hiện ở đây trong vài giây.</div>`}
          ${queue && queue.map(item => html`
            <${ProductCard} key=${item.product_id} item=${item} shops=${shops}
              categories=${categories} onDone=${reload} />`)}
        <//>

        <${Card} title="Đơn chờ thu tiền">
          <${MoneyQueue} orders=${moneyRows} onDone=${reload} />
        <//>

        <${Card} title="Việc đi mua">
          <${BuyingQueue} tasks=${tasks} onDone=${reload} />
        <//>

        <${Card} title="Chìa khoá của bạn">
          <div class="note">Chìa khoá thay cho đăng nhập. Nó cũng là cách hệ thống ghi lại <b>ai</b> đã xác
            nhận sản phẩm và <b>ai</b> đã nhận tiền, nên đừng dùng chung chìa với người khác. Bản chạy thử để
            sẵn <span class="mono">dev-operator</span>; bản thật mỗi người một chìa.</div>
          <div class="grid2">
            <${Field} ...${{ label: "Chìa khoá", value: token, onChange: e => setToken(e.target.value) }} />
          </div>
          <div class="actions">
            <button onClick=${() => { setTokens({ operator: token }); reload(); }}>Lưu và tải lại</button>
          </div>
        <//>
      </div>
      <${Toasts} items=${toasts} onDismiss=${dismiss} />
    <//>`;
}

window.ReactDOM.createRoot(document.getElementById("root")).render(html`<${App} />`);
