import type { EpisodeSpec } from "../../registry";
import { M1Hook } from "./scenes/M1Hook";
import { M2Promote } from "./scenes/M2Promote";
import { M3NotInherit } from "./scenes/M3NotInherit";
import { M4TypedID } from "./scenes/M4TypedID";
import { M5Recap } from "./scenes/M5Recap";

/**
 * SEASON 1, EPISODE 7 — "Embedding: giống kế thừa nhưng không phải"
 *
 * Learning objective (exactly one): the viewer can read a nameless field, knows
 * it forwards rather than inherits, and can explain why this repository has
 * nine typed ids instead of passing uuids around.
 */
export const GO_EP07 = {
  id: "Go-Ep07-Embedding",
  title: "Embedding: giống kế thừa nhưng không phải",
  scenePrefix: "Go-Ep07-S",
  outDir: "mua1-go/ep07-embedding",
  scenes: [
    { name: "1 · Field không có tên", component: M1Hook, durationInFrames: 330 },
    { name: "2 · Tìm method xuống một tầng", component: M2Promote, durationInFrames: 480 },
    { name: "3 · Từ vựng: embedding", component: M3NotInherit, durationInFrames: 520 },
    { name: "4 · Nhờ nó mà có typed ID", component: M4TypedID, durationInFrames: 540 },
    { name: "5 · Nhớ lại", component: M5Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
