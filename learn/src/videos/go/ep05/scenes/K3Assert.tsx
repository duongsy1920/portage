import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — buying back the safety the missing keyword took away.
 *
 * Implicit interfaces have one genuine cost: delete a method and the break
 * appears at the call site, in another package, possibly months later. The one
 * line in this scene moves that error back to the file that caused it.
 */
export const K3Assert: React.FC = () => (
  <Stage
    eyebrow="MÙA 1 · TẬP 5 · CẢNH 3 · TỪ VỰNG"
    source={`repo có ${GO_USAGE.assertions} dòng khẳng định như thế này`}
  >
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Một dòng để lỗi hiện ra đúng chỗ gây lỗi
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={780}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.assertion} delay={20} accent={COLOR.ok} highlight={[4]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
              <TermCard
                delay={140}
                accent={COLOR.ok}
                term="var _ X = (*Y)(nil)"
                plain="Bảo trình biên dịch kiểm ngay tại file này: Y có đủ method của X không."
                symfony="implements XInterface — cùng tác dụng, nhưng ở Go nó là tuỳ chọn."
              />
              <Bullets
                gap={20}
                items={[
                  {
                    at: 300,
                    tag: "không có nó",
                    text: "Xoá nhầm một method thì lỗi nổ ở package khác, lúc ai đó gọi — có khi vài tháng sau.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 360,
                    tag: "đọc thế nào",
                    text: "Dấu _ là vứt giá trị đi: mình không cần biến, chỉ cần trình biên dịch kiểm hộ.",
                    accent: COLOR.go,
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
