import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, TYPE } from "../design/tokens";
import { MONO } from "../design/fonts";
import { EASE } from "../design/motion";

export type Bullet = {
  /** The sentence. Full sentences, not keyword fragments. */
  text: string;
  /** Optional lead-in shown before the sentence, e.g. "Vì sao". */
  tag?: string;
  /** Frame this bullet appears. */
  at: number;
  accent?: string;
};

/**
 * The explanation column.
 *
 * Bullets arrive one at a time and stay. For a viewer meeting Go and DDD for
 * the first time, the failure mode is a wall of text that is all present at
 * once and therefore read in no order at all; timing the lines IS the
 * explanation's structure.
 */
export const Bullets: React.FC<{ items: readonly Bullet[]; gap?: number }> = ({
  items,
  gap = 28,
}) => {
  const frame = useCurrentFrame();

  return (
    <div style={{ display: "flex", flexDirection: "column", gap }}>
      {items.map((b, i) => {
        const accent = b.accent ?? COLOR.go;
        return (
          <Interactive.Div
            key={i}
            name={`Bullet ${i + 1}`}
            style={{
              display: "flex",
              gap: 18,
              alignItems: "flex-start",
              opacity: interpolate(frame, [b.at, b.at + 18], [0, 1], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
              }),
              translate: interpolate(
                frame,
                [b.at, b.at + 24],
                ["0px 16px", "0px 0px"],
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
                width: 10,
                height: 10,
                borderRadius: 999,
                backgroundColor: accent,
                flexShrink: 0,
                marginTop: 15,
              }}
            />
            <span style={{ fontSize: TYPE.body, lineHeight: 1.45 }}>
              {b.tag ? (
                <span
                  style={{
                    fontFamily: MONO,
                    fontSize: TYPE.caption,
                    fontWeight: 700,
                    letterSpacing: 1,
                    color: accent,
                    marginRight: 12,
                  }}
                >
                  {b.tag.toUpperCase()}
                </span>
              ) : null}
              <span style={{ color: COLOR.text }}>{b.text}</span>
            </span>
          </Interactive.Div>
        );
      })}
    </div>
  );
};
