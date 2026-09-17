import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 2, episode 1 — recall. */
export const A4Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 1 · NHỚ LẠI"
    accent="#9D7CD8"
    points={[
      { at: 30, text: "Mô hình thiếu máu: entity chỉ có getter và setter, luật nghiệp vụ nằm ở tầng service." },
      { at: 80, text: "Nó không sai về kỹ thuật. Vấn đề là cùng một luật bị chép ra nhiều nơi, rồi lệch nhau." },
      { at: 130, text: "Cách chữa nhỏ hơn tên gọi: method mang tên việc nghiệp vụ, luật kiểm ngay ở đầu method." },
      { at: 180, text: "Trong repo không có setter công khai nào trên aggregate — nên không ai lách qua luật được." },
    ]}
    asked={[
      "Mô hình thiếu máu là gì, và vì sao nó phổ biến đến vậy?",
      "Đặt luật nghiệp vụ vào entity thì tầng service còn làm gì?",
      "Vì sao không nên có setter công khai trên aggregate?",
    ]}
    nextLabel="MÙA 2 · TẬP 2"
    nextText="Value Object và Entity: cái gì cần ID, cái gì không — và vì sao Money không cần."
  />
);
