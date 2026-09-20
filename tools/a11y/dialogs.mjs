import { chromium } from 'playwright';
import { AxeBuilder } from '@axe-core/playwright';
const base = process.env.SEI_URL ?? 'http://127.0.0.1:8090';
const user = process.env.SEI_USER ?? 'ops';
const pass = process.env.SEI_PASS ?? '';
if (!pass) {
  console.error('Set SEI_PASS to the password of an account that can see every page.');
  process.exit(2);
}

const browser = await chromium.launch();
let issues = 0;

for (const scheme of ['dark', 'light']) {
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, colorScheme: scheme });
  const page = await ctx.newPage();
  await page.goto(base + '/login', { waitUntil: 'networkidle' });
  await page.fill('#username', user);
  await page.fill('#password', pass);
  await page.click('button[type=submit]');
  await page.waitForURL('**/dashboard');
  await page.goto(base + '/hosts/winserver', { waitUntil: 'networkidle' });
  await page.waitForTimeout(600);

  for (const [label, name] of [['acknowledge', /^Acknowledge$/], ['downtime', /^Downtime$/],
                               ['passive result', /Submit result/], ['notify', /^Notify$/]]) {
    const btn = page.getByRole('button', { name }).first();
    if (!(await btn.count())) { console.log(`[${scheme}] ${label}: button not present`); continue; }
    await btn.click();
    await page.waitForTimeout(350);
    const res = await new AxeBuilder({ page }).withTags(['wcag2a','wcag2aa','wcag21a','wcag21aa']).analyze();
    for (const v of res.violations) {
      issues++;
      console.log(`[${scheme}] ${label}: ${v.id} (${v.impact}) x${v.nodes.length} - ${v.help}`);
      console.log(`    ${v.nodes[0].html.slice(0, 130)}`);
    }
    // Escape must close it: that is the whole point of using <dialog>.
    await page.keyboard.press('Escape');
    await page.waitForTimeout(250);
    const stillOpen = await page.evaluate(() => !!document.querySelector('dialog[open]'));
    if (stillOpen) { issues++; console.log(`[${scheme}] ${label}: Escape did not close the dialog`); }
  }
  await ctx.close();
}
console.log(issues === 0 ? 'dialogs: no violations, Escape closes each one' : `${issues} issue(s)`);
await browser.close();
