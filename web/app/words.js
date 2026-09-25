// words.js — every value the API sends as a code, and the words a person reads.
//
// Two rules this file exists to enforce, both learned the hard way:
//
//   1. NO machine word reaches the screen. "issued", "open", "none" and
//      "branded" mean something to whoever wrote the API and nothing to the
//      person using it.
//   2. Every number and every state can say WHERE IT CAME FROM. If somebody
//      looks at a figure and has to ask, the screen failed, not the person.
//
// So each entry has words, a tone for the pill, and where it helps, a `why`
// the screen can show next to it.

/* ── where the box is (reporting.Tracking) ───────────────────────────────── */
export const TRACKING = {
  none:     { words: "chưa có gì để gửi", tone: "flat",
              why: "Chưa mua nên chưa có kiện hàng nào." },
  expected: { words: "kho đang chờ hàng về", tone: "info",
              why: "Đã mua ở shop, kho bên Mỹ đang đợi shop giao tới." },
  received: { words: "đã về kho, đã cân", tone: "info",
              why: "Hàng tới kho và được cân thật. Số cân này là số tính cước." },
  shipped:  { words: "đã bay", tone: "ok",
              why: "Đã lên máy bay trong một lô gom nhiều đơn, nên cước rẻ hơn gửi lẻ." },
};

/* ── the order itself (ordering.OrderStatus) ─────────────────────────────── */
export const ORDER = {
  awaiting_deposit: { words: "chờ bạn chuyển cọc", tone: "warn",
                      why: "Bên mình ứng trước 100% tiền hàng, nên cần cọc trước khi đi mua." },
  deposited:        { words: "đã nhận cọc, đang đi mua", tone: "info",
                      why: "Nhân viên đang mua hộ ở shop." },
  purchased:        { words: "đã mua xong", tone: "info",
                      why: "Từ lúc này không huỷ được nữa: hàng đã mua riêng cho bạn." },
  in_transit:       { words: "đang trên đường về", tone: "info",
                      why: "Đã rời kho Mỹ." },
  delivered:        { words: "đã giao", tone: "ok", why: "" },
  cancelled:        { words: "đã huỷ", tone: "bad",
                      why: "Số tiền hoàn lại tuỳ thời điểm huỷ." },
  purchase_failed:  { words: "shop không bán được", tone: "bad",
                      why: "Hết hàng hoặc shop từ chối. Bạn được hoàn đủ cọc nếu huỷ." },
};

/* ── a quote (pricing.QuoteStatus) ───────────────────────────────────────── */
export const QUOTE = {
  issued:   { words: "còn hiệu lực", tone: "ok",
              why: "Giá đã chốt cứng trong báo giá này, tỷ giá mai đổi cũng không ảnh hưởng." },
  accepted: { words: "bạn đã đồng ý", tone: "info", why: "" },
  expired:  { words: "đã hết hạn", tone: "bad",
              why: "Báo giá chỉ giữ 48 giờ, vì tỷ giá và giá cước đổi. Xin lại một cái mới." },
};

/* ── a box at the Denver warehouse (logistics.ParcelStatus) ───────────────── */
export const PARCEL = {
  expected: { words: "chờ shop giao tới kho", tone: "warn",
              why: "Đã mua ở shop. Khi hộp tới kho, cân và đo nó: số đó là số tính cước thật." },
  received: { words: "đã cân, chờ xếp lô", tone: "info",
              why: "Kiện nằm ở kho. Xếp vào lô đang gom để bay chung, cước rẻ hơn gửi lẻ." },
  batched:  { words: "đã xếp vào lô", tone: "info", why: "" },
  shipped:  { words: "đã bay", tone: "ok", why: "" },
};

