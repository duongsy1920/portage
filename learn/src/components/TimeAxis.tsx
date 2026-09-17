import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, CONTEXT, TYPE, type ContextKey } from "../design/tokens";
import { MONO } from "../design/fonts";
import { EASE } from "../design/motion";

export type Milestone = {
  /** The real day number in the business, used to place the mark. */
  day: number;
  label: string;
  what: string;
  ctx: ContextKey;
  at: number;
};

/**
 * A time axis where marks sit at their REAL position in the month.
 *
 * A list of seven bullets says "thirty days" and teaches nothing. Placing the
 * marks proportionally makes the shape of the problem visible on its own: four
 * things happen in the first two days, and then there are three weeks in which
 * the business is waiting on a shop, a warehouse and an airline it does not
 * control. That gap is the reason the whole architecture exists, so the
 * picture has to contain it.
 */
export const TimeAxis: React.FC<{
  milestones: readonly Milestone[];
  width: number;
  /** Day the axis ends on. */
  span?: number;
  /**
   * Horizontal breathing room at both ends. The default suits short labels;
   * a scene whose first mark carries a long caption needs more, or the caption
   * centres itself off the left edge of the frame.
   */
  pad?: number;
}> = ({ milestones, width, span = 32, pad = 70 }) => {
  const frame = useCurrentFrame();
  const usable = width - pad * 2;
  const xOf = (day: number) => pad + (day / span) * usable;

  const first = milestones[0];
  const last = milestones[milestones.length - 1];

  return (
    <Interactive.Div
      name="TimeAxis"
      style={{ position: "relative", width, height: 360 }}
    >
      {/* The axis itself draws left to right, so time reads as time. */}
      <Interactive.Div
        name="Axis"
        style={{
          position: "absolute",
          left: pad,
          top: 178,
          height: 4,
          borderRadius: 999,
          backgroundColor: COLOR.line,
          width: interpolate(frame, [0, 40], [0, usable], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: EASE,
          }),
        }}
      />

      {/* The waiting gap, called out because it is the argument. */}
      <Interactive.Div
        name="Waiting gap"
        style={{
          position: "absolute",
          left: xOf(first.day),
          top: 150,
          width: xOf(last.day) - xOf(first.day),
          height: 60,
          borderRadius: 8,
          backgroundColor: "rgba(242,179,61,0.09)",
          border: `1px dashed ${COLOR.money}`,
          opacity: interpolate(frame, [180, 210], [0, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: EASE,
          }),
        }}
      />

      {milestones.map((m, i) => {
        const x = xOf(m.day);
        const above = i % 2 === 0;
        const c = CONTEXT[m.ctx];

        return (
          <Interactive.Div
            key={i}
            name={`${m.label} ${m.what}`}
            style={{
              position: "absolute",
              left: x,
              top: above ? 40 : 216,
              width: 230,
              translate: "-50% 0",
              textAlign: "center",
              opacity: interpolate(frame, [m.at, m.at + 18], [0, 1], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
              }),
            }}
          >
            {above ? null : <Tick color={c.color} at={m.at} flip />}
            <div
              style={{
                fontFamily: MONO,
                fontSize: TYPE.caption,
                fontWeight: 700,
                letterSpacing: 1.5,
                color: c.color,
                marginBottom: 6,
              }}
            >
              {m.label.toUpperCase()}
            </div>
            <div style={{ fontSize: TYPE.label, lineHeight: 1.25, color: COLOR.text }}>
              {m.what}
            </div>
            {above ? <Tick color={c.color} at={m.at} /> : null}
          </Interactive.Div>
        );
      })}
    </Interactive.Div>
  );
};

/** The short stem joining a label to the axis. */
const Tick: React.FC<{ color: string; at: number; flip?: boolean }> = ({
  color,
  at,
  flip,
}) => {
  const frame = useCurrentFrame();
  return (
    <div
      style={{
        width: 3,
        height: 44,
        margin: flip ? "0 auto 10px" : "10px auto 0",
        borderRadius: 999,
        backgroundColor: color,
        opacity: interpolate(frame, [at + 8, at + 22], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: EASE,
        }),
      }}
    />
  );
};
