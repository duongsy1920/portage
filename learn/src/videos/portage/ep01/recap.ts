import type { Recap } from "../../registry";

/**
 * The closing screen's data. This episode's recap is a five-step CHAIN, drawn by its own
 * scene (S11Recap) rather than the shared Recap component; the interview questions are
 * not on screen — they exist for the learning hub, like every other episode's.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Một đơn sống 30 ngày, qua ba bên mình không điều khiển được." },
    { at: 80, text: "Nên không transaction nào giữ nổi nó từ đầu tới cuối." },
    { at: 130, text: "Nên phải tách thành nhiều aggregate, mỗi cái một lần ghi." },
    { at: 180, text: "Mà chúng gọi cùng một từ bằng nhiều nghĩa, nên tách tiếp thành bounded context." },
    { at: 230, text: "Không gọi hàm nhau được nữa, nên chúng kể lại bằng event." },
  ],
  asked: [
    "Vì sao một đơn hàng không thể nằm trong một transaction?",
    "Aggregate và bounded context khác nhau ở chỗ nào?",
    "Khi hai vùng không gọi hàm nhau được, chúng nói chuyện bằng gì?",
  ],
  next: { label: "TẬP 2", text: "Event đi từ vùng này sang vùng kia bằng đường nào, và vì sao không bao giờ mất?" },
} as const satisfies Recap;
