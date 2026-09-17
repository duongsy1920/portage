import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, RADIUS, TYPE } from "../design/tokens";
import { MONO } from "../design/fonts";
import { EASE } from "../design/motion";

export type CompareRow = {
  /** What the two columns are being compared on. */
  about: string;
  php: string;
  go: string;
  at: number;
};

/**
 * The PHP / Go comparison table — the spine of season one.
 *
 * A learner with six years of Symfony does not need Go explained from zero;
 * they need the DELTA. Two columns with the same row label make the delta the
 * thing on screen, and rows arrive one at a time so each difference lands on
 * its own instead of as a wall.
 */
export const CompareTable: React.FC<{
  rows: readonly CompareRow[];
  width: number;
  leftTitle?: string;
  rightTitle?: string;
}> = ({ rows, width, leftTitle = "PHP / Symfony", rightTitle = "Go" }) => {
  const frame = useCurrentFrame();
  const aboutW = 300;
  const colW = (width - aboutW - 40) / 2;

  const head = (text: string, color: string) => (
    <div
      style={{
        width: colW,
        fontFamily: MONO,
        fontSize: TYPE.caption,
        fontWeight: 700,
        letterSpacing: 2,
        color,
        paddingBottom: 14,
      }}
    >
      {text.toUpperCase()}
    </div>
  );

  return (
    <Interactive.Div name="CompareTable" style={{ width }}>
      <div style={{ display: "flex", gap: 20, borderBottom: `2px solid ${COLOR.line}` }}>
        <div style={{ width: aboutW }} />
        {head(leftTitle, COLOR.php)}
        {head(rightTitle, COLOR.go)}
      </div>

      {rows.map((r, i) => (
        <Interactive.Div
          key={i}
          name={r.about}
          style={{
            display: "flex",
            gap: 20,
            alignItems: "flex-start",
            padding: "18px 0",
            borderBottom: `1px solid ${COLOR.line}`,
            opacity: interpolate(frame, [r.at, r.at + 18], [0, 1], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            }),
            translate: interpolate(frame, [r.at, r.at + 24], ["0px 14px", "0px 0px"], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            }),
          }}
        >
          <div
            style={{
              width: aboutW,
              fontSize: TYPE.label,
              color: COLOR.textDim,
              lineHeight: 1.3,
            }}
          >
            {r.about}
          </div>
          <div
            style={{
              width: colW,
              fontSize: TYPE.label,
              color: COLOR.text,
              lineHeight: 1.35,
            }}
          >
            {r.php}
          </div>
          <div
            style={{
              width: colW,
              fontSize: TYPE.label,
              color: COLOR.text,
              lineHeight: 1.35,
              fontWeight: 600,
            }}
          >
            {r.go}
          </div>
        </Interactive.Div>
      ))}
    </Interactive.Div>
  );
};

/** A single big statement, used to close a comparison. */
export const Punchline: React.FC<{ children: React.ReactNode; at: number }> = ({
  children,
  at,
}) => {
  const frame = useCurrentFrame();
  return (
    <Interactive.Div
      name="Punchline"
      style={{
        marginTop: 30,
        padding: "26px 34px",
        borderRadius: RADIUS.md,
        backgroundColor: COLOR.surface,
        borderLeft: `8px solid ${COLOR.go}`,
        fontSize: TYPE.sub,
        fontWeight: 600,
        lineHeight: 1.35,
        opacity: interpolate(frame, [at, at + 20], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
      }}
    >
      {children}
    </Interactive.Div>
  );
};
