import { useEffect, useState } from "react";

/**
 * Hash routing on purpose: the hub is a static page that must work at the
 * site root, under /portage/ on GitHub Pages, and from a file:// double-click.
 * A path router would need a server that rewrites every URL to index.html.
 *
 * "home" is not a page of its own: it opens the next episode to watch, so the
 * first click of a session is always "continue".
 */
export type Route =
  | { page: "home" }
  | { page: "episode"; id: string }
  | { page: "review" }
  | { page: "roadmap" };

export const parse = (hash: string): Route => {
  const h = hash.replace(/^#\/?/, "");
  if (h.startsWith("ep/")) return { page: "episode", id: decodeURIComponent(h.slice(3)) };
  if (h === "on") return { page: "review" };
  if (h === "roadmap") return { page: "roadmap" };
  return { page: "home" };
};

export const href = (r: Route): string => {
  switch (r.page) {
    case "episode":
      return `#/ep/${encodeURIComponent(r.id)}`;
    case "review":
      return "#/on";
    case "roadmap":
      return "#/roadmap";
    default:
      return "#/";
  }
};

export const useRoute = (): Route => {
  const [route, setRoute] = useState<Route>(() => parse(window.location.hash));
  useEffect(() => {
    const on = () => setRoute(parse(window.location.hash));
    window.addEventListener("hashchange", on);
    return () => window.removeEventListener("hashchange", on);
  }, []);
  return route;
};
