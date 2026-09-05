import { describe, expect, it } from "vitest";

import { alternatePath, copy } from "../src/i18n/copy";

describe("locale routing", () => {
  it("maps alternate paths symmetrically", () => {
    expect(alternatePath("/projects/vpn-hub/", "ru")).toBe("/ru/projects/vpn-hub/");
    expect(alternatePath("/ru/projects/vpn-hub/", "zh-cn")).toBe("/zh-cn/projects/vpn-hub/");
    expect(alternatePath("/zh-cn/projects/vpn-hub/", "ru")).toBe("/ru/projects/vpn-hub/");
    expect(alternatePath("/ru/projects/vpn-hub/", "en")).toBe("/projects/vpn-hub/");
  });

  it("keeps query strings and hashes", () => {
    expect(alternatePath("/projects/?kind=cli#catalog", "ru")).toBe(
      "/ru/projects/?kind=cli#catalog",
    );
  });

  it("exposes complete non-empty copy in all locales", () => {
    const keys = Object.keys(copy.en) as Array<keyof typeof copy.en>;
    expect(keys.length).toBeGreaterThan(10);
    for (const key of keys) {
      expect(copy.en[key].trim()).not.toBe("");
      expect(copy.ru[key].trim()).not.toBe("");
      expect(copy["zh-cn"][key].trim()).not.toBe("");
    }
  });
});
