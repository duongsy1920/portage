import type { EpisodeSpec } from "../../registry";
import { P7S1 } from "./scenes/P7S1";
import { P7S2 } from "./scenes/P7S2";
import { P7S3 } from "./scenes/P7S3";

/**
 * SEASON 3, EPISODE 7 — "Màn hình cần bốn context: đừng JOIN"
 *
 * Learning objective (exactly one): the viewer can build a screen that spans
 * several contexts without querying across them, and can state the cost of the
 * read model in plain words.
 */
export const EP07 = {
  id: "Ep07-ReadModel",
  title: "Màn hình cần bốn context: đừng JOIN",
  scenePrefix: "Ep07-S",
  outDir: "mua3-portage/ep07-read-model",
  scenes: [
    { name: "1 · Bốn vùng, không JOIN", component: P7S1, durationInFrames: 400 },
    { name: "2 · Nghe rồi tự ghi", component: P7S2, durationInFrames: 420 },
    { name: "3 · Nhớ lại", component: P7S3, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
