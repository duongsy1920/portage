import React, { useState } from "react";
import { BookOpen, CheckCircle2, Circle, CircleCheckBig, ExternalLink, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { EPISODES, SEASON_INFO, isConfident, mmss, type Episode } from "../catalog";
import { PATH } from "../../../src/data/path";
import { PLATES, PLATE_FILES, EXERCISES, EXERCISES_DOC } from "../../../src/data/roadmap";
import { docURL } from "../../../src/data/meta";
import { href } from "../router";
import { useProgress } from "../store";
import { cn } from "@/lib/utils";

/**
 * The course outline: every episode on one line, grouped by season or by the
 * study session of docs/HOC.md. The learner's place is the highlighted row;
 * what they have watched and learned is the icon in front of it.
 */
type Grouping = "season" | "session";

/** inSheet: the drawer draws its own close button top-right; the progress line must leave it room. */
export const Sidebar: React.FC<{ currentId?: string; onNavigate?: () => void; inSheet?: boolean }> = ({ currentId, onNavigate, inSheet }) => {
  const { progress } = useProgress();
  const [grouping, setGrouping] = useState<Grouping>("season");

  const watched = EPISODES.filter((e) => progress.watched[e.id]).length;
  const learned = EPISODES.filter((e) => isConfident(e, progress.quiz)).length;

  const Row: React.FC<{ e: Episode }> = ({ e }) => {
    const done = isConfident(e, progress.quiz);
    const seen = Boolean(progress.watched[e.id]);
    const current = e.id === currentId;
    return (
      <li>
        <a
          href={href({ page: "episode", id: e.id })}
          onClick={onNavigate}
          aria-current={current ? "page" : undefined}
          data-testid="episode-row"
          data-watched={seen ? "1" : undefined}
          title={`${e.label} · ${e.spec.title}`}
          className={cn(
            "flex items-center gap-2.5 rounded-md px-2 py-1.5 text-sm text-sidebar-foreground/90 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
            current && "bg-sidebar-accent font-medium text-sidebar-accent-foreground",
          )}
        >
          {done ? (
            <CircleCheckBig className="size-4 shrink-0 text-done" aria-label="đã thuộc" />
          ) : seen ? (
            <CheckCircle2 className="size-4 shrink-0 text-muted-foreground" aria-label="đã xem" />
          ) : (
            <Circle className="size-4 shrink-0 text-muted-foreground/50" aria-hidden="true" />
          )}
          <span className="w-5 shrink-0 font-mono text-xs text-muted-foreground">{e.number}</span>
          <span className="min-w-0 flex-1 truncate">{e.spec.title}</span>
          <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{mmss(e.seconds)}</span>
        </a>
      </li>
    );
  };

  return (
    <div className="flex h-full flex-col" data-testid="sidebar">
      <div className={cn("px-4 pt-4 pb-3", inSheet && "pr-12")}>
        <div className="flex flex-wrap items-baseline justify-between gap-x-3 text-sm">
          <span className="font-medium">Tiến độ</span>
          <span className="text-muted-foreground" data-testid="progress-text">
            Đã xem {watched}/{EPISODES.length} · đã thuộc {learned}/{EPISODES.length}
          </span>
        </div>
        <Progress value={(watched / EPISODES.length) * 100} className="mt-2 h-1.5" data-testid="progress-bar" aria-label={`Đã xem ${watched} trên ${EPISODES.length} bài`} />
        <div className="mt-3 flex gap-1" role="group" aria-label="Nhóm danh sách theo">
          <Button size="sm" variant={grouping === "season" ? "secondary" : "ghost"} className="h-7 px-2.5 text-xs" aria-pressed={grouping === "season"} onClick={() => setGrouping("season")}>
            Theo mùa
          </Button>
          <Button size="sm" variant={grouping === "session" ? "secondary" : "ghost"} className="h-7 px-2.5 text-xs" aria-pressed={grouping === "session"} onClick={() => setGrouping("session")}>
            Theo buổi học
          </Button>
        </div>
      </div>
      <Separator />
      <ScrollArea className="min-h-0 flex-1">
        <nav aria-label="Danh sách bài" className="px-2 py-3">
          {grouping === "season"
            ? SEASON_INFO.map((season) => {
                const eps = EPISODES.filter((e) => e.season.index === season.index);
                const seen = eps.filter((e) => progress.watched[e.id]).length;
                return (
                  <section key={season.folder} className="mb-4" aria-label={season.label}>
                    <div className="flex items-baseline justify-between px-2 pb-1">
                      <h2 className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{season.label}</h2>
                      <span className="font-mono text-[11px] text-muted-foreground">
                        {seen}/{eps.length}
                      </span>
                    </div>
                    <p className="px-2 pb-1.5 text-xs text-muted-foreground">{season.blurb}</p>
                    <ul className="space-y-0.5">
                      {eps.map((e) => (
                        <Row key={e.id} e={e} />
                      ))}
                    </ul>
                  </section>
                );
              })
            : PATH.map((s) => {
                const eps = s.episodes.map((id) => EPISODES.find((e) => e.id === id)).filter((e): e is Episode => Boolean(e));
                return (
                  <section key={s.n} className="mb-4" aria-label={`Buổi ${s.n}`}>
                    <div className="px-2 pb-1">
                      <h2 className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Buổi {s.n}</h2>
                      <div className="text-sm font-medium">{s.title}</div>
                      <p className="text-xs text-muted-foreground">{s.goal}</p>
                    </div>
                    {eps.length ? (
                      <ul className="space-y-0.5">
                        {eps.map((e) => (
                          <Row key={e.id} e={e} />
                        ))}
                      </ul>
                    ) : null}
                    <ul className="mt-1 list-disc space-y-1 px-2 pl-6 text-xs text-muted-foreground">
                      {s.read.map((r) => (
                        <li key={r}>Đọc {r}</li>
                      ))}
                      {s.do.map((d) => (
                        <li key={d}>{d}</li>
                      ))}
                    </ul>
                  </section>
                );
              })}

          <Separator className="my-3" />
          <section aria-label="Không phải video" className="px-2">
            <h2 className="pb-1.5 text-xs font-semibold tracking-wide text-muted-foreground uppercase">Không phải video</h2>
            <ul className="space-y-0.5">
              {(Object.keys(PLATES) as (keyof typeof PLATES)[]).map((k) => (
                <li key={k}>
                  <a
                    href={`cheatsheet/${PLATE_FILES[k]}.html`}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-sm hover:bg-sidebar-accent"
                    data-testid="plate-link"
                  >
                    <FileText className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                    <span className="min-w-0 flex-1 truncate">Cheatsheet: {PLATES[k]}</span>
                    <ExternalLink className="size-3.5 shrink-0 text-muted-foreground" aria-label="mở tab mới" />
                  </a>
                </li>
              ))}
              <li>
                <a href={docURL(EXERCISES_DOC)} target="_blank" rel="noreferrer" className="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-sm hover:bg-sidebar-accent">
                  <BookOpen className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                  <span className="min-w-0 flex-1 truncate">{Object.keys(EXERCISES).length} bài tập tự chạy trong HOC.md</span>
                  <ExternalLink className="size-3.5 shrink-0 text-muted-foreground" aria-label="mở tab mới" />
                </a>
              </li>
            </ul>
          </section>
        </nav>
      </ScrollArea>
    </div>
  );
};
