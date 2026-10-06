import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Anemic model: entity chỉ có getter và setter, luật nghiệp vụ nằm ở tầng service." },
    { at: 80, text: "Nó không sai về kỹ thuật. Vấn đề là cùng một luật bị chép ra nhiều nơi, rồi lệch nhau." },
    { at: 130, text: "Cách chữa nhỏ hơn tên gọi: method mang tên việc nghiệp vụ, luật kiểm ngay ở đầu method." },
    { at: 180, text: "Trong repo không có setter công khai nào trên aggregate — nên không ai lách qua luật được." },
  ],
  asked: [
    "Anemic model là gì, và vì sao nó phổ biến đến vậy?",
    "Đặt luật nghiệp vụ vào entity thì tầng service còn làm gì?",
    "Vì sao không nên có setter công khai trên aggregate?",
  ],
  next: { label: "MÙA 2 · TẬP 2", text: "Value Object và Entity: cái gì cần ID, cái gì không — và vì sao Money không cần." },
} as const satisfies Recap;
