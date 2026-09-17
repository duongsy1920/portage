import React from "react";
import { Easing, Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, RADIUS, TYPE } from "../design/tokens";
import { MONO, SANS } from "../design/fonts";

export const EASE = Easing.bezier(0.16, 1, 0.3, 1);

export type NodeKind = "box" | "store" | "actor" | "ghost" | "boundary";

export type FlowNode = {
  id: string;
  label: string;
  /** Second line, smaller — what the box IS in plain words. */
  sub?: string;
  /** Top-left corner inside the board. */
  x: number;
  y: number;
  w?: number;
  h?: number;
  color?: string;
  kind?: NodeKind;
  /** Frame the node appears. */
  at?: number;
};

export type FlowEdge = {
  from: string;
  to: string;
  /** Frame the arrow draws itself. */
  at: number;
  color?: string;
  /** Draw dashed — used for "this is not a function call". */
  dashed?: boolean;
};

export type FlowPacket = {
  /** Text carried along the wire, e.g. an event name. */
  label: string;
  from: string;
  to: string;
  /** Frame the packet leaves `from`. It arrives BEAT.travel frames later. */
  at: number;
  duration?: number;
  color?: string;
  /** Dịch đường bay lên/xuống, để packet không chồng lên hộp. */
  offsetY?: number;
};

const DEFAULT_W = 250;
const DEFAULT_H = 96;

function centre(n: FlowNode) {
  return {
    cx: n.x + (n.w ?? DEFAULT_W) / 2,
    cy: n.y + (n.h ?? DEFAULT_H) / 2,
  };
}

/**
 * Trim a segment so it starts and ends at the node borders rather than at the
 * node centres, otherwise every arrow disappears under the boxes it connects.
 */
function trim(a: FlowNode, b: FlowNode, pad = 14) {
  const A = centre(a);
  const B = centre(b);
  const dx = B.cx - A.cx;
  const dy = B.cy - A.cy;
  const len = Math.hypot(dx, dy) || 1;
  const ux = dx / len;
  const uy = dy / len;

  const shrink = (n: FlowNode) => {
    const hw = (n.w ?? DEFAULT_W) / 2;
    const hh = (n.h ?? DEFAULT_H) / 2;
    // Distance from the centre to the border along this direction.
    const tx = Math.abs(ux) < 1e-6 ? Infinity : hw / Math.abs(ux);
    const ty = Math.abs(uy) < 1e-6 ? Infinity : hh / Math.abs(uy);
    return Math.min(tx, ty) + pad;
  };

  const sa = shrink(a);
  const sb = shrink(b);
  return {
    x1: A.cx + ux * sa,
    y1: A.cy + uy * sa,
    x2: B.cx - ux * sb,
    y2: B.cy - uy * sb,
  };
}

/**
 * An animated flow diagram: boxes, arrows that draw themselves, and packets
 * that travel along the arrows carrying a label.
 *
 * This is the component the series leans on hardest. A learner can be told
 * "contexts talk by events" and nod without understanding; watching a labelled
 * packet leave one box, cross a gap, and land in another is the same sentence
 * in a form the eye can check.
 */
