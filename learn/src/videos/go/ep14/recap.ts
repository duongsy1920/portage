import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: `select chờ nhiều channel, nhánh nào có trước thì chạy. Repo có ${GO_USAGE.selects} khối: ctx.Done() hoặc ticker.C.` },
    { at: 80, text: "ctx.Done() là một channel, đóng khi context bị huỷ — cách một việc đang chạy biết lúc dừng." },
    { at: 130, text: "WaitGroup đếm goroutine; close(start) thả cả bốn cùng lúc — test đua migrate đợt 18." },
    { at: 180, text: "Không đệm: gửi chờ nhận. Đệm 1: gửi xong đi luôn — goroutine không treo khi hết người đọc." },
  ],
  asked: [
    "select làm gì, và điều gì xảy ra khi hai channel cùng sẵn sàng?",
    "Channel có đệm khác không đệm ở chỗ nào, và khi nào cần đệm 1?",
    "Fan-in, fan-out, pipeline là gì? Vì sao repo này không có?",
  ],
  next: { label: "TẬP 15", text: "struct tag, JSON, và đường ra khỏi domain: 364 cái tag, và không cái nào nằm trong domain." },
} as const satisfies Recap;
