import type { EpisodeSpec } from "../../registry";
import { C1Hook } from "./scenes/C1Hook";
import { C2Invariant } from "./scenes/C2Invariant";
import { C3Recap } from "./scenes/C3Recap";

/**
 * SEASON 2, EPISODE 3 — "Aggregate và invariant"
 *
 * Learning objective (exactly one): the viewer can decide where an aggregate
 * boundary goes by asking which facts must be true at the same instant — and
 * can say why that answer also decides the shape of their transactions.
 */
export const DDD_EP03 = {
  id: "Ddd-Ep03-Aggregate",
  title: "Aggregate và invariant",
  scenePrefix: "Ddd-Ep03-S",
  outDir: "mua2-ddd/ep03-aggregate",
  scenes: [
    { name: "1 · Một cụm, một cửa", component: C1Hook, durationInFrames: 380 },
    { name: "2 · Từ vựng: invariant", component: C2Invariant, durationInFrames: 520 },
    { name: "3 · Nhớ lại", component: C3Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
