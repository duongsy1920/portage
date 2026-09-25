// customer.js — the screen a customer uses.
//
// Three questions, in the order a person actually has them: can you buy me
// this, how much, and where is it. Everything else the API can do is somebody
// else's screen.
//
// No uuid is ever typed here. The shop and the category come from the API as
// lists, the size comes from what staff published, and the ids the API needs
// travel in the background.

import { React, html, Top, Section, Sheet, Problem, Field, Select, Journey, Keys, Toasts, PageHead,
  Skeleton, Empty, Icon, useToasts, usePoll, useAction } from "./ui.js";
import { call, money, retryOn, getTokens, setTokens } from "./portage.js";
import { QUOTE, ORDER, GOODS, NEXT_STEP, say, friendly, categoryWords } from "./words.js";

const { useState, useEffect, useCallback, useMemo } = React;

const POLL_MS = 4000;
const LANE = "us_forwarder"; // the one lane seeded today; POST /lanes adds more
const QUOTES_KEY = "portage.customer.quotes.v1";

/* Quotes a customer has asked for, kept in the browser: the API has no read
 * model of "quotes I asked for", and inventing one to hold a number the
 * customer can re-request in a click would be the wrong trade. */
function loadQuotes() {
  try { return JSON.parse(localStorage.getItem(QUOTES_KEY) || "{}"); } catch { return {}; }
}
function saveQuotes(q) {
  try { localStorage.setItem(QUOTES_KEY, JSON.stringify(q)); } catch { /* private mode */ }
}

/* The shop and category picked last time. Someone who shops for shoes picks
 * "giày dép" every time, and a form that forgets it on reload makes them. */
const PASTE_KEY = "portage.customer.paste.v1";
function lastPicked() {
  try { return JSON.parse(localStorage.getItem(PASTE_KEY) || "{}"); } catch { return {}; }
}
function rememberPicked(patch) {
  try { localStorage.setItem(PASTE_KEY, JSON.stringify({ ...lastPicked(), ...patch })); } catch { /* private mode */ }
}

