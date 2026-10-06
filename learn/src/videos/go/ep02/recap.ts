import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Tên trước, kiểu sau — kể cả kiểu trả về, nằm ở cuối chữ ký hàm." },
    { at: 80, text: "Chữ HOA cho package khác dùng, chữ thường thì không. Không có từ khoá public." },
    { at: 130, text: "Riêng tư tính theo package, không theo struct. Muốn giấu kỹ hơn thì tách package." },
    { at: 180, text: "Struct chỉ giữ dữ liệu. Hàm viết ở ngoài, gắn vào kiểu bằng cụm trước tên hàm." },
  ],
  asked: [
    "Vì sao Go viết kiểu sau tên biến?",
    "Chữ thường trong Go giấu được khỏi ai, và không giấu được khỏi ai?",
    "Go tạo object kiểu gì, khi không có từ khoá new?",
  ],
  next: { label: "TẬP 3", text: "Lỗi là giá trị trả về: vì sao Go không có try/catch, và (Money, error) nghĩa là gì?" },
} as const satisfies Recap;
