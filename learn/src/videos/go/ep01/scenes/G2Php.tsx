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
 * Season 1, scene 2 — the PHP-FPM life the viewer already lives in.
 *
 * Shown first and shown accurately, because the whole episode is a delta and a
 * delta needs a correct starting point. The important beat is the skull: the
 * worker does not keep anything, so nothing the request built survives it.
 */
const ROW = 150;
const NODES: FlowNode[] = [
  { id: "req", label: "request", x: 0, y: ROW, w: 200, h: 110, color: COLOR.textFaint, at: 20 },
  { id: "fpm", label: "php-fpm worker", x: 260, y: ROW, w: 260, h: 110, color: COLOR.php, at: 45 },
  {
    id: "boot",
    label: "Kernel::boot()",
    sub: "autoload, dựng container",
    x: 580,
    y: ROW,
    w: 300,
    h: 110,
    color: COLOR.php,
    at: 70,
  },
  {
    id: "ctrl",
    label: "Controller",
    sub: "Doctrine, response",
    x: 940,
    y: ROW,
    w: 280,
    h: 110,
    color: COLOR.php,
    at: 95,
  },
  {
    id: "dead",
    label: "💀 TIẾN TRÌNH CHẾT",
    sub: "mọi biến biến mất",
    x: 1280,
    y: ROW,
    w: 320,
    h: 110,
    color: COLOR.danger,
    at: 150,
  },
];

const EDGES: FlowEdge[] = [
  { from: "req", to: "fpm", at: 40 },
  { from: "fpm", to: "boot", at: 65 },
  { from: "boot", to: "ctrl", at: 90 },
  { from: "ctrl", to: "dead", at: 145, color: COLOR.danger },
];

const PACKETS: FlowPacket[] = [
  { label: "request 2", from: "req", to: "fpm", at: 300, duration: 40, offsetY: -120, color: COLOR.php },
];

export const G2Php: React.FC = () => {
  return (
    <Stage eyebrow="MÙA 1 · TẬP 1 · CẢNH 2">
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 45, sound: "appear" },
          { at: 70, sound: "appear" },
          { at: 95, sound: "appear" },
          { at: 150, sound: "crack", volume: 0.35 },
          { at: 300, sound: "travel" },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Thế giới bạn đang sống: một request, một đời
        </Headline>

        <div style={{ marginTop: 30 }}>
          <FlowBoard
            nodes={NODES}
            edges={EDGES}
            packets={PACKETS}
            width={SAFE.contentWidth}
            height={290}
          />
        </div>

        <div style={{ display: "flex", gap: 70, marginTop: 20 }}>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 190,
                  tag: "mỗi lần",
                  text: "Nạp autoload, dựng lại container, dựng lại mọi service. Xong request thì bỏ hết.",
                  accent: COLOR.php,
                },
                {
                  at: 240,
                  tag: "được gì",
                  text: "Không bao giờ rò rỉ bộ nhớ, không bao giờ có hai request giẫm lên nhau.",
                  accent: COLOR.ok,
                },
              ]}
            />
          </div>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 310,
                  tag: "mất gì",
                  text: "Request thứ hai không nhớ gì từ request thứ nhất. Muốn nhớ phải nhờ Redis hay database.",
                  accent: COLOR.money,
                },
                {
                  at: 360,
                  text: "Và mỗi request lại mở kết nối database từ đầu, hoặc phải có một pool nằm ngoài PHP.",
                  accent: COLOR.money,
                },
              ]}
            />
          </div>
        </div>
      </div>
    </Stage>
  );
};
