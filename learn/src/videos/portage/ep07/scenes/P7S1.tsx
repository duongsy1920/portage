import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, CONTEXT, SAFE } from "../../../../design/tokens";

/**
 * Season 3, episode 7, scene 1 — one screen, four contexts.
 *
 * The customer's order page needs the product name, the price, the order
 * status and where the parcel is. Those live in four packages that are not
 * allowed to import each other, so the JOIN a Symfony developer would reach
 * for is not merely discouraged here — it has nothing to join.
 */
const NODES: FlowNode[] = [
  { id: "ui", label: "một màn hình của khách", sub: "tên hàng · giá · trạng thái · kiện đang ở đâu", x: 560, y: 0, w: 640, h: 100, color: COLOR.text, at: 20 },
  { id: "c", label: "catalog", sub: "tên hàng", x: 0, y: 220, w: 380, h: 88, color: CONTEXT.catalog.color, at: 90 },
  { id: "p", label: "pricing", sub: "giá", x: 430, y: 220, w: 380, h: 88, color: CONTEXT.pricing.color, at: 120 },
  { id: "o", label: "ordering", sub: "trạng thái đơn", x: 860, y: 220, w: 380, h: 88, color: CONTEXT.ordering.color, at: 150 },
  { id: "l", label: "logistics", sub: "kiện đang ở đâu", x: 1290, y: 220, w: 390, h: 88, color: CONTEXT.logistics.color, at: 180 },
];

const EDGES: FlowEdge[] = [
  { from: "c", to: "ui", at: 230, dashed: true, color: COLOR.danger },
  { from: "p", to: "ui", at: 245, dashed: true, color: COLOR.danger },
  { from: "o", to: "ui", at: 260, dashed: true, color: COLOR.danger },
  { from: "l", to: "ui", at: 275, dashed: true, color: COLOR.danger },
];

export const P7S1: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 7" source="bốn vùng không được import lẫn nhau — nên không có gì để JOIN">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 230, sound: "crack", volume: 0.2 }, { at: 320, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Một màn hình, bốn vùng, không có JOIN nào
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={320} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={320} accent={COLOR.danger}>
          Bốn mũi tên đứt kia là thứ không tồn tại. Bốn vùng có bốn bảng riêng và một bài
          test cấm chúng import nhau — nên câu JOIN quen thuộc không có chỗ để viết.
        </Callout>
      </div>
    </div>
  </Stage>
);
