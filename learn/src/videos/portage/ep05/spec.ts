import type { EpisodeSpec } from "../../registry";
import { P5S1 } from "./scenes/P5S1";
import { P5S2 } from "./scenes/P5S2";
import { P5S3 } from "./scenes/P5S3";

/**
 * SEASON 3, EPISODE 5 — "Điểm không thể quay đầu"
 *
 * Learning objective (exactly one): the viewer can find the moment in their own
 * domain after which an action cannot be undone, and knows that the rule
 * protecting it belongs inside the aggregate.
 */
export const EP05 = {
  id: "Ep05-KhongQuayDau",
  title: "Điểm không thể quay đầu",
  scenePrefix: "Ep05-S",
  outDir: "mua3-portage/ep05-khong-quay-dau",
  scenes: [
    { name: "1 · Trên trục thời gian", component: P5S1, durationInFrames: 380 },
    { name: "2 · Luật viết ra được", component: P5S2, durationInFrames: 520 },
    { name: "3 · Nhớ lại", component: P5S3, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
