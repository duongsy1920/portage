import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 3, episode 3 — recall. */
export const P3S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 3 · NHỚ LẠI"
    points={[
      { at: 30, text: "Bảng đọc chỉ đầy sau khi relay chạy, nên 404 ngay sau khi tạo là chuyện bình thường." },
      { at: 80, text: "404 ở đây nghĩa là “chưa tới”, không phải “không có”. Khách nên thử lại, không nên sửa request." },
      { at: 130, text: "Repo có một bảng duy nhất biến tên lỗi thành mã HTTP: 92 dòng, ba mã cho ba câu chuyện." },
      { at: 180, text: "Không nằm trong bảng thì là 500 — nghĩa là bug của mình, và chỉ ghi log, không kể cho khách." },
    ]}
    asked={[
      "Nhất quán sau cùng là gì, và nó buộc API phải đổi thế nào?",
      "Khi nào trả 404 và khi nào trả 409?",
      "Client nên xử lý 404 tạm thời kiểu gì?",
    ]}
    nextLabel="MÙA 3 · TẬP 4"
    nextText="Port và Adapter: vì sao 278 test chạy dưới 2 giây mà không cần Docker."
  />
);
