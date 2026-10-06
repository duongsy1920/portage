import React, { createContext, useCallback, useContext, useMemo, useState } from "react";
import { load, save, type Progress } from "./progress";

/** Progress as React state, saved on every change. */
const Ctx = createContext<{ progress: Progress; update: (fn: (p: Progress) => Progress) => void } | null>(null);

export const ProgressProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [progress, setProgress] = useState<Progress>(load);
  const update = useCallback((fn: (p: Progress) => Progress) => {
    setProgress((prev) => {
      const next = fn(prev);
      save(next);
      return next;
    });
  }, []);
  const value = useMemo(() => ({ progress, update }), [progress, update]);
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
};

export const useProgress = () => {
  const v = useContext(Ctx);
  if (!v) throw new Error("useProgress ngoài ProgressProvider");
  return v;
};
