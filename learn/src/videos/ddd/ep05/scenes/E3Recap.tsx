import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 2, episode 5 — recall. */
export const E3Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 5 · NHỚ LẠI"
    accent="#9D7CD8"
    points={[
      { at: 30, text: "Bounded context là một vùng mà trong đó mỗi từ chỉ có một nghĩa." },
      { at: 80, text: "Repo này có ba kiểu cùng tên Variant, ở ba package, với ba hình dạng khác nhau." },
      { at: 130, text: "Gộp thành một định nghĩa chung nghe gọn, nhưng khiến mỗi thay đổi nhỏ phải xin phép bốn vùng." },
      { at: 180, text: "Chúng đồng bộ bằng event (catalog.variant_added), không import nhau. Ranh giới là package, và có test canh." },
    ]}
    asked={[
      "Bounded context là gì, và nó khác module ở chỗ nào?",
      "Hai context cùng nói về một thứ thì đồng bộ kiểu gì?",
      "Vì sao không nên có một mô hình chung cho cả hệ thống?",
    ]}
    nextLabel="MÙA 2 · TẬP 6"
    nextText="Domain Event: aggregate chỉ GHI lại việc đã xảy ra, tầng app mới là nơi phát đi."
  />
);