/* ── paste a link ─────────────────────────────────────────────────────────── */
function PasteForm({ shops, categories, onSent }) {
  const act = useAction();
  const canPaste = useMemo(
    () => shops.filter(s => s.status === "active" && s.sourcing.includes("customer")),
    [shops]);
  const [shopId, setShopId] = useState("");
  const [category, setCategory] = useState("");
  const [url, setUrl] = useState("");
  const [name, setName] = useState("");
  const [price, setPrice] = useState("");
  const [wanted, setWanted] = useState("");

  useEffect(() => {
    if (shopId || !canPaste.length) return;
    const last = lastPicked().shop;
    setShopId(canPaste.some(x => x.id === last) ? last : canPaste[0].id);
  }, [canPaste, shopId]);
  useEffect(() => {
    if (category || !categories.length) return;
    const last = lastPicked().category;
    setCategory(categories.some(c => c.code === last) ? last : categories[0].code);
  }, [categories, category]);

  const shop = canPaste.find(s => s.id === shopId);
  const cat = categories.find(c => c.code === category);

  const send = () => act.go(async () => {
    await call("POST", "/products", {
      as: "customer",
      body: {
        name: name.trim(), merchant_id: shopId, category,
        source_url: url.trim(), price: price.trim(), currency: shop.currency,
        requested_variant: wanted.trim(),
      },
    });
    setUrl(""); setName(""); setPrice(""); setWanted("");
    await onSent();
  });

  if (!canPaste.length) {
    return html`<${Empty} icon="bag">Chưa có shop nào nhận link khách tự gửi. Nhân viên cần thêm shop trước.<//>`;
  }
  return html`
    <${React.Fragment}>
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}
      <div class="grid2">
        <${Select} ...${{ label: "Shop", value: shopId, onChange: e => { setShopId(e.target.value); rememberPicked({ shop: e.target.value }); },
          options: canPaste.map(s => ({ value: s.id, label: s.name })),
          hint: shop ? `Shop này bán bằng ${shop.currency}` : "" }} />
        <${Select} ...${{ label: "Ngành hàng", value: category, onChange: e => { setCategory(e.target.value); rememberPicked({ category: e.target.value }); },
          options: categories.map(c => ({ value: c.code, label: categoryWords(c.code) })),
          hint: cat ? `Chọn đúng nhóm vì nó quyết định thuế và giá cước. Khi chưa ai cân hộp thật, bên mình tạm tính theo hộp mẫu của nhóm này: ${cat.weight_g} g, ${cat.length_mm}×${cat.width_mm}×${cat.height_mm} mm.` : "" }} />
      </div>
      <${Field} ...${{ label: "Link sản phẩm", value: url, placeholder: "https://…",
        onChange: e => setUrl(e.target.value),
        hint: "Dán nguyên link trang sản phẩm. Bên mình không tự đọc trang, nên hai ô dưới bạn điền giúp." }} />
      <div class="grid2">
        <${Field} ...${{ label: "Tên sản phẩm", value: name, placeholder: "Air Force 1 '07",
          onChange: e => setName(e.target.value) }} />
        <${Field} ...${{ label: `Giá trên web (${shop ? shop.currency : "?"})`, value: price, placeholder: "115.00",
          onChange: e => setPrice(e.target.value),
          hint: "Đúng số trên trang, chưa gồm thuế và cước" }} />
      </div>
      <${Field} ...${{ label: "Size hoặc màu bạn muốn", value: wanted,
        placeholder: "US 9, M 8 / W 9.5, 1Y, 10C, L, bản 256GB",
        onChange: e => setWanted(e.target.value),
        hint: "Copy đúng chữ trên trang shop, kể cả chữ cái của hệ size. Nhân viên đọc dòng này để biết mua cái nào, nên bỏ trống là họ phải hỏi lại bạn." }} />
      <div class="note">Chữ cái đứng cạnh số quyết định đôi nào, đừng bỏ: <b>M</b> nam, <b>W</b> nữ, <b>Y</b> thiếu niên, <b>C</b> trẻ nhỏ, áo thì <b>S/M/L</b>. Cùng số 1 mà <b>1Y</b> và <b>1C</b> là
        hai đôi khác nhau. Trang nào ghi hai hệ cùng lúc, ví dụ <b>M 8 / W 9.5</b>, thì copy nguyên cả dòng.</div>
      <div class="actions">
        <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !url.trim() || !name.trim() || !price.trim()}
          onClick=${send}>Gửi cho nhân viên</button>
        <span class="dim">Gửi xong nhân viên sẽ kiểm rồi mới báo giá được.</span>
      </div>
    <//>`;
}

