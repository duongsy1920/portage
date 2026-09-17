import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { PHP_SNIPPETS, GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 3, scene 1 — the two habits, side by side.
 *
 * The PHP block on the left is written for the comparison and labelled PHP, so
 * nobody mistakes it for repository code. It is the honest picture of the
 * habit being replaced: a catch block you can forget to write, and the program
 * runs anyway.
 */
export const I1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 3" source="internal/domain/shared/money.go:102 — code thật">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 200, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Go không có try/catch. Không phải thiếu — là cố ý.
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={780}
          gap={56}
          align="center"
          left={<CodeCard snippet={PHP_SNIPPETS.tryCatch} delay={20} accent={COLOR.php} highlight={[3]} />}
          right={<CodeCard snippet={GO_SNIPPETS.parseMoney} delay={90} accent={COLOR.go} highlight={[0, 2]} />}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={200} accent={COLOR.go}>
          Bên trái, quên khối catch thì chương trình vẫn chạy. Bên phải, lỗi là một
          giá trị trả về — muốn bỏ qua thì phải tự tay viết ra là mình đang bỏ qua.
        </Callout>
      </div>
    </div>
  </Stage>
);
