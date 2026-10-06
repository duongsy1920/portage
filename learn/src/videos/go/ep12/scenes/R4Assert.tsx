import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — the type assertion, in the form the repository always uses.
 *
 * `Decode` hands back `any` because it cannot know which V1 struct the caller
 * wants. The caller asks with `msg.(T)` — and asks WITH ok, because without
 * it a wrong type is a panic in a worker, which episode 9 showed kills the
 * whole process.
 */
export const R4Assert: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 12 · CẢNH 4" source={`internal/platform/wire/subscribe.go:111 — repo có ${GO_USAGE.typeAssertOk} assertion, cả hai đều có ok`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 120, sound: "reveal" }, { at: 330, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        x.(T): hỏi lại kiểu — và luôn hỏi kèm ok
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={760}
          gap={48}
          left={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <CodeCard snippet={GO_SNIPPETS.decodeAny} delay={20} accent={COLOR.money} highlight={[0]} />
              <CodeCard snippet={GO_SNIPPETS.assertOk} delay={120} accent={COLOR.go} highlight={[0, 1]} />
            </div>
          }
          right={
            <Bullets
              gap={20}
              items={[
                { at: 220, tag: "không ok", text: "m := msg.(T) — sai kiểu là panic. Trong worker, panic là chết cả tiến trình (tập 9).", accent: COLOR.danger },
                { at: 280, tag: "có ok", text: "m, ok := msg.(T) — sai kiểu thì ok là false, m là zero value. Trả lỗi, worker giữ hàng, thử lại sau.", accent: COLOR.ok },
                { at: 340, tag: "cái giá", text: "Compiler không biết tên event khớp kiểu nào. Lỗi chỉ lộ lúc chạy — nên có test cho từng decoder.", accent: COLOR.money },
                { at: 400, tag: "comma-ok", text: "Cùng hình với đọc map: một giá trị, rồi một bool nói có hay không.", accent: COLOR.go },
              ]}
            />
          }
        />
      </div>

    </div>
  </Stage>
);
