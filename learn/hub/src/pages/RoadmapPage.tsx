import React, { useMemo, useState } from "react";
import { ROADMAP, PLANNED, PLATES, EXERCISES, type RoadmapStatus } from "../../../src/data/roadmap";
import { byId } from "../catalog";
import { href } from "../router";

const LABEL: Record<RoadmapStatus, string> = {
  taught: "đã dạy",
  planned: "sẽ có",
  plate: "bản in",
  exercise: "bài tập",
  dropped: "cố ý bỏ",
};
const ORDER: RoadmapStatus[] = ["taught", "planned", "plate", "exercise", "dropped"];

export const RoadmapPage: React.FC = () => {
  const [filter, setFilter] = useState<RoadmapStatus | "all">("all");
  const counts = useMemo(() => {
    const c: Record<string, number> = {};
    for (const r of ROADMAP) c[r.status] = (c[r.status] ?? 0) + 1;
    return c;
  }, []);
  const rows = filter === "all" ? ROADMAP : ROADMAP.filter((r) => r.status === filter);

  const refOf = (r: (typeof ROADMAP)[number]): React.ReactNode => {
    if (r.status === "taught" && r.ref) {
      const e = byId(r.ref);
      return e ? (
        <a href={href({ page: "episode", id: e.id })}>
          {e.season.short} {e.number} · {e.spec.title}
        </a>
      ) : (
        <span className="bad">{r.ref}?</span>
      );
    }
    if (r.status === "planned" && r.ref) return `${r.ref} · ${PLANNED[r.ref as keyof typeof PLANNED]}`;
    if (r.status === "plate" && r.ref) return `bản in “${PLATES[r.ref as keyof typeof PLATES]}”`;
    if (r.status === "exercise" && r.ref) return `bài tập “${EXERCISES[r.ref as keyof typeof EXERCISES]}”`;
    return null;
  };

  return (
    <>
      <h1>
        Đối chiếu <em>roadmap.sh</em>
      </h1>
      <p className="sub">
        {ROADMAP.length} mục của roadmap Go (bản PDF 06/09/2025), mỗi mục đúng một nhãn: dạy ở tập nào, sẽ có ở tập nào, tra ở bản in nào, phải
        tự chạy, hay cố ý bỏ và vì sao. Lý do của từng nhóm ở <code>learn/PLAN-BO-SUNG.md</code> §4–§8.
      </p>
      <div className="filters" role="group" aria-label="Lọc theo nhãn">
        <button aria-pressed={filter === "all"} onClick={() => setFilter("all")}>
          tất cả {ROADMAP.length}
        </button>
        {ORDER.map((s) => (
          <button key={s} aria-pressed={filter === s} onClick={() => setFilter(s)}>
            <span className={`tag ${s}`}>{LABEL[s]}</span> {counts[s] ?? 0}
          </button>
        ))}
      </div>
      <table className="road">
        <thead>
          <tr>
            <th className="g">Nhóm</th>
            <th>Mục</th>
            <th>Nhãn</th>
            <th>Ở đâu / vì sao</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr key={`${r.group}/${r.name}`}>
              <td className="g">{r.group}</td>
              <td>{r.name}</td>
              <td>
                <span className={`tag ${r.status}`}>{LABEL[r.status]}</span>
              </td>
              <td className="note">
                {refOf(r)}
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
    </>
  );
};
