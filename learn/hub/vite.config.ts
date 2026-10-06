import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { createReadStream, cpSync, existsSync, statSync } from "node:fs";
import { extname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

/**
 * The learning hub is a Vite app that lives INSIDE learn/ and reuses its
 * node_modules, its scenes and its data. Nothing here imports the Go repo.
 *
 * Two things make it a little unusual:
 *   - publicDir points at learn/public so the scenes' sound files resolve the
 *     same way they do under Remotion Studio;
 *   - ../cheatsheet (32 static A3 plates + plate.css + fonts) is served at
 *     /cheatsheet/ in dev and copied into dist/ at build, so the hub can show
 *     a plate in an iframe without a second copy of it anywhere.
 */
const HUB = fileURLToPath(new URL(".", import.meta.url));
const LEARN = resolve(HUB, "..");
const CHEATSHEET = join(LEARN, "cheatsheet");

const TYPES: Record<string, string> = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".ttf": "font/ttf",
  ".png": "image/png",
  ".pdf": "application/pdf",
};

const plates = (): Plugin => ({
  name: "portage-plates",
  configureServer(server) {
    server.middlewares.use((req, res, next) => {
      const url = (req.url ?? "").split("?")[0];
      if (!url.startsWith("/cheatsheet/")) return next();
      const file = resolve(CHEATSHEET, decodeURIComponent(url.slice("/cheatsheet/".length)));
      if (!file.startsWith(CHEATSHEET) || !existsSync(file) || statSync(file).isDirectory()) return next();
      res.setHeader("Content-Type", TYPES[extname(file)] ?? "application/octet-stream");
      createReadStream(file).pipe(res);
    });
  },
  closeBundle() {
    cpSync(CHEATSHEET, join(HUB, "dist", "cheatsheet"), { recursive: true });
  },
});

export default defineConfig({
  root: HUB,
  // Relative URLs, so the same build works at the site root and under
  // /portage/ on GitHub Pages. Routing is hash-based for the same reason.
  base: "./",
  publicDir: join(LEARN, "public"),
  plugins: [react(), tailwindcss(), plates()],
  resolve: { alias: { "@": resolve(HUB, "src") } },
  server: { port: 5173, fs: { allow: [LEARN] } },
  build: { outDir: join(HUB, "dist"), emptyOutDir: true },
});
