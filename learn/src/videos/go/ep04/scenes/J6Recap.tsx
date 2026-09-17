import React from "react";
import { Recap } from "../../../../components/Recap";
import { GO_USAGE } from "../../../../data/portage";

/** Season 1, episode 4 — recall. */
export const J6Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 4 · NHỚ LẠI"
    points={[
      { at: 30, text: "Cụm trước tên hàm là receiver — nó là $this của PHP, nhưng do bạn đặt tên." },
      { at: 80, text: "Không có dấu sao thì Go chép cả struct, nên method không sửa được bản gốc." },
      { at: 130, text: "Có dấu sao thì method nhận địa chỉ bản gốc, sửa vào đó là sửa thật." },
      { at: 180, text: `Repo có ${GO_USAGE.valueReceivers} receiver giá trị và ${GO_USAGE.pointerReceivers} receiver con trỏ — value object nhóm đầu, thứ có trạng thái nhóm sau.` },
    ]}
    asked={[
      "Khi nào dùng receiver con trỏ, khi nào dùng receiver giá trị?",
      "Vì sao một method gán vào field mà không đổi được gì?",
      "Receiver giá trị có làm chậm không, khi struct lớn?",
    ]}
    nextLabel="TẬP 5"
    nextText="Interface ngầm: vì sao Go không có từ khoá implements, và vì sao port khai ở nơi CẦN?"
  />
);
