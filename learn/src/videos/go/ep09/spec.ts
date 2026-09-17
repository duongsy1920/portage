import type { EpisodeSpec } from "../../registry";
import { O1Hook } from "./scenes/O1Hook";
import { O2When } from "./scenes/O2When";
import { O3Panic } from "./scenes/O3Panic";
import { O4Recover } from "./scenes/O4Recover";
import { O5Recap } from "./scenes/O5Recap";

/**
 * SEASON 1, EPISODE 9 — "defer, panic, recover"
 *
 * Learning objective (exactly one): the viewer can say exactly when a deferred
 * call runs, and why a panic in this codebase is a production incident rather
 * than a failed request.
 */
export const GO_EP09 = {
  id: "Go-Ep09-DeferPanic",
  title: "defer, panic, recover",
  scenePrefix: "Go-Ep09-S",
  outDir: "mua1-go/ep09-defer-panic",
  scenes: [
    { name: "1 · Hai dòng đi cùng nhau", component: O1Hook, durationInFrames: 330 },
    { name: "2 · Thoát kiểu gì cũng chạy", component: O2When, durationInFrames: 520 },
    { name: "3 · Từ vựng: panic", component: O3Panic, durationInFrames: 540 },
    { name: "4 · recover để dọn dẹp", component: O4Recover, durationInFrames: 540 },
    { name: "5 · Nhớ lại", component: O5Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
