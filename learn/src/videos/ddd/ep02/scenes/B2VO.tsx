import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the value object, and the thing Go gives away for free.
 *
 * Season 1 already paid for this: a value receiver copies, so immutability is
 * not a discipline here, it is the default. Pointing back at episode 4 is the
 * whole argument, and it is why season 1 came first.
 */
export const B2VO: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 2 · CẢNH 2 · TỪ VỰNG" source="internal/domain/shared/money.go — value object thật">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 310, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Value Object: định nghĩa bằng giá trị, không bằng ID
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={760}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.valueReceiver} delay={20} accent={COLOR.ok} highlight={[0, 2]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={150}
                accent={COLOR.ok}
                term="value object"
                plain="Bằng nhau khi mọi field bằng nhau. Không có ID, không sửa được sau khi tạo."
                symfony="Brick\\Money, hoặc một class có #[Embeddable] và không có setter."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 310,
                    tag: "go cho không",
                    text: "Receiver giá trị chép cả struct, nên bất biến mà không phải cố gắng gì.",
                    accent: COLOR.go,
                  },
                  {
                    at: 370,
                    tag: "trong repo",
                    text: "Money, Weight, Percent, Currency, ParcelSpec, ExchangeRate — tất cả đều không có ID.",
                    accent: COLOR.ok,
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
