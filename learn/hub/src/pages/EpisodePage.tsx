import React, { useEffect, useMemo, useRef, useState } from "react";
import { Player, type PlayerRef } from "@remotion/player";
import { FPS, HEIGHT, WIDTH } from "../../../src/design/tokens";
import { EPISODES, byId, componentOf, mmss, type Episode } from "../catalog";
import { resolveAnchor } from "../anchors";
import { VERIFIED } from "../../../src/data/meta";
import { href } from "../router";
import { useProgress } from "../store";
import { dueAfter, type Grade } from "../progress";

type Tab = "video" | "plate" | "code" | "quiz";
const TABS: { key: Tab; label: string }[] = [
  { key: "video", label: "Video" },
  { key: "plate", label: "Bản in" },
  { key: "code", label: "Neo vào code" },
  { key: "quiz", label: "Tự kiểm tra" },
];

const GRADES: { g: Grade; label: string }[] = [
  { g: "nho", label: "nhớ" },
  { g: "mo-ho", label: "mơ hồ" },
  { g: "chua", label: "chưa" },
];

export const EpisodePage: React.FC<{ id: string }> = ({ id }) => {
  const e = byId(id);
  if (!e) {
    return (
      <>
        <h1>Không có tập nào tên {id}</h1>
        <p className="sub">
          <a href={href({ page: "map" })}>Về bản đồ</a>
        </p>
      </>
    );
  }
  return <EpisodeView key={e.id} e={e} />;
};

