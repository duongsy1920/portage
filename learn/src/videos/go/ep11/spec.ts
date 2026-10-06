import type { EpisodeSpec } from "../../registry";
import { RECAP } from "./recap";
import { Q1Hook } from "./scenes/Q1Hook";
import { Q2Why } from "./scenes/Q2Why";
import { Q3Closure } from "./scenes/Q3Closure";
import { Q4Named } from "./scenes/Q4Named";
import { Q5Table } from "./scenes/Q5Table";
import { Q6Recap } from "./scenes/Q6Recap";

/**
 * SEASON 1, EPISODE 11 — "Closure: hàm nhận hàm làm tham số"
 *
 * Learning objective (exactly one): the viewer can read `InTx(ctx, func(ctx)
 * error { … })` — a function handed to a function — and explain why a result
 * escapes it through a captured variable instead of a return value.
 *
 * Why first of the supplement (PLAN-BO-SUNG §5): every use case in the
 * repository has this shape, and episodes 12–13 build on it — on[T] is a
 * generic that takes a closure and returns one.
 */
export const GO_EP11 = {
  id: "Go-Ep11-Closure",
  title: "Closure: hàm nhận hàm làm tham số",
  scenePrefix: "Go-Ep11-S",
  outDir: "mua1-go/ep11-closure",
  recap: RECAP,
  anchors: ["GO_SNIPPETS.closureResult", "GO_SNIPPETS.inTxNamed", "GO_SNIPPETS.inTxPort", "GO_SNIPPETS.recoverTx", "GO_USAGE.funcLiterals", "GO_USAGE.inTxCalls", "GO_USAGE.namedReturns"],
  scenes: [
    { name: "1 · Tham số thứ hai là một hàm", component: Q1Hook, durationInFrames: 360 },
    { name: "2 · Vì sao nhận hàm, không trả transaction", component: Q2Why, durationInFrames: 520 },
    { name: "3 · Từ vựng: closure", component: Q3Closure, durationInFrames: 540 },
    { name: "4 · Trả về có tên, defer ghi đè", component: Q4Named, durationInFrames: 520 },
    { name: "5 · Bảng so sánh", component: Q5Table, durationInFrames: 480 },
    { name: "6 · Nhớ lại", component: Q6Recap, durationInFrames: 460 },
  ],
} as const satisfies EpisodeSpec;
