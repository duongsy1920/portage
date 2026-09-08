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
};

export function friendly(e) {
  if (!e) return "";
  return ERRORS[e.code] || `Lỗi chưa có mô tả: ${e.code}. ${e.message || ""}`;
}
