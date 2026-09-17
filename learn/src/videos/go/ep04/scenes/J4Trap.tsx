import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Bullets } from "../../../../components/Bullets";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 4 — the bug this actually causes.
 *
 * Every Go developer writes this bug once: a method on a value receiver that
 * assigns to a field, compiles fine, runs fine, and silently does nothing. It
 * is worth a scene because the compiler will not catch it and neither will a
 * quick reading.
 */
export const J4Trap: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 4 · CẢNH 4" source="lỗi ai học Go cũng viết đúng một lần">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 90, sound: "reveal" },
        { at: 240, sound: "crack", volume: 0.24 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cái bẫy: biên dịch sạch, chạy êm, và không làm gì cả
      </Headline>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 66 }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={34}
          parts={[
            { text: "func (c Counter) Inc() { " },
            { text: "c.n++", note: "sửa vào bản sao, rồi bản sao bị vứt", at: 90, color: COLOR.danger },
            { text: " }" },
          ]}
        />

        <Bullets
          gap={22}
          items={[
            {
              at: 180,
              tag: "chuyện gì xảy ra",
              text: "go build không báo gì. go vet cũng không. Chạy thì c.n vẫn bằng 0 mãi mãi.",
              accent: COLOR.danger,
            },
            {
              at: 240,
              tag: "cách nhận ra",
              text: "Method có gán vào field mà receiver không có dấu * thì gần như chắc chắn là sai.",
              accent: COLOR.money,
            },
          ]}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.php}>
          Bên PHP không có bẫy này, vì object luôn truyền theo tham chiếu — đổi lại,
          bên PHP không có cách nào bảo đảm một method không sửa được <span style={{ color: COLOR.php }}>$this</span>.
        </Callout>
      </div>
    </div>
  </Stage>
);
