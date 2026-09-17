import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 4 — %w, watched travelling.
 *
 * Wrapping is the one idea in this episode that a static screenshot cannot
 * teach, because the whole point is what happens to the error on its way up
 * through three layers. So it gets the flow board: the message grows at each
 * hop, and the original cause is still in there at the top.
 */
const NODES: FlowNode[] = [
  { id: "a", label: "parseDecimal", sub: "nơi hỏng thật sự", x: 0, y: 60, w: 360, h: 104, color: COLOR.danger, at: 20 },
  { id: "b", label: "ParseMoney", sub: "domain", x: 440, y: 60, w: 360, h: 104, color: COLOR.go, at: 70 },
  { id: "c", label: "CreateQuote", sub: "tầng app", x: 880, y: 60, w: 360, h: 104, color: COLOR.go, at: 140 },
  { id: "d", label: "HTTP 400", sub: "câu trả lời cho khách", x: 1320, y: 60, w: 360, h: 104, color: COLOR.ok, at: 210 },
];

const EDGES: FlowEdge[] = [
  { from: "a", to: "b", at: 60 },
  { from: "b", to: "c", at: 130 },
  { from: "c", to: "d", at: 200 },
];

const PACKETS: FlowPacket[] = [
  { label: "ErrMalformedAmount", from: "a", to: "b", at: 62, offsetY: -74, color: COLOR.danger },
  { label: 'parse "abc" as USD: %w', from: "b", to: "c", at: 132, offsetY: -74, color: COLOR.money },
  { label: "create quote: %w", from: "c", to: "d", at: 202, offsetY: -74, color: COLOR.money },
];

export const I4Wrap: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 3 · CẢNH 4" source="repo có 456 chỗ bọc lỗi bằng %w">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 62, sound: "travel", volume: 0.16 },
        { at: 132, sound: "travel", volume: 0.16 },
        { at: 202, sound: "travel", volume: 0.16 },
        { at: 300, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        %w: mỗi tầng thêm một câu, không xoá câu cũ
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={240} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.money}>
          Tới nơi, câu lỗi đọc được cả hành trình:{" "}
          <span style={{ color: COLOR.text }}>create quote: parse "abc" as USD: malformed amount</span>.
          Nguyên nhân gốc vẫn nằm nguyên trong đó — đó là việc của chữ %w.
        </Callout>
      </div>
    </div>
  </Stage>
);
