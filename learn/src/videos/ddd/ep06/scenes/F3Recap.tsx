import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 2, episode 6 — recall, and the end of the season. */
export const F3Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 2 · TẬP 6 · HẾT MÙA 2"
    accent="#9D7CD8"
    title={RECAP.title}
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
