import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the rule that follows from it, and the word for the boundary.
 *
 * Two anti-corruption layers exist here: one wraps a model that may be wrong,
 * one wraps shops that have no API. Both translate an outside world that does
 * not share this domain's vocabulary — and both stop short of deciding
 * anything, which is the part that matters when money is involved.
 */
const NODES: FlowNode[] = [
  { id: "ai", label: "máy đọc URL", sub: "có thể sai", x: 0, y: 30, w: 360, h: 88, color: COLOR.money, at: 40 },
  { id: "draft", label: "tạo bản nháp", sub: "được phép", x: 480, y: 30, w: 320, h: 88, color: COLOR.ok, at: 110 },
  { id: "conf", label: "xác nhận để bán", sub: "không được phép", x: 480, y: 150, w: 320, h: 88, color: COLOR.danger, at: 180 },
  { id: "human", label: "người xác nhận", sub: "và chịu trách nhiệm", x: 0, y: 150, w: 360, h: 88, color: COLOR.go, at: 230 },
];

const EDGES: FlowEdge[] = [
  { from: "ai", to: "draft", at: 100, color: COLOR.ok },
  { from: "ai", to: "conf", at: 170, dashed: true, color: COLOR.danger },
  { from: "human", to: "conf", at: 260, color: COLOR.go },
];

export const P8S2: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 8 · CẢNH 2 · TỪ VỰNG" source="catalogapp.ListingExtractor và procurement.MerchantACL">
    <Cues items={[{ at: 40, sound: "appear" }, { at: 170, sound: "crack", volume: 0.2 }, { at: 280, sound: "reveal" }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Máy được tạo nháp, không được xác nhận
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="center"
          leftWidth={820}
          gap={52}
          left={<FlowBoard nodes={NODES} edges={EDGES} width={820} height={240} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
              <TermCard
                delay={280}
                accent={COLOR.money}
                term="anti-corruption layer"
                plain="Lớp dịch giữa thế giới bên ngoài và domain, để từ vựng của họ không lọt vào mình."
                symfony="Một Gateway bọc SDK bên thứ ba — nhưng ở đây nó giữ cả ranh giới trách nhiệm."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 380,
                    tag: "hai lớp",
                    text: "Một lớp bọc mô hình đọc URL, một lớp bọc shop không có API. Cả hai chỉ tạo nháp.",
                    accent: COLOR.money,
                  },
                  {
                    at: 440,
                    tag: "vì sao dừng ở nháp",
                    text: "Xác nhận là lúc có người chịu trách nhiệm. Máy đoán sai cỡ giày là mua nhầm hàng.",
                    accent: COLOR.danger,
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
