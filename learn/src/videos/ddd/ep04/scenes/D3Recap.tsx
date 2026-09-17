import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 2, episode 4 — recall. */
export const D3Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 4 · NHỚ LẠI"
    accent="#9D7CD8"
    points={[
      { at: 30, text: "Repository là một interface do domain khai báo, không phải một class kế thừa từ ORM." },
      { at: 80, text: "Chiều phụ thuộc bị đảo: adapter biết domain, domain không biết adapter nào tồn tại." },
      { at: 130, text: "Luật vàng: internal/domain chỉ import stdlib và uuid. Có 7 test canh bằng go/ast." },
      { at: 180, text: "Đổi lại: domain test được mà không cần database — 278 test chạy dưới 2 giây." },
    ]}
    asked={[
      "Repository trong DDD khác repository của Doctrine ở chỗ nào?",
      "Đảo chiều phụ thuộc để làm gì?",
      "Làm sao ép được luật kiến trúc mà không dựa vào kỷ luật?",
    ]}
    nextLabel="MÙA 2 · TẬP 5"
    nextText="Bounded Context: một từ, một nghĩa — và repo này có ba kiểu tên là Variant."
  />
);
