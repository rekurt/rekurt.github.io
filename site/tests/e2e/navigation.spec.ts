import { expect, test } from "@playwright/test";

test("primary navigation and locale round trip", async ({ page }) => {
  await page.goto("/");
  await page.keyboard.press("Tab");
  await expect(page.getByRole("link", { name: "Skip to content" })).toBeFocused();
  await page.getByRole("link", { name: "Русский" }).click();
  await expect(page).toHaveURL(/\/ru\/$/);
  await page.getByRole("link", { name: "English" }).click();
  await expect(page).toHaveURL(/\/$/);
});

test("project actions map only to declared public surfaces", async ({ page }) => {
  await page.goto("/projects/vpn-hub/");
  await expect(page.getByRole("heading", { level: 1, name: "vpn-hub" })).toBeVisible();
  await expect(page.getByRole("link", { name: /^Website/ })).toHaveAttribute("href", "https://rekurt.github.io/vpn-hub/");
  await expect(page.getByRole("link", { name: /^Documentation/ }).first()).toHaveAttribute("href", "https://rekurt.github.io/vpn-hub/docs/");
});

test("catalog filters without hiding content by default", async ({ page }) => {
  await page.goto("/projects/");
  await expect(page.locator("[data-project-card]:visible")).toHaveCount(14);
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
  await expect(page.locator("tbody tr")).toHaveCount(49);
  await expect(page.getByRole("link", { name: "tsql", exact: true })).toBeVisible();
});
