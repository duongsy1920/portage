import React from "react";
import { Audio, Sequence, staticFile } from "remotion";

/**
 * The sound palette. Deliberately small and deliberately boring.
 *
 * These are UI sounds, not jokes: the package also ships meme clips, and a
 * meme sting in the middle of an explanation costs the viewer the thread they
 * were holding. Each sound here means exactly one thing, every episode:
 */
export const CUE = {
  /** A box, a row, a step appears. The quietest sound in the set. */
  appear: "mouse-click",
  /** A new idea is introduced — a term card, a section. */
  reveal: "switch",
  /** Something travels: a packet on a wire, a value moving. */
  travel: "whoosh",
  /** A fast cut or a snap change of subject. */
  snap: "whip",
  /** A point lands, a conclusion, a test goes green. */
  land: "ding",
  /** Something breaks. Used once per episode at most. */
  crack: "bone-crack",
  /** Time passing, waiting, a process that is slow. */
  waiting: "loading-lag",
  /** Turning to a new scene. */
  turn: "page-turn",
} as const;

export type CueName = keyof typeof CUE;

/**
 * Default volumes, tuned so the explanation stays the foreground.
 *
 * A learning video is watched with the eyes; the sound is there to mark WHEN
 * something happened so the eye knows where to go. Anything loud enough to
 * notice on its own is too loud.
 */
const VOLUME: Record<CueName, number> = {
  appear: 0.18,
  reveal: 0.3,
  travel: 0.28,
  snap: 0.3,
  land: 0.4,
  crack: 0.5,
  waiting: 0.22,
  turn: 0.25,
};

/**
 * One sound, fired at one frame.
 *
 * `at` is the frame the sound STARTS, and it is meant to be the same number
 * the matching animation uses, so the ear and the eye agree. Reading a scene
 * file, a cue sits next to the thing it marks.
 */
export const Cue: React.FC<{
  at: number;
  sound: CueName;
  /** Override the default level for this one hit. */
  volume?: number;
}> = ({ at, sound, volume }) => {
  // <Audio> has no `from` of its own, so a headless <Sequence> places it on
  // the timeline. durationInFrames also clips a long sample so a slow one does
  // not bleed into the next scene.
  return (
    <Sequence from={at} durationInFrames={90} layout="none" name={`sfx ${sound}`}>
      <Audio src={staticFile(`sfx/${CUE[sound]}.wav`)} volume={volume ?? VOLUME[sound]} />
    </Sequence>
  );
};

/**
 * Several cues at once, so a scene can declare its whole sound track in one
 * place instead of scattering <Cue> tags through the markup.
 */
export const Cues: React.FC<{
  items: readonly { at: number; sound: CueName; volume?: number }[];
}> = ({ items }) => (
  <>
    {items.map((c, i) => (
      <Cue key={i} at={c.at} sound={c.sound} volume={c.volume} />
    ))}
  </>
);
