import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { SketchCard } from "../../../../components/SketchCard";
import { Cues } from "../../../../components/Sfx";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the design that gets proposed instead, and why it loses.
 *
 * "One canonical Variant shared by everyone" is the instinct, so it is drawn as
 * a rejected sketch rather than argued against. The cost is concrete: catalog
 * cannot change a field without a meeting, and ordering carries five fields it
 * has no use for.
 */
const SHARED = [
  "// cách ai cũng thử trước:",
  "package shared",
  "type Variant struct {",
  "    ID, Product shared.ID",
  "    Size, Color, MerchantRef string",
  "    AddedAt time.Time",
  "}",
  "",
  "// mọi context dùng chung một định nghĩa",
];

export const E2Why: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 5 · CẢNH 2 · TỪ VỰNG" source="thiết kế đã từ chối, và lý do">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 310, sound: "crack", volume: 0.2 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Vì sao không gộp thành một định nghĩa chung
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={780}
          gap={52}
          left={<SketchCard label="cách ai cũng thử trước" lines={SHARED} delay={20} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={150}
                accent={COLOR.go}
                term="bounded context"
                plain="Một vùng mà trong đó mỗi từ chỉ có một nghĩa."
                symfony="Gần với một bundle có mô hình riêng — nhưng ở đây ranh giới là package, và có test canh."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 310,
                    tag: "cái giá",
                    text: "Catalog muốn thêm một field thì phải xin phép bốn vùng khác, vì ai cũng đang dùng chung kiểu đó.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 370,
                    tag: "cái giá thứ hai",
                    text: "Ordering chỉ cần biết variant nào của sản phẩm nào, nhưng phải mang theo cả size, màu, mã shop.",
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
