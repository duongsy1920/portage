import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — WaitGroup and a closed channel as a starting gun, in the test
 * that caught a real bug.
 *
 * Round 18: two processes migrating a cold database at once broke on a
 * Postgres index. The regression test has to make four goroutines collide on
 * purpose, and the two tools for that are exactly the two an interview asks
 * about. Teaching them on this test means teaching them on something that
 * mattered.
 */
export const T3Race: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 14 · CẢNH 3" source={`internal/adapter/postgres/postgres_test.go:103 — WaitGroup duy nhất của repo (${GO_USAGE.waitGroupsInTests}), trong test`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 170, sound: "reveal" }, { at: 380, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Bốn goroutine cùng rời vạch: WaitGroup và một channel đóng
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={860}
          gap={44}
          left={<CodeCard snippet={GO_SNIPPETS.raceStart} delay={20} accent={COLOR.money} highlight={[2, 3, 7, 11, 15, 16]} lineStagger={2} />}
          right={
            <Bullets
              gap={18}
              items={[
                { at: 170, tag: "vì sao có test này", text: "Đợt 18: hai process cùng migrate database trống thì một cái vỡ. Test phải làm bốn cái đâm nhau thật.", accent: COLOR.danger },
                { at: 230, tag: "WaitGroup", text: "Bộ đếm: Add(1) trước khi go, Done() khi xong, Wait() chờ về 0. Không có nó, test kết thúc trước goroutine.", accent: COLOR.go },
                { at: 290, tag: "súng lệnh", text: "Mỗi goroutine kết nối xong rồi đứng ở <-start. close(start) một lần: cả bốn nhận cùng lúc.", accent: COLOR.money },
                { at: 350, tag: "chan struct{}", text: "Channel không mang dữ liệu, chỉ mang tín hiệu; struct{} tốn 0 byte. Và go func(i int): mỗi goroutine một i riêng (tập 11).", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
