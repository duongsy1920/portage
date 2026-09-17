import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { ContextMap } from "../../../../components/ContextMap";
import { CodeCard } from "../../../../components/CodeCard";
import { SNIPPETS, REPO } from "../../../../data/portage";
import { COLOR, SAFE } from "../../../../design/tokens";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 10 — the map, revealed whole for the first time.
 *
 * This picture closes every episode of the series with a different part lit,
 * so it is shown complete once, here, to give the viewer something to come
 * back to. The absence of arrows between the cards is the lesson.
 */
export const S10Map: React.FC = () => {
  return (
    <Stage
      eyebrow="TẬP 1 · CẢNH 10"
      source={`${REPO.subscribeLines} dòng định tuyến · ${REPO.events} event · internal/platform/wire/subscribe.go`}
    >
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 180, sound: "reveal" },
          { at: 250, sound: "land", volume: 0.26 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Vậy chúng nói chuyện với nhau thế nào?
        </Headline>

        <div style={{ marginTop: 26 }}>
          <ContextMap delay={20} width={SAFE.contentWidth} />
        </div>

        <div style={{ display: "flex", gap: 60, marginTop: 30, alignItems: "flex-start" }}>
          <div style={{ width: 700, flexShrink: 0 }}>
            <CodeCard snippet={SNIPPETS.subscribe} delay={180} accent={COLOR.go} />
          </div>
          <div style={{ flex: 1 }}>
            <Bullets
              gap={20}
              items={[
                {
                  at: 200,
                  text: "Ordering ghi lại việc 'đã nhận cọc'. Nó không biết ai sẽ đọc, và không cần biết.",
                  accent: COLOR.go,
                },
                {
                  at: 250,
                  text: "Procurement đăng ký nghe đúng việc đó. Thêm người nghe thứ hai không phải sửa bên gửi.",
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
