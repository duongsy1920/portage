// api.js — the only file that talks to the Portage API.
//
// Everything the console does goes through call(), so one place decides how a
// token is attached, how an error becomes something the UI can branch on, and
// what gets written to the call log. Served from the same origin as the API
// (cmd/api -web), so paths are relative and no CORS is involved.
//
// [PHP] Đây là cái ApiClient/HttpClient bạn hay bọc quanh Guzzle: gắn header,
// [PHP] dịch lỗi HTTP thành exception có mã, ghi log — không có logic nghiệp vụ.

const KEY = "portage.console.v1";

const DEFAULTS = {
  base: "",                 // "" = same origin as this page; the console is served by cmd/api -web
  op: "dev-operator",       // wire.DevOperatorToken — memory mode only
  cust: "dev-customer",     // wire.DevCustomerToken
};

/* ── config + session, both in localStorage so a reload keeps the run ──── */
function read(name, fallback) {
  try {
    const raw = localStorage.getItem(KEY + "." + name);
    return raw ? { ...fallback, ...JSON.parse(raw) } : { ...fallback };
  } catch {
    return { ...fallback };
  }
}
function write(name, value) {
  try { localStorage.setItem(KEY + "." + name, JSON.stringify(value)); } catch { /* private mode */ }
}

let _cfg = read("cfg", DEFAULTS);
let _session = read("session", {});

export const cfg = () => _cfg;
export function setCfg(patch) { _cfg = { ..._cfg, ...patch }; write("cfg", _cfg); fire(); }

/** The ids one run has produced so far: merchant, product, variant, quote, order, … */
export const session = () => _session;
export function setSession(patch) { _session = { ..._session, ...patch }; write("session", _session); fire(); }
export function resetSession() { _session = { run: newRunTag() }; write("session", _session); fire(); }

/** A tag unique to this run, because merchants.site is UNIQUE and a repeated
 *  source_url would make the next publish fail with suspected_duplicate. */
export function newRunTag() { return Date.now().toString(36); }
export function runTag() {
  if (!_session.run) setSession({ run: newRunTag() });
  return _session.run;
}

/* ── subscribers: the whole UI re-renders on change ───────────────────── */
const subs = new Set();
export function onChange(fn) { subs.add(fn); return () => subs.delete(fn); }
function fire() { for (const fn of subs) fn(); }

/* ── call log: every request, so the UI is never a black box ──────────── */
export const log = [];
export function clearLog() { log.length = 0; fire(); }

export class ApiError extends Error {
  constructor(status, code, message, entry) {
    super(message || code);
    this.status = status;   // 0 when the request never left the browser
    this.code = code;       // the stable slug from errors.go, e.g. "wrong_amount"
    this.entry = entry;
  }
}

/**
 * call performs one API request.
 *   as   "operator" | "customer" | "none"  → which bearer token to send
 *   lang sets Accept-Language, which is what decides "2.696.860" vs "2,696,860"
 * Throws ApiError on anything that is not 2xx, with .code taken from the
 * body's stable error code so callers can retry on exactly one reason.
 */
export async function call(method, path, opts = {}) {
  const { body, as = "operator", lang } = opts;
  const headers = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (lang) headers["Accept-Language"] = lang;
  const token = as === "operator" ? _cfg.op : as === "customer" ? _cfg.cust : "";
  if (token && as !== "none") headers.Authorization = "Bearer " + token;

  const started = performance.now();
  const entry = { method, path, as, body, at: new Date(), status: 0, ms: 0, json: null, text: "", error: null };
  log.unshift(entry);
  if (log.length > 150) log.pop();

  let res;
  try {
    res = await fetch(_cfg.base + path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch (e) {
    entry.ms = Math.round(performance.now() - started);
    entry.error = String((e && e.message) || e);
    fire();
    throw new ApiError(0, "unreachable",
      "Không gọi được API (" + entry.error + "). API còn chạy không? Trang này phải mở qua http://localhost:8080/ui/, không phải file://", entry);
  }

  entry.status = res.status;
  entry.ms = Math.round(performance.now() - started);
  entry.text = await res.text();
  if (entry.text) { try { entry.json = JSON.parse(entry.text); } catch { /* not JSON */ } }
  fire();

  if (!res.ok) {
    const err = entry.json && entry.json.error;
    throw new ApiError(res.status, (err && err.code) || "http_" + res.status,
      (err && err.message) || entry.text || res.statusText, entry);
  }
  return entry.json;
}

/* ── helpers the steps and screens share ──────────────────────────────── */
export const sleep = ms => new Promise(r => setTimeout(r, ms));
export const short = id => (id ? String(id).slice(0, 8) : "—");

/** money as the API sends it: {amount:"5393720", currency:"VND"} */
export function money(m) {
  if (!m || !m.amount) return "—";
  if (m.currency === "VND") {
    const n = Number(m.amount);
    return (Number.isFinite(n) ? n.toLocaleString("vi-VN") : m.amount) + " ₫";
  }
  return m.amount + " " + m.currency;
}

/**
 * poll calls fn until it returns something truthy. Used where a projection
 * has its own read endpoint, so the console can SHOW the wait instead of
 * hiding it — that wait is the outbox relay, and it is real.
 */
export async function poll(report, label, fn, tries = 12, gap = 400) {
  for (let i = 1; i <= tries; i++) {
    report(`${label} — thử lần ${i}/${tries}`);
    const v = await fn();
    if (v) return v;
    await sleep(gap);
  }
  throw new ApiError(0, "not_relayed",
    `${label}: hết ${tries} lần thử. Với -dsn thì cmd/worker phải đang chạy; in-memory thì relay nằm trong cmd/api.`, null);
}

/**
 * retryOn re-runs a WRITE while the API refuses for one of the given codes.
 * Those codes mean "the relay has not delivered yet" — anything else is a
 * real refusal and must surface immediately.
 */
export async function retryOn(report, codes, fn, tries = 12, gap = 400) {
  let last;
  for (let i = 1; i <= tries; i++) {
    try {
      return await fn();
    } catch (e) {
      last = e;
      if (!(e instanceof ApiError) || !codes.includes(e.code)) throw e;
      report(`còn ${e.code} (lần ${i}/${tries}) — chờ worker relay rồi thử lại`);
      await sleep(gap);
    }
  }
  throw last;
}

/** health: one cheap operator read, so the top bar can say why nothing works. */
export async function health() {
  try {
    await call("GET", "/purchase-tasks", { as: "operator" });
    return { state: "up", text: "API ok · token operator ok" };
  } catch (e) {
    if (e.status === 401) return { state: "warn", text: "API ok · token operator sai (401)" };
    if (e.status === 403) return { state: "warn", text: "API ok · token này không phải operator (403)" };
    if (e.status === 0) return { state: "down", text: "Không gọi được API" };
    return { state: "warn", text: `API trả ${e.status} ${e.code}` };
  }
}
