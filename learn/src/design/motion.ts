import { Easing } from "remotion";

/**
 * The series easing. One curve everywhere: quick out, long settle.
 * Defined once so a scene never invents its own feel.
 */
export const EASE = Easing.bezier(0.16, 1, 0.3, 1);

/** Linear travel for things moving along a wire, where ease-out looks wrong. */
export const EASE_TRAVEL = Easing.bezier(0.4, 0, 0.2, 1);
