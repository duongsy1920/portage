import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  title: "Bốn câu mang theo, và hết mùa một",
  points: [
    { at: 30, text: `Repo có ${GO_USAGE.goroutines} goroutine, ${GO_USAGE.mutex} chỗ khoá, ${GO_USAGE.channelsDeclared} channel tự tạo và ${GO_USAGE.channelsRead} lần đọc channel — nói đúng vì đó là sự thật.` },
    { at: 80, text: "Dữ liệu dùng chung thì dùng khoá. Chuyển giao việc thì dùng channel. Chọn theo hình dạng bài toán." },
    { at: 130, text: "“Đừng chia sẻ bộ nhớ, hãy giao tiếp” là lời khuyên, không phải luật cấm." },
    { at: 180, text: "Giờ bạn đọc được mọi file trong repo. Mùa 2 nói về DDD, bằng chính ví dụ Symfony đã quen." },
  ],
  asked: [
    "Khi nào dùng channel, khi nào dùng mutex?",
    "Channel không có bộ đệm khác channel có bộ đệm ở chỗ nào?",
    "Deadlock trong Go xảy ra kiểu gì?",
  ],
  next: { label: "MÙA 2 · TẬP 1", text: "Mô hình thiếu máu: entity toàn getter setter, và luật nghiệp vụ rơi mất ở đâu." },
} as const satisfies Recap;
