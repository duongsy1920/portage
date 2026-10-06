import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 3, episode 2 — recall. */
export const P2S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 2 · NHỚ LẠI"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
