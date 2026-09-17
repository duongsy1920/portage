import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { SketchCard } from "../../../../components/SketchCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — where the `extends` reading breaks.
 *
 * The sketch card is the right component here precisely because the code on it
 * is wrong: it is the thing a Symfony developer expects to be able to write,
 * shown as a rejected design rather than as a lesson in syntax.
 */
const WRONG = [
  "// mong đợi kiểu PHP:",
  "func handle(o CustomerOrder) { … }",
  "handle(parcel)  // Parcel cũng nhúng shared.Events",
  "",
  "// Go: không biên dịch được.",
  "// Nhúng chung một kiểu không làm hai kiểu",
  "// thành họ hàng với nhau.",
];

export const M3NotInherit: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 7 · CẢNH 3 · TỪ VỰNG" source="chỗ cách đọc “kế thừa” gãy">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 300, sound: "crack", volume: 0.2 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Nhưng nó không phải kế thừa
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={800}
          gap={52}
          left={<SketchCard label="điều bạn sẽ thử, và nó hỏng" lines={WRONG} delay={20} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={150}
                accent={COLOR.money}
                term="embedding"
                plain="Nhúng một kiểu vào struct mà không đặt tên field, để method của nó gọi thẳng được."
                symfony="Gần với trait hơn là extends: mượn hành vi, không tạo quan hệ cha con."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 300,
                    tag: "không có",
                    text: "Không có đa hình theo lớp cha. Muốn nhận nhiều kiểu thì phải dùng interface.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 360,
                    tag: "đổi lại",
                    text: "Không có chuỗi kế thừa nhiều tầng để mà lạc. Muốn biết method ở đâu thì nhìn xuống đúng một tầng.",
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
