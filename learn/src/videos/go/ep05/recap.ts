import { GO_USAGE, REPO } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Không có từ khoá implements. Có đủ method là dùng được, không cần khai báo gì." },
    { at: 80, text: "Interface khai ở package cần nó, nên mũi tên phụ thuộc chạy vào trong, không chạy ra." },
    { at: 130, text: `Muốn lỗi hiện đúng chỗ thì thêm var _ X = (*Y)(nil) — repo có ${GO_USAGE.assertions} dòng như vậy.` },
    { at: 180, text: `Interface nhỏ là lý do ${REPO.testsNoDocker} test chạy dưới 2 giây mà không cần Docker.` },
  ],
  asked: [
    "Interface của Go khác interface của PHP ở chỗ nào?",
    "Vì sao nên khai interface ở phía người dùng nó?",
    "var _ X = (*Y)(nil) để làm gì?",
  ],
  next: { label: "TẬP 6", text: "Zero value: Go không có null, nên mọi kiểu đều có sẵn một giá trị — và constructor phải tự vệ." },
} as const satisfies Recap;
