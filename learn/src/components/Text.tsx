import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, TYPE } from "../design/tokens";
import { EASE } from "./Stage";

/**
 * Headline: the one thing the viewer should read first in a scene.
 *
 * It rises 24px while fading in. That small movement is what separates "a
 * slide appeared" from "an idea arrived" — but it is the ONLY movement, so it
 * never competes with the diagram below it.
 */
export const Headline: React.FC<{
  children: React.ReactNode;
  delay?: number;
  color?: string;
  size?: number;
}> = ({ children, delay = 0, color = COLOR.text, size = TYPE.headline }) => {
  const frame = useCurrentFrame();

  return (
    <Interactive.Div
      name="Headline"
      style={{
        fontSize: size,
        fontWeight: 800,
        lineHeight: 1.16,
        letterSpacing: -1,
        color,
        opacity: interpolate(frame, [delay, delay + 16], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
        translate: interpolate(frame, [delay, delay + 20], ["0px 24px", "0px 0px"], {
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

/** Supporting line under a headline. Never smaller than the 44px floor. */
export const Body: React.FC<{
  children: React.ReactNode;
  delay?: number;
  color?: string;
  size?: number;
}> = ({ children, delay = 0, color = COLOR.textDim, size = TYPE.body }) => {
  const frame = useCurrentFrame();

  return (
    <Interactive.Div
      name="Body"
      style={{
        fontSize: size,
        fontWeight: 400,
        lineHeight: 1.45,
        color,
        opacity: interpolate(frame, [delay, delay + 16], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
        translate: interpolate(frame, [delay, delay + 20], ["0px 16px", "0px 0px"], {
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

/**
 * A short statement the scene is built around — the "therefore" line.
 * It gets a left rule in an accent colour so it reads as a conclusion, not
 * as more prose.
 */
export const Callout: React.FC<{
  children: React.ReactNode;
  delay?: number;
  accent?: string;
}> = ({ children, delay = 0, accent = COLOR.go }) => {
  const frame = useCurrentFrame();

  return (
    <Interactive.Div
      name="Callout"
      style={{
        borderLeft: `6px solid ${accent}`,
        paddingLeft: 28,
        fontSize: TYPE.sub,
        fontWeight: 600,
        lineHeight: 1.32,
        color: COLOR.text,
        opacity: interpolate(frame, [delay, delay + 18], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
        translate: interpolate(frame, [delay, delay + 22], ["-14px 0px", "0px 0px"], {
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
