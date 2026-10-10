import { execFileSync } from "node:child_process";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../", import.meta.url));
const scratch = mkdtempSync(join(tmpdir(), "rekurt-layout-"));
try {
  const binary = join(scratch, "project-site");
  execFileSync("go", ["build", "-o", binary, "./cmd/project-site"], { cwd: root, stdio: "inherit" });
  const repository = join(scratch, "repository");
  mkdirSync(repository);
  const long = "LongRepositoryOrPackageName".repeat(12);
  const source = [
    "# Layout regression fixture", "", "Documentation with [" + long + "](https://github.com/rekurt).", "",
    "## Source and terminal output", "", "```text",
    "┌──────────────┬──────────────┐",
    "│ source       │ destination  │",
    "└──────────────┴──────────────┘",
    long, "```", "", "| Field | Source | Destination | Version | Status | Notes |",
    "| --- | --- | --- | --- | --- | --- |",
    "| " + long + " | stable | release | 1.0.0 | ready | checked |", "",
    "## TypeScript example", "", "```ts", "const count: number = 42;", "```"
  ].join("\n");
  for (const name of ["README.md", "README.ru.md", "README.zh-CN.md"]) writeFileSync(join(repository, name), source);
  const profiles = ["cortex-forge", "dbdiff", "depth", "git-barber", "gitlab-downloader", "go-propisyu", "gost-crypto", "prt", "sprint-velocity", "ymsdk"];
  for (const slug of profiles) {
    execFileSync(binary, ["build", "--slug", slug, "--snapshot", join(root, "site/src/data/generated/catalog.json"), "--repo", repository,
      "--out", resolve(root, "site/dist/__project-layout", slug), "--base-url", "https://rekurt.github.io/" + slug + "/"], { cwd: root, stdio: "inherit" });
  }
} finally {
  rmSync(scratch, { recursive: true, force: true });
}
