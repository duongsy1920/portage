import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 1, episode 6, scene 1 — the line that should crash and does not.
 *
 * `var m Money` with no value is the shortest way to show that Go has no null.
 * A PHP developer reads it as "uninitialised" and expects a fatal error on the
 * next line; instead it is a perfectly usable zero. That gap is the episode.
 */
export const L1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 6" source="Go không có null — mọi kiểu đều có sẵn một giá trị">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 90, sound: "reveal" }, { at: 240, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Dòng này lẽ ra phải nổ. Nó không nổ.
      </Headline>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 70 }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "var m Money" , note: "không gán gì cả — và đây là code hợp lệ", at: 90, color: COLOR.go },
            { text: "        " },
            { text: "m.Minor()", note: "chạy được, trả về 0", at: 170, color: COLOR.ok },
          ]}
        />

        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "$m = null;" },
            { text: "   " },
            { text: "$m->minor()", note: "bên PHP: lỗi ngay, call on null", at: 240, color: COLOR.danger },
          ]}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.go}>
          Go không có null. Mỗi kiểu có sẵn một giá trị mặc định, gọi là{" "}
          <span style={{ color: COLOR.go }}>zero value</span>, và nó luôn dùng được ngay.
        </Callout>
      </div>
    </div>
  </Stage>
);
