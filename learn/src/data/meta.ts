/**
 * The commit of the Go repository that every snippet and figure on screen was
 * last checked against (scripts/verify-numbers.sh, scripts/verify-snippets.py).
 * The hub links "neo vào code" to THIS commit, not to main: line numbers drift,
 * and a link that opens the wrong line teaches the wrong thing.
 *
 * Update it when the two checks are re-run green against a newer commit.
 */
export const VERIFIED = {
  commit: "ef5661dd291f2a20d780609b9c1c8b5295dc935a",
  short: "ef5661d",
  date: "2026-10-06",
  repo: "https://github.com/duongsy1920/portage",
} as const;

export const sourceURL = (path: string, line?: number): string =>
  `${VERIFIED.repo}/blob/${VERIFIED.commit}/${path}${line ? `#L${line}` : ""}`;
