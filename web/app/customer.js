// customer.js — the screen a customer uses.
//
// Three questions, in the order a person actually has them: can you buy me
// this, how much, and where is it. Everything else the API can do is somebody
// else's screen.
//
// No uuid is ever typed here. The shop and the category come from the API as
// lists, the size comes from what staff published, and the ids the API needs
// travel in the background.

import { React, html, Top, Card, Field, Select, Steps, Toasts, useToasts, usePoll, useAction } from "./ui.js";
import { call, money, retryOn, getTokens, setTokens } from "./portage.js";
import { QUOTE, ORDER, GOODS, NEXT_STEP, TRACKING, say, friendly, categoryWords } from "./words.js";

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

  useEffect(() => { if (!shopId && canPaste.length) setShopId(canPaste[0].id); }, [canPaste, shopId]);
  useEffect(() => { if (!category && categories.length) setCategory(categories[0].code); }, [categories, category]);

  const shop = canPaste.find(s => s.id === shopId);
  const cat = categories.find(c => c.code === category);

  const send = () => act.go(async () => {
    await call("POST", "/products", {
      as: "customer",
      body: {
        name: name.trim(), merchant_id: shopId, category,
        source_url: url.trim(), price: price.trim(), currency: shop.currency,
      },
    });
    setUrl(""); setName(""); setPrice("");
    await onSent();
  });

  if (!canPaste.length) {
    return html`<div class="empty">Chưa có shop nào cho khách tự gửi link. Nhân viên phải thêm shop trước.</div>`;
  }
  return html`
    <${React.Fragment}>
      ${act.problem && html`<div class="note bad">${friendly(act.problem)}</div>`}
      <div class="grid2">
        <${Select} ...${{ label: "Shop", value: shopId, onChange: e => setShopId(e.target.value),
          options: canPaste.map(s => ({ value: s.id, label: s.name })),
          hint: shop ? `Shop này bán bằng ${shop.currency}` : "" }} />
        <${Select} ...${{ label: "Ngành hàng", value: category, onChange: e => setCategory(e.target.value),
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
      <div class="actions">
        <button class="primary" disabled=${act.busy || !url.trim() || !name.trim() || !price.trim()}
          onClick=${send}>Gửi cho nhân viên</button>
        <span class="dim">Gửi xong nhân viên sẽ kiểm rồi mới báo giá được.</span>
      </div>
    <//>`;
}

