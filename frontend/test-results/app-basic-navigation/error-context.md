# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: app.spec.js >> basic navigation
- Location: src/test/app.spec.js:3:1

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByTitle('Activity Center')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByTitle('Activity Center')

```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  |
  3  | test('basic navigation', async ({ page }) => {
  4  |   await page.goto('/');
  5  |
  6  |   // Verify main navigation is present
> 7  |   await expect(page.getByTitle('Activity Center')).toBeVisible();
     |                                                    ^ Error: expect(locator).toBeVisible() failed
  8  |   await expect(page.getByTitle('Plugin Management')).toBeVisible();
  9  |   await expect(page.getByTitle('SSH & Remote Files')).toBeVisible();
  10 |
  11 |   // Switch to Plugin Management
  12 |   await page.getByTitle('Plugin Management').click();
  13 |   // Check if the header title changed
  14 |   await expect(page.getByRole('heading', { name: /Plugin Management/i })).toBeVisible();
  15 | });
  16 |
  17 | test('theme toggle', async ({ page }) => {
  18 |   await page.goto('/');
  19 |
  20 |   const themeToggle = page.getByTitle(/Switch to (Dark|Light) Mode/);
  21 |   await expect(themeToggle).toBeVisible();
  22 |
  23 |   await themeToggle.click();
  24 |   // Verify class change on html element or local storage if applicable
  25 |   const html = page.locator('html');
  26 |   await expect(html).toHaveClass(/(dark|light)/);
  27 | });
  28 |
```