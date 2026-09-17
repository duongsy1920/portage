import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { SNIPPETS } from "../../../../data/portage";
import { COLOR, CONTEXT } from "../../../../design/tokens";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 7 — the answer, in the project's own code.
 *
 * Landscape earns its keep here: the two structs sit side by side, so the
 * split is something the eye sees rather than something the narration claims.
 * Colours are the ones scene 1 already trained.
 */
export const S7TwoStructs: React.FC = () => {
  return (
    <Stage eyebrow="TẬP 1 · CẢNH 7" source="Code thật, chép nguyên từ repo">
      <Cues
        items={[
          { at: 30, sound: "appear" },
          { at: 120, sound: "reveal" },
          { at: 310, sound: "appear", volume: 0.14 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Nên trong Portage, nó là hai thứ
        </Headline>

        <div style={{ display: "flex", gap: 60, marginTop: 34 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <CodeCard
              snippet={SNIPPETS.customerOrder}
              delay={30}
              accent={CONTEXT.ordering.color}
              highlight={[0]}
            />
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <CodeCard
              snippet={SNIPPETS.purchaseTask}
              delay={120}
              accent={CONTEXT.procurement.color}
              highlight={[0]}
            />
          </div>
        </div>

        <div style={{ display: "flex", gap: 70, marginTop: 34 }}>
          <div style={{ width: 810 }}>
            <Bullets
              gap={20}
              items={[
                {
                  at: 220,
                  text: "Đơn của khách: tiền, cam kết, huỷ thì hoàn bao nhiêu. Khách nhìn thấy nó.",
                  accent: CONTEXT.ordering.color,
                },
                {
                  at: 260,
                  text: "Việc đi mua: mã đơn ở shop, đã trả thật bao nhiêu. Nhân viên nhìn thấy nó.",
                  accent: CONTEXT.procurement.color,
                },
              ]}
            />
          </div>
          <div style={{ width: 810 }}>
            <Bullets
              gap={20}
              items={[
                {
                  at: 310,
                  tag: "chú ý",
                  text: "Hai struct ở hai package khác nhau, và không package nào import package kia.",
                  accent: COLOR.go,
                },
                {
                  at: 360,
                  text: "Chúng chỉ biết nhau qua một id trần, không qua kiểu dữ liệu của nhau.",
                  accent: COLOR.go,
                },
              ]}
            />
          </div>
        </div>
      </div>
    </Stage>
  );
};
