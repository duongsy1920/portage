import React, { useMemo, useState } from "react";
import { EPISODES } from "../catalog";
import { HOC_QUIZ } from "../../../src/data/quiz";
import { href } from "../router";
import { useProgress } from "../store";
import { dueAfter, type Grade } from "../progress";

/**
 * Review cards: every interview question of every episode watched, plus the
 * 17 questions of HOC.md. A card is due when it has never been graded or its
 * due date has passed. One session is at most ten cards (HUB-PLAN §4.3(c)).
 */
type Card = { id: string; from: string; q: string; hints: readonly string[]; episode?: string };

const GRADES: { g: Grade; label: string; days: number }[] = [
  { g: "chua", label: "chưa", days: 1 },
  { g: "mo-ho", label: "mơ hồ", days: 3 },
  { g: "nho", label: "nhớ", days: 7 },
];

export const ReviewPage: React.FC = () => {
  const { progress, update } = useProgress();
  const [showHints, setShowHints] = useState(false);
  const [done, setDone] = useState<string[]>([]);

  const cards = useMemo<Card[]>(() => {
    const out: Card[] = [];
    for (const e of EPISODES) {
      if (!progress.watched[e.id]) continue;
      e.spec.recap.asked.forEach((q, i) =>
        out.push({ id: `${e.id}#${i}`, from: `${e.label} · ${e.spec.title}`, q, hints: e.spec.recap.points.map((p) => p.text), episode: e.id }),
      );
    }
    for (const q of HOC_QUIZ) out.push({ id: `hoc#${q.n}`, from: `HOC.md · ${q.group} · câu ${q.n}`, q: q.q, hints: [], episode: q.episodes[0] });
    return out;
  }, [progress.watched]);

  const now = Date.now();
  const due = cards.filter((c) => !done.includes(c.id) && (!progress.review[c.id] || Date.parse(progress.review[c.id].due) <= now));
  const session = due.slice(0, 10);
  const card = session[0];
  const total = cards.length;
  const seen = cards.filter((c) => progress.review[c.id]).length;

  const grade = (c: Card, g: Grade) => {
    update((s) => ({ ...s, review: { ...s.review, [c.id]: { due: dueAfter(g), grade: g, at: new Date().toISOString() } } }));
    setDone((d) => [...d, c.id]);
    setShowHints(false);
  };

  return (
    <>
      <h1>
        Ôn tập — <em>trả lời không nhìn code</em>
      </h1>
      <p className="sub">
        {total} thẻ: câu phỏng vấn của các tập đã xem, cộng 17 câu của HOC.md. Đã chấm {seen}/{total}. Hôm nay tới hạn: {due.length}
        {due.length > 10 ? " (một buổi tối đa 10)" : ""}.
      </p>

      {total === HOC_QUIZ.length && EPISODES.every((e) => !progress.watched[e.id]) ? (
        <p className="hint">
          Chưa xem tập nào, nên mới chỉ có 17 câu của HOC.md. <a href={href({ page: "map" })}>Về bản đồ</a> xem tập 1 trước.
        </p>
      ) : null}

      {card ? (
        <div className="card review-card" key={card.id}>
          <div className="from">{card.from}</div>
          <div>{card.q}</div>
          <div className="hints">
            {card.hints.length ? (
              <button className="btn ghost" onClick={() => setShowHints((v) => !v)}>
                {showHints ? "Ẩn gợi ý" : "Gợi ý (ý nhớ lại của tập)"}
              </button>
            ) : card.episode ? (
              <a className="btn ghost" href={href({ page: "episode", id: card.episode })}>
                Xem lại tập liên quan
              </a>
            ) : (
              <span className="hint">Câu này chỉ có trong docs — trả lời bằng miệng.</span>
            )}
            {showHints && card.hints.length ? (
              <ol className="points" style={{ marginTop: 10, fontSize: 15 }}>
                {card.hints.map((h) => (
                  <li key={h}>{h}</li>
                ))}
              </ol>
            ) : null}
          </div>
          <div className="btn-row" style={{ marginTop: 16 }}>
            {GRADES.map(({ g, label, days }) => (
              <button key={g} className="btn" onClick={() => grade(card, g)} title={`quay lại sau ${days} ngày`}>
                {label}
              </button>
            ))}
            <span className="hint">
              còn {session.length - 1} thẻ trong buổi này
            </span>
          </div>
        </div>
      ) : (
        <div className="card">
          <div className="clinical" style={{ color: "var(--ok)" }}>
            Hết thẻ tới hạn
          </div>
          <div style={{ marginTop: 8 }}>
            {done.length ? `Vừa chấm ${done.length} thẻ. ` : ""}
            Thẻ “chưa” quay lại sau 1 ngày, “mơ hồ” sau 3, “nhớ” sau 7. Xem thêm tập mới thì có thêm thẻ.
          </div>
        </div>
      )}
    </>
  );
};
