import React from "react";
import { Recap } from "../../../../components/Recap";
import { REPO } from "../../../../data/portage";

/** Season 3, episode 4 — recall. */
export const P4S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 4 · NHỚ LẠI"
    points={[
      { at: 30, text: "Port là interface do phía cần khai báo. Adapter là bản cài đặt, nằm ở vòng ngoài." },
      { at: 80, text: "Mọi mũi tên phụ thuộc chạy vào trong. Domain không biết Postgres, HTTP hay OpenAI tồn tại." },
      { at: 130, text: `Nhờ vậy ${REPO.testsNoDocker} test chạy dưới hai giây mà không cần Docker — đó là con số đo được của kiến trúc.` },
      { at: 180, text: "Và có 7 bài test viết bằng go/ast canh luật đó, nên nó không phụ thuộc vào trí nhớ ai cả." },
    ]}
    asked={[
      "Port khác adapter ở chỗ nào?",
      "Vì sao domain không được import package của adapter?",
      "Làm sao ép được luật kiến trúc trong một codebase Go?",
    ]}
    nextLabel="MÙA 3 · TẬP 5"
    nextText="Điểm không thể quay đầu: luật nghiệp vụ có tiền thật của khách đằng sau."
  />
);
