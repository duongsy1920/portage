import React, { useEffect, useMemo, useRef, useState } from "react";
import { Player, type PlayerRef } from "@remotion/player";
import { ChevronLeft, ChevronRight, ExternalLink } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FPS, HEIGHT, WIDTH } from "../../../src/design/tokens";
import { EPISODES, byId, componentOf, isConfident, mmss, type Episode } from "../catalog";
import { resolveAnchor } from "../anchors";
import { VERIFIED } from "../../../src/data/meta";
import { href } from "../router";
import { useProgress } from "../store";
import { dueAfter, INTERVAL_DAYS, type Grade } from "../progress";
import { cn } from "@/lib/utils";

const GRADES: { g: Grade; label: string }[] = [
  { g: "nho", label: "nhớ" },
  { g: "mo-ho", label: "mơ hồ" },
  { g: "chua", label: "chưa" },
];

export const EpisodePage: React.FC<{ id: string; continuing?: boolean }> = ({ id, continuing }) => {
  const e = byId(id);
  if (!e) {
    return (
      <div className="py-10">
        <h1 className="text-2xl font-semibold tracking-tight">Không có bài nào tên {id}</h1>
        <p className="mt-2 text-muted-foreground">
          Chọn một bài trong danh sách bên trái, hoặc{" "}
          <a className="underline underline-offset-4" href={href({ page: "home" })}>
            mở bài kế tiếp
          </a>
          .
        </p>
      </div>
    );
  }
  return <EpisodeView key={e.id} e={e} continuing={continuing} />;
};

