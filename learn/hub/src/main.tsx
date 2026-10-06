import React from "react";
import { createRoot } from "react-dom/client";
import "./hub.css";
import { useRoute, href, type Route } from "./router";
import { ProgressProvider } from "./store";
import { MapPage } from "./pages/MapPage";
import { EpisodePage } from "./pages/EpisodePage";
import { ReviewPage } from "./pages/ReviewPage";
import { RoadmapPage } from "./pages/RoadmapPage";
import { VERIFIED } from "../../src/data/meta";

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

const NAV: { route: Route; label: string }[] = [
  { route: { page: "map" }, label: "Bản đồ" },
  { route: { page: "review" }, label: "Ôn tập" },
  { route: { page: "roadmap" }, label: "roadmap.sh" },
];

const App: React.FC = () => {
  const route = useRoute();
  return (
    <>
      <div className="ground" aria-hidden="true" />
      <div className="wrap">
        <header className="top">
          <a className="brand" href={href({ page: "map" })}>
            Học <em>Go</em> · Portage
          </a>
          <nav aria-label="Trang">
            {NAV.map((n) => (
              <a key={n.label} href={href(n.route)} aria-current={route.page === n.route.page ? "page" : undefined}>
                {n.label}
              </a>
            ))}
          </nav>
        </header>
        <main>
          {route.page === "map" ? <MapPage /> : null}
          {route.page === "episode" ? <EpisodePage id={route.id} /> : null}
          {route.page === "review" ? <ReviewPage /> : null}
          {route.page === "roadmap" ? <RoadmapPage /> : null}
        </main>
        <footer className="foot">
          <span>Video và bản in dựng bằng Remotion từ chính code của Portage.</span>
          <a href={`${VERIFIED.repo}/tree/${VERIFIED.commit}/learn`}>Nguồn trên GitHub ↗</a>
          <span>Số và snippet kiểm lần cuối: {VERIFIED.date}.</span>
        </footer>
      </div>
    </>
  );
};

createRoot(document.getElementById("root")!).render(
  <ProgressProvider>
    <App />
  </ProgressProvider>,
);
