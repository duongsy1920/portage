/**
 * Design tokens — the visual vocabulary shared by every episode.
 *
 * Changing a value here changes every video at once. That is the point:
 * a learner should recognise "purple box = ordering" without being told,
 * and that only works if the mapping never drifts between episodes.
 *
 * Format: 1920x1080 landscape. This is a LESSON, watched on a laptop next to
 * the open repository, not a Reel watched on a phone. Landscape buys room for
 * a diagram and its explanation side by side, which is the whole point of
 * teaching a flow: the picture and the words about it must be visible at once.
 */

export const FPS = 30;
export const WIDTH = 1920;
export const HEIGHT = 1080;

/** Safe area, scaled from the 1080-wide guidance to this frame. */
export const SAFE = {
  x: 110,
  top: 76,
  bottom: 76,
  get contentWidth() {
    return WIDTH - this.x * 2; // 1700
  },
  get contentHeight() {
    return HEIGHT - this.top - this.bottom; // 928
  },
} as const;

/** Two-column split used by most scenes: picture left, words right. */
export const SPLIT = {
  left: 1000,
  gap: 64,
  get right() {
    return SAFE.contentWidth - this.left - this.gap; // 636
  },
} as const;

export const COLOR = {
  bg: "#080B10",
  bgLift: "#0E141C",
  surface: "#141C26",
  surfaceHi: "#1B2632",
  line: "#243243",
  lineHi: "#35485E",

  text: "#E8EFF6",
  textDim: "#93A4B8",
  textFaint: "#5C6E85",

  /** Go's own brand cyan — this is a Go project, the accent should say so. */
  go: "#00ADD8",
  goDim: "#0B7E9E",

  money: "#F2B33D",
  danger: "#FF6B6B",
  ok: "#4ADE80",
  /** Used only for the "this is PHP/Symfony" side of a comparison. */
  php: "#B39DFF",
} as const;

/**
 * One colour per bounded context, fixed for the whole series.
 * These are the five directories under internal/domain (plus the read side).
 */
export const CONTEXT = {
  catalog: { label: "CATALOG", color: "#7AA2F7" },
  pricing: { label: "PRICING", color: "#F2B33D" },
  ordering: { label: "ORDERING", color: "#9D7CD8" },
  procurement: { label: "PROCUREMENT", color: "#E06C75" },
  logistics: { label: "LOGISTICS", color: "#4ADE80" },
  reporting: { label: "REPORTING", color: "#7D8FA6" },
} as const;

export type ContextKey = keyof typeof CONTEXT;

/**
 * Type scale for a 1920-wide frame.
 *
 * Smaller relative to width than the Reels guidance, on purpose: this video
 * carries detailed explanation, and detailed explanation at Reels type size
 * either does not fit or turns into forty scenes of three words each.
 */
export const TYPE = {
  hero: 92,
  headline: 68,
  sub: 44,
  body: 36,
  label: 28,
  code: 27,
  caption: 22,
} as const;

export const RADIUS = { sm: 10, md: 16, lg: 24 } as const;

/** Shared animation rhythm, in frames at 30fps. */
export const BEAT = {
  in: 14,
  out: 10,
  stagger: 8,
  transition: 14,
  /** How long a packet takes to travel one edge of a flow diagram. */
  travel: 26,
} as const;
