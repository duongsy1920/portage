import type { EpisodeSpec } from "../../registry";
import { E1Hook } from "./scenes/E1Hook";
import { E2Why } from "./scenes/E2Why";
import { E3Recap } from "./scenes/E3Recap";

/**
 * SEASON 2, EPISODE 5 — "Bounded Context"
 *
 * Learning objective (exactly one): the viewer stops trying to build one shared
 * model of a word and can name the cost of doing so, using three real types
 * from this repository that all happen to be called Variant.
 */
export const DDD_EP05 = {
  id: "Ddd-Ep05-BoundedContext",
  title: "Bounded Context",
  scenePrefix: "Ddd-Ep05-S",
  outDir: "mua2-ddd/ep05-bounded-context",
  scenes: [
    { name: "1 · Một chữ, ba nghĩa", component: E1Hook, durationInFrames: 380 },
    { name: "2 · Từ vựng: bounded context", component: E2Why, durationInFrames: 520 },
    { name: "3 · Nhớ lại", component: E3Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
