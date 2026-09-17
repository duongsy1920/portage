import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 5, scene 1 — the missing keyword.
 *
 * A Symfony developer's first question about Go interfaces is "where does it
 * say it implements that?". The answer is that nothing says it, anywhere, and
 * the file on the right is the whole proof: a struct, a pool, and no mention
 * of the interface it satisfies.
 */
export const K1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 5" source="app/ports.go:46 và adapter/postgres/product_repo.go:23">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 210, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Go không có từ khoá implements
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={790}
          gap={56}
          align="center"
          left={<CodeCard snippet={GO_SNIPPETS.portInterface} delay={20} accent={COLOR.go} />}
          right={<CodeCard snippet={GO_SNIPPETS.assertion} delay={90} accent={COLOR.ok} highlight={[4]} />}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={210} accent={COLOR.go}>
          Bên phải không nhắc tên interface nào ở chỗ khai báo struct. Nó vẫn dùng được
          ở mọi nơi cần <span style={{ color: COLOR.go }}>ProductRepository</span> — chỉ vì
          nó có đủ method.
        </Callout>
      </div>
    </div>
  </Stage>
);
