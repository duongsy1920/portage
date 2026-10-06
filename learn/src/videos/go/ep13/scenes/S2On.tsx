import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, REPO } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — on[T], read line by line.
 *
 * The interview answer lives in this function: without generics, every one of
 * the 35 subscribe lines would need its own hand-written adapter from
 * worker.Handler to a typed method — thirty-five near-identical functions.
 * With one type parameter, one function does it for all of them.
 */
export const S2On: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 13 · CẢNH 2" source="internal/platform/wire/subscribe.go:105">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 160, sound: "reveal" }, { at: 380, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        on[T]: closure có kiểu đi vào, handler không kiểu đi ra
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={1000}
          gap={40}
          left={<CodeCard snippet={GO_SNIPPETS.onGeneric} delay={20} accent={COLOR.go} highlight={[0, 6, 10]} />}
          right={
            <Bullets
              gap={18}
              items={[
                { at: 160, tag: "[T any]", text: "T là một kiểu để trống. Mỗi handler khác đưa vào on(…) là một T khác.", accent: COLOR.go },
                { at: 220, tag: "vào · ra", text: "Vào: hàm nhận T thật (OnLaneDefined nhận LaneDefinedV1). Ra: worker.Handler nhận Entry thô — Bus không cần biết T.", accent: COLOR.money },
                { at: 300, tag: "suy ra T", text: "on(lanes.OnLaneDefined) không ghi T: Go nhìn tham số, thấy nó nhận LaneDefinedV1, điền T.", accent: COLOR.ok },
                { at: 380, tag: "không có generics", text: `${REPO.subscribeLines} dòng đăng ký là ${REPO.subscribeLines} adapter viết tay gần giống hệt nhau. Đó là câu trả lời phỏng vấn.`, accent: COLOR.danger },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