/* ── a consolidation batch (logistics.BatchStatus) ───────────────────────── */
export const BATCH = {
  open:    { words: "đang gom", tone: "info",
             why: "Còn nhận thêm kiện. Khi đủ chuyến thì dán kín." },
  closed:  { words: "đã dán kín, chờ bay", tone: "warn",
             why: "Không thêm kiện được nữa. Khi hãng bay gửi hoá đơn, nhập tổng cước để chia cho từng kiện." },
  shipped: { words: "đã bay", tone: "ok",
             why: "Cước cả lô đã chia cho từng kiện theo cân tính cước. Phần của mỗi đơn đi vào bảng đối chiếu lời lỗ." },
};

/* ── work at the shop (procurement.TaskStatus) ───────────────────────────── */
export const TASK = {
  open:      { words: "chưa mua", tone: "warn", why: "" },
  confirmed: { words: "đã mua", tone: "ok", why: "" },
  failed:    { words: "không mua được", tone: "bad", why: "" },
};

/* ── the forwarder's price list (pricing.GoodsClass) ─────────────────────── */
export const GOODS = {
  standard:    { words: "hàng thường", why: "Nhóm rẻ nhất trong bảng giá của bên vận chuyển." },
  branded:     { words: "hàng có thương hiệu",
                 why: "Giày và quần áo có thương hiệu bị bên vận chuyển tính giá cao hơn hàng thường." },
  electronics: { words: "hàng điện tử",
                 why: "Đồ điện tử tính giá riêng, và có loại còn bị hạn chế bay." },
  sensitive:   { words: "hàng cần lưu ý", why: "Nhóm bị hạn chế, giá cao nhất." },
};

/* ── our own category codes ──────────────────────────────────────────────────
 * The code is business DATA: an operator can define a new one through
 * POST /categories, so this dictionary can always fall behind. It falls back to
 * the code rather than hiding it, and the real fix is a display name on the
 * category itself — noted, not done, because that is a domain change.
 */
export const CATEGORY = {
  footwear:    "giày dép",
  apparel:     "quần áo",
  electronics: "đồ điện tử",
  luggage:     "vali, túi",
};

export function categoryWords(code) {
  return CATEGORY[code] || code;
}

/* ── what staff still have to do (reportingapp.Step) ─────────────────────── */
export const NEXT_STEP = {
  variant: "nhân viên đang nhập size",
  listing: "nhân viên đang kiểm đúng sản phẩm",
  measure: "nhân viên đang cân và đo",
  publish: "sắp xong, chờ đăng bán",
  "":      "đã xong",
};

/** say(dict, code) never shows a raw code, but never lies either: an unknown
 *  value is shown as-is with a note, because silently hiding it would make a
 *  new state invisible instead of merely ugly. */
export function say(dict, code) {
  return dict[code] || { words: code || "—", tone: "flat", why: "Trạng thái mới, màn hình chưa có mô tả." };
}

