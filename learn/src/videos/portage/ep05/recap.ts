import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Trước lúc mua ở shop Mỹ, huỷ đơn là hoàn cọc. Sau đó thì tiền đã đi và hàng đang bay." },
    { at: 80, text: "Luật đó là ba dòng ở đầu method: sai trạng thái thì trả error và không đổi gì cả." },
    { at: 130, text: "Mỗi lý do từ chối một tên lỗi riêng, nên khách nhận 409 kèm lý do đọc được, không phải 500." },
    { at: 180, text: "Không có setter, không có UPDATE thẳng. Một cửa duy nhất, và cửa đó kiểm luật trước." },
  ],
  asked: [
    "Điểm không thể quay đầu trong nghiệp vụ của bạn nằm ở đâu?",
    "Vì sao luật đó phải nằm trong aggregate chứ không ở tầng service?",
    "Aggregate từ chối thì tầng HTTP biết trả mã nào?",
  ],
  next: { label: "MÙA 3 · TẬP 6", text: "Báo giá là một bức ảnh, và đối soát: con số variance trả lời câu hỏi kinh doanh." },
} as const satisfies Recap;
