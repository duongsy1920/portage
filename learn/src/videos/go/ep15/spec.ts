import type { EpisodeSpec } from "../../registry";
import { RECAP } from "./recap";
import { U1Hook } from "./scenes/U1Hook";
import { U2Tag } from "./scenes/U2Tag";
import { U3Published } from "./scenes/U3Published";
import { U4Exit } from "./scenes/U4Exit";
import { U5Table } from "./scenes/U5Table";
import { U6Recap } from "./scenes/U6Recap";

/**
 * SEASON 1, EPISODE 15 — "struct tag, JSON, và đường ra khỏi domain"
 *
 * Learning objective (exactly one): the viewer can read a `json:"…"` tag,
 * say who reads it (not the compiler), and explain why the domain has none
 * while the contracts and the HTTP views have hundreds — the Published
 * Language, by its mechanism.
 */
export const GO_EP15 = {
  id: "Go-Ep15-StructTag",
  title: "struct tag, JSON, và đường ra khỏi domain",
  scenePrefix: "Go-Ep15-S",
  outDir: "mua1-go/ep15-struct-tag",
  recap: RECAP,
  anchors: ["GO_SNIPPETS.contractV1", "GO_SNIPPETS.domainEvent", "GO_SNIPPETS.moneyView", "GO_SNIPPETS.projectorReads", "GO_USAGE.contractTypes", "GO_USAGE.jsonFiles", "GO_USAGE.jsonTags"],
  scenes: [
    { name: "1 · Cùng một sự kiện, hai struct", component: U1Hook, durationInFrames: 380 },
    { name: "2 · Đọc một tag", component: U2Tag, durationInFrames: 520 },
    { name: "3 · Từ vựng: Published Language", component: U3Published, durationInFrames: 540 },
    { name: "4 · Hai cửa ra, một cửa vào", component: U4Exit, durationInFrames: 540 },
    { name: "5 · Bảng so sánh", component: U5Table, durationInFrames: 480 },
    { name: "6 · Nhớ lại", component: U6Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
