import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, TYPE } from "../design/tokens";
import { MONO, SANS } from "../design/fonts";
import { EASE } from "../design/motion";

/** One contiguous piece of the line. A piece with a note gets annotated. */
export type Part = {
  text: string;
  /** Plain-words explanation. Omit for punctuation and glue. */
  note?: string;
  /** Frame the piece lights up and its note arrives. */
  at?: number;
  color?: string;
};

/**
 * Take one real line of Go apart, piece by piece.
 *
 * This is the shape of season one: the learner is not stuck on concepts, they
 * are stuck on where to put their eyes inside `func (m Money) Add(o Money)
 * (Money, error)`. So the line stays whole on screen and the pieces light up
 * underneath it, in reading order, each with its own leader — never a rewrite
 * of the line into something friendlier, which teaches nothing.
 *
 * Positions are computed from character counts rather than measured, which is
 * exact here because the line is monospaced: JetBrains Mono advances 0.6em per
 * character. [PHP] Không có DOM measure trong Remotion — nhưng font đều nét
 * thì đếm ký tự là đủ.
 */
const CH = 0.6;
const ROW_H = 52;
const NOTE_W = 700;
/** Repeating a long fragment beside the line is noise; the leader already says which piece. */
const PREFIX_MAX = 12;

export const Anatomy: React.FC<{
  parts: readonly Part[];
  fontSize?: number;
  /** Width available, so a note near the right edge can be pulled back in. */
  width: number;
}> = ({ parts, fontSize = 34, width }) => {
  const frame = useCurrentFrame();
  const chW = fontSize * CH;

  // Left edge of each piece, in pixels, from the character counts before it.
  const offsets: number[] = [];
  let chars = 0;
  for (const p of parts) {
    offsets.push(chars * chW);
    chars += p.text.length;
  }

  const annotated = parts
    .map((p, i) => ({ p, i }))
    .filter(({ p }) => Boolean(p.note));

  // The code line, then a gap, then one row per note. Declared as a real
  // height so the block occupies its space in normal flow — the notes are
  // positioned absolutely, and a container that does not account for them
  // lets the next block land on top of this one.
  const lineH = fontSize * 1.5;
  const rowsTop = lineH + 26;
  const height = rowsTop + annotated.length * ROW_H;

  return (
    <Interactive.Div name="Anatomy" style={{ position: "relative", width, height }}>
      <div
        style={{
          fontFamily: MONO,
          fontSize,
          lineHeight: 1.5,
          whiteSpace: "pre",
          display: "flex",
        }}
      >
        {parts.map((p, i) => {
          const on = p.at === undefined || frame >= p.at;
          return (
            <span
              key={i}
              style={{
                color: on ? (p.color ?? COLOR.text) : COLOR.textFaint,
                fontWeight: on && p.note ? 700 : 400,
              }}
            >
              {p.text}
            </span>
          );
        })}
      </div>

      <svg
        width={width}
        height={height}
        style={{ position: "absolute", left: 0, top: 0, overflow: "visible" }}
      >
        {annotated.map(({ p, i }, row) => {
          const at = p.at ?? 0;
          const x = offsets[i] + (p.text.length * chW) / 2;
          const y = rowsTop + row * ROW_H;
          const draw = interpolate(frame, [at, at + 16], [0, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: EASE,
          });
          const colour = p.color ?? COLOR.go;
          const noteLeft = Math.min(x + 18, width - NOTE_W);
          return (
            <g key={i} opacity={draw}>
              <line x1={x} y1={lineH + 2} x2={x} y2={y} stroke={colour} strokeWidth={1.5} opacity={0.45} />
              <circle cx={x} cy={y} r={3.5} fill={colour} />
              <line x1={x} y1={y} x2={noteLeft - 8} y2={y} stroke={colour} strokeWidth={1.5} opacity={0.45} />
            </g>
          );
        })}
      </svg>

      {annotated.map(({ p, i }, row) => {
        const at = p.at ?? 0;
        const x = offsets[i] + (p.text.length * chW) / 2;
        const y = rowsTop + row * ROW_H;
        return (
          <div
            key={i}
            style={{
              position: "absolute",
              left: Math.min(x + 18, width - NOTE_W),
              top: y - 16,
              width: NOTE_W,
              fontFamily: SANS,
              fontSize: TYPE.label,
              color: COLOR.text,
              lineHeight: 1.25,
              whiteSpace: "nowrap",
              opacity: interpolate(frame, [at + 6, at + 22], [0, 1], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
              }),
            }}
          >
            {p.text.trim().length <= PREFIX_MAX ? (
              <>
                <span style={{ color: p.color ?? COLOR.go, fontFamily: MONO, fontSize: TYPE.caption }}>
                  {p.text.trim()}
                </span>
                {"  "}
              </>
            ) : null}
            {p.note}
          </div>
        );
      })}
    </Interactive.Div>
  );
};
