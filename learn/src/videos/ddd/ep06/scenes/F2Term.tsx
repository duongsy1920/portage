import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, REPO } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the vocabulary, and the tense that carries it.
 *
 * Past tense is not a style preference: a command can be refused, an event
 * cannot. Naming one `DepositPaid` rather than `PayDeposit` is what stops a
 * listener from thinking it has a say in the matter.
 */
export const F2Term: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 6 · CẢNH 2 · TỪ VỰNG" source="internal/domain/shared/event.go:78">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 310, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Tên event luôn ở thì quá khứ
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={760}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.pointerReceiver} delay={20} accent={COLOR.money} highlight={[0, 1]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <TermCard
                delay={150}
                accent={COLOR.money}
                term="domain event"
                plain="Một việc đã xảy ra rồi, ghi lại bằng tên ở thì quá khứ. Không ai từ chối được nữa."
                symfony="Symfony Messenger — nhưng ở đó bạn dispatch ngay; ở đây aggregate chỉ ghi, chưa gửi."
              />
              <Bullets
                gap={18}
                items={[
                  {
                    at: 310,
                    tag: "vì sao quá khứ",
                    text: "Lệnh thì từ chối được, việc đã xảy ra thì không. Tên gọi nói rõ người nghe có quyền gì.",
                    accent: COLOR.money,
                  },
                  {
                    at: 370,
                    tag: "trong repo",
                    text: `${REPO.events} loại event, và ${REPO.subscribeLines} dòng đăng ký nghe — không listener nào gọi ngược lại vùng đã phát.`,
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
