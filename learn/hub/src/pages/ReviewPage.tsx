import React, { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card";
import { EPISODES } from "../catalog";
import { href } from "../router";
import { useProgress } from "../store";
import { dueAfter, INTERVAL_DAYS, type Grade } from "../progress";
import { buildCards, dueCards, SITTING, type ReviewCard } from "../review";
import { HOC_QUIZ } from "../../../src/data/quiz";

const GRADES: { g: Grade; label: string }[] = [
  { g: "chua", label: "chưa" },
  { g: "mo-ho", label: "mơ hồ" },
  { g: "nho", label: "nhớ" },
];

export const ReviewPage: React.FC = () => {
  const { progress, update } = useProgress();
  const [showHints, setShowHints] = useState(false);
  const [done, setDone] = useState<string[]>([]);

  const cards = useMemo(() => buildCards(progress), [progress]);
  const due = dueCards(cards, progress).filter((c) => !done.includes(c.id));
  const sitting = due.slice(0, SITTING);
  const card = sitting[0];
  const total = cards.length;
  const seen = cards.filter((c) => progress.review[c.id]).length;
  const nothingWatched = EPISODES.every((e) => !progress.watched[e.id]);

  const grade = (c: ReviewCard, g: Grade) => {
    update((s) => ({ ...s, review: { ...s.review, [c.id]: { due: dueAfter(g), grade: g, at: new Date().toISOString() } } }));
    setDone((d) => [...d, c.id]);
    setShowHints(false);
  };

  return (
    <div>
      <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">Ôn tập</h1>
      <p className="mt-1 text-sm text-muted-foreground">Trả lời thành tiếng, không nhìn code, rồi tự chấm.</p>
      <p className="mt-3 text-sm text-muted-foreground" data-testid="review-summary">
        {total} câu hỏi: câu phỏng vấn của các bài đã xem, cộng {HOC_QUIZ.length} câu của HOC.md. Đã chấm {seen}/{total}. Hôm nay đến hạn: {due.length}
        {due.length > SITTING ? ` (một lượt tối đa ${SITTING})` : ""}.
      </p>

      {nothingWatched ? (
        <p className="mt-2 text-sm text-muted-foreground">
          Chưa xem bài nào, nên mới chỉ có {HOC_QUIZ.length} câu của HOC.md.{" "}
          <a className="underline underline-offset-4" href={href({ page: "home" })}>
            Xem bài đầu tiên
          </a>{" "}
          thì có thêm câu hỏi.
        </p>
      ) : null}

      {card ? (
        <Card className="mt-5" data-testid="review-card" key={card.id}>
          <CardHeader>
            <p className="text-xs text-muted-foreground">{card.from}</p>
            <p className="text-lg leading-snug font-medium">{card.q}</p>
          </CardHeader>
          <CardContent>
            {card.hints.length ? (
              <Button variant="outline" size="sm" onClick={() => setShowHints((v) => !v)} aria-expanded={showHints}>
                {showHints ? "Ẩn ý chính" : "Xem ý chính của bài"}
              </Button>
            ) : card.episode ? (
              <Button asChild variant="outline" size="sm">
                <a href={href({ page: "episode", id: card.episode })}>Mở lại bài liên quan</a>
              </Button>
            ) : (
              <p className="text-xs text-muted-foreground">Câu này chỉ có trong docs, không có video đi kèm.</p>
            )}
            {showHints && card.hints.length ? (
              <ol className="mt-3 list-decimal space-y-1.5 rounded-lg border bg-muted/40 px-4 py-3 pl-9 text-sm">
                {card.hints.map((h) => (
                  <li key={h}>{h}</li>
                ))}
              </ol>
            ) : null}
          </CardContent>
          <CardFooter className="flex flex-wrap items-center gap-2">
            {GRADES.map(({ g, label }) => (
              <Button key={g} variant="outline" size="sm" onClick={() => grade(card, g)} data-testid={`review-grade-${g}`}>
                {label}
                <span className="font-mono text-[11px] font-normal text-muted-foreground">{INTERVAL_DAYS[g]} ngày</span>
              </Button>
            ))}
            <span className="ml-auto text-xs text-muted-foreground">còn {sitting.length - 1} câu trong lượt này</span>
          </CardFooter>
        </Card>
      ) : (
        <Card className="mt-5" data-testid="review-card">
          <CardHeader>
            <p className="text-lg font-medium">Hết câu đến hạn hôm nay</p>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            {done.length ? `Vừa chấm ${done.length} câu. ` : ""}
            Câu “chưa” quay lại sau {INTERVAL_DAYS.chua} ngày, “mơ hồ” sau {INTERVAL_DAYS["mo-ho"]} ngày, “nhớ” sau {INTERVAL_DAYS.nho} ngày. Xem thêm bài mới thì có thêm câu hỏi.
          </CardContent>
        </Card>
      )}
      <p className="mt-3 text-xs text-muted-foreground">Số ngày ghi cạnh mỗi nút là lúc câu đó quay lại.</p>
    </div>
  );
};
