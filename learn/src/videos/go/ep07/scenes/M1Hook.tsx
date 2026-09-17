import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 7, scene 1 — the thing that looks like inheritance.
 *
 * A field with no name is the single strangest line a Symfony developer meets
 * in this repository, and it appears in every aggregate root. Reading it as
 * `extends` works for about ten minutes and then breaks, so it gets an episode.
 */
export const M1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 7" source="internal/domain/ordering/order.go:25 và :109">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 210, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Một field không có tên
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={760}
          gap={56}
          align="center"
          left={<CodeCard snippet={GO_SNIPPETS.typedID} delay={20} accent={COLOR.go} highlight={[1]} />}
          right={<CodeCard snippet={GO_SNIPPETS.embedEvents} delay={90} accent={COLOR.money} highlight={[1]} />}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={210} accent={COLOR.money}>
          Hai dòng in đậm chỉ có kiểu, không có tên field. Đó gọi là{" "}
          <span style={{ color: COLOR.money }}>embedding</span>, và nó trông như kế thừa
          cho tới lúc bạn cần nó cư xử như kế thừa.
        </Callout>
      </div>
    </div>
  </Stage>
);
