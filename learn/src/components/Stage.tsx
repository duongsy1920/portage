import React from "react";
import { AbsoluteFill, Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, SAFE, SPLIT, TYPE } from "../design/tokens";
import { SANS } from "../design/fonts";
import { EASE } from "../design/motion";

export { EASE };

/**
 * Stage is the frame every scene sits in: the background, the safe area, and
 * the two persistent labels (episode marker at the top, source path at the
 * bottom).
 *
 * Every scene uses it, so the viewer's eye never has to re-learn where things
 * are between scenes — only the middle changes.
 */
export const Stage: React.FC<{
  eyebrow?: string;
  /** File path shown at the bottom when the scene quotes real code. */
  source?: string;
  children: React.ReactNode;
}> = ({ eyebrow, source, children }) => {
  const frame = useCurrentFrame();

  return (
    <AbsoluteFill
      name="Stage"
      style={{ backgroundColor: COLOR.bg, fontFamily: SANS, color: COLOR.text }}
    >
      <AbsoluteFill
        name="Grid"
        style={{
          backgroundImage: `linear-gradient(${COLOR.bgLift} 1px, transparent 1px), linear-gradient(90deg, ${COLOR.bgLift} 1px, transparent 1px)`,
          backgroundSize: "64px 64px",
          opacity: 0.5,
        }}
      />

      {eyebrow ? (
        <Interactive.Div
          name="Eyebrow"
          style={{
            position: "absolute",
            top: SAFE.top - 44,
            left: SAFE.x,
            fontSize: TYPE.caption,
            letterSpacing: 5,
            fontWeight: 600,
            color: COLOR.textFaint,
            opacity: interpolate(frame, [0, 14], [0, 1], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            }),
          }}
        >
          {eyebrow}
        </Interactive.Div>
      ) : null}

      <AbsoluteFill
        name="Content"
        style={{
          paddingLeft: SAFE.x,
          paddingRight: SAFE.x,
          paddingTop: SAFE.top,
          paddingBottom: SAFE.bottom,
        }}
      >
        {children}
      </AbsoluteFill>

      {source ? (
        <Interactive.Div
          name="Source"
          style={{
            position: "absolute",
            bottom: SAFE.bottom - 46,
            left: SAFE.x,
            fontSize: TYPE.caption,
            color: COLOR.textFaint,
            opacity: interpolate(frame, [12, 30], [0, 1], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            }),
          }}
        >
          {source}
        </Interactive.Div>
      ) : null}
    </AbsoluteFill>
  );
};

/**
 * The layout most scenes use: the picture on the left, the words about it on
 * the right. Keeping the split fixed across scenes means the eye already knows
 * where the explanation lives before it arrives.
 */
export const Split: React.FC<{
  left: React.ReactNode;
  right: React.ReactNode;
  /** Vertical alignment of the two columns. */
  align?: "center" | "flex-start";
  /**
   * Width of the left column. The right column takes the rest. A vocabulary
   * scene wants a wide right side (the words are the content); a flow scene
   * wants a wide left side (the picture is).
   */
  leftWidth?: number;
  gap?: number;
}> = ({ left, right, align = "center", leftWidth = SPLIT.left, gap = SPLIT.gap }) => {
  const rightWidth = SAFE.contentWidth - leftWidth - gap;
  return (
    <div style={{ display: "flex", gap, alignItems: align, height: "100%" }}>
      <div style={{ width: leftWidth, flexShrink: 0 }}>{left}</div>
      <div style={{ width: rightWidth, flexShrink: 0 }}>{right}</div>
    </div>
  );
};
