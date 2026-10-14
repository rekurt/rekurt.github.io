import { expect, test } from '@playwright/test';

test('published English article is reachable, readable and honest about validation', async ({ page }) => {
  await page.goto('/articles/');
  await page.getByRole('heading', { name: 'prt 0.6.0: find the process behind a local port' }).getByRole('link').click();
  await expect(page).toHaveURL(/\/articles\/prt-0-6-0-port-ownership\/$/);
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('prt 0.6.0: find the process behind a local port');
  await expect(page.locator('html')).toHaveAttribute('lang', 'en');
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://rekurt.github.io/articles/prt-0-6-0-port-ownership/');
  await expect(page.locator('link[hreflang]')).toHaveCount(0);
  await expect(page.getByText('It did not test Linux,', { exact: false })).toBeVisible();
  await expect(page.getByText('Disclosure: this article was drafted with AI assistance.', { exact: false })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
});
