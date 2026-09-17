#!/usr/bin/env python3
"""Check every Go snippet against the file it claims to come from.

Why this exists: a snippet in src/data/portage.ts carries a path and a line
number, which is a promise that the viewer can open that file and see that
code. One snippet was written from memory and said `StatusPlaced` where the
repository says `StatusAwaitingDeposit` — plausible, wrong, and invisible to
tsc. Now the promise is checked.

Lines containing `…` are elisions and are skipped; everything else must appear
in the file, ignoring indentation.

Usage:  python3 scripts/verify-snippets.py
"""

import re
import sys
from pathlib import Path

LEARN = Path(__file__).resolve().parent.parent
REPO = LEARN.parent
DATA = LEARN / "src/data/portage.ts"

BLOCK = re.compile(
    r'(\w+):\s*\{\s*'
    r'path:\s*"([^"]+)",\s*'
    r'(?:line:\s*(\d+),\s*)?'
    r'lang:\s*"(go|php)",\s*'
    r'code:\s*\[(.*?)\n\s*\],',
    re.S,
)
STRING = re.compile(r'''^\s*(?:'((?:[^'\\]|\\.)*)'|"((?:[^"\\]|\\.)*)")\s*,?\s*$''')


def code_lines(raw: str) -> list[str]:
    out = []
    for ln in raw.splitlines():
        m = STRING.match(ln)
        if m:
            s = m.group(1) if m.group(1) is not None else m.group(2)
            out.append(s.replace('\\"', '"').replace("\\'", "'").replace("\\\\", "\\"))
    return out


def blob(lines: list[str]) -> str:
    """Whitespace and comments removed, so the comparison is about the code.

    Two differences are deliberately allowed and therefore normalised away:
      - a long line re-wrapped to fit the frame (whitespace)
      - an English comment translated to Vietnamese for the lesson (comments)
      - the trailing comma Go requires when a line is split (",}" and ",)")
    Anything else — a renamed constant, a different call — still shows up.
    """
    out = []
    for ln in lines:
        code = ln.split("//", 1)[0]
        out.append("".join(code.split()))
    blob = "".join(out)
    # Splitting one line into several forces a trailing comma in Go, so a
    # re-wrapped snippet legitimately gains commas the file does not have.
    return blob.replace(",}", "}").replace(",)", ")")


def main() -> int:
    text = DATA.read_text()
    checked = skipped = bad = 0

    for name, path, line, lang, raw in BLOCK.findall(text):
        if lang != "go":
            skipped += 1
            continue
        src = REPO / path
        if not src.exists():
            print(f"THIẾU FILE  {name:<18} {path}")
            bad += 1
            continue

        file_blob = blob(src.read_text().splitlines())
        # `…` marks an elision: each run between markers must appear on its own.
        fragments = [f for f in blob(code_lines(raw)).split("…") if f]
        missing = [f for f in fragments if f not in file_blob]

        checked += 1
        if missing:
            bad += 1
            print(f"LỆCH        {name:<18} {path}:{line or '?'}")
            for m in missing[:2]:
                print(f"              không có trong file: {m[:90]}")
        else:
            print(f"ok          {name:<18} {path}:{line or '?'}")

    print()
    print(f"{checked} snippet Go đối chiếu · {skipped} snippet PHP bỏ qua (minh hoạ, không phải code repo)")
    if bad:
        print(f"{bad} snippet không khớp file — sửa src/data/portage.ts")
    else:
        print("mọi dòng code lên hình đều có thật trong file nó khai.")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
