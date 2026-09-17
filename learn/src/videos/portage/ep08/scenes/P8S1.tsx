import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 3, episode 8, scene 1 — the legal constraint, drawn as a boundary.
 *
 * Nike's terms forbid scraping, and most shops say the same. So the system
 * never fetches a shop page: a person pastes the details, or an adapter reads
 * a URL through a model. This is a design constraint with a lawyer behind it,
 * not a footnote — and it is why the machine may draft but never confirm.
 */
const NODES: FlowNode[] = [
  { id: "shop", label: "trang shop Mỹ", sub: "điều khoản cấm cào", x: 0, y: 100, w: 400, h: 100, color: COLOR.danger, at: 20 },
  { id: "acl", label: "adapter/openai", sub: "đọc từ URL người dùng đưa", x: 560, y: 30, w: 440, h: 92, color: COLOR.money, at: 90 },
  { id: "man", label: "người nhập tay", sub: "khách hoặc nhân viên", x: 560, y: 170, w: 440, h: 92, color: COLOR.go, at: 130 },
  { id: "draft", label: "bản nháp trong catalog", sub: "chưa ai mua được", x: 1160, y: 100, w: 520, h: 100, color: COLOR.textDim, at: 200 },
];

const EDGES: FlowEdge[] = [
  { from: "shop", to: "acl", at: 180, dashed: true, color: COLOR.danger },
  { from: "acl", to: "draft", at: 190 },
  { from: "man", to: "draft", at: 210 },
];

export const P8S1: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 8" source="ràng buộc pháp lý: hệ thống không tự fetch trang shop">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 180, sound: "crack", volume: 0.2 }, { at: 280, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Hệ thống không được phép tự đọc trang shop
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <FlowBoard nodes={NODES} edges={EDGES} width={SAFE.contentWidth} height={290} />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={280} accent={COLOR.danger}>
          Mũi tên đứt kia là thứ không được làm. Điều khoản của Nike và phần lớn shop cấm
          cào trang — nên dữ liệu vào hệ thống bằng tay người, hoặc qua một adapter đọc URL.
        </Callout>
      </div>
    </div>
  </Stage>
);
