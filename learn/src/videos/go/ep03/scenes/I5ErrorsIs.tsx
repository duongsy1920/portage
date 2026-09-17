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
 * Scene 5 — reading the cause back out, through however many wraps.
 *
 * The example is a real one and it is not tidy: accepting an expired quote
 * changes the quote (issued → expired) and still refuses. That is why the code
 * saves first and reports after. Showing the messy real case beats a clean
 * invented one, because the viewer will meet the messy case.
 */
export const I5ErrorsIs: React.FC = () => (
  <Stage
    eyebrow="MÙA 1 · TẬP 3 · CẢNH 5 · TỪ VỰNG"
    source={`internal/app/pricing/accept_quote.go:35 — repo có ${GO_USAGE.errorsIs} chỗ`}
  >
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        errors.Is: hỏi “gốc của nó có phải cái này không?”
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={740}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.errorsIs} delay={20} accent={COLOR.go} highlight={[1]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
              <TermCard
                delay={140}
                accent={COLOR.go}
                term="errors.Is"
                plain="Lần ngược chuỗi %w để xem nguyên nhân gốc có đúng cái lỗi mình đang hỏi không."
                symfony="catch (QuoteExpired $e) — cũng là hỏi “lỗi này loại nào”, nhưng bằng if."
              />
              <Bullets
                gap={15}
                items={[
                  {
                    at: 300,
                    tag: "vì sao không ==",
                    text: "So sánh bằng == chỉ đúng khi lỗi chưa bị bọc. Bọc một lần là == sai ngay.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 350,
                    tag: "ví dụ thật",
                    text: "Báo giá hết hạn vẫn phải lưu lại rồi mới từ chối. Nên mới cần biết đây là lỗi nào.",
                    accent: COLOR.money,
                  },
                  {
                    at: 410,
                    tag: "còn một cái nữa",
                    text: "errors.As dùng khi cần lấy chính đối tượng lỗi ra để đọc field.",
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
