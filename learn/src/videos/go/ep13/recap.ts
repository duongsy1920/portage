import { GO_USAGE, REPO } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: `Repo có ${GO_USAGE.generics} hàm generic, và chỉ cần ${GO_USAGE.generics}. Generics là công cụ hiếm dùng, không phải rắc vào mọi nơi.` },
    { at: 80, text: `on[T] nhận closure có kiểu T, trả handler không kiểu — ${REPO.subscribeLines} dòng Subscribe, 0 adapter viết tay.` },
    { at: 130, text: "T suy ra từ tham số khi có tham số để suy; không có thì ghi rõ: into[contracts.X]." },
    { at: 180, text: "Interface nói về hành vi. Generic giữ nguyên kiểu đi qua một hàm. Hai câu hỏi, hai công cụ." },
  ],
  asked: [
    "Generics trong Go dùng để làm gì, và khác interface ở chỗ nào?",
    "Type inference: khi nào Go tự suy T, khi nào phải ghi rõ?",
    "Type constraint là gì? any và comparable khác nhau thế nào?",
  ],
  next: { label: "TẬP 14", text: "select, ticker, và dừng một việc đang chạy: bốn lần đọc channel của tập 10, đọc kỹ từng dòng." },
} as const satisfies Recap;
