import { expect, test } from "@playwright/test";

for (const path of ["/", "/projects/", "/registry/"]) {
  test(`stays within the viewport: ${path}`, async ({ page }, testInfo) => {
    await page.goto(path);
    const dimensions = await page.evaluate(() => ({ width: document.documentElement.clientWidth, scrollWidth: document.documentElement.scrollWidth }));
    expect(dimensions.scrollWidth).toBe(dimensions.width);
    await page.screenshot({ path: testInfo.outputPath(`${path === "/" ? "home" : path.split("/")[1]}-${testInfo.project.name}.png`), fullPage: true });
  });
}
