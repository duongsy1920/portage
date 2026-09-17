import type { EpisodeSpec } from "../../registry";
import { J1Hook } from "./scenes/J1Hook";
import { J2Copy } from "./scenes/J2Copy";
import { J3Pointer } from "./scenes/J3Pointer";
import { J4Trap } from "./scenes/J4Trap";
import { J5Table } from "./scenes/J5Table";
import { J6Recap } from "./scenes/J6Recap";

/**
 * SEASON 1, EPISODE 4 — "Receiver: (m Money) và (e *Events)"
 *
 * Learning objective (exactly one): the viewer can decide, for any method they
 * write, whether the receiver should be a value or a pointer — and can explain
 * why the repository's value objects are immutable without any effort.
 *
 * Why fourth: episodes 2 and 3 both pointed at `(m Money)` and said "later".
 * This is later. It also closes season one's first arc — after this the viewer
 * can read every declaration and every method signature in the repository.
 */
export const GO_EP04 = {
  id: "Go-Ep04-Receiver",
  title: "Receiver: (m Money) và (e *Events)",
  scenePrefix: "Go-Ep04-S",
  outDir: "mua1-go/ep04-receiver",
  scenes: [
    { name: "1 · Một dấu sao", component: J1Hook, durationInFrames: 330 },
    { name: "2 · Go chép cả struct", component: J2Copy, durationInFrames: 540 },
    { name: "3 · Từ vựng: pointer receiver", component: J3Pointer, durationInFrames: 520 },
    { name: "4 · Cái bẫy", component: J4Trap, durationInFrames: 520 },
    { name: "5 · Bảng so sánh", component: J5Table, durationInFrames: 460 },
    { name: "6 · Nhớ lại", component: J6Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
