import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Bullets } from "../../../../components/Bullets";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 2, scene 1 — the promise.
 *
 * The hook is the code itself, unexplained, because the honest starting point
 * is the feeling of looking at four lines and not knowing where to put your
 * eyes. Naming that feeling first makes the next six scenes feel like relief
 * rather than like a syntax lecture.
 */
export const H1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 2" source="internal/domain/shared/money.go — code thật">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 120, sound: "reveal" }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column", justifyContent: "center", gap: 40 }}>
      <Headline delay={4} size={64}>
        Bốn dòng này, cuối tập bạn đọc được không cần đoán
      </Headline>

      <div style={{ display: "flex", gap: 60, alignItems: "flex-start" }}>
        <CodeCard snippet={GO_SNIPPETS.moneyStruct} delay={30} accent={COLOR.go} />
        <div style={{ flex: 1, paddingTop: 10 }}>
          <Bullets
            gap={22}
            items={[
              {
                at: 120,
                tag: "lạ chỗ nào",
                text: "int64 đứng sau tên biến. Bên PHP kiểu đứng trước: private int $minor.",
                accent: COLOR.php,
              },
              {
                at: 180,
                tag: "lạ chỗ nào",
                text: "Không thấy private, không thấy public, mà vẫn có cái riêng cái chung.",
                accent: COLOR.php,
              },
              {
                at: 240,
                tag: "lạ chỗ nào",
                text: "Không có method nào bên trong dấu ngoặc. Vậy Money làm gì được?",
                accent: COLOR.php,
              },
            ]}
          />
        </div>
      </div>
    </div>
  </Stage>
);
