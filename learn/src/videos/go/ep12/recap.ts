import { GO_USAGE } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: "any là interface không đòi method nào, nên nhận mọi giá trị — mixed của PHP, và tốn cái giá như mixed." },
    { at: 80, text: `shared.Event chỉ đòi hai method, nên ${GO_USAGE.typeSwitchCases} loại event là Event. Mã hoá ở một chỗ: một type switch.` },
    { at: 130, text: "Trong mỗi case, e đã có kiểu cụ thể. Thiếu một case là một test đỏ, không phải event mất tích." },
    { at: 180, text: "x.(T) sai kiểu: không ok là panic, có ok là false. Repo chỉ dùng dạng có ok." },
  ],
  asked: [
    "Interface rỗng là gì, và vì sao dùng nó là mất kiểm tra lúc biên dịch?",
    "Type switch khác chuỗi instanceof ở chỗ nào?",
    "Type assertion x.(T) khi sai kiểu thì xảy ra gì, và làm sao tránh panic?",
  ],
  next: { label: "TẬP 13", text: "Generics: hai chỗ duy nhất trong repo, và vì sao không có generics thì phải viết 35 adapter giống hệt nhau." },
} as const satisfies Recap;
