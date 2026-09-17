import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — why append behaves that way.
 *
 * A slice is a small struct (pointer, length, capacity) passed by value, which
 * is episode 4's lesson applied to the standard library. Once that lands, the
 * rule stops being arbitrary and never has to be memorised again.
 */
const NODES: FlowNode[] = [
  { id: "slice", label: "slice = 3 con số", sub: "con trỏ · độ dài · sức chứa", x: 0, y: 30, w: 460, h: 104, color: COLOR.go, at: 20 },
  { id: "arr", label: "mảng thật nằm chỗ khác", sub: "chỉ có con trỏ trỏ tới", x: 0, y: 200, w: 460, h: 96, color: COLOR.textDim, at: 70 },
  { id: "call", label: "truyền vào append", sub: "chép 3 con số — luật tập 4", x: 660, y: 30, w: 440, h: 104, color: COLOR.money, at: 140 },
  { id: "new", label: "trả về 3 con số mới", sub: "độ dài đã +1", x: 1280, y: 30, w: 400, h: 104, color: COLOR.ok, at: 220 },
];

const EDGES: FlowEdge[] = [
  { from: "slice", to: "arr", at: 110, dashed: true },
  { from: "slice", to: "call", at: 130 },
  { from: "call", to: "new", at: 210 },
];

export const N2Why: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 8 · CẢNH 2" source="slice là một struct nhỏ, nên nó bị chép như mọi struct">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 130, sound: "travel", volume: 0.16 },
        { at: 210, sound: "travel", volume: 0.16 },
        { at: 300, sound: "land", volume: 0.24 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Vì sao append phải gán lại
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={310} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.go}>
          Không có luật mới nào ở đây. Slice là struct, struct thì bị chép khi truyền đi —
          đúng bài tập 4. Nên append không có cách nào sửa slice của người gọi.
        </Callout>
      </div>
    </div>
  </Stage>
);
