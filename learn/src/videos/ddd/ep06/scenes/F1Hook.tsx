import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, CONTEXT, SAFE } from "../../../../design/tokens";

/**
 * Season 2, episode 6, scene 1 — record here, publish there.
 *
 * The split is the entire lesson and it is easy to get wrong: if the aggregate
 * published directly, an event would escape for a change that later rolled
 * back. Drawn as two hops with the transaction boundary in between, because
 * the ordering is what matters.
 */
const NODES: FlowNode[] = [
  { id: "agg", label: "CustomerOrder", sub: "GHI lại: DepositPaid", x: 0, y: 60, w: 430, h: 104, color: CONTEXT.ordering.color, at: 20 },
  { id: "app", label: "tầng app", sub: "lưu đơn, rồi lấy event ra", x: 600, y: 60, w: 430, h: 104, color: COLOR.go, at: 90 },
  { id: "out", label: "outbox", sub: "cùng một transaction với đơn", x: 1230, y: 60, w: 450, h: 104, color: COLOR.money, at: 170 },
];

const EDGES: FlowEdge[] = [
  { from: "agg", to: "app", at: 80 },
  { from: "app", to: "out", at: 160 },
];

const PACKETS: FlowPacket[] = [
  { label: "PullEvents()", from: "agg", to: "app", at: 82, offsetY: -70, color: CONTEXT.ordering.color },
  { label: "Append(events)", from: "app", to: "out", at: 162, offsetY: -70, color: COLOR.money },
];

export const F1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 6" source="quy ước 6 của repo: domain GHI event, không PHÁT event">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 82, sound: "travel", volume: 0.16 },
        { at: 162, sound: "travel", volume: 0.16 },
        { at: 250, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Aggregate ghi lại, tầng app mới phát đi
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={220} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={250} accent={COLOR.money}>
          Nếu aggregate phát thẳng, một thay đổi bị rollback vẫn kịp bắn event ra ngoài —
          và không có cách nào thu lại. Nên nó chỉ ghi vào sổ, còn ai phát là việc của tầng trên.
        </Callout>
      </div>
    </div>
  </Stage>
);
