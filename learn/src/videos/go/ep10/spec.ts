import type { EpisodeSpec } from "../../registry";
import { P1Hook } from "./scenes/P1Hook";
import { P2Why } from "./scenes/P2Why";
import { P3Channel } from "./scenes/P3Channel";
import { P4Table } from "./scenes/P4Table";
import { P5Recap } from "./scenes/P5Recap";

/**
 * SEASON 1, EPISODE 10 — "Đồng thời: goroutine, mutex, channel"
 *
 * Learning objective (exactly one): the viewer can choose between a lock and a
 * channel from the shape of the problem, and can answer the channel questions
 * an interview will ask even though this repository uses none.
 *
 * The season's honesty rule is loudest here: what the repo does NOT use is said
 * out loud and then taught separately, labelled "ngoài Portage".
 */
export const GO_EP10 = {
  id: "Go-Ep10-DongThoi",
  title: "Đồng thời: goroutine, mutex, channel",
  scenePrefix: "Go-Ep10-S",
  outDir: "mua1-go/ep10-dong-thoi",
  scenes: [
    { name: "1 · Bao nhiêu channel?", component: P1Hook, durationInFrames: 360 },
    { name: "2 · Vì sao khoá", component: P2Why, durationInFrames: 540 },
    { name: "3 · Channel, ngoài Portage", component: P3Channel, durationInFrames: 520 },
    { name: "4 · Khi nào dùng cái nào", component: P4Table, durationInFrames: 480 },
    { name: "5 · Hết mùa 1", component: P5Recap, durationInFrames: 480 },
  ],
} as const satisfies EpisodeSpec;
