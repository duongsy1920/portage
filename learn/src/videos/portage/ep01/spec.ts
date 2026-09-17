import type { EpisodeSpec } from "../../registry";
import { S1Hook } from "./scenes/S1Hook";
import { S2Transaction } from "./scenes/S2Transaction";
import { S3Journey } from "./scenes/S3Journey";
import { S4Thirty } from "./scenes/S4Thirty";
import { S5Break } from "./scenes/S5Break";
import { S6Aggregate } from "./scenes/S6Aggregate";
import { S7TwoStructs } from "./scenes/S7TwoStructs";
import { S8Context } from "./scenes/S8Context";
import { S9Guard } from "./scenes/S9Guard";
import { S10Map } from "./scenes/S10Map";
import { S11Recap } from "./scenes/S11Recap";

/**
 * EPISODE 1 — "Vì sao project Go này bị chia làm năm?"
 *
 * Learning objective (exactly one): the viewer can explain why CustomerOrder
 * and PurchaseTask cannot live in one table, using the business as the
 * argument rather than DDD vocabulary.
 *
 * Audience: has never written Go and never studied DDD, but has six years of
 * PHP/Symfony. So every term gets a TermCard BEFORE it is used, and every term
 * is anchored to the Symfony thing it maps onto. Three terms are introduced in
 * this episode and no more: transaction, aggregate, bounded context.
 *
 * The episode is DATA: a list of scenes and how long each is held. Adding an
 * episode means adding a file like this one, not writing animation code.
 */
export const EP01 = {
  id: "Ep01-NamVung",
  title: "Vì sao project Go này bị chia làm năm?",
  scenePrefix: "Ep01-S",
  outDir: "mua3-portage/ep01-nam-vung",
  scenes: [
    { name: "1 · Câu hỏi", component: S1Hook, durationInFrames: 290 },
    { name: "2 · Từ vựng: transaction", component: S2Transaction, durationInFrames: 470 },
    { name: "3 · Đi qua năm vùng", component: S3Journey, durationInFrames: 620 },
    { name: "4 · Ba mươi ngày", component: S4Thirty, durationInFrames: 500 },
    { name: "5 · Chỗ vỡ", component: S5Break, durationInFrames: 520 },
    { name: "6 · Từ vựng: aggregate", component: S6Aggregate, durationInFrames: 480 },
    { name: "7 · Hai struct thật", component: S7TwoStructs, durationInFrames: 470 },
    { name: "8 · Từ vựng: bounded context", component: S8Context, durationInFrames: 520 },
    { name: "9 · Máy canh ranh giới", component: S9Guard, durationInFrames: 440 },
    { name: "10 · Bản đồ", component: S10Map, durationInFrames: 400 },
    { name: "11 · Nhớ lại", component: S11Recap, durationInFrames: 400 },
  ],
} as const satisfies EpisodeSpec;