/* ── API error codes, in words a person can act on ───────────────────────── */
export const ERRORS = {
  unauthenticated:         "Chưa có chìa khoá, hoặc chìa sai. Kiểm ở phần Chìa khoá cuối trang.",
  forbidden:               "Chìa khoá này không có quyền làm việc đó.",
  category_not_found:      "Ngành hàng đó chưa được mở.",
  merchant_not_found:      "Không tìm thấy shop đó. Có thể máy chủ vừa khởi động lại.",
  merchant_inactive:       "Shop này đang tạm ngưng, chưa mua hộ được.",
  sourcing_not_allowed:    "Shop này không nhận link do khách tự gửi. Nhắn nhân viên để họ thêm hộ.",
  price_currency:          "Giá phải cùng loại tiền với shop.",
  malformed_amount:        "Số tiền không đúng dạng. Tiền Việt không có phần thập phân.",
  listing_not_found:       "Sản phẩm chưa đăng bán xong. Vài giây nữa thử lại.",
  no_variants:             "Phải có ít nhất một size trước khi đăng bán.",
  unverified:              "Còn thiếu bước xác nhận hoặc bước cân.",
  duplicate_variant:       "Size và màu này đã có rồi.",
  unnamed_variant:         "Sản phẩm đã có size khác, nên size mới phải có tên.",
  suspected_duplicate:     "Link này đã có người gửi trước. Nhân viên cần xử lý trùng trước khi đăng bán.",
  quote_not_issued:        "Báo giá không còn hiệu lực.",
  quote_expired:           "Báo giá đã hết hạn, xin lại cái mới.",
  quote_not_accepted:      "Chưa kịp ghi nhận việc bạn đồng ý. Thử lại sau một nhịp.",
  quote_already_used:      "Báo giá này đã tạo đơn rồi.",
  variant_unknown:         "Size này hệ thống chưa phát ra. Tải lại trang rồi chọn lại.",
  variant_not_for_product: "Size đó thuộc sản phẩm khác.",
  wrong_amount:            "Phải đúng số tiền, không thiếu không thừa.",
  not_awaiting_deposit:    "Đơn này không còn chờ cọc.",
  paid_currency:           "Số tiền phải cùng loại tiền với shop.",
  empty_reference:         "Phải có mã đơn của shop.",
  unreachable:             "Không gọi được máy chủ. Máy chủ còn chạy không?",
  incomplete_parcel_spec:  "Phải đủ bốn số: cân nặng, dài, rộng, cao.",
  parcel_not_expected:     "Kiện này đã được cân rồi.",
  parcel_not_received:     "Kiện chưa được cân ở kho, chưa xếp lô được.",
  parcel_not_batched:      "Kiện chưa nằm trong lô nào.",
  batch_not_open:          "Lô đã dán kín, không thêm kiện được nữa. Mở lô mới.",
  batch_not_closed:        "Phải dán kín lô trước khi cho bay.",
  batch_empty:             "Lô chưa có kiện nào.",
  duplicate_parcel:        "Kiện này đã nằm trong lô rồi.",
  invalid_freight:         "Cước phải là một số dương.",
  lane_rule_not_found:     "Tuyến bay này chưa có bảng giá cước.",
  not_in_transit:          "Đơn chưa bay nên chưa thu phần còn lại hay giao được.",
  balance_unpaid:          "Chưa thu đủ phần còn lại nên chưa giao được.",
  already_delivered:       "Đơn này đã giao rồi.",
};

export function friendly(e) {
  if (!e) return "";
  return ERRORS[e.code] || `Lỗi chưa có mô tả: ${e.code}. ${e.message || ""}`;
}

/* ── the journey strip: the order of the stages, and what each one holds ─────
 * ORDER and TRACKING answer two different questions (money and commitment vs
 * where the box is). The strip puts them back on one line in the order a
 * person lives through them. Each stage shows ONLY a figure or a date the API
 * actually sent for it; a stage the API says nothing about stays blank rather
 * than showing a number somebody made up (rule 2, turned into a picture).
 */
export const JOURNEY = [
  { key: "quote",     label: "Báo giá",    icon: "receipt" },
  { key: "deposit",   label: "Đã cọc",     icon: "wallet" },
  { key: "bought",    label: "Đã mua",     icon: "bag" },
  { key: "warehouse", label: "Kho Denver", icon: "scale" },
  { key: "flown",     label: "Đã bay",     icon: "plane" },
  { key: "delivered", label: "Đã giao",    icon: "home" },
];

const WAS_BOUGHT = ["purchased", "in_transit", "delivered"];

/** grams(1250) → "1.250 g" */
export function grams(g) {
  const n = Number(g);
  return Number.isFinite(n) && n > 0 ? n.toLocaleString("vi-VN") + " g" : "";
}

/** dayMonth("2026-09-20T…") → "20/09". Empty for a missing date. */
export function dayMonth(iso) {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return String(d.getDate()).padStart(2, "0") + "/" + String(d.getMonth() + 1).padStart(2, "0");
}

