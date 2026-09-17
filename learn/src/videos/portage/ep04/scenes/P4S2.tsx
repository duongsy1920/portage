import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — the dependency rule as one picture.
 *
 * Every arrow points inward, and the outermost ring is the only place that
 * knows a database exists. Drawn as three bands rather than a hexagon, because
 * the hexagon shape is decoration and the direction is the lesson.
 */
const NODES: FlowNode[] = [
  { id: "dom", label: "internal/domain", sub: "chỉ stdlib + uuid", x: 1180, y: 60, w: 500, h: 116, color: COLOR.ok, at: 20 },
  { id: "app", label: "internal/app", sub: "use case + port (interface)", x: 600, y: 60, w: 500, h: 116, color: COLOR.go, at: 80 },
  { id: "ad", label: "internal/adapter", sub: "postgres · memory · http · openai", x: 0, y: 0, w: 500, h: 110, color: COLOR.money, at: 140 },
  { id: "pl", label: "internal/platform", sub: "wire · auth · clock", x: 0, y: 130, w: 500, h: 110, color: COLOR.money, at: 180 },
];

const EDGES: FlowEdge[] = [
  { from: "app", to: "dom", at: 130, color: COLOR.go },
  { from: "ad", to: "app", at: 220, color: COLOR.money },
  { from: "pl", to: "app", at: 250, color: COLOR.money },
];

export const P4S2: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 4 · CẢNH 2" source="mọi mũi tên chạy vào trong, không có ngoại lệ">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 130, sound: "travel", volume: 0.16 },
        { at: 220, sound: "travel", volume: 0.16 },
        { at: 300, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Luật chiều phụ thuộc, vẽ ra một lần
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={250} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.ok}>
          Domain nằm trong cùng và không biết gì về ba vòng ngoài. Đổi Postgres sang bản
          trong bộ nhớ là đổi một dòng ở main() — domain không đọc lại một chữ nào.
        </Callout>
      </div>
    </div>
  </Stage>
);
