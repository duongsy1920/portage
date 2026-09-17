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
 * Scene 3 — a named error per reason.
 *
 * This is where the pattern stops being a syntax quirk and starts being
 * design: the HTTP layer answers 409 instead of 500 because the domain said
 * which refusal it was. A viewer who has written typed exceptions already
 * knows the idea; only the spelling is new.
 */
export const I3Sentinel: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 3 · CẢNH 3 · TỪ VỰNG" source="internal/domain/ordering/errors.go:7">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Mỗi lý do từ chối một cái tên
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={760}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.sentinels} delay={20} accent={COLOR.danger} highlight={[2, 4]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
              <TermCard
                delay={140}
                accent={COLOR.danger}
                term="sentinel error"
                plain="Một biến lỗi đặt tên sẵn ở cấp package, để nơi khác so sánh đúng lý do."
                symfony="final class QuoteNotAccepted extends DomainException — cùng vai trò."
              />
              <Bullets
                gap={20}
                items={[
                  {
                    at: 300,
                    tag: "quy ước",
                    text: "Tên luôn bắt đầu bằng Err, chữ E viết hoa nên package khác so sánh được.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 350,
                    tag: "để làm gì",
                    text: "Tầng HTTP đọc được lý do nên trả 409 thay vì 500.",
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
