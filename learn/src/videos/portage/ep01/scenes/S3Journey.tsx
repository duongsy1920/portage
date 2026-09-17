import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import {
  FlowBoard,
  type FlowNode,
  type FlowEdge,
  type FlowPacket,
} from "../../../../components/Diagram";
import { CONTEXT, SAFE } from "../../../../design/tokens";
import { CONTEXTS } from "../../../../data/portage";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 3 — one order actually moving.
 *
 * The packets are the point. "Contexts talk by events" is a sentence a learner
 * can repeat without understanding; a labelled packet leaving one box and
 * landing in the next is the same claim in a form the eye can verify.
 *
 * The diagram runs the full width rather than sitting in a column: five boxes
 * squeezed into half the frame leaves gaps too small for an arrow to read, and
 * the arrow IS the lesson here.
 *
 * Packets fly in a lane above the boxes, because an event name is wider than
 * the gap between two boxes and a label covering the boxes it travels between
 * teaches the wrong picture.
 */
const NODE_W = 250;
const STEP = 362;
const ROW_Y = 150;

const NODES: FlowNode[] = CONTEXTS.map((c, i) => ({
  id: c.key,
  label: CONTEXT[c.key].label,
  sub: c.short,
  x: i * STEP,
  y: ROW_Y,
  w: NODE_W,
  h: 140,
  color: CONTEXT[c.key].color,
  at: 20 + i * 12,
}));

const EDGES: FlowEdge[] = [
  { from: "catalog", to: "pricing", at: 90, dashed: true },
  { from: "pricing", to: "ordering", at: 96, dashed: true },
  { from: "ordering", to: "procurement", at: 102, dashed: true },
  { from: "procurement", to: "logistics", at: 108, dashed: true },
];

/** Real event names, from internal/platform/wire/subscribe.go. */
const LANE = -(ROW_Y - 40);
const PACKETS: FlowPacket[] = [
  { label: "catalog.product_published", from: "catalog", to: "pricing", at: 140, duration: 46, offsetY: LANE },
  { label: "pricing.quote_accepted", from: "pricing", to: "ordering", at: 250, duration: 46, offsetY: LANE },
  { label: "ordering.deposit_paid", from: "ordering", to: "procurement", at: 360, duration: 46, offsetY: LANE },
  { label: "procurement.purchase_confirmed", from: "procurement", to: "logistics", at: 470, duration: 46, offsetY: LANE },
];

export const S3Journey: React.FC = () => {
  return (
    <Stage
      eyebrow="TẬP 1 · CẢNH 3"
      source="Tên event lấy nguyên từ internal/platform/wire/subscribe.go"
    >
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 140, sound: "travel", volume: 0.16 },
          { at: 250, sound: "travel", volume: 0.16 },
          { at: 360, sound: "travel", volume: 0.16 },
          { at: 470, sound: "travel", volume: 0.16 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Một đơn hàng đi qua cả năm vùng
        </Headline>

        <div style={{ marginTop: 26 }}>
          <FlowBoard
            nodes={NODES}
            edges={EDGES}
            packets={PACKETS}
            width={SAFE.contentWidth}
            height={310}
          />
        </div>

        <div style={{ display: "flex", gap: 70, marginTop: 20 }}>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 150,
                  tag: "1",
                  text: "Nhân viên đăng bán một món. Catalog kể lại việc đó, Pricing nghe được nên mới tính giá được.",
                  accent: CONTEXT.catalog.color,
                },
                {
                  at: 260,
                  tag: "2",
                  text: "Khách đồng ý giá. Pricing kể lại, Ordering nghe và mở một đơn hàng.",
                  accent: CONTEXT.pricing.color,
                },
              ]}
            />
          </div>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 370,
                  tag: "3",
                  text: "Khách trả cọc. Ordering kể lại, Procurement nghe và mở một việc đi mua cho người thật.",
                  accent: CONTEXT.ordering.color,
                },
                {
                  at: 480,
                  tag: "4",
                  text: "Mua xong. Procurement kể lại, Logistics nghe và cho kho bên Mỹ chờ một cái thùng.",
                  accent: CONTEXT.procurement.color,
                },
              ]}
            />
          </div>
        </div>
      </div>
    </Stage>
  );
};
