import { test, expect } from '@playwright/test';

test('basic navigation', async ({ page }) => {
  // Since we can't run the actual Wails app here, we point to the dev server
  // This is just a placeholder to show how E2E tests would be structured
  await page.goto('http://localhost:5173');

  // Expect a title "to contain" a substring.
  // await expect(page).toHaveTitle(/Ostenia/);
});
