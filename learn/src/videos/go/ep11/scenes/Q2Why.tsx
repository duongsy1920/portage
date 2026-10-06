import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — why the port takes a function instead of returning a transaction.
 *
 * The design reason is the lesson: when the transaction is an object you hand
 * back, every caller must remember to commit it. When the transaction is a
 * function you hand in, nobody can forget — the end of the function IS the
 * commit, and a panic inside it IS the rollback (episode 9).
 */
export const Q2Why: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 11 · CẢNH 2" source="internal/app/ports.go:46">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 130, sound: "reveal" }, { at: 330, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Vì sao InTx nhận hàm, mà không trả về transaction
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={760}
          gap={52}
          left={<CodeCard snippet={GO_SNIPPETS.inTxPort} delay={20} accent={COLOR.go} highlight={[2]} />}
          right={
            <Bullets
              gap={22}
              items={[
                {
                  at: 130,
                  tag: "nếu trả transaction",
                  text: "Người gọi phải nhớ Commit. Quên một chỗ là dữ liệu treo trong transaction mở, database giữ khoá.",
                  accent: COLOR.danger,
                },
                {
                  at: 200,
                  tag: "nếu nhận hàm",
                  text: "Hàm chạy xong là commit, trả lỗi là rollback. Không có đường nào để quên, vì không có bước nào để làm.",
                  accent: COLOR.go,
                },
                {
                  at: 270,
                  tag: "panic thì sao",
                  text: "Tập 9: defer trong InTx bắt panic, rollback, rồi ném lại. Người gọi không phải lo.",
                  accent: COLOR.money,
                },
                {
                  at: 330,
                  tag: "bên symfony",
                  text: "$em->wrapInTransaction(fn () => …) — cùng một ý. Khác ở chỗ đây là interface, và bản in-memory không có transaction thật.",
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
