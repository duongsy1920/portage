import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE, REPO } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 12, scene 1 — a two-method interface, and the door all
 * events go through.
 *
 * Episode 5 taught implicit interfaces. This episode starts from the other
 * end: once 37 different structs are all "an Event", something has to turn
 * each one back into its own type to write it down. That something is a type
 * switch, and the honest framing is that it is a cost, paid once.
 */
export const R1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 12" source="internal/domain/shared/event.go:26 — code thật">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 120, sound: "reveal" }, { at: 250, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Hai method là đủ để là một Event. Rồi {REPO.events} loại chui qua một cái cửa.
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={700}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.eventInterface} delay={16} accent={COLOR.go} highlight={[1, 2]} />}
          right={
            <Bullets
              gap={22}
              items={[
                { at: 120, tag: "tập 5 đã nói", text: "Struct nào có đủ hai method này là Event, không cần khai báo gì. Năm context, mỗi context vài loại.", accent: COLOR.go },
                { at: 190, tag: "câu hỏi mới", text: "Event đi vào outbox phải thành JSON. Mà JSON của OrderPlaced khác JSON của ParcelReceived. Ai biết đang cầm cái nào?", accent: COLOR.money },
                { at: 250, tag: "câu trả lời", text: `Một chỗ duy nhất hỏi lại kiểu thật: type switch với ${GO_USAGE.typeSwitchCases} case trong eventcodec.`, accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.money}>
          Interface cho phép quên kiểu thật. Tập này nói về lúc phải nhớ lại, và cái giá của việc đó.
        </Callout>
      </div>
    </div>
  </Stage>
);
