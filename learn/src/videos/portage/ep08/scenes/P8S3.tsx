import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 3, episode 8 — recall, and the end of the series. */
export const P8S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 8 · HẾT LOẠT"
    title="Bốn câu mang theo, và hết loạt"
    points={[
      { at: 30, text: "Điều khoản của Nike và phần lớn shop cấm cào trang, nên hệ thống không tự đọc trang shop." },
      { at: 80, text: "Dữ liệu vào bằng tay người, hoặc qua một adapter đọc URL mà người dùng tự đưa." },
      { at: 130, text: "Hai anti-corruption layer dịch thế giới bên ngoài sang từ vựng của domain — và dừng ở bản nháp." },
      { at: 180, text: "Xác nhận là lúc có người chịu trách nhiệm. Máy đoán sai một cỡ giày là mua nhầm hàng, tiền thật." },
    ]}
    asked={[
      "Anti-corruption layer để làm gì?",
      "Vì sao không để máy tự xác nhận sản phẩm?",
      "Ràng buộc pháp lý ảnh hưởng tới kiến trúc kiểu gì?",
    ]}
  />
);
