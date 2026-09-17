import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 4 — the convention that turns the cost into a benefit.
 *
 * "Zero value phải an toàn" is rule 9 of the repository, and Events is the
 * proof: nobody ever constructs one, every aggregate just embeds it, and
 * pulling from a fresh one returns nothing instead of crashing. Shown as a
 * flow because the point is what happens on the path, not what the code says.
 */
const NODES: FlowNode[] = [
  { id: "new", label: "CustomerOrder mới", sub: "chưa ai gán Events", x: 0, y: 40, w: 460, h: 104, color: COLOR.go, at: 20 },
  { id: "pull", label: "PullEvents()", sub: "gọi trên bản mặc định", x: 680, y: 40, w: 420, h: 104, color: COLOR.money, at: 90 },
  { id: "ok", label: "trả về slice rỗng", sub: "không panic, không null check", x: 1300, y: 40, w: 380, h: 104, color: COLOR.ok, at: 180 },
];

const EDGES: FlowEdge[] = [
  { from: "new", to: "pull", at: 80 },
  { from: "pull", to: "ok", at: 170 },
];

export const L4Safe: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 6 · CẢNH 4" source="quy ước 9 của repo: zero value phải an toàn">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 80, sound: "travel", volume: 0.16 },
        { at: 170, sound: "travel", volume: 0.16 },
        { at: 260, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Quy ước bù lại: zero value phải dùng được ngay
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={200} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={260} accent={COLOR.ok}>
          Không aggregate nào trong repo phải khởi tạo Events. Chúng chỉ nhúng nó vào, và
          bản mặc định đã chạy đúng — nên không có dòng kiểm tra null nào ở domain.
        </Callout>
      </div>
    </div>
  </Stage>
);
