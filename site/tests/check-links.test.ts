import { mkdtemp, mkdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { afterEach, describe, expect, it } from "vitest";

import { checkInternalLinks } from "../scripts/check-links.mjs";

const temporaryDirectories: string[] = [];

async function fixture(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "portfolio-links-"));
  temporaryDirectories.push(directory);
  return directory;
}

afterEach(async () => {
  await Promise.all(temporaryDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })));
});

describe("internal link checker", () => {
  it("reports a missing directory route", async () => {
    const directory = await fixture();
    await writeFile(join(directory, "index.html"), '<a href="/missing/">missing</a>');
    await expect(checkInternalLinks(directory)).rejects.toThrow("index.html: broken internal link /missing/");
  });

  it("accepts directory indexes and ignores external links", async () => {
    const directory = await fixture();
    await mkdir(join(directory, "projects"));
    await writeFile(join(directory, "index.html"), '<a href="/projects/">projects</a><a href="https://example.com">external</a>');
    await writeFile(join(directory, "projects", "index.html"), '<a href="/#top">home</a>');
    await expect(checkInternalLinks(directory)).resolves.toEqual({ checkedFiles: 2, checkedLinks: 2 });
  });
});
