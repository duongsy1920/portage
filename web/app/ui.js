// ui.js — the shared React pieces both screens are built from.
//
// React without a build step: the UMD bundles define the globals, and htm
// turns tagged template literals into React.createElement calls. Same
// components, same hooks, no npm and no bundler — see docs/UI-GUIDE.md for why
// this repository chose that.

import htmFactory from "../vendor/htm.module.js";

export const React = window.React;
export const html = htmFactory.bind(React.createElement);
const { useState, useEffect, useRef, useCallback, useId } = React;

/* ── top bar ─────────────────────────────────────────────────────────────── */
export function Top({ title, role, waiting, children }) {
  return html`
    <div class="top">
      <h1>${title}</h1>
      <span class="role">${role}</span>
      ${waiting > 0 && html`<span class="badge" title="đang chờ bạn">${waiting}</span>`}
      <span class="spacer"></span>
      ${children}
    </div>`;
}

export function Card({ title, right, children }) {
  return html`
    <div class="card">
      ${title && html`
        <div class="card-head">
          <h2>${title}</h2>
          <span class="spacer"></span>
          ${right}
        </div>`}
      <div class="card-body">${children}</div>
    </div>`;
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
    <div class="steps">
      ${steps.map((s, i) => html`
        <div class="step ${s.state}" key=${s.key || i}>
          <div class="num">${s.state === "done" ? "✓" : i + 1}</div>
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
        </div>`)}
    </div>`;
}

/* ── toasts: how one role tells the other something happened ─────────────────
 * There is no websocket. Each screen polls its own queue and compares with
 * what it saw last time, which is honest about the fact that the read model is
 * eventually consistent anyway — a push would arrive before the projection.
 */
export function Toasts({ items, onDismiss }) {
  return html`
    <div class="toasts">
      ${items.map(t => html`
        <div class="toast" key=${t.id}>
          <div><span class="k">${t.title}</span><br />${t.text}</div>
          <span class="spacer" style=${{ marginLeft: "auto" }}></span>
          <button onClick=${() => onDismiss(t.id)} aria-label="đóng">✕</button>
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
