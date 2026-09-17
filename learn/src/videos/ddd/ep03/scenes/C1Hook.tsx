import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, CONTEXT, SAFE } from "../../../../design/tokens";

const ORDERING = CONTEXT.ordering.color;

/**
 * Season 2, episode 3, scene 1 — the one door, drawn.
 *
 * "Aggregate" is a word; the useful thing is the shape. Everything outside has
 * to knock on the root, and the root is the only thing that can say no. The
 * boundary box is what makes that visible in one glance.
 */
const NODES: FlowNode[] = [
  {
    id: "b", label: "aggregate CustomerOrder", kind: "boundary",
    x: 560, y: 0, w: 700, h: 300, color: ORDERING, at: 30,
  },
  { id: "root", label: "CustomerOrder", sub: "cửa duy nhất — root", x: 600, y: 56, w: 620, h: 92, color: ORDERING, at: 60 },
  { id: "in1", label: "trạng thái · cọc · số dư", x: 600, y: 180, w: 300, h: 76, color: COLOR.lineHi, at: 110 },
  { id: "in2", label: "lý do huỷ · hoàn tiền", x: 920, y: 180, w: 300, h: 76, color: COLOR.lineHi, at: 130 },
  { id: "out", label: "tầng app", sub: "chỉ gọi được method của root", x: 0, y: 100, w: 420, h: 92, color: COLOR.go, at: 170 },
  { id: "no", label: "chạm thẳng vào field bên trong", sub: "không có cách nào", x: 1400, y: 100, w: 280, h: 92, color: COLOR.danger, at: 220 },
];

const EDGES: FlowEdge[] = [
  { from: "out", to: "root", at: 210, color: COLOR.go },
  { from: "no", to: "in2", at: 250, color: COLOR.danger, dashed: true },
];

export const C1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 3" source="internal/domain/ordering/order.go — không có setter công khai nào">
    <Cues items={[{ at: 30, sound: "appear" }, { at: 210, sound: "travel", volume: 0.16 }, { at: 250, sound: "crack", volume: 0.2 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Một cụm dữ liệu, một cửa vào
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={310} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={ORDERING}>
          Mọi field bên trong đều viết thường, nên package khác không chạm tới được.
          Muốn đổi gì cũng phải đi qua một method của root — và method đó kiểm luật trước.
        </Callout>
      </div>
    </div>
  </Stage>
);
