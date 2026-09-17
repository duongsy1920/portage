import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 3, episode 3, scene 1 — the 404 that is not a bug.
 *
 * A viewer coming from a synchronous world reads "created, then not found" as
 * broken. It is the outbox relay not having run yet. Showing the worker as its
 * own step — the thing the simulation in web/flow.html teaches best — is what
 * makes the status code read as a design decision.
 */
const NODES: FlowNode[] = [
  { id: "post", label: "POST /products", sub: "201 — đã tạo", x: 0, y: 40, w: 400, h: 100, color: COLOR.ok, at: 20 },
  { id: "ob", label: "outbox", sub: "event nằm đây, chưa ai đọc", x: 560, y: 40, w: 400, h: 100, color: COLOR.money, at: 80 },
  { id: "get", label: "GET /products/{id}", sub: "404 — chưa có trong bảng đọc", x: 1120, y: 0, w: 430, h: 92, color: COLOR.danger, at: 150 },
  { id: "relay", label: "relay chạy (200ms)", sub: "một bước riêng, không tự động", x: 1120, y: 128, w: 430, h: 92, color: COLOR.go, at: 220 },
  { id: "ok", label: "GET lần hai: 200", x: 1620, y: 60, w: 60, h: 60, color: COLOR.ok, at: 300 },
];

const EDGES: FlowEdge[] = [
  { from: "post", to: "ob", at: 70 },
  { from: "ob", to: "get", at: 140, dashed: true, color: COLOR.danger },
  { from: "ob", to: "relay", at: 210 },
];

const PACKETS: FlowPacket[] = [
  { label: "product_published", from: "post", to: "ob", at: 72, offsetY: -66, color: COLOR.money },
];

export const P3S1: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 3" source="404 ngay sau khi tạo — đây là thiết kế">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 72, sound: "travel", volume: 0.16 },
        { at: 150, sound: "crack", volume: 0.2 },
        { at: 300, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Tạo xong, đọc lại ngay: 404
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={240} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={330} accent={COLOR.go}>
          Bảng đọc chỉ đầy sau khi relay chạy. Nên 404 ở đây không có nghĩa là “không có”,
          mà là “chưa tới”. Khác nhau, và mã trạng thái phải nói được sự khác nhau đó.
        </Callout>
      </div>
    </div>
  </Stage>
);
