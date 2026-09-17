import type { EpisodeSpec } from "../../registry";
import { P2S1 } from "./scenes/P2S1";
import { P2S2 } from "./scenes/P2S2";
import { P2S3 } from "./scenes/P2S3";

/**
 * SEASON 3, EPISODE 2 — "Outbox: hai việc, một transaction"
 *
 * Learning objective (exactly one): the viewer can explain the gap in
 * "save then publish" and name the table that closes it.
 */
export const EP02 = {
  id: "Ep02-Outbox",
  title: "Outbox: hai việc, một transaction",
  scenePrefix: "Ep02-S",
  outDir: "mua3-portage/ep02-outbox",
  scenes: [
    { name: "1 · Khe hở", component: P2S1, durationInFrames: 360 },
    { name: "2 · Từ vựng: outbox", component: P2S2, durationInFrames: 520 },
    { name: "3 · Nhớ lại", component: P2S3, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
