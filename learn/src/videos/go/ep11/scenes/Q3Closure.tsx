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
 * Scene 3 — the word "closure", on the one line where it does real work.
 *
 * fn returns only error, so a handler that needs a result cannot get it back
 * as a return value. It gets it back by writing to a variable declared
 * OUTSIDE the function — `var id` above, `id = m.ID()` inside. That is a
 * closure, and it is the whole reason the pattern works.
 */
export const Q3Closure: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 11 · CẢNH 3 · TỪ VỰNG" source="internal/app/catalog/register_merchant.go:31">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 330, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Kết quả thoát ra ngoài bằng một biến bị bắt
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={820}
          gap={48}
          left={<CodeCard snippet={GO_SNIPPETS.closureResult} delay={20} accent={COLOR.money} highlight={[4, 8, 11]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
              <TermCard
                delay={150}
                accent={COLOR.money}
                term="closure"
                plain="Hàm không tên thấy và sửa được biến của hàm bao quanh — cùng một biến, không phải bản sao."
                symfony="function () use (&$id) { … } — Go bắt tự động, không cần use."
              />
              <Bullets
                gap={16}
                items={[
                  {
                    at: 330,
                    tag: "vì sao cần",
                    text: "fn chỉ trả error (Go không có method generic). Kết quả đi ra bằng biến id khai bên ngoài.",
                    accent: COLOR.go,
                  },
                  {
                    at: 390,
                    tag: "đọc thế nào",
                    text: "Biến gán trong func() mà khai bên ngoài — đó là closure đang làm việc.",
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
