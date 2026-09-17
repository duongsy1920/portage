#!/usr/bin/env python3
"""
Catch content that runs off the frame, or crowds its edge.

Why this exists: `tsc` is happy, the Studio preview looks fine at the frame you
happen to be scrubbed to, and a bullet that appears at frame 360 can still be
sitting half outside the picture. The only reliable detector is to look at the
pixels of the LAST frame of every scene, when everything has finished arriving.

Two failures are reported separately, because they have different causes:

  CLIPPED   content in the outermost rows/columns — it is being cut off
  CROWDED   content inside the margin but not at the very edge — it fits, but
            breaks the safe area the layout rules promise

Usage:  python3 scripts/check-overflow.py            # every scene
        python3 scripts/check-overflow.py Go-Ep01     # only matching ids
"""

import subprocess
import sys
from pathlib import Path

from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parent))
from episodes import all_episodes  # noqa: E402

LEARN = Path(__file__).resolve().parent.parent
SHOTS = LEARN / "out" / ".overflow-check"

# From src/design/tokens.ts. Kept here as plain numbers on purpose: this script
# is a second opinion, and a second opinion that imports the thing it is
# checking is not a second opinion.
SAFE_X, SAFE_TOP, SAFE_BOTTOM = 110, 76, 76

# Anything brighter than this counts as content. The background is #080B10
# (luma ~11) and the faint grid is #0E141C (luma ~19), so this sits clear of
# both while still catching dim text like #5C6E85 (luma ~105).
INK = 45

# The Stage deliberately places two labels inside the margin band: the episode
# eyebrow near the top and the source path near the bottom. They are part of the
# design, so the CROWDED check ignores the rows they live in.
# Ranh giới đo bằng pixel thật, không đoán: dấu thanh tiếng Việt của nhãn
# (MÙA 1 · TẬP 1) rơi xuống tới hàng 60, nên dải phải rộng tới đó.
EYEBROW_BAND = (16, 64)
SOURCE_BAND = (-66, -14)

# Font metric làm chữ nhô vào lề vài pixel mà mắt không thấy. Chỉ báo khi
# phần nhô đủ lớn để thật sự phá vùng an toàn.
TOLERANCE = 8


def last_frame(comp_id: str) -> int:
    """Durations live in the spec files; read them rather than guess."""
    return DURATIONS[comp_id] - 1


def shoot(comp_id: str, frame: int) -> Path:
    SHOTS.mkdir(parents=True, exist_ok=True)
    png = SHOTS / f"{comp_id}.png"
    subprocess.run(
        ["npx", "remotion", "still", comp_id, str(png),
         f"--frame={frame}", "--log=error", "--overwrite"],
        cwd=LEARN, check=True, capture_output=True,
    )
    return png


def scan(png: Path) -> dict:
    im = Image.open(png).convert("L")
    w, h = im.size
    px = im.load()

    def has_ink(xs, ys) -> bool:
        for y in ys:
            for x in xs:
                if px[x, y] > INK:
                    return True
        return False

    edge = 4  # the outermost band: ink here means it is being cut off
    clipped = {
        "top": has_ink(range(w), range(0, edge)),
        "bottom": has_ink(range(w), range(h - edge, h)),
        "left": has_ink(range(0, edge), range(h)),
        "right": has_ink(range(w - edge, w), range(h)),
    }

    eb = range(EYEBROW_BAND[0], EYEBROW_BAND[1])
    sb = range(h + SOURCE_BAND[0], h + SOURCE_BAND[1])
    crowded = {
        "top": has_ink(range(w), (y for y in range(edge, SAFE_TOP - TOLERANCE) if y not in eb)),
        "bottom": has_ink(range(w), (y for y in range(h - SAFE_BOTTOM + TOLERANCE, h - edge) if y not in sb)),
        "left": has_ink(range(edge, SAFE_X - TOLERANCE), range(h)),
        "right": has_ink(range(w - SAFE_X + TOLERANCE, w - edge), range(h)),
    }
    return {"clipped": clipped, "crowded": crowded}


DURATIONS: dict[str, int] = {}


def load_durations() -> None:
    """Every scene of every episode, found by reading the spec files."""
    for ep in all_episodes():
        for i, scene_id in enumerate(ep.scene_ids, start=1):
            DURATIONS[scene_id] = ep.durations[i - 1]


def main() -> int:
    load_durations()
    wanted = sys.argv[1] if len(sys.argv) > 1 else ""
    ids = [i for i in DURATIONS if wanted in i]
    if not ids:
        print("không có composition nào khớp", wanted)
        return 2

    bad = 0
    for comp_id in ids:
        frame = last_frame(comp_id)
        png = shoot(comp_id, frame)
        r = scan(png)
        clip = [k for k, v in r["clipped"].items() if v]
        crowd = [k for k, v in r["crowded"].items() if v]

        if clip:
            print(f"CẮT MẤT  {comp_id:<14} frame {frame:<5} mép: {', '.join(clip)}")
            bad += 1
        elif crowd:
            print(f"SÁT LỀ   {comp_id:<14} frame {frame:<5} mép: {', '.join(crowd)}")
            bad += 1
        else:
            print(f"ok       {comp_id:<14} frame {frame}")

    print()
    print(f"{len(ids)} cảnh · {bad} cảnh có vấn đề")
    print(f"ảnh đã lưu ở {SHOTS.relative_to(LEARN)}")
    return 1 if bad else 0


if __name__ == "__main__":
    raise SystemExit(main())
