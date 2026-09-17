import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — the one honest use of recover in this repository.
 *
 * It does not swallow the panic. It rolls the transaction back and then panics
 * again with the same value, so the database is not left holding an open
 * transaction while the process dies. That re-panic is the line worth pausing
 * on: recover used to clean up, never to continue.
 */
export const O4Recover: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 9 · CẢNH 4" source="internal/adapter/postgres/postgres.go:91">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        recover dùng để dọn dẹp, không dùng để đi tiếp
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={840}
          gap={52}
          left={<CodeCard snippet={GO_SNIPPETS.recoverTx} delay={20} accent={COLOR.money} highlight={[1, 2, 3]} scale={0.9} />}
          right={
            <Bullets
              gap={20}
              items={[
                {
                  at: 140,
                  tag: "bắt được",
                  text: "recover() chỉ chạy được bên trong defer. Ngoài defer thì nó luôn trả về nil.",
                  accent: COLOR.go,
                },
                {
                  at: 210,
                  tag: "rồi panic lại",
                  text: "Dòng panic(p) ném lại đúng cái vừa bắt. Nó không nuốt lỗi — chỉ kịp rollback trước khi chết.",
                  accent: COLOR.danger,
                },
                {
                  at: 280,
                  tag: "vì sao cần",
                  text: "Không có nó, tiến trình chết mà transaction vẫn mở, database giữ khoá cho tới lúc hết giờ.",
                  accent: COLOR.money,
                },
                {
                  at: 350,
                  tag: "err có tên",
                  text: "Chữ (err error) ở chữ ký hàm cho phép defer ghi đè giá trị trả về — nhờ vậy lỗi lúc commit vẫn báo ra ngoài được.",
                  accent: COLOR.ok,
                },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
