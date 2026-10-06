import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — the one type switch in the repository, with its real first cases.
 *
 * Two things to see: inside a case, `e` already has the concrete type (no
 * cast); and the keys are typed by hand on purpose — json.Marshal of the
 * struct would make a field rename change the wire format while everything
 * still compiles (codec.go's own header says so).
 */
export const R3Switch: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 12 · CẢNH 3" source={`internal/adapter/eventcodec/codec.go:70 — ${GO_USAGE.typeSwitchCases} case`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 340, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        switch e := ev.(type): hỏi một lần, {GO_USAGE.typeSwitchCases} câu trả lời
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={960}
          gap={44}
          left={<CodeCard snippet={GO_SNIPPETS.typeSwitch} delay={20} accent={COLOR.go} highlight={[1, 2, 6]} />}
          right={
            <Bullets
              gap={20}
              items={[
                { at: 150, tag: "trong case", text: "e đã là catalog.MerchantRegistered thật. Đọc e.Name, e.Site ngay — không ép kiểu thêm.", accent: COLOR.go },
                { at: 210, tag: "thay cho gì", text: "Chuỗi if ($ev instanceof A) … elseif (… B). Switch gọn hơn, compiler biết hết các case.", accent: COLOR.php },
                { at: 270, tag: "thiếu case", text: "Thêm event mà quên case: không phải event mất tích — là một test đỏ trong codec_test.go.", accent: COLOR.danger },
                { at: 340, tag: "vì sao viết tay", text: "json.Marshal thẳng struct thì đổi tên field là đổi contract mà vẫn biên dịch. Viết tay, đổi là không biên dịch.", accent: COLOR.money },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
