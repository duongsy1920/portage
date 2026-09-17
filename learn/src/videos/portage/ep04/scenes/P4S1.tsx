import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { SNIPPETS, REPO } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 3, episode 4, scene 1 — the number that proves the architecture.
 *
 * Hexagonal architecture is usually argued for. Here it can be measured: the
 * test suite runs without Docker because nothing in the domain or the app layer
 * can reach a database. The guard test on screen is what keeps that true.
 */
export const P4S1: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 4" source="internal/domain/decisions_test.go — 7 test canh kiến trúc">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 140, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        {REPO.testsNoDocker} test, dưới hai giây, không cần Docker
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={880}
          gap={56}
          align="center"
          left={<CodeCard snippet={SNIPPETS.guard} delay={20} accent={COLOR.ok} scale={0.88} />}
          right={
            <div style={{ fontSize: 34, lineHeight: 1.45, color: COLOR.textDim }}>
              Không phải vì test viết khéo.
              <br />
              Mà vì <span style={{ color: COLOR.ok, fontWeight: 700 }}>không có gì</span> trong
              domain chạm được tới database — và có bài test canh điều đó.
            </div>
          }
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={140} accent={COLOR.go}>
          Kiến trúc lục giác thường phải đi thuyết phục. Ở đây nó đo được: một con số,
          và một bài test làm nó đỏ nếu ai đó phá luật.
        </Callout>
      </div>
    </div>
  </Stage>
);
