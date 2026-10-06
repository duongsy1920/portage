import type { EpisodeSpec } from "../../registry";
import { RECAP } from "./recap";
import { S1Hook } from "./scenes/S1Hook";
import { S2On } from "./scenes/S2On";
import { S3Into } from "./scenes/S3Into";
import { S4When } from "./scenes/S4When";
import { S5Table } from "./scenes/S5Table";
import { S6Recap } from "./scenes/S6Recap";

/**
 * SEASON 1, EPISODE 13 — "Generics: hai chỗ duy nhất trong repo"
 *
 * Learning objective (exactly one): the viewer can read `func on[T any](…)`,
 * say what T is and where it comes from at each call, and answer the
 * interview question "generics or interface?" with the repository's own two
 * uses as the argument.
 *
 * Depends on 11 (on takes a closure) and 12 (the assertion inside it).
 */
export const GO_EP13 = {
  id: "Go-Ep13-Generics",
  title: "Generics: hai chỗ duy nhất trong repo",
  scenePrefix: "Go-Ep13-S",
  outDir: "mua1-go/ep13-generics",
  recap: RECAP,
  anchors: ["GO_SNIPPETS.decoders", "GO_SNIPPETS.intoGeneric", "GO_SNIPPETS.onGeneric", "GO_USAGE.generics"],
  scenes: [
    { name: "1 · Đúng hai hàm", component: S1Hook, durationInFrames: 360 },
    { name: "2 · on[T]: closure có kiểu vào, handler không kiểu ra", component: S2On, durationInFrames: 560 },
    { name: "3 · into[T]: ghi rõ T", component: S3Into, durationInFrames: 540 },
    { name: "4 · Khi nào generic, khi nào interface", component: S4When, durationInFrames: 520 },
    { name: "5 · Bảng so sánh", component: S5Table, durationInFrames: 480 },
    { name: "6 · Nhớ lại", component: S6Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
