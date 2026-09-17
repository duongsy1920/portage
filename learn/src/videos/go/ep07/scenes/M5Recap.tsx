import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 1, episode 7 — recall. */
export const M5Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 7 · NHỚ LẠI"
    points={[
      { at: 30, text: "Field chỉ có kiểu, không có tên, là embedding — nhúng một kiểu vào struct." },
      { at: 80, text: "Cơ chế là tìm method xuống một tầng, không phải kế thừa: không lớp cha, không ghi đè." },
      { at: 130, text: "Hai kiểu cùng nhúng một thứ không thành họ hàng. Muốn nhận nhiều kiểu thì dùng interface." },
      { at: 180, text: "Nhờ nó mà 9 kiểu ID chỉ tốn một dòng mỗi cái, và 8 aggregate nhúng shared.Events." },
    ]}
    asked={[
      "Embedding khác kế thừa ở chỗ nào?",
      "Method promotion là gì?",
      "Vì sao dùng OrderID riêng thay vì truyền uuid.UUID?",
    ]}
    nextLabel="TẬP 8"
    nextText="slice, map và ba cái bẫy: append phải gán lại, và map duyệt theo thứ tự ngẫu nhiên."
  />
);
