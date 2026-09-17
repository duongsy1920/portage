import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — what embedding actually does: it forwards.
 *
 * The mechanism is a lookup, not an inheritance chain: CustomerOrder has no
 * Record method, so Go looks one level in and calls the embedded one. Drawn as
 * a hop rather than described, because "method promotion" is a phrase that
 * explains nothing to someone who has only ever had `extends`.
 */
const NODES: FlowNode[] = [
  { id: "call", label: "order.Record(ev)", sub: "chỗ gọi", x: 0, y: 50, w: 420, h: 104, color: COLOR.go, at: 20 },
  { id: "order", label: "CustomerOrder", sub: "không có method Record", x: 620, y: 50, w: 420, h: 104, color: COLOR.textDim, at: 80 },
  { id: "events", label: "shared.Events", sub: "có Record — Go gọi cái này", x: 1240, y: 50, w: 440, h: 104, color: COLOR.money, at: 150 },
];

const EDGES: FlowEdge[] = [
  { from: "call", to: "order", at: 70 },
  { from: "order", to: "events", at: 140 },
];

const PACKETS: FlowPacket[] = [
  { label: "tìm Record ở đây trước", from: "call", to: "order", at: 72, offsetY: -70, color: COLOR.go },
  { label: "không có → tìm xuống một tầng", from: "order", to: "events", at: 142, offsetY: -70, color: COLOR.money },
];

export const M2Promote: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 7 · CẢNH 2" source="cơ chế: Go tìm method, không phải kế thừa">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 72, sound: "travel", volume: 0.16 },
        { at: 142, sound: "travel", volume: 0.16 },
        { at: 240, sound: "land", volume: 0.24 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cơ chế: Go tìm method xuống một tầng
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={220} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={240} accent={COLOR.go}>
          Không có lớp cha, không có bảng ảo, không ghi đè. Chỉ là: kiểu ngoài không có
          method đó thì tìm ở kiểu được nhúng vào.
        </Callout>
      </div>
    </div>
  </Stage>
);
