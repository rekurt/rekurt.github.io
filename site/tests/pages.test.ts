import { describe, expect, it } from "vitest";

import { getCatalog, getProduct } from "../src/lib/catalog";
import { allSitePaths, escapeXml, productPaths } from "../src/lib/routes";

describe("static route contracts", () => {
  it("builds all localized product paths", () => {
    const en = productPaths("en");
    const ru = productPaths("ru");
    const zh = productPaths("zh-cn");
    expect(en).toHaveLength(13);
    expect(ru).toHaveLength(13);
    expect(zh).toHaveLength(13);
    expect(en.map((path) => path.params.slug)).toEqual(ru.map((path) => path.params.slug));
    expect(en.map((path) => path.params.slug)).toEqual(zh.map((path) => path.params.slug));
    expect(en.every((path) => path.props.locale === "en")).toBe(true);
    expect(ru.every((path) => path.props.locale === "ru")).toBe(true);
    expect(zh.every((path) => path.props.locale === "zh-cn")).toBe(true);
  });

  it("creates a unique trilingual route set", () => {
    const paths = allSitePaths();
    expect(new Set(paths).size).toBe(paths.length);
    expect(paths).toContain("/projects/prt/");
    expect(paths).toContain("/ru/projects/prt/");
    expect(paths).toContain("/zh-cn/projects/prt/");
    expect(paths).toHaveLength(53);
  });

  it("escapes sitemap XML values", () => {
    expect(escapeXml(`https://example.test/?a=1&b=<tag>\"`)).toBe(
      "https://example.test/?a=1&amp;b=&lt;tag&gt;&quot;",
    );
  });

  it("never promotes a simple fork homepage to an author website", () => {
    const catalog = getCatalog();
    const tsql = catalog.repositories.find((repository) => repository.nameWithOwner === "rekurt/tsql");
    expect(tsql?.role).toBe("fork");
    expect(tsql?.links.some((link) => link.kind === "website")).toBe(false);
  });

  it("preserves maintained-fork attribution", () => {
    const product = getProduct("mac-coffee", "en");
    expect(product.maintainedFork).toBe(true);
    expect(product.upstream).toBe("Elliotwu-7/Mac-Coffee");
  });
});
