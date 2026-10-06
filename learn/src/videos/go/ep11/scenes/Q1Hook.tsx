import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Cues } from "../../../../components/Sfx";
import { GO_USAGE } from "../../../../data/portage";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 1, episode 11, scene 1 — the line every use case opens with.
 *
 * The viewer has read `if err != nil` for ten episodes; the one shape they
 * still skim over is a function passed as an argument. So the first thing on
 * screen is that exact line from the repository, taken apart piece by piece.
 */
export const Q1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 11" source="internal/app/catalog/register_merchant.go:35 — code thật">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 240, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Tham số thứ hai là một hàm. Mọi use case trong repo mở đầu như thế.
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={34}
          parts={[
            { text: "err := " },
            { text: "h.deps.UoW.InTx", note: "một method bình thường, nhận hai tham số", at: 90, color: COLOR.go },
            { text: "(ctx, " },
            { text: "func(ctx context.Context) error", note: "tham số thứ hai là một HÀM, viết ngay tại chỗ", at: 150, color: COLOR.money },
            { text: " {", note: "thân hàm — chạy trong transaction", at: 200, color: COLOR.ok },
          ]}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={240} accent={COLOR.money}>
          Repo gọi InTx như vậy ở {GO_USAGE.inTxCalls} chỗ, và có {GO_USAGE.funcLiterals} hàm không tên như cái
          ở trên. Đọc được dòng này là đọc được mọi use case.
        </Callout>
      </div>
    </div>
  </Stage>
);
