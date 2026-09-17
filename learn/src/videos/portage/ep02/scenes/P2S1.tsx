import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 3, episode 2, scene 1 — the design that loses one event a week.
 *
 * "Save, then publish" is what everyone writes first and it is correct
 * ninety-nine times out of a hundred. The hundredth is a process that dies in
 * the gap, and the order is now paid with nobody told to buy it. Drawn as a
 * gap, because the gap is the whole bug.
 */
const NODES: FlowNode[] = [
  { id: "save", label: "COMMIT: đơn = đã cọc", sub: "database đã ghi xong", x: 0, y: 60, w: 470, h: 104, color: COLOR.ok, at: 20 },
  { id: "gap", label: "khe hở", sub: "tiến trình chết ở đây", x: 700, y: 60, w: 340, h: 104, color: COLOR.danger, at: 110 },
  { id: "pub", label: "publish deposit_paid", sub: "không bao giờ chạy", x: 1280, y: 60, w: 400, h: 104, color: COLOR.textFaint, at: 200 },
];

const EDGES: FlowEdge[] = [
  { from: "save", to: "gap", at: 100 },
  { from: "gap", to: "pub", at: 190, dashed: true, color: COLOR.danger },
];

export const P2S1: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 2" source="cách ai cũng viết trước: lưu xong rồi publish">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 110, sound: "crack", volume: 0.24 }, { at: 260, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        “Lưu xong rồi publish” — hỏng ở đúng một chỗ
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={220} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={260} accent={COLOR.danger}>
          Khách đã trả cọc, database đã ghi, và không ai được bảo đi mua. Không có lỗi nào
          để lần, vì về mặt kỹ thuật thì mọi thứ đã thành công.
        </Callout>
      </div>
    </div>
  </Stage>
);
