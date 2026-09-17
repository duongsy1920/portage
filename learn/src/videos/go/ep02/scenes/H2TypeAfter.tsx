import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Callout } from "../../../../components/Text";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — type after name.
 *
 * The rule is trivial; the reason is not, and the reason is what makes it
 * stick. Go reads left to right: the thing being named first, what it is
 * second. A function signature obeys the same order, which is why the return
 * type is at the far end instead of in front.
 */
export const H2TypeAfter: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 2 · CẢNH 2" source="internal/domain/shared/money.go:57 và :71">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 250, sound: "reveal" }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Luật một: tên trước, kiểu sau
      </Headline>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 74 }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "minor", note: "tên field — cái đang được đặt tên", at: 90, color: COLOR.go },
            { text: "    " },
            { text: "int64", note: "kiểu của nó. Đọc: “minor thuộc kiểu int64”.", at: 150, color: COLOR.money },
          ]}
        />
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "func NewMoney(" },
            { text: "minor int64, c Currency", note: "tham số — vẫn tên trước, kiểu sau.", at: 250, color: COLOR.go },
            { text: ") " },
            { text: "Money", note: "kiểu trả về — nằm ở cuối chữ ký.", at: 310, color: COLOR.ok },
            { text: " {" },
          ]}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={380} accent={COLOR.php}>
          Cả dòng đọc xuôi một chiều: đang nói về cái gì, rồi cái đó là gì. PHP thì
          ngược — <span style={{ color: COLOR.php }}>private int $minor</span> nói kiểu trước, tên sau.
        </Callout>
      </div>
    </div>
  </Stage>
);
