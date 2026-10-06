import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { COLOR, RADIUS, TYPE } from "../../../../design/tokens";
import { MONO } from "../../../../design/fonts";
import { EASE } from "../../../../design/motion";
import { Cues } from "../../../../components/Sfx";
import { RECAP } from "../recap";

/**
 * Scene 11 — recall, not summary.
 *
 * The chain is shown in the order it was derived, because the chain itself is
 * the thing worth remembering: each step forces the next, and a learner who
 * can replay the forcing can rebuild the architecture from the business.
 */
const CHAIN = RECAP.points;

export const S11Recap: React.FC = () => {
  const frame = useCurrentFrame();

  return (
    <Stage eyebrow="TẬP 1 · NHỚ LẠI">
      <Cues
        items={[
          { at: 30, sound: "appear", volume: 0.12 },
          { at: 130, sound: "appear", volume: 0.12 },
          { at: 230, sound: "reveal" },
          { at: 290, sound: "turn" },
        ]}
      />
      <div
        style={{
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          gap: 44,
        }}
      >
        <Headline delay={4} size={60}>
          Một chuỗi, không phải năm lựa chọn rời rạc
        </Headline>

        <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
          {CHAIN.map((c, i) => (
            <Interactive.Div
              key={i}
              name={`Bước ${i + 1}`}
              style={{
                display: "flex",
                gap: 26,
                alignItems: "flex-start",
                opacity: interpolate(frame, [c.at, c.at + 20], [0, 1], {
                  extrapolateLeft: "clamp",
                  extrapolateRight: "clamp",
                  easing: EASE,
                }),
                translate: interpolate(
                  frame,
                  [c.at, c.at + 26],
                  ["-20px 0px", "0px 0px"],
                  {
                    extrapolateLeft: "clamp",
                    extrapolateRight: "clamp",
                    easing: EASE,
                  },
                ),
              }}
            >
              <span
                style={{
                  fontFamily: MONO,
                  fontSize: TYPE.label,
                  fontWeight: 700,
                  color: COLOR.go,
                  flexShrink: 0,
                  width: 54,
                }}
              >
                {String(i + 1).padStart(2, "0")}
              </span>
              <span style={{ fontSize: TYPE.sub, lineHeight: 1.35, color: COLOR.text }}>
                {c.text}
              </span>
            </Interactive.Div>
          ))}
        </div>

        <Interactive.Div
          name="Next"
          style={{
            marginTop: 20,
            alignSelf: "flex-start",
            padding: "28px 36px",
            borderRadius: RADIUS.md,
            border: `1px solid ${COLOR.line}`,
            backgroundColor: COLOR.surface,
            opacity: interpolate(frame, [290, 320], [0, 1], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            }),
          }}
        >
          <div
            style={{
              fontSize: TYPE.caption,
              letterSpacing: 3,
              color: COLOR.textFaint,
              marginBottom: 10,
              fontWeight: 700,
            }}
          >
            {RECAP.next.label}
          </div>
          <div style={{ fontSize: TYPE.sub, fontWeight: 600 }}>
            {RECAP.next.text}
          </div>
        </Interactive.Div>
      </div>
    </Stage>
  );
};
