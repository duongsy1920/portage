// ui.js — the shared React pieces both screens are built from.
//
// React without a build step: the UMD bundles define the globals, and htm
// turns tagged template literals into React.createElement calls. Same
// components, same hooks, no npm and no bundler — see docs/UI-GUIDE.md for why
// this repository chose that.

import htmFactory from "../vendor/htm.module.js";
import { Icon } from "./icons.js";
import { journeyOf, say, ORDER, ORDER_FOR_STAFF, TRACKING, YOURS as YOURS_WORDS } from "./words.js";
import { money } from "./portage.js";

export { Icon };
export const React = window.React;
export const html = htmFactory.bind(React.createElement);
const { useState, useEffect, useRef, useCallback, useId } = React;

/* ── top bar ─────────────────────────────────────────────────────────────── */
export function Top({ role, waiting, children }) {
  return html`
    <header class="top">
      <div class="top-in">
        <div class="brand"><span class="brand-mark"><${Icon} name="plane" size=${17} /></span><h1>Portage</h1></div>
        <span class="role">${role}</span>
        ${waiting > 0 && html`
          <span class="badge" role="status">${waiting}<span class="sr"> việc đang chờ bạn</span></span>`}
        <span class="spacer"></span>
        <nav aria-label="Màn hình khác">${children}</nav>
      </div>
    </header>`;
}

/* A section is a heading, an optional count and its rows: no box around it.
 * Boxes are for the one piece of work that is open (Sheet), so the eye can
 * find it without reading anything. */
export function Section({ title, icon, count, hot, right, children, id, first }) {
  return html`
    <section class=${"panel" + (first ? " first" : "")} id=${id} aria-label=${title}>
      <div class="section-head">
        ${icon && html`<${Icon} name=${icon} />`}
        <h2>${title}</h2>
        ${count !== undefined && count !== "" && html`<span class=${"count" + (hot ? " hot" : "")} key=${count}>${count}</span>`}
        <span class="spacer"></span>
        ${right}
      </div>
      <div class="section-body">${children}</div>
    </section>`;
}

export function Sheet({ children, label }) {
  return html`<div class="sheet" role="region" aria-label=${label}>${children}</div>`;
}

/** PageHead: the page's title, one line of what it is for, and live counts. */
export function PageHead({ title, lead, stats = [] }) {
  return html`
    <div class="page-head">
      <div><h2>${title}</h2>${lead && html`<p>${lead}</p>`}</div>
      <div class="stats">
        ${stats.map(s => html`
          <div class=${"stat" + (s.hot ? " hot" : "")} key=${s.label}><b key=${s.value}>${s.value}</b>${s.label}</div>`)}
      </div>
    </div>`;
}

/** Skeleton stands in while a list loads: its shape, not the word "loading". */
export function Skeleton({ rows = 3 }) {
  return html`<div class="skeleton" aria-label="Đang tải">${Array.from({ length: rows }, (_, i) => html`<i key=${i}></i>`)}</div>`;
}

/** Empty is an invitation to act, with a picture of what will appear here. */
export function Empty({ icon = "box", children }) {
  return html`<div class="empty"><${Icon} name=${icon} size=${28} /><div>${children}</div></div>`;
}

/** Problem is an error said in words, with an icon, never by colour alone. */
export function Problem({ children }) {
  return html`<div class="note bad" role="alert"><${Icon} name="alert" /><div>${children}</div></div>`;
}

// The label is TIED to the input with for/id. Not decoration: without it a
// screen reader reads an unnamed box, and a test cannot find the field by the
// only name a human knows it by either. Both symptoms, one cause.
export function Field({ label, hint, ...rest }) {
  const id = useId();
  return html`
    <div class="field">
      <label htmlFor=${id}>${label}</label>
      <input id=${id} ...${rest} />
      ${hint && html`<span class="hint">${hint}</span>`}
    </div>`;
}

