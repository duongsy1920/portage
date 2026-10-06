import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Field chỉ có kiểu, không có tên, là embedding — nhúng một kiểu vào struct." },
    { at: 80, text: "Cơ chế là tìm method xuống một tầng, không phải kế thừa: không lớp cha, không ghi đè." },
    { at: 130, text: "Hai kiểu cùng nhúng một thứ không thành họ hàng. Muốn nhận nhiều kiểu thì dùng interface." },
    { at: 180, text: "Nhờ nó mà 9 kiểu ID chỉ tốn một dòng mỗi cái, và 8 aggregate nhúng shared.Events." },
  ],
  asked: [
    "Embedding khác kế thừa ở chỗ nào?",
    "Method promotion là gì?",
    "Vì sao dùng OrderID riêng thay vì truyền uuid.UUID?",
  ],
  next: { label: "TẬP 8", text: "slice, map và ba cái bẫy: append phải gán lại, và map duyệt theo thứ tự ngẫu nhiên." },
} as const satisfies Recap;
