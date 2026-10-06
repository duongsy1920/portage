import React from "react";
import { Recap } from "../../../../components/Recap";
import { RECAP } from "../recap";

/** Season 3, episode 8 — recall, and the end of the series. */
export const P8S3: React.FC = () => (
  <Recap
    eyebrow="MÙA 3 · TẬP 8 · HẾT LOẠT"
    title={RECAP.title}
    points={RECAP.points}
    asked={RECAP.asked}
  />
);
