import type { EpisodeSpec } from "../../registry";
import { RECAP } from "./recap";
import { T1Hook } from "./scenes/T1Hook";
import { T2Select } from "./scenes/T2Select";
import { T3Race } from "./scenes/T3Race";
import { T4Buffered } from "./scenes/T4Buffered";
import { T5Outside } from "./scenes/T5Outside";
import { T6Recap } from "./scenes/T6Recap";

/**
 * SEASON 1, EPISODE 14 — "select, ticker, và dừng một việc đang chạy"
 *
 * Learning objective (exactly one): the viewer can read the select loop at
 * the heart of the worker, explain how a running job learns it must stop,
 * and answer the buffered/unbuffered and WaitGroup questions an interview
 * asks — with the repository's own test as the example.
 *
 * This is where episode 10's "4 reads from channels" is finally read.
 */
export const GO_EP14 = {
  id: "Go-Ep14-Select",
  title: "select, ticker, và dừng một việc đang chạy",
  scenePrefix: "Go-Ep14-S",
  outDir: "mua1-go/ep14-select",
  recap: RECAP,
  anchors: ["GO_SNIPPETS.bufferedDone", "GO_SNIPPETS.raceStart", "GO_SNIPPETS.selectLoop", "GO_USAGE.channelsRead", "GO_USAGE.chansInTests", "GO_USAGE.selects", "GO_USAGE.waitGroupsInTests"],
  scenes: [
    { name: "1 · Bốn lần đọc channel, đây là chúng", component: T1Hook, durationInFrames: 400 },
    { name: "2 · Từ vựng: select", component: T2Select, durationInFrames: 520 },
    { name: "3 · WaitGroup và súng lệnh", component: T3Race, durationInFrames: 560 },
    { name: "4 · Đệm hay không đệm", component: T4Buffered, durationInFrames: 520 },
    { name: "5 · Ngoài Portage: fan-out, fan-in", component: T5Outside, durationInFrames: 500 },
    { name: "6 · Nhớ lại", component: T6Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
