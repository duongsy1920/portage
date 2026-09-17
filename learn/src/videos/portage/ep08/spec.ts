import type { EpisodeSpec } from "../../registry";
import { P8S1 } from "./scenes/P8S1";
import { P8S2 } from "./scenes/P8S2";
import { P8S3 } from "./scenes/P8S3";

/**
 * SEASON 3, EPISODE 8 — "Máy được tạo nháp, không được xác nhận"
 *
 * Learning objective (exactly one): the viewer can name what an
 * anti-corruption layer protects, and can explain why a legal constraint ended
 * up shaping the architecture rather than sitting in a policy document.
 *
 * This closes the series.
 */
export const EP08 = {
  id: "Ep08-ACL",
  title: "Máy được tạo nháp, không được xác nhận",
  scenePrefix: "Ep08-S",
  outDir: "mua3-portage/ep08-acl",
  scenes: [
    { name: "1 · Không được tự đọc trang shop", component: P8S1, durationInFrames: 380 },
    { name: "2 · Từ vựng: anti-corruption layer", component: P8S2, durationInFrames: 560 },
    { name: "3 · Hết loạt", component: P8S3, durationInFrames: 480 },
  ],
} as const satisfies EpisodeSpec;
