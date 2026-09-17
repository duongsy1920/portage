import type { EpisodeSpec } from "../../registry";
import { P6S1 } from "./scenes/P6S1";
import { P6S2 } from "./scenes/P6S2";

/**
 * SEASON 3, EPISODE 6 — "Báo giá là một bức ảnh, và đối soát"
 *
 * Learning objective (exactly one): the viewer can say why a quote is stored
 * rather than recomputed, and what the variance column is for.
 */
export const EP06 = {
  id: "Ep06-BaoGia",
  title: "Báo giá là một bức ảnh, và đối soát",
  scenePrefix: "Ep06-S",
  outDir: "mua3-portage/ep06-bao-gia",
  scenes: [
    { name: "1 · Ảnh chụp, không phải phép tính", component: P6S1, durationInFrames: 400 },
    { name: "2 · Nhớ lại", component: P6S2, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
