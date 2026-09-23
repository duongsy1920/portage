// icons.js — the few pictures the two screens use, drawn as inline SVG.
//
// Hand-drawn in the Lucide manner (24px grid, 1.5 stroke, round joins) and
// kept in the repository for the same reason React is: the screens work with
// no network. No emoji anywhere: an emoji is drawn by the reader's operating
// system, so the same "✓" is a different picture on every machine.
//
// An icon next to words is decoration and is hidden from screen readers. An
// icon standing alone must be given a label, and Icon refuses to render one
// without it rather than ship an unnamed button.

import htmFactory from "../vendor/htm.module.js";

// Its own html binding, not ui.js's: ui.js imports this file, and a module
// cycle would hand PATHS an html that does not exist yet.
const html = htmFactory.bind(window.React.createElement);

const PATHS = {
  check: html`<path d="M20 6 9 17l-5-5" />`,
  x: html`<path d="M18 6 6 18M6 6l12 12" />`,
  link: html`<path d="M15 3h6v6M10 14 21 3M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />`,
  alert: html`<circle cx="12" cy="12" r="9" /><path d="M12 8v4.5M12 16h.01" />`,
  info: html`<circle cx="12" cy="12" r="9" /><path d="M12 11v5M12 8h.01" />`,
  key: html`<circle cx="7.5" cy="15.5" r="4.5" /><path d="m10.7 12.3 9.8-9.8M17 6l3 3M14.5 8.5l2.5 2.5" />`,
  scale: html`<path d="M4 20h16M6 20l2-11h8l2 11M12 9V5M9 5h6" />`,
  plane: html`<path d="M17.8 19.2 16 11l3.5-3.5C21 6 21.5 4 21 3c-1-.5-3 0-4.5 1.5L13 8 4.8 6.2c-.5-.1-.9.1-1.1.5l-.3.5c-.2.5-.1 1 .3 1.3L9 12l-2 3H4l-1 1 3 2 2 3 1-1v-3l3-2 3.5 5.3c.3.4.8.5 1.3.3l.5-.2c.4-.3.6-.7.5-1.2z" />`,
  box: html`<path d="M21 8 12 3 3 8v8l9 5 9-5V8zM3 8l9 5 9-5M12 13v8" />`,
  wallet: html`<path d="M19 7V5a2 2 0 0 0-2-2H5a2 2 0 0 0 0 4h14a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5" /><path d="M16 14h.01" />`,
  bag: html`<path d="M6 7h12l1 14H5L6 7zM9 7V6a3 3 0 0 1 6 0v1" />`,
  home: html`<path d="M3 11 12 4l9 7M5 10v10h14V10" />`,
  receipt: html`<path d="M6 3h12v18l-3-2-3 2-3-2-3 2V3zM9 8h6M9 12h6" />`,
  chevron: html`<path d="m9 6 6 6-6 6" />`,
  coffee: html`<path d="M4 9h13v5a5 5 0 0 1-5 5H9a5 5 0 0 1-5-5V9zM17 11h1.5a2.5 2.5 0 0 1 0 5H17M8 3v3M12 3v3" />`,
};

/**
 * Icon draws one picture. `label` makes it a standalone, announced image;
 * without one it is hidden from assistive technology as pure decoration.
 */
export function Icon({ name, label, size = 18 }) {
  const body = PATHS[name];
  if (!body) throw new Error(`icons.js: no icon named "${name}"`); // programmer error, fail loudly
  const a11y = label ? { role: "img", "aria-label": label } : { "aria-hidden": "true", focusable: "false" };
  return html`
    <svg class="icon" width=${size} height=${size} viewBox="0 0 24 24" fill="none" stroke="currentColor"
      strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" ...${a11y}>${body}</svg>`;
}
