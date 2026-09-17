import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 1, episode 8, scene 1 — the line that loses your data.
 *
 * `append` returning a new slice instead of mutating is the single most common
 * first-week Go bug, and the compiler says nothing. Showing the forgotten
 * assignment first, before any theory about backing arrays, is the order that
 * makes the theory worth listening to.
 */
export const N1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 8" source="bẫy số một của tuần đầu học Go">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 100, sound: "crack", volume: 0.22 }, { at: 220, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Dòng này mất sạch dữ liệu, và không ai báo
      </Headline>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 70 }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "append(e.pending, ev)", note: "sai: append trả về slice mới, ở đây bị vứt", at: 100, color: COLOR.danger },
          ]}
        />
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "e.pending = ", note: "đúng: phải gán lại", at: 220, color: COLOR.ok },
            { text: "append(e.pending, ev)" },
          ]}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.danger}>
          Cả hai dòng đều biên dịch được. Dòng trên chạy im lặng và không thêm được gì
          vào danh sách sự kiện — đơn hàng mất event, và không có lỗi nào để lần.
        </Callout>
      </div>
    </div>
  </Stage>
);
