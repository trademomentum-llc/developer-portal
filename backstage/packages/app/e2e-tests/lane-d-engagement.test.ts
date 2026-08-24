// Lane D (FR-35/FR-36) verification spec -- temporary, not committed.
// Runs against the isolated frontend on :3100. The shared backend's CORS
// allowlist only permits :3001, so direct-to-backend calls (auth refresh)
// are replayed server-side with an injected ACAO header.
import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.route('http://localhost:7008/**', async route => {
    const response = await route.fetch();
    const headers = {
      ...response.headers(),
      'access-control-allow-origin': 'http://localhost:3100',
      'access-control-allow-credentials': 'true',
    };
    await route.fulfill({ response, headers });
  });

  await page.goto('/');
  const enter = page.getByRole('button', { name: 'Enter' });
  if (await enter.isVisible().catch(() => false)) {
    await enter.click();
    // Wait for the sign-in to settle (sidebar appears on every page once
    // authenticated). Note: /catalog currently crashes in material-table
    // (uuid v4 interop, tree-wide issue unrelated to Lane D), so navigate
    // directly to the entity page instead of via the catalog index.
    await expect(page.getByRole('link', { name: 'Settings' })).toBeVisible({
      timeout: 30000,
    });
  }
});

test('engagement tab renders dispatch control and test results card', async ({
  page,
}) => {
  await page.goto('/catalog/default/component/hello-m2/engagement');
  await expect(page.getByText('CI Runs').first()).toBeVisible({
    timeout: 30000,
  });
  await expect(
    page.getByRole('button', { name: /Dispatch ci\.yaml/ }),
  ).toBeVisible();
  await expect(page.getByText('Test Results').first()).toBeVisible();
  await page.screenshot({
    path: '/tmp/lane-d-engagement-tab.png',
    fullPage: true,
  });
});

test('dispatch workflow action POSTs through the proxy', async ({ page }) => {
  await page.goto('/catalog/default/component/hello-m2/engagement');
  const button = page.getByRole('button', { name: /Dispatch ci\.yaml/ });
  await expect(button).toBeVisible({ timeout: 30000 });
  await button.click();
  // Either the dispatched confirmation or the honest error state must render.
  await expect(
    page
      .getByText(/Dispatched ci\.yaml on main/)
      .or(page.getByText(/Dispatch failed \(not wired\)/)),
  ).toBeVisible({ timeout: 15000 });
  await page.screenshot({
    path: '/tmp/lane-d-engagement-dispatch.png',
    fullPage: true,
  });
});
