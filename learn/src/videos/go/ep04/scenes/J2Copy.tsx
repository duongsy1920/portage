import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — what a value receiver actually does, watched happening.
 *
 * "Go copies the struct" is a sentence a viewer will nod at and not believe
 * until they see the copy appear as its own box and the original stay put.
 * Hence the flow board rather than another bullet.
 */
const NODES: FlowNode[] = [
  { id: "orig", label: "m := MustParseMoney(…)", sub: "bản gốc 150.00, ở chỗ gọi", x: 0, y: 30, w: 430, h: 104, color: COLOR.go, at: 20 },
  { id: "copy", label: "m (bản sao)", sub: "bên trong method Add", x: 620, y: 30, w: 430, h: 104, color: COLOR.textDim, at: 90 },
  { id: "out", label: "một Money 175.00", sub: "giá trị mới trả về", x: 1240, y: 30, w: 430, h: 104, color: COLOR.ok, at: 190 },
  { id: "still", label: "m vẫn là 150.00", sub: "không ai đụng tới nó", x: 0, y: 214, w: 430, h: 88, color: COLOR.ok, at: 260 },
];

const EDGES: FlowEdge[] = [
  { from: "orig", to: "copy", at: 80 },
  { from: "copy", to: "out", at: 180 },
  { from: "orig", to: "still", at: 255, dashed: true, color: COLOR.ok },
];

const PACKETS: FlowPacket[] = [
  { label: "Go chép cả struct", from: "orig", to: "copy", at: 82, offsetY: -66, color: COLOR.go },
  { label: "sửa gì cũng chỉ sửa bản sao", from: "copy", to: "out", at: 182, offsetY: -66, color: COLOR.money },
];

export const J2Copy: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 4 · CẢNH 2" source="func (m Money) Add — receiver dạng giá trị">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 82, sound: "travel", volume: 0.16 },
        { at: 182, sound: "travel", volume: 0.16 },
        { at: 260, sound: "land", volume: 0.24 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Không có dấu sao: Go chép cả struct
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={320} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={320} accent={COLOR.ok}>
          Đây là lý do value object trong repo bất biến mà không cần cố gắng gì:
          method không có cách nào sửa bản gốc, kể cả khi lập trình viên muốn.
        </Callout>
      </div>
    </div>
  </Stage>
);
