"""Discover every episode by reading the spec files.

The scripts deliberately parse src/videos/*/ep*/spec.ts rather than importing
anything TypeScript: a check that imports the thing it is checking is not an
independent check. Adding an episode needs no change here.
"""

import re
from pathlib import Path

LEARN = Path(__file__).resolve().parent.parent


class Episode:
    def __init__(self, spec_path: Path):
        text = spec_path.read_text()
        self.path = spec_path

        def field(name: str) -> str:
            m = re.search(rf'{name}:\s*"([^"]+)"', text)
            if not m:
                raise SystemExit(f"{spec_path}: thiếu trường {name}")
            return m.group(1)

        self.id = field("id")
        self.title = field("title")
        self.scene_prefix = field("scenePrefix")
        self.out_dir = field("outDir")
        self.durations = [int(m) for m in re.findall(r"durationInFrames: (\d+)", text)]
        if not self.durations:
            raise SystemExit(f"{spec_path}: không có cảnh nào")

    @property
    def scene_ids(self) -> list[str]:
        return [f"{self.scene_prefix}{i}" for i in range(1, len(self.durations) + 1)]

    def last_frame(self, i: int) -> int:
        """Last frame of scene i (1-based) — when everything has arrived."""
        return self.durations[i - 1] - 1


def all_episodes() -> list[Episode]:
    specs = sorted((LEARN / "src/videos").glob("*/ep*/spec.ts"))
    if not specs:
        raise SystemExit("không tìm thấy spec.ts nào dưới src/videos/")
    return [Episode(p) for p in specs]
