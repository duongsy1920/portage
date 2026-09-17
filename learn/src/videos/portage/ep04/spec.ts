import type { EpisodeSpec } from "../../registry";
import { P4S1 } from "./scenes/P4S1";
import { P4S2 } from "./scenes/P4S2";
import { P4S3 } from "./scenes/P4S3";

/**
 * SEASON 3, EPISODE 4 — "Port & Adapter, và luật chiều phụ thuộc"
 *
 * Learning objective (exactly one): the viewer can state the dependency rule
 * and point at the measurable consequence of following it in this repository.
 */
export const EP04 = {
  id: "Ep04-PortAdapter",
  title: "Port & Adapter, và luật chiều phụ thuộc",
  scenePrefix: "Ep04-S",
  outDir: "mua3-portage/ep04-port-adapter",
  scenes: [
    { name: "1 · Con số đo được", component: P4S1, durationInFrames: 360 },
    { name: "2 · Mũi tên chạy vào trong", component: P4S2, durationInFrames: 500 },
    { name: "3 · Nhớ lại", component: P4S3, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
