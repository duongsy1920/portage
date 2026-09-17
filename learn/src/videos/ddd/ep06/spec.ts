import type { EpisodeSpec } from "../../registry";
import { F1Hook } from "./scenes/F1Hook";
import { F2Term } from "./scenes/F2Term";
import { F3Recap } from "./scenes/F3Recap";

/**
 * SEASON 2, EPISODE 6 — "Domain Event"
 *
 * Learning objective (exactly one): the viewer can say why an aggregate records
 * rather than publishes, and what would go wrong on the day a transaction rolls
 * back if it did.
 *
 * This closes season 2 and hands straight to season 3, where the same events
 * are watched travelling through a real order.
 */
export const DDD_EP06 = {
  id: "Ddd-Ep06-DomainEvent",
  title: "Domain Event",
  scenePrefix: "Ddd-Ep06-S",
  outDir: "mua2-ddd/ep06-domain-event",
  scenes: [
    { name: "1 · Ghi lại, không phát đi", component: F1Hook, durationInFrames: 360 },
    { name: "2 · Từ vựng: domain event", component: F2Term, durationInFrames: 520 },
    { name: "3 · Hết mùa 2", component: F3Recap, durationInFrames: 480 },
  ],
} as const satisfies EpisodeSpec;
