import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { SNIPPETS, GOLDEN } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 3, episode 6, scene 1 — quoted beside actual, with the real numbers.
 *
 * The struct has the two halves side by side on purpose: a quote is a snapshot
 * of what the world looked like on the day it was made, and the whole point of
 * keeping both is that the world moves.
 */
export const P6S1: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 6" source="internal/domain/pricing/reconciliation.go:21">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 260, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Báo giá là một bức ảnh, không phải một phép tính
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={840}
          gap={56}
          align="center"
          left={<CodeCard snippet={SNIPPETS.reconciliation} delay={20} accent={COLOR.money} highlight={[4, 5, 7, 8]} scale={0.9} />}
          right={
            <div style={{ fontSize: 30, lineHeight: 1.6, color: COLOR.textDim }}>
              Báo giá:{" "}
              <span style={{ color: COLOR.text, fontWeight: 700 }}>
                163.22 + {GOLDEN.freightQuotedUSD} USD
              </span>
              <br />
              Thực tế:{" "}
              <span style={{ color: COLOR.text, fontWeight: 700 }}>
                {GOLDEN.actualGoodsUSD} + {GOLDEN.actualFreightUSD} USD
              </span>
              <br />
              <br />
              variance:{" "}
              <span style={{ color: COLOR.danger, fontWeight: 700, fontSize: 40 }}>
                {GOLDEN.varianceUSD} USD
              </span>
            </div>
          }
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={260} accent={COLOR.money}>
          Tỷ giá đổi, cước hãng bay đổi, cân thật khác cân đo. Báo giá giữ nguyên con số
          đã hứa với khách — còn chênh lệch thì phải đo được, chứ không được lờ đi.
        </Callout>
      </div>
    </div>
  </Stage>
);
