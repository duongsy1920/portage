import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 9, scene 1 — the two lines that always travel together.
 *
 * Lock then defer Unlock is the most repeated shape in the repository. Showing
 * it before defining `defer` means the word arrives as a name for something
 * already seen working, which is the pattern the whole season uses.
 */
export const O1Hook: React.FC = () => (
  <Stage
    eyebrow="MÙA 1 · TẬP 9"
    source={`repo có ${GO_USAGE.defers} lệnh defer và ${GO_USAGE.mutex} chỗ khoá`}
  >
    <Cues items={[{ at: 16, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 210, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Hai dòng luôn đi cùng nhau
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={760}
          gap={56}
          align="center"
          left={<CodeCard snippet={GO_SNIPPETS.mutex} delay={20} accent={COLOR.danger} highlight={[10, 11]} />}
          right={<CodeCard snippet={GO_SNIPPETS.recoverTx} delay={90} accent={COLOR.money} highlight={[0, 1]} scale={0.86} />}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={210} accent={COLOR.go}>
          <span style={{ color: COLOR.go }}>defer</span> hoãn một lệnh lại tới lúc hàm
          thoát — thoát kiểu gì cũng chạy, kể cả return giữa chừng, kể cả khi đang panic.
        </Callout>
      </div>
    </div>
  </Stage>
);
