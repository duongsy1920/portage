import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — why a lock, and not a channel.
 *
 * The honest answer is boring and therefore worth saying out loud: the shared
 * thing here is a map that several goroutines read and write, and that is what
 * a mutex is for. Channels are for handing work over, which this code does not
 * do. Choosing by the shape of the problem beats choosing by fashion.
 */
export const P2Why: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 10 · CẢNH 2" source="internal/adapter/memory/merchant_repo.go:24">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 130, sound: "reveal" }, { at: 300, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Vì sao khoá, mà không phải channel
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={800}
          gap={52}
          left={<CodeCard snippet={GO_SNIPPETS.mutex} delay={20} accent={COLOR.danger} highlight={[1, 2, 10, 11]} />}
          right={
            <Bullets
              gap={22}
              items={[
                {
                  at: 130,
                  tag: "hình dạng bài toán",
                  text: "Ở đây có một cái map dùng chung, nhiều goroutine đọc và ghi vào nó. Đó đúng là việc của khoá.",
                  accent: COLOR.danger,
                },
                {
                  at: 200,
                  tag: "channel để làm gì",
                  text: "Channel là để chuyển giao việc từ goroutine này sang goroutine khác. Code này không chuyển giao gì cả.",
                  accent: COLOR.go,
                },
                {
                  at: 270,
                  tag: "câu hay bị nhớ sai",
                  text: "“Đừng chia sẻ bộ nhớ, hãy giao tiếp” là lời khuyên, không phải luật. Bộ nhớ dùng chung có khoá vẫn là Go đúng chuẩn.",
                  accent: COLOR.money,
                },
                {
                  at: 340,
                  tag: "kiểm bằng gì",
                  text: "go test -race chạy thật và báo nếu có hai goroutine chạm cùng một ô nhớ mà không khoá.",
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
