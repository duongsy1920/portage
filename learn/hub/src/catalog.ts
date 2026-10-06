import { SEASONS, componentFor } from "../../src/videos";
import { totalFrames, type EpisodeSpec } from "../../src/videos/registry";
import { BEAT, FPS } from "../../src/design/tokens";
import type { Progress } from "./progress";

/**
 * The series as the hub sees it: SEASONS flattened, each episode knowing its
 * season, its number, its length and where its cheatsheet lives. Everything
 * here is DERIVED from the specs — the hub never keeps a second list of
 * episodes (HUB-PLAN §3.3), so a new spec.ts shows up in the sidebar by itself.
 */
export type SeasonInfo = { index: number; folder: string; label: string; short: string; blurb: string };

export const SEASON_INFO: readonly SeasonInfo[] = [
  { index: 0, folder: "Mua1-Go", label: "Mùa 1 · Go", short: "Go", blurb: "Go cho người viết PHP" },
  { index: 1, folder: "Mua2-DDD", label: "Mùa 2 · DDD", short: "DDD", blurb: "DDD, bằng ví dụ Symfony đã quen" },
  { index: 2, folder: "Mua3-Portage", label: "Mùa 3 · Portage", short: "Portage", blurb: "Vì sao project thật này trông như vậy" },
];

export type Episode = {
  spec: EpisodeSpec;
  id: string;
  season: SeasonInfo;
  number: number;
  /** "Mùa 1 · Tập 3" */
  label: string;
  frames: number;
  seconds: number;
  /** cheatsheet/<plate>.html, derived from outDir: mua1-go/ep03-… → go-ep03 */
  plate: string;
  /** first frame of each scene, transitions accounted for */
  sceneStarts: readonly number[];
};

const PLATE_PREFIX: Record<string, string> = { "mua1-go": "go", "mua2-ddd": "ddd", "mua3-portage": "portage" };

const plateOf = (outDir: string): string => {
  const [season, ep] = outDir.split("/");
  const n = /^ep(\d+)/.exec(ep ?? "")?.[1];
  const prefix = PLATE_PREFIX[season ?? ""];
  if (!prefix || !n) throw new Error(`outDir lạ: ${outDir}`);
  return `${prefix}-ep${n}`;
};

const sceneStartsOf = (spec: EpisodeSpec): number[] => {
  const starts: number[] = [];
  let at = 0;
  spec.scenes.forEach((s, i) => {
    starts.push(at);
    at += s.durationInFrames - (i < spec.scenes.length - 1 ? BEAT.transition : 0);
  });
  return starts;
};

export const EPISODES: readonly Episode[] = SEASONS.flatMap((season, si) =>
  season.specs.map((spec, ei) => {
    const frames = totalFrames(spec);
    const info = SEASON_INFO[si];
    if (!info || info.folder !== season.folder) throw new Error(`SEASONS và SEASON_INFO lệch nhau ở ${season.folder}`);
    return {
      spec,
      id: spec.id,
      season: info,
      number: ei + 1,
      label: `Mùa ${si + 1} · Tập ${ei + 1}`,
      frames,
      // floor, not round: a 128.47 s video reads 2:08 on every player, and in CURRICULUM.md §11
      seconds: Math.floor(frames / FPS),
      plate: plateOf(spec.outDir),
      sceneStarts: sceneStartsOf(spec),
    };
  }),
);

export const byId = (id: string): Episode | undefined => EPISODES.find((e) => e.id === id);

export const mmss = (seconds: number): string =>
  `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;

/** "Đã thuộc": every interview question of the episode was last graded "nhớ". */
export const isConfident = (e: Episode, quiz: Progress["quiz"]): boolean => {
  const g = quiz[e.id] ?? {};
  return e.spec.recap.asked.length > 0 && e.spec.recap.asked.every((_, i) => g[String(i)]?.grade === "nho");
};

/** What "#/" opens: the first unwatched episode, else the first not yet learned, else the first. */
export const nextEpisode = (p: Progress): Episode =>
  EPISODES.find((e) => !p.watched[e.id]) ?? EPISODES.find((e) => !isConfident(e, p.quiz)) ?? EPISODES[0]!;

/** Episode components, built once per spec: buildEpisode returns a new component each call. */
const built = new Map<string, React.FC>();
export const componentOf = (e: Episode): React.FC => {
  let c = built.get(e.id);
  if (!c) {
    c = componentFor(e.spec);
    built.set(e.id, c);
  }
  return c;
};
