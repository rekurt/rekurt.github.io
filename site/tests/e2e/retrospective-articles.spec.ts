import { expect, test } from "@playwright/test";

const articles = [
  ["git-barber-merge-squash-rebase", "Branch cleanup after merge, squash and rebase with git-barber", "gone does not mean integrated"],
  ["go-propisyu-grammar-precision", "Russian number wording in Go: test the grammar and the precision", "Precision is not rounding"],
  ["ymsdk-retry-payload-identity", "Testing message retries without sending a message", "What the test actually proves"],
];
for (const [slug, title, boundary] of articles) {
  test(`retrospective: ${slug}`, async ({ page }) => {
    await page.goto("/articles/");
    await page.getByRole("link", { name: title, exact: true }).click();
    await expect(page.getByRole("heading", { level: 1 })).toHaveText(title);
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute("href", `https://rekurt.github.io/articles/${slug}/`);
    await expect(page.getByText("Checked 18 September 2026", { exact: false }).first()).toBeVisible();
    await expect(page.locator("article time")).toHaveAttribute("datetime", "2026-10-15");
    await expect(page.getByRole("heading", { name: boundary, exact: false })).toBeVisible();
    await expect(page.getByText("Disclosure: this article was drafted with AI assistance.", { exact: false })).toBeVisible();
    const schema = JSON.parse((await page.locator("script[data-content-schema]").textContent())!);
    expect(schema.datePublished).toBe("2026-10-15");
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth)).toBe(true);
  });
}
