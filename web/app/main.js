// main.js — the console: a step runner plus the screens, both on the real API.
//
// One render() rebuilds the active tab from state; every button carries a
// data-act and one delegated listener dispatches it. Typed input is kept in
// `forms` so a re-render never eats what you were writing.
//
// [PHP] Không framework, cùng tinh thần với backend: net/http viết tay ở
// [PHP] server, DOM viết tay ở client — đọc là thấy hết, không có magic.

import * as A from "./api.js";
import { STEPS, CTX, ACTOR } from "./steps.js";

/* ── state ────────────────────────────────────────────────────────────── */
let tab = "runner";
let cur = 0;
const status = {};      // step id → "done" | "failed" | "running"
const results = {};     // step id → { ok, text, entry }
let trace = [];
let busy = false;
let openLog = -1;
const forms = {};       // input id → what you typed
const data = {};        // panel id → last response
const msg = {};         // panel id → { kind, text }

const el = id => document.getElementById(id);
const esc = s => String(s ?? "").replace(/[&<>"']/g, c => ({ "&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;" }[c]));
const val = (id, def = "") => (forms[id] !== undefined ? forms[id] : def);

function jsonHtml(v){
  const raw = JSON.stringify(v, null, 2);
  const re = /("(?:[^"\\]|\\.)*")(\s*:)?|(-?\d+(?:\.\d+)?)|\b(true|false|null)\b/g;
  let out = "", last = 0, m;
  while ((m = re.exec(raw)) !== null){
    out += esc(raw.slice(last, m.index));
    if (m[1] !== undefined) out += m[2] !== undefined
      ? `<span class="k">${esc(m[1])}</span>${esc(m[2])}`
      : `<span class="s">${esc(m[1])}</span>`;
    else if (m[3] !== undefined) out += `<span class="n">${esc(m[3])}</span>`;
    else out += `<span class="b">${esc(m[4])}</span>`;
    last = re.lastIndex;
  }
  return out + esc(raw.slice(last));
}

/* ── little building blocks ───────────────────────────────────────────── */
const chip = (ctx, text) => `<span class="chip" style="--ctx:${CTX[ctx].hue}">${esc(text || CTX[ctx].name)}</span>`;
const method = m => `<span class="method ${m.toLowerCase()}">${m}</span>`;
function panel(id, title, body, right = ""){
  const m = msg[id];
  return `<div class="card" data-panel="${id}">
    <div class="card-head"><h3>${esc(title)}</h3><span style="margin-left:auto">${right}</span></div>
    <div class="card-body">
      ${m ? `<div class="note ${m.kind === "err" ? "bad" : m.kind === "warn" ? "warn" : ""}">${m.text}</div>` : ""}
      ${body}
    </div></div>`;
}
function field(id, label, def = "", opts = {}){
  return `<div class="field${opts.grow ? " grow" : ""}"><label for="${id}">${esc(label)}</label>
    <input id="${id}" class="${opts.mono === false ? "" : "mono"}" value="${esc(val(id, def))}"
      placeholder="${esc(opts.ph || "")}" ${opts.type ? `type="${opts.type}"` : ""}></div>`;
}
function select(id, label, options, def){
  return `<div class="field"><label for="${id}">${esc(label)}</label><select id="${id}">
    ${options.map(o => `<option value="${esc(o)}" ${val(id, def) === o ? "selected" : ""}>${esc(o)}</option>`).join("")}
  </select></div>`;
}
const btn = (act, label, opts = {}) =>
  `<button class="btn ${opts.cls || ""}" data-act="${act}" ${opts.arg !== undefined ? `data-arg="${esc(opts.arg)}"` : ""}
    ${busy || opts.off ? "disabled" : ""}>${esc(label)}</button>`;
function table(cols, rows, cells){
  if (!rows || !rows.length) return `<div class="empty">Không có dòng nào.</div>`;
  return `<div class="wrap-x"><table class="t">
    <thead><tr>${cols.map(c => `<th class="${c.r ? "r" : ""}">${esc(c.h)}</th>`).join("")}</tr></thead>
    <tbody>${rows.map((r, i) => `<tr>${cells(r, i).map((v, j) =>
      `<td class="${cols[j] && cols[j].r ? "r" : ""}">${v}</td>`).join("")}</tr>`).join("")}</tbody>
  </table></div>`;
}
const statusPill = s => `<span class="pill ${
  ["delivered","published","confirmed","shipped","accepted","received"].includes(s) ? "ok" :
  ["cancelled","failed","expired","purchase_failed"].includes(s) ? "bad" : "warn"}">${esc(s)}</span>`;

/* ── runner ───────────────────────────────────────────────────────────── */
function report(line){
  trace = [...trace, line];
  render();
}

async function runStep(i){
  const st = STEPS[i];
  if (!st) return false;
  busy = true; cur = i; trace = []; status[st.id] = "running"; render();
  try {
    const text = await st.run(report);
    status[st.id] = "done";
    results[st.id] = { ok:true, text, entry:A.log[0] || null };
    busy = false; render();
    return true;
  } catch (e) {
    status[st.id] = "failed";
    results[st.id] = { ok:false, text:(e && e.message) || String(e), code:e && e.code, entry:A.log[0] || null };
    busy = false; render();
    return false;
  }
}

async function runFrom(i){
  for (let k = i; k < STEPS.length; k++){
    const ok = await runStep(k);
    if (!ok) return;
    if (k + 1 < STEPS.length) cur = k + 1;
  }
}

function runnerView(){
  const S = A.session();
  const done = STEPS.filter(s => status[s.id] === "done").length;
  const st = STEPS[cur];
  const res = results[st.id];

  let list = "", group = null;
  STEPS.forEach((s, i) => {
    if (s.ctx !== group){
      if (group !== null) list += "</div>";
      list += `<div style="--ctx:${CTX[s.ctx].hue}"><div class="sgroup-label"><i></i>${CTX[s.ctx].name}</div>`;
      group = s.ctx;
    }
    const mark = status[s.id] === "done" ? "✓" : status[s.id] === "failed" ? "✕" : status[s.id] === "running" ? "…" : "";
    list += `<button class="sitem ${status[s.id] || ""}" data-act="goto" data-arg="${i}"
      ${i === cur ? 'aria-current="true"' : ""} ${busy ? "disabled" : ""}>
      <span class="n">${s.wait ? "⚙" : i + 1}</span><span>${esc(s.title)}</span><span class="s">${mark}</span></button>`;
  });
  list += "</div>";

  const ids = [["run", S.run], ["merchant", S.merchant], ["product", S.product], ["variant", S.variant],
    ["quote", S.quote], ["order", S.order], ["task", S.task], ["parcel", S.parcel], ["batch", S.batch]];

  return `<div class="runner">
    <div>
      <div class="card"><div class="card-body" style="gap:9px">
        <div class="actions">
          ${btn("run", "Chạy bước này", { cls:"btn-primary" })}
          ${btn("runall", "Chạy tới hết")}
        </div>
        <div class="actions">
          ${btn("prev", "←", { off:cur === 0 })}${btn("next", "→", { off:cur >= STEPS.length - 1 })}
          ${btn("reset", "Đặt lại", { cls:"btn-danger" })}
        </div>
        <div class="progress"><i style="width:${Math.round(done / STEPS.length * 100)}%"></i></div>
        <div class="dim mono" style="font-size:11.5px">${done}/${STEPS.length} bước xong</div>
      </div></div>
      <div class="card" style="margin-top:12px"><div class="card-body tight steplist">${list}</div></div>
    </div>
    <div>
      ${panel("step", `${cur + 1}. ${st.title}`, `
        <div class="reqline">${chip(st.ctx)}<span class="pill">${ACTOR[st.actor].icon} ${ACTOR[st.actor].label}</span>
          ${st.api ? `<span class="path">${esc(st.api)}</span>` : `<span class="dim">không ai bấm — chỉ chờ</span>`}</div>
        <p class="muted" style="max-width:70ch">${esc(st.hint)}</p>
        ${trace.length ? `<div class="trace">${trace.map(t => `<span class="l">› ${esc(t)}</span>`).join("")}</div>` : ""}
        ${res ? (res.ok
          ? `<div class="note"><b>OK.</b> ${esc(res.text)}</div>`
          : `<div class="note bad"><b>${esc(res.code || "lỗi")}</b> — ${esc(res.text)}</div>`) : ""}
        ${res && res.entry ? `
          <div class="hr"></div>
          <div class="reqline">${method(res.entry.method)}<span class="path">${esc(res.entry.path)}</span>
            <span class="pill ${res.entry.status < 300 ? "ok" : "bad"}">${res.entry.status || "—"}</span>
            <span class="dim mono">${res.entry.ms} ms · token ${esc(res.entry.as)}</span></div>
          <div class="cols">
            ${res.entry.body !== undefined ? `<div><div class="eyebrow">Đã gửi</div><pre class="code">${jsonHtml(res.entry.body)}</pre></div>` : "<div></div>"}
            ${res.entry.json ? `<div><div class="eyebrow">Nhận về</div><pre class="code">${jsonHtml(res.entry.json)}</pre></div>` : "<div></div>"}
          </div>` : ""}
      `, busy ? `<span class="pill warn">đang chạy…</span>` : "")}
      ${panel("ids", "Id của lần chạy này", `<dl class="kv">${ids.map(([k, v]) =>
        `<dt>${k}</dt><dd>${v ? esc(v) : '<span class="dim">—</span>'}</dd>`).join("")}</dl>`)}
    </div>
  </div>`;
}

/* ── customer screens ─────────────────────────────────────────────────── */
function customerView(){
  const S = A.session();
  const q = data.quote;
  const site = S.site || `www.example-${S.run || "run"}.com`;
  return `
  ${panel("paste", "Dán link sản phẩm", `
    ${field("f-name", "Tên sản phẩm", "Air Trainer 90", { mono:false })}
    <div class="grid3">
      ${field("f-price", "Giá trên web", "150.00")}
      ${select("f-cur", "Tiền tệ", ["USD", "VND"], "USD")}
      ${select("f-cat", "Ngành hàng", ["footwear", "apparel", "electronics"], "footwear")}
    </div>
    ${field("f-url", "Link", `https://${site}/t/air-trainer-90/abc`, { grow:true })}
    ${field("f-merchant", "Shop (merchant_id)", S.merchant || "", { ph:"chạy bước 1 hoặc dán id" })}
    <div class="actions">${btn("paste", "Gửi cho nhân viên", { cls:"btn-primary" })}
      <span class="dim">gửi bằng token khách → provenance = customer</span></div>`,
    S.product ? `<span class="mono dim">product ${A.short(S.product)}</span>` : "")}

  ${panel("quote", "Báo giá", `
    <div class="actions">
      ${btn("quote-new", "Xin báo giá", { cls:"btn-primary" })}
      ${btn("quote-get", "Tải lại")}
      ${btn("quote-accept", "Đồng ý báo giá")}
      <span class="dim">${S.quote ? "quote " + A.short(S.quote) : "chưa có quote"}</span>
    </div>
    ${q ? `
      <div class="reqline">${statusPill(q.status)}<span class="pill">${esc(q.class)}</span>
        <span class="mono dim">${q.chargeable_g} g tính phí · hết hạn ${esc(q.expires_at)}</span></div>
      ${table([{ h:"Khoản" }, { h:"USD", r:true }, { h:"₫", r:true }], [
        ["Giá hàng", q.lines.item, null], ["Thuế bán hàng", q.lines.sales_tax, null],
        ["Cước", q.lines.freight, null], ["Phụ thu", q.lines.surcharge, null],
        ["Thuế nhập khẩu", q.lines.duty, null], ["Tạm tính", q.lines.subtotal, q.home.subtotal],
        ["Phí dịch vụ", null, q.home.service_fee], ["Khách trả", null, q.home.total],
        ["Cọc 50 %", null, q.home.deposit],
      ], r => [esc(r[0]), r[1] ? A.money(r[1]) : "", r[2] ? A.money(r[2]) : ""])}
    ` : `<div class="empty">Chưa tải báo giá. Bấm “Xin báo giá” — nếu vừa publish mà chưa relay thì API trả 404 listing_not_found.</div>`}`)}

  ${panel("me", "Đơn của tôi", `
    <div class="actions">${btn("me", "Tải danh sách", { cls:"btn-primary" })}</div>
    ${table([{ h:"Đơn" }, { h:"Sản phẩm" }, { h:"Trạng thái" }, { h:"Vận chuyển" }, { h:"Tổng", r:true }, { h:"Đã cọc" }],
      data.me, r => [`<span class="mono">${A.short(r.order_id)}</span>`, esc(r.product_name || "—"),
        statusPill(r.status), esc(r.tracking), A.money(r.total), r.deposit_paid ? "rồi" : "chưa"])}
    <p class="dim" style="font-size:12.5px">Khách không thấy <span class="mono">shop_reference</span> — đó là số nội bộ.</p>`)}`;
}

/* ── operator screens ─────────────────────────────────────────────────── */
function operatorView(){
  const S = A.session();
  const o = data.order, b = data.batch, r = data.recon;
  return `
  ${panel("product", "Sản phẩm nháp → publish", `
    ${field("f-pid", "product_id", S.product || "")}
    <div class="actions">
      ${btn("p-variant", "Thêm variant")}${btn("p-confirm", "Xác nhận listing")}
      ${btn("p-measure", "Cân 1250 g")}${btn("p-publish", "Publish", { cls:"btn-primary" })}
    </div>
    <div class="note">API <b>không có</b> <span class="mono">GET /products/{id}</span> — đọc là việc của read model.
      Trạng thái sản phẩm chỉ hiện ra qua lỗi: bấm Publish sớm sẽ nhận 409 <span class="mono">no_variants</span> rồi
      409 <span class="mono">unverified</span>, đúng theo thứ tự còn thiếu.</div>`)}

  ${panel("tasks", "Việc cần đi mua", `
    <div class="actions">${btn("tasks", "Tải danh sách", { cls:"btn-primary" })}</div>
    ${table([{ h:"Task" }, { h:"Mua gì" }, { h:"Đơn" }, { h:"Tiền tệ" }, { h:"Trạng thái" }, { h:"" }], data.tasks, t => [
      `<span class="mono">${A.short(t.id)}</span>`,
      `<div>${esc(t.product_name || "—")}${t.variant_label ? ` · <b>${esc(t.variant_label)}</b>` : ""}</div>
       <div class="note dim">${t.variant_ref ? `<span class="mono">${esc(t.variant_ref)}</span>` : ""}
         ${t.source ? `<a class="mono" href="${esc(t.source)}" target="_blank" rel="noopener noreferrer">link shop</a>` : ""}</div>`,
      `<span class="mono">${A.short(t.order_id)}</span>`,
      `<span class="mono">${esc(t.currency)}</span>`, statusPill(t.status),
      t.status === "open"
        ? btn("task-pick", "Chọn", { arg:t.id })
        : `<span class="mono dim">${esc(t.reference || t.reason || "")}</span>`])}
    <div class="note">Cột <b>Mua gì</b> là bốn chữ procurement tự giữ, chép vào task lúc mở: tên sản phẩm,
      size/màu, mã của shop và link gốc. Không phải join sang catalog — hai context không được đọc bảng của nhau,
      nên catalog phải <i>kể</i> qua <span class="mono">catalog.variant_added</span>.</div>
    <div class="hr"></div>
    ${field("f-task", "task_id", S.task || "")}
    <div class="grid3">${field("f-ref", "Mã đơn ở shop", `NK-${S.run || ""}-001`)}
      ${field("f-paid", "Đã trả thật", "163.22")}${select("f-paidcur", "Tiền tệ", ["USD", "VND"], "USD")}</div>
    <div class="actions">${btn("task-confirm", "Xác nhận đã mua", { cls:"btn-primary" })}
      ${field("f-reason", "Lý do nếu không mua được", "hết size US 9", { mono:false })}
      ${btn("task-fail", "Không mua được", { cls:"btn-danger" })}</div>`)}

  ${panel("parcels", "Kho: thùng chưa bay", `
    <div class="actions">${btn("parcels", "Tải danh sách", { cls:"btn-primary" })}</div>
    ${table([{ h:"Thùng" }, { h:"Đơn" }, { h:"Mã shop" }, { h:"Cân thật", r:true }, { h:"Trạng thái" }, { h:"" }],
      data.parcels, p => [`<span class="mono">${A.short(p.id)}</span>`, `<span class="mono">${A.short(p.order_id)}</span>`,
        `<span class="mono">${esc(p.reference)}</span>`, p.actual ? p.actual.weight_g + " g" : "—", statusPill(p.status),
        p.status === "expected" ? btn("parcel-receive", "Cân", { arg:p.id }) : ""])}
    <div class="note">Cân tính phí = max(cân thật, thể tích) làm tròn lên bước 500 g. 1250 g và hộp 34×23×13 cm → <b>2500 g</b>.</div>`)}

  ${panel("batch", "Lô hàng", `
    <div class="actions">${btn("batch-open", "Mở lô us_forwarder", { cls:"btn-primary" })}
      ${field("f-batch", "batch_id", S.batch || "")}${btn("batch-get", "Tải lô")}</div>
    <div class="actions">${field("f-parcel", "parcel_id", S.parcel || "")}
      ${btn("batch-add", "Xếp thùng vào lô")}${btn("batch-close", "Dán kín lô")}</div>
    <div class="actions">${field("f-freight", "Hoá đơn carrier", "27.50")}
      ${select("f-freightcur", "Tiền tệ", ["USD", "VND"], "USD")}${btn("batch-ship", "Ship lô", { cls:"btn-primary" })}</div>
    ${b ? `<div class="reqline">${statusPill(b.status)}<span class="mono dim">${esc(b.lane)} ·
        ${(b.parcels || []).length} thùng · ${b.freight ? A.money(b.freight) : "chưa có hoá đơn"}</span></div>
      ${table([{ h:"Thùng" }, { h:"Đơn" }, { h:"Cân tính phí", r:true }, { h:"Cước chia", r:true }],
        b.allocations || [], a => [`<span class="mono">${A.short(a.parcel_id)}</span>`,
          `<span class="mono">${A.short(a.order_id)}</span>`, a.chargeable_g + " g", A.money(a.freight)])}` : ""}`)}

  ${panel("order", "Một đơn hàng", `
    <div class="actions">${field("f-order", "order_id", S.order || "")}${btn("order-get", "Tải đơn", { cls:"btn-primary" })}</div>
    ${o ? `<div class="reqline">${statusPill(o.status)}
        <span class="mono dim">tổng ${A.money(o.total)} · cọc ${A.money(o.deposit)} · còn ${A.money(o.balance)}
        ${o.balance_paid ? "· đã trả nốt" : ""}${o.forfeited ? "· MẤT CỌC" : ""}</span></div>
      ${o.refund ? `<div class="note warn">Đã huỷ — hoàn ${A.money(o.refund)}${o.forfeited ? " (mất cọc)" : ""}. Lý do: ${esc(o.cancel_reason || "")}</div>` : ""}` : ""}
    <div class="actions">
      ${field("f-amount", "Số tiền", S.deposit || "2696860")}
      ${btn("order-deposit", "Thu cọc")}${btn("order-balance", "Thu nốt")}${btn("order-deliver", "Đã giao")}
    </div>
    <div class="actions">${field("f-cancel", "Lý do huỷ", "khách đổi ý", { mono:false })}
      ${btn("order-cancel", "Huỷ đơn", { cls:"btn-danger" })}</div>
    <div class="note">Huỷ trước <span class="mono">purchased</span> hoàn đủ cọc; từ <span class="mono">purchased</span>
      trở đi hoàn 0 và khách mất cọc — ranh giới đó là điểm không thể quay đầu.</div>`)}

  ${panel("queue", "Hàng đợi theo trạng thái", `
    <div class="actions">${select("f-status", "status",
      ["awaiting_deposit", "deposited", "purchased", "in_transit", "delivered", "cancelled", "purchase_failed"], "deposited")}
      ${btn("queue", "Tải hàng đợi", { cls:"btn-primary" })}</div>
    ${table([{ h:"Đơn" }, { h:"Sản phẩm" }, { h:"Trạng thái" }, { h:"Vận chuyển" }, { h:"Mã shop" }, { h:"Tổng", r:true }],
      data.queue, s => [`<span class="mono">${A.short(s.order_id)}</span>`, esc(s.product_name || "—"),
        statusPill(s.status), esc(s.tracking), `<span class="mono">${esc(s.shop_reference || "—")}</span>`, A.money(s.total)])}`)}

  ${panel("recon", "Đối soát Quote vs Actual", `
    <div class="actions">${field("f-recon", "order_id", S.order || "")}${btn("recon", "Tải đối soát", { cls:"btn-primary" })}</div>
    ${r ? `${table([{ h:"Khoản" }, { h:"Ước tính", r:true }, { h:"Thực tế", r:true }], [
        ["Tiền hàng", r.quoted.goods, r.actual.goods], ["Cước", r.quoted.freight, r.actual.freight],
      ], row => [esc(row[0]), row[1] ? A.money(row[1]) : "—", row[2] ? A.money(row[2]) : "—"])}
      <div class="note ${r.variance && String(r.variance.amount).startsWith("-") ? "warn" : ""}">
        ${r.complete ? `Lệch <b>${A.money(r.variance)}</b> — âm là mình chịu, dương là báo giá dư.`
                     : "Chưa đủ hai nửa actual nên chưa có variance."}</div>` : ""}`)}`;
}

/* ── keys ─────────────────────────────────────────────────────────────── */
function keysView(){
  return `
  ${panel("keys", "Chìa khoá API", `
    <div class="actions">${btn("keys", "Tải danh sách", { cls:"btn-primary" })}</div>
    ${table([{ h:"Hash" }, { h:"Loại" }, { h:"Chủ" }, { h:"Nhãn" }, { h:"Còn dùng" }, { h:"" }], data.keys, k => [
      `<span class="mono">${esc(String(k.hash).slice(0, 12))}…</span>`, `<span class="pill">${esc(k.kind)}</span>`,
      `<span class="mono">${A.short(k.subject)}</span>`, esc(k.label || "—"),
      k.active ? `<span class="pill ok">có</span>` : `<span class="pill bad">đã thu hồi</span>`,
      k.active ? btn("key-revoke", "Thu hồi", { arg:k.hash, cls:"btn-danger" }) : ""])}
    <div class="hr"></div>
    <div class="grid3">${select("f-kind", "Loại", ["customer", "operator"], "customer")}
      ${field("f-label", "Nhãn", "console", { mono:false })}
      ${field("f-subject", "Chủ (để trống = tạo mới)", "")}</div>
    <div class="actions">${btn("key-new", "Phát chìa mới", { cls:"btn-primary" })}</div>
    ${data.newToken ? `<div class="note"><b>Token chỉ hiện một lần:</b>
      <span class="mono">${esc(data.newToken.token)}</span><br>
      ${btn("key-use", "Dùng làm token " + (data.newToken.kind === "operator" ? "operator" : "khách"))}</div>` : ""}
    <div class="note warn">Store chỉ giữ <b>hash</b>, nên token mất là phát lại chứ không lấy lại được.
      Không thu hồi được chìa operator cuối cùng — API mà không ai quản được thì chỉ còn cách sửa DB bằng tay.</div>`)}`;
}

/* ── log ──────────────────────────────────────────────────────────────── */
function logView(){
  return panel("log", `Call log (${A.log.length})`, `
    <div class="actions">${btn("log-clear", "Xoá log")}
      <span class="dim">mọi request UI gửi, kể cả cái thất bại</span></div>
    <div class="card-body tight" style="border:1px solid var(--line);border-radius:var(--radius-sm)">
      ${A.log.length ? A.log.map((e, i) => `
        <div class="logrow ${e.status && e.status < 300 ? "ok2" : "err"}" data-act="log-open" data-arg="${i}">
          <span class="st">${e.status || "—"}</span>
          <span class="mono">${esc(e.method)}</span>
          <span class="p">${esc(e.path)}</span>
          <span class="who">${esc(e.as)}</span>
          <span class="ms">${e.ms} ms</span>
        </div>
        ${openLog === i ? `<div class="logdetail">
          <div><div class="eyebrow">Đã gửi</div><pre class="code">${e.body !== undefined ? jsonHtml(e.body) : "(không body)"}</pre></div>
          <div><div class="eyebrow">Nhận về</div><pre class="code">${e.json ? jsonHtml(e.json) : esc(e.error || e.text || "(rỗng)")}</pre></div>
        </div>` : ""}`).join("") : `<div class="empty">Chưa gọi API lần nào.</div>`}
    </div>`);
}

/* ── actions ──────────────────────────────────────────────────────────── */
const ACTIONS = {
  goto: a => { if (!busy){ cur = +a; trace = []; } },
  prev: () => { cur = Math.max(0, cur - 1); trace = []; },
  next: () => { cur = Math.min(STEPS.length - 1, cur + 1); trace = []; },
  run: () => runStep(cur),
  runall: () => runFrom(cur),
  reset: () => {
    A.resetSession();
    for (const k of Object.keys(status)) delete status[k];
    for (const k of Object.keys(results)) delete results[k];
    for (const k of Object.keys(data)) delete data[k];
    for (const k of Object.keys(forms)) if (k.startsWith("f-")) delete forms[k];
    for (const k of Object.keys(msg)) delete msg[k];
    cur = 0; trace = [];
  },

  paste: () => run("paste", async () => {
    const r = await A.call("POST", "/products", { as:"customer", body:{
      name:val("f-name", "Air Trainer 90"), merchant_id:val("f-merchant", A.session().merchant || ""),
      category:val("f-cat", "footwear"), source_url:val("f-url", ""),
      price:val("f-price", "150.00"), currency:val("f-cur", "USD") } });
    A.setSession({ product:r.id });
    return { kind:"ok", text:`Đã tạo sản phẩm nháp <span class="mono">${r.id}</span>` };
  }),
  "quote-new": () => run("quote", async () => {
    const r = await A.call("POST", "/quotes", { as:"customer",
      body:{ product_id:val("f-pid", A.session().product || ""), lane:"us_forwarder" } });
    A.setSession({ quote:r.id });
    data.quote = await A.call("GET", `/quotes/${r.id}`, { as:"customer" });
    return { kind:"ok", text:`Quote <span class="mono">${r.id}</span>` };
  }),
  "quote-get": () => run("quote", async () => {
    data.quote = await A.call("GET", `/quotes/${A.session().quote}`, { as:"customer" });
    return null;
  }),
  "quote-accept": () => run("quote", async () => {
    await A.call("POST", `/quotes/${A.session().quote}/accept`, { as:"customer" });
    data.quote = await A.call("GET", `/quotes/${A.session().quote}`, { as:"customer" });
    return { kind:"ok", text:"Khách đã đồng ý — ordering sẽ biết sau khi worker relay." };
  }),
  me: () => run("me", async () => { data.me = await A.call("GET", "/me/orders", { as:"customer" }); return null; }),

  "p-variant": () => run("product", async () => {
    const r = await A.call("POST", `/products/${val("f-pid", A.session().product)}/variants`,
      { as:"operator", body:{ size:"US 9", color:"black" } });
    A.setSession({ variant:r.id });
    return { kind:"ok", text:`variant <span class="mono">${r.id}</span>` };
  }),
  "p-confirm": () => run("product", async () => {
    await A.call("POST", `/products/${val("f-pid", A.session().product)}/confirm-listing`, { as:"operator" });
    return { kind:"ok", text:"listing verified" };
  }),
  "p-measure": () => run("product", async () => {
    await A.call("POST", `/products/${val("f-pid", A.session().product)}/measure`, { as:"operator",
      body:{ weight_g:1250, length_mm:340, width_mm:230, height_mm:130 } });
    return { kind:"ok", text:"đã ghi số cân" };
  }),
  "p-publish": () => run("product", async () => {
    await A.call("POST", `/products/${val("f-pid", A.session().product)}/publish`, { as:"operator" });
    return { kind:"ok", text:"published" };
  }),

  tasks: () => run("tasks", async () => { data.tasks = await A.call("GET", "/purchase-tasks", { as:"operator" }); return null; }),
  "task-pick": a => { forms["f-task"] = a; },
  "task-confirm": () => run("tasks", async () => {
    await A.call("POST", `/purchase-tasks/${val("f-task", A.session().task)}/confirm`, { as:"operator",
      body:{ reference:val("f-ref", ""), paid:val("f-paid", "163.22"), currency:val("f-paidcur", "USD") } });
    data.tasks = await A.call("GET", "/purchase-tasks", { as:"operator" });
    return { kind:"ok", text:"Đã xác nhận mua — ordering sẽ sang purchased sau khi relay." };
  }),
  "task-fail": () => run("tasks", async () => {
    await A.call("POST", `/purchase-tasks/${val("f-task", A.session().task)}/fail`, { as:"operator",
      body:{ reason:val("f-reason", "hết size") } });
    data.tasks = await A.call("GET", "/purchase-tasks", { as:"operator" });
    return { kind:"warn", text:"Đã báo không mua được — đơn sẽ về purchase_failed, huỷ lúc đó hoàn ĐỦ cọc." };
  }),

  parcels: () => run("parcels", async () => { data.parcels = await A.call("GET", "/parcels", { as:"operator" }); return null; }),
  "parcel-receive": a => run("parcels", async () => {
    await A.call("POST", `/parcels/${a}/receive`, { as:"operator",
      body:{ weight_g:1250, length_mm:340, width_mm:230, height_mm:130 } });
    A.setSession({ parcel:a });
    data.parcels = await A.call("GET", "/parcels", { as:"operator" });
    return { kind:"ok", text:"đã cân" };
  }),

  "batch-open": () => run("batch", async () => {
    const r = await A.call("POST", "/batches", { as:"operator", body:{ lane:"us_forwarder" } });
    A.setSession({ batch:r.id }); forms["f-batch"] = r.id;
    data.batch = await A.call("GET", `/batches/${r.id}`, { as:"operator" });
    return { kind:"ok", text:`lô <span class="mono">${r.id}</span>` };
  }),
  "batch-get": () => run("batch", async () => {
    data.batch = await A.call("GET", `/batches/${val("f-batch", A.session().batch)}`, { as:"operator" }); return null;
  }),
  "batch-add": () => run("batch", async () => {
    const b = val("f-batch", A.session().batch);
    await A.call("POST", `/batches/${b}/parcels`, { as:"operator", body:{ parcel_id:val("f-parcel", A.session().parcel) } });
    data.batch = await A.call("GET", `/batches/${b}`, { as:"operator" });
    return { kind:"ok", text:"đã xếp thùng" };
  }),
  "batch-close": () => run("batch", async () => {
    const b = val("f-batch", A.session().batch);
    await A.call("POST", `/batches/${b}/close`, { as:"operator" });
    data.batch = await A.call("GET", `/batches/${b}`, { as:"operator" });
    return { kind:"ok", text:"lô đã dán kín" };
  }),
  "batch-ship": () => run("batch", async () => {
    const b = val("f-batch", A.session().batch);
    const allocs = await A.call("POST", `/batches/${b}/ship`, { as:"operator",
      body:{ freight:val("f-freight", "27.50"), currency:val("f-freightcur", "USD") } });
    data.batch = await A.call("GET", `/batches/${b}`, { as:"operator" });
    const a = (allocs || [])[0];
    return { kind:"ok", text:a ? `chia cước: ${A.money(a.freight)} cho đơn ${A.short(a.order_id)} (${a.chargeable_g} g)` : "đã ship" };
  }),

  "order-get": () => run("order", async () => {
    data.order = await A.call("GET", `/orders/${val("f-order", A.session().order)}`, { as:"operator" }); return null;
  }),
  "order-deposit": () => run("order", async () => {
    const id = val("f-order", A.session().order);
    await A.call("POST", `/orders/${id}/deposit`, { as:"operator", body:{ amount:val("f-amount", ""), currency:"VND" } });
    data.order = await A.call("GET", `/orders/${id}`, { as:"operator" });
    return { kind:"ok", text:"đã thu cọc" };
  }),
  "order-balance": () => run("order", async () => {
    const id = val("f-order", A.session().order);
    await A.call("POST", `/orders/${id}/balance`, { as:"operator", body:{ amount:val("f-amount", ""), currency:"VND" } });
    data.order = await A.call("GET", `/orders/${id}`, { as:"operator" });
    return { kind:"ok", text:"đã thu nốt" };
  }),
  "order-deliver": () => run("order", async () => {
    const id = val("f-order", A.session().order);
    await A.call("POST", `/orders/${id}/deliver`, { as:"operator" });
    data.order = await A.call("GET", `/orders/${id}`, { as:"operator" });
    return { kind:"ok", text:"đã giao" };
  }),
  "order-cancel": () => run("order", async () => {
    const id = val("f-order", A.session().order);
    await A.call("POST", `/orders/${id}/cancel`, { as:"operator", body:{ reason:val("f-cancel", "khách đổi ý") } });
    data.order = await A.call("GET", `/orders/${id}`, { as:"operator" });
    return { kind:"warn", text:`đã huỷ — hoàn ${data.order.refund ? A.money(data.order.refund) : "0"}${data.order.forfeited ? " (mất cọc)" : ""}` };
  }),
  queue: () => run("queue", async () => {
    data.queue = await A.call("GET", `/orders?status=${encodeURIComponent(val("f-status", "deposited"))}`, { as:"operator" });
    return null;
  }),
  recon: () => run("recon", async () => {
    data.recon = await A.call("GET", `/reconciliations/${val("f-recon", A.session().order)}`, { as:"operator" }); return null;
  }),

  keys: () => run("keys", async () => { data.keys = await A.call("GET", "/tokens", { as:"operator" }); return null; }),
  "key-new": () => run("keys", async () => {
    const body = { kind:val("f-kind", "customer"), label:val("f-label", "console") };
    const subj = val("f-subject", ""); if (subj) body.subject = subj;
    data.newToken = await A.call("POST", "/tokens", { as:"operator", body });
    data.keys = await A.call("GET", "/tokens", { as:"operator" });
    return { kind:"ok", text:"đã phát chìa — copy ngay, store chỉ giữ hash" };
  }),
  "key-use": () => {
    const t = data.newToken; if (!t) return;
    A.setCfg(t.kind === "operator" ? { op:t.token } : { cust:t.token });
    msg.keys = { kind:"ok", text:`Đã đặt làm token ${t.kind}.` };
  },
  "key-revoke": a => run("keys", async () => {
    await A.call("DELETE", `/tokens/${a}`, { as:"operator" });
    data.keys = await A.call("GET", "/tokens", { as:"operator" });
    return { kind:"ok", text:"đã thu hồi" };
  }),

  "log-open": a => { openLog = openLog === +a ? -1 : +a; },
  "log-clear": () => A.clearLog(),
  "cfg-save": () => {
    A.setCfg({ base:val("c-base", A.cfg().base).trim(), op:val("c-op", A.cfg().op).trim(), cust:val("c-cust", A.cfg().cust).trim() });
    checkHealth();
  },
};

/** run wraps a panel action: one place turns an ApiError into a readable note. */
async function run(panelId, fn){
  busy = true; msg[panelId] = null; render();
  try {
    const out = await fn();
    if (out) msg[panelId] = out;
  } catch (e) {
    msg[panelId] = { kind:"err", text:`<b>${esc(e.code || "lỗi")}</b> — ${esc((e && e.message) || String(e))}` };
  }
  busy = false; render();
}

/* ── health + shell ───────────────────────────────────────────────────── */
let healthState = { state:"", text:"chưa kiểm" };
async function checkHealth(){
  healthState = { state:"", text:"đang kiểm…" }; render();
  healthState = await A.health();
  render();
}

function render(){
  const c = A.cfg();
  el("tokenbar").innerHTML = `
    ${field("c-base", "API base (rỗng = cùng origin)", c.base, { ph:"http://localhost:8080" })}
    ${field("c-op", "Token operator", c.op, { grow:true })}
    ${field("c-cust", "Token khách", c.cust, { grow:true })}
    <div class="actions">${btn("cfg-save", "Lưu + kiểm", { cls:"btn-primary" })}</div>`;
  el("health").className = "health " + (healthState.state || "");
  el("health").innerHTML = `<span class="dot"></span>${esc(healthState.text)}`;
  document.querySelectorAll("#tabs button").forEach(b => b.setAttribute("aria-selected", String(b.dataset.tab === tab)));
  el("view").innerHTML =
    tab === "runner" ? runnerView() :
    tab === "customer" ? customerView() :
    tab === "operator" ? operatorView() :
    tab === "keys" ? keysView() : logView();
  if (tab === "runner"){
    const box = document.querySelector(".steplist");
    const active = box && box.querySelector('[aria-current="true"]');
    if (box && active && box.scrollHeight > box.clientHeight){
      box.scrollTop = Math.max(0, active.offsetTop - box.clientHeight / 2);
    }
  }
}

document.addEventListener("click", e => {
  const t = e.target.closest("[data-act],[data-tab]");
  if (!t) return;
  if (t.dataset.tab){ tab = t.dataset.tab; render(); return; }
  const fn = ACTIONS[t.dataset.act];
  if (!fn) return;
  const out = fn(t.dataset.arg);
  if (out && typeof out.then === "function") out.then(() => render());
  else render();
});
// Keep what you typed: re-rendering a panel must not eat an id you pasted.
document.addEventListener("input", e => {
  if (e.target.id) forms[e.target.id] = e.target.value;
});
document.addEventListener("change", e => {
  if (e.target.id) forms[e.target.id] = e.target.value;
});

if (location.protocol === "file:"){
  document.body.insertAdjacentHTML("afterbegin",
    `<div class="note bad" style="margin:12px">Trang này đang mở bằng <span class="mono">file://</span> nên module JS và
     fetch cùng origin đều không chạy. Mở qua <span class="mono">http://localhost:8080/ui/console.html</span>
     (chạy <span class="mono">go run ./cmd/api -web ./web</span>).</div>`);
}

// The log tab shows requests as they happen; the other tabs render on action.
A.onChange(() => { if (tab === "log") render(); });

render();
checkHealth();
