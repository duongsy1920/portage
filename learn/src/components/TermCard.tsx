import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, RADIUS, TYPE } from "../design/tokens";
import { MONO } from "../design/fonts";
import { EASE } from "../design/motion";

/**
 * A "new word" screen.
 *
 * The viewer has never studied Go or DDD, so every term gets introduced before
 * it is used, in a fixed shape: the word, what it means in plain Vietnamese,
 * the thing they ALREADY know that it maps to in Symfony, and where it lives
 * in this repository.
 *
 * The Symfony row is the load-bearing one. A definition lands on an empty
 * shelf; an equivalence lands on six years of existing experience.
 */
export const TermCard: React.FC<{
  /** The term itself, e.g. "Aggregate". */
  term: string;
  /** Plain-language meaning. No jargon allowed in this string. */
  plain: string;
  /** The Symfony/Doctrine/PHP thing this maps onto. */
  symfony?: string;
  /** Where it lives in the Go repository. */
  inRepo?: string;
  delay?: number;
  accent?: string;
}> = ({ term, plain, symfony, inRepo, delay = 0, accent = COLOR.go }) => {
  const frame = useCurrentFrame();

  const row = (
    label: string,
    value: string,
    at: number,
    labelColor: string,
    mono = false,
  ) => (
    <Interactive.Div
      name={label}
      style={{
        display: "flex",
        gap: 20,
        alignItems: "baseline",
        opacity: interpolate(frame, [at, at + 16], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
        translate: interpolate(frame, [at, at + 22], ["0px 14px", "0px 0px"], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
      }}
    >
      <span
        style={{
          fontSize: TYPE.caption,
          fontWeight: 700,
          letterSpacing: 1.5,
          color: labelColor,
          width: 158,
          flexShrink: 0,
          textAlign: "right",
        }}
      >
        {label}
      </span>
      <span
        style={{
          fontSize: mono ? TYPE.label : TYPE.body,
          fontFamily: mono ? MONO : undefined,
          lineHeight: 1.4,
          color: COLOR.text,
        }}
      >
        {value}
      </span>
    </Interactive.Div>
  );

  return (
    <Interactive.Div
      name={`Term: ${term}`}
      style={{
        padding: "44px 48px",
        borderRadius: RADIUS.lg,
        backgroundColor: COLOR.surface,
        border: `1px solid ${COLOR.line}`,
        borderLeft: `8px solid ${accent}`,
        display: "flex",
        flexDirection: "column",
        gap: 26,
        opacity: interpolate(frame, [delay, delay + 14], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
      }}
    >
      <Interactive.Div
        name="Term word"
        style={{
          display: "flex",
          alignItems: "baseline",
          gap: 18,
          opacity: interpolate(frame, [delay + 4, delay + 20], [0, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: EASE,
          }),
        }}
      >
        <span
          style={{
            fontSize: TYPE.caption,
            letterSpacing: 4,
            fontWeight: 700,
            color: COLOR.textFaint,
          }}
        >
          TỪ MỚI
        </span>
        <span
          style={{
            fontFamily: MONO,
            fontSize: TYPE.sub,
            fontWeight: 700,
            color: accent,
          }}
        >
          {term}
        </span>
      </Interactive.Div>

      {row("NGHĨA ĐEN", plain, delay + 22, COLOR.textFaint)}
      {symfony
        ? row("BÊN SYMFONY", symfony, delay + 48, COLOR.php)
        : null}
      {inRepo ? row("Ở ĐÂY", inRepo, delay + 74, accent, true) : null}
    </Interactive.Div>
  );
};
