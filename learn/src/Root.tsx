import React from "react";
import { Composition, Folder } from "remotion";
import "./index.css";
import { FPS, WIDTH, HEIGHT } from "./design/tokens";
import { SEASONS, componentFor } from "./videos";
import { totalFrames } from "./videos/registry";

/**
 * Every episode, plus every scene registered on its own, so a scene can be
 * opened and re-timed in Studio without scrubbing through the whole video to
 * reach it. Everything here is derived from src/videos/index.ts — nothing in
 * this file knows the name of any particular episode.
 */
export const RemotionRoot: React.FC = () => (
  <>
    {SEASONS.map((season) => (
      <Folder key={season.folder} name={season.folder}>
        {season.specs.map((spec) => (
          <React.Fragment key={spec.id}>
            <Composition
              id={spec.id}
              component={componentFor(spec)}
              durationInFrames={totalFrames(spec)}
              fps={FPS}
              width={WIDTH}
              height={HEIGHT}
            />
            {spec.scenes.map((scene, i) => (
              <Composition
                key={scene.name}
                id={`${spec.scenePrefix}${i + 1}`}
                component={scene.component}
                durationInFrames={scene.durationInFrames}
                fps={FPS}
                width={WIDTH}
                height={HEIGHT}
              />
            ))}
          </React.Fragment>
        ))}
      </Folder>
    ))}
  </>
);
