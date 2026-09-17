import React from "react";
import { AbsoluteFill } from "remotion";
import { TransitionSeries, linearTiming } from "@remotion/transitions";
import { slide } from "@remotion/transitions/slide";
import { BEAT, COLOR } from "../design/tokens";
import type { EpisodeSpec } from "./registry";

/**
 * Build an episode's composition from its spec.
 *
 * Every episode assembles the same way, so it is assembled once here rather
 * than in twenty-four near-identical index.tsx files — the kind of duplication
 * that stays correct right up until one copy drifts.
 *
 * Slide, not cross-fade: in a video whose payload is text, a cross-fade leaves
 * two headlines readable in the same spot for a moment, which reads as a
 * rendering mistake rather than as a transition.
 */
export const buildEpisode = (spec: EpisodeSpec): React.FC => {
  const Episode: React.FC = () => (
    <AbsoluteFill name={spec.id} style={{ backgroundColor: COLOR.bg }}>
      <TransitionSeries>
        {spec.scenes.map((scene, i) => {
          const Scene = scene.component;
          return (
            <React.Fragment key={scene.name}>
              <TransitionSeries.Sequence name={scene.name} durationInFrames={scene.durationInFrames}>
                <Scene />
              </TransitionSeries.Sequence>
              {i < spec.scenes.length - 1 ? (
                <TransitionSeries.Transition
                  presentation={slide({ direction: "from-bottom" })}
                  timing={linearTiming({ durationInFrames: BEAT.transition })}
                />
              ) : null}
            </React.Fragment>
          );
        })}
      </TransitionSeries>
    </AbsoluteFill>
  );
  Episode.displayName = spec.id;
  return Episode;
};
