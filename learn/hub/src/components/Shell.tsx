import React, { useMemo, useRef, useState } from "react";
import { Download, Menu, MoreHorizontal, Moon, Sun, Trash2, Upload } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Sidebar } from "./Sidebar";
import { href, type Route } from "../router";
import { useProgress } from "../store";
import { useTheme } from "../theme";
import { buildCards, dueCards } from "../review";
import { exportJSON, importJSON, empty } from "../progress";
import { VERIFIED } from "../../../src/data/meta";

/**
 * The frame every page sits in: header, the course outline on the left (a
 * drawer under 1024 px), the page on the right. Progress import/export and the
 * theme live in the header so no page has to carry them.
 */
export const Shell: React.FC<{ route: Route; currentId?: string; children: React.ReactNode }> = ({ route, currentId, children }) => {
  const { progress, update } = useProgress();
  const { isDark, set } = useTheme();
  const [open, setOpen] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  const due = useMemo(() => dueCards(buildCards(progress), progress).length, [progress]);

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
      alert(`Không nạp được: ${r.why}`);
      return;
    }
    update(() => r.progress);
  };

  return (
    <div className="flex min-h-svh flex-col">
      <header className="sticky top-0 z-30 flex h-14 items-center gap-2 border-b bg-background/95 px-3 backdrop-blur supports-[backdrop-filter]:bg-background/80 sm:px-4">
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="lg:hidden" aria-label="Mở danh sách bài" data-testid="open-sidebar">
              <Menu />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[320px] max-w-[88vw] gap-0 p-0">
            <SheetHeader className="sr-only">
              <SheetTitle>Danh sách bài</SheetTitle>
              <SheetDescription>Ba mùa, ba mươi bài, và tiến độ của bạn.</SheetDescription>
            </SheetHeader>
            <Sidebar currentId={currentId} onNavigate={() => setOpen(false)} inSheet />
          </SheetContent>
        </Sheet>

        <a href={href({ page: "home" })} className="flex items-baseline gap-2 rounded-sm font-semibold tracking-tight">
          Học Go
          <span className="hidden text-sm font-normal text-muted-foreground sm:inline">từ Symfony sang Go, trên project Portage</span>
        </a>

        <nav className="ml-auto flex items-center gap-1" aria-label="Trang">
          <Button asChild variant={route.page === "review" ? "secondary" : "ghost"} size="sm">
            <a href={href({ page: "review" })} aria-current={route.page === "review" ? "page" : undefined}>
              Ôn tập
              {due > 0 ? (
                <Badge variant="secondary" className="ml-0.5 h-5 min-w-5 justify-center px-1 font-mono text-[11px]" data-testid="due-badge">
                  {due}
                </Badge>
              ) : null}
            </a>
          </Button>
          <Button asChild variant={route.page === "roadmap" ? "secondary" : "ghost"} size="sm" className="hidden sm:inline-flex">
            <a href={href({ page: "roadmap" })} aria-current={route.page === "roadmap" ? "page" : undefined}>
              So với roadmap.sh
            </a>
          </Button>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon" aria-label={isDark ? "Chuyển sang nền sáng" : "Chuyển sang nền tối"} onClick={() => set(isDark ? "light" : "dark")} data-testid="theme-toggle">
                {isDark ? <Sun /> : <Moon />}
              </Button>
            </TooltipTrigger>
            <TooltipContent>{isDark ? "Nền sáng" : "Nền tối"}</TooltipContent>
          </Tooltip>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" aria-label="Thêm" data-testid="more-menu">
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-64">
              <DropdownMenuLabel className="font-normal text-muted-foreground">Tiến độ lưu trong trình duyệt này. Không có tài khoản, không có server.</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onSelect={download}>
                <Download /> Tải tiến độ về (file JSON)
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={() => fileRef.current?.click()}>
                <Upload /> Nạp tiến độ từ file
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild className="sm:hidden">
                <a href={href({ page: "roadmap" })}>So với roadmap.sh</a>
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={() => set("system")}>Nền sáng/tối theo hệ điều hành</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                onSelect={() => {
                  if (confirm("Xoá toàn bộ tiến độ trên trình duyệt này?")) update(() => empty());
                }}
              >
                <Trash2 /> Xoá tiến độ trên máy này
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <input ref={fileRef} type="file" accept="application/json" hidden onChange={(ev) => void onImport(ev.target.files?.[0])} />
        </nav>
      </header>

      <div className="flex flex-1">
        <aside className="sticky top-14 hidden h-[calc(100svh-3.5rem)] w-[300px] shrink-0 border-r bg-sidebar text-sidebar-foreground lg:block" aria-label="Danh sách bài">
          <Sidebar currentId={currentId} />
        </aside>
        <main className="min-w-0 flex-1 px-4 py-5 sm:px-6 lg:px-8">
          <div className="mx-auto w-full max-w-5xl">{children}</div>
        </main>
      </div>

      <footer className="border-t px-4 py-4 text-xs text-muted-foreground lg:pl-[calc(300px+2rem)]">
        <div className="mx-auto flex max-w-5xl flex-wrap gap-x-4 gap-y-1">
          <span>Video và cheatsheet dựng bằng Remotion từ chính code của Portage.</span>
          <a className="underline-offset-4 hover:underline" href={`${VERIFIED.repo}/tree/${VERIFIED.commit}/learn`} target="_blank" rel="noreferrer">
            Mã nguồn trên GitHub
          </a>
          <span>Số liệu và snippet kiểm lần cuối: {VERIFIED.date}.</span>
        </div>
      </footer>
    </div>
  );
};
