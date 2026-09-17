import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 4, scene 1 — one asterisk, two behaviours.
 *
 * The two snippets are shown together and unexplained. The whole episode is
 * about a single character of difference, so the character had better be on
 * screen before anything is said about it.
 */
export const J1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 4" source="money.go:201 và event.go:78 — code thật">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 210, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Một dấu sao, và hai hành vi khác hẳn nhau
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={780}
          gap={56}
          align="center"
          left={<CodeCard snippet={GO_SNIPPETS.valueReceiver} delay={20} accent={COLOR.go} highlight={[0]} />}
          right={<CodeCard snippet={GO_SNIPPETS.pointerReceiver} delay={90} accent={COLOR.money} highlight={[0, 4]} />}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={210} accent={COLOR.money}>
          Bên trái <span style={{ color: COLOR.go }}>(m Money)</span>, bên phải{" "}
          <span style={{ color: COLOR.money }}>(e *Events)</span>. Khác đúng một ký tự, và
          nó quyết định method có sửa được bản gốc hay không.
        </Callout>
      </div>
    </div>
  </Stage>
);
