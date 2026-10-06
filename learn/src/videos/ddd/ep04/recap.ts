import { REPO } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Repository là một interface do domain khai báo, không phải một class kế thừa từ ORM." },
    { at: 80, text: "Chiều phụ thuộc bị đảo: adapter biết domain, domain không biết adapter nào tồn tại." },
    { at: 130, text: "Luật vàng: internal/domain chỉ import stdlib và uuid. Có 7 test canh bằng go/ast." },
    { at: 180, text: `Đổi lại: domain test được mà không cần database — ${REPO.testsNoDocker} test chạy dưới 2 giây.` },
  ],
  asked: [
    "Repository trong DDD khác repository của Doctrine ở chỗ nào?",
    "Đảo chiều phụ thuộc để làm gì?",
    "Làm sao ép được luật kiến trúc mà không dựa vào kỷ luật?",
  ],
  next: { label: "MÙA 2 · TẬP 5", text: "Bounded Context: một từ, một nghĩa — và repo này có ba kiểu tên là Variant." },
} as const satisfies Recap;
