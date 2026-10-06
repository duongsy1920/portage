import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { Cues } from "../../../../components/Sfx";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { COLOR, TYPE, RADIUS } from "../../../../design/tokens";
import { MONO } from "../../../../design/fonts";
import { EASE } from "../../../../design/motion";
import { GO_USAGE, REPO } from "../../../../data/portage";

/**
 * Season 1, episode 16, scene 1 — four figures, including a zero.
 *
 * Same opening as episode 10, on purpose: numbers first, including the one
 * that is missing. The suite is fast and wide, and it has no benchmark at
 * all — both facts are worth a learner's attention before any explanation.
 */
const FIGURES = [
  { n: GO_USAGE.testFiles, label: "file _test.go", color: COLOR.go, at: 30 },
  { n: GO_USAGE.testFuncs, label: "hàm Test", color: COLOR.go, at: 80 },
  { n: REPO.testsNoDocker, label: "PASS không Docker", color: COLOR.ok, at: 130 },
  { n: GO_USAGE.benchmarks, label: "Benchmark", color: COLOR.textFaint, at: 180 },
] as const;

export const V1Hook: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <Stage eyebrow="MÙA 1 · TẬP 16" source="đếm trên internal/ và cmd/ — go test ./... không PORTAGE_TEST_DSN">
      <Cues items={[{ at: 30, sound: "appear" }, { at: 180, sound: "snap", volume: 0.2 }, { at: 250, sound: "land", volume: 0.26 }]} />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={58}>
          Toàn bộ test chạy dưới hai giây. Vì sao?
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
                opacity: interpolate(frame, [f.at, f.at + 20], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE }),
              }}
            >
              <div style={{ fontFamily: MONO, fontSize: 108, fontWeight: 700, color: f.color, lineHeight: 1 }}>{f.n}</div>
              <div style={{ fontFamily: MONO, fontSize: TYPE.label, color: COLOR.textDim, marginTop: 18 }}>{f.label}</div>
            </Interactive.Div>
          ))}
        </div>

        <div style={{ marginBottom: 26 }}>
          <Callout delay={250} accent={COLOR.money}>
            Không Docker, không mạng, không database. Câu trả lời là tập 5 nhìn từ phía test: mỗi port có một adapter chạy trong RAM.
            Và số 0 kia là thật — repo chưa có benchmark nào.
          </Callout>
        </div>
      </div>
    </Stage>
  );
};
