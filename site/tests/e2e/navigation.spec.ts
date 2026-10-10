import catalog from "../../src/data/generated/catalog.json" with { type: "json" };
import { expect, test } from "@playwright/test";

test("primary navigation and locale round trip", async ({ page }) => {
  await page.goto("/");
  await expect(page.locator('link[rel="alternate"]')).toHaveCount(4);
  await expect(page.locator('link[rel="alternate"][hreflang="zh-CN"]')).toHaveAttribute("href", "https://rekurt.github.io/zh-cn/");
  await page.keyboard.press("Tab");
  await expect(page.getByRole("link", { name: "Skip to content" })).toBeFocused();
  await page.getByRole("link", { name: "Русский" }).click();
  await expect(page).toHaveURL(/\/ru\/$/);
  await page.getByRole("link", { name: "简体中文" }).click();
  await expect(page).toHaveURL(/\/zh-cn\/$/);
  await page.getByRole("link", { name: "English" }).click();
  await expect(page).toHaveURL(/\/$/);
});

test("home page exposes author structured data", async ({ page }) => {
  await page.goto("/");
  const payload = await page.locator('script[type="application/ld+json"]').textContent();
  expect(payload).not.toBeNull();
  const schema = JSON.parse(payload ?? "{}");
  expect(schema).toMatchObject({
    "@context": "https://schema.org",
    "@type": "Person",
    name: "rekurt",
    url: "https://rekurt.github.io/",
  });
  expect(schema.sameAs).toContain("https://github.com/rekurt");
});

test("project actions map only to declared public surfaces", async ({ page }) => {
  await page.goto("/projects/prt/");
  await expect(page.locator(".project-hero").getByRole("heading", { level: 1, name: "prt", exact: true })).toBeVisible();
  const actions = page.locator(".project-actions");
  await expect(actions.getByRole("link", { name: /^Website/ })).toHaveAttribute("href", "https://rekurt.github.io/prt/");
  await expect(actions.getByRole("link", { name: /^Documentation/ })).toHaveAttribute("href", "https://crates.io/crates/prt");
});

test("catalog filters without hiding content by default", async ({ page }) => {
  await page.goto("/projects/");
  await expect(page.locator("[data-project-card]:visible")).toHaveCount(13);
  await page.getByRole("button", { name: "fintech", exact: true }).click();
  const visible = page.locator("[data-project-card]:visible");
  await expect(visible).not.toHaveCount(0);
  await expect(visible).toHaveCount(await page.locator('[data-project-card][data-domain="fintech"]').count());
});

test("maintained fork keeps upstream attribution", async ({ page }) => {
  await page.goto("/projects/mac-coffee/");
  await expect(page.getByText("Based on the upstream project")).toBeVisible();
  await expect(page.getByRole("link", { name: "Elliotwu-7/Mac-Coffee" }).first()).toHaveAttribute("href", "https://github.com/Elliotwu-7/Mac-Coffee");
});

test("registry contains the complete snapshot", async ({ page }) => {
  await page.goto("/registry/");
  await expect(page.locator("tbody tr")).toHaveCount(catalog.repositories.length);
  await expect(page.getByRole("link", { name: "tsql", exact: true })).toBeVisible();
});

for (const prefix of ["", "/ru", "/zh-cn"]) {
  test(`npm packages keep separate GitHub provenance in ${prefix || "en"}`, async ({ page }) => {
    const packages = [
      { slug: "depth", name: "@rekurt/depth", version: "0.1.0" },
      { slug: "openkline", name: "@rekurt/openkline-core", version: "0.2.0" },
      { slug: "gitlab-downloader", name: "gitlab-dump-cli", version: "0.2.0" },
    ];
    await page.goto(`${prefix}/projects/`);
    for (const pkg of packages) {
      const card = page.locator(`[data-project-card]`).filter({ has: page.locator(`h3 a[href="${prefix}/projects/${pkg.slug}/"]`) });
      await expect(card.locator(".npm-package")).toHaveAttribute("href", `https://www.npmjs.com/package/${pkg.name}/v/${pkg.version}`);
      await expect(card.locator(".npm-package")).toHaveText(`npm · ${pkg.name} · ${pkg.version} ↗`);
    }
    for (const pkg of packages) {
      await page.goto(`${prefix}/projects/${pkg.slug}/`);
      const status = page.locator(".project-status");
      await expect(status.locator(".npm-package")).toHaveAttribute("href", `https://www.npmjs.com/package/${pkg.name}/v/${pkg.version}`);
      const provenance = catalog.products.find(product => product.slug === pkg.slug)?.version;
      if (!provenance) throw new Error(`Missing GitHub provenance for ${pkg.slug}`);
      if (provenance.source !== "none") await expect(status.locator(".version-badge")).toHaveAttribute("title", provenance.source);
      if (provenance.value) await expect(status.locator(".version-badge")).toHaveText(provenance.value);
      await expect(status).toContainText("GitHub");
      await expect(status.locator("strong")).toHaveText(provenance.source === "none" ? "—" : provenance.source);
    }
  });
}