export function Select({ label, hint, options, ...rest }) {
  const id = useId();
  return html`
    <div class="field">
      <label htmlFor=${id}>${label}</label>
      <select id=${id} ...${rest}>
        ${options.map(o => html`<option key=${o.value} value=${o.value}>${o.label}</option>`)}
      </select>
      ${hint && html`<span class="hint">${hint}</span>`}
    </div>`;
}


/* ── the numbered checklist ──────────────────────────────────────────────────
 * A step is done, current, or not yet. Done steps say WHAT THEY RECORDED
 * rather than going grey, because "US 9 · black" is the proof the step
 * happened and a grey button is not. Only the current step is interactive:
 * three buttons that all look available is how the first console taught
 * nobody anything.
 */
export function Steps({ steps }) {
  return html`
    <ol class="steps">
      ${steps.map((s, i) => html`
        <li class="step ${s.state}" key=${s.key || i} aria-current=${s.state === "now" ? "step" : undefined}>
          <div class="num">${s.state === "done" ? html`<${Icon} name="check" label="xong" size=${16} />` : i + 1}</div>
          <div>
            <div class="title">${s.title}</div>
            <div class="body">
              ${s.state === "done" && s.recorded &&
                html`<div class="recorded">${s.recorded}</div>`}
              ${s.state === "now" && html`
                <${React.Fragment}>
                  ${s.why && html`<div class="note why">${s.why}</div>`}
                  ${s.children}
                <//>`}
              ${s.state === "todo" && s.after && html`<div>${s.after}</div>`}
            </div>
          </div>
        </li>`)}
    </ol>`;
}

/* ── the journey strip ───────────────────────────────────────────────────────
 * The one loud element on both screens: an air waybill label with one box per
 * stage. Which stage is which comes from journeyOf in words.js, so this
 * component only draws. A box shows a figure or date only when the API sent
 * one for that stage; a blank box is honest, an invented one is not.
 */
export function Journey({ order, viewer = "customer", facts }) {
  const { cells, reached, stopped } = journeyOf(order, viewer, facts);
  // where the plane sits on the route: the middle of the stage it is in,
  // the end of the last one reached when the order stopped, home when delivered
  const p = order.status === "delivered" ? 1 : stopped ? (reached + 1) / cells.length : (reached + 1.5) / cells.length;
  const said = say(ORDER, order.status);
  const st = viewer === "staff" && ORDER_FOR_STAFF[order.status] ? { ...said, words: ORDER_FOR_STAFF[order.status] } : said;
  const mine = cells.find(c => c.state === "yours");
  const asks = mine && (YOURS_WORDS[viewer] || {})[mine.key];
  const where = order.tracking && order.tracking !== "none" && !["delivered", "cancelled", "purchase_failed"].includes(order.status)
    ? `${st.words}, ${say(TRACKING, order.tracking).words}` : st.words;
  const why = order.tracking && order.tracking !== "none" && !stopped ? say(TRACKING, order.tracking).why : st.why;

  return html`
    <div class="journey-box">
      <ol class="journey jr" style=${{ "--cells": cells.length }} aria-label="Hành trình của đơn">
        ${cells.map(c => html`
          <li class="jr-cell ${c.state}" key=${c.key}
            aria-current=${c.state === "now" || c.state === "yours" ? "step" : undefined}>
            <div class="jr-label"><${Icon} name=${c.state === "stopped" ? "x" : c.icon} size=${15} />${c.label}</div>
            ${(c.value || c.text) && html`<div class="jr-value">${c.value ? money(c.value) : c.text}</div>`}
            ${c.sub && html`<div class="jr-sub">${c.sub}</div>`}
            <span class="sr">${STATE_WORDS[c.state]}</span>
          </li>`)}
      </ol>
      <div class=${"jr-route" + (stopped ? " stopped" : "")} style=${{ "--p": p }} aria-hidden="true">
        <div class="jr-fill"></div>
        <span class="jr-plane"><${Icon} name=${stopped ? "x" : "plane"} size=${14} /></span>
      </div>
      <p class="jr-caption">
        <span class="where">${where.charAt(0).toUpperCase() + where.slice(1)}.</span>
        ${asks && html` <span class="yours">Việc của bạn: ${asks}.</span>`}
        ${why && html` <span class="muted">${why}</span>`}
        ${stopped && order.refund && html` <span>Hoàn lại <b>${money(order.refund)}</b>.</span>`}
      </p>
    </div>`;
}

