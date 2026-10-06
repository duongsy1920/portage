import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — the word "Published Language", by its mechanism.
 *
 * Season 2 named the concept; this scene shows what it is made of in Go:
 * a package of plain structs with tags and a version in every type name,
 * imported by the codec and by consumers, never by the domain.
 */
export const U3Published: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 15 · CẢNH 3 · TỪ VỰNG" source={`internal/contracts — ${GO_USAGE.contractTypes} kiểu V1`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 340, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Published Language: contract là struct riêng, có tag, có version
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={820}
          gap={52}
          left={
            <TermCard
              delay={20}
              accent={COLOR.money}
              term="Published Language"
              plain="Hình dạng dữ liệu đi qua ranh giới hai context: chỉ kiểu nguyên thuỷ và V1 khác, không bao giờ kiểu domain. Đổi nghĩa là kiểu mới."
              symfony="Một package Messages/DTO đánh version dùng chung giữa các bundle."
              inRepo="internal/contracts — package riêng"
            />
          }
          right={
            <Bullets
              gap={20}
              items={[
                { at: 150, tag: "vì sao không marshal domain", text: "Đổi tên một field của OrderPlaced là đổi JSON trên dây — mà mọi thứ vẫn biên dịch. Hợp đồng vỡ im lặng.", accent: COLOR.danger },
                { at: 220, tag: "struct riêng thì", text: "Đổi domain, contract không đổi. Đổi contract phải sửa codec (tập 12) — không biên dịch cho tới khi sửa xong.", accent: COLOR.go },
                { at: 290, tag: "V1 trong tên", text: "Thêm field có mặc định: vẫn V1, người đọc cũ bỏ qua. Đổi nghĩa một field: kiểu V2 mới, V1 vẫn sống.", accent: COLOR.money },
                { at: 360, tag: "guard 7", text: "Domain không import context khác, nên người nhận chỉ biết đúng những gì nằm trong JSON. Contract là tất cả họ có.", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
