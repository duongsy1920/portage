import React from "react";
import { Recap } from "../../../../components/Recap";
import { GO_USAGE } from "../../../../data/portage";

/** Season 1, episode 5 — recall. */
export const K6Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 5 · NHỚ LẠI"
    points={[
      { at: 30, text: "Không có từ khoá implements. Có đủ method là dùng được, không cần khai báo gì." },
      { at: 80, text: "Interface khai ở package cần nó, nên mũi tên phụ thuộc chạy vào trong, không chạy ra." },
      { at: 130, text: `Muốn lỗi hiện đúng chỗ thì thêm var _ X = (*Y)(nil) — repo có ${GO_USAGE.assertions} dòng như vậy.` },
      { at: 180, text: "Interface nhỏ là lý do 278 test chạy dưới 2 giây mà không cần Docker." },
    ]}
    asked={[
      "Interface của Go khác interface của PHP ở chỗ nào?",
      "Vì sao nên khai interface ở phía người dùng nó?",
      "var _ X = (*Y)(nil) để làm gì?",
    ]}
    nextLabel="TẬP 6"
    nextText="Zero value: Go không có null, nên mọi kiểu đều có sẵn một giá trị — và constructor phải tự vệ."
  />
);
