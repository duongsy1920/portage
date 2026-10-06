import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: `defer hoãn một lệnh tới lúc hàm thoát — thoát kiểu gì cũng chạy. Repo có ${GO_USAGE.defers} lệnh.` },
    { at: 80, text: "Nhờ vậy khoá không bao giờ kẹt: chỉ có một chỗ mở khoá, không phải nhớ ở từng đường return." },
    { at: 130, text: "panic không phải exception. Vì tiến trình phục vụ mọi request, panic không chặn là chết cả hệ." },
    { at: 180, text: "recover chỉ chạy trong defer, và ở đây nó rollback rồi panic lại — dọn dẹp, không đi tiếp." },
  ],
  asked: [
    "defer chạy vào lúc nào, và theo thứ tự nào khi có nhiều cái?",
    "Vì sao panic ở Go nguy hiểm hơn fatal error ở PHP?",
    "Khi nào được dùng recover?",
  ],
  next: { label: "TẬP 10", text: "Đồng thời: 3 goroutine, 25 mutex, và 0 channel — vì sao repo không dùng channel, và phỏng vấn vẫn hỏi." },
} as const satisfies Recap;
