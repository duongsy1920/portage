import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 3, episode 2 — recall. */
export const P2S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 2 · NHỚ LẠI"
    points={[
      { at: 30, text: "“Lưu xong rồi publish” hỏng ở khe hở giữa hai bước: tiến trình chết là event mất luôn." },
      { at: 80, text: "Outbox là một bảng trong chính database đó. Event ghi vào nó cùng transaction với dữ liệu." },
      { at: 130, text: "Nhờ transaction đi trong ctx, INSERT vào outbox dùng đúng kết nối mà đơn hàng đang dùng." },
      { at: 180, text: "Bài toán đổi từ “hai hệ thống phải cùng thành công” thành “một database phải commit”." },
    ]}
    asked={[
      "Outbox pattern giải quyết vấn đề gì?",
      "Vì sao không publish thẳng sau khi lưu?",
      "Làm sao bảo đảm INSERT vào outbox cùng transaction với dữ liệu?",
    ]}
    nextLabel="MÙA 3 · TẬP 3"
    nextText="Nhất quán sau cùng: vì sao 404 ngay sau khi tạo là thiết kế, không phải bug."
  />
);
