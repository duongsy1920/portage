import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — where the rule actually ends up, drawn.
 *
 * The scattering is the problem, not any one of the places. Four boxes each
 * holding a copy of the same condition is a picture every Symfony developer
 * recognises immediately, including the part where one copy is out of date.
 */
const NODES: FlowNode[] = [
  { id: "rule", label: "“chỉ cọc được khi đơn vừa đặt”", sub: "một luật nghiệp vụ", x: 560, y: 0, w: 560, h: 96, color: COLOR.money, at: 20 },
  { id: "c", label: "Controller", sub: "kiểm trước khi gọi", x: 0, y: 200, w: 380, h: 88, color: COLOR.php, at: 90 },
  { id: "s", label: "Service", sub: "kiểm lại lần nữa", x: 430, y: 200, w: 380, h: 88, color: COLOR.php, at: 120 },
  { id: "v", label: "Validator của form", sub: "kiểm theo cách khác", x: 860, y: 200, w: 380, h: 88, color: COLOR.php, at: 150 },
  { id: "j", label: "Một job chạy nền", sub: "quên mất luật này", x: 1290, y: 200, w: 390, h: 88, color: COLOR.danger, at: 190 },
];

const EDGES: FlowEdge[] = [
  { from: "rule", to: "c", at: 80, dashed: true },
  { from: "rule", to: "s", at: 110, dashed: true },
  { from: "rule", to: "v", at: 140, dashed: true },
  { from: "rule", to: "j", at: 180, dashed: true, color: COLOR.danger },
];

export const A2Where: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 1 · CẢNH 2" source="mô hình thiếu máu: dữ liệu một nơi, luật ở bốn nơi">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 190, sound: "crack", volume: 0.2 }, { at: 280, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Nó rơi ra ngoài, và rơi thành nhiều bản
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={300} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={280} accent={COLOR.danger}>
          Hộp cuối là chỗ đau thật: job chạy nền không đi qua Controller nào cả. Nó gọi
          thẳng setStatus, và luật kia không có mặt ở đó.
        </Callout>
      </div>
    </div>
  </Stage>
);
