import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 2, episode 3 — recall. */
export const C3Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 3 · NHỚ LẠI"
    accent="#9D7CD8"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
