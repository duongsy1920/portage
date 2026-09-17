import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 2, episode 2 — recall. */
export const B3Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 2 · NHỚ LẠI"
    accent="#9D7CD8"
    points={[
      { at: 30, text: "Một câu hỏi tách xong: đổi nó bằng một cái y hệt, có ai nhận ra không?" },
      { at: 80, text: "Không ai nhận ra thì là Value Object — bằng nhau khi mọi field bằng nhau, và không có ID." },
      { at: 130, text: "Có người nhận ra thì là Entity — nó có danh tính riêng, sống qua nhiều lần thay đổi." },
      { at: 180, text: "Value Object trong Go bất biến gần như miễn phí, nhờ receiver giá trị ở tập 4 mùa 1." },
    ]}
    asked={[
      "Value Object khác Entity ở chỗ nào?",
      "Vì sao Value Object nên bất biến?",
      "Money có nên có ID không, và vì sao?",
    ]}
    nextLabel="MÙA 2 · TẬP 3"
    nextText="Aggregate và invariant: vì sao một transaction chỉ nên đụng vào một aggregate."
  />
);
