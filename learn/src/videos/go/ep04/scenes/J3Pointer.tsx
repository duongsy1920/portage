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
 * Scene 3 — the pointer receiver, and why an aggregate needs one.
 *
 * Events is the perfect example because it is the one thing in the shared
 * package that is deliberately NOT a value: it is bookkeeping, it must
 * accumulate, and accumulating is exactly what a copy cannot do.
 */
export const J3Pointer: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 4 · CẢNH 3 · TỪ VỰNG" source="internal/domain/shared/event.go:78">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Có dấu sao: method sửa được bản gốc
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={720}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.pointerReceiver} delay={20} accent={COLOR.money} highlight={[0, 1]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
              <TermCard
                delay={140}
                accent={COLOR.money}
                term="pointer receiver"
                plain="Dấu * nghĩa là method nhận địa chỉ của bản gốc, nên sửa vào đó là sửa thật."
                symfony="Đúng bằng hành vi mặc định của object PHP: $this luôn là bản gốc."
              />
              <Bullets
                gap={20}
                items={[
                  {
                    at: 300,
                    tag: "vì sao ở đây",
                    text: "Events phải cộng dồn. Cộng dồn vào một bản sao thì sự kiện mất khi hàm kết thúc.",
                    accent: COLOR.money,
                  },
                  {
                    at: 360,
                    tag: "quy tắc",
                    text: "Value object dùng (m Money). Thứ có trạng thái thay đổi dùng (e *Events).",
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
