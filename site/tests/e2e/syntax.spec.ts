import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";

const asset = (name: string) => readFileSync(new URL("../../public/syntax/" + name, import.meta.url), "utf8");
const familyScript = readFileSync(new URL("../../../internal/projectsite/assets/family.js", import.meta.url), "utf8");

test("portfolio installation and README code are highlighted", async ({ page }) => {
  await page.goto("/projects/depth/");
  const install = page.locator(".install-list code").first();
  await expect(install).toHaveAttribute("data-syntax-highlighted", "cli");
  await expect(install.locator(".hljs-title")).toContainText("npm");
  const source = page.locator(".readme code.language-ts, .readme code.language-typescript").first();
  await expect(source).toHaveClass(/hljs/);
  await expect(source.locator(".hljs-keyword").first()).toBeVisible();
});

for (const theme of ["theme-liquidity", "theme-words"]) {
  test("syntax preserves copy, escaping and layout in " + theme, async ({ page }) => {
    const command = "npm install @rekurt/depth --save-dev";
    const code = 'const markup: string = "<img src=x onerror=alert(1)>";';
    await page.goto("/");
    await page.setContent('<main class="' + theme + '"><div class="command"><code data-syntax-language="cli"></code><button data-copy="' + command + '" data-copy-label="Copy"><span>Copy</span></button></div><p class="copy-status" data-copied="Copied" data-copy-failed="Copy manually"></p><pre><code class="language-typescript"></code></pre><pre><code data-syntax-language="plaintext">Friday report: 8 tasks</code></pre></main>');
    await page.locator(".command code").evaluate((el, value) => { el.textContent = value; }, command);
    await page.locator("code.language-typescript").evaluate((el, value) => { el.textContent = value; }, code);
    await page.addStyleTag({ content: "body { margin: 0; padding: 16px; background: " + (theme === "theme-words" ? "#fffaf0; color: #362c21;" : "#101b24; color: #e9f3f7;") + " } code { white-space: pre-wrap; overflow-wrap: anywhere; } pre { max-width: 100%; } " + asset("syntax.css") });
    await page.addScriptTag({ content: asset("highlight-11.12.0.min.js") });
    await page.addScriptTag({ content: asset("syntax.js") });
    await expect(page.locator(".command code")).toHaveText(command);
    await expect(page.locator("code.language-typescript")).toHaveText(code);
    await expect(page.locator("code.language-typescript img")).toHaveCount(0);
    await expect(page.locator('[data-syntax-language="plaintext"]')).not.toHaveClass(/hljs/);
    await page.evaluate(() => {
      Object.defineProperty(navigator, "clipboard", { configurable: true, value: { writeText: async (text: string) => {
        (window as unknown as { copied: string }).copied = text;
      } } });
    });
    await page.addScriptTag({ content: familyScript });
    await page.getByRole("button", { name: "Copy", exact: true }).click();
    await expect(page.locator(".copy-status")).toHaveText("Copied");
    expect(await page.evaluate(() => (window as unknown as { copied: string }).copied)).toBe(command);
    const width = await page.evaluate(() => ({ viewport: innerWidth, document: document.documentElement.scrollWidth }));
    expect(width.document).toBeLessThanOrEqual(width.viewport);
    const before = await page.locator("code.language-typescript").innerHTML();
    await page.addScriptTag({ content: asset("syntax.js") });
    expect(await page.locator("code.language-typescript").innerHTML()).toBe(before);
  });
}
