import type React from "react";
import { BEAT } from "../design/tokens";

/**
 * One place that knows every episode.
 *
 * Why this exists: Root.tsx used to repeat a fifteen-line registration block
 * per episode, and the still/overflow scripts each re-derived scene ids by
 * hand. With a two-episode series that is merely repetitive; with twenty-four
 * it is a source of drift. Adding an episode is now one entry here.
 */
export type Scene = {
  name: string;
  component: React.FC;
  durationInFrames: number;
};

export type RecapPoint = { at: number; text: string };

/**
 * What the closing screen says — four things to remember, three questions an
 * interview asks, what the next episode is about. It lives in the spec (as
 * recap.ts beside it) rather than inside the scene's JSX, so the learning hub
 * can quiz on it without parsing React. The scene reads the same object.
 */
export type Recap = {
  title?: string;
  points: readonly RecapPoint[];
  asked: readonly string[];
  /** Omitted on the last episode of the series. */
  next?: { label: string; text: string };
};

export type EpisodeSpec = {
  /** Composition id of the whole episode, e.g. "Go-Ep02-DocMotDong". */
  id: string;
  title: string;
  /** Scene compositions are registered as `${scenePrefix}1`, `${scenePrefix}2`… */
  scenePrefix: string;
  /** Where the rendered pair lands, relative to out/. */
  outDir: string;
  recap: Recap;
  /**
   * The pieces of src/data/portage.ts this episode puts on screen, as
   * "GO_SNIPPETS.mutex" / "GO_USAGE.defers". Generated from the scene files;
   * scripts/verify-hub.py recomputes and compares, so the list cannot go stale.
   */
  anchors: readonly string[];
  scenes: readonly Scene[];
};

/** Total length once the overlapping transitions are accounted for. */
export const totalFrames = (spec: EpisodeSpec): number =>
  spec.scenes.reduce((sum, s) => sum + s.durationInFrames, 0) -
  (spec.scenes.length - 1) * BEAT.transition;
