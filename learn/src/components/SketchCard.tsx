import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, RADIUS, TYPE } from "../design/tokens";
import { MONO, SANS } from "../design/fonts";
import { EASE } from "./Stage";

/**
 * A design the project REJECTED, drawn so it can never be mistaken for the
 * project's own code: dashed border, no file path, and a label saying so.
 *
 * Teaching why an architecture exists needs the wrong version on screen. The
 * risk is that a learner screenshots the wrong version and remembers it as
 * the codebase — hence the loud disclaimer rather than a subtle one.
 */
export const SketchCard: React.FC<{
  /** What this sketch is, e.g. "cách làm quen thuộc". */
  label: string;
  lines: readonly string[];
  delay?: number;
  lineStagger?: number;
}> = ({ label, lines, delay = 0, lineStagger = 3 }) => {
  const frame = useCurrentFrame();

  return (
    <Interactive.Div
      name="SketchCard"
      style={{
        backgroundColor: "rgba(255,107,107,0.05)",
        border: `2px dashed ${COLOR.danger}`,
        borderRadius: RADIUS.md,
        overflow: "hidden",
        opacity: interpolate(frame, [delay, delay + 14], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
      }}
    >
      <div
        style={{
          padding: "14px 24px",
          borderBottom: `1px dashed ${COLOR.danger}`,
          fontFamily: SANS,
          fontSize: TYPE.caption,
          fontWeight: 700,
          letterSpacing: 2,
          color: COLOR.danger,
        }}
      >
        ✕ {label.toUpperCase()} — KHÔNG PHẢI CODE CỦA PORTAGE
      </div>

      <div style={{ padding: "22px 24px" }}>
        {lines.map((line: string, i: number) => (
          <Interactive.Div
            key={i}
            name={`Sketch line ${i + 1}`}
            style={{
              fontFamily: MONO,
              fontSize: TYPE.code,
              lineHeight: 1.6,
              whiteSpace: "pre",
              color: COLOR.textDim,
              opacity: interpolate(
                frame,
                [delay + 8 + i * lineStagger, delay + 20 + i * lineStagger],
                [0, line.trim() === "" ? 0 : 1],
                {
                  extrapolateLeft: "clamp",
                  extrapolateRight: "clamp",
                  easing: EASE,
                },
              ),
            }}
          >
            {line === "" ? " " : line}
          </Interactive.Div>
        ))}
      </div>
    </Interactive.Div>
  );
};
