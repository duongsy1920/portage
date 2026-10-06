import { useCallback, useEffect, useState } from "react";

/**
 * Light / dark / follow-the-OS. index.html applies the saved choice before the
 * first paint; this hook keeps it in sync afterwards. The videos are dark by
 * design; the page around them follows the reader, like any other site.
 */
export type Theme = "light" | "dark" | "system";

const KEY = "learn.hub.theme";

const read = (): Theme => {
  try {
    const t = localStorage.getItem(KEY);
    return t === "light" || t === "dark" ? t : "system";
  } catch {
    return "system";
  }
};

const apply = (t: Theme) => {
  const dark = t === "dark" || (t === "system" && matchMedia("(prefers-color-scheme: dark)").matches);
  document.documentElement.classList.toggle("dark", dark);
};

export const useTheme = () => {
  const [theme, setTheme] = useState<Theme>(read);
  useEffect(() => {
    apply(theme);
    if (theme !== "system") return;
    const mq = matchMedia("(prefers-color-scheme: dark)");
    const on = () => apply("system");
    mq.addEventListener("change", on);
    return () => mq.removeEventListener("change", on);
  }, [theme]);
  const set = useCallback((t: Theme) => {
    setTheme(t);
    try {
      if (t === "system") localStorage.removeItem(KEY);
      else localStorage.setItem(KEY, t);
    } catch {
      /* private mode: the choice lives for this page load only */
    }
  }, []);
  const isDark = typeof document !== "undefined" && document.documentElement.classList.contains("dark");
  return { theme, set, isDark };
};
