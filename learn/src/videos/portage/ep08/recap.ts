import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  title: "Bốn câu mang theo, và hết loạt",
  points: [
    { at: 30, text: "Điều khoản của Nike và phần lớn shop cấm cào trang, nên hệ thống không tự đọc trang shop." },
    { at: 80, text: "Dữ liệu vào bằng tay người, hoặc qua một adapter đọc URL mà người dùng tự đưa." },
    { at: 130, text: "Hai anti-corruption layer dịch thế giới bên ngoài sang từ vựng của domain — và dừng ở bản nháp." },
    { at: 180, text: "Xác nhận là lúc có người chịu trách nhiệm. Máy đoán sai một cỡ giày là mua nhầm hàng, tiền thật." },
  ],
  asked: [
    "Anti-corruption layer để làm gì?",
    "Vì sao không để máy tự xác nhận sản phẩm?",
    "Ràng buộc pháp lý ảnh hưởng tới kiến trúc kiểu gì?",
  ],
} as const satisfies Recap;
