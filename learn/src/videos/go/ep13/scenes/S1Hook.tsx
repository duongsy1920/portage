import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 13, scene 1 — two generic functions, counted.
 *
 * The number is the lesson. A learner coming from PHP will have heard that
 * generics are the big thing Go got in 1.18; a real codebase with thirty use
 * cases needs them twice, and both times for the same job: getting a typed
 * value back out of bytes.
 */
export const S1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 13" source="internal/adapter/eventcodec/decode.go:19 — code thật">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 120, sound: "reveal" }, { at: 260, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Repo có đúng {GO_USAGE.generics} hàm generic. Và chỉ cần {GO_USAGE.generics}.
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={860}
          gap={52}
          left={<CodeCard snippet={GO_SNIPPETS.decoders} delay={16} accent={COLOR.go} highlight={[1, 2]} />}
          right={
            <Bullets
              gap={22}
              items={[
                { at: 120, tag: "dấu hiệu", text: "Ngoặc vuông sau tên hàm: into[…]. Cái trong ngoặc là một KIỂU, điền lúc gọi.", accent: COLOR.go },
                { at: 190, tag: "cả hai làm gì", text: "on[T] và into[T] cùng một việc: từ bytes hoặc any về một kiểu T cụ thể, mà không viết lại hàm cho từng kiểu.", accent: COLOR.money },
                { at: 260, tag: "nói thẳng", text: "Hai chỗ trong hơn một trăm file. Generics là công cụ hiếm dùng — phỏng vấn hỏi, nên học; code thường, nên ít gặp.", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={320} accent={COLOR.money}>
          Tập 11 đã học hàm nhận hàm, tập 12 đã học x.(T). on[T] là cả hai cộng lại, với một kiểu để trống.
        </Callout>
      </div>
    </div>
  </Stage>
);
