import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 1, episode 8 — recall. */
export const N5Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 8 · NHỚ LẠI"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
