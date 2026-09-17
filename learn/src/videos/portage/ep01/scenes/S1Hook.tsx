import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Body } from "../../../../components/Text";
import { FlowBoard, type FlowNode } from "../../../../components/Diagram";
import { CONTEXT } from "../../../../design/tokens";
import { CONTEXTS } from "../../../../data/portage";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 1 — the question.
 *
 * The five names go up unexplained on purpose: the viewer should feel the
 * "why five?" itch before any answer is offered. Their colours are set here
 * and never change again for the rest of the series.
 */
const NODES: FlowNode[] = CONTEXTS.map((c, i) => ({
  id: c.key,
  label: CONTEXT[c.key].label,
  sub: c.short,
  x: i * 205,
  y: 0,
  w: 180,
  h: 130,
  color: CONTEXT[c.key].color,
  at: 40 + i * 14,
}));

export const S1Hook: React.FC = () => {
  return (
    <Stage eyebrow="TẬP 1 · CẢNH 1">
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 140, sound: "reveal" },
        ]}
      />
      <div
        style={{
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          gap: 72,
        }}
      >
        <div style={{ textAlign: "center", maxWidth: 1400 }}>
          <Headline delay={4}>Vì sao project Go này bị chia làm năm?</Headline>
        </div>

        <FlowBoard nodes={NODES} width={1000} height={130} />

        <div style={{ textAlign: "center", maxWidth: 1180 }}>
          <Body delay={140}>
            Năm thư mục dưới internal/domain. Chúng không gọi hàm của nhau và
            không đọc bảng của nhau. Ba phút nữa bạn sẽ thấy vì sao, và lý do
            không nằm trong sách DDD.
          </Body>
        </div>
      </div>
    </Stage>
  );
};
