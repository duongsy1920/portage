import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Lỗi là một giá trị trả về, không phải một cú nhảy. Hàm nào hỏng được thì trả (T, error)." },
    { at: 80, text: "Người gọi phải nhận cả hai. Muốn vứt lỗi thì phải viết dấu _ ra — không quên ngầm được." },
    { at: 130, text: `Mỗi lý do một biến ErrXxx, và ${GO_USAGE.wrap} chỗ trong repo bọc lỗi bằng %w để giữ nguyên nhân gốc.` },
    { at: 180, text: `Đọc nguyên nhân gốc bằng errors.Is, ${GO_USAGE.errorsIs} chỗ — không bao giờ so sánh bằng ==.` },
  ],
  asked: [
    "Vì sao Go chọn lỗi là giá trị thay vì exception?",
    "%w khác %v ở chỗ nào?",
    "Khi nào dùng errors.Is, khi nào errors.As?",
  ],
  next: { label: "TẬP 4", text: "Receiver: vì sao (m Money) chép cả struct, còn (e *Events) thì không?" },
} as const satisfies Recap;
