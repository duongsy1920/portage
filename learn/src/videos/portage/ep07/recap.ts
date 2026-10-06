import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Một màn hình của khách cần dữ liệu từ bốn vùng, mà bốn vùng không được import lẫn nhau." },
    { at: 80, text: "Nên không ai đi hỏi ai. Mỗi vùng kể lại việc đã xảy ra, và một projector nghe rồi ghi." },
    { at: 130, text: "Kết quả là một dòng rộng trong order_summaries: màn hình đọc đúng một câu SELECT." },
    { at: 180, text: "Cái giá là dòng đó chậm hơn sự thật vài trăm mili giây — chính là cái 404 của tập 3." },
  ],
  asked: [
    "Vì sao không JOIN giữa các bounded context?",
    "Read model được dựng lại từ đâu nếu nó hỏng?",
    "CQRS giải quyết vấn đề gì, và tốn gì?",
  ],
  next: { label: "MÙA 3 · TẬP 8", text: "Anti-corruption layer: máy được tạo nháp, nhưng không được xác nhận." },
} as const satisfies Recap;
