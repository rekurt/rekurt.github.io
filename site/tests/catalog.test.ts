import { describe, expect, it } from "vitest";

import { getCatalog, getProduct, getProducts, getRegistryRepositories } from "../src/lib/catalog";

describe("catalog selectors", () => {
  it("loads the generated schema", () => {
    const catalog = getCatalog();
    expect(catalog.schemaVersion).toBe(1);
    expect(catalog.owner).toBe("rekurt");
    expect(catalog.products).toHaveLength(14);
    expect(catalog.repositories.length).toBeGreaterThan(0);
    const repositoryNames = new Set(catalog.repositories.map((repository) => repository.nameWithOwner));
    expect(repositoryNames.size).toBe(catalog.repositories.length);
    for (const product of catalog.products) {
      expect(repositoryNames.has(product.primaryRepo)).toBe(true);
    }
    expect(catalog.repositories.find((repository) => repository.nameWithOwner === "rekurt/rekurt.github.io")?.role).toBe("portfolio-hub");
  });

  it("excludes all forks from the public registry after a snapshot refresh", () => {
    const visible = getRegistryRepositories();
    const all = getCatalog().repositories;
    expect(all.some(repository => repository.fork)).toBe(true);
    expect(visible.length).toBeLessThan(all.length);
    expect(visible.every(repository => !repository.fork && !["fork", "maintained-fork"].includes(repository.role))).toBe(true);
    expect(visible.some(repository => repository.name === "awesome-backup")).toBe(false);
    expect(visible.some(repository => repository.name === "openkline")).toBe(true);
  });

  it("localizes without changing project identity", () => {
    const en = getProduct("prt", "en");
    const ru = getProduct("prt", "ru");
    const zh = getProduct("prt", "zh-cn");
    expect(en.slug).toBe(ru.slug);
    expect(en.slug).toBe(zh.slug);
    expect(en.summary).not.toBe(ru.summary);
    expect(en.summary).toBe(en.summaries.en);
    expect(ru.summary).toBe(ru.summaries.ru);
    expect(zh.summary).toBe(zh.summaries.zhCN);
  });

  it("returns detached product collections", () => {
    const first = getProducts("en");
    const second = getProducts("en");
    expect(first).not.toBe(second);
    expect(first[0]).not.toBe(second[0]);
  });

  it("rejects an unknown product", () => {
    expect(() => getProduct("missing", "en")).toThrow("Unknown product: missing");
  });
});
