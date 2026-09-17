import { buildEpisode } from "./episode";
import type { EpisodeSpec } from "./registry";

import { GO_EP01 } from "./go/ep01/spec";
import { GO_EP02 } from "./go/ep02/spec";
import { GO_EP03 } from "./go/ep03/spec";
import { GO_EP04 } from "./go/ep04/spec";
import { GO_EP05 } from "./go/ep05/spec";
import { GO_EP06 } from "./go/ep06/spec";
import { GO_EP07 } from "./go/ep07/spec";
import { GO_EP08 } from "./go/ep08/spec";
import { GO_EP09 } from "./go/ep09/spec";
import { GO_EP10 } from "./go/ep10/spec";
import { DDD_EP01 } from "./ddd/ep01/spec";
import { DDD_EP02 } from "./ddd/ep02/spec";
import { DDD_EP03 } from "./ddd/ep03/spec";
import { DDD_EP04 } from "./ddd/ep04/spec";
import { DDD_EP05 } from "./ddd/ep05/spec";
import { DDD_EP06 } from "./ddd/ep06/spec";
import { EP01 } from "./portage/ep01/spec";
import { EP02 } from "./portage/ep02/spec";
import { EP03 } from "./portage/ep03/spec";
import { EP04 } from "./portage/ep04/spec";
import { EP05 } from "./portage/ep05/spec";
import { EP06 } from "./portage/ep06/spec";
import { EP07 } from "./portage/ep07/spec";
import { EP08 } from "./portage/ep08/spec";

export type Season = { folder: string; specs: readonly EpisodeSpec[] };

/**
 * The whole series, in teaching order.
 *
 * Season 1 is listed first because it comes first in the roadmap: the Portage
 * episodes show Go code, and Go code is unreadable until season 1 is done.
 * Adding an episode is one import and one line — the composition itself is
 * built from the spec by buildEpisode().
 */
export const SEASONS: readonly Season[] = [
  {
    folder: "Mua1-Go",
    specs: [GO_EP01, GO_EP02, GO_EP03, GO_EP04, GO_EP05, GO_EP06, GO_EP07, GO_EP08, GO_EP09, GO_EP10],
  },
  {
    folder: "Mua2-DDD",
    specs: [DDD_EP01, DDD_EP02, DDD_EP03, DDD_EP04, DDD_EP05, DDD_EP06],
  },
  {
    folder: "Mua3-Portage",
    specs: [EP01, EP02, EP03, EP04, EP05, EP06, EP07, EP08],
  },
];

/** Episode component, built once per spec. */
export const componentFor = buildEpisode;
