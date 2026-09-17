import React from "react";
import { Recap } from "../../../../components/Recap";
import { GO_USAGE } from "../../../../data/portage";

/** Season 1, episode 10 — recall, and the end of the season. */
export const P5Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 10 · HẾT MÙA 1"
    title="Bốn câu mang theo, và hết mùa một"
    points={[
      { at: 30, text: `Repo có ${GO_USAGE.goroutines} goroutine, ${GO_USAGE.mutex} chỗ khoá, và ${GO_USAGE.channels} channel — nói thẳng vì đó là sự thật.` },
      { at: 80, text: "Dữ liệu dùng chung thì dùng khoá. Chuyển giao việc thì dùng channel. Chọn theo hình dạng bài toán." },
      { at: 130, text: "“Đừng chia sẻ bộ nhớ, hãy giao tiếp” là lời khuyên, không phải luật cấm." },
      { at: 180, text: "Giờ bạn đọc được mọi file trong repo. Mùa 2 nói về DDD, bằng chính ví dụ Symfony đã quen." },
    ]}
    asked={[
      "Khi nào dùng channel, khi nào dùng mutex?",
      "Channel không có bộ đệm khác channel có bộ đệm ở chỗ nào?",
      "Deadlock trong Go xảy ra kiểu gì?",
    ]}
    nextLabel="MÙA 2 · TẬP 1"
    nextText="Mô hình thiếu máu: entity toàn getter setter, và luật nghiệp vụ rơi mất ở đâu."
  />
);
