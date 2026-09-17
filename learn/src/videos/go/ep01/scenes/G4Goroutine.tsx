import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, scene 4 — the word for the thing just seen.
 *
 * The term arrives AFTER the picture on purpose. A definition given first is
 * a label with nothing under it; given second it is a name for something the
 * viewer already watched happen.
 */
export const G4Goroutine: React.FC = () => {
  return (
    <Stage eyebrow="MÙA 1 · TẬP 1 · CẢNH 4 · TỪ VỰNG" source="cmd/api/main.go — code thật">
      <Cues
        items={[
          { at: 30, sound: "appear" },
          { at: 150, sound: "reveal" },
          { at: 300, sound: "appear", volume: 0.14 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Từ mới: goroutine
        </Headline>

        <div style={{ marginTop: 30, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={700}
            gap={60}
            left={<CodeCard snippet={GO_SNIPPETS.mainOnce} delay={30} accent={COLOR.go} highlight={[7, 8, 9]} />}
            right={
              <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
                <TermCard
                  delay={150}
                  accent={COLOR.go}
                  term="goroutine"
                  plain="Một việc chạy song song, nhẹ tới mức chạy vài trăm nghìn cái cùng lúc vẫn bình thường."
                  symfony="Không có. Process PHP-FPM tốn vài megabyte, goroutine chỉ tốn hai kilobyte."
                  inRepo={`go func() { … }() — repo có ${GO_USAGE.goroutines} chỗ`}
                />
                <Bullets
                  gap={16}
                  items={[
                    {
                      at: 310,
                      tag: "đừng nhầm",
                      text: "Goroutine không phải thread của hệ điều hành. Go xếp nghìn goroutine lên vài thread.",
                      accent: COLOR.go,
                    },
                    {
                      at: 360,
                      tag: "vì sao hợp",
                      text: "Gọi API mất 30 giây thì goroutine chỉ nằm chờ, gần như không tốn gì.",
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
};
