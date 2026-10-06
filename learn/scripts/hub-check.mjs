// Browser check of the learning hub (HUB-PLAN §6.2, redone for the two-pane
// shadcn layout): every page, every episode, three widths, keyboard, reduced
// motion, progress persistence. Needs Chrome.
//
//   node scripts/hub-check.mjs http://localhost:4173/ <outDir>
//
// Exit 1 on any page error, console error (favicon aside), failed request, or
// failed assertion. Screenshots land in <outDir> for a human to look at —
// this script measures, it does not see. It finds elements by data-testid,
// never by utility class: Tailwind classes are not an interface.
import puppeteer from "puppeteer-core";
import { readFileSync } from "node:fs";

const base = (process.argv[2] ?? "http://localhost:4173/").replace(/\/?$/, "/");
const out = process.argv[3] ?? ".";
const meta = JSON.parse(readFileSync(new URL("../hub/.episode-ids.json", import.meta.url), "utf8"));
const ids = meta.ids;

const browser = await puppeteer.launch({
  executablePath: "/usr/bin/google-chrome",
  headless: true,
  args: ["--no-sandbox", "--autoplay-policy=no-user-gesture-required", "--mute-audio"],
});
const page = await browser.newPage();
const problems = [];
const note = (m) => problems.push(m);
page.on("pageerror", (e) => note(`pageerror: ${String(e).slice(0, 200)}`));
// "Failed to load resource" carries no URL here; the response listener below reports the same 404 WITH its URL.
page.on("console", (m) => { if (m.type() === "error" && !m.text().startsWith("Failed to load resource")) note(`console.error: ${m.text().slice(0, 200)}`); });
page.on("requestfailed", (r) => { if (!r.url().includes("favicon")) note(`requestfailed: ${r.url().slice(0, 140)}`); });
page.on("response", (r) => { if (r.status() >= 400 && !r.url().includes("favicon")) note(`${r.status()} ${r.url().slice(0, 140)}`); });

const T = (id) => `[data-testid="${id}"]`;
const text = () => page.evaluate(() => document.body.innerText.replace(/\s+/g, " "));
const hscroll = () => page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth + 1);
// A hash-only change to the SAME url is a same-document navigation: React state (the open tab) survives it.
// The checks want a fresh page each time, so an unchanged url is reloaded instead.
const go = async (hash) => {
  const url = base + hash;
  if (page.url() === url) await page.reload({ waitUntil: "networkidle0", timeout: 60000 });
  else await page.goto(url, { waitUntil: "networkidle0", timeout: 60000 });
};
const expect = (cond, msg) => { if (!cond) note(`assert: ${msg}`); };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const click = async (sel) => { await page.waitForSelector(sel, { timeout: 15000 }); await page.click(sel); };
// Visible only: the desktop sidebar is display:none under 1024 px but still in the DOM.
const count = (sel) => page.$$eval(sel, (els) => els.filter((e) => e.getClientRects().length > 0).length);

// Start from nothing — ONCE. (evaluateOnNewDocument would clear on every navigation and
// make the persistence checks below fail for the wrong reason.)
await go("#/");
await page.evaluate(() => { try { localStorage.clear(); localStorage.setItem("learn.hub.theme", "light"); } catch {} });

// ── home = next episode, three widths; the sidebar lists every episode ──────
for (const [w, h] of [[1440, 900], [768, 1024], [390, 844]]) {
  await page.setViewport({ width: w, height: h });
  await go("#/");
  await page.waitForSelector(T("episode-header"));
  if (w < 1024) {
    expect((await count(T("episode-row"))) === 0, `home@${w}: danh sách bài phải nằm trong ngăn kéo, không hiện sẵn`);
    await click(T("open-sidebar"));
    await page.waitForSelector(T("episode-row"));
    await sleep(600); // let the drawer finish sliding in before the screenshot
    await page.screenshot({ path: `${out}/drawer-${w}.png` });
  }
  const rows = await count(T("episode-row"));
  expect(rows === ids.length, `home@${w}: ${rows} dòng bài, muốn ${ids.length}`);
  expect(!(await hscroll()), `home@${w}: cuộn ngang`);
  await page.screenshot({ path: `${out}/home-${w}.png`, fullPage: w !== 1440 });
  if (w < 1024) await page.keyboard.press("Escape");
}
await page.setViewport({ width: 1440, height: 900 });
await go("#/");
await page.waitForSelector(T("progress-text"));
expect((await page.$eval(T("progress-text"), (el) => el.textContent)).includes(`Đã xem 0/${ids.length}`), `tiến độ: khởi đầu phải là 0/${ids.length}`);
expect((await text()).includes("Bắt đầu ở đây"), "home: chưa xem gì thì phải nói 'Bắt đầu ở đây'");