/* ── one thing I asked for ───────────────────────────────────────────────── */
function MyProduct({ item, quotes, setQuotes, onChanged }) {
  const act = useAction();
  const saved = quotes[item.product_id] || {};
  const [variantId, setVariantId] = useState(saved.variant || "");
  const [quote, setQuote] = useState(null);

  useEffect(() => {
    if (!variantId && item.variants.length) setVariantId(item.variants[0].id);
  }, [item.variants, variantId]);

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

  return html`
    <div class="item">
      <div class="item-head">
        <span class="name">${item.name}</span>
        <span class="pill flat">${categoryWords(item.category)}</span>
        <span class="spacer"></span>
        ${item.published
          ? html`<span class="pill ok">sẵn sàng báo giá</span>`
          : html`<span class="pill warn">${waiting || "đang xử lý"}</span>`}
      </div>
      <div class="item-body">
        <div class="meta" style=${{ marginBottom: "10px" }}>
          <span>Giá trên web: <b>${money(item.price)}</b></span>
          ${item.source && html`<a href=${item.source} target="_blank" rel="noreferrer noopener">trang bạn đã gửi</a>`}
        </div>
        ${act.problem && html`<div class="note bad">${friendly(act.problem)}</div>`}

        ${!item.published && html`
          <div class="note">Bên mình chỉ báo giá sau khi có người thật xem trang và cân hộp, vì cước tính theo
            cân thật. Trang này tự cập nhật, không cần tải lại.</div>`}

        ${item.published && !saved.quote && html`
          <${React.Fragment}>
            <div class="grid2">
              <${Select} ...${{ label: "Chọn size", value: variantId, onChange: e => setVariantId(e.target.value),
                options: item.variants.map(v => ({ value: v.id, label: v.label || "một phiên bản" })) }} />
            </div>
            <div class="actions">
              <button class="primary" disabled=${act.busy || !variantId} onClick=${ask}>Xin báo giá</button>
            </div>
          <//>`}

        ${saved.quote && quote && html`
          <${React.Fragment}>
            <${QuoteBreakdown} quote=${quote} />
            ${!saved.order && quote.status === "issued" && html`
              <div class="actions">
                <button class="primary" disabled=${act.busy} onClick=${agree}>Đồng ý và đặt hàng</button>
                <span class="dim">Đồng ý rồi mới tạo đơn. Báo giá hết hạn sau 48 giờ.</span>
              </div>`}
            ${saved.order && html`<div class="note ok">Đã đặt đơn. Xem phần Đơn của tôi bên dưới.</div>`}
            ${quote.status === "expired" && html`<div class="note bad">Báo giá đã hết hạn. Xin lại một cái mới.</div>`}
          <//>`}
      </div>
    </div>`;
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
  const row = (label, why, value, strong) => html`
    <tr key=${label}>
      <th style=${{ fontWeight: 500 }}>${label}<div class="hint" style=${{ fontWeight: 400 }}>${why}</div></th>
      <td style=${{ whiteSpace: "nowrap", textAlign: "right" }}>${strong ? html`<b>${value}</b>` : value}</td>
    </tr>`;

  return html`
    <${React.Fragment}>
      <div class="actions" style=${{ marginBottom: "8px" }}>
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
              `Không tính theo cân thật mà theo số lớn hơn giữa cân thật và cân quy đổi từ kích thước hộp, rồi làm tròn lên bước 500 g. Hộp to mà nhẹ vẫn chiếm chỗ trên máy bay. Nhóm giá: ${cls.words} — ${cls.why}`,
              money(l.freight))}
            ${nonZero(l.surcharge) && row("Phụ phí", "Phụ phí của bên vận chuyển cho nhóm hàng này.", money(l.surcharge))}
            ${nonZero(l.duty) && row("Thuế nhập khẩu", "Tính theo nhóm hàng và giá trị đơn.", money(l.duty))}
            ${row("Cộng lại bên Mỹ", "Ba dòng trên gộp lại, vẫn tính bằng tiền của shop.", money(l.subtotal), true)}
            ${row("Quy ra tiền Việt", `Theo tỷ giá ${Number(quote.fx || 0).toLocaleString("vi-VN")} ₫, đã khoá cứng vào báo giá này. Tỷ giá mai đổi cũng không đổi số của bạn.`, money(h.subtotal))}
            ${row("Phí dịch vụ mua hộ", "Phần của bên mình cho việc mua, kiểm, gom hàng và gửi về.", money(h.service_fee))}
            ${row("Tổng bạn trả", "Đã gồm mọi thứ ở trên. Không có phí nào phát sinh sau.", money(h.total), true)}
            ${row("Cọc trước", "Một nửa tổng. Bên mình ứng đủ tiền hàng cho shop nên cần cọc trước khi đi mua.", money(h.deposit), true)}
          </tbody>
        </table>
      </div>
    <//>`;
}

