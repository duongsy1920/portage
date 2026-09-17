import React from "react";
import { Recap } from "../../../../components/Recap";
import { GOLDEN } from "../../../../data/portage";

/** Season 3, episode 6 — recall. */
export const P6S2: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 6 · NHỚ LẠI"
    points={[
      { at: 30, text: "Báo giá là ảnh chụp của thế giới lúc nó được lập: tỷ giá, cước, cân đo đều chốt lại ở đó." },
      { at: 80, text: "Thực tế thì đổi. Nên bảng reconciliations giữ cả hai nửa: đã hứa bao nhiêu, đã trả bao nhiêu." },
      { at: 130, text: `Chênh lệch của đơn mẫu là ${GOLDEN.varianceUSD} USD — cước hãng bay thu 27.50 trong khi báo giá 25.00.` },
      { at: 180, text: "Variance âm liên tục qua nhiều đơn nghĩa là bảng giá cước hoặc phép đo đang sai, không phải xui." },
    ]}
    asked={[
      "Vì sao báo giá phải lưu lại thay vì tính lại mỗi lần?",
      "Đối soát quote với actual trả lời câu hỏi kinh doanh nào?",
      "Hạn 48 giờ của báo giá để làm gì?",
    ]}
    nextLabel="MÙA 3 · TẬP 7"
    nextText="Màn hình cần bốn context: vì sao không JOIN, và bảng đọc dựng bằng event."
  />
);
