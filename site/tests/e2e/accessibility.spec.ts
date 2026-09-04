import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

for (const path of ["/", "/projects/", "/projects/mac-coffee/", "/registry/", "/ru/"]) {
  test(`has no serious accessibility violations: ${path}`, async ({ page }) => {
    await page.goto(path);
    const report = await new AxeBuilder({ page }).analyze();
    const violations = report.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical");
    expect(violations, violations.map((violation) => `${violation.id}: ${violation.help}`).join("\n")).toEqual([]);
  });
}

test("catalog controls and cards are keyboard reachable", async ({ page }) => {
  await page.goto("/projects/");
  const focusable = page.locator(":focus");
  await page.keyboard.press("Tab");
  await expect(focusable).toHaveAttribute("href", "#content");
  for (let index = 0; index < 8; index += 1) await page.keyboard.press("Tab");
  await expect(focusable).toBeVisible();
});