export const FlowBoard: React.FC<{
  nodes: readonly FlowNode[];
  edges?: readonly FlowEdge[];
  packets?: readonly FlowPacket[];
  width: number;
  height: number;
}> = ({ nodes, edges = [], packets = [], width, height }) => {
  const frame = useCurrentFrame();
  const byId = React.useMemo(
    () => Object.fromEntries(nodes.map((n) => [n.id, n])),
    [nodes],
  );

  return (
    <Interactive.Div
      name="FlowBoard"
      style={{ position: "relative", width, height }}
    >
      <svg
        width={width}
        height={height}
        style={{ position: "absolute", inset: 0, overflow: "visible" }}
      >
        <defs>
          <marker
            id="arrowhead"
            markerUnits="userSpaceOnUse"
            markerWidth="14"
            markerHeight="12"
            refX="12"
            refY="6"
            orient="auto"
          >
            <path d="M0,0 L14,6 L0,12 z" fill={COLOR.lineHi} />
          </marker>
        </defs>

        {edges.map((e, i) => {
          const a = byId[e.from];
          const b = byId[e.to];
          if (!a || !b) return null;
          const { x1, y1, x2, y2 } = trim(a, b);
          const len = Math.hypot(x2 - x1, y2 - y1);
          const drawn = interpolate(frame, [e.at, e.at + 20], [0, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: EASE,
          });

          return (
            <line
              key={i}
              x1={x1}
              y1={y1}
              x2={x2}
              y2={y2}
              stroke={e.color ?? COLOR.lineHi}
              strokeWidth={3}
              strokeDasharray={e.dashed ? "10 9" : `${len} ${len}`}
              strokeDashoffset={e.dashed ? 0 : len * (1 - drawn)}
              opacity={e.dashed ? drawn : 1}
              markerEnd="url(#arrowhead)"
            />
          );
        })}
      </svg>

      {nodes.map((n) => {
        const at = n.at ?? 0;
        const kind = n.kind ?? "box";
        const color = n.color ?? COLOR.lineHi;
        const ghost = kind === "ghost";
        const boundary = kind === "boundary";

        return (
          <Interactive.Div
            key={n.id}
            name={n.label}
            style={{
              position: "absolute",
              left: n.x,
              top: n.y,
              width: n.w ?? DEFAULT_W,
              height: n.h ?? DEFAULT_H,
              borderRadius: kind === "store" || boundary ? RADIUS.lg : RADIUS.md,
              backgroundColor: ghost || boundary ? "transparent" : COLOR.surface,
              border: `2px ${ghost || boundary || kind === "store" ? "dashed" : "solid"} ${color}`,
              display: "flex",
              flexDirection: "column",
              alignItems: boundary ? "flex-start" : "center",
              justifyContent: boundary ? "flex-start" : "center",
              paddingTop: boundary ? 14 : 0,
              gap: 6,
              padding: "0 16px",
              textAlign: "center",
              opacity: interpolate(frame, [at, at + 16], [0, ghost ? 0.45 : boundary ? 0.85 : 1], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
              }),
              scale: interpolate(frame, [at, at + 22], [0.9, 1], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
                easing: EASE,
                output: "perceptual-scale",
              }),
            }}
          >
            <div
              style={{
                fontFamily: MONO,
                fontSize: boundary ? TYPE.caption : TYPE.label,
                fontWeight: 700,
                letterSpacing: boundary ? 2 : 0.5,
                color,
                lineHeight: 1.15,
              }}
            >
              {n.label}
            </div>
            {n.sub ? (
              <div
                style={{
                  fontFamily: SANS,
                  fontSize: TYPE.caption,
                  color: COLOR.textDim,
                  lineHeight: 1.25,
                }}
              >
                {n.sub}
              </div>
            ) : null}
          </Interactive.Div>
        );
      })}

      {packets.map((p, i) => {
        const a = byId[p.from];
        const b = byId[p.to];
        if (!a || !b) return null;
        const dur = p.duration ?? 26;
        const { x1, y1, x2, y2 } = trim(a, b, 0);

        const t = interpolate(frame, [p.at, p.at + dur], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: Easing.bezier(0.4, 0, 0.2, 1),
        });
        // Fade in as it leaves, fade out as it lands.
        const alpha = interpolate(
          frame,
          [p.at, p.at + 6, p.at + dur - 6, p.at + dur],
          [0, 1, 1, 0],
          { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
        );
        const color = p.color ?? COLOR.go;

        return (
          <Interactive.Div
            key={i}
            name={`Packet ${p.label}`}
            style={{
              position: "absolute",
              left: x1 + (x2 - x1) * t,
              top: y1 + (y2 - y1) * t + (p.offsetY ?? 0),
              translate: "-50% -50%",
              padding: "8px 16px",
              borderRadius: 999,
              backgroundColor: COLOR.bg,
              border: `2px solid ${color}`,
              fontFamily: MONO,
              fontSize: TYPE.caption,
              fontWeight: 700,
              color,
              whiteSpace: "nowrap",
              opacity: alpha,
            }}
          >
            {p.label}
          </Interactive.Div>
        );
      })}
    </Interactive.Div>
  );
};
