import { expect, test } from "@playwright/test";
import { author } from "../../src/lib/author";
import { copy } from "../../src/i18n/copy";

for (const locale of ["en", "ru", "zh-cn"] as const) {
  test(`approved author introduction and avatar: ${locale}`, async ({ page, request }, info) => {
    test.skip(info.project.name !== "chromium-desktop", "The test covers all four viewport widths.");
    const image = await request.get(author.avatarUrl);
    expect(image.ok()).toBe(true);
    expect(image.headers()["content-type"]).toMatch(/^image\//);
    let releaseImage: () => void = () => {};
    let ready = Promise.resolve();
    await page.route(author.avatarUrl, async route => { await ready; await route.fulfill({ response: image }); });
    const prefix = locale === "en" ? "" : "/" + locale;
    for (const width of [320,390,768,1280]) for (const path of ["/", "/projects/", "/about/"]) {
      ready = new Promise<void>(resolve => { releaseImage = resolve; });
      await page.setViewportSize({ width, height: 900 });
      await page.goto(prefix + path, { waitUntil: "domcontentloaded" });
      const avatar = page.locator(".author-avatar");
      await expect(avatar).toHaveAttribute("src", author.avatarUrl);
      await expect(avatar).toHaveAttribute("alt", copy[locale].authorAvatarAlt);
      await expect(avatar).toHaveAttribute("width", "96");
      await expect(avatar).toHaveAttribute("height", "96");
      const before = await avatar.boundingBox();
      const copyBefore = await page.locator(".author-copy").boundingBox();
      expect(before?.width).toBe(96);
      expect(before?.height).toBe(96);
      releaseImage();
      await expect.poll(() => avatar.evaluate((el: HTMLImageElement) => el.complete && el.naturalWidth > 0)).toBe(true);
      expect(await avatar.boundingBox()).toEqual(before);
      expect(await page.locator(".author-copy").boundingBox()).toEqual(copyBefore);
      await expect(page.locator(".author-hobby")).toHaveText(copy[locale].authorHobby);
      const sizes = await page.evaluate(() => ({viewport: document.documentElement.clientWidth, document: document.documentElement.scrollWidth}));
      expect(sizes.document).toBeLessThanOrEqual(sizes.viewport);
      if (path === "/projects/") {
        const names = await page.locator(".project-card h3 a").allTextContents();
        expect(names.slice(0,7)).toEqual(["OpenKline","Depth","matching-engine","dbdiff","GitLab Dump","ymsdk","prt"]);
        await expect(page.locator("[data-project-card]")).toHaveCount(14);
        if (locale === "en" && (width === 390 || width === 1280)) await page.screenshot({ path: info.outputPath(`catalog-author-${width}.png`), fullPage: true });
      }
    }
  });
}
