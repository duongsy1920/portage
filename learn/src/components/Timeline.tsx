import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, CONTEXT, RADIUS, TYPE, type ContextKey } from "../design/tokens";
import { MONO } from "../design/fonts";
import { EASE } from "./Stage";

export type TimelineItem = {
  readonly day: string;
  readonly what: string;
  readonly ctx: ContextKey;
};

const ROW_H = 118;

/**
 * The real life of one order, as time — not as boxes.
 *
 * This is the evidence scene 2 rests on: the viewer has to FEEL that thirty
 * days pass between the deposit and the delivery before the argument about
 * transactions can land.
 */
export const Timeline: React.FC<{
  items: readonly TimelineItem[];
  delay?: number;
  stagger?: number;
  /** Rows above this index are dimmed — used to point at one moment. */
  focus?: number;
}> = ({ items, delay = 0, stagger = 10, focus }) => {
  const frame = useCurrentFrame();

  return (
    <Interactive.Div name="Timeline" style={{ position: "relative" }}>
      {/* The spine the dots sit on. */}
      <Interactive.Div
        name="Spine"
        style={{
          position: "absolute",
          left: 17,
          top: 20,
          width: 3,
          backgroundColor: COLOR.line,
          height: interpolate(
            frame,
            [delay, delay + items.length * stagger + 20],
            [0, (items.length - 1) * ROW_H],
            {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            },
          ),
        }}
      />

      {items.map((item, i) => {
        const at = delay + i * stagger;
        const on = focus === undefined || focus === i;
        const c = CONTEXT[item.ctx];

        return (
          <Interactive.Div
            key={i}
            name={`${item.day} ${item.what}`}
            style={{
              height: ROW_H,
              display: "flex",
              alignItems: "flex-start",
              gap: 30,
              opacity: interpolate(frame, [at, at + 16], [0, on ? 1 : 0.28], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
              }),
              translate: interpolate(
                frame,
                [at, at + 22],
                ["0px 18px", "0px 0px"],
                {
                  extrapolateLeft: "clamp",
                  extrapolateRight: "clamp",
                  easing: EASE,
                },
              ),
            }}
          >
            <div
              style={{
                width: 36,
                height: 36,
                borderRadius: 999,
                backgroundColor: COLOR.bg,
                border: `4px solid ${c.color}`,
                flexShrink: 0,
                marginTop: 6,
              }}
            />
            <div>
              <div
                style={{
                  fontFamily: MONO,
                  fontSize: TYPE.caption,
                  letterSpacing: 2,
                  color: c.color,
                  fontWeight: 700,
                  marginBottom: 6,
                }}
              >
                {item.day.toUpperCase()}
              </div>
              <div
                style={{
                  fontSize: TYPE.body,
                  fontWeight: 600,
                  color: COLOR.text,
                  lineHeight: 1.2,
                }}
              >
                {item.what}
              </div>
            </div>
          </Interactive.Div>
        );
      })}
    </Interactive.Div>
  );
};

/**
 * A `BEGIN … COMMIT` bracket drawn alongside the timeline, which snaps.
 *
 * The break is the whole argument of episode 1 in one image: a transaction is
 * a promise that everything inside it happens together, and nothing that
 * takes thirty days and involves a shop in another country can keep it.
 */
export const TxBracket: React.FC<{
  height: number;
  /** Frame at which the bracket starts drawing. */
  delay?: number;
  /** Frame at which it snaps. Omit to leave it intact. */
  breakAt?: number;
}> = ({ height, delay = 0, breakAt }) => {
  const frame = useCurrentFrame();
  const broken = breakAt !== undefined && frame >= breakAt;

  const grow = interpolate(frame, [delay, delay + 26], [0, height], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: EASE,
  });

  const split = breakAt
    ? interpolate(frame, [breakAt, breakAt + 16], [0, 40], {
        extrapolateLeft: "clamp",
        extrapolateRight: "clamp",
        easing: EASE,
      })
    : 0;

  const color = broken ? COLOR.danger : COLOR.money;

  return (
    <Interactive.Div
      name="Transaction bracket"
      style={{ width: 170, flexShrink: 0 }}
    >
      <Interactive.Div
        name="BEGIN"
        style={{
          fontFamily: MONO,
          fontSize: TYPE.caption,
          fontWeight: 700,
          color,
          marginBottom: 10,
          opacity: interpolate(frame, [delay, delay + 14], [0, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: EASE,
          }),
        }}
      >
        BEGIN
      </Interactive.Div>

      {/* Upper half of the bracket. */}
      <div
        style={{
          borderLeft: `4px solid ${color}`,
          borderTop: `4px solid ${color}`,
          borderBottom: broken ? "none" : undefined,
          borderTopLeftRadius: RADIUS.sm,
          width: 140,
          height: Math.max(0, grow / 2 - split / 2),
          marginLeft: 8,
        }}
      />
      {/* Lower half — the gap between the two is the failure. */}
      <div
        style={{
          borderLeft: `4px solid ${color}`,
          borderBottom: `4px solid ${color}`,
          borderBottomLeftRadius: RADIUS.sm,
          width: 140,
          height: Math.max(0, grow / 2 - split / 2),
          marginLeft: 8,
          marginTop: split,
        }}
      />

      <Interactive.Div
        name="COMMIT"
        style={{
          fontFamily: MONO,
          fontSize: TYPE.caption,
          fontWeight: 700,
          color,
          marginTop: 10,
          whiteSpace: "nowrap",
          opacity: interpolate(
            frame,
            [delay + 20, delay + 34],
            [0, 1],
            {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            },
          ),
        }}
      >
        {broken ? "✕ VỠ" : "COMMIT"}
      </Interactive.Div>
    </Interactive.Div>
  );
};
