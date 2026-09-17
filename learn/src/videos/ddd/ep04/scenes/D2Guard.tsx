import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the rule, and the test that enforces it.
 *
 * A convention nobody checks is a convention that lasts until the first busy
 * week. This repository has seven architecture tests written with go/ast, and
 * the one on screen fails the build if anything under internal/domain imports
 * something it should not.
 */
export const D2Guard: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 4 · CẢNH 2 · TỪ VỰNG" source="internal/domain/decisions_test.go — 7 test canh kiến trúc">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 320, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Luật vàng, và bài test canh nó
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={820}
          gap={52}
          left={<CodeCard snippet={SNIPPETS.guard} delay={20} accent={COLOR.ok} scale={0.9} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={150}
                accent={COLOR.ok}
                term="port"
                plain="Interface do phía cần khai báo, để phía làm cắm vào. Repository là một cái port."
                symfony="Interface của bạn cũng là port — nhưng container tự cắm, nên ít khi phải nghĩ tới chiều."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 320,
                    tag: "luật vàng",
                    text: "internal/domain chỉ import stdlib và uuid. Vi phạm là build đỏ, không phải lời nhắc.",
                    accent: COLOR.ok,
                  },
                  {
                    at: 380,
                    tag: "đổi lại được gì",
                    text: "Domain test được mà không cần database — 278 test chạy dưới 2 giây.",
                    accent: COLOR.go,
                  },
                ]}
              />
            </div>
          }
        />
      </div>
    </div>
  </Stage>
);
