import { expect, test } from "@playwright/test";

const slugs = ["cortex-forge", "dbdiff", "depth", "git-barber", "gitlab-downloader", "go-propisyu", "gost-crypto", "prt", "sprint-velocity", "ymsdk"];
for (const slug of slugs) for (const locale of ["", "ru/", "zh-cn/"]) {
  test(`project layout: ${slug}/${locale || "en"}`, async ({ page }, info) => {
    test.skip(info.project.name !== "chromium-desktop", "This test covers its own phone, tablet and desktop widths.");
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.route("https://**", route => route.abort());
    for (const width of [320, 390, 768, 1280]) {
      await page.setViewportSize({ width, height: 900 });
      await page.goto(`/__project-layout/${slug}/${locale}`);
      await expect(page.locator("body")).toHaveAttribute("data-product-profile", slug);
      await expect(page.locator(".project-navigation")).toBeVisible();
      const layout = await page.evaluate(() => {
        const rect = (el: Element) => { const r = el.getBoundingClientRect(); return { x: r.x, y: r.y, right: r.right, bottom: r.bottom, width: r.width }; };
        const targets = [...document.querySelectorAll(".shell, .product-demo, .benefit-grid, .readme pre, .readme table")];
        const links = [...document.querySelectorAll(".project-navigation a")].map(rect);
        return {
          viewport: document.documentElement.clientWidth, document: document.documentElement.scrollWidth,
          escaping: targets.map(rect).filter(r => r.x < -1 || r.right > innerWidth + 1),
          overlappingLinks: links.some((a, i) => links.slice(i + 1).some(b => Math.min(a.right,b.right)-Math.max(a.x,b.x)>1 && Math.min(a.bottom,b.bottom)-Math.max(a.y,b.y)>1)),
          code: getComputedStyle(document.querySelector(".readme pre")!).whiteSpace,
        };
      });
      expect(layout.document, `${slug} ${locale} at ${width}px`).toBeLessThanOrEqual(layout.viewport);
      expect(layout.escaping).toEqual([]);
      expect(layout.overlappingLinks).toBe(false);
      expect(layout.code).toBe("pre");
      await expect(page.locator(".readme pre").first()).toContainText("│ source       │ destination  │");
      await page.locator('.project-sections a[href="#install"]').click();
      const anchor = await page.locator("#install").boundingBox();
      const nav = await page.locator(".project-navigation").boundingBox();
      expect(anchor!.y).toBeGreaterThanOrEqual(nav!.height - 1);
      expect(anchor!.y).toBeLessThan(600);
      if (!locale && (width === 390 || width === 1280)) {
        await page.evaluate(() => scrollTo(0,0));
        await page.screenshot({ path: info.outputPath(`${slug}-${width}.png`), fullPage: true });
      }
    }
  });
}
