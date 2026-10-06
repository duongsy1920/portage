import { SNIPPETS, GO_SNIPPETS, DDD_SNIPPETS, PHP_SNIPPETS, GO_USAGE, type Snippet } from "../../src/data/portage";
import { sourceURL } from "../../src/data/meta";

/**
 * `spec.anchors` names the pieces of src/data/portage.ts an episode puts on
 * screen, as "GO_SNIPPETS.mutex" or "GO_USAGE.defers". This resolves a name to
 * the thing — a snippet with its file and line, or a counted figure — so the
 * episode page can show "neo vào code" and link to the verified commit.
 */
const SNIPPET_SETS: Record<string, Record<string, Snippet>> = {
  SNIPPETS: SNIPPETS as unknown as Record<string, Snippet>,
  GO_SNIPPETS: GO_SNIPPETS as unknown as Record<string, Snippet>,
  DDD_SNIPPETS: DDD_SNIPPETS as unknown as Record<string, Snippet>,
  PHP_SNIPPETS: PHP_SNIPPETS as unknown as Record<string, Snippet>,
};

export type Anchor =
  | { kind: "snippet"; name: string; snippet: Snippet; url?: string }
  | { kind: "figure"; name: string; value: number }
  | { kind: "unknown"; name: string };

export const resolveAnchor = (name: string): Anchor => {
  const [set, key] = name.split(".");
  if (!set || !key) return { kind: "unknown", name };
  if (set === "GO_USAGE") {
    const value = (GO_USAGE as Record<string, number>)[key];
    return typeof value === "number" ? { kind: "figure", name, value } : { kind: "unknown", name };
  }
  const snippet = SNIPPET_SETS[set]?.[key];
  if (!snippet) return { kind: "unknown", name };
  const url = snippet.lang === "go" ? sourceURL(snippet.path, snippet.line) : undefined;
  return { kind: "snippet", name, snippet, url };
};
