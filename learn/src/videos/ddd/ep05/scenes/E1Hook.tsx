import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { DDD_SNIPPETS } from "../../../../data/portage";
import { COLOR, CONTEXT } from "../../../../design/tokens";

/**
 * Season 2, episode 5, scene 1 — the same word, three times, for real.
 *
 * This is the single best argument for bounded contexts in the repository and
 * it needs no explaining: three files, three packages, one word, three
 * different shapes. Nobody had to agree on a definition, and that is the point.
 */
export const E1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 5" source="ba file thật trong repo, cùng tên kiểu là Variant">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 90, sound: "appear" }, { at: 160, sound: "appear" }, { at: 250, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Một chữ “Variant”, ba nghĩa khác nhau
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center", gap: 24 }}>
        <div style={{ flex: 1 }}>
          <CodeCard snippet={DDD_SNIPPETS.variantCatalog} delay={20} accent={CONTEXT.catalog.color} scale={0.72} />
        </div>
        <div style={{ flex: 1 }}>
          <CodeCard snippet={DDD_SNIPPETS.variantOrdering} delay={90} accent={CONTEXT.ordering.color} scale={0.72} />
        </div>
        <div style={{ flex: 1 }}>
          <CodeCard snippet={DDD_SNIPPETS.variantProcurement} delay={160} accent={CONTEXT.procurement.color} scale={0.72} />
        </div>
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={250} accent={COLOR.money}>
          Không ai phải họp để thống nhất một định nghĩa chung. Mỗi vùng giữ đúng phần
          nó cần — và đó chính là điều bounded context cho phép.
        </Callout>
      </div>
    </div>
  </Stage>
);
