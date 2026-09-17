import { loadFont as loadSans } from "@remotion/google-fonts/BeVietnamPro";
import { loadFont as loadMono } from "@remotion/google-fonts/JetBrainsMono";

/**
 * Be Vietnam Pro is drawn for Vietnamese diacritics — stacked marks on ế, ộ, ữ
 * stay legible at small sizes, which a latin-only face does not guarantee.
 */
export const SANS = loadSans("normal", {
  weights: ["400", "600", "800"],
  subsets: ["latin", "vietnamese"],
}).fontFamily;

/** Code is ASCII, so latin is enough. */
export const MONO = loadMono("normal", {
  weights: ["400", "700"],
  subsets: ["latin"],
}).fontFamily;
