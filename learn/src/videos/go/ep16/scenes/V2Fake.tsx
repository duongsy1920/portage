import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the word "fake", on the test world the pricing use cases run in.
 *
 * Every field of `world` is an in-memory adapter behind the same port the
 * Postgres adapter implements. That is why the same behaviour is tested here
 * in microseconds and again against Postgres in CI: one port, two adapters,
 * one set of expectations.
 */
export const V2Fake: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 16 · CẢNH 2 · TỪ VỰNG" source="internal/app/pricing/pricing_test.go:19">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 340, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Fake: adapter thật chạy trong RAM, không phải kịch bản
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={660}
          gap={52}
          left={<CodeCard snippet={GO_SNIPPETS.worldFake} delay={20} accent={COLOR.go} highlight={[2, 3, 7]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
              <TermCard
                delay={150}
                accent={COLOR.go}
                term="fake"
                plain="Cài đặt thật của port, sống trong bộ nhớ. Khác mock: không ai dặn trước nó phải trả gì."
                symfony="InMemoryRepository tự viết, thay cho createMock()->willReturn(…)."
                inRepo="adapter/memory — cùng port với postgres"
              />
              <Bullets
                gap={16}
                items={[
                  { at: 340, tag: "vì sao không mock", text: "Mock kiểm kịch bản (“được gọi với đối số kia”). Fake kiểm hành vi: lưu rồi tìm lại phải ra.", accent: COLOR.danger },
                  { at: 400, tag: "một test, hai adapter", text: "Cùng khẳng định chạy trên RAM lúc dev, trên Postgres ở CI. Khác nhau là adapter sai.", accent: COLOR.ok },
                ]}
              />
            </div>
          }
        />
      </div>
    </div>
  </Stage>
);
