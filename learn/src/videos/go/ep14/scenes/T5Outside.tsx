import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge, type FlowPacket } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 5 — the three pattern names, outside Portage and labelled so.
 *
 * Same honesty rule as episode 10: the repository has no fan-out, no fan-in,
 * no pipeline, because no job in it needs them. The interview will still ask,
 * so the shapes are drawn once with their names, and nothing more.
 */
const NODES: FlowNode[] = [
  { id: "src", label: "nguồn việc", x: 0, y: 70, w: 300, h: 92, color: COLOR.go, at: 20 },
  { id: "jobs", label: "chan Job", sub: "fan-out: một kênh, nhiều người lấy", x: 400, y: 70, w: 380, h: 92, color: COLOR.money, at: 60 },
  { id: "w1", label: "worker 1", x: 880, y: 0, w: 260, h: 70, color: COLOR.ok, at: 120 },
  { id: "w2", label: "worker 2", x: 880, y: 81, w: 260, h: 70, color: COLOR.ok, at: 135 },
  { id: "w3", label: "worker 3", x: 880, y: 162, w: 260, h: 70, color: COLOR.ok, at: 150 },
  { id: "out", label: "chan Result", sub: "fan-in: nhiều người ghi, một kênh", x: 1240, y: 70, w: 400, h: 92, color: COLOR.money, at: 200 },
];

const EDGES: FlowEdge[] = [
  { from: "src", to: "jobs", at: 50 },
  { from: "jobs", to: "w1", at: 110 },
  { from: "jobs", to: "w2", at: 125 },
  { from: "jobs", to: "w3", at: 140 },
  { from: "w1", to: "out", at: 190 },
  { from: "w2", to: "out", at: 205 },
  { from: "w3", to: "out", at: 220 },
];

const PACKETS: FlowPacket[] = [
  { label: "jobs <- j", from: "src", to: "jobs", at: 52, offsetY: -64, color: COLOR.money },
  { label: "j := <-jobs", from: "jobs", to: "w2", at: 150, offsetY: -60, color: COLOR.ok },
  { label: "out <- r", from: "w2", to: "out", at: 230, offsetY: -60, color: COLOR.money },
];

export const T5Outside: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 14 · CẢNH 5 · NGOÀI PORTAGE" source="repo không có ba mẫu này — dạy riêng vì phỏng vấn sẽ hỏi">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 52, sound: "travel", volume: 0.16 },
        { at: 150, sound: "travel", volume: 0.16 },
        { at: 230, sound: "travel", volume: 0.16 },
        { at: 330, sound: "land", volume: 0.24 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Fan-out, fan-in, pipeline: ba cái tên, một hình
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} packets={PACKETS} width={SAFE.contentWidth} height={240} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={330} accent={COLOR.money}>
          Fan-out: nhiều goroutine cùng lấy từ một channel. Fan-in: nhiều goroutine cùng ghi vào một channel. Pipeline: nối các
          bước bằng channel, mỗi bước một goroutine. Repo chưa có việc nào cần — relay của nó là một goroutine, đủ cho một kho.
        </Callout>
      </div>
    </div>
  </Stage>
);
