import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { DDD_SNIPPETS } from "../../../../data/portage";
import { COLOR, CONTEXT } from "../../../../design/tokens";

const ORDERING = CONTEXT.ordering.color;

/**
 * Scene 2 — invariant, defined by what it is not.
 *
 * "Something that must always be true" is too vague to act on. The operational
 * version is: it is the thing that decides where the boundary goes. If two
 * facts must agree at the same instant, they are in one aggregate; if they can
 * disagree for a second, they are not.
 */
export const C2Invariant: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 3 · CẢNH 2 · TỪ VỰNG" source="internal/domain/ordering/order.go:225">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 310, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Invariant là thứ quyết định ranh giới nằm ở đâu
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={840}
          gap={52}
          left={<CodeCard snippet={DDD_SNIPPETS.oneDoor} delay={20} accent={ORDERING} highlight={[2, 3]} scale={0.92} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
              <TermCard
                delay={150}
                accent={ORDERING}
                term="invariant"
                plain="Một điều phải đúng ngay lập tức, không được phép sai dù chỉ một khoảnh khắc."
                symfony="Thường rơi vào Validator — mà Validator chạy trước khi ghi, không phải lúc ghi."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 310,
                    tag: "cách chia",
                    text: "Hai điều phải đúng cùng lúc thì ở chung. Cho phép lệch một lúc thì tách ra.",
                    accent: ORDERING,
                  },
                  {
                    at: 370,
                    tag: "hệ quả",
                    text: "Một transaction một aggregate. Cần sửa hai cái là chia sai, hoặc cần event.",
                    accent: COLOR.money,
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
