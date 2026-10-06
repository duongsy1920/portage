import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 1, episode 10 — recall, and the end of the season. */
export const P5Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 10 · HẾT MÙA 1"
    title={RECAP.title}
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
