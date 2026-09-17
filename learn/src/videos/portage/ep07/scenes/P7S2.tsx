import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, CONTEXT, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — the read model, built by listening.
 *
 * Nobody queries the four contexts. They each announce what happened, and a
 * projector writes one wide row that the screen reads in a single SELECT. The
 * cost is honest and worth naming: that row is a few hundred milliseconds
 * behind, which is episode 3's 404 seen from the other side.
 */
const NODES: FlowNode[] = [
  { id: "c", label: "catalog", x: 0, y: 0, w: 320, h: 76, color: CONTEXT.catalog.color, at: 20 },
  { id: "p", label: "pricing", x: 0, y: 100, w: 320, h: 76, color: CONTEXT.pricing.color, at: 40 },
  { id: "o", label: "ordering", x: 0, y: 200, w: 320, h: 76, color: CONTEXT.ordering.color, at: 60 },
  { id: "l", label: "logistics", x: 0, y: 300, w: 320, h: 76, color: CONTEXT.logistics.color, at: 80 },
  { id: "ob", label: "outbox", sub: "mọi việc đã xảy ra rơi vào đây", x: 560, y: 140, w: 420, h: 96, color: COLOR.money, at: 130 },
  { id: "proj", label: "projector", sub: "nghe, rồi ghi một dòng rộng", x: 1160, y: 140, w: 400, h: 96, color: COLOR.go, at: 220 },
  { id: "tb", label: "order_summaries", sub: "đọc một câu SELECT", x: 1290, y: 300, w: 390, h: 90, color: COLOR.ok, at: 300 },
];

const EDGES: FlowEdge[] = [
  { from: "c", to: "ob", at: 120 },
  { from: "p", to: "ob", at: 130 },
  { from: "o", to: "ob", at: 140 },
  { from: "l", to: "ob", at: 150 },
  { from: "ob", to: "proj", at: 210 },
  { from: "proj", to: "tb", at: 290 },
];

const PACKETS: FlowPacket[] = [
  { label: "event", from: "ob", to: "proj", at: 212, offsetY: -62, color: COLOR.money },
];

export const P7S2: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 7 · CẢNH 2" source="bảng đọc dựng bằng event — CQRS, nhưng không cần gọi tên">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 212, sound: "travel", volume: 0.16 },
        { at: 300, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Không hỏi ai cả — nghe rồi tự ghi lại
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={390} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={330} accent={COLOR.go}>
          Cái giá phải nói thẳng: dòng đó chậm hơn sự thật vài trăm mili giây. Đó chính là
          cái 404 của tập 3, nhìn từ phía bên kia.
        </Callout>
      </div>
    </div>
  </Stage>
);
