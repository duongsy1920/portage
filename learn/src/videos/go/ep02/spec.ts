import type { EpisodeSpec } from "../../registry";
import { H1Hook } from "./scenes/H1Hook";
import { H2TypeAfter } from "./scenes/H2TypeAfter";
import { H3Case } from "./scenes/H3Case";
import { H4NoMethods } from "./scenes/H4NoMethods";
import { H5Literal } from "./scenes/H5Literal";
import { H6Hardest } from "./scenes/H6Hardest";
import { H7Table } from "./scenes/H7Table";
import { H8Recap } from "./scenes/H8Recap";

/**
 * SEASON 1, EPISODE 2 — "Đọc một dòng Go"
 *
 * Learning objective (exactly one): the viewer can read any declaration in the
 * repository out loud, without guessing — fields, function signatures, and the
 * visibility of every name in them.
 *
 * Why it is second: episode 1 needed almost no syntax, and every episode after
 * this one shows code. Nothing else can be taught while the reader is still
 * stuck on where the type went.
 *
 * All four rules are anchored in one file, shared/money.go, on purpose: a
 * single real file the viewer can open afterwards beats four tidy fragments
 * that exist nowhere.
 */
export const GO_EP02 = {
  id: "Go-Ep02-DocMotDong",
  title: "Đọc một dòng Go",
  scenePrefix: "Go-Ep02-S",
  outDir: "mua1-go/ep02-doc-mot-dong",
  scenes: [
    { name: "1 · Bốn dòng chưa đọc được", component: H1Hook, durationInFrames: 300 },
    { name: "2 · Tên trước, kiểu sau", component: H2TypeAfter, durationInFrames: 520 },
    { name: "3 · HOA và thường", component: H3Case, durationInFrames: 560 },
    { name: "4 · Struct không chứa hàm", component: H4NoMethods, durationInFrames: 520 },
    { name: "5 · Tạo mới, không có new", component: H5Literal, durationInFrames: 520 },
    { name: "6 · Dòng khó nhất file", component: H6Hardest, durationInFrames: 540 },
    { name: "7 · Bảng so sánh", component: H7Table, durationInFrames: 480 },
    { name: "8 · Nhớ lại", component: H8Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
