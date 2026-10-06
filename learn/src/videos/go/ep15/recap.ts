import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "Struct tag là một chuỗi đứng sau kiểu của field. Compiler không đọc nó; encoding/json đọc bằng reflection để biết tên key." },
    { at: 80, text: `Repo có ${GO_USAGE.jsonTags} tag json trong code chạy thật và 0 trong internal/domain: domain không biết JSON là gì.` },
    { at: 130, text: `JSON ra ngoài qua hai cửa — outbox và HTTP — và cả hai đều là struct riêng: ${GO_USAGE.contractTypes} kiểu V1 trong contracts, các view trong adapter/http.` },
    { at: 180, text: "Đổi tên field domain không đổi contract; đổi contract là một kiểu V2. Đó là Published Language." },
  ],
  asked: [
    "Struct tag là gì, ai đọc nó, và nó khác annotation của PHP ở chỗ nào?",
    "Vì sao không json.Marshal thẳng struct của domain?",
    "Published Language là gì, và version trong tên kiểu để làm gì?",
  ],
  next: { label: "TẬP 16", text: "Vì sao toàn bộ test chạy dưới hai giây: port, adapter và cái fake thật trong RAM." },
} as const satisfies Recap;
