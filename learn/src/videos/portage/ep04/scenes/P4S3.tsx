import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 3, episode 4 — recall. */
export const P4S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 4 · NHỚ LẠI"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
