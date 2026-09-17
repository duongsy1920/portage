import type { EpisodeSpec } from "../../registry";
import { K1Hook } from "./scenes/K1Hook";
import { K2Duck } from "./scenes/K2Duck";
import { K3Assert } from "./scenes/K3Assert";
import { K4Small } from "./scenes/K4Small";
import { K5Table } from "./scenes/K5Table";
import { K6Recap } from "./scenes/K6Recap";

/**
 * SEASON 1, EPISODE 5 — "Interface ngầm: Go không có implements"
 *
 * Learning objective (exactly one): the viewer can explain why the dependency
 * arrows in this repository point inward, and can name the one line that buys
 * back the compile-time check the missing keyword took away.
 *
 * Why fifth: it is the last piece of Go needed to read the architecture, and
 * it is the piece that makes season 3 possible — ports and adapters stop being
 * a pattern to memorise once you see they fall out of the language.
 */
export const GO_EP05 = {
  id: "Go-Ep05-InterfaceNgam",
  title: "Interface ngầm: Go không có implements",
  scenePrefix: "Go-Ep05-S",
  outDir: "mua1-go/ep05-interface-ngam",
  scenes: [
    { name: "1 · Không có implements", component: K1Hook, durationInFrames: 330 },
    { name: "2 · Mũi tên chạy vào trong", component: K2Duck, durationInFrames: 520 },
    { name: "3 · Từ vựng: var _ X", component: K3Assert, durationInFrames: 520 },
    { name: "4 · Interface nhỏ", component: K4Small, durationInFrames: 540 },
    { name: "5 · Bảng so sánh", component: K5Table, durationInFrames: 460 },
    { name: "6 · Nhớ lại", component: K6Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
