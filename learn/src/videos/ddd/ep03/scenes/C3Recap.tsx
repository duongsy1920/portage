import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 2, episode 3 — recall. */
export const C3Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 3 · NHỚ LẠI"
    accent="#9D7CD8"
    points={[
      { at: 30, text: "Aggregate là một cụm dữ liệu phải luôn đúng cùng nhau, và chỉ sửa được qua một cửa." },
      { at: 80, text: "Cửa đó là root. Field bên trong viết thường nên package khác không chạm được — luật tập 2 mùa 1." },
      { at: 130, text: "Invariant là điều phải đúng ngay lập tức. Nó quyết định ranh giới aggregate nằm ở đâu." },
      { at: 180, text: "Một transaction một aggregate. Phải sửa hai cái cùng lúc nghĩa là chia sai, hoặc cần event." },
    ]}
    asked={[
      "Aggregate root để làm gì?",
      "Invariant là gì, và nó khác validation ở chỗ nào?",
      "Vì sao một transaction chỉ nên đụng một aggregate?",
    ]}
    nextLabel="MÙA 2 · TẬP 4"
    nextText="Repository là cổng: vì sao domain khai interface, còn Postgres cắm vào từ bên ngoài."
  />
);
