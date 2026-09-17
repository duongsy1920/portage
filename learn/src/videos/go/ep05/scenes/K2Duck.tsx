import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — who points at whom.
 *
 * This is the scene that makes hexagonal architecture click, so it is a
 * picture rather than a paragraph: the arrows all run INTO the interface, and
 * neither side knows the other exists. In Symfony the container knows both; in
 * Go nothing knows anything until main() wires it.
 */
const NODES: FlowNode[] = [
  { id: "app", label: "app/pricing", sub: "cần một chỗ lưu", x: 0, y: 40, w: 420, h: 104, color: COLOR.go, at: 20 },
  { id: "port", label: "interface", sub: "khai ở nơi cần, không ở nơi làm", x: 640, y: 40, w: 420, h: 104, color: COLOR.money, at: 70 },
  { id: "pg", label: "adapter/postgres", sub: "có đủ method", x: 1280, y: 0, w: 400, h: 92, color: COLOR.ok, at: 140 },
  { id: "mem", label: "adapter/memory", sub: "cũng có đủ method", x: 1280, y: 128, w: 400, h: 92, color: COLOR.ok, at: 180 },
];

const EDGES: FlowEdge[] = [
  { from: "app", to: "port", at: 110, color: COLOR.go },
  { from: "pg", to: "port", at: 220, color: COLOR.ok },
  { from: "mem", to: "port", at: 250, color: COLOR.ok },
];

export const K2Duck: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 5 · CẢNH 2" source="đây là lý do 278 test chạy dưới 2 giây">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 110, sound: "travel", volume: 0.16 },
        { at: 220, sound: "travel", volume: 0.16 },
        { at: 300, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Mũi tên chạy vào interface, không chạy ra
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={240} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.ok}>
          Đổi Postgres sang bản trong bộ nhớ là đổi một dòng trong main(). Tầng app
          không biết Postgres tồn tại, nên test nó không cần Docker.
        </Callout>
      </div>
    </div>
  </Stage>
);
