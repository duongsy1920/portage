import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Body } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Bullets } from "../../../../components/Bullets";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — what `(Money, error)` actually means at the call site.
 *
 * Episode 2 pointed at this pair and deferred it. Paying that debt in the first
 * two minutes of episode 3 is the point: a roadmap that promises and delivers
 * is the reason a viewer keeps watching.
 */
export const I2TwoValues: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 3 · CẢNH 2" source="cách gọi một hàm trả hai giá trị">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 80, sound: "reveal" }, { at: 230, sound: "snap", volume: 0.18 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Một hàm trả về hai thứ, và bạn phải nhận cả hai
      </Headline>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 62 }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "m, err" , note: "hai biến, nhận hai giá trị trả về", at: 80, color: COLOR.go },
            { text: " := " },
            { text: "ParseMoney" },
            { text: '("150.00", USD)' },
          ]}
        />

        <Body delay={150} size={36}>
          Trình biên dịch <span style={{ color: COLOR.danger, fontWeight: 600 }}>từ chối build</span> nếu
          bạn khai báo <span style={{ color: COLOR.go }}>err</span> rồi không dùng tới nó.
        </Body>

        <Bullets
          gap={20}
          items={[
            {
              at: 230,
              tag: "muốn bỏ qua",
              text: "m, _ := ParseMoney(...) — dấu gạch dưới là lời khai: tôi biết có lỗi, tôi cố tình vứt.",
              accent: COLOR.money,
            },
            {
              at: 290,
              tag: "khác gì php",
              text: "Bỏ quên catch thì không ai biết. Ở đây bỏ qua lỗi là một ký tự người review nhìn thấy ngay.",
              accent: COLOR.php,
            },
          ]}
        />
      </div>
    </div>
  </Stage>
);
