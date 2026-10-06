import { REPO } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  title: "Bốn câu mang theo, và hết mùa hai",
  points: [
    { at: 30, text: "Domain event là một việc đã xảy ra, đặt tên ở thì quá khứ — không ai từ chối được nữa." },
    { at: 80, text: "Aggregate chỉ ghi event vào sổ. Tầng app lấy ra và ghi vào outbox, trong cùng transaction." },
    { at: 130, text: "Phát thẳng từ aggregate thì một thay đổi bị rollback vẫn kịp bắn event ra ngoài." },
    { at: 180, text: `Repo có ${REPO.events} loại event và ${REPO.subscribeLines} dòng đăng ký nghe. Mùa 3 xem chúng chạy thật.` },
  ],
  asked: [
    "Domain event khác message của queue ở chỗ nào?",
    "Vì sao aggregate không tự publish event?",
    "Vì sao tên event luôn ở thì quá khứ?",
  ],
  next: { label: "MÙA 3 · TẬP 1", text: "Portage: vì sao project thật này bị chia làm năm vùng." },
} as const satisfies Recap;
