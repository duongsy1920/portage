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
 * Scene 2 — the fix is a table, not a queue.
 *
 * The outbox turns "two systems must both succeed" into "one database must
 * commit", which is a problem databases already solved. The INSERT on screen
 * runs on the same connection as the order, inside the same transaction.
 */
export const P2S2: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 2 · CẢNH 2 · TỪ VỰNG" source="internal/adapter/postgres/outbox.go:35">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 310, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Ghi event vào một cái bảng, cùng transaction
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={800}
          gap={52}
          left={<CodeCard snippet={SNIPPETS.outboxAppend} delay={20} accent={COLOR.money} highlight={[2, 5]} scale={0.84} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
              <TermCard
                delay={150}
                accent={COLOR.money}
                term="outbox"
                plain="Một bảng trong chính database đó. Event ghi cùng lúc với dữ liệu — hoặc có cả hai, hoặc không có gì."
                symfony="Messenger với doctrine transport — cùng ý tưởng."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 310,
                    tag: "mẹo nằm ở đâu",
                    text: "db(ctx, o.pool) lấy transaction đang mở từ ctx — nên INSERT này đi chung chuyến với đơn.",
                    accent: COLOR.go,
                  },
                  {
                    at: 370,
                    tag: "hết khe hở",
                    text: "Hết “hai hệ thống phải cùng thành công”. Chỉ còn một database phải commit.",
                    accent: COLOR.ok,
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
