import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { PHP_SNIPPETS, DDD_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 2, episode 1, scene 1 — the model everyone has written.
 *
 * The left card is not a strawman. It is the shape Doctrine, form handling and
 * fifteen years of tutorials all push you toward, and it works. The question
 * is not whether it is bad code; it is where the business rule went.
 */
export const A1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 1" source="bên trái: PHP minh hoạ · bên phải: ordering/order.go:225">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 100, sound: "reveal" }, { at: 220, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Luật nghiệp vụ nằm ở đâu?
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={800}
          gap={52}
          align="center"
          left={<CodeCard snippet={PHP_SNIPPETS.anemic} delay={20} accent={COLOR.php} highlight={[8]} scale={0.92} />}
          right={<CodeCard snippet={DDD_SNIPPETS.oneDoor} delay={100} accent={COLOR.go} highlight={[0, 2, 6]} scale={0.92} />}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={220} accent={COLOR.php}>
          Bên trái, ai gọi setStatus cũng được, lúc nào cũng được. Luật “chỉ cọc được khi
          đơn còn ở trạng thái vừa đặt” phải nằm đâu đó khác — và thường là nằm ở vài chỗ khác.
        </Callout>
      </div>
    </div>
  </Stage>
);
