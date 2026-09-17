import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 2, episode 2, scene 1 — one question, asked of four things.
 *
 * The distinction is usually taught as two definitions. It is faster as a
 * single test applied to real things from this repository: if you replace it
 * with an identical one, does anything break?
 */
const NODES: FlowNode[] = [
  { id: "q", label: "“Đổi nó bằng một cái y hệt — có ai nhận ra không?”", x: 360, y: 0, w: 960, h: 88, color: COLOR.money, at: 20 },
  { id: "m", label: "150.00 USD", sub: "không ai nhận ra → Value Object", x: 0, y: 176, w: 400, h: 96, color: COLOR.ok, at: 90 },
  { id: "w", label: "500 gram", sub: "không ai nhận ra → Value Object", x: 430, y: 176, w: 400, h: 96, color: COLOR.ok, at: 130 },
  { id: "o", label: "đơn hàng #A17", sub: "khách nhận ra ngay → Entity", x: 860, y: 176, w: 400, h: 96, color: COLOR.php, at: 180 },
  { id: "p", label: "kiện hàng đang bay", sub: "kho nhận ra ngay → Entity", x: 1290, y: 176, w: 390, h: 96, color: COLOR.php, at: 220 },
];

export const B1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 2" source="một câu hỏi, và nó tách xong mọi kiểu trong domain">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 180, sound: "snap", volume: 0.18 }, { at: 280, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Cái gì cần ID, cái gì không
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} width={SAFE.contentWidth} height={290} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={280} accent={COLOR.money}>
          Hai tờ 150 đô là như nhau. Hai đơn hàng cùng giá trị thì không — khách đợi
          đúng đơn của họ. Đó là toàn bộ sự khác nhau.
        </Callout>
      </div>
    </div>
  </Stage>
);
