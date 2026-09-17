import type { EpisodeSpec } from "../../registry";
import { B1Hook } from "./scenes/B1Hook";
import { B2VO } from "./scenes/B2VO";
import { B3Recap } from "./scenes/B3Recap";

/**
 * SEASON 2, EPISODE 2 — "Value Object và Entity"
 *
 * Learning objective (exactly one): the viewer can sort any type in their own
 * domain into the two buckets using one question, and knows why Go makes the
 * value-object half almost free.
 */
export const DDD_EP02 = {
  id: "Ddd-Ep02-ValueObject",
  title: "Value Object và Entity",
  scenePrefix: "Ddd-Ep02-S",
  outDir: "mua2-ddd/ep02-value-object",
  scenes: [
    { name: "1 · Một câu hỏi tách xong", component: B1Hook, durationInFrames: 360 },
    { name: "2 · Từ vựng: value object", component: B2VO, durationInFrames: 520 },
    { name: "3 · Nhớ lại", component: B3Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
