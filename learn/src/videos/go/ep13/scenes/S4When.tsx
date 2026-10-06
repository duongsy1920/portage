import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — the question an interview actually asks: generic, or interface?
 *
 * The repository answers it by example. Repositories are interfaces because
 * callers want BEHAVIOUR (Save, ByID) and do not care which type implements
 * it. on[T] is generic because the caller wants the SAME type to come out
 * that went in. Constraints are introduced here because `any` is the only
 * one the repo uses, and that too is worth saying.
 */
export const S4When: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 13 · CẢNH 4 · TỪ VỰNG" source="chọn theo câu hỏi, không theo mốt">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 360, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Interface: hành vi. Generic: giữ nguyên kiểu qua một hàm.
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={780}
          gap={52}
          left={
            <TermCard
              delay={20}
              accent={COLOR.money}
              term="type constraint"
              plain="Điều kiện cho T. any: gì cũng được. comparable: so == được. Một interface: T phải có method đó."
              symfony="@template T of Foo — lại là chú thích; Go từ chối biên dịch nếu T không thoả."
              inRepo="cả hai hàm của repo: [T any]"
            />
          }
          right={
            <Bullets
              gap={20}
              items={[
                { at: 140, tag: "dùng interface khi", text: "Người gọi cần hành vi: Save, ByID, Now — không cần biết kiểu thật. Repository, Clock (tập 5).", accent: COLOR.go },
                { at: 210, tag: "dùng generic khi", text: "Kiểu đi vào phải bằng kiểu đi ra: handler nhận T thì nhận đúng T, không phải any phải hỏi lại.", accent: COLOR.money },
                { at: 280, tag: "đừng dùng khi", text: "Chỉ có một kiểu. Một hàm cho []Quote không cần [T any] — đó là generic vì mốt.", accent: COLOR.danger },
                { at: 360, tag: "câu phỏng vấn", text: "“Go 1.18 có generics rồi, interface còn cần không?” — Còn. Hai câu hỏi khác nhau: làm được gì, và là kiểu gì.", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={420} accent={COLOR.go}>
          Một repo ba mươi use case, hai hàm generic. Con số đó là câu trả lời thật thà nhất về generics.
        </Callout>
      </div>
    </div>
  </Stage>
);
