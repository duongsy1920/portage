import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 3, episode 7 — recall. */
export const P7S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 7 · NHỚ LẠI"
    points={[
      { at: 30, text: "Một màn hình của khách cần dữ liệu từ bốn vùng, mà bốn vùng không được import lẫn nhau." },
      { at: 80, text: "Nên không ai đi hỏi ai. Mỗi vùng kể lại việc đã xảy ra, và một projector nghe rồi ghi." },
      { at: 130, text: "Kết quả là một dòng rộng trong order_summaries: màn hình đọc đúng một câu SELECT." },
      { at: 180, text: "Cái giá là dòng đó chậm hơn sự thật vài trăm mili giây — chính là cái 404 của tập 3." },
    ]}
    asked={[
      "Vì sao không JOIN giữa các bounded context?",
      "Read model được dựng lại từ đâu nếu nó hỏng?",
      "CQRS giải quyết vấn đề gì, và tốn gì?",
    ]}
    nextLabel="MÙA 3 · TẬP 8"
    nextText="Anti-corruption layer: máy được tạo nháp, nhưng không được xác nhận."
  />
);
