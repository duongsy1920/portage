import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — the payoff, in the repository's own code.
 *
 * Nine typed ids exist because passing a QuoteID where an OrderID belongs is a
 * bug that costs real money in this business, and embedding makes each of them
 * one line instead of forty. This is the moment the syntax becomes a decision.
 */
export const M4TypedID: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 7 · CẢNH 4" source="repo có 9 kiểu ID dựng kiểu này và 8 chỗ nhúng shared.Events">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 130, sound: "reveal" }, { at: 300, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Nhờ nó mà có OrderID, QuoteID, ParcelID
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={740}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.typedID} delay={20} accent={COLOR.go} highlight={[1]} />}
          right={
            <Bullets
              gap={22}
              items={[
                {
                  at: 130,
                  tag: "được gì",
                  text: "Truyền nhầm QuoteID vào chỗ cần OrderID thì không biên dịch được. Cả hai vẫn chỉ là một uuid.",
                  accent: COLOR.ok,
                },
                {
                  at: 200,
                  tag: "giá bao nhiêu",
                  text: "Một dòng. Mọi method của shared.ID (String, Equal, MarshalJSON) tự dùng được.",
                  accent: COLOR.go,
                },
                {
                  at: 270,
                  tag: "vì sao đáng",
                  text: "Đây là hệ mua hộ có tiền thật. Gán nhầm một id là gán nhầm tiền cọc của khách khác.",
                  accent: COLOR.money,
                },
                {
                  at: 340,
                  tag: "bên symfony",
                  text: "Cũng làm được bằng class UuidValueObject riêng cho từng loại, nhưng phải viết lại constructor và các method mỗi lần.",
                  accent: COLOR.php,
                },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
