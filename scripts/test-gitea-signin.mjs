#!/usr/bin/env node
/**
 * Test end-to-end Backstage sign-in via Gitea.
 * Expects Backstage dev server on localhost:3001 and Gitea on localhost:3333.
 */
import { chromium } from '@playwright/test';
import fs from 'fs';
import os from 'os';
import path from 'path';

const BACKSTAGE_URL = process.env.BACKSTAGE_URL || 'http://localhost:3001';
const GITEA_USER = process.env.GITEA_USER || 'gitea_admin';
const GITEA_PASSWORD_FILE = path.join(os.homedir(), '.rational-reserve', 'm1-gitea-admin-password');

if (!fs.existsSync(GITEA_PASSWORD_FILE)) {
  console.error('Gitea admin password file not found');
  process.exit(1);
}
const GITEA_PASSWORD = fs.readFileSync(GITEA_PASSWORD_FILE, 'utf-8').trim();

const ARTIFACT_DIR = process.env.ARTIFACT_DIR
  ? (fs.mkdirSync(process.env.ARTIFACT_DIR, { recursive: true, mode: 0o700 }), process.env.ARTIFACT_DIR)
  : fs.mkdtempSync(path.join(os.tmpdir(), 'gitea-signin-'));
try {
  fs.chmodSync(ARTIFACT_DIR, 0o700);
} catch {
  // best-effort
}

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext();
const page = await context.newPage();

console.log(`Opening ${BACKSTAGE_URL}`);
await page.goto(BACKSTAGE_URL);
await page.waitForLoadState('networkidle');

await page.waitForSelector('text=Gitea', { timeout: 10000 });
console.log('Clicking Gitea sign-in');

const [popup] = await Promise.all([
  page.waitForEvent('popup'),
  page.click('text=Gitea >> .. >> button'),
]);
await popup.waitForLoadState('networkidle');
console.log(`Popup URL: ${popup.url()}`);

if (popup.url().includes('login')) {
  await popup.fill("input[name='user_name']", GITEA_USER);
  await popup.fill("input[name='password']", GITEA_PASSWORD);
  await popup.click("button[type='submit']");
  await popup.waitForLoadState('networkidle');
  console.log(`After login URL: ${popup.url()}`);
}

if (popup.url().includes('/oauth/authorize') || popup.url().includes('/login/oauth/authorize')) {
  await popup.click('button:has-text("Authorize")');
  await popup.waitForLoadState('networkidle');
  console.log(`After authorize URL: ${popup.url()}`);
}

await popup.waitForEvent('close', { timeout: 20000 });
console.log('Popup closed');

await page.waitForLoadState('networkidle');
await page.waitForTimeout(2000);

const identity = await page.evaluate(() => {
  try {
    const item = localStorage.getItem('@backstage/core-plugin-api_GiteaAuth');
    return item ? JSON.parse(item) : null;
  } catch (e) {
    return null;
  }
});
console.log('Gitea auth session in localStorage:', identity);

const profile = await page.evaluate(async () => {
  try {
    const resp = await fetch('/api/auth/gitea/session');
    return await resp.json();
  } catch (e) {
    return { error: String(e) };
  }
});
console.log('Profile via /api/auth/gitea/session:', profile);

await page.screenshot({
  path: path.join(ARTIFACT_DIR, 'backstage-after-gitea-signin.png'),
  fullPage: true,
});
console.log('Artifacts written under', ARTIFACT_DIR);

await browser.close();
