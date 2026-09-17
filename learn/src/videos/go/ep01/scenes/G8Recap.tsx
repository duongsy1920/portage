import React from "react";
import { Recap } from "../../../../components/Recap";

/** Season 1, episode 1 — recall, plus the questions an interviewer asks. */
export const G8Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 1 · NHỚ LẠI"
    points={[
      { at: 30, text: "PHP dựng lại thế giới mỗi request. Go dựng một lần rồi phục vụ mọi request trên đó." },
      { at: 80, text: "Nên mỗi request là một goroutine, và chúng dùng chung mọi thứ main() đã dựng." },
      { at: 130, text: "Dùng chung thì phải khoá: đó là lý do có sync.Mutex ở 25 chỗ trong repo." },
      { at: 180, text: "Sống lâu thì phải huỷ được từng việc: đó là lý do ctx có mặt ở 361 chỗ." },
    ]}
    asked={[
      "Goroutine khác thread ở chỗ nào?",
      "go test -race dò ra cái gì?",
      "context.Context dùng để làm gì?",
    ]}
    nextLabel="TẬP 2"
    nextText="Đọc một dòng Go: vì sao kiểu viết sau tên, và class đi đâu mất?"
  />
);
