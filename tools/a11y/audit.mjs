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

const pages = [
  ['login', '/login', false],
  ['dashboard', '/dashboard', true],
  ['problems', '/problems', true],
  ['hosts', '/hosts', true],
  ['services', '/services', true],
  ['host detail', '/hosts/winserver', true],
  ['service detail', '/service?host=localhost&service=Current%20Load', true],
  ['downtimes', '/downtimes', true],
  ['acknowledgements', '/acknowledgements', true],
  ['log entries', '/logentries', true],
  ['check history', '/history/checks', true],
  ['state changes', '/history/statechanges', true],
  ['notifications', '/history/notifications', true],
  ['command log', '/audit', true],
];

let total = 0;
for (const scheme of ['dark', 'light']) {
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, colorScheme: scheme });
  const page = await ctx.newPage();
  await page.goto(base + '/login', { waitUntil: 'networkidle' });
  await page.fill('#username', user);
  await page.fill('#password', pass);
  await page.click('button[type=submit]');
  await page.waitForURL('**/dashboard');

  for (const [name, path, authed] of pages) {
    if (!authed) continue;
    await page.goto(base + path, { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);
    const res = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();
    for (const v of res.violations) {
      total++;
      console.log(`[${scheme}] ${name}: ${v.id} (${v.impact}) x${v.nodes.length}`);
      console.log(`    ${v.help}`);
      console.log(`    e.g. ${v.nodes[0].html.slice(0, 140)}`);
    }
  }
  await ctx.close();
}

// The login page has no session, so it gets its own pass.
for (const scheme of ['dark', 'light']) {
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, colorScheme: scheme });
  const page = await ctx.newPage();
  await page.goto(base + '/login', { waitUntil: 'networkidle' });
  const res = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']).analyze();
  for (const v of res.violations) {
    total++;
    console.log(`[${scheme}] login: ${v.id} (${v.impact}) x${v.nodes.length}`);
    console.log(`    ${v.help}`);
    console.log(`    e.g. ${v.nodes[0].html.slice(0, 140)}`);
  }
  await ctx.close();
}

console.log(total === 0 ? 'no WCAG 2.1 AA violations' : `${total} violation group(s)`);
await browser.close();
