#!/usr/bin/env python3
"""The fifth gate: the learning hub's data cannot drift from the episodes it describes.

Checks, each born from a way this could go wrong without anyone noticing:
  1. every spec.ts has recap.ts beside it with ≥3 points and exactly 3 interview questions
  2. spec.anchors equals what the scene files actually reference in data/portage.ts
     (the list is generated once; if a scene changes, this is what catches the stale list)
  3. every anchor names something that exists in data/portage.ts
  4. every roadmap item has one status, and a "taught" item points at a real episode
  5. every episode id in path.ts and quiz.ts is real
  6. the 17 quiz questions are the ones in docs/HOC.md, verbatim
  7. the verified commit in data/meta.ts exists in the Go repository
  8. CURRICULUM.md §11 (scene count, length) agrees with the specs

Usage: python3 scripts/verify-hub.py      (from learn/; exit 1 on any failure)
"""
import re
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from episodes import LEARN, all_episodes  # noqa: E402

REPO = LEARN.parent
DATA = LEARN / "src/data"
bad = 0


def fail(msg: str) -> None:
    global bad
    bad += 1
    print(f"  ✗ {msg}")


def ok(msg: str) -> None:
    print(f"  ✓ {msg}")


ANCHOR = re.compile(r"\b(SNIPPETS|GO_SNIPPETS|DDD_SNIPPETS|PHP_SNIPPETS|GO_USAGE)\.(\w+)")

# what data/portage.ts actually exports, by set
portage = (DATA / "portage.ts").read_text()
sets: dict[str, set[str]] = {}
for m in re.finditer(r"^export const (SNIPPETS|GO_SNIPPETS|DDD_SNIPPETS|PHP_SNIPPETS|GO_USAGE) = \{(.*?)^\} as const", portage, re.S | re.M):
    sets[m.group(1)] = set(re.findall(r"^  (\w+): ", m.group(2), re.M))

episodes = all_episodes()
ids = {e.id for e in episodes}

print("1–3. recap.ts và anchors của từng tập")
for ep in episodes:
    folder = ep.path.parent
    recap = folder / "recap.ts"
    if not recap.exists():
        fail(f"{ep.id}: thiếu recap.ts"); continue
    rt = recap.read_text()
    points = len(re.findall(r"\{ at: \d+, text:", rt))
    asked_block = re.search(r"asked: \[(.*?)\n  \]", rt, re.S)
    asked = len(re.findall(r'^\s+"', asked_block.group(1), re.M)) if asked_block else 0
    if points < 3 or asked != 3:
        fail(f"{ep.id}: recap có {points} ý, {asked} câu hỏi (cần ≥3 và đúng 3)")
    spec_text = ep.path.read_text()
    if "recap: RECAP" not in spec_text:
        fail(f"{ep.id}: spec.ts không có recap: RECAP")
    m = re.search(r"anchors: \[(.*?)\]", spec_text, re.S)
    declared = set(re.findall(r'"([^"]+)"', m.group(1))) if m else set()
    found: set[str] = set()
    for scene in (folder / "scenes").glob("*.tsx"):
        found |= {f"{a}.{b}" for a, b in ANCHOR.findall(scene.read_text())}
    found |= {f"GO_USAGE.{b}" for a, b in ANCHOR.findall(rt) if a == "GO_USAGE"}
    if declared != found:
        fail(f"{ep.id}: anchors khai {sorted(declared)} ≠ thực tế {sorted(found)}")
    for name in declared:
        s, k = name.split(".", 1)
        if k not in sets.get(s, set()):
            fail(f"{ep.id}: anchor {name} không có trong data/portage.ts")
ok(f"{len(episodes)} tập có recap.ts; anchors khớp scene; mọi anchor có thật")

print("4. roadmap.ts")
road = (DATA / "roadmap.ts").read_text()
items = re.findall(r'^\s+([TPSXD])\("([^"]+)", "([^"]+)"(?:, "([^"]+)")?', road, re.M)
taught_refs = [ref for kind, _, _, ref in items if kind == "T"]
for ref in taught_refs:
    if ref not in ids:
        fail(f"roadmap: 'taught' trỏ tới tập không có: {ref}")
