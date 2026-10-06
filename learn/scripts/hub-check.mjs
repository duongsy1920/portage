// Browser check of the learning hub (HUB-PLAN §6.2): every page, every episode,
// three widths, keyboard, reduced motion, progress persistence. Needs Chrome.
//
//   node scripts/hub-check.mjs http://localhost:4173/ <outDir>
//
// Exit 1 on any page error, console error (favicon aside), failed request, or
// failed assertion. Screenshots land in <outDir> for a human to look at —
// this script measures, it does not see.
import puppeteer from "puppeteer-core";
import { readFileSync } from "node:fs";

const base = (process.argv[2] ?? "http://localhost:4173/").replace(/\/?$/, "/");
const out = process.argv[3] ?? ".";
const ids = JSON.parse(readFileSync(new URL("../hub/.episode-ids.json", import.meta.url), "utf8"));

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

const text = () => page.evaluate(() => document.body.innerText.replace(/\s+/g, " "));
const hscroll = () => page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth + 1);
const go = async (hash) => { await page.goto(base + hash, { waitUntil: "networkidle0", timeout: 60000 }); };
const expect = (cond, msg) => { if (!cond) note(`assert: ${msg}`); };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// Start from nothing — ONCE. (evaluateOnNewDocument would clear on every navigation and
// make the persistence checks below fail for the wrong reason; that was the first run's bug.)
await go("#/");
await page.evaluate(() => { try { localStorage.clear(); } catch {} });

// ── map, three widths ───────────────────────────────────────────────────────
for (const [w, h] of [[1440, 900], [768, 1024], [390, 844]]) {
  await page.setViewport({ width: w, height: h });
  await go("#/");
  await page.waitForSelector(".dots .dot");
  const dots = await page.$$eval(".dots .dot", (els) => els.length);
  expect(dots === ids.length + 6, `map@${w}: ${dots} chấm, muốn ${ids.length + 6} (24 tập + 6 T1x)`);
  expect(!(await hscroll()), `map@${w}: cuộn ngang`);
  await page.screenshot({ path: `${out}/map-${w}.png`, fullPage: w !== 1440 });
}
await page.setViewport({ width: 1440, height: 900 });
await go("#/");
expect((await text()).includes(`Đã xem 0/${ids.length}`), "map: khởi đầu phải là 0/24");

// keyboard: Tab from the top must reach a dot within 20 presses
let reached = false;
for (let i = 0; i < 20 && !reached; i++) {
  await page.keyboard.press("Tab");
  reached = await page.evaluate(() => document.activeElement?.classList.contains("dot") ?? false);
}
expect(reached, "map: Tab không tới được chấm nào trong 20 lần");

// reduced motion turns the page's own animation off
await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "reduce" }]);
await go("#/");
await page.waitForSelector(".dot .n");
const anim = await page.$eval(".dot .n", (el) => getComputedStyle(el).animationName);
expect(anim === "none", `reduced-motion: .dot .n animation-name = ${anim}, muốn none`);
await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "no-preference" }]);