// read aloud after each box, since the bar colour says it only to the eye
const STATE_WORDS = {
  done: "đã qua", now: "đang ở bước này", yours: "đang chờ bạn",
  todo: "chưa tới", stopped: "dừng ở đây", off: "sẽ không tới",
};

/* ── the key, folded into one line at the foot of the page ───────────────── */
export function Keys({ token, onSave, children }) {
  const [draft, setDraft] = useState(token);
  useEffect(() => setDraft(token), [token]);
  return html`
    <details class="keys">
      <summary><${Icon} name="key" /><span>Chìa khoá đang dùng: <b>${token || "chưa có"}</b></span>
        <span class="link">Đổi chìa</span></summary>
      <div class="keys-body">
        <div class="note">${children}</div>
        <${Field} ...${{ label: "Chìa khoá", value: draft, onChange: e => setDraft(e.target.value) }} />
        <div class="actions">
          <button onClick=${() => onSave(draft.trim())}>Lưu và tải lại</button>
          <span class="dim">Đổi chìa là đổi người, nên mọi danh sách trên trang tải lại theo.</span>
        </div>
      </div>
    </details>`;
}

/* ── toasts: how one role tells the other something happened ─────────────────
 * There is no websocket. Each screen polls its own queue and compares with
 * what it saw last time, which is honest about the fact that the read model is
 * eventually consistent anyway — a push would arrive before the projection.
 */
export function Toasts({ items, onDismiss }) {
  return html`
    <div class="toasts" role="status" aria-live="polite">
      ${items.map(t => html`
        <div class="toast" key=${t.id}>
          <${Icon} name="info" />
          <div><span class="k">${t.title}</span><br />${t.text}</div>
          <button onClick=${() => onDismiss(t.id)} aria-label="Đóng thông báo"><${Icon} name="x" size=${16} /></button>
        </div>`)}
    </div>`;
}

let toastSeq = 0;

/** useToasts returns [list, push, dismiss]; a toast fades itself after 9s. */
export function useToasts() {
  const [items, setItems] = useState([]);
  const dismiss = useCallback(id => setItems(xs => xs.filter(x => x.id !== id)), []);
  const push = useCallback((title, text) => {
    const id = ++toastSeq;
    setItems(xs => [...xs, { id, title, text }]);
    setTimeout(() => dismiss(id), 9000);
  }, [dismiss]);
  return [items, push, dismiss];
}

/**
 * usePoll re-runs load() every ms and hands the result to the component. It
 * also calls onChange(next, previous) so a screen can notice that work
 * appeared or a step got done, which is the whole notification mechanism.
 */
export function usePoll(load, ms, onChange) {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const prev = useRef(null);
  const changed = useRef(onChange);
  changed.current = onChange;

  const run = useCallback(async () => {
    try {
      const next = await load();
      setError(null);
      if (changed.current) changed.current(next, prev.current);
      prev.current = next;
      setData(next);
    } catch (e) {
      setError(e);
    }
  }, [load]);

  useEffect(() => {
    run();
    const t = setInterval(run, ms);
    return () => clearInterval(t);
  }, [run, ms]);

  return [data, error, run];
}

/** Busy wraps an async action so a button cannot be double-clicked. */
export function useAction() {
  const [busy, setBusy] = useState(false);
  const [problem, setProblem] = useState(null);
  const go = useCallback(async fn => {
    setBusy(true);
    setProblem(null);
    try {
      return await fn();
    } catch (e) {
      setProblem(e);
      return null;
    } finally {
      setBusy(false);
    }
  }, []);
  return { busy, problem, go, clear: () => setProblem(null) };
}
