import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Cụm trước tên hàm là receiver — nó là $this của PHP, nhưng do bạn đặt tên." },
    { at: 80, text: "Không có dấu sao thì Go chép cả struct, nên method không sửa được bản gốc." },
    { at: 130, text: "Có dấu sao thì method nhận địa chỉ bản gốc, sửa vào đó là sửa thật." },
    { at: 180, text: `Repo có ${GO_USAGE.valueReceivers} receiver giá trị và ${GO_USAGE.pointerReceivers} receiver con trỏ — value object nhóm đầu, thứ có trạng thái nhóm sau.` },
  ],
  asked: [
    "Khi nào dùng receiver con trỏ, khi nào dùng receiver giá trị?",
    "Vì sao một method gán vào field mà không đổi được gì?",
    "Receiver giá trị có làm chậm không, khi struct lớn?",
  ],
  next: { label: "TẬP 5", text: "Interface ngầm: vì sao Go không có từ khoá implements, và vì sao port khai ở nơi CẦN?" },
} as const satisfies Recap;