const EpisodeView: React.FC<{ e: Episode }> = ({ e }) => {
  const { progress, update } = useProgress();
  const ref = useRef<PlayerRef>(null);
  const [tab, setTab] = useState<Tab>("video");
  const [frame, setFrame] = useState(0);
  const [showPoints, setShowPoints] = useState(false);
  const Episode = useMemo(() => componentOf(e), [e]);

  const idx = EPISODES.findIndex((x) => x.id === e.id);
  const prev = EPISODES[idx - 1];
  const next = EPISODES[idx + 1];

  // Watched = reached the last frame. No button to press: finishing is the proof.
  useEffect(() => {
    const p = ref.current;
    if (!p) return;
    const onFrame = (ev: { detail: { frame: number } }) => {
      setFrame(ev.detail.frame);
      if (ev.detail.frame >= e.frames - 2 && !progress.watched[e.id]) {
        update((s) => ({ ...s, watched: { ...s.watched, [e.id]: new Date().toISOString() } }));
      }
    };
    const onEnded = () => {
      if (!progress.watched[e.id]) update((s) => ({ ...s, watched: { ...s.watched, [e.id]: new Date().toISOString() } }));
    };
    p.addEventListener("frameupdate", onFrame);
    p.addEventListener("ended", onEnded);
    return () => {
      p.removeEventListener("frameupdate", onFrame);
      p.removeEventListener("ended", onEnded);
    };
  }, [e, progress.watched, update]);

  const sceneNow = e.sceneStarts.reduce((acc, start, i) => (frame >= start ? i : acc), 0);

  // Test hook for scripts/hub-check.mjs: seek without driving the UI. Harmless in use.
  useEffect(() => {
    (window as unknown as { __hub?: unknown }).__hub = {
      seek: (f: number) => {
        ref.current?.seekTo(f);
        ref.current?.play();
      },
      frames: e.frames,
    };
    return () => {
      delete (window as unknown as { __hub?: unknown }).__hub;
    };
  }, [e]);

  const grade = (i: number, g: Grade) =>
    update((s) => ({
      ...s,
      quiz: { ...s.quiz, [e.id]: { ...(s.quiz[e.id] ?? {}), [String(i)]: { grade: g, at: new Date().toISOString() } } },
      review: { ...s.review, [`${e.id}#${i}`]: { due: dueAfter(g), grade: g, at: new Date().toISOString() } },
    }));
  const graded = progress.quiz[e.id] ?? {};
  const allGood = e.spec.recap.asked.every((_, i) => graded[String(i)]?.grade === "nho");

  return (
    <>
      <div className="ep-head">
        <div>
          <div className="k">
            {e.season.label} · tập {e.number}
          </div>
          <h1 style={{ marginTop: 6 }}>{e.spec.title}</h1>
          <p className="sub" style={{ marginBottom: 0 }}>
            {mmss(e.seconds)} · {e.spec.scenes.length} cảnh
            {progress.watched[e.id] ? <span className="ok"> · đã xem</span> : null}
            {allGood ? <span className="ok"> · tự tin</span> : null}
          </p>
        </div>
        <div className="btn-row">
          {prev ? (
            <a className="btn ghost" href={href({ page: "episode", id: prev.id })}>
              ← {prev.season.short} {prev.number}
            </a>
          ) : null}
          {next ? (
            <a className="btn" href={href({ page: "episode", id: next.id })}>
              {next.season.short} {next.number} →
            </a>
          ) : null}
        </div>
      </div>

      <div className="ep-layout" style={{ marginTop: 18 }}>
        <div>
          <div className="player">
            <Player
              ref={ref}
              component={Episode}
              durationInFrames={e.frames}
              fps={FPS}
              compositionWidth={WIDTH}
              compositionHeight={HEIGHT}
              controls
              clickToPlay
              style={{ width: "100%" }}
            />
          </div>
        </div>
        <ol className="scenes" aria-label="Cảnh">
          {e.spec.scenes.map((s, i) => (
            <li key={s.name} className={i === sceneNow ? "now" : ""}>
              <button
                onClick={() => {
                  ref.current?.seekTo(e.sceneStarts[i] ?? 0);
                  ref.current?.play();
                }}
              >
                <span className="t">{s.name}</span>
                <span className="at">{mmss(Math.round((e.sceneStarts[i] ?? 0) / FPS))}</span>
              </button>
            </li>
          ))}
        </ol>
      </div>

      <div className="tabs" role="tablist">
        {TABS.map((t) => (
          <button key={t.key} role="tab" aria-selected={tab === t.key} onClick={() => setTab(t.key)}>
            {t.label}
          </button>
        ))}
      </div>

      {tab === "video" ? (
        <p className="hint">
          Xem hết cảnh cuối thì tập tự thành “đã xem”. Bấm một cảnh ở cột bên để tua tới đúng chỗ. Bản in và câu hỏi ở hai tab kế.
        </p>
      ) : null}

      {tab === "plate" ? <PlateView e={e} /> : null}

      {tab === "code" ? (
        <div>
          <p className="sub">
            Mọi dòng code và con số lên hình đều được kiểm lại bằng máy ở commit{" "}
            <a href={`${VERIFIED.repo}/tree/${VERIFIED.commit}`}>{VERIFIED.short}</a> ({VERIFIED.date}). Link mở đúng commit đó, không phải{" "}
            <code>main</code>: số dòng trôi.
          </p>
          {e.spec.anchors.length === 0 ? <p className="hint">Tập này không chiếu snippet nào từ data/portage.ts.</p> : null}
          {e.spec.anchors.map((name) => {
            const a = resolveAnchor(name);
            if (a.kind === "figure")
              return (
                <div className="figure" key={name}>
                  <span className="n">{a.value}</span>
                  <span className="l">{name.replace("GO_USAGE.", "")}</span>
                </div>
              );
            if (a.kind === "snippet")
              return (
                <div className={`card anchor${a.snippet.lang === "php" ? " php" : ""}`} key={name}>
                  <div className="p">
                    <span>{a.snippet.path}</span>
                    {a.snippet.line ? <span>dòng {a.snippet.line}</span> : null}
                    {a.url ? <a href={a.url}>mở trên GitHub ↗</a> : <span className="hint">minh hoạ, không phải code repo</span>}
                  </div>
                  <pre>
                    <code>{a.snippet.code.join("\n")}</code>
                  </pre>
                </div>
              );
            return (
              <div className="card anchor bad" key={name}>
                <div className="p">{name} — không tìm thấy trong data/portage.ts</div>
              </div>
            );
          })}
        </div>
      ) : null}

      {tab === "quiz" ? (
        <div>
          <div className="btn-row" style={{ marginBottom: 10 }}>
            <button className="btn ghost" onClick={() => setShowPoints((v) => !v)}>
              {showPoints ? "Ẩn" : "Hiện"} {e.spec.recap.points.length} ý nhớ lại
            </button>
            <span className="hint">Thử trả lời trước, rồi mới hiện.</span>
          </div>
          {showPoints ? (
            <ol className="points">
              {e.spec.recap.points.map((p) => (
                <li key={p.at}>{p.text}</li>
              ))}
            </ol>
          ) : null}
          <div className="card ask">
            <div className="clinical" style={{ color: "var(--money)", marginBottom: 6 }}>
              Phỏng vấn sẽ hỏi
            </div>
            {e.spec.recap.asked.map((q, i) => (
              <div className="q" key={q}>
                <div className="text">{q}</div>
                <div className="grade" role="group" aria-label={`Tự chấm câu ${i + 1}`}>
                  {GRADES.map(({ g, label }) => (
                    <button key={g} className={`${g}${graded[String(i)]?.grade === g ? " on" : ""}`} onClick={() => grade(i, g)}>
                      {label}
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </div>
          <p className="hint" style={{ marginTop: 10 }}>
            Chấm “nhớ” cả {e.spec.recap.asked.length} câu thì tập thành “tự tin”. Câu “chưa” quay lại trong ôn tập sau 1 ngày, “mơ hồ” sau 3, “nhớ” sau 7.
          </p>
        </div>
      ) : null}
    </>
  );
};

/** The A3 plate, scaled to the available width. It is the same HTML the PDF is printed from. */
const PlateView: React.FC<{ e: Episode }> = ({ e }) => {
  const box = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(0.5);
  useEffect(() => {
    const el = box.current;
    if (!el) return;
    const fit = () => setScale(el.clientWidth / 1587);
    fit();
    const ro = new ResizeObserver(fit);
    ro.observe(el);
    return () => ro.disconnect();
  }, []);
  const src = `cheatsheet/${e.plate}.html`;
  return (
    <div>
      <div className="btn-row" style={{ marginBottom: 10 }}>
        <a className="btn" href={src} target="_blank" rel="noreferrer">
          Mở toàn màn ↗
        </a>
        <span className="hint">Khổ A3 ngang, 1587×1123. In từ trang mở toàn màn (Ctrl+P), hoặc lấy PDF trong out/ sau khi render.</span>
      </div>
      <div className="plate-frame" ref={box} style={{ height: Math.round(1123 * scale) }}>
        <iframe title={`Bản in ${e.label}`} src={src} style={{ transform: `scale(${scale})` }} loading="lazy" />
      </div>
    </div>
  );
};
