// Audit aksesibilitas otomatis dengan axe-core pada halaman yang sedang berjalan.
//
// Pakai:
//   node scripts/a11y.mjs http://localhost:5173 / /katalog /print
//
// Butuh Chrome terpasang. Set CHROME_PATH bila lokasinya tidak standar.
import { existsSync } from 'node:fs';
import { chromium } from 'playwright-core';
import AxeBuilder from '@axe-core/playwright';

const CHROME_CANDIDATES = [
  process.env.CHROME_PATH,
  'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Google/Chrome/Application/chrome.exe',
  `${process.env.LOCALAPPDATA ?? ''}/Google/Chrome/Application/chrome.exe`,
  '/usr/bin/google-chrome',
  '/usr/bin/chromium',
  '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
].filter(Boolean);

const executablePath = CHROME_CANDIDATES.find((p) => existsSync(p));
if (!executablePath) {
  console.error('Chrome tidak ditemukan. Set CHROME_PATH ke executable Chrome.');
  process.exit(2);
}

const [base = 'http://localhost:5173', ...paths] = process.argv.slice(2);
const targets = paths.length > 0 ? paths : ['/'];

const browser = await chromium.launch({ executablePath, headless: true });
// axe-core butuh halaman dari browser.newContext(), bukan browser.newPage().
const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });

// Untuk halaman yang butuh login: A11Y_TOKEN (dan opsional A11Y_USER) diisi ke
// localStorage sebelum skrip halaman berjalan.
if (process.env.A11Y_TOKEN) {
  const token = process.env.A11Y_TOKEN;
  const user = process.env.A11Y_USER || '';
  await context.addInitScript(
    ({ token, user }) => {
      localStorage.setItem('printku_token', token);
      if (user) localStorage.setItem('printku_user', user);
    },
    { token, user },
  );
}

const page = await context.newPage();

let total = 0;
for (const path of targets) {
  const url = new URL(path, base).href;
  // Vite dev server mempertahankan koneksi HMR, jadi 'networkidle' tak pernah
  // tercapai. Tunggu 'load' lalu beri waktu untuk hidrasi & fetch data.
  await page.goto(url, { waitUntil: 'load', timeout: 60000 });
  await page.waitForTimeout(2500);
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .analyze();

  const v = results.violations;
  total += v.length;
  console.log(`\n=== ${path} — ${v.length} pelanggaran ===`);
  for (const rule of v) {
    console.log(`[${rule.impact}] ${rule.id}: ${rule.help} (${rule.nodes.length} elemen)`);
    for (const node of rule.nodes.slice(0, 3)) {
      console.log(`   ${node.target.join(' ')}`);
      const msg = (node.failureSummary || '').split('\n').filter(Boolean).slice(0, 2).join(' | ');
      if (msg) console.log(`     ${msg}`);
    }
    if (rule.nodes.length > 3) console.log(`   … dan ${rule.nodes.length - 3} elemen lain`);
  }
}

await browser.close();
console.log(`\nTotal pelanggaran: ${total}`);
process.exit(total === 0 ? 0 : 1);
