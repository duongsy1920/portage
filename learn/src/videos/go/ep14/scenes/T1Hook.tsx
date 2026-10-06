import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, episode 14, scene 1 — the loop episode 10 counted.
 *
 * Episode 10 said the repository reads from four channels and creates none.
 * This is the function those reads live in, shown whole, because a learner
 * who opens worker.go will see exactly this and should recognise every line.
 */
export const T1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 14" source="internal/worker/worker.go:127 — code thật">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 300, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Tập 10 đếm {GO_USAGE.channelsRead} lần đọc channel. Đây là hai trong số đó.
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={900}
          gap={48}
          left={<CodeCard snippet={GO_SNIPPETS.selectLoop} delay={16} accent={COLOR.go} highlight={[7, 8, 10]} />}
          right={
            <Bullets
              gap={20}
              items={[
                { at: 140, tag: "vòng lặp", text: "Chạy một lượt relay, rồi đứng chờ. Chờ gì? Một trong hai channel có tin.", accent: COLOR.go },
                { at: 200, tag: "ticker.C", text: "Mỗi Interval, ticker bỏ một nhịp vào channel C. Nhận được là tới lượt chạy tiếp.", accent: COLOR.money },
                { at: 260, tag: "ctx.Done()", text: "Channel thứ hai, của context. Nó đóng khi ai đó huỷ ctx — và đọc từ channel đã đóng là nhận ngay.", accent: COLOR.danger },
                { at: 320, tag: "sweep.go", text: "Hai lần đọc còn lại nằm trong Sweeper.Run, cùng hình y hệt. Hai khối select, cả repo.", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={360} accent={COLOR.money}>
          Không tự tạo channel nào, nhưng hai thứ Go đưa sẵn — context và ticker — đều nói chuyện bằng channel.
        </Callout>
      </div>
    </div>
  </Stage>
);
