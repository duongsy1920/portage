import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { DDD_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — the fix, which is smaller than it sounds.
 *
 * There is no framework here and no new pattern: a method named after the
 * business action, the rule at the top, the event at the bottom. Naming it
 * "rich model" afterwards is fine; leading with the name would not have helped.
 */
export const A3Rich: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 1 · CẢNH 3 · TỪ VỰNG" source="internal/domain/ordering/order.go:225 — code thật">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 310, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Đặt luật vào chính chỗ giữ dữ liệu
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={840}
          gap={52}
          left={<CodeCard snippet={DDD_SNIPPETS.oneDoor} delay={20} accent={COLOR.go} highlight={[2, 3, 6]} scale={0.92} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={150}
                accent={COLOR.go}
                term="mô hình thiếu máu"
                plain="Entity chỉ có getter và setter, còn luật nghiệp vụ nằm rải ở tầng service."
                symfony="Đúng cái Doctrine và form đẩy bạn tới. Nó chạy được, chỉ là luật không ở cùng dữ liệu."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 310,
                    tag: "đặt tên",
                    text: "Method mang tên việc nghiệp vụ: PayDeposit, không phải setStatus.",
                    accent: COLOR.go,
                  },
                  {
                    at: 360,
                    tag: "luật ở trên cùng",
                    text: "Sai trạng thái thì trả error và không đổi gì — không có setter để mà lách.",
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
