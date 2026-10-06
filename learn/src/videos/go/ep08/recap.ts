import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "append trả về slice mới. Quên gán lại thì mất dữ liệu, và không có lỗi nào báo." },
    { at: 80, text: "Vì slice là một struct nhỏ, nên nó bị chép khi truyền đi — đúng luật receiver của tập 4." },
    { at: 130, text: "Map duyệt theo thứ tự ngẫu nhiên, cố ý. Cần thứ tự thì phải sort, không có cách khác." },
    { at: 180, text: "Struct so sánh bằng == được, trừ khi bên trong có slice hoặc map." },
  ],
  asked: [
    "Vì sao append phải gán lại kết quả?",
    "Vì sao Go xáo thứ tự khi duyệt map?",
    "Slice và array khác nhau ở chỗ nào?",
  ],
  next: { label: "TẬP 9", text: "defer, panic, recover: chạy lúc nào, và vì sao panic giết cả process chứ không chỉ một request." },
} as const satisfies Recap;
