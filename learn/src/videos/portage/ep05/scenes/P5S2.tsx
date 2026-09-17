import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { DDD_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the rule written down, and what it refuses.
 *
 * The status check at the top of PayDeposit is three lines and it is the whole
 * safety mechanism: no setter, no second path, and a named error so the HTTP
 * layer can tell the customer which rule stopped them.
 */
export const P5S2: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 5 · CẢNH 2" source="internal/domain/ordering/order.go:225">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Luật viết ra được, và từ chối được
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={860}
          gap={52}
          left={<CodeCard snippet={DDD_SNIPPETS.oneDoor} delay={20} accent={COLOR.money} highlight={[2, 3, 4]} scale={0.9} />}
          right={
            <Bullets
              gap={22}
              items={[
                {
                  at: 140,
                  tag: "ba dòng đầu",
                  text: "Sai trạng thái thì trả error và không đổi gì. Đó là toàn bộ cơ chế an toàn.",
                  accent: COLOR.money,
                },
                {
                  at: 210,
                  tag: "tên lỗi",
                  text: "ErrNotAwaitingDeposit đi thẳng vào bảng mã lỗi, nên khách nhận 409 kèm lý do đọc được.",
                  accent: COLOR.go,
                },
                {
                  at: 280,
                  tag: "không có đường vòng",
                  text: "Không có setStatus, không có UPDATE thẳng từ tầng app. Một cửa, và cửa đó kiểm luật.",
                  accent: COLOR.ok,
                },
                {
                  at: 350,
                  tag: "vì sao quan trọng",
                  text: "Đây là chỗ một dòng code sai đồng nghĩa với mất tiền thật của khách.",
                  accent: COLOR.danger,
                },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
