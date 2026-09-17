import type { EpisodeSpec } from "../../registry";
import { L1Hook } from "./scenes/L1Hook";
import { L2Table } from "./scenes/L2Table";
import { L3Guard } from "./scenes/L3Guard";
import { L4Safe } from "./scenes/L4Safe";
import { L5Recap } from "./scenes/L5Recap";

/**
 * SEASON 1, EPISODE 6 — "Zero value: Go không có null"
 *
 * Learning objective (exactly one): the viewer stops writing null checks that
 * Go does not need, and understands why every constructor in this repository
 * either panics or returns an error rather than trusting its input.
 */
export const GO_EP06 = {
  id: "Go-Ep06-ZeroValue",
  title: "Zero value: Go không có null",
  scenePrefix: "Go-Ep06-S",
  outDir: "mua1-go/ep06-zero-value",
  scenes: [
    { name: "1 · Dòng lẽ ra phải nổ", component: L1Hook, durationInFrames: 360 },
    { name: "2 · Bảng zero value", component: L2Table, durationInFrames: 500 },
    { name: "3 · Từ vựng: IsZero", component: L3Guard, durationInFrames: 520 },
    { name: "4 · Zero value phải an toàn", component: L4Safe, durationInFrames: 460 },
    { name: "5 · Nhớ lại", component: L5Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
