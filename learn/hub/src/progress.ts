/**
 * One learner's progress, in this browser only (HUB-PLAN §3.5): what was
 * watched, how each self-test question was graded, when each review card is
 * due. No server, no account. Export/import as JSON so a new machine does not
 * start from zero.
 *
 * Every read is wrapped: localStorage can be absent or throw (private mode,
 * cleared site data), and the page must still render.
 */
export type Grade = "nho" | "mo-ho" | "chua";

export type Progress = {
  v: 1;
  /** episode id → ISO time it was watched to the end */
  watched: Record<string, string>;
  /** episode id → question index → latest grade */
  quiz: Record<string, Record<string, { grade: Grade; at: string }>>;
  /** review card id → when it is due again, and the last grade */
  review: Record<string, { due: string; grade: Grade; at: string }>;
};

const KEY = "learn.hub.progress.v1";

export const empty = (): Progress => ({ v: 1, watched: {}, quiz: {}, review: {} });

export const load = (): Progress => {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return empty();
    const p = JSON.parse(raw) as Partial<Progress>;
    if (p.v !== 1) return empty();
    return { v: 1, watched: p.watched ?? {}, quiz: p.quiz ?? {}, review: p.review ?? {} };
  } catch {
    return empty();
  }
};

export const save = (p: Progress): void => {
  try {
    localStorage.setItem(KEY, JSON.stringify(p));
  } catch {
    /* private mode: the session still works, it just does not persist */
  }
};

/** Days until a card comes back, by how well it went. HUB-PLAN §4.3(c). */
export const INTERVAL_DAYS: Record<Grade, number> = { chua: 1, "mo-ho": 3, nho: 7 };

export const dueAfter = (grade: Grade, from = new Date()): string => {
  const d = new Date(from);
  d.setDate(d.getDate() + INTERVAL_DAYS[grade]);
  return d.toISOString();
};

export const exportJSON = (p: Progress): string => JSON.stringify(p, null, 2);

/** Returns the parsed progress, or a reason it was refused. */
export const importJSON = (text: string): { ok: true; progress: Progress } | { ok: false; why: string } => {
  try {
    const p = JSON.parse(text) as Partial<Progress>;
    if (p.v !== 1 || typeof p.watched !== "object") return { ok: false, why: "không phải file tiến độ của trang này (thiếu v: 1)" };
    return { ok: true, progress: { v: 1, watched: p.watched ?? {}, quiz: p.quiz ?? {}, review: p.review ?? {} } };
  } catch (e) {
    return { ok: false, why: `JSON hỏng: ${String(e)}` };
  }
};