/* ── my orders ───────────────────────────────────────────────────────────── */
function MyOrders({ orders }) {
  if (!orders || !orders.length) {
    return html`<div class="empty">Chưa có đơn nào. Đơn xuất hiện sau khi bạn đồng ý một báo giá.</div>`;
  }
  return html`
    <${React.Fragment}>
      <div class="wrap-x">
        <table class="t">
          <thead><tr>
            <th>Sản phẩm</th>
            <th>Đơn đang ở đâu</th>
            <th>Tổng phải trả</th>
            <th>Cọc</th>
            <th>Kiện hàng</th>
          </tr></thead>
          <tbody>
            ${orders.map(o => {
              const st = say(ORDER, o.status);
              const tr = say(TRACKING, o.tracking);
              return html`
                <tr key=${o.order_id}>
                  <td>${o.product_name || "—"}</td>
                  <td><span class="pill ${st.tone}">${st.words}</span>
                    ${st.why && html`<div class="hint">${st.why}</div>`}</td>
                  <td style=${{ whiteSpace: "nowrap" }}>${money(o.total)}</td>
                  <td style=${{ whiteSpace: "nowrap" }}>${o.deposit_paid
                    ? html`<span class="dim">đã nhận</span>`
                    : html`${money(o.deposit)}<div class="hint">cần chuyển</div>`}</td>
                  <td><span class="pill ${tr.tone}">${tr.words}</span></td>
                </tr>`;
            })}
          </tbody>
        </table>
      </div>
      <div class="note">Hai cột giữa trả lời hai câu khác nhau: <b>đơn đang ở đâu</b> là chuyện tiền và
        cam kết, còn <b>kiện hàng</b> là chuyện cái hộp đang nằm đâu. Chúng đổi trạng thái vào những lúc
        khác nhau nên tách ra hai cột.</div>
    <//>`;
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
        pushToast("Nhân viên xử lý xong", `${item.name} — xin báo giá được rồi`);
      } else if (item.next_step !== before.next_step) {
        pushToast("Có tiến triển", `${item.name} — đang ở bước ${item.next_step || "xong"}`);
      }
    }
  });
  const [orders, , reloadOrders] = usePoll(loadOrders, POLL_MS, (next, prev) => {
    if (!prev) return;
    for (const o of next) {
      const before = prev.find(p => p.order_id === o.order_id);
      if (before && before.status !== o.status) {
        pushToast("Đơn của bạn đổi trạng thái", `${o.product_name || "đơn"} — ${say(ORDER, o.status).words}`);
      }
    }
  });

  const reload = useCallback(async () => {
    await Promise.all([reloadMine(), reloadOrders()]);
  }, [reloadMine, reloadOrders]);

  const readyCount = mine ? mine.filter(m => m.published && !(quotes[m.product_id] || {}).order).length : 0;

  return html`
    <${React.Fragment}>
      <${Top} title="Portage" role="Bạn đang xem với vai khách" waiting=${readyCount}>
        <a href="./staff.html">Màn hình nhân viên →</a>
      <//>
      <div class="wrap stack">
        ${mineErr && html`<div class="note bad">${friendly(mineErr)}</div>`}

        <${Card} title="Gửi link sản phẩm bạn muốn mua">
          <${PasteForm} shops=${shops} categories=${categories} onSent=${reload} />
        <//>

        <${Card} title="Hàng của tôi" right=${html`<span class="dim">${mine ? mine.length : 0} món</span>`}>
          ${!mine && html`<div class="empty">Đang tải…</div>`}
          ${mine && !mine.length && html`<div class="empty">Chưa gửi món nào. Dán một link ở trên.</div>`}
          ${mine && mine.map(item => html`
            <${MyProduct} key=${item.product_id} item=${item} quotes=${quotes}
              setQuotes=${setQuotes} onChanged=${reload} />`)}
        <//>

        <${Card} title="Đơn của tôi">
          <${MyOrders} orders=${orders} />
          <div class="note">Cọc và tiền còn lại chuyển cho nhân viên, họ xác nhận trong hệ thống. Chưa có cổng
            thanh toán tự động.</div>
        <//>

        <${Card} title="Chìa khoá của bạn">
          <div class="note">Chìa khoá là cách hệ thống biết đây là bạn, thay cho đăng nhập bằng mật khẩu.
            Mọi thứ trang này hỏi đều gắn kèm nó, nên "hàng của tôi" và "đơn của tôi" không bao giờ trả về
            của người khác. Bản chạy thử để sẵn <span class="mono">dev-customer</span>; bản thật thì nhân
            viên phát cho bạn một chìa riêng.</div>
          <div class="grid2">
            <${Field} ...${{ label: "Chìa khoá", value: token, onChange: e => setToken(e.target.value) }} />
          </div>
          <div class="actions">
            <button onClick=${() => { setTokens({ customer: token }); reload(); }}>Lưu và tải lại</button>
            <span class="dim">Đổi chìa là đổi người, nên danh sách bên trên sẽ tải lại theo.</span>
          </div>
        <//>
      </div>
      <${Toasts} items=${toasts} onDismiss=${dismiss} />
    <//>`;
}

window.ReactDOM.createRoot(document.getElementById("root")).render(html`<${App} />`);