// keyboard: Tab from the top must reach an episode row within 25 presses
let reached = false;
for (let i = 0; i < 25 && !reached; i++) {
  await page.keyboard.press("Tab");
  reached = await page.evaluate(() => document.activeElement?.getAttribute("data-testid") === "episode-row");
}
expect(reached, "home: Tab không tới được dòng bài nào trong 25 lần");

// reduced motion: the progress bar must not animate
await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "reduce" }]);
await go("#/");
await page.waitForSelector(T("progress-bar"));
const dur = await page.$eval(`${T("progress-bar")} [data-slot="progress-indicator"]`, (el) => getComputedStyle(el).transitionDuration);
const durS = dur.endsWith("ms") ? parseFloat(dur) / 1000 : parseFloat(dur); // Chrome reports seconds ("1e-05s")
expect(durS <= 0.0001, `reduced-motion: thanh tiến độ transition-duration = ${dur}, muốn ≤ 0.01ms`);
await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "no-preference" }]);

// dark mode: the toggle flips the html class and survives a reload
await click(T("theme-toggle"));
await sleep(100);
expect(await page.evaluate(() => document.documentElement.classList.contains("dark")), "theme: bấm nút phải sang nền tối");
await page.reload({ waitUntil: "networkidle0" });
expect(await page.evaluate(() => document.documentElement.classList.contains("dark")), "theme: tải lại phải vẫn tối");
await page.screenshot({ path: `${out}/home-dark-1440.png` });
await click(T("theme-toggle"));
await sleep(100);
expect(!(await page.evaluate(() => document.documentElement.classList.contains("dark"))), "theme: bấm lại phải về sáng");

