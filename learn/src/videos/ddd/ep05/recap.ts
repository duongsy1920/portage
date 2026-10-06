import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Bounded context là một vùng mà trong đó mỗi từ chỉ có một nghĩa." },
    { at: 80, text: "Repo này có ba kiểu cùng tên Variant, ở ba package, với ba hình dạng khác nhau." },
    { at: 130, text: "Gộp thành một định nghĩa chung nghe gọn, nhưng khiến mỗi thay đổi nhỏ phải xin phép bốn vùng." },
    { at: 180, text: "Chúng đồng bộ bằng event (catalog.variant_added), không import nhau. Ranh giới là package, và có test canh." },
  ],
  asked: [
    "Bounded context là gì, và nó khác module ở chỗ nào?",
    "Hai context cùng nói về một thứ thì đồng bộ kiểu gì?",
    "Vì sao không nên có một mô hình chung cho cả hệ thống?",
  ],
  next: { label: "MÙA 2 · TẬP 6", text: "Domain Event: aggregate chỉ GHI lại việc đã xảy ra, tầng app mới là nơi phát đi." },
} as const satisfies Recap;
