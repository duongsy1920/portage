import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — panic is not an exception, and the difference costs money.
 *
 * In PHP a fatal error kills one request and the next one is unaffected. Here
 * the process serves everyone, so an unrecovered panic takes every in-flight
 * order down with it. That is the whole reason episode 1 had to come first.
 */
const NODES: FlowNode[] = [
  { id: "p", label: "panic ở một request", x: 0, y: 40, w: 340, h: 96, color: COLOR.danger, at: 40 },
  { id: "php", label: "PHP: chết một process", sub: "request sau vẫn chạy", x: 460, y: 0, w: 420, h: 92, color: COLOR.php, at: 110 },
  { id: "go", label: "Go: chết cả tiến trình", sub: "mọi request đang dở mất", x: 460, y: 128, w: 420, h: 92, color: COLOR.danger, at: 160 },
];

const EDGES: FlowEdge[] = [
  { from: "p", to: "php", at: 100 },
  { from: "p", to: "go", at: 150 },
];

export const O3Panic: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 9 · CẢNH 3 · TỪ VỰNG" source="repo có 33 chỗ panic — tất cả đều là lỗi lập trình viên">
    <Cues items={[{ at: 40, sound: "appear" }, { at: 160, sound: "crack", volume: 0.24 }, { at: 260, sound: "reveal" }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        panic không phải exception
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="center"
          leftWidth={880}
          gap={52}
          left={<FlowBoard nodes={NODES} edges={EDGES} width={880} height={240} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
              <TermCard
                delay={260}
                accent={COLOR.danger}
                term="panic"
                plain="Dừng hàm và đi ngược lên, chạy mọi defer trên đường. Không ai chặn thì cả tiến trình chết."
                symfony="Gần fatal error hơn exception — và ở Go nó giết cả tiến trình."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 360,
                    tag: "dùng khi nào",
                    text: "Chỉ khi lập trình viên sai. Người dùng sai thì trả error — luật 1 của repo.",
                    accent: COLOR.money,
                  },
                  {
                    at: 420,
                    tag: "trong repo",
                    text: "33 chỗ panic, gần hết nằm trong constructor Must* và hàm dựng lúc chạy.",
                    accent: COLOR.go,
                  },
                ]}
              />
            </div>
          }
        />
      </div>
    </div>
  </Stage>
);
