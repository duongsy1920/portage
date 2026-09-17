import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { Cues } from "../../../../components/Sfx";
import {
  FlowBoard,
  type FlowNode,
  type FlowEdge,
  type FlowPacket,
} from "../../../../components/Diagram";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 1, scene 3 — the same picture in Go, and the shape is different.
 *
 * The boot row happens ONCE, at the top. Requests arrive underneath it and
 * share everything it built. Three packets land on three goroutines while the
 * built world sits still above them — that stillness is the lesson.
 */
const NODES: FlowNode[] = [
  // The boundary is listed FIRST so it paints behind, and it WRAPS the row it
  // describes: the point is that these three boxes are one long-lived thing,
  // not three steps that finish and go away.
  {
    id: "world",
    label: "THẾ GIỚI ĐÃ DỰNG — sống suốt đời tiến trình",
    x: 0,
    y: 0,
    w: 1160,
    h: 205,
    color: COLOR.go,
    kind: "boundary",
    at: 110,
  },
  {
    id: "build",
    label: "go build",
    sub: "một file nhị phân",
    x: 28,
    y: 76,
    w: 240,
    h: 100,
    color: COLOR.textFaint,
    at: 20,
  },
  {
    id: "main",
    label: "main() chạy MỘT LẦN",
    sub: "dựng pool DB, handler, router",
    x: 330,
    y: 76,
    w: 400,
    h: 100,
    color: COLOR.go,
    at: 45,
  },
  {
    id: "serve",
    label: "ListenAndServe",
    sub: "đứng chờ, không chết",
    x: 800,
    y: 76,
    w: 330,
    h: 100,
    color: COLOR.ok,
    at: 75,
  },
  { id: "g1", label: "goroutine #1", x: 1270, y: 0, w: 240, h: 76, color: COLOR.go, at: 200 },
  { id: "g2", label: "goroutine #2", x: 1270, y: 96, w: 240, h: 76, color: COLOR.go, at: 250 },
  { id: "g3", label: "goroutine #3", x: 1270, y: 192, w: 240, h: 76, color: COLOR.go, at: 300 },
];

const EDGES: FlowEdge[] = [
  { from: "build", to: "main", at: 40 },
  { from: "main", to: "serve", at: 70 },
];

const PACKETS: FlowPacket[] = [
  { label: "request 1", from: "serve", to: "g1", at: 190, duration: 34 },
  { label: "request 2", from: "serve", to: "g2", at: 240, duration: 34 },
  { label: "request 3", from: "serve", to: "g3", at: 290, duration: 34 },
];

export const G3Go: React.FC = () => {
  return (
    <Stage eyebrow="MÙA 1 · TẬP 1 · CẢNH 3">
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 45, sound: "appear" },
          { at: 75, sound: "land", volume: 0.3 },
          { at: 110, sound: "reveal" },
          { at: 190, sound: "travel" },
          { at: 240, sound: "travel" },
          { at: 290, sound: "travel" },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Go: dựng thế giới một lần, rồi phục vụ mọi request trên đó
        </Headline>

        <div style={{ marginTop: 26 }}>
          <FlowBoard
            nodes={NODES}
            edges={EDGES}
            packets={PACKETS}
            width={SAFE.contentWidth}
            height={280}
          />
        </div>

        <div style={{ display: "flex", gap: 70, marginTop: 16 }}>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 130,
                  tag: "chú ý",
                  text: "Hàng trên chỉ chạy đúng một lần, lúc khởi động. Không lặp lại cho request nào.",
                  accent: COLOR.go,
                },
                {
                  at: 340,
                  tag: "mỗi request",
                  text: "Go tự sinh một goroutine mới, và nó dùng chung mọi thứ main() đã dựng.",
                  accent: COLOR.go,
                },
              ]}
            />
          </div>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 390,
                  tag: "được gì",
                  text: "Một pool kết nối duy nhất. Cache trong RAM giữ được. Không tốn công dựng lại gì cả.",
                  accent: COLOR.ok,
                },
                {
                  at: 440,
                  tag: "mất gì",
                  text: "Biến chung sống mãi, nên hai request có thể giẫm lên nhau. Và rò rỉ bộ nhớ là chuyện có thật.",
                  accent: COLOR.danger,
                },
              ]}
            />
          </div>
        </div>
      </div>
    </Stage>
  );
};