// ── every episode ───────────────────────────────────────────────────────────
let shot = false;
for (const id of ids) {
  await go(`#/ep/${encodeURIComponent(id)}`);
  await page.waitForSelector(".scenes li button", { timeout: 60000 });
  await page.waitForFunction(() => typeof window.__hub?.seek === "function", { timeout: 30000 });
  const scenes = await page.$$eval(".scenes li", (els) => els.length);
  const frames = await page.evaluate(() => window.__hub.frames);
  expect(scenes > 0 && frames > 0, `${id}: ${scenes} cảnh, ${frames} frame`);
  // seek to the end, let it play out: the episode must mark itself watched
  await page.evaluate(() => window.__hub.seek(window.__hub.frames - 3));
  // The header's own "· đã xem" badge — NOT body text, which always contains the hint "tự thành “đã xem”".
  await page.waitForSelector(".ep-head .ok", { timeout: 15000 }).catch(() => note(`${id}: không thành 'đã xem' sau khi tới cảnh cuối`));
  const saved = await page.evaluate((i) => { try { return Boolean(JSON.parse(localStorage.getItem("learn.hub.progress.v1") ?? "{}").watched?.[i]); } catch { return false; } }, id);
  expect(saved, `${id}: 'đã xem' không được ghi vào localStorage`);
  if (!shot) {
    await page.screenshot({ path: `${out}/episode-1440.png` });
    // plate tab: the iframe must load the real plate
    await page.$$eval(".tabs button", (bs) => bs.find((b) => b.textContent?.includes("Bản in"))?.click());
    await page.waitForSelector(".plate-frame iframe");
    await sleep(1500);
    const title = await page.evaluate(() => document.querySelector(".plate-frame iframe")?.contentDocument?.title ?? "");
    expect(title.includes("Plate"), `plate: iframe title "${title}" không phải bản in`);
    await page.screenshot({ path: `${out}/episode-plate-1440.png` });
    // code tab: a GitHub link to the verified commit
    await page.$$eval(".tabs button", (bs) => bs.find((b) => b.textContent?.includes("Neo"))?.click());
    await sleep(300);
    const links = await page.$$eval(".anchor a", (as) => as.map((a) => a.href));
    expect(links.some((l) => /github\.com\/.*\/blob\/[0-9a-f]{40}\//.test(l)), "code: không có link GitHub tới commit đã kiểm");
    // quiz tab: grade three questions "nhớ" → episode becomes confident
    await page.$$eval(".tabs button", (bs) => bs.find((b) => b.textContent?.includes("Tự kiểm"))?.click());
    await sleep(300);
    const nho = await page.$$(".grade button.nho");
    for (const b of nho) await b.click();
    await sleep(200);
    expect((await text()).includes("tự tin"), "quiz: chấm 'nhớ' 3 câu mà tập chưa 'tự tin'");
    await page.screenshot({ path: `${out}/episode-quiz-1440.png` });
    await page.setViewport({ width: 390, height: 844 });
    await go(`#/ep/${encodeURIComponent(id)}`);
    await page.waitForSelector(".scenes li button");
    expect(!(await hscroll()), "episode@390: cuộn ngang");
    await page.screenshot({ path: `${out}/episode-390.png`, fullPage: true });
    await page.setViewport({ width: 1440, height: 900 });
    shot = true;
  }
}

// ── progress persists, and the map now says 24/24 ───────────────────────────
await go("#/");
await page.waitForSelector(".dots .dot");
expect((await text()).includes(`Đã xem ${ids.length}/${ids.length}`), "map: sau 24 tập phải là 24/24");
await page.reload({ waitUntil: "networkidle0" });
expect((await text()).includes(`Đã xem ${ids.length}/${ids.length}`), "map: tải lại mất tiến độ");
const exported = await page.evaluate(() => localStorage.getItem("learn.hub.progress.v1"));
expect(exported && JSON.parse(exported).v === 1, "progress: localStorage không có bản v1");
await page.evaluate(() => localStorage.clear());
await page.reload({ waitUntil: "networkidle0" });
expect((await text()).includes(`Đã xem 0/${ids.length}`), "map: xoá localStorage phải về 0");
await page.evaluate((json) => localStorage.setItem("learn.hub.progress.v1", json), exported);
await page.reload({ waitUntil: "networkidle0" });
expect((await text()).includes(`Đã xem ${ids.length}/${ids.length}`), "map: nhập lại JSON phải về 24/24");

// ── review ──────────────────────────────────────────────────────────────────
await go("#/on");
await page.waitForSelector(".review-card, .card");
const t1 = await text();
const want = ids.length * 3 + 17;
expect(t1.includes(`${want} thẻ`), `review: muốn "${want} thẻ", thấy: ${t1.slice(0, 120)}`);
// The three "nhớ" grades given on the episode page above already count as graded cards,
// so the check is "+1", not "= 1".
const gradedBefore = Number(/Đã chấm (\d+)\//.exec(t1)?.[1] ?? -1);
await page.$$eval(".review-card .btn-row .btn", (bs) => bs.find((b) => b.textContent?.trim() === "nhớ")?.click());
await sleep(200);
const t2 = await text();
expect(t2.includes(`Đã chấm ${gradedBefore + 1}/${want}`), `review: chấm một thẻ phải thành ${gradedBefore + 1}/${want}, thấy: ${t2.slice(0, 160)}`);
await page.screenshot({ path: `${out}/review-1440.png` });

// ── roadmap ─────────────────────────────────────────────────────────────────
await go("#/roadmap");
await page.waitForSelector("table.road tbody tr");
const rows = await page.$$eval("table.road tbody tr", (els) => els.length);
await page.$$eval(".filters button", (bs) => bs.find((b) => b.textContent?.includes("cố ý bỏ"))?.click());
await sleep(200);
const dropped = await page.$$eval("table.road tbody tr", (els) => els.length);
expect(rows > dropped && dropped > 0, `roadmap: ${rows} dòng, lọc 'cố ý bỏ' còn ${dropped}`);
await page.screenshot({ path: `${out}/roadmap-1440.png` });
console.log(`roadmap: ${rows} dòng, cố ý bỏ ${dropped}`);

await browser.close();
console.log(`episodes: ${ids.length} mở được, tới cảnh cuối, thành 'đã xem'`);
if (problems.length) {
  console.log(`✗ ${problems.length} vấn đề:`);
  for (const p of problems.slice(0, 30)) console.log("  " + p);
  process.exit(1);
}
console.log("hub-check: sạch.");
