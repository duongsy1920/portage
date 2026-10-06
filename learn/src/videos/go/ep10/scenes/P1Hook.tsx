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
 * Season 1, episode 10, scene 1 — the numbers, including the zero AND the four.
 *
 * The zero is the point: every Go tutorial opens with channels, and this
 * codebase never creates one. But it does READ from four — <-ctx.Done() and
 * <-ticker.C inside two select blocks — and a learner who opens worker.go on
 * day one meets exactly that line. Saying "0 channel" alone was true of
 * declarations and false of what they would see, so both numbers are shown.
 */
const FIGURES = [
  { n: GO_USAGE.goroutines, label: "goroutine", color: COLOR.go, at: 30 },
  { n: GO_USAGE.mutex, label: "sync.Mutex", color: COLOR.danger, at: 80 },
  { n: GO_USAGE.channelsDeclared, label: "channel tự tạo", color: COLOR.textFaint, at: 140 },
  { n: GO_USAGE.channelsRead, label: "lần đọc channel", color: COLOR.money, at: 185 },
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

        <div style={{ flex: 1, display: "flex", alignItems: "center", gap: 44 }}>
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
            Không tự tạo cái nào — hệ thật này chạy bằng khoá. Nhưng mở worker.go sẽ gặp{" "}
            <span style={{ color: COLOR.money }}>{"case <-ctx.Done():"}</span> — bốn lần đọc từ
            channel của context và ticker, chỉ để biết khi nào dừng. Và phỏng vấn thì vẫn sẽ hỏi channel.
          </Callout>
        </div>
      </div>
    </Stage>
  );
};
