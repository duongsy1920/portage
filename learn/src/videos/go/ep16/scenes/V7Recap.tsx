import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 1, episode 16 — recall, and the end of the extended season. */
export const V7Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 16 · HẾT MÙA 1"
    title="Bốn câu mang theo, và hết mùa một"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel="MÙA 2 · TẬP 1"
    nextText="Anemic model: entity toàn getter setter, và luật nghiệp vụ rơi mất ở đâu."
  />
);
