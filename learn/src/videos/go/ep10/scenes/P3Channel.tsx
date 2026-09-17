import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 3 — channels taught on their own terms, because the interview asks.
 *
 * The season's honesty rule: what the repository does not use gets taught
 * separately and labelled as such. This scene is explicitly outside Portage —
 * a worker pool, the shape channels are actually for.
 */
const NODES: FlowNode[] = [
  { id: "prod", label: "một goroutine tạo việc", x: 0, y: 60, w: 420, h: 96, color: COLOR.go, at: 20 },
  { id: "ch", label: "chan Job", sub: "hàng đợi có kiểu, an toàn sẵn", x: 620, y: 60, w: 420, h: 96, color: COLOR.money, at: 70 },
  { id: "w1", label: "worker 1", x: 1260, y: 0, w: 420, h: 84, color: COLOR.ok, at: 140 },
  { id: "w2", label: "worker 2", x: 1260, y: 112, w: 420, h: 84, color: COLOR.ok, at: 160 },
];

const EDGES: FlowEdge[] = [
  { from: "prod", to: "ch", at: 110 },
  { from: "ch", to: "w1", at: 200 },
  { from: "ch", to: "w2", at: 220 },
];

const PACKETS: FlowPacket[] = [
  { label: "ch <- job", from: "prod", to: "ch", at: 112, offsetY: -66, color: COLOR.money },
  { label: "job := <-ch", from: "ch", to: "w1", at: 202, offsetY: -66, color: COLOR.ok },
];

export const P3Channel: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 10 · CẢNH 3 · NGOÀI PORTAGE" source="repo không dùng cái này — dạy riêng vì phỏng vấn sẽ hỏi">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 112, sound: "travel", volume: 0.16 },
        { at: 202, sound: "travel", volume: 0.16 },
        { at: 300, sound: "land", volume: 0.24 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Channel hợp khi có việc cần chuyển giao
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={220} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.money}>
          Mũi tên ngược trong <span style={{ color: COLOR.money }}>{"<-ch"}</span> là lấy ra,
          xuôi trong <span style={{ color: COLOR.money }}>{"ch <- job"}</span> là bỏ vào. Channel
          rỗng thì bên lấy đứng chờ — đó vừa là tiện lợi vừa là chỗ hay bị treo.
        </Callout>
      </div>
    </div>
  </Stage>
);
