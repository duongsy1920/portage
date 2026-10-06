import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: `Một hàm có thể nhận hàm làm tham số. InTx nhận một hàm ở ${GO_USAGE.inTxCalls} chỗ — đó là hình dạng của mọi use case trong repo.` },
    { at: 80, text: "Hàm không tên viết ngay tại chỗ dùng. Nó nhìn thấy và sửa được biến của hàm bao quanh — không cần khai use như PHP." },
    { at: 130, text: "InTx nhận hàm thay vì trả transaction, để commit và rollback là việc của InTx, không ai quên được." },
    { at: 180, text: "Trả về có tên (err error) cho defer ghi đè được kết quả — đó là cách lỗi commit vẫn báo ra ngoài." },
  ],
  asked: [
    "Closure là gì, và Go bắt biến ngoài khác PHP use ở chỗ nào?",
    "Vì sao một hàm lại nhận hàm làm tham số thay vì trả về một đối tượng?",
    "Named return value dùng để làm gì?",
  ],
  next: { label: "TẬP 12", text: "interface rỗng và type switch: 37 loại event chui qua một cái cửa thế nào?" },
} as const satisfies Recap;
