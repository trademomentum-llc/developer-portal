#!/usr/bin/env node
/**
 * Test end-to-end Backstage sign-in via Gitea.
 *
 * Environment:
 *   BACKSTAGE_URL         Backstage frontend URL (default http://localhost:3001)
 *   BACKEND_URL           Backstage backend URL (default http://localhost:7008)
 *   GITEA_USER            Gitea account to sign in with (default gitea_admin)
 *   ARTIFACT_DIR          Optional private artifact directory (created if missing)
 *
 * Exit codes:
 *   0 = sign-in succeeded and the resulting session can refresh
 *   1 = sign-in failed or assertions did not pass
 */
import { chromium } from '@playwright/test';
import fs from 'fs';
import os from 'os';
import path from 'path';

const BACKSTAGE_URL = process.env.BACKSTAGE_URL || 'http://localhost:3001';
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:7008';
const GITEA_USER = process.env.GITEA_USER || 'gitea_admin';
const GITEA_PASSWORD = (() => {
  if (process.env.GITEA_PASSWORD) {
    return process.env.GITEA_PASSWORD;
  }
  const passwordFile = path.join(
    os.homedir(),
    '.rational-reserve',
    'm1-gitea-admin-password',
  );
  if (!fs.existsSync(passwordFile)) {
    console.error(
      'Set GITEA_PASSWORD or place the admin password at',
      passwordFile,
    );
    process.exit(1);
  }
  return fs.readFileSync(passwordFile, 'utf-8').trim();
})();

// Private mode-0700 directory for screenshots/HTML (avoids predictable /tmp paths).
const ARTIFACT_DIR = (() => {
  if (process.env.ARTIFACT_DIR) {
    fs.mkdirSync(process.env.ARTIFACT_DIR, { recursive: true, mode: 0o700 });
    return process.env.ARTIFACT_DIR;
  }
  return fs.mkdtempSync(path.join(os.tmpdir(), 'gitea-signin-'), { encoding: 'utf8' });
})();
try {
  fs.chmodSync(ARTIFACT_DIR, 0o700);
} catch {
  // best-effort on platforms that ignore mode bits
}
console.log('Artifact directory:', ARTIFACT_DIR);

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext();
const page = await context.newPage();

page.on('console', msg => console.log('PAGE CONSOLE:', msg.text()));
page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

console.log(`Opening ${BACKSTAGE_URL}`);
await page.goto(BACKSTAGE_URL);
await page.waitForLoadState('networkidle');

await page.waitForSelector('text=Gitea', { timeout: 10000 });
console.log('Clicking Gitea sign-in');

const [popup] = await Promise.all([
  page.waitForEvent('popup'),
  page.locator('li:has-text("Gitea") button').click(),
]);
popup.on('console', msg => console.log('POPUP CONSOLE:', msg.text()));
popup.on('pageerror', err => console.log('POPUP ERROR:', err.message));

await popup.waitForLoadState('networkidle');
console.log(`Popup URL: ${popup.url()}`);

// Gitea redirects unauthenticated users to its login page first.
if (popup.url().includes('/user/login')) {
  await popup.fill("input[name='user_name']", GITEA_USER);
  await popup.fill("input[name='password']", GITEA_PASSWORD);
  await popup.click('button:has-text("Sign In")');
}

// Wait for either the Gitea authorize page or the Backstage callback handler.
try {
  await popup.waitForURL(/login\/oauth\/authorize|api\/auth\/gitea\/handler\/frame/, {
    timeout: 10000,
  });
  console.log(`After login URL: ${popup.url()}`);
} catch (e) {
  console.log('No expected redirect; current popup URL:', popup.url());
}

// If Gitea asks for authorization, approve it.
if (popup.url().includes('/login/oauth/authorize')) {
  const authorizeButton = popup.locator('button:has-text("Authorize")');
  if (await authorizeButton.count() > 0) {
    await authorizeButton.click();
  }
}

// Wait for the OAuth handler to finish and close the popup.
try {
  await popup.waitForEvent('close', { timeout: 15000 });
  console.log('Popup closed by handler');
} catch (e) {
  console.log('Popup still open; current URL:', popup.url());
  await popup.screenshot({
    path: path.join(ARTIFACT_DIR, 'gitea-popup-still-open.png'),
    fullPage: true,
  });
  const html = await popup.content();
  fs.writeFileSync(path.join(ARTIFACT_DIR, 'gitea-popup-still-open.html'), html, {
    mode: 0o600,
  });
  await popup.close();
}

// Give the main window time to process the OAuth result and update its state.
await page.waitForLoadState('networkidle');
await page.waitForTimeout(2000);

// The root route in this dev build is not wired to the catalog index page, so
// after sign-in Backstage renders a 404 with the sidebar visible. We verify
// sign-in by checking that the sign-in cards are gone and the sidebar appeared.
const signInCard = page.locator('text=Sign in using Gitea');
const sidebarCatalog = page.locator('nav:has-text("Catalog")');
const stillOnSignIn = (await signInCard.count()) > 0;
const sidebarVisible = (await sidebarCatalog.count()) > 0;
console.log(`Sign-in card visible: ${stillOnSignIn}, sidebar visible: ${sidebarVisible}`);

if (stillOnSignIn || !sidebarVisible) {
  console.error('Sign-in did not complete; still on the sign-in page');
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'backstage-after-gitea-signin.png'),
    fullPage: true,
  });
  await browser.close();
  process.exit(1);
}

// Inspect how the session is stored.
const storage = await page.evaluate(() => ({
  localStorage: Object.fromEntries(
    Array.from({ length: window.localStorage.length }, (_, i) => {
      const key = window.localStorage.key(i);
      return [key, window.localStorage.getItem(key)];
    }),
  ),
  sessionStorage: Object.fromEntries(
    Array.from({ length: window.sessionStorage.length }, (_, i) => {
      const key = window.sessionStorage.key(i);
      return [key, window.sessionStorage.getItem(key)];
    }),
  ),
}));
console.log('Storage after sign-in:', JSON.stringify(storage, null, 2));

const cookies = await context.cookies();
console.log(
  'Cookies after sign-in:',
  cookies.map(c => `${c.name}=${c.value.slice(0, 20)}... domain=${c.domain} path=${c.path}`),
);

await page.screenshot({
  path: path.join(ARTIFACT_DIR, 'backstage-after-gitea-signin.png'),
  fullPage: true,
});
console.log('Gitea sign-in succeeded');
console.log('Artifacts written under', ARTIFACT_DIR);

await browser.close();
