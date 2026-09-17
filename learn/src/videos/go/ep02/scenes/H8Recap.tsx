import React from "react";
import { Recap } from "../../../../components/Recap";

/**
 * Season 1, episode 2 — recall.
 *
 * Point three is the one worth the whole episode: package-level privacy is the
 * rule a Symfony developer will get wrong for months, because `private` looks
 * like a word they already know.
 */
export const H8Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 2 · NHỚ LẠI"
    points={[
      { at: 30, text: "Tên trước, kiểu sau — kể cả kiểu trả về, nằm ở cuối chữ ký hàm." },
      { at: 80, text: "Chữ HOA cho package khác dùng, chữ thường thì không. Không có từ khoá public." },
      { at: 130, text: "Riêng tư tính theo package, không theo struct. Muốn giấu kỹ hơn thì tách package." },
      { at: 180, text: "Struct chỉ giữ dữ liệu. Hàm viết ở ngoài, gắn vào kiểu bằng cụm trước tên hàm." },
    ]}
    asked={[
      "Vì sao Go viết kiểu sau tên biến?",
      "Chữ thường trong Go giấu được khỏi ai, và không giấu được khỏi ai?",
      "Go tạo object kiểu gì, khi không có từ khoá new?",
    ]}
    nextLabel="TẬP 3"
    nextText="Lỗi là giá trị trả về: vì sao Go không có try/catch, và (Money, error) nghĩa là gì?"
  />
);
