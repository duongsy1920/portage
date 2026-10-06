import { EPISODES } from "./catalog";
import { HOC_QUIZ } from "../../src/data/quiz";
import type { Progress } from "./progress";

/**
 * Review questions: every interview question of every episode watched, plus
 * the 17 questions of HOC.md. A question is due when it has never been graded
 * or its due date has passed. One sitting is at most ten (HUB-PLAN §4.3(c)).
 * Shared by the review page and the header badge, so both count the same way.
 */
export type ReviewCard = { id: string; from: string; q: string; hints: readonly string[]; episode?: string };

export const SITTING = 10;

export const buildCards = (p: Progress): ReviewCard[] => {
  const out: ReviewCard[] = [];
  for (const e of EPISODES) {
    if (!p.watched[e.id]) continue;
    e.spec.recap.asked.forEach((q, i) =>
      out.push({ id: `${e.id}#${i}`, from: `${e.label} · ${e.spec.title}`, q, hints: e.spec.recap.points.map((x) => x.text), episode: e.id }),
    );
  }
  for (const q of HOC_QUIZ) out.push({ id: `hoc#${q.n}`, from: `HOC.md · ${q.group} · câu ${q.n}`, q: q.q, hints: [], episode: q.episodes[0] });
  return out;
};

export const dueCards = (cards: readonly ReviewCard[], p: Progress, now = Date.now()): ReviewCard[] =>
  cards.filter((c) => !p.review[c.id] || Date.parse(p.review[c.id]!.due) <= now);
