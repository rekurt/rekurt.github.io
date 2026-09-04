import { describe, expect, it } from "vitest";

import { getCatalog, getProduct, getProducts } from "../src/lib/catalog";

describe("catalog selectors", () => {
  it("loads the generated schema", () => {
    const catalog = getCatalog();
    expect(catalog.schemaVersion).toBe(1);
    expect(catalog.owner).toBe("rekurt");
    expect(catalog.products).toHaveLength(14);
    expect(catalog.repositories).toHaveLength(49);
  });

  it("localizes without changing project identity", () => {
    const en = getProduct("vpn-hub", "en");
    const ru = getProduct("vpn-hub", "ru");
    expect(en.slug).toBe(ru.slug);
    expect(en.summary).not.toBe(ru.summary);
    expect(en.summary).toBe(en.summaries.en);
    expect(ru.summary).toBe(ru.summaries.ru);
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
