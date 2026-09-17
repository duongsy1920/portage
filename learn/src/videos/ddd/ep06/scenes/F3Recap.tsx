import React from "react";
import { Recap } from "../../../../components/Recap";
import { REPO } from "../../../../data/portage";

/** Season 2, episode 6 — recall, and the end of the season. */
export const F3Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 6 · HẾT MÙA 2"
    accent="#9D7CD8"
    title="Bốn câu mang theo, và hết mùa hai"
    points={[
      { at: 30, text: "Domain event là một việc đã xảy ra, đặt tên ở thì quá khứ — không ai từ chối được nữa." },
      { at: 80, text: "Aggregate chỉ ghi event vào sổ. Tầng app lấy ra và ghi vào outbox, trong cùng transaction." },
      { at: 130, text: "Phát thẳng từ aggregate thì một thay đổi bị rollback vẫn kịp bắn event ra ngoài." },
      { at: 180, text: `Repo có ${REPO.events} loại event và ${REPO.subscribeLines} dòng đăng ký nghe. Mùa 3 xem chúng chạy thật.` },
    ]}
    asked={[
      "Domain event khác message của queue ở chỗ nào?",
      "Vì sao aggregate không tự publish event?",
      "Vì sao tên event luôn ở thì quá khứ?",
    ]}
    nextLabel="MÙA 3 · TẬP 1"
    nextText="Portage: vì sao project thật này bị chia làm năm vùng."
  />
);
