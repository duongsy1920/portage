import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Aggregate là một cụm dữ liệu phải luôn đúng cùng nhau, và chỉ sửa được qua một cửa." },
    { at: 80, text: "Cửa đó là root. Field bên trong viết thường nên package khác không chạm được — luật tập 2 mùa 1." },
    { at: 130, text: "Invariant là điều phải đúng ngay lập tức. Nó quyết định ranh giới aggregate nằm ở đâu." },
    { at: 180, text: "Một transaction một aggregate. Phải sửa hai cái cùng lúc nghĩa là chia sai, hoặc cần event." },
  ],
  asked: [
    "Aggregate root để làm gì?",
    "Invariant là gì, và nó khác validation ở chỗ nào?",
    "Vì sao một transaction chỉ nên đụng một aggregate?",
  ],
  next: { label: "MÙA 2 · TẬP 4", text: "Repository là một port: vì sao domain khai interface, còn Postgres cắm vào từ bên ngoài." },
} as const satisfies Recap;
