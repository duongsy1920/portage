import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — the struct holds fields, and nothing else.
 *
 * The receiver `(m Money)` is episode 4's subject, so it is named here and
 * deliberately left unexplained: "this is what attaches the function to the
 * type, we open it later". Half-teaching it now would cost a minute and leave
 * the viewer with a half-idea.
 */
export const H4NoMethods: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 2 · CẢNH 4" source="money.go:57 và money.go:145 — cách nhau 88 dòng">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 120, sound: "reveal" }, { at: 260, sound: "snap", volume: 0.18 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Luật ba: struct chỉ chứa dữ liệu, hàm viết ở ngoài
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={660}
          gap={56}
          left={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <CodeCard snippet={GO_SNIPPETS.moneyStruct} delay={20} accent={COLOR.go} />
              <CodeCard snippet={GO_SNIPPETS.minorMethod} delay={120} accent={COLOR.ok} highlight={[0]} />
            </div>
          }
          right={
            <Bullets
              gap={22}
              items={[
                {
                  at: 60,
                  tag: "trên",
                  text: "Trong dấu ngoặc của struct chỉ có field. Không có hàm nào cả.",
                  accent: COLOR.go,
                },
                {
                  at: 140,
                  tag: "dưới",
                  text: "Hàm viết tách hẳn ra, cách đó 88 dòng, vẫn thuộc về Money.",
                  accent: COLOR.ok,
                },
                {
                  at: 220,
                  tag: "gắn bằng gì",
                  text: "Cụm (m Money) đứng trước tên hàm là thứ gắn nó vào kiểu Money. Tập 4 mở cụm đó ra.",
                  accent: COLOR.ok,
                },
                {
                  at: 300,
                  tag: "bên symfony",
                  text: "PHP gom field và method trong một cặp ngoặc của class. Go tách đôi, nên một kiểu có thể có method nằm rải ở nhiều file.",
                  accent: COLOR.php,
                },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
