import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Bảng đọc chỉ đầy sau khi relay chạy, nên 404 ngay sau khi tạo là chuyện bình thường." },
    { at: 80, text: "404 ở đây nghĩa là “chưa tới”, không phải “không có”. Khách nên thử lại, không nên sửa request." },
    { at: 130, text: "Repo có một bảng duy nhất biến tên lỗi thành mã HTTP: 92 dòng, ba mã cho ba câu chuyện." },
    { at: 180, text: "Không nằm trong bảng thì là 500 — nghĩa là bug của mình, và chỉ ghi log, không kể cho khách." },
  ],
  asked: [
    "Nhất quán sau cùng là gì, và nó buộc API phải đổi thế nào?",
    "Khi nào trả 404 và khi nào trả 409?",
    "Client nên xử lý 404 tạm thời kiểu gì?",
  ],
  next: { label: "MÙA 3 · TẬP 4", text: "Port và Adapter: vì sao 278 test chạy dưới 2 giây mà không cần Docker." },
} as const satisfies Recap;
