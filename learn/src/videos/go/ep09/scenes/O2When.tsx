import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — when exactly it runs, drawn as a path.
 *
 * "At the end of the function" is not precise enough to be useful: the three
 * exits behave the same, and that sameness is the whole value. So all three
 * paths are on screen at once, converging on the same box.
 */
const NODES: FlowNode[] = [
  { id: "open", label: "mu.Lock()", sub: "mở khoá vào", x: 620, y: 0, w: 420, h: 92, color: COLOR.go, at: 20 },
  { id: "def", label: "defer mu.Unlock()", sub: "hẹn: chạy lúc hàm thoát", x: 620, y: 132, w: 420, h: 92, color: COLOR.money, at: 60 },
  { id: "ret", label: "return bình thường", x: 0, y: 292, w: 400, h: 84, color: COLOR.ok, at: 130 },
  { id: "err", label: "return giữa chừng vì lỗi", x: 440, y: 292, w: 400, h: 84, color: COLOR.money, at: 160 },
  { id: "pan", label: "panic", x: 880, y: 292, w: 400, h: 84, color: COLOR.danger, at: 190 },
  { id: "run", label: "mu.Unlock() vẫn chạy", sub: "cả ba đường đều qua đây", x: 1320, y: 292, w: 360, h: 84, color: COLOR.ok, at: 250 },
];

const EDGES: FlowEdge[] = [
  { from: "open", to: "def", at: 50 },
  { from: "ret", to: "run", at: 240 },
  { from: "err", to: "run", at: 250 },
  { from: "pan", to: "run", at: 260 },
];

export const O2When: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 9 · CẢNH 2" source="ba đường thoát, một chỗ dọn dẹp">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 190, sound: "crack", volume: 0.18 },
        { at: 250, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Thoát kiểu gì cũng chạy
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={390} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={310} accent={COLOR.ok}>
          Đây là lý do khoá trong repo không bao giờ kẹt: không ai phải nhớ mở khoá ở
          từng đường return, vì chỉ có một chỗ mở.
        </Callout>
      </div>
    </div>
  </Stage>
);
