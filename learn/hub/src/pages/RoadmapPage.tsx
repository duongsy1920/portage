import React, { useMemo, useState } from "react";
import { ExternalLink } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ROADMAP, PLANNED, PLATES, PLATE_FILES, EXERCISES, EXERCISES_DOC, type RoadmapStatus } from "../../../src/data/roadmap";
import { docURL } from "../../../src/data/meta";
import { byId } from "../catalog";
import { href } from "../router";
import { cn } from "@/lib/utils";

const LABEL: Record<RoadmapStatus, string> = {
  taught: "có bài",
  planned: "sẽ có",
  plate: "cheatsheet",
  exercise: "bài tập",
  dropped: "cố ý bỏ",
};
const ORDER: RoadmapStatus[] = ["taught", "planned", "plate", "exercise", "dropped"];
const TONE: Record<RoadmapStatus, "default" | "secondary" | "outline"> = {
  taught: "default",
  planned: "outline",
  plate: "secondary",
  exercise: "secondary",
  dropped: "outline",
};

export const RoadmapPage: React.FC = () => {
  const [filter, setFilter] = useState<RoadmapStatus | "all">("all");
  const counts = useMemo(() => {
    const c: Partial<Record<RoadmapStatus, number>> = {};
    for (const r of ROADMAP) c[r.status] = (c[r.status] ?? 0) + 1;
    return c;
  }, []);
  const rows = filter === "all" ? ROADMAP : ROADMAP.filter((r) => r.status === filter);

  const where = (r: (typeof ROADMAP)[number]): React.ReactNode => {
    if (r.status === "taught" && r.ref) {
      const e = byId(r.ref);
      return e ? (
        <a className="underline underline-offset-4" href={href({ page: "episode", id: e.id })}>
          {e.season.short} {e.number} · {e.spec.title}
        </a>
      ) : (
        <span className="text-destructive">{r.ref}?</span>
      );
    }
    if (r.status === "planned" && r.ref) return `${r.ref} · ${PLANNED[r.ref] ?? "chưa dựng"}`;
    if (r.status === "plate" && r.ref) {
      const k = r.ref as keyof typeof PLATES;
      return (
        <a className="inline-flex items-center gap-1 underline underline-offset-4" href={`cheatsheet/${PLATE_FILES[k]}.html`} target="_blank" rel="noreferrer">
          cheatsheet “{PLATES[k]}” <ExternalLink className="size-3" aria-hidden="true" />
        </a>
      );
    }
    if (r.status === "exercise" && r.ref) {
      const k = r.ref as keyof typeof EXERCISES;
      return (
        <a className="inline-flex items-center gap-1 underline underline-offset-4" href={docURL(EXERCISES_DOC)} target="_blank" rel="noreferrer">
          bài tập “{EXERCISES[k]}” trong {EXERCISES_DOC} <ExternalLink className="size-3" aria-hidden="true" />
        </a>
      );
    }
    return null;
  };

  return (
    <div>
      <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">So với roadmap.sh</h1>
      <p className="mt-1 max-w-3xl text-sm text-muted-foreground">
        {ROADMAP.length} mục của roadmap Go trên roadmap.sh (bản PDF 06/09/2025). Mỗi mục đúng một nhãn: có bài nào dạy, tra ở cheatsheet nào, phải tự chạy, hay cố
        ý bỏ và vì sao. Lý do của từng nhóm ở <code className="font-mono text-xs">learn/PLAN-BO-SUNG.md</code> §4–§8.
      </p>

      <div className="mt-4 flex flex-wrap gap-1.5" role="group" aria-label="Lọc theo nhãn">
        <Button size="sm" variant={filter === "all" ? "secondary" : "ghost"} aria-pressed={filter === "all"} onClick={() => setFilter("all")} data-testid="roadmap-filter-all">
          tất cả <span className="font-mono text-xs text-muted-foreground">{ROADMAP.length}</span>
        </Button>
        {ORDER.filter((s) => (counts[s] ?? 0) > 0).map((s) => (
          <Button key={s} size="sm" variant={filter === s ? "secondary" : "ghost"} aria-pressed={filter === s} onClick={() => setFilter(s)} data-testid={`roadmap-filter-${s}`}>
            {LABEL[s]} <span className="font-mono text-xs text-muted-foreground">{counts[s]}</span>
          </Button>
        ))}
      </div>

      <div className="mt-4 overflow-x-auto rounded-lg border">
        <table className="w-full text-sm" data-testid="roadmap-table">
          <thead className="bg-muted/50 text-left text-xs text-muted-foreground">
            <tr>
              <th className="px-3 py-2 font-medium">Nhóm</th>
              <th className="px-3 py-2 font-medium">Mục</th>
              <th className="px-3 py-2 font-medium">Nhãn</th>
              <th className="px-3 py-2 font-medium">Ở đâu, vì sao</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {rows.map((r) => (
              <tr key={`${r.group}/${r.name}`} className="align-top" data-testid="roadmap-row">
                <td className="px-3 py-2 text-xs whitespace-nowrap text-muted-foreground">{r.group}</td>
                <td className="px-3 py-2 font-medium">{r.name}</td>
                <td className="px-3 py-2">
                  <Badge variant={TONE[r.status]} className={cn(r.status === "dropped" && "text-muted-foreground", r.status === "planned" && "border-dashed")}>
                    {LABEL[r.status]}
                  </Badge>
                </td>
                <td className="px-3 py-2 text-muted-foreground">
                  {where(r)}
                  {r.note ? (
                    <>
                      {r.status === "dropped" ? "" : " — "}
                      {r.note}
                    </>
                  ) : null}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
