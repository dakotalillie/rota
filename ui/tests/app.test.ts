import { test, expect } from '@playwright/test';

test('homepage', async ({ page }) => {
  await page.goto('http://localhost:5173/');

  await expect(page).toHaveScreenshot('homepage.png', { animations: 'disabled', fullPage: true });
});

test('settings page', async ({ page }) => {
  await page.goto('http://localhost:5173/');

  await page.getByRole('button', { name: 'Settings' }).click();

  await expect(page).toHaveScreenshot('settings.png', { animations: 'disabled', fullPage: true });
});
