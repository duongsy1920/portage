import type { EpisodeSpec } from "../../registry";
import { RECAP } from "./recap";
import { R1Hook } from "./scenes/R1Hook";
import { R2Any } from "./scenes/R2Any";
import { R3Switch } from "./scenes/R3Switch";
import { R4Assert } from "./scenes/R4Assert";
import { R5Table } from "./scenes/R5Table";
import { R6Recap } from "./scenes/R6Recap";

/**
 * SEASON 1, EPISODE 12 — "interface rỗng và type switch"
 *
 * Learning objective (exactly one): the viewer can read `switch e := ev.(type)`
 * and `m, ok := msg.(T)`, say what each costs (compile-time checking), and
 * say why the repository pays that cost in exactly one place — the codec.
 */
export const GO_EP12 = {
  id: "Go-Ep12-TypeSwitch",
  title: "interface rỗng và type switch",
  scenePrefix: "Go-Ep12-S",
  outDir: "mua1-go/ep12-type-switch",
  recap: RECAP,
  anchors: ["GO_SNIPPETS.assertOk", "GO_SNIPPETS.decodeAny", "GO_SNIPPETS.eventInterface", "GO_SNIPPETS.payloadMap", "GO_SNIPPETS.typeSwitch", "GO_USAGE.anyUses", "GO_USAGE.typeAssertOk", "GO_USAGE.typeSwitchCases"],
  scenes: [
    { name: "1 · Hai method là đủ", component: R1Hook, durationInFrames: 360 },
    { name: "2 · Từ vựng: any", component: R2Any, durationInFrames: 520 },
    { name: "3 · Một type switch, 37 case", component: R3Switch, durationInFrames: 540 },
    { name: "4 · Hỏi lại kiểu bằng ok", component: R4Assert, durationInFrames: 520 },
    { name: "5 · Bảng so sánh", component: R5Table, durationInFrames: 480 },
    { name: "6 · Nhớ lại", component: R6Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