dups = {n for n in [f"{g}/{nm}" for _, g, nm, _ in items] if [f"{g}/{nm}" for _, g, nm, _ in items].count(n) > 1}
if dups:
    fail(f"roadmap: mục trùng: {dups}")
ok(f"{len(items)} mục, {len(taught_refs)} 'đã dạy' đều trỏ tới tập có thật")

print("5. path.ts và quiz.ts trỏ tới tập có thật")
for name in ("path.ts", "quiz.ts"):
    text = (DATA / name).read_text()
    for ref in re.findall(r'"((?:Go|Ddd)-Ep\d\d-[A-Za-z]+|Ep\d\d-[A-Za-z]+)"', text):
        if ref not in ids:
            fail(f"{name}: tập không có: {ref}")
ok("mọi id trong path.ts / quiz.ts có thật")

print("6. 17 câu của quiz.ts đúng chữ HOC.md")
hoc = (REPO / "docs/HOC.md").read_text()
sec = hoc.split("## Tự kiểm tra — trả lời không nhìn code", 1)[1].split("## Nói gì khi phỏng vấn", 1)[0]
doc_qs = [" ".join(q.split()) for q in re.findall(r"^\d+\. (.+?)(?=^\d+\. |\n\n|\Z)", sec, re.M | re.S)]
quiz = (DATA / "quiz.ts").read_text()
ts_qs = [bytes(q, "utf-8").decode("unicode_escape").encode("latin1").decode("utf-8") if "\\u" in q else q.replace('\\"', '"') for q in re.findall(r'q: "((?:[^"\\]|\\.)*)"', quiz)]
if len(doc_qs) != 17 or len(ts_qs) != 17:
    fail(f"đếm được {len(doc_qs)} câu trong HOC.md, {len(ts_qs)} trong quiz.ts")
for i, (d, t) in enumerate(zip(doc_qs, ts_qs), 1):
    if " ".join(d.split()) != " ".join(t.split()):
        fail(f"câu {i} lệch:\n      HOC: {d}\n      ts : {t}")
ok("17/17 câu giống HOC.md")

print("8. bảng CURRICULUM.md §11 (cảnh, thời lượng) khớp spec")
cur = (LEARN / "CURRICULUM.md").read_text()
rows = re.findall(r"^\| MÙA (\d) · (\d+) \| (.+?) \| (\d+) \| (\d+):(\d+) \|", cur, re.M)
season_of = {"go": 1, "ddd": 2, "portage": 3}
for ep in episodes:
    n = len(ep.durations)
    secs = (sum(ep.durations) - (n - 1) * 14) // 30  # BEAT.transition = 14, FPS = 30, floor like a player
    season = season_of[ep.path.parts[-3]]
    num = int(re.search(r"ep(\d+)", ep.path.parts[-2]).group(1))
    row = next((r for r in rows if int(r[0]) == season and int(r[1]) == num), None)
    if not row:
        fail(f"CURRICULUM §11 thiếu dòng MÙA {season} · {num}"); continue
    if int(row[3]) != n or int(row[4]) * 60 + int(row[5]) != secs:
        fail(f"CURRICULUM §11 MÙA {season} · {num}: bảng {row[3]} cảnh {row[4]}:{row[5]}, spec {n} cảnh {secs // 60}:{secs % 60:02d}")
ok(f"{len(rows)} dòng của bảng khớp số cảnh và thời lượng")

print("7. commit đã kiểm (meta.ts) tồn tại trong repo Go")
meta = (DATA / "meta.ts").read_text()
commit = re.search(r'commit: "([0-9a-f]{40})"', meta).group(1)
r = subprocess.run(["git", "-C", str(REPO), "cat-file", "-e", f"{commit}^{{commit}}"], capture_output=True)
if r.returncode != 0:
    fail(f"commit {commit[:7]} không có trong repo")
else:
    ok(f"commit {commit[:7]} có thật")

print()
if bad:
    print(f"{bad} lỗi — hub đang nói sai về video.")
    sys.exit(1)
print("verify-hub: khớp hết.")
