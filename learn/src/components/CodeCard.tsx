import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, RADIUS, TYPE } from "../design/tokens";
import { MONO, SANS } from "../design/fonts";
import { EASE } from "./Stage";
import type { Snippet } from "../data/portage";

/**
 * Real code from the Go repository, with the file and line it came from.
 *
 * The header is not decoration. A learner who cannot find the code on disk
 * will not go looking for it, and a video that shows code without an address
 * teaches the code instead of teaching the codebase.
 *
 * Lines appear one after another so the eye reads in order instead of
 * scanning a wall of text.
 */
export const CodeCard: React.FC<{
  snippet: Snippet;
  delay?: number;
  /** Zero-based line indexes to tint with the accent colour. */
  highlight?: number[];
  accent?: string;
  /** Frames between consecutive lines appearing. 0 shows the block at once. */
  lineStagger?: number;
  scale?: number;
}> = ({
  snippet,
  delay = 0,
  highlight = [],
  accent = COLOR.go,
  lineStagger = 3,
  scale = 1,
}) => {
  const frame = useCurrentFrame();

  return (
    <Interactive.Div
      name="CodeCard"
      style={{
        backgroundColor: COLOR.surface,
        border: `1px solid ${COLOR.line}`,
        borderRadius: RADIUS.md,
        overflow: "hidden",
        opacity: interpolate(frame, [delay, delay + 14], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
        scale: interpolate(frame, [delay, delay + 20], [0.96 * scale, scale], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
          output: "perceptual-scale",
        }),
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 14,
          padding: "16px 24px",
          backgroundColor: COLOR.surfaceHi,
          borderBottom: `1px solid ${COLOR.line}`,
          fontFamily: SANS,
          fontSize: TYPE.caption,
          color: COLOR.textDim,
        }}
      >
        <span
          style={{
            width: 10,
            height: 10,
            borderRadius: 999,
            backgroundColor: accent,
            flexShrink: 0,
          }}
        />
        <span style={{ fontFamily: MONO, minWidth: 0, wordBreak: "break-all" }}>
          {snippet.line === undefined ? snippet.path : `${snippet.path}:${snippet.line}`}
        </span>
        {snippet.lang === "php" ? (
          <span
            style={{
              marginLeft: "auto",
              fontFamily: MONO,
              fontSize: TYPE.caption - 2,
              letterSpacing: 1.5,
              color: COLOR.php,
            }}
          >
            PHP
          </span>
        ) : null}
      </div>

      <div style={{ padding: "22px 24px" }}>
        {snippet.code.map((line: string, i: number) => (
          <Interactive.Div
            key={i}
            name={`Line ${i + 1}`}
            style={{
              fontFamily: MONO,
              fontSize: TYPE.code,
              lineHeight: 1.62,
              whiteSpace: "pre",
              color: highlight.includes(i) ? accent : COLOR.text,
              fontWeight: highlight.includes(i) ? 700 : 400,
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
