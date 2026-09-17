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
 * Scene 3 — the price of no null, and who pays it.
 *
 * Because `Currency{}` is a legal value, nothing stops a caller from building
 * money in a currency that does not exist. Go will not catch it, so the
 * constructor has to. This is where the repository's convention "programmer
 * mistake → panic" comes from, and it is worth naming out loud.
 */
export const L3Guard: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 6 · CẢNH 3 · TỪ VỰNG" source="internal/domain/shared/money.go:71 và currency.go:87">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 310, sound: "crack", volume: 0.2 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Zero value hợp lệ, nên constructor phải tự vệ
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={820}
          gap={52}
          left={<CodeCard snippet={GO_SNIPPETS.zeroGuard} delay={20} accent={COLOR.money} highlight={[1, 2]} scale={0.88} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={150}
                accent={COLOR.money}
                term="IsZero()"
                plain="Hỏi xem giá trị này có phải bản mặc định chưa ai gán không."
                symfony="if ($c === null) — ở đây không có null, nên phải so với Currency{}."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 310,
                    tag: "cái giá",
                    text: "Currency{} là giá trị hợp lệ, nên không ai chặn được việc tạo tiền bằng loại tiền không tồn tại.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 370,
                    tag: "luật của repo",
                    text: "Lập trình viên sai thì panic, người dùng nhập sai thì trả error — dòng panic kia là loại đầu.",
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
