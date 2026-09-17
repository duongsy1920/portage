import type { EpisodeSpec } from "../../registry";
import { I1Hook } from "./scenes/I1Hook";
import { I2TwoValues } from "./scenes/I2TwoValues";
import { I3Sentinel } from "./scenes/I3Sentinel";
import { I4Wrap } from "./scenes/I4Wrap";
import { I5ErrorsIs } from "./scenes/I5ErrorsIs";
import { I6Table } from "./scenes/I6Table";
import { I7Recap } from "./scenes/I7Recap";

/**
 * SEASON 1, EPISODE 3 — "Lỗi là giá trị trả về"
 *
 * Learning objective (exactly one): the viewer can read, write and explain the
 * error path of any function in the repository — the pair return, the sentinel,
 * the wrap, and the check.
 *
 * Why third: it pays the debt episode 2 left at `(Money, error)`, and every
 * remaining episode shows code with `if err != nil` in it. Left unexplained,
 * that line reads as noise for the rest of the season.
 */
export const GO_EP03 = {
  id: "Go-Ep03-LoiLaGiaTri",
  title: "Lỗi là giá trị trả về",
  scenePrefix: "Go-Ep03-S",
  outDir: "mua1-go/ep03-loi-la-gia-tri",
  scenes: [
    { name: "1 · Không có try/catch", component: I1Hook, durationInFrames: 320 },
    { name: "2 · Trả về hai giá trị", component: I2TwoValues, durationInFrames: 520 },
    { name: "3 · Từ vựng: sentinel", component: I3Sentinel, durationInFrames: 540 },
    { name: "4 · %w đi qua ba tầng", component: I4Wrap, durationInFrames: 520 },
    { name: "5 · Từ vựng: errors.Is", component: I5ErrorsIs, durationInFrames: 540 },
    { name: "6 · Bảng so sánh", component: I6Table, durationInFrames: 480 },
    { name: "7 · Nhớ lại", component: I7Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
