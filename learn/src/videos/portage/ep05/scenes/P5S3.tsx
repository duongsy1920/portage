import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 3, episode 5 — recall. */
export const P5S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 5 · NHỚ LẠI"
    points={[
      { at: 30, text: "Trước lúc mua ở shop Mỹ, huỷ đơn là hoàn cọc. Sau đó thì tiền đã đi và hàng đang bay." },
      { at: 80, text: "Luật đó là ba dòng ở đầu method: sai trạng thái thì trả error và không đổi gì cả." },
      { at: 130, text: "Mỗi lý do từ chối một tên lỗi riêng, nên khách nhận 409 kèm lý do đọc được, không phải 500." },
      { at: 180, text: "Không có setter, không có UPDATE thẳng. Một cửa duy nhất, và cửa đó kiểm luật trước." },
    ]}
    asked={[
      "Điểm không thể quay đầu trong nghiệp vụ của bạn nằm ở đâu?",
      "Vì sao luật đó phải nằm trong aggregate chứ không ở tầng service?",
      "Aggregate từ chối thì tầng HTTP biết trả mã nào?",
    ]}
    nextLabel="MÙA 3 · TẬP 6"
    nextText="Báo giá là một bức ảnh, và đối soát: con số variance trả lời câu hỏi kinh doanh."
  />
);
