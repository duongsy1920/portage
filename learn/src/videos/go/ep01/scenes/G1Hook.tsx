import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Body } from "../../../../components/Text";
import { Cues } from "../../../../components/Sfx";
import { COLOR, TYPE } from "../../../../design/tokens";
import { MONO } from "../../../../design/fonts";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { EASE } from "../../../../design/motion";

/**
 * Season 1, scene 1 — the promise.
 *
 * The series opens on the ONE difference that explains most of what looks
 * strange in Go to a PHP developer. Everything after this episode (pools,
 * mutexes, ctx, main() as a container) is a consequence of it, so it is worth
 * the whole first episode.
 */
export const G1Hook: React.FC = () => {
  const frame = useCurrentFrame();

  return (
    <Stage eyebrow="MÙA 1 · GO CHO NGƯỜI VIẾT PHP · TẬP 1">
      <Cues items={[{ at: 6, sound: "reveal" }, { at: 96, sound: "appear" }]} />
      <div
        style={{
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          gap: 56,
          textAlign: "center",
        }}
      >
        <div style={{ maxWidth: 1450 }}>
          <Headline delay={6} size={80}>
            PHP chết sau mỗi request. Go thì không.
          </Headline>
        </div>

        <Interactive.Div
          name="Two words"
          style={{
            display: "flex",
            gap: 26,
            alignItems: "center",
            fontFamily: MONO,
            fontSize: TYPE.sub,
            fontWeight: 700,
            opacity: interpolate(frame, [96, 120], [0, 1], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: EASE,
            }),
          }}
        >
          <span style={{ color: COLOR.php }}>dựng lại mỗi request</span>
          <span style={{ color: COLOR.textFaint, fontSize: TYPE.label }}>⟷</span>
          <span style={{ color: COLOR.go }}>dựng một lần, sống mãi</span>
        </Interactive.Div>

        <div style={{ maxWidth: 1280 }}>
          <Body delay={150}>
            Đây là khác biệt gốc, và gần như mọi thứ trông lạ trong code Go đều
            là hệ quả của nó: vì sao có pool kết nối, vì sao phải khoá biến
            chung, vì sao hàm nào cũng nhận ctx, vì sao không có container.
          </Body>
        </div>
      </div>
    </Stage>
  );
};
