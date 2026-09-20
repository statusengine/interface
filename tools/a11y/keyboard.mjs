import { chromium } from 'playwright';
const base = process.env.SEI_URL ?? 'http://127.0.0.1:8090';
const user = process.env.SEI_USER ?? 'ops';
const pass = process.env.SEI_PASS ?? '';
if (!pass) {
  console.error('Set SEI_PASS to the password of an account that can see every page.');
  process.exit(2);
}

const browser = await chromium.launch();
const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, colorScheme: 'dark' });
const page = await ctx.newPage();
const problems = [];

// Log in with the keyboard alone.
await page.goto(base + '/login', { waitUntil: 'networkidle' });
await page.keyboard.press('Tab');
const first = await page.evaluate(() => document.activeElement?.textContent?.trim().slice(0, 30));
console.log('first tab stop on login:', JSON.stringify(first));

await page.fill('#username', user);
await page.fill('#password', pass);
await page.keyboard.press('Enter');          // submit from the password field
await page.waitForURL('**/dashboard', { timeout: 10000 });
console.log('signed in with Enter from the password field: yes');

// Walk the hosts list and check every stop is visible and has a ring.
await page.goto(base + '/hosts', { waitUntil: 'networkidle' });
await page.waitForTimeout(600);
const stops = [];
for (let i = 0; i < 40; i++) {
  await page.keyboard.press('Tab');
  const info = await page.evaluate(() => {
    const el = document.activeElement;
    if (!el || el === document.body) return null;
    const r = el.getBoundingClientRect();
    const cs = getComputedStyle(el);
    return {
      tag: el.tagName.toLowerCase(),
      label: (el.getAttribute('aria-label') || el.textContent || '').trim().slice(0, 28),
      visible: r.width > 0 && r.height > 0,
      offscreen: r.top < -5 || r.left < -5,
      outline: cs.outlineStyle !== 'none' && cs.outlineWidth !== '0px',
    };
  });
  if (!info) break;
  stops.push(info);
  if (!info.visible && !info.label.includes('Skip')) problems.push(`invisible stop: ${info.tag} "${info.label}"`);
  if (info.offscreen && !info.label.includes('Skip')) problems.push(`offscreen stop: ${info.tag} "${info.label}"`);
  if (!info.outline) problems.push(`no focus ring: ${info.tag} "${info.label}"`);
}
console.log(`tab stops walked on /hosts: ${stops.length}`);
console.log('  ' + stops.slice(0, 12).map((s) => s.label || s.tag).join(' > '));

// The skip link has to actually land on the main region.
await page.goto(base + '/hosts', { waitUntil: 'networkidle' });
await page.keyboard.press('Tab');
await page.keyboard.press('Enter');
await page.waitForTimeout(200);
const landed = await page.evaluate(() => document.activeElement?.id || location.hash);
console.log('skip link lands on:', landed);
if (!String(landed).includes('main')) problems.push('the skip link does not move focus to the main region');

console.log(problems.length ? problems : 'keyboard pass: no problems');
await browser.close();