/**
 * journeyOf turns one order summary into the strip's cells.
 *   viewer   "customer" | "staff": decides which stage is "waiting for you"
 *   facts    what the caller knows beyond the summary, all of it read from the
 *            API by a screen allowed to: { boughtAt, paid } from the purchase
 *            task, { weighedG, receivedAt } from the parcel, { shippedAt,
 *            freight } from the batch. Absent facts leave the cell blank.
 * Returns { cells, reached, stopped } where each cell has
 *   state  done | now | yours | todo | stopped | off
 *   value  a money object or ""   text  a figure already in words   sub  a short line
 * A stage that is done but has nothing to show says "xong", so a finished
 * stage never looks like a missing one.
 */
export function journeyOf(o, viewer = "customer", facts = {}) {
  const tr = o.tracking;
  let reached = 0;
  if (o.deposit_paid) reached = 1;
  if (WAS_BOUGHT.includes(o.status) || ["expected", "received", "shipped"].includes(tr)) reached = 2;
  if (tr === "received" || tr === "shipped") reached = 3;
  if (tr === "shipped" || o.status === "in_transit" || o.status === "delivered") reached = 4;
  if (o.status === "delivered") reached = 5;

  const stopped = o.status === "cancelled" || o.status === "purchase_failed";
  const now = o.status === "delivered" ? -1 : reached + 1;

  let yours = -1;
  if (!stopped) {
    if (viewer === "customer" && o.status === "awaiting_deposit") yours = 1;
    if (viewer === "staff" && o.status === "deposited") yours = 2;
    if (o.status === "in_transit" && !o.balance_paid) yours = 5; // customer pays, staff takes it
  }

  const balance = o.balance; // from the API since T-002: the read model does the subtraction, not this screen
  const fill = {
    quote:     { value: o.total, sub: o.placed_at ? "đặt " + dayMonth(o.placed_at) : "" },
    deposit:   { value: o.deposit, sub: o.deposit_paid ? "đã nhận" : (stopped ? "" : "cần chuyển") },
    bought:    { value: facts.paid || "", sub: facts.boughtAt ? "mua " + dayMonth(facts.boughtAt) : "" },
    warehouse: { value: "", text: facts.weighedG ? grams(facts.weighedG) : "",
                 sub: facts.receivedAt ? "cân " + dayMonth(facts.receivedAt) : "" },
    flown:     { value: facts.freight || "", sub: facts.shippedAt ? "bay " + dayMonth(facts.shippedAt) : "" },
    delivered: o.status === "delivered"
      ? { value: "", sub: dayMonth(o.delivered_at) }
      : o.status === "in_transit" && !o.balance_paid && balance
        ? { value: balance, sub: "phần còn lại" }
        : { value: "", sub: "" },
  };

  const cells = JOURNEY.map((step, i) => {
    let state = i <= reached ? "done" : "todo";
    if (i === now) state = "now";
    if (i === yours) state = "yours";
    if (stopped && i === now) state = "stopped";
    if (stopped && i > now) state = "off";
    const f = fill[step.key];
    // a later stage's figure means nothing before the order gets there
    const shown = state === "todo" || state === "off" || state === "stopped" ? { value: "", text: "", sub: "" } : { text: "", ...f };
    if (state === "done" && !shown.value && !shown.text && !shown.sub) shown.sub = "xong";
    return { ...step, state, ...shown };
  });
  return { cells, reached, stopped };
}

/* what the stage marked "waiting for you" asks the person looking at it */
export const YOURS = {
  customer: { deposit: "chuyển cọc cho nhân viên", delivered: "chuyển phần còn lại cho nhân viên" },
  staff:    { bought: "đi mua ở shop", delivered: "thu phần còn lại khi khách chuyển" },
};

/* ORDER is written to the customer ("chờ bạn chuyển cọc"). Where that "bạn"
 * would be wrong on the staff screen, this says it from the desk's side. */
export const ORDER_FOR_STAFF = {
  awaiting_deposit: "chờ khách chuyển cọc",
};
