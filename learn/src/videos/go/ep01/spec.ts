import type { EpisodeSpec } from "../../registry";
import { G1Hook } from "./scenes/G1Hook";
import { G2Php } from "./scenes/G2Php";
import { G3Go } from "./scenes/G3Go";
import { G4Goroutine } from "./scenes/G4Goroutine";
import { G5Race } from "./scenes/G5Race";
import { G6Context } from "./scenes/G6Context";
import { G7Table } from "./scenes/G7Table";
import { G8Recap } from "./scenes/G8Recap";

/**
 * SEASON 1, EPISODE 1 — "PHP chết sau mỗi request. Go thì không."
 *
 * Learning objective (exactly one): the viewer can explain how a Go process
 * lives compared to PHP-FPM, and can name three things in Go code that exist
 * BECAUSE of that difference.
 *
 * Why this is episode one of the whole roadmap: it needs almost no syntax, it
 * is the largest single surprise for a Symfony developer, and it makes the
 * next six episodes feel inevitable rather than arbitrary — pools, locks, ctx
 * and a hand-wired main() are all consequences of it.
 */
export const GO_EP01 = {
  id: "Go-Ep01-VongDoi",
  title: "PHP chết sau mỗi request. Go thì không.",
  scenePrefix: "Go-Ep01-S",
  outDir: "mua1-go/ep01-vong-doi",
  scenes: [
    { name: "1 · Lời hứa", component: G1Hook, durationInFrames: 300 },
    { name: "2 · PHP-FPM: một request một đời", component: G2Php, durationInFrames: 560 },
    { name: "3 · Go: dựng một lần", component: G3Go, durationInFrames: 620 },
    { name: "4 · Từ vựng: goroutine", component: G4Goroutine, durationInFrames: 500 },
    { name: "5 · Cái giá: biến chung", component: G5Race, durationInFrames: 520 },
    { name: "6 · Từ vựng: context", component: G6Context, durationInFrames: 520 },
    { name: "7 · Bảng so sánh", component: G7Table, durationInFrames: 480 },
    { name: "8 · Nhớ lại", component: G8Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
