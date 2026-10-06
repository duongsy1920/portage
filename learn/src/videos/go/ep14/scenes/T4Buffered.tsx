import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — buffered or not, on the one place the repository chooses a
 * buffer of one.
 *
 * The test gives up after three seconds via time.After and walks away. The
 * goroutine may still finish later and send its error; with an unbuffered
 * channel that send would block forever — a leaked goroutine in a test.
 * A buffer of one lets the send complete with nobody listening. That is the
 * whole interview answer, and it is real.
 */
export const T4Buffered: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 14 · CẢNH 4" source={`internal/adapter/openai/openai_test.go:120 — ${GO_USAGE.chansInTests} channel tự tạo trong repo, đều trong test`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 350, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        make(chan error, 1): vì sao đúng một ô đệm
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={840}
          gap={48}
          left={<CodeCard snippet={GO_SNIPPETS.bufferedDone} delay={20} accent={COLOR.go} highlight={[0, 3, 8]} />}
          right={
            <Bullets
              gap={20}
              items={[
                { at: 150, tag: "không đệm", text: "make(chan error): gửi đứng chờ tới khi có người nhận. Hai bên bắt tay — đó là cách đồng bộ mặc định.", accent: COLOR.go },
                { at: 210, tag: "đệm 1", text: "make(chan error, 1): gửi xong đi luôn, miễn còn chỗ. Người nhận lấy sau cũng được.", accent: COLOR.money },
                { at: 270, tag: "vì sao ở đây", text: "Test chờ 3 giây rồi bỏ đi. Goroutine về muộn vẫn gửi được, không treo mãi ở done <- err. Không đệm là rò một goroutine.", accent: COLOR.danger },
                { at: 350, tag: "time.After", text: "Trả một channel sẽ có tin sau 3 giây. Đặt nó vào select là có timeout — không cần cờ, không cần vòng lặp đếm.", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
