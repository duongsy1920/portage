import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, scene 6 — the parameter that is in 361 places and has no PHP twin.
 *
 * ctx looks like noise until you know the process outlives the request. Once
 * that is established, it is obviously the missing piece: a long-lived process
 * needs a way to be told "this particular piece of work is over".
 */
export const G6Context: React.FC = () => {
  return (
    <Stage
      eyebrow="MÙA 1 · TẬP 1 · CẢNH 6 · TỪ VỰNG"
      source={`repo có ${GO_USAGE.ctx} chỗ nhận context.Context`}
    >
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 140, sound: "reveal" },
          { at: 300, sound: "appear", volume: 0.14 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Vì sao hàm nào cũng có ctx ở tham số đầu
        </Headline>

        <div style={{ marginTop: 30, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={700}
            gap={60}
            left={<CodeCard snippet={GO_SNIPPETS.ctx} delay={20} accent={COLOR.go} highlight={[1, 4]} />}
            right={
              <div style={{ display: "flex", flexDirection: "column", gap: 28 }}>
                <TermCard
                  delay={140}
                  accent={COLOR.go}
                  term="context.Context"
                  plain="ctx mang lệnh dừng và hạn chót đi xuống mọi hàm bên dưới nó."
                  symfony="PHP không có. max_execution_time giết cả script, không dừng riêng một việc."
                />
                <Bullets
                  gap={20}
                  items={[
                    {
                      at: 310,
                      tag: "ví dụ",
                      text: "Khách đóng tab. ctx bị huỷ, câu SQL đang chạy dừng luôn, không chạy phí nữa.",
                      accent: COLOR.go,
                    },
                    {
                      at: 360,
                      tag: "quy ước",
                      text: "ctx luôn là tham số đầu tiên, luôn tên là ctx, và không bao giờ lưu vào struct.",
                      accent: COLOR.go,
                    },
                    {
                      at: 410,
                      tag: "trong repo",
                      text: "Domain không nhận ctx: nó chỉ tính toán, không chạm mạng hay đĩa.",
                      accent: COLOR.ok,
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
};
