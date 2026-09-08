// steps.js — the 26 steps of one order, as REAL API calls.
//
// Same sequence as web/flow.html, but nothing is simulated: each run() calls
// the API, captures the ids it hands back, and the waits are genuine waits on
// the outbox relay. That is why some steps poll a read endpoint and others
// retry a write on exactly one error code — the two honest ways to find out
// whether the worker has caught up.
//
// [PHP] Gần nhất bên Symfony là một Behat feature file: mỗi step một hành
// [PHP] động HTTP thật, chạy tuần tự, giữ id từ bước trước.

import { call, poll, retryOn, session, setSession, runTag, short, money, sleep } from "./api.js";

export const CTX = {
  catalog:     { name: "Catalog",     hue: "var(--c-catalog)" },
  pricing:     { name: "Pricing",     hue: "var(--c-pricing)" },
  ordering:    { name: "Ordering",    hue: "var(--c-ordering)" },
  procurement: { name: "Procurement", hue: "var(--c-procurement)" },
  logistics:   { name: "Logistics",   hue: "var(--c-logistics)" },
  reporting:   { name: "Reporting",   hue: "var(--c-reporting)" },
  worker:      { name: "Worker",      hue: "var(--c-worker)" },
};
export const ACTOR = {
  customer: { icon: "◍", label: "Khách" },
  operator: { icon: "◆", label: "Operator" },
  worker:   { icon: "⚙", label: "Worker" },
};

/* Values that must be unique per run, or the second run fails on a UNIQUE
   index (merchants.site) and on the duplicate-product flag (source_url). */
const site = () => `www.example-${runTag()}.com`;
const sourceURL = () => `https://${site()}/t/air-trainer-90/abc`;
const reference = () => `NK-${runTag()}-001`;

const need = (...keys) => {
  const S = session();
  for (const k of keys) if (!S[k]) throw new Error(`thiếu ${k} — chạy lại từ bước trước (hoặc bấm "Đặt lại")`);
  return S;
};

