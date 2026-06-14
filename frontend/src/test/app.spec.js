import { test, expect } from '@playwright/test';

test('basic navigation', async ({ page }) => {
  await page.goto('/');

  // Verify main navigation is present
  await expect(page.getByTitle('Activity Center')).toBeVisible();
  await expect(page.getByTitle('Plugin Management')).toBeVisible();
  await expect(page.getByTitle('SSH & Remote Files')).toBeVisible();

  // Switch to Plugin Management
  await page.getByTitle('Plugin Management').click();
  // Check if the header title changed
  await expect(page.getByRole('heading', { name: /Plugin Management/i })).toBeVisible();
});

test('theme toggle', async ({ page }) => {
  await page.goto('/');

  const themeToggle = page.getByTitle(/Switch to (Dark|Light) Mode/);
  await expect(themeToggle).toBeVisible();

  await themeToggle.click();
  // Verify class change on html element or local storage if applicable
  const html = page.locator('html');
  await expect(html).toHaveClass(/(dark|light)/);
});
