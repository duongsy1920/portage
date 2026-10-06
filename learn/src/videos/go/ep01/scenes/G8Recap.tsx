import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 1, episode 1 — recall, plus the questions an interviewer asks. */
export const G8Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 1 · NHỚ LẠI"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
