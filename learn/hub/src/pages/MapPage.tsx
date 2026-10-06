import React, { useMemo, useRef } from "react";
import { EPISODES, PLANNED_EPISODES, SEASON_INFO, mmss, type Episode } from "../catalog";
import { PATH } from "../../../src/data/path";
import { href } from "../router";
import { useProgress } from "../store";
import { exportJSON, importJSON, empty } from "../progress";

/** An episode is "confident" when every interview question was last graded "nhớ". */
export const isConfident = (e: Episode, quiz: Record<string, Record<string, { grade: string }>>): boolean => {
  const g = quiz[e.id] ?? {};
  return e.spec.recap.asked.length > 0 && e.spec.recap.asked.every((_, i) => g[String(i)]?.grade === "nho");
};

export const MapPage: React.FC = () => {
  const { progress, update } = useProgress();
  const fileRef = useRef<HTMLInputElement>(null);

  const watched = EPISODES.filter((e) => progress.watched[e.id]).length;
  const confident = EPISODES.filter((e) => isConfident(e, progress.quiz)).length;
  const next = useMemo(
    () => EPISODES.find((e) => !progress.watched[e.id]) ?? EPISODES.find((e) => !isConfident(e, progress.quiz)),
    [progress],
  );

  const download = () => {
    const blob = new Blob([exportJSON(progress)], { type: "application/json" });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = `hoc-go-tien-do-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(a.href);
  };
  const onImport = async (file: File | undefined) => {
    if (!file) return;
    const r = importJSON(await file.text());
    if (!r.ok) {
      alert(`Không nhập được: ${r.why}`);
      return;
    }
    update(() => r.progress);
  };

  return (
    <>
      <h1>
        Học Go — <em>từ Symfony sang Go</em>
      </h1>
      <p className="sub">
        {EPISODES.length} tập, ba mùa, mỗi tập một video và một bản in. Đã xem <b>{watched}/{EPISODES.length}</b> · tự tin{" "}
        <b>{confident}/{EPISODES.length}</b>.
      </p>

      {next ? (
        <div className="card next">
          <div>
            <div className="k">{watched === 0 ? "Bắt đầu ở đây" : "Tiếp theo"}</div>
            <div className="t">
              {next.label} · {next.spec.title}
            </div>
            <div className="m">
              {mmss(next.seconds)} · {next.spec.scenes.length} cảnh
              {progress.watched[next.id] ? " · đã xem, chưa tự kiểm tra" : ""}
            </div>
          </div>
          <a className="btn primary go" href={href({ page: "episode", id: next.id })}>
            Xem
          </a>
        </div>
      ) : (
        <div className="card next">
          <div className="t">Đi hết rồi. Còn lại là ôn và làm thật.</div>
          <a className="btn primary go" href={href({ page: "review" })}>
            Ôn tập
          </a>
        </div>
      )}

      {SEASON_INFO.map((season) => {
        const eps = EPISODES.filter((e) => e.season.index === season.index);
        const done = eps.filter((e) => progress.watched[e.id]).length;
        return (
          <section className="season" key={season.folder} aria-label={season.label}>
            <div className="head">
              <span className="name">{season.label}</span>
              <span className="count">
                {done}/{eps.length} đã xem · {mmss(eps.reduce((s, e) => s + e.seconds, 0))}
              </span>
            </div>
            <ul className="dots">
              {eps.map((e, i) => {
                const cls = ["dot", isConfident(e, progress.quiz) ? "confident" : progress.watched[e.id] ? "watched" : "", next?.id === e.id ? "next" : ""]
                  .filter(Boolean)
                  .join(" ");
                return (
                  <li key={e.id}>
                    <a className={cls} href={href({ page: "episode", id: e.id })} title={`${e.spec.title} · ${mmss(e.seconds)}`} style={{ ["--i" as string]: i }}>
                      <span className="n">{e.number}</span>
                      <span className="t">{e.spec.title}</span>
                    </a>
                  </li>
                );
              })}
              {season.index === 0
                ? PLANNED_EPISODES.map((p, i) => (
                    <li key={p.key}>
                      <span className="dot planned" title={`${p.title} — đã lên kế hoạch (PLAN-BO-SUNG §5), chưa dựng`} style={{ ["--i" as string]: eps.length + i }}>
                        <span className="n">{p.key}</span>
                        <span className="t">{p.title}</span>
                      </span>
                    </li>
                  ))
                : null}
            </ul>
          </section>
        );
      })}
      <div className="legend">
        <span className="w">đã xem</span>
        <span className="c">tự tin (đã tự kiểm tra)</span>
        <span className="x">tiếp theo</span>
        <span>chưa</span>
        <span className="p">đã lên kế hoạch, chưa dựng</span>
      </div>

      <h2>Theo buổi của HOC.md</h2>
      <p className="sub">
        Mười buổi của <code>docs/HOC.md</code>, và tập nào đi với buổi nào. Quy tắc duy nhất của file đó: mỗi buổi phải sửa được một dòng
        và làm test đỏ.
      </p>
      <div className="sessions">
        {PATH.map((s) => (
          <div className="card session" key={s.n}>
            <div className="k">Buổi {s.n}</div>
            <div className="t">{s.title}</div>
            <div className="g">{s.goal}</div>
            <ul>
              {s.read.map((r) => (
                <li key={r}>đọc {r}</li>
              ))}
              {s.do.map((d) => (
                <li key={d}>{d}</li>
              ))}
            </ul>
            {s.episodes.length ? (
              <div className="eps">
                {s.episodes.map((id) => {
                  const e = EPISODES.find((x) => x.id === id);
                  return e ? (
                    <a key={id} className={`chip${progress.watched[id] ? " watched" : ""}`} href={href({ page: "episode", id })} title={e.spec.title}>
                      {e.season.short} {e.number}
                    </a>
                  ) : (
                    <span key={id} className="chip bad">
                      {id}?
                    </span>
                  );
                })}
              </div>
            ) : null}
          </div>
        ))}
      </div>

      <h2>Tiến độ của bạn</h2>
      <p className="sub">
        Lưu trong trình duyệt này, không có tài khoản, không có server. Xuất ra file để mang sang máy khác.
      </p>
      <div className="btn-row">
        <button className="btn" onClick={download}>
          Xuất JSON
        </button>
        <button className="btn" onClick={() => fileRef.current?.click()}>
          Nhập JSON
        </button>
        <input ref={fileRef} type="file" accept="application/json" hidden onChange={(ev) => void onImport(ev.target.files?.[0])} />
        <button
          className="btn ghost"
          onClick={() => {
            if (confirm("Xoá toàn bộ tiến độ trên trình duyệt này?")) update(() => empty());
          }}
        >
          Xoá
        </button>
      </div>
    </>
  );
};
