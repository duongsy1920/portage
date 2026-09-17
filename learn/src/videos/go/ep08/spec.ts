import type { EpisodeSpec } from "../../registry";
import { N1Hook } from "./scenes/N1Hook";
import { N2Why } from "./scenes/N2Why";
import { N3MapOrder } from "./scenes/N3MapOrder";
import { N4Compare } from "./scenes/N4Compare";
import { N5Recap } from "./scenes/N5Recap";

/**
 * SEASON 1, EPISODE 8 — "slice, map và ba cái bẫy"
 *
 * Learning objective (exactly one): the viewer stops writing the three bugs
 * that PHP's single `array` type makes almost inevitable, and can say why each
 * one is a consequence of something already learned rather than a rule.
 */
export const GO_EP08 = {
  id: "Go-Ep08-SliceMap",
  title: "slice, map và ba cái bẫy",
  scenePrefix: "Go-Ep08-S",
  outDir: "mua1-go/ep08-slice-map",
  scenes: [
    { name: "1 · Dòng mất dữ liệu", component: N1Hook, durationInFrames: 360 },
    { name: "2 · Vì sao phải gán lại", component: N2Why, durationInFrames: 500 },
    { name: "3 · Map duyệt ngẫu nhiên", component: N3MapOrder, durationInFrames: 540 },
    { name: "4 · Một array, ba thứ", component: N4Compare, durationInFrames: 500 },
    { name: "5 · Nhớ lại", component: N5Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