// ── every episode ───────────────────────────────────────────────────────────
let shot = false;
for (const id of ids) {
  await go(`#/ep/${encodeURIComponent(id)}`);
  await page.waitForSelector(T("toc-item"), { timeout: 60000 });
  await page.waitForFunction(() => typeof window.__hub?.seek === "function", { timeout: 30000 });
  const scenes = await count(T("toc-item"));
  const frames = await page.evaluate(() => window.__hub.frames);
  expect(scenes > 0 && frames > 0, `${id}: ${scenes} đoạn, ${frames} frame`);
  expect(await page.evaluate(() => document.querySelector('[data-testid="episode-row"][aria-current="page"]') !== null), `${id}: dòng bài không được đánh dấu đang xem`);
  // seek to the end, let it play out: the episode must mark itself watched
  await page.evaluate(() => window.__hub.seek(window.__hub.frames - 3));
  await page.waitForSelector(T("watched-badge"), { timeout: 15000 }).catch(() => note(`${id}: không thành 'Đã xem' sau khi tới đoạn cuối`));
  const saved = await page.evaluate((i) => { try { return Boolean(JSON.parse(localStorage.getItem("learn.hub.progress.v1") ?? "{}").watched?.[i]); } catch { return false; } }, id);
  expect(saved, `${id}: 'đã xem' không được ghi vào localStorage`);
  if (!shot) {
    await page.screenshot({ path: `${out}/episode-1440.png` });
    // cheatsheet tab: the iframe must load the real plate
    await click(T("tab-cheatsheet"));
    await page.waitForSelector(".plate-frame iframe");
    await sleep(1500);
    const title = await page.evaluate(() => document.querySelector(".plate-frame iframe")?.contentDocument?.title ?? "");
    expect(title.includes("Plate"), `cheatsheet: iframe title "${title}" không phải bản A3`);
    await page.screenshot({ path: `${out}/episode-cheatsheet-1440.png` });
    // code tab: a GitHub link to the verified commit
    await click(T("tab-code"));
    await sleep(300);
    const links = await page.$$eval(`${T("anchor")} a`, (as) => as.map((a) => a.href));
    expect(links.some((l) => /github\.com\/.*\/blob\/[0-9a-f]{40}\//.test(l)), "code: không có link GitHub tới commit đã kiểm");
    // quiz tab: grade three questions "nhớ" → episode becomes "Đã thuộc"
    await click(T("tab-quiz"));
    await sleep(300);
    const nho = await page.$$(T("grade-nho"));
    for (const b of nho) await b.click();
    await sleep(200);
    expect((await count(T("confident-badge"))) === 1, "quiz: chấm 'nhớ' 3 câu mà bài chưa 'Đã thuộc'");
    await page.screenshot({ path: `${out}/episode-quiz-1440.png` });
    await page.setViewport({ width: 390, height: 844 });
    await go(`#/ep/${encodeURIComponent(id)}`);
    await page.waitForSelector(T("toc-item"));
    expect(!(await hscroll()), "episode@390: cuộn ngang");
    await page.screenshot({ path: `${out}/episode-390.png`, fullPage: true });
    await page.setViewport({ width: 1440, height: 900 });
    shot = true;
  }
}

// ── progress persists, and the sidebar now says 30/30 ───────────────────────
const progressText = async () => { await page.waitForSelector(T("progress-text")); return page.$eval(T("progress-text"), (el) => el.textContent ?? ""); };
await go("#/");
expect((await progressText()).includes(`Đã xem ${ids.length}/${ids.length}`), `tiến độ: sau ${ids.length} bài phải là ${ids.length}/${ids.length}`);
await page.reload({ waitUntil: "networkidle0" });
expect((await progressText()).includes(`Đã xem ${ids.length}/${ids.length}`), "tiến độ: tải lại mất tiến độ");
const exported = await page.evaluate(() => localStorage.getItem("learn.hub.progress.v1"));
expect(exported && JSON.parse(exported).v === 1, "progress: localStorage không có bản v1");
await page.evaluate(() => localStorage.removeItem("learn.hub.progress.v1"));
await page.reload({ waitUntil: "networkidle0" });
expect((await progressText()).includes(`Đã xem 0/${ids.length}`), "tiến độ: xoá localStorage phải về 0");
await page.evaluate((json) => localStorage.setItem("learn.hub.progress.v1", json), exported);
await page.reload({ waitUntil: "networkidle0" });
expect((await progressText()).includes(`Đã xem ${ids.length}/${ids.length}`), `tiến độ: nạp lại JSON phải về ${ids.length}/${ids.length}`);

// ── review ──────────────────────────────────────────────────────────────────
await go("#/on");
await page.waitForSelector(T("review-card"));
const t1 = await page.$eval(T("review-summary"), (el) => el.textContent ?? "");
const want = ids.length * 3 + 17;
expect(t1.includes(`${want} câu hỏi`), `ôn tập: muốn "${want} câu hỏi", thấy: ${t1.slice(0, 120)}`);
// The three "nhớ" grades given on the episode page above already count as graded,
// so the check is "+1", not "= 1".
const gradedBefore = Number(/Đã chấm (\d+)\//.exec(t1)?.[1] ?? -1);
await click(T("review-grade-nho"));
await sleep(200);
const t2 = await page.$eval(T("review-summary"), (el) => el.textContent ?? "");
expect(t2.includes(`Đã chấm ${gradedBefore + 1}/${want}`), `ôn tập: chấm một câu phải thành ${gradedBefore + 1}/${want}, thấy: ${t2.slice(0, 160)}`);
await page.screenshot({ path: `${out}/review-1440.png` });

// ── roadmap ─────────────────────────────────────────────────────────────────
await go("#/roadmap");
await page.waitForSelector(T("roadmap-row"));
const rows = await count(T("roadmap-row"));
await click(T("roadmap-filter-dropped"));
await sleep(200);
const dropped = await count(T("roadmap-row"));
expect(rows > dropped && dropped > 0, `roadmap: ${rows} dòng, lọc 'cố ý bỏ' còn ${dropped}`);
await page.screenshot({ path: `${out}/roadmap-1440.png` });
console.log(`roadmap: ${rows} dòng, cố ý bỏ ${dropped}`);
await click(T("roadmap-filter-plate"));
await sleep(200);
const plateLinks = [...new Set(await page.$$eval(`${T("roadmap-table")} tbody a[href*='cheatsheet/']`, (as) => as.map((a) => a.href)))];
expect(plateLinks.length === 2, `roadmap: ${plateLinks.length} link cheatsheet tra cứu (cần 2)`);
for (const u of plateLinks) {
  const r = await fetch(u);
  expect(r.ok, `cheatsheet ${u}: HTTP ${r.status}`);
}
console.log(`roadmap: ${plateLinks.length} cheatsheet tra cứu mở được`);

// ── the words: nothing of the old vocabulary may leak back into the page ─────
for (const hash of ["#/", "#/on", "#/roadmap"]) {
  await go(hash);
  await sleep(300);
  const t = await text();
  for (const bad of ["Bản in", "bản in", "Neo vào", "neo vào", "tự tin", " thẻ ", "cảnh"]) {
    expect(!t.includes(bad), `${hash}: còn chữ cũ "${bad}" trên trang`);
  }
}

await browser.close();
console.log(`episodes: ${ids.length} mở được, tới đoạn cuối, thành 'Đã xem'`);
if (problems.length) {
  console.log(`✗ ${problems.length} vấn đề:`);
  for (const p of problems.slice(0, 30)) console.log("  " + p);
  process.exit(1);
}
console.log("hub-check: sạch.");
