#!/usr/bin/env python3
"""Refresh one review still per scene, at the LAST frame of that scene.

The last frame is the one worth inspecting: it is the only moment when every
staggered bullet, row and packet has finished arriving.

Usage:  python3 scripts/stills.py            # every episode
        python3 scripts/stills.py Go-Ep02    # only matching ids
"""

import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from episodes import LEARN, all_episodes  # noqa: E402


def main() -> int:
    wanted = sys.argv[1] if len(sys.argv) > 1 else ""
    eps = [e for e in all_episodes() if wanted in e.id]
    if not eps:
        print("không có tập nào khớp", wanted)
        return 2

    for ep in eps:
        out = LEARN / "out" / ep.out_dir / "canh"
        out.mkdir(parents=True, exist_ok=True)
        for png in out.glob("canh-*.png"):
            png.unlink()
        print(f"{ep.id} — {ep.title}")
        for i, scene_id in enumerate(ep.scene_ids, start=1):
            frame = ep.last_frame(i)
            subprocess.run(
                ["npx", "remotion", "still", scene_id, str(out / f"canh-{i}.png"),
                 f"--frame={frame}", "--scale=0.5", "--log=error", "--overwrite"],
                cwd=LEARN, check=True, capture_output=True,
            )
            print(f"  canh-{i}.png  frame {frame}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
