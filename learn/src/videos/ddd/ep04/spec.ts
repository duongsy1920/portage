import type { EpisodeSpec } from "../../registry";
import { D1Hook } from "./scenes/D1Hook";
import { D2Guard } from "./scenes/D2Guard";
import { D3Recap } from "./scenes/D3Recap";

/**
 * SEASON 2, EPISODE 4 — "Repository là cổng"
 *
 * Learning objective (exactly one): the viewer can explain why the domain
 * declares the repository interface rather than implementing it, and can name
 * the mechanism that keeps that true after the first busy week.
 */
export const DDD_EP04 = {
  id: "Ddd-Ep04-Repository",
  title: "Repository là cổng",
  scenePrefix: "Ddd-Ep04-S",
  outDir: "mua2-ddd/ep04-repository",
  scenes: [
    { name: "1 · Domain khai, không cài", component: D1Hook, durationInFrames: 340 },
    { name: "2 · Từ vựng: port", component: D2Guard, durationInFrames: 540 },
    { name: "3 · Nhớ lại", component: D3Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
