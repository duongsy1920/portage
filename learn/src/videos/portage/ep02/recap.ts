import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "“Lưu xong rồi publish” hỏng ở khe hở giữa hai bước: tiến trình chết là event mất luôn." },
    { at: 80, text: "Outbox là một bảng trong chính database đó. Event ghi vào nó cùng transaction với dữ liệu." },
    { at: 130, text: "Nhờ transaction đi trong ctx, INSERT vào outbox dùng đúng kết nối mà đơn hàng đang dùng." },
    { at: 180, text: "Bài toán đổi từ “hai hệ thống phải cùng thành công” thành “một database phải commit”." },
  ],
  asked: [
    "Outbox pattern giải quyết vấn đề gì?",
    "Vì sao không publish thẳng sau khi lưu?",
    "Làm sao bảo đảm INSERT vào outbox cùng transaction với dữ liệu?",
  ],
  next: { label: "MÙA 3 · TẬP 3", text: "Nhất quán sau cùng: vì sao 404 ngay sau khi tạo là thiết kế, không phải bug." },
} as const satisfies Recap;
