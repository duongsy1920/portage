#!/usr/bin/env python3
"""Re-runnable version of the roadmap comparison PLAN-BO-SUNG §1 lost in /tmp.

For every roadmap item marked "taught" in src/data/roadmap.ts, grep its keywords
in everything that reaches the screen of the referenced episode: every string in
its scene files plus its recap.ts. A "taught" item whose words never appear on
that episode's screen is a lie, and is reported. Items with other statuses are
counted, not checked — there is nothing on screen to check them against.

Usage: python3 scripts/roadmap-gap.py        (from learn/; exit 1 if any lie)
"""
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from episodes import LEARN, all_episodes  # noqa: E402

road = (LEARN / "src/data/roadmap.ts").read_text()
ITEM = re.compile(r'^\s+T\("([^"]+)", "([^"]+)", "([^"]+)", \[([^\]]*)\]', re.M)
by_id = {e.id: e for e in all_episodes()}

def screen_text(ep) -> str:
    folder = ep.path.parent
    parts = [p.read_text() for p in (folder / "scenes").glob("*.tsx")]
    parts.append((folder / "recap.ts").read_text())
    parts.append(ep.path.read_text())  # scene names
    return "\n".join(parts)

counts = {k: len(re.findall(rf"^\s+{k}\(", road, re.M)) for k in "TPSXD"}
lies = 0
for group, name, ref, kw in ITEM.findall(road):
    keywords = re.findall(r'"([^"]+)"', kw)
    ep = by_id.get(ref)
    if not ep:
        print(f"  ✗ {name}: tập {ref} không có"); lies += 1; continue
    text = screen_text(ep)
    hits = [k for k in keywords if k in text]
    if not hits:
        print(f"  ✗ {name} → {ref}: không từ khoá nào {keywords} lên hình"); lies += 1
total = sum(counts.values())
print(f"{total} mục: đã dạy {counts['T']} · sẽ có {counts['P']} · bản in {counts['S']} · bài tập {counts['X']} · cố ý bỏ {counts['D']}")
if lies:
    print(f"{lies} mục 'đã dạy' mà không thấy từ khoá trên hình — sửa nhãn hoặc từ khoá.")
    sys.exit(1)
print("mọi mục 'đã dạy' đều có từ khoá trên hình của tập nó trỏ tới.")
