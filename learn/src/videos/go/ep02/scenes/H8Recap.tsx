import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/**
 * Season 1, episode 2 — recall.
 *
 * Point three is the one worth the whole episode: package-level privacy is the
 * rule a Symfony developer will get wrong for months, because `private` looks
 * like a word they already know.
 */
export const H8Recap: React.FC = () => (
  <Recap
    eyebrow="MÙA 1 · TẬP 2 · NHỚ LẠI"
    points={RECAP.points}
    asked={RECAP.asked}
    nextLabel={RECAP.next.label}
    nextText={RECAP.next.text}
  />
);