export const STEPS = [
{ id:"merchant", ctx:"catalog", actor:"operator", api:"POST /merchants",
  title:"Đăng ký shop",
  hint:"Shop quyết định tiền tệ; biên nhận lúc đi mua buộc phải cùng loại tiền đó.",
  async run(){
    const r = await call("POST", "/merchants", { as:"operator", body:{
      name:"Example Sports", site:site(), currency:"USD",
      free_shipping:{ kind:"over", threshold:"50.00" }, sourcing:["operator","customer"] } });
    setSession({ merchant:r.id, site:site() });
    return `shop ${short(r.id)} · ${site()}`;
  } },

{ id:"product", ctx:"catalog", actor:"customer", api:"POST /products",
  title:"Khách dán link",
  hint:"Gửi bằng token KHÁCH: không có field sourced_by, provenance suy từ token nên là customer — chưa ai xác minh.",
  async run(){
    const S = need("merchant");
    const r = await call("POST", "/products", { as:"customer", body:{
      name:"Air Trainer 90", merchant_id:S.merchant, category:"footwear",
      source_url:sourceURL(), price:"150.00", currency:"USD" } });
    setSession({ product:r.id });
    return `sản phẩm nháp ${short(r.id)}`;
  } },

{ id:"variant", ctx:"catalog", actor:"operator", api:"POST /products/{id}/variants",
  title:"Thêm variant",
  hint:"US 9 / black. Phát catalog.variant_added: ordering cần nó để từ chối size lạ, procurement cần nó để hiện 'M 8 / W 9.5' thay vì uuid.",
  async run(){
    const S = need("product");
    const r = await call("POST", `/products/${S.product}/variants`, { as:"operator",
      body:{ size:"US 9", color:"black", merchant_ref:"EX-AT90-9-BLK" } });
    setSession({ variant:r.id });
    return `variant ${short(r.id)}`;
  } },

{ id:"confirm", ctx:"catalog", actor:"operator", api:"POST /products/{id}/confirm-listing",
  title:"Xác nhận listing",
  hint:"Operator lấy từ Bearer token, không phải header — 'ai bảo đảm' không để người gọi tự khai.",
  async run(){
    const S = need("product");
    await call("POST", `/products/${S.product}/confirm-listing`, { as:"operator" });
    return "listing đã verified";
  } },

{ id:"measure", ctx:"catalog", actor:"operator", api:"POST /products/{id}/measure",
  title:"Cân và đo thật",
  hint:"1250 g · 340×230×130 mm. Số này feed của shop không có.",
  async run(){
    const S = need("product");
    await call("POST", `/products/${S.product}/measure`, { as:"operator",
      body:{ weight_g:1250, length_mm:340, width_mm:230, height_mm:130 } });
    return "đã ghi 1250 g · 34×23×13 cm";
  } },

{ id:"publish", ctx:"catalog", actor:"operator", api:"POST /products/{id}/publish",
  title:"Publish",
  hint:"5 điều kiện. Event mang cả giá lẫn hộp, vì pricing không được import catalog để hỏi lại.",
  async run(){
    const S = need("product");
    await call("POST", `/products/${S.product}/publish`, { as:"operator" });
    return "published — outbox có product_published";
  } },

{ id:"relay1", ctx:"worker", actor:"worker", wait:true,
  title:"Chờ worker relay #1",
  hint:"4 event đi ra. Bước sau sẽ tự thử lại nếu chưa kịp: đó chính là 404 listing_not_found.",
  async run(report){
    report("đợi relay… (in-memory: 200 ms/lượt · -dsn: theo cờ -every của cmd/worker)");
    await sleep(900);
    return "đã chờ — nếu chưa kịp thì bước 8 sẽ thử lại";
  } },

{ id:"quote", ctx:"pricing", actor:"customer", api:"POST /quotes",
  title:"Xin báo giá",
  hint:"Chưa relay thì API trả 404 listing_not_found — bước này thử lại cho tới khi pricing nghe được sản phẩm.",
  async run(report){
    const S = need("product");
    const r = await retryOn(report, ["listing_not_found"], () =>
      call("POST", "/quotes", { as:"customer", body:{ product_id:S.product, lane:"us_forwarder" } }));
    setSession({ quote:r.id });
    return `quote ${short(r.id)}`;
  } },

{ id:"quote-read", ctx:"pricing", actor:"customer", api:"GET /quotes/{id}",
  title:"Khách đọc báo giá",
  hint:"Ghi trả id, đọc trả view. Số cọc lấy từ đây để bước 13 gửi ĐÚNG số.",
  async run(){
    const S = need("quote");
    const v = await call("GET", `/quotes/${S.quote}`, { as:"customer" });
    setSession({ total:v.home.total.amount, deposit:v.home.deposit.amount, chargeable:v.chargeable_g });
    return `${v.class} · ${v.chargeable_g} g · tổng ${money(v.home.total)} · cọc ${money(v.home.deposit)}`;
  } },

{ id:"accept", ctx:"pricing", actor:"customer", api:"POST /quotes/{id}/accept",
  title:"Khách đồng ý",
  hint:"Quá 48 h thì quote vừa chuyển expired vừa trả 409 — commit rồi mới báo lỗi.",
  async run(){
    const S = need("quote");
    await call("POST", `/quotes/${S.quote}/accept`, { as:"customer" });
    return "accepted — outbox có quote_accepted";
  } },

{ id:"relay2", ctx:"worker", actor:"worker", wait:true,
  title:"Chờ worker relay #2",
  hint:"quote_accepted sang ordering. Không có endpoint đọc projection này, nên bước 12 thử lại trên 409 quote_not_accepted.",
  async run(report){
    report("đợi relay…");
    await sleep(900);
    return "đã chờ";
  } },

{ id:"order", ctx:"ordering", actor:"customer", api:"POST /orders",
  title:"Đặt hàng",
  hint:"Không có customer_id trong body — đơn thuộc người giữ token. Một quote một đơn. variant_id phải là variant của đúng sản phẩm quote đã báo giá.",
  async run(report){
    const S = need("quote", "variant");
    // variant_unknown cũng là "chưa relay tới", không phải từ chối thật:
    // ordering chỉ biết variant sau khi catalog.variant_added được giao.
    const r = await retryOn(report, ["quote_not_accepted", "variant_unknown"], () =>
      call("POST", "/orders", { as:"customer", body:{ quote_id:S.quote, variant_id:S.variant } }));
    setSession({ order:r.id });
    return `đơn ${short(r.id)}`;
  } },

{ id:"deposit", ctx:"ordering", actor:"operator", api:"POST /orders/{id}/deposit",
  title:"Thu cọc 50 %",
  hint:"Phải đúng số. Sai một đồng là 409 wrong_amount — cọc thiếu không phải cam kết nhỏ hơn.",
  async run(){
    const S = need("order", "deposit");
    await call("POST", `/orders/${S.order}/deposit`, { as:"operator",
      body:{ amount:S.deposit, currency:"VND" } });
    return `đã thu ${S.deposit} ₫`;
  } },

{ id:"relay3", ctx:"worker", actor:"worker", wait:true, api:"GET /purchase-tasks (poll)",
  title:"Chờ procurement mở việc mua",
  hint:"deposit_paid sang procurement → mở PurchaseTask, hỏi shop qua ACL. Chưa shop nào có API nên task nằm chờ người. Bước này POLL danh sách việc, nên bạn thấy đúng lúc nó xuất hiện.",
  async run(report){
    const S = need("order");
    const task = await poll(report, "chờ task", async () => {
      const rows = await call("GET", "/purchase-tasks", { as:"operator" });
      return (rows || []).find(t => t.order_id === S.order) || null;
    });
    setSession({ task:task.id });
    return `task ${short(task.id)} · shop tính bằng ${task.currency}`;
  } },

{ id:"tasks", ctx:"procurement", actor:"operator", api:"GET /purchase-tasks",
  title:"Người mua xem việc",
  hint:"Không có route tạo task: luật 'không mua khi chưa có cọc' giữ bằng endpoint không tồn tại.",
  async run(){
    const rows = await call("GET", "/purchase-tasks", { as:"operator" });
    return `${(rows || []).length} việc đang mở`;
  } },

{ id:"buy", ctx:"procurement", actor:"operator", api:"POST /purchase-tasks/{id}/confirm",
  title:"Mua xong, nhập biên nhận",
  hint:"163.22 USD là Actual đầu tiên. Trả bằng VND ở đây là 409 paid_currency.",
  async run(){
    const S = need("task");
    await call("POST", `/purchase-tasks/${S.task}/confirm`, { as:"operator",
      body:{ reference:reference(), paid:"163.22", currency:"USD" } });
    setSession({ reference:reference() });
    return `${reference()} · đã trả 163.22 USD`;
  } },

{ id:"relay4", ctx:"worker", actor:"worker", wait:true, api:"GET /orders/{id} + GET /parcels (poll)",
  title:"Chờ 4 listener của purchase_confirmed",
  hint:"Một event, bốn người nghe: ordering → purchased (không quay đầu), logistics → chờ thùng, pricing → tiền hàng thật, reporting → mã đơn shop.",
  async run(report){
    const S = need("order");
    await poll(report, "chờ đơn sang purchased", async () => {
      const o = await call("GET", `/orders/${S.order}`, { as:"operator" });
      return o.status === "purchased" ? o : null;
    });
    const parcel = await poll(report, "chờ kho biết có thùng", async () => {
      const rows = await call("GET", "/parcels", { as:"operator" });
      return (rows || []).find(p => p.order_id === S.order) || null;
    });
    setSession({ parcel:parcel.id });
    return `đơn purchased · thùng ${short(parcel.id)} (${parcel.reference}) đang chờ về kho`;
  } },

{ id:"receive", ctx:"logistics", actor:"operator", api:"POST /parcels/{id}/receive",
  title:"Cân thùng ở kho",
  hint:"Cân tính phí = max(1250 g, 2034 g thể tích) → 2500 g. Carrier tính tiền trên 2500 g.",
  async run(){
    const S = need("parcel");
    await call("POST", `/parcels/${S.parcel}/receive`, { as:"operator",
      body:{ weight_g:1250, length_mm:340, width_mm:230, height_mm:130 } });
    return "đã cân — outbox có parcel_received";
  } },

{ id:"batch", ctx:"logistics", actor:"operator", api:"POST /batches",
  title:"Mở lô hàng",
  hint:"Tuyến phải có LaneRule (từ pricing.lane_defined lúc seed) — không thì chặn ngay lúc mở, không đợi tới lúc chia cước.",
  async run(){
    const r = await call("POST", "/batches", { as:"operator", body:{ lane:"us_forwarder" } });
    setSession({ batch:r.id });
    return `lô ${short(r.id)}`;
  } },

{ id:"pack", ctx:"logistics", actor:"operator", api:"POST /batches/{id}/parcels → /close",
  title:"Xếp thùng vào lô, dán kín",
  hint:"Xếp thùng sửa hai aggregate trong một transaction; không phát event — lô lên tiếng khi ship.",
  async run(report){
    const S = need("batch", "parcel");
    await call("POST", `/batches/${S.batch}/parcels`, { as:"operator", body:{ parcel_id:S.parcel } });
    report("đã xếp thùng, giờ dán kín lô");
    await call("POST", `/batches/${S.batch}/close`, { as:"operator" });
    return "lô closed";
  } },

{ id:"ship", ctx:"logistics", actor:"operator", api:"POST /batches/{id}/ship",
  title:"Ship lô, nhập hoá đơn carrier",
  hint:"27.50 USD cho cả lô. Đây là route ghi duy nhất trả body: phần chia cước.",
  async run(){
    const S = need("batch");
    const allocs = await call("POST", `/batches/${S.batch}/ship`, { as:"operator",
      body:{ freight:"27.50", currency:"USD" } });
    const a = (allocs || [])[0];
    return a ? `chia: đơn ${short(a.order_id)} · ${a.chargeable_g} g · ${money(a.freight)}` : "đã ship";
  } },

{ id:"relay5", ctx:"worker", actor:"worker", wait:true, api:"GET /orders/{id} (poll)",
  title:"Chờ ordering nghe batch_shipped",
  hint:"Pending() chụp một lần đầu lượt, nên event sinh ra trong lượt đợi lượt sau — vì thế các lượt ra 5/2/2/2/6/3.",
  async run(report){
    const S = need("order");
    await poll(report, "chờ đơn sang in_transit", async () => {
      const o = await call("GET", `/orders/${S.order}`, { as:"operator" });
      return o.status === "in_transit" ? o : null;
    });
    return "đơn in_transit";
  } },

{ id:"balance", ctx:"ordering", actor:"operator", api:"POST /orders/{id}/balance",
  title:"Khách trả nốt 50 %",
  hint:"Chỉ thu được khi hàng đang bay; trước đó là 409 not_in_transit.",
  async run(){
    const S = need("order", "deposit");
    await call("POST", `/orders/${S.order}/balance`, { as:"operator",
      body:{ amount:S.deposit, currency:"VND" } });
    return `đã thu nốt ${S.deposit} ₫`;
  } },

{ id:"deliver", ctx:"ordering", actor:"operator", api:"POST /orders/{id}/deliver",
  title:"Giao tận nhà",
  hint:"Giao mà chưa thu đủ là 409 balance_unpaid.",
  async run(){
    const S = need("order");
    await call("POST", `/orders/${S.order}/deliver`, { as:"operator" });
    return "delivered";
  } },

{ id:"relay6", ctx:"worker", actor:"worker", wait:true, api:"GET /me/orders (poll)",
  title:"Chờ màn hình khách phản chiếu",
  hint:"Bảng ghi đúng ngay, bảng đọc đúng sau một nhịp — đó là hình dạng bình thường của hệ event-driven.",
  async run(report){
    const S = need("order");
    const row = await poll(report, "chờ /me/orders", async () => {
      const rows = await call("GET", "/me/orders", { as:"customer" });
      const r = (rows || []).find(x => x.order_id === S.order);
      return r && r.status === "delivered" ? r : null;
    });
    return `khách thấy: ${row.product_name} · ${row.status} · ${row.tracking}`;
  } },

{ id:"recon", ctx:"reporting", actor:"operator", api:"GET /reconciliations/{order}",
  title:"Đối soát: lời hay lỗ",
  hint:"Ước tính 163.22 + 25.00, thật 163.22 + 27.50 → −2.50 USD. Hai con số nằm hai bảng, không ghi đè.",
  async run(){
    const S = need("order");
    const v = await call("GET", `/reconciliations/${S.order}`, { as:"operator" });
    return v.complete
      ? `variance ${money(v.variance)} (ước tính ${v.quoted.goods.amount}+${v.quoted.freight.amount} · thật ${v.actual.goods.amount}+${v.actual.freight.amount})`
      : "chưa đủ hai nửa actual";
  } },
];

export const STEP_INDEX = Object.fromEntries(STEPS.map((s, i) => [s.id, i]));
