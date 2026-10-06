import React from "react";
import { createRoot } from "react-dom/client";
import "./hub.css";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useRoute } from "./router";
import { ProgressProvider, useProgress } from "./store";
import { Shell } from "./components/Shell";
import { EpisodePage } from "./pages/EpisodePage";
import { ReviewPage } from "./pages/ReviewPage";
import { RoadmapPage } from "./pages/RoadmapPage";
import { byId, nextEpisode } from "./catalog";

/**
 * staticFile() in a scene resolves "sfx/ding.wav" to "/sfx/ding.wav" — the
 * site root. Under GitHub Pages the site root is /portage/, so the sound would
 * 404. Remotion consults window.remotion_staticFiles first; filling it with
 * the eight sounds, relative to the page, keeps the scenes untouched and the
 * build relocatable (vite base "./").
 */
const SOUNDS = ["mouse-click", "switch", "whoosh", "whip", "ding", "bone-crack", "loading-lag", "page-turn"];
(window as unknown as { remotion_staticFiles: { name: string; src: string }[] }).remotion_staticFiles = SOUNDS.map((s) => ({
  name: `sfx/${s}.wav`,
  src: `${import.meta.env.BASE_URL}sfx/${s}.wav`,
}));

const App: React.FC = () => {
  const route = useRoute();
  const { progress } = useProgress();
  // "#/" is "continue": the next episode to watch, chosen from progress.
  const current = route.page === "home" ? nextEpisode(progress) : route.page === "episode" ? byId(route.id) : undefined;
  return (
    <Shell route={route} currentId={current?.id}>
      {route.page === "home" ? <EpisodePage id={current!.id} continuing /> : null}
      {route.page === "episode" ? <EpisodePage id={route.id} /> : null}
      {route.page === "review" ? <ReviewPage /> : null}
      {route.page === "roadmap" ? <RoadmapPage /> : null}
    </Shell>
  );
};

createRoot(document.getElementById("root")!).render(
  <ProgressProvider>
    <TooltipProvider delayDuration={300}>
      <App />
    </TooltipProvider>
  </ProgressProvider>,
);
