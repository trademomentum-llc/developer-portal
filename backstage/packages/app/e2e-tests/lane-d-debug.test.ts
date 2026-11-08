// Lane D debug -- temporary, not committed.
import { test } from '@playwright/test';

test('debug catalog on 3100', async ({ page }) => {
  await page.route('http://localhost:7008/**', async route => {
    const response = await route.fetch();
    const headers = {
      ...response.headers(),
      'access-control-allow-origin': 'http://localhost:3100',
      'access-control-allow-credentials': 'true',
    };
    await route.fulfill({ response, headers });
  });
  page.on('pageerror', err => console.log('[pageerror]', String(err).slice(0, 300)));

  await page.goto('/');
  const enter = page.getByRole('button', { name: 'Enter' });
  if (await enter.isVisible().catch(() => false)) {
    await enter.click();
    await page.waitForTimeout(3000);
  }
  await page.goto('/catalog');
  await page.waitForTimeout(8000);
  console.log('url:', page.url());
  const body = await page.locator('body').innerText();
  console.log('body head:', body.slice(0, 300).replace(/\n/g, ' | '));
  await page.screenshot({ path: '/tmp/lane-d-catalog-3100.png', fullPage: true });
});
