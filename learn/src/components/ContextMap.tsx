import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, CONTEXT, RADIUS, TYPE, type ContextKey } from "../design/tokens";
import { MONO } from "../design/fonts";
import { EASE } from "../design/motion";
import { CONTEXTS } from "../data/portage";

const CARD_W = 250;
const CARD_H = 140;
const STEP = 362;

/**
 * THE MASTER MENTAL MODEL of the series.
 *
 * Five write-side contexts in a row; the outbox as a river running underneath
 * them; the read side as a bar below that. The one thing it must teach, and
 * the reason there is not a single arrow BETWEEN the cards: contexts never
 * call each other. Everything one context wants to say falls into the river,
 * and whoever cares picks it up there.
 *
 * Every episode closes on this picture with a different part lit, so the map
 * is learned by repetition instead of being re-explained each time.
 */
export const ContextMap: React.FC<{
  /** Contexts drawn in full colour. Everything else is dimmed. */
  highlight?: readonly ContextKey[];
  delay?: number;
  showRiver?: boolean;
  showReporting?: boolean;
  stagger?: number;
  width: number;
}> = ({
  highlight,
  delay = 0,
  showRiver = true,
  showReporting = true,
  stagger = 8,
  width,
}) => {
  const frame = useCurrentFrame();
  const lit = (k: ContextKey) => !highlight || highlight.includes(k);
  const riverY = CARD_H + 56;
  const repY = riverY + 104;

  return (
    <Interactive.Div
      name="ContextMap"
      style={{
        position: "relative",
        width,
        height: showReporting ? repY + 88 : riverY + 88,
      }}
    >
      {CONTEXTS.map((c, i) => {
        const key = c.key as ContextKey;
        const meta = CONTEXT[key];
        const at = delay + i * stagger;
        const on = lit(key);

        return (
          <Interactive.Div
            key={key}
            name={meta.label}
            style={{
              position: "absolute",
              left: i * STEP,
              top: 0,
              width: CARD_W,
              height: CARD_H,
              borderRadius: RADIUS.md,
              backgroundColor: COLOR.surface,
              border: `2px solid ${on ? meta.color : COLOR.line}`,
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              gap: 8,
              padding: "0 14px",
              textAlign: "center",
              opacity: interpolate(frame, [at, at + 16], [0, on ? 1 : 0.3], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
              }),
              translate: interpolate(frame, [at, at + 22], ["0px -20px", "0px 0px"], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
              }),
            }}
          >
            <div
              style={{
                fontFamily: MONO,
                fontSize: TYPE.label,
                fontWeight: 700,
                color: on ? meta.color : COLOR.textFaint,
              }}
            >
              {meta.label}
            </div>
            <div style={{ fontSize: TYPE.caption, color: COLOR.textDim, lineHeight: 1.25 }}>
              {c.short}
            </div>
          </Interactive.Div>
        );
      })}

      {showRiver ? (
        <>
          {/* Stems from each context down into the river. */}
          {CONTEXTS.map((c, i) => (
            <div
              key={`stem-${c.key}`}
              style={{
                position: "absolute",
                left: i * STEP + CARD_W / 2 - 1.5,
                top: CARD_H,
                width: 3,
                height: 56,
                backgroundColor: COLOR.goDim,
                opacity: interpolate(frame, [delay + 46, delay + 66], [0, 0.8], {
                  extrapolateLeft: "clamp",
                  extrapolateRight: "clamp",
                  easing: EASE,
                }),
              }}
            />
          ))}

          <Interactive.Div
            name="Outbox river"
            style={{
              position: "absolute",
              left: 0,
              top: riverY,
              width,
              height: 78,
              borderRadius: RADIUS.lg,
              border: `2px dashed ${COLOR.goDim}`,
              backgroundColor: "rgba(0,173,216,0.08)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 20,
              opacity: interpolate(frame, [delay + 46, delay + 70], [0, 1], {
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
                letterSpacing: 5,
                color: COLOR.go,
              }}
            >
              OUTBOX
            </span>
            <span style={{ fontSize: TYPE.caption, color: COLOR.textFaint }}>
              mọi việc đã xảy ra đều rơi xuống đây, rồi ai cần thì nhặt lên
            </span>
          </Interactive.Div>
        </>
      ) : null}

      {showReporting ? (
        <Interactive.Div
          name="Reporting"
          style={{
            position: "absolute",
            left: 0,
            top: repY,
            width,
            height: 78,
            borderRadius: RADIUS.md,
            border: `2px solid ${COLOR.line}`,
            backgroundColor: COLOR.bgLift,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            gap: 20,
            opacity: interpolate(frame, [delay + 76, delay + 100], [0, 1], {
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
              letterSpacing: 2,
              color: CONTEXT.reporting.color,
            }}
          >
            REPORTING
          </span>
          <span style={{ fontSize: TYPE.caption, color: COLOR.textFaint }}>
            màn hình cho khách và nhân viên, dựng chỉ bằng event
          </span>
        </Interactive.Div>
      ) : null}
    </Interactive.Div>
  );
};
