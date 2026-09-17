import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Body, Callout } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 6 — the line that looks hardest in the whole file, taken apart.
 *
 * Everything needed to read it has now been said, so this scene proves it
 * rather than teaching anything new. Two pieces point forward on purpose: the
 * receiver to episode 4, the second return value to episode 3. Saying "that one
 * is later" is more honest than a rushed half-explanation, and it also tells
 * the viewer the roadmap is real.
 */
export const H6Hardest: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 2 · CẢNH 6" source="internal/domain/shared/money.go:201">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 80, sound: "reveal" },
        { at: 260, sound: "reveal" },
        { at: 420, sound: "land", volume: 0.26 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Dòng khó nhất file — thật ra chỉ có bốn mảnh
      </Headline>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 58 }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "func " },
            { text: "(m Money)", note: "gắn hàm vào kiểu Money. Mở ra ở tập 4.", at: 80, color: COLOR.ok },
            { text: " " },
            { text: "Add", note: "tên hàm. A viết HOA nên ai cũng gọi được.", at: 160, color: COLOR.go },
            { text: "(" },
            { text: "o Money", note: "tham số: tên o, kiểu Money.", at: 260, color: COLOR.go },
            { text: ") " },
            { text: "(Money, error)", note: "trả về hai thứ: kết quả và lỗi. Tập 3.", at: 350, color: COLOR.money },
            { text: " {" },
          ]}
        />

        <Body delay={430} size={40}>
          Đọc thành lời:{" "}
          <span style={{ color: COLOR.text, fontWeight: 600 }}>
            “Hàm Add, thuộc về Money, nhận một Money tên o, trả về một Money và một error.”
          </span>
        </Body>
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={480} accent={COLOR.ok}>
          Không có gì mới trong dòng này. Chỉ là bốn luật vừa học, xếp cạnh nhau.
        </Callout>
      </div>
    </div>
  </Stage>
);
