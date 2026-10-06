import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "PHP dựng lại thế giới mỗi request. Go dựng một lần rồi phục vụ mọi request trên đó." },
    { at: 80, text: "Nên mỗi request là một goroutine, và chúng dùng chung mọi thứ main() đã dựng." },
    { at: 130, text: "Dùng chung thì phải khoá: đó là lý do có sync.Mutex ở 25 chỗ trong repo." },
    { at: 180, text: "Sống lâu thì phải huỷ được từng việc: đó là lý do ctx có mặt ở 361 chỗ." },
  ],
  asked: [
    "Goroutine khác thread ở chỗ nào?",
    "go test -race dò ra cái gì?",
    "context.Context dùng để làm gì?",
  ],
  next: { label: "TẬP 2", text: "Đọc một dòng Go: vì sao kiểu viết sau tên, và class đi đâu mất?" },
} as const satisfies Recap;
