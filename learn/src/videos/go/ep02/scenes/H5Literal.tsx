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
 * Scene 5 — making one, without `new`.
 *
 * Two things land together because they always appear together in real code:
 * the composite literal `Money{...}` and the short declaration `:=`. Both are
 * everywhere in the repository, and both look like typos until named.
 */
export const H5Literal: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 2 · CẢNH 5 · TỪ VỰNG" source="internal/domain/shared/money.go:71">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Luật bốn: tạo một cái mới, không có từ new
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={700}
          gap={60}
          left={<CodeCard snippet={GO_SNIPPETS.newMoney} delay={20} accent={COLOR.go} highlight={[4]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={140}
                accent={COLOR.go}
                term="composite literal"
                plain="Money{minor: 5, currency: usd} vừa tạo vừa gán field, ngay tại chỗ."
                symfony="new Money(minor: 5, currency: $usd) — named arguments, bỏ chữ new đi là ra."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 300,
                    tag: "không có",
                    text: "Go không có __construct. Quy ước là một hàm thường tên NewMoney.",
                    accent: COLOR.go,
                  },
                  {
                    at: 350,
                    tag: "hay gặp",
                    text: "m := NewMoney(5, usd) — dấu := vừa khai báo biến vừa gán, kiểu tự suy ra.",
                    accent: COLOR.go,
                  },
                  {
                    at: 400,
                    tag: "chỉ trong hàm",
                    text: "Ngoài hàm thì phải viết var. Đó là lý do ErrCurrencyMismatch dùng var chứ không dùng :=.",
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
