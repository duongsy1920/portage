import type { EpisodeSpec } from "../../registry";
import { RECAP } from "./recap";
import { V1Hook } from "./scenes/V1Hook";
import { V2Fake } from "./scenes/V2Fake";
import { V3Httptest } from "./scenes/V3Httptest";
import { V4Table } from "./scenes/V4Table";
import { V5Guard } from "./scenes/V5Guard";
import { V6Compare } from "./scenes/V6Compare";
import { V7Recap } from "./scenes/V7Recap";

/**
 * SEASON 1, EPISODE 16 — "Vì sao toàn bộ test chạy dưới hai giây"
 *
 * Learning objective (exactly one): the viewer can explain the speed of the
 * suite from the architecture — ports with in-memory adapters, httptest,
 * tables — and can name the two things the repository does NOT have
 * (benchmarks, generated mocks) and say why.
 *
 * Independent of 11–15; the one to make if only one could be made
 * (PLAN-BO-SUNG §5), because it is what an interview asks most.
 */
export const GO_EP16 = {
  id: "Go-Ep16-Testing",
  title: "Vì sao toàn bộ test chạy dưới hai giây",
  scenePrefix: "Go-Ep16-S",
  outDir: "mua1-go/ep16-testing",
  recap: RECAP,
  anchors: ["GO_SNIPPETS.callHelper", "GO_SNIPPETS.guardAllowlist", "GO_SNIPPETS.tableTest", "GO_SNIPPETS.worldFake", "GO_USAGE.benchmarks", "GO_USAGE.httptestFiles", "GO_USAGE.subtests", "GO_USAGE.tableTests", "GO_USAGE.testFiles", "GO_USAGE.testFuncs"],
  scenes: [
    { name: "1 · Bốn con số", component: V1Hook, durationInFrames: 380 },
    { name: "2 · Từ vựng: fake", component: V2Fake, durationInFrames: 540 },
    { name: "3 · httptest: không mở cổng", component: V3Httptest, durationInFrames: 520 },
    { name: "4 · Bảng test và t.Run", component: V4Table, durationInFrames: 540 },
    { name: "5 · Luật sống trong test", component: V5Guard, durationInFrames: 520 },
    { name: "6 · PHPUnit và testing", component: V6Compare, durationInFrames: 500 },
    { name: "7 · Hết mùa 1", component: V7Recap, durationInFrames: 480 },
  ],
} as const satisfies EpisodeSpec;