/* ── one thing I asked for ───────────────────────────────────────────────── */
function MyProduct({ item, anchor, quotes, setQuotes, onChanged, ordered }) {
  const act = useAction();
  // An order for this product may exist without this browser knowing (placed
  // from another device, or by staff on the customer's behalf): the order list
  // says so by product, and it outranks what localStorage remembers.
  const saved = { ...(quotes[item.product_id] || {}), ...(ordered ? { order: true } : {}) };
  const [variantId, setVariantId] = useState(saved.variant || "");
  const [quote, setQuote] = useState(null);

  // Default to the size the customer asked for, not to the first one staff
  // happened to create. Matching ignores case and every space, the same way
  // catalog decides two variants are the same thing.
  useEffect(() => {
    if (variantId || !item.variants.length) return;
    const key = s => String(s || "").toLowerCase().replace(/\s+/g, "");
    const wish = key(item.requested_variant);
    const match = wish && item.variants.find(v => key(v.label).includes(wish));
    setVariantId((match || item.variants[0]).id);
  }, [item.variants, item.requested_variant, variantId]);

  // Re-read a quote we already asked for, so a reload does not lose the price.
  useEffect(() => {
    if (!saved.quote) { setQuote(null); return; }
    call("GET", `/quotes/${saved.quote}`, { as: "customer" }).then(setQuote).catch(() => setQuote(null));
  }, [saved.quote]);

  const ask = () => act.go(async () => {
    const q = await retryOn(["listing_not_found"], () =>
      call("POST", "/quotes", { as: "customer", body: { product_id: item.product_id, lane: LANE } }));
    const next = { ...quotes, [item.product_id]: { quote: q.id, variant: variantId } };
    setQuotes(next); saveQuotes(next);
  });

  const agree = () => act.go(async () => {
    await call("POST", `/quotes/${saved.quote}/accept`, { as: "customer" });
    const order = await retryOn(["quote_not_accepted", "variant_unknown"], () =>
      call("POST", "/orders", { as: "customer", body: { quote_id: saved.quote, variant_id: variantId } }));
    const next = { ...quotes, [item.product_id]: { ...saved, order: order.id } };
    setQuotes(next); saveQuotes(next);
    await onChanged();
  });

  const waiting = NEXT_STEP[item.next_step];
  // The one piece being priced right now is lifted onto a sheet; the rest are
  // rows. Once the order exists the product has done its job on this list.
  const lifted = item.published && !saved.order;

  const body = html`
    <${React.Fragment}>
      <div class="row-head">
        ${lifted ? html`<h3>${item.name}</h3>` : html`<span class="name">${item.name}</span>`}
        <span class="pill">${categoryWords(item.category)}</span>
        <span class="spacer"></span>
        ${saved.order
          ? html`<span class="pill info">đã đặt đơn</span>`
          : item.published
            ? html`<span class="pill ok">sẵn sàng báo giá</span>`
            : html`<span class="pill warn">${waiting || "đang xử lý"}</span>`}
      </div>
      <div class="facts">
        <span>Giá trên web <b>${money(item.price)}</b></span>
        ${item.requested_variant && html`<span>Bạn yêu cầu <b>${item.requested_variant}</b></span>`}
        ${item.source && html`<a href=${item.source} target="_blank" rel="noreferrer noopener">Trang bạn đã gửi</a>`}
      </div>
      ${act.problem && html`<${Problem}>${friendly(act.problem)}<//>`}

      ${!item.published && html`
        <div class="note">Bên mình chỉ báo giá sau khi có người thật xem trang và cân hộp, vì cước tính theo
          cân thật. Trang này tự cập nhật, không cần tải lại.</div>`}

      ${item.published && !saved.quote && !saved.order && html`
        <${React.Fragment}>
          <${Select} ...${{ label: "Chọn size", value: variantId, onChange: e => setVariantId(e.target.value),
            options: item.variants.map(v => ({ value: v.id, label: v.label || "một phiên bản" })),
            hint: item.requested_variant
              ? `Đã chọn sẵn theo yêu cầu của bạn (${item.requested_variant}). Nếu shop ghi khác thì đây là các size nhân viên tìm thấy.`
              : "Đây là các size nhân viên tìm thấy trên trang shop." }} />
          <div class="actions">
            <button class="primary" aria-busy=${act.busy} disabled=${act.busy || !variantId} onClick=${ask}>Xin báo giá</button>
          </div>
        <//>`}

      ${saved.quote && quote && !saved.order && html`
        <${React.Fragment}>
          <${QuoteBreakdown} quote=${quote} />
          ${quote.status === "issued" && html`
            <div class="actions">
              <button class="primary" aria-busy=${act.busy} disabled=${act.busy} onClick=${agree}>Đồng ý và đặt hàng</button>
              <span class="dim">Đồng ý rồi mới tạo đơn. Báo giá hết hạn sau 48 giờ.</span>
            </div>`}
          ${quote.status === "expired" && html`<${Problem}>Báo giá đã hết hạn. Xin lại một cái mới.<//>`}
        <//>`}
      ${saved.order && html`<div class="note ok">Đã đặt hàng. Đơn nằm ở phần Đơn của tôi bên dưới.</div>`}
    <//>`;

  return lifted
    ? html`<div id=${anchor}><${Sheet} label=${item.name}>${body}<//></div>`
    : html`<div class="row" id=${anchor}>${body}</div>`;
}

/* ── the price, explained line by line ────────────────────────────────────────
 * Every figure here says where it came from. A total nobody can take apart is
 * a number a customer has to trust blindly, and that is not a price, it is a
 * demand.
 */
// A line that is always zero is noise: the reader has to work out that it does
// not apply, every time. So it is left out until it is not zero.
function nonZero(m) {
  return m && Number(m.amount) > 0;
}

function QuoteBreakdown({ quote }) {
  const st = say(QUOTE, quote.status);
  const cls = GOODS[quote.class] || { words: quote.class, why: "" };
  const l = quote.lines || {};
  const h = quote.home || {};
  const row = (label, why, value, kind) => html`
    <tr key=${label} class=${kind || undefined}>
      <th>${label}<div class="hint">${why}</div></th>
      <td class="amount">${value}</td>
    </tr>`;

  return html`
    <${React.Fragment}>
      <div class="actions">
        <span class="pill ${st.tone}">Báo giá ${st.words}</span>
        ${quote.estimated && html`<span class="pill warn">tạm tính, chưa cân thật</span>`}
      </div>
      ${st.why && html`<div class="note why">${st.why}</div>`}

      <div class="wrap-x">
        <table class="t">
          <tbody>
            ${row("Tiền hàng trên web", "Đúng giá bạn thấy trên trang shop.", money(l.item))}
            ${row("Thuế bán hàng bên Mỹ", "Shop Mỹ tính thêm thuế bán hàng khi thanh toán, mình trả hộ.", money(l.sales_tax))}
            ${row(`Cước bay, tính trên ${quote.chargeable_g} g`,
              `Không tính theo cân thật mà theo số lớn hơn giữa cân thật và cân quy đổi từ kích thước hộp, rồi làm tròn lên bước 500 g. Hộp to mà nhẹ vẫn chiếm chỗ trên máy bay. Nhóm giá: ${cls.words}. ${cls.why}`,
              money(l.freight))}
            ${nonZero(l.surcharge) && row("Phụ phí", "Phụ phí của bên vận chuyển cho nhóm hàng này.", money(l.surcharge))}
            ${nonZero(l.duty) && row("Thuế nhập khẩu", "Tính theo nhóm hàng và giá trị đơn.", money(l.duty))}
            ${row("Cộng lại bên Mỹ", "Ba dòng trên gộp lại, vẫn tính bằng tiền của shop.", money(l.subtotal), "sum")}
            ${row("Quy ra tiền Việt", `Theo tỷ giá ${Number(quote.fx || 0).toLocaleString("vi-VN")} ₫, đã khoá cứng vào báo giá này. Tỷ giá mai đổi cũng không đổi số của bạn.`, money(h.subtotal))}
            ${row("Phí dịch vụ mua hộ", "Phần của bên mình cho việc mua, kiểm, gom hàng và gửi về.", money(h.service_fee))}
            ${row("Tổng bạn trả", "Đã gồm mọi thứ ở trên. Không có phí nào phát sinh sau.", money(h.total), "sum")}
            ${row("Cọc trước", "Một nửa tổng. Bên mình ứng đủ tiền hàng cho shop nên cần cọc trước khi đi mua.", money(h.deposit), "due")}
          </tbody>
        </table>
      </div>
    <//>`;
}

/* ── my orders: one journey strip each ───────────────────────────────────── */
function MyOrders({ orders }) {
  if (!orders) return html`<${Skeleton} />`;
  if (!orders.length) {
    return html`<${Empty} icon="plane">Chưa có đơn nào. Đơn xuất hiện sau khi bạn đồng ý một báo giá.<//>`;
  }
  return html`
    <div class="list">
      ${orders.map((o, i) => html`
        <div class="row" key=${o.order_id} id=${"don-" + (i + 1)}>
          <div class="row-head">
            <span class="name">${o.product_name || "Sản phẩm chưa có tên"}</span>
            ${o.placed_by_id && html`<span class="hint">nhân viên đặt hộ bạn</span>`}
          </div>
          <${Journey} order=${o} viewer="customer" />
        </div>`)}
    </div>`;
}

/* ── what is waiting for this customer, in one place at the top ──────────────
 * Built only from what the screen already knows: a product staff finished, a
 * quote not yet agreed, a deposit or a balance the order says is unpaid.
 * Anchors are positions ("hang-2"), not ids, so no uuid reaches the page.
 */
function waitingFor(mine, quotes, orders) {
  const out = [];
  const ordered = new Set((orders || []).map(o => o.product_id));
  (mine || []).forEach((item, i) => {
    const saved = quotes[item.product_id] || {};
    if (!item.published || saved.order || ordered.has(item.product_id)) return;
    out.push({ key: "p" + i, target: "hang-" + (i + 1), name: item.name,
      text: saved.quote ? "báo giá đã có, chờ bạn đồng ý" : "nhân viên xử lý xong, xin báo giá được rồi" });
  });
  (orders || []).forEach((o, i) => {
    const target = "don-" + (i + 1);
    if (o.status === "awaiting_deposit") {
      out.push({ key: "d" + i, target, name: o.product_name, text: `chuyển cọc ${money(o.deposit)} cho nhân viên` });
    } else if (o.status === "in_transit" && !o.balance_paid) {
      out.push({ key: "b" + i, target, name: o.product_name, text: `chuyển phần còn lại ${money(o.balance)} cho nhân viên` });
    }
  });
  return out;
}

function jumpTo(id) {
  const el = document.getElementById(id);
  if (!el) return;
  el.scrollIntoView({ block: "center", behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth" });
  el.classList.remove("flash"); void el.offsetWidth; el.classList.add("flash");
  const focusable = el.querySelector("button:not(:disabled), select, input");
  if (focusable) focusable.focus({ preventScroll: true });
}

/* ── the screen ──────────────────────────────────────────────────────────── */
function App() {
  const [toasts, pushToast, dismiss] = useToasts();
  const [shops, setShops] = useState([]);
  const [categories, setCategories] = useState([]);
  const [quotes, setQuotes] = useState(loadQuotes);
  const [token, setToken] = useState(getTokens().customer);

  useEffect(() => {
    Promise.all([
      call("GET", "/merchants", { as: "customer" }),
      call("GET", "/categories", { as: "customer" }),
    ]).then(([m, c]) => { setShops(m); setCategories(c); }).catch(() => {});
  }, []);

  const loadMine = useCallback(() => call("GET", "/me/products", { as: "customer" }), []);
  const loadOrders = useCallback(() => call("GET", "/me/orders", { as: "customer" }), []);

  const [mine, mineErr, reloadMine] = usePoll(loadMine, POLL_MS, (next, prev) => {
    if (!prev) return;
    for (const item of next) {
      const before = prev.find(p => p.product_id === item.product_id);
      if (!before) continue;
      if (item.published && !before.published) {
        pushToast("Nhân viên xử lý xong", `${item.name}: xin báo giá được rồi.`);
      } else if (item.next_step !== before.next_step) {
        pushToast("Có tiến triển", `${item.name}: ${NEXT_STEP[item.next_step] || "đang xử lý"}.`);
      }
    }
  });
  const [orders, , reloadOrders] = usePoll(loadOrders, POLL_MS, (next, prev) => {
    if (!prev) return;
    for (const o of next) {
      const before = prev.find(p => p.order_id === o.order_id);
      if (before && before.status !== o.status) {
        pushToast("Đơn của bạn đổi trạng thái", `${o.product_name || "Đơn"}: ${say(ORDER, o.status).words}.`);
      }
    }
  });

  const reload = useCallback(async () => {
    await Promise.all([reloadMine(), reloadOrders()]);
  }, [reloadMine, reloadOrders]);
  // the lists are a relay tick behind a write: ask again a few quick times
  const settle = useCallback(async () => {
    await reload();
    for (const ms of [250, 700, 1500]) setTimeout(() => { reload(); }, ms);
  }, [reload]);

  const waiting = waitingFor(mine, quotes, orders);

  return html`
    <${React.Fragment}>
      <${Top} role="Bạn đang xem với vai khách" waiting=${waiting.length}>
        <a href="./staff.html">Màn hình nhân viên</a>
      <//>
      <div class="wrap">
        <${PageHead} title="Hàng bạn nhờ mua"
          lead="Dán link món ở shop Mỹ, bên mình kiểm, báo giá từng dòng, mua hộ và gửi về tận nhà."
          stats=${[
            { label: "việc chờ bạn", value: waiting.length, hot: waiting.length > 0 },
            { label: "món đang xử lý", value: mine ? mine.filter(m => !m.published).length : 0 },
            { label: "đơn đang chạy", value: orders ? orders.filter(o => !["delivered", "cancelled", "purchase_failed"].includes(o.status)).length : 0 },
          ]} />
        ${mineErr && html`<${Problem}>${friendly(mineErr)}<//>`}

        <div class="cust">
          <aside class="aside enter">
            <${Section} title="Gửi link sản phẩm bạn muốn mua" icon="link">
              <${PasteForm} shops=${shops} categories=${categories} onSent=${settle} />
            <//>
          </aside>

          <div class="main enter">
            ${waiting.length > 0 && html`
              <${Section} title="Việc đang chờ bạn" icon="alert" count=${waiting.length} hot first>
                <ul class="waiting">
                  ${waiting.map(w => html`
                    <li key=${w.key}>
                      <span><b>${w.name || "Đơn"}</b>: ${w.text}</span>
                      <button class="link" onClick=${() => jumpTo(w.target)}>Xem</button>
                    </li>`)}
                </ul>
              <//>`}

            <${Section} title="Hàng của tôi" icon="bag" count=${mine ? mine.length : ""}>
              ${!mine && html`<${Skeleton} />`}
              ${mine && !mine.length && html`<${Empty} icon="link">Chưa gửi món nào. Dán một link ở ô bên cạnh.<//>`}
              ${mine && mine.length > 0 && html`
                <div class="list">
                  ${mine.map((item, i) => html`
                    <${MyProduct} key=${item.product_id} item=${item} anchor=${"hang-" + (i + 1)} quotes=${quotes}
                      setQuotes=${setQuotes} onChanged=${settle}
                      ordered=${!!orders && orders.some(o => o.product_id === item.product_id)} />`)}
                </div>`}
            <//>

            <${Section} title="Đơn của tôi" icon="plane" count=${orders ? orders.length : ""}>
              <${MyOrders} orders=${orders} />
              <div class="note">Cọc và tiền còn lại chuyển cho nhân viên, họ xác nhận trong hệ thống. Chưa có cổng
                thanh toán tự động.</div>
            <//>
          </div>
        </div>

        <${Keys} token=${token} onSave=${t => { setTokens({ customer: t }); setToken(t); reload(); }}>
          Chìa khoá là cách hệ thống biết đây là bạn, thay cho đăng nhập bằng mật khẩu. Mọi thứ trang này
          hỏi đều gắn kèm nó, nên "hàng của tôi" và "đơn của tôi" không bao giờ trả về của người khác. Bản
          chạy thử để sẵn <b>dev-customer</b>; bản thật thì nhân viên phát cho bạn một chìa riêng.
        <//>
      </div>
      <${Toasts} items=${toasts} onDismiss=${dismiss} />
    <//>`;
}

window.ReactDOM.createRoot(document.getElementById("root")).render(html`<${App} />`);
