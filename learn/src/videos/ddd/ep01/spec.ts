import type { EpisodeSpec } from "../../registry";
import { A1Hook } from "./scenes/A1Hook";
import { A2Where } from "./scenes/A2Where";
import { A3Rich } from "./scenes/A3Rich";
import { A4Recap } from "./scenes/A4Recap";

/**
 * SEASON 2, EPISODE 1 — "Mô hình thiếu máu"
 *
 * Learning objective (exactly one): the viewer can point at the place a
 * business rule leaked out of their own Symfony code, and can name the move
 * that pulls it back.
 *
 * Season 2 opens here rather than with vocabulary because the viewer has
 * already lived this problem for six years. Starting from their experience
 * makes every later term — aggregate, invariant, bounded context — an answer
 * to something instead of a definition to memorise.
 */
export const DDD_EP01 = {
  id: "Ddd-Ep01-ThieuMau",
  title: "Mô hình thiếu máu",
  scenePrefix: "Ddd-Ep01-S",
  outDir: "mua2-ddd/ep01-thieu-mau",
  scenes: [
    { name: "1 · Luật nằm ở đâu", component: A1Hook, durationInFrames: 340 },
    { name: "2 · Nó rơi thành nhiều bản", component: A2Where, durationInFrames: 480 },
    { name: "3 · Từ vựng: mô hình thiếu máu", component: A3Rich, durationInFrames: 520 },
    { name: "4 · Nhớ lại", component: A4Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
