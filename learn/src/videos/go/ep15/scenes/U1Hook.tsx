import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 15, scene 1 — the same fact, written twice.
 *
 * Left, the domain event: typed fields, not one tag. Right, the contract:
 * strings and primitives, a tag on every line, a V1 in the name. Putting them
 * side by side first means the rest of the episode is explaining a picture
 * the viewer has already seen.
 */
export const U1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 15" source="internal/domain/ordering/events.go:12 · internal/contracts/ordering_v1.go:11 — code thật">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 110, sound: "reveal" }, { at: 280, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một sự kiện, hai struct: một không có tag, một toàn tag
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={760}
          gap={56}
          align="center"
          left={<CodeCard snippet={GO_SNIPPETS.domainEvent} delay={16} accent={COLOR.go} highlight={[1, 7]} />}
          right={<CodeCard snippet={GO_SNIPPETS.contractV1} delay={110} accent={COLOR.money} highlight={[0, 1, 7]} />}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={280} accent={COLOR.money}>
          {GO_USAGE.jsonTags} tag json trong code chạy thật. Trong internal/domain: 0. Tập này nói về cái ranh giới đó.
        </Callout>
      </div>
    </div>
  </Stage>
);