const EpisodeView: React.FC<{ e: Episode; continuing?: boolean }> = ({ e, continuing }) => {
  const { progress, update } = useProgress();
  const ref = useRef<PlayerRef>(null);
  const [frame, setFrame] = useState(0);
  const [showPoints, setShowPoints] = useState(false);
  const Episode = useMemo(() => componentOf(e), [e]);

  const idx = EPISODES.findIndex((x) => x.id === e.id);
  const prev = EPISODES[idx - 1];
  const next = EPISODES[idx + 1];
  const seen = Boolean(progress.watched[e.id]);
  const learned = isConfident(e, progress.quiz);
  const anyWatched = Object.keys(progress.watched).length > 0;

  // Watched = reached the last frame. No button to press: finishing is the proof.
  useEffect(() => {
    const p = ref.current;
    if (!p) return;
    const mark = () => update((s) => (s.watched[e.id] ? s : { ...s, watched: { ...s.watched, [e.id]: new Date().toISOString() } }));
    const onFrame = (ev: { detail: { frame: number } }) => {
      setFrame(ev.detail.frame);
      if (ev.detail.frame >= e.frames - 2) mark();
    };
    p.addEventListener("frameupdate", onFrame);
    p.addEventListener("ended", mark);
    return () => {
      p.removeEventListener("frameupdate", onFrame);
      p.removeEventListener("ended", mark);
    };
  }, [e, update]);

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

  const seekTo = (i: number) => {
    ref.current?.seekTo(e.sceneStarts[i] ?? 0);
    ref.current?.play();
  };

  const grade = (i: number, g: Grade) =>
    update((s) => ({
      ...s,
      quiz: { ...s.quiz, [e.id]: { ...(s.quiz[e.id] ?? {}), [String(i)]: { grade: g, at: new Date().toISOString() } } },
      review: { ...s.review, [`${e.id}#${i}`]: { due: dueAfter(g), grade: g, at: new Date().toISOString() } },
    }));
  const graded = progress.quiz[e.id] ?? {};

  const PrevNext: React.FC<{ className?: string }> = ({ className }) => (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {prev ? (
        <Button asChild variant="outline" size="sm">
          <a href={href({ page: "episode", id: prev.id })}>
            <ChevronLeft /> {prev.season.short} {prev.number}
          </a>
        </Button>
      ) : null}
      {next ? (
        <Button asChild variant="outline" size="sm">
          <a href={href({ page: "episode", id: next.id })}>
            {next.season.short} {next.number} <ChevronRight />
          </a>
        </Button>
      ) : null}
    </div>
  );

  return (
    <article>
      <header className="flex flex-wrap items-end justify-between gap-x-6 gap-y-3" data-testid="episode-header">
        <div className="min-w-0">
          {continuing ? <p className="mb-1 text-xs font-medium text-muted-foreground">{seen ? "Xem lại để tự kiểm tra" : anyWatched ? "Tiếp theo" : "Bắt đầu ở đây"}</p> : null}
          <p className="text-sm text-muted-foreground">
            {e.season.label} · Tập {e.number}
          </p>
          <h1 className="mt-0.5 text-2xl font-semibold tracking-tight sm:text-3xl">{e.spec.title}</h1>
          <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <span className="font-mono text-xs">{mmss(e.seconds)}</span>
            <span aria-hidden="true">·</span>
            <span>{e.spec.scenes.length} đoạn</span>
            {seen ? (
              <Badge variant="secondary" data-testid="watched-badge">
                Đã xem
              </Badge>
            ) : null}
            {learned ? (
              <Badge variant="outline" className="border-done text-done" data-testid="confident-badge">
                Đã thuộc
              </Badge>
            ) : null}
          </div>
        </div>
        <PrevNext />
      </header>

      <div className="mt-4 overflow-hidden rounded-lg border bg-black shadow-sm" data-testid="player">
        <Player ref={ref} component={Episode} durationInFrames={e.frames} fps={FPS} compositionWidth={WIDTH} compositionHeight={HEIGHT} controls clickToPlay style={{ width: "100%" }} />
      </div>

      <Tabs defaultValue="toc" className="mt-5">
        <TabsList className="grid h-auto w-full grid-cols-2 sm:inline-flex sm:w-auto">
          <TabsTrigger value="toc" data-testid="tab-toc">
            Mục lục
          </TabsTrigger>
          <TabsTrigger value="cheatsheet" data-testid="tab-cheatsheet">
            Cheatsheet
          </TabsTrigger>
          <TabsTrigger value="code" data-testid="tab-code">
            Code trong repo
          </TabsTrigger>
          <TabsTrigger value="quiz" data-testid="tab-quiz">
            Tự kiểm tra
          </TabsTrigger>
        </TabsList>

        <TabsContent value="toc" className="mt-3">
          <ol className="divide-y rounded-lg border" aria-label="Các đoạn của video">
            {e.spec.scenes.map((s, i) => (
              <li key={s.name}>
                <button
                  type="button"
                  onClick={() => seekTo(i)}
                  aria-current={i === sceneNow ? "step" : undefined}
                  data-testid="toc-item"
                  className={cn(
                    "flex w-full items-baseline gap-3 px-3 py-2 text-left text-sm hover:bg-accent focus-visible:bg-accent focus-visible:outline-none",
                    i === sceneNow && "bg-accent font-medium",
                  )}
                >
                  <span className="min-w-0 flex-1">{s.name}</span>
                  <span className="shrink-0 font-mono text-xs text-muted-foreground">{mmss(Math.round((e.sceneStarts[i] ?? 0) / FPS))}</span>
                </button>
              </li>
            ))}
          </ol>
          <p className="mt-2 text-xs text-muted-foreground">Bấm một đoạn để tua tới đó. Xem hết đoạn cuối thì bài tự được đánh dấu “đã xem”.</p>
        </TabsContent>

        <TabsContent value="cheatsheet" className="mt-3">
          <PlateView e={e} />
        </TabsContent>

        <TabsContent value="code" className="mt-3 space-y-3">
          <p className="text-sm text-muted-foreground">
            Mọi dòng code và con số trong video đều được script kiểm lại trên commit{" "}
            <a className="font-mono underline underline-offset-4" href={`${VERIFIED.repo}/tree/${VERIFIED.commit}`} target="_blank" rel="noreferrer">
              {VERIFIED.short}
            </a>{" "}
            ({VERIFIED.date}). Link mở đúng commit đó chứ không phải <code className="font-mono">main</code>, vì số dòng sẽ trôi theo thời gian.
          </p>
          {e.spec.anchors.length === 0 ? <p className="text-sm text-muted-foreground">Bài này không chiếu snippet nào từ repo.</p> : null}
          <div className="flex flex-wrap gap-x-6 gap-y-2">
            {e.spec.anchors.map((name) => {
              const a = resolveAnchor(name);
              return a.kind === "figure" ? (
                <div key={name} className="flex items-baseline gap-2" data-testid="figure">
                  <span className="font-mono text-2xl font-semibold">{a.value}</span>
                  <span className="text-xs text-muted-foreground">{name.replace("GO_USAGE.", "")}</span>
                </div>
              ) : null;
            })}
          </div>
          {e.spec.anchors.map((name) => {
            const a = resolveAnchor(name);
            if (a.kind === "snippet")
              return (
                <Card key={name} className="gap-2 py-4" data-testid="anchor">
                  <CardHeader className="px-4">
                    <CardTitle className="flex flex-wrap items-baseline gap-x-3 gap-y-1 font-mono text-xs font-normal text-muted-foreground">
                      <span className="text-foreground">{a.snippet.path}</span>
                      {a.snippet.line ? <span>dòng {a.snippet.line}</span> : null}
                      {a.url ? (
                        <a className="inline-flex items-center gap-1 underline underline-offset-4" href={a.url} target="_blank" rel="noreferrer">
                          Mở trên GitHub <ExternalLink className="size-3" aria-hidden="true" />
                        </a>
                      ) : (
                        <span>minh hoạ, không phải code trong repo</span>
                      )}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="px-4">
                    <pre className="overflow-x-auto rounded-md bg-muted p-3 font-mono text-[13px] leading-relaxed">
                      <code>{a.snippet.code.join("\n")}</code>
                    </pre>
                  </CardContent>
                </Card>
              );
            if (a.kind === "unknown")
              return (
                <Card key={name} className="border-destructive py-3" data-testid="anchor">
                  <CardContent className="px-4 font-mono text-xs text-destructive">{name} — không có trong data/portage.ts</CardContent>
                </Card>
              );
            return null;
          })}
        </TabsContent>

        <TabsContent value="quiz" className="mt-3 space-y-3">
          <div className="flex flex-wrap items-center gap-3">
            <Button variant="outline" size="sm" onClick={() => setShowPoints((v) => !v)} aria-expanded={showPoints}>
              {showPoints ? "Ẩn ý chính" : `Xem ${e.spec.recap.points.length} ý chính của bài`}
            </Button>
            <span className="text-xs text-muted-foreground">Thử trả lời trước, rồi mới xem.</span>
          </div>
          {showPoints ? (
            <ol className="list-decimal space-y-1.5 rounded-lg border bg-muted/40 px-4 py-3 pl-9 text-sm">
              {e.spec.recap.points.map((p) => (
                <li key={p.at}>{p.text}</li>
              ))}
            </ol>
          ) : null}
          <Card className="gap-3 py-4">
            <CardHeader className="px-4">
              <CardTitle className="text-base">Phỏng vấn sẽ hỏi</CardTitle>
            </CardHeader>
            <CardContent className="divide-y px-4">
              {e.spec.recap.asked.map((q, i) => (
                <div key={q} className="flex flex-wrap items-start justify-between gap-x-6 gap-y-2 py-3 first:pt-0 last:pb-0">
                  <p className="min-w-0 flex-1 basis-72 text-sm">{q}</p>
                  <div className="flex gap-1" role="group" aria-label={`Tự chấm câu ${i + 1}`}>
                    {GRADES.map(({ g, label }) => {
                      const on = graded[String(i)]?.grade === g;
                      return (
                        <Button
                          key={g}
                          size="sm"
                          variant={on ? "default" : "outline"}
                          className={cn("h-7 px-2.5 text-xs", on && g === "nho" && "bg-done text-white hover:bg-done/90")}
                          aria-pressed={on}
                          data-testid={`grade-${g}`}
                          onClick={() => grade(i, g)}
                        >
                          {label}
                        </Button>
                      );
                    })}
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
          <p className="text-xs text-muted-foreground">
            Chấm “nhớ” cả {e.spec.recap.asked.length} câu thì bài này tính là đã thuộc. Câu “chưa” sẽ quay lại trong Ôn tập sau {INTERVAL_DAYS.chua} ngày, “mơ hồ” sau {INTERVAL_DAYS["mo-ho"]} ngày, “nhớ” sau {INTERVAL_DAYS.nho} ngày.
          </p>
        </TabsContent>
      </Tabs>

      <PrevNext className="mt-6 justify-end" />
    </article>
  );
};

/** The A3 cheatsheet, scaled to the available width. It is the same HTML the PDF is printed from. */
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
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-3">
        <Button asChild variant="outline" size="sm">
          <a href={src} target="_blank" rel="noreferrer">
            Mở toàn màn hình <ExternalLink />
          </a>
        </Button>
        <span className="text-xs text-muted-foreground">Một trang A3 ngang. Muốn in thì mở toàn màn hình rồi Ctrl+P.</span>
      </div>
      <div className="plate-frame rounded-lg border" ref={box} style={{ height: Math.round(1123 * scale) }}>
        {/* eslint-disable-next-line @remotion/warn-native-media-tag -- a page, not a composition: nothing renders this frame to video */}
        <iframe title={`Cheatsheet ${e.label}`} src={src} style={{ transform: `scale(${scale})` }} loading="lazy" />
      </div>
    </div>
  );
};
