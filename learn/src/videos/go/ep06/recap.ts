import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Go không có null. Mỗi kiểu có sẵn một giá trị mặc định, và nó luôn dùng được ngay." },
    { at: 80, text: "Số là 0, chuỗi là rỗng, struct là mọi field bằng zero, con trỏ và interface là nil." },
    { at: 130, text: "Slice nil vẫn append và range được. Map nil đọc được nhưng ghi vào thì panic." },
    { at: 180, text: "Vì zero value hợp lệ nên constructor phải tự chặn — và quy ước là zero value phải an toàn." },
  ],
  asked: [
    "Zero value là gì, và vì sao Go chọn có nó thay vì null?",
    "Slice nil khác slice rỗng ở chỗ nào?",
    "Vì sao NewMoney panic thay vì trả error?",
  ],
  next: { label: "TẬP 7", text: "Embedding: trông như kế thừa, nhưng không phải — và đó là lý do có OrderID, QuoteID, ParcelID." },
} as const satisfies Recap;
