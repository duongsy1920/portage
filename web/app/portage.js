// portage.js — the only file on the two screens that talks to the API.
//
// One place decides how a token is attached, how an HTTP error becomes
// something a component can branch on, and how polling works. Served from the
// same origin as the API (cmd/api -web), so paths are relative and there is no
// CORS anywhere.
//
// [PHP] Cái ApiClient bọc quanh Guzzle: gắn header, dịch lỗi thành exception
// [PHP] có mã, không chứa logic nghiệp vụ.

const KEY = "portage.screens.v1";

const DEFAULTS = {
  operator: "dev-operator", // wire.DevOperatorToken — in-memory mode only
  customer: "dev-customer", // wire.DevCustomerToken
};

function read() {
  try {
    const raw = localStorage.getItem(KEY);
    return raw ? { ...DEFAULTS, ...JSON.parse(raw) } : { ...DEFAULTS };
  } catch {
    return { ...DEFAULTS };
  }
}

let tokens = read();

export function getTokens() {
  return { ...tokens };
}

export function setTokens(patch) {
  tokens = { ...tokens, ...patch };
  try { localStorage.setItem(KEY, JSON.stringify(tokens)); } catch { /* private mode */ }
}

export class ApiError extends Error {
  constructor(status, code, message) {
    super(message || code);
    this.status = status; // 0 when the request never left the browser
    this.code = code;     // the stable slug from errors.go, e.g. "wrong_amount"
  }
}

/**
 * call performs one request.
 *   as    "operator" | "customer" | "none"
 *   lang  sets Accept-Language, which decides "2.696.860" vs "2,696,860"
 * Throws ApiError on anything that is not 2xx, with .code taken from the body's
 * stable error code so a caller can retry on exactly one reason.
 */
export async function call(method, path, opts = {}) {
  const { body, as = "operator", lang } = opts;
  const headers = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (lang) headers["Accept-Language"] = lang;
  const token = as === "operator" ? tokens.operator : as === "customer" ? tokens.customer : "";
  if (token && as !== "none") headers.Authorization = "Bearer " + token;

  let res;
  try {
    res = await fetch(path, { method, headers, body: body === undefined ? undefined : JSON.stringify(body) });
  } catch (e) {
    throw new ApiError(0, "unreachable",
      "Không gọi được API. Trang này phải mở qua http://…/ui/, không phải file://");
  }
  const text = await res.text();
  let json = null;
  if (text) { try { json = JSON.parse(text); } catch { /* not JSON */ } }
  if (!res.ok) {
    const err = json && json.error;
    throw new ApiError(res.status, (err && err.code) || "http_" + res.status,
      (err && err.message) || text || res.statusText);
  }
  return json;
}

export const sleep = ms => new Promise(r => setTimeout(r, ms));

/** short id, for the rare place a uuid still has to be shown at all. */
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
 * retryOn re-runs a write while the API refuses for one of the given codes.
 * Those codes mean "the relay has not delivered yet" — anything else is a real
 * refusal and must surface immediately.
 */
export async function retryOn(codes, fn, tries = 12, gap = 400) {
  let last;
  for (let i = 0; i < tries; i++) {
    try {
      return await fn();
    } catch (e) {
      last = e;
      if (!(e instanceof ApiError) || !codes.includes(e.code)) throw e;
      await sleep(gap);
    }
  }
  throw last;
}
