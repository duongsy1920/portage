import type { EpisodeSpec } from "../../registry";
import { P3S1 } from "./scenes/P3S1";
import { P3S2 } from "./scenes/P3S2";
import { P3S3 } from "./scenes/P3S3";

/**
 * SEASON 3, EPISODE 3 — "Nhất quán sau cùng có mã trạng thái"
 *
 * Learning objective (exactly one): the viewer stops reading a 404 right after
 * a successful create as a bug, and can say what the client should do instead.
 */
export const EP03 = {
  id: "Ep03-NhatQuan",
  title: "Nhất quán sau cùng có mã trạng thái",
  scenePrefix: "Ep03-S",
  outDir: "mua3-portage/ep03-nhat-quan",
  scenes: [
    { name: "1 · Tạo xong, đọc lại 404", component: P3S1, durationInFrames: 400 },
    { name: "2 · Ba mã, ba câu chuyện", component: P3S2, durationInFrames: 480 },
    { name: "3 · Nhớ lại", component: P3S3, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
