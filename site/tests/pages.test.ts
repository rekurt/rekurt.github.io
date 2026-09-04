import { describe, expect, it } from "vitest";

import { getCatalog, getProduct } from "../src/lib/catalog";
import { allSitePaths, escapeXml, productPaths } from "../src/lib/routes";

describe("static route contracts", () => {
  it("builds all localized product paths", () => {
    const en = productPaths("en");
    const ru = productPaths("ru");
    expect(en).toHaveLength(14);
    expect(ru).toHaveLength(14);
    expect(en.map((path) => path.params.slug)).toEqual(ru.map((path) => path.params.slug));
    expect(en.every((path) => path.props.locale === "en")).toBe(true);
    expect(ru.every((path) => path.props.locale === "ru")).toBe(true);
  });

  it("creates a unique bilingual route set", () => {
    const paths = allSitePaths();
    expect(new Set(paths).size).toBe(paths.length);
    expect(paths).toContain("/projects/vpn-hub/");
    expect(paths).toContain("/ru/projects/vpn-hub/");
    expect(paths).toHaveLength(38);
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
