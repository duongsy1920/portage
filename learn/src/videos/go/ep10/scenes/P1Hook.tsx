import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Cues } from "../../../../components/Sfx";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, TYPE, RADIUS } from "../../../../design/tokens";
import { MONO } from "../../../../design/fonts";
import { EASE } from "../../../../design/motion";
import { GO_USAGE } from "../../../../data/portage";

/**
 * Season 1, episode 10, scene 1 — the numbers, including the zero.
 *
 * The zero is the point. Every Go tutorial opens with channels; this codebase
 * has none, and saying so first is the honest way in. It also sets up the
 * episode's real question: when would you reach for one?
 */
const FIGURES = [
  { n: GO_USAGE.goroutines, label: "goroutine", color: COLOR.go, at: 30 },
  { n: GO_USAGE.mutex, label: "sync.Mutex", color: COLOR.danger, at: 80 },
  { n: GO_USAGE.channels, label: "channel", color: COLOR.textFaint, at: 140 },
] as const;

export const P1Hook: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <Stage eyebrow="MÙA 1 · TẬP 10" source="đếm trên internal/ và cmd/, không tính file test">
      <Cues items={[{ at: 30, sound: "appear" }, { at: 140, sound: "snap", volume: 0.2 }, { at: 230, sound: "land", volume: 0.26 }]} />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={58}>
          Repo này có bao nhiêu channel?
        </Headline>

        <div style={{ flex: 1, display: "flex", alignItems: "center", gap: 90 }}>
          {FIGURES.map((f) => (
            <Interactive.Div
              key={f.label}
              name={f.label}
              style={{
                flex: 1,
                padding: "44px 40px",
                borderRadius: RADIUS.md,
                backgroundColor: COLOR.surface,
                borderLeft: `8px solid ${f.color}`,
                opacity: interpolate(frame, [f.at, f.at + 20], [0, 1], {
                  extrapolateLeft: "clamp",
                  extrapolateRight: "clamp",
                  easing: EASE,
                }),
              }}
            >
              <div style={{ fontFamily: MONO, fontSize: 108, fontWeight: 700, color: f.color, lineHeight: 1 }}>
                {f.n}
              </div>
              <div style={{ fontFamily: MONO, fontSize: TYPE.label, color: COLOR.textDim, marginTop: 18 }}>
                {f.label}
              </div>
            </Interactive.Div>
          ))}
        </div>

        <div style={{ marginBottom: 26 }}>
          <Callout delay={230} accent={COLOR.money}>
            Không. Một cái cũng không. Mọi bài dạy Go đều mở đầu bằng channel, còn hệ
            thật này chạy bằng khoá — và phỏng vấn thì vẫn sẽ hỏi channel.
          </Callout>
        </div>
      </div>
    </Stage>
  );
};
