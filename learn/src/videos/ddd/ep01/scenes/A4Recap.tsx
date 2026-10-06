import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 2, episode 1 — recall. */
export const A4Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 1 · NHỚ LẠI"
    accent="#9D7CD8"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
