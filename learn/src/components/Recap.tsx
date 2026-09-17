import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { Stage } from "./Stage";
import { Headline } from "./Text";
import { Cues } from "./Sfx";
import { COLOR, RADIUS, TYPE } from "../design/tokens";
import { MONO } from "../design/fonts";
import { EASE } from "../design/motion";

/**
 * The closing screen every episode ends with.
 *
 * Extracted because there are twenty-four of these to come and they must all
 * look and behave identically — recall is built on repetition, and a closing
 * screen that drifts in layout from episode to episode trains nothing. The
 * interview box is not decoration: the stated goal of season one is to answer
 * questions about Go out loud, and a lesson that never rehearses the answer
 * does not get that done.
 */
export type RecapPoint = { at: number; text: string };

const fade = (frame: number, at: number, span = 20) =>
  interpolate(frame, [at, at + span], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: EASE,
  });

export const Recap: React.FC<{
  eyebrow: string;
  title?: string;
  points: readonly RecapPoint[];
  asked: readonly string[];
  /** What the next episode is about. Omit on the last episode of a season. */
  nextLabel?: string;
  nextText?: string;
  askedAt?: number;
  nextAt?: number;
  accent?: string;
}> = ({
  eyebrow,
  title = "Bốn câu mang theo",
  points,
  asked,
  nextLabel,
  nextText,
  askedAt = 230,
  nextAt = 320,
  accent = COLOR.go,
}) => {
  const frame = useCurrentFrame();

  return (
    <Stage eyebrow={eyebrow}>
      <Cues
        items={[
          { at: points[0]?.at ?? 30, sound: "appear", volume: 0.12 },
          { at: askedAt, sound: "reveal" },
          { at: nextAt, sound: "turn" },
        ]}
      />
      <div
        style={{
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          gap: 38,
        }}
      >
        <Headline delay={4} size={60}>
          {title}
        </Headline>

        <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          {points.map((p, i) => (
            <Interactive.Div
              key={i}
              name={`Điểm ${i + 1}`}
              style={{
                display: "flex",
                gap: 24,
                alignItems: "flex-start",
                opacity: fade(frame, p.at),
                translate: interpolate(frame, [p.at, p.at + 26], ["-18px 0px", "0px 0px"], {
                  extrapolateLeft: "clamp",
                  extrapolateRight: "clamp",
                  easing: EASE,
                }),
              }}
            >
              <span
                style={{
                  fontFamily: MONO,
                  fontSize: TYPE.label,
                  fontWeight: 700,
                  color: accent,
                  width: 50,
                  flexShrink: 0,
                }}
              >
                {String(i + 1).padStart(2, "0")}
              </span>
              <span style={{ fontSize: TYPE.body, lineHeight: 1.4 }}>{p.text}</span>
            </Interactive.Div>
          ))}
        </div>

        <div style={{ display: "flex", gap: 40, marginTop: 10 }}>
          <Interactive.Div
            name="Interview"
            style={{
              flex: 1,
              padding: "28px 34px",
              borderRadius: RADIUS.md,
              backgroundColor: COLOR.surface,
              borderLeft: `8px solid ${COLOR.money}`,
              opacity: fade(frame, askedAt, 28),
            }}
          >
            <div
              style={{
                fontSize: TYPE.caption,
                letterSpacing: 3,
                fontWeight: 700,
                color: COLOR.money,
                marginBottom: 14,
              }}
            >
              PHỎNG VẤN SẼ HỎI
            </div>
            {asked.map((q, i) => (
              <div key={i} style={{ fontSize: TYPE.label, lineHeight: 1.5, color: COLOR.text }}>
                • {q}
              </div>
            ))}
          </Interactive.Div>

          {nextText ? (
            <Interactive.Div
              name="Next"
              style={{
                flex: 1,
                padding: "28px 34px",
                borderRadius: RADIUS.md,
                border: `1px solid ${COLOR.line}`,
                backgroundColor: COLOR.bgLift,
                opacity: fade(frame, nextAt, 28),
              }}
            >
              <div
                style={{
                  fontSize: TYPE.caption,
                  letterSpacing: 3,
                  fontWeight: 700,
                  color: COLOR.textFaint,
                  marginBottom: 14,
                }}
              >
                {nextLabel}
              </div>
              <div style={{ fontSize: TYPE.sub, fontWeight: 600, lineHeight: 1.3 }}>{nextText}</div>
            </Interactive.Div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
