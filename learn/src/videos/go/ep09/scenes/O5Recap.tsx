import React from "react";
import { Recap } from "../../../../components/Recap";
import { GO_USAGE } from "../../../../data/portage";

/** Season 1, episode 9 — recall. */
export const O5Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 9 · NHỚ LẠI"
    points={[
      { at: 30, text: `defer hoãn một lệnh tới lúc hàm thoát — thoát kiểu gì cũng chạy. Repo có ${GO_USAGE.defers} lệnh.` },
      { at: 80, text: "Nhờ vậy khoá không bao giờ kẹt: chỉ có một chỗ mở khoá, không phải nhớ ở từng đường return." },
      { at: 130, text: "panic không phải exception. Vì tiến trình phục vụ mọi request, panic không chặn là chết cả hệ." },
      { at: 180, text: "recover chỉ chạy trong defer, và ở đây nó rollback rồi panic lại — dọn dẹp, không đi tiếp." },
    ]}
    asked={[
      "defer chạy vào lúc nào, và theo thứ tự nào khi có nhiều cái?",
      "Vì sao panic ở Go nguy hiểm hơn fatal error ở PHP?",
      "Khi nào được dùng recover?",
    ]}
    nextLabel="TẬP 10"
    nextText="Đồng thời: 3 goroutine, 25 mutex, và 0 channel — vì sao repo không dùng channel, và phỏng vấn vẫn hỏi."
  />
);
