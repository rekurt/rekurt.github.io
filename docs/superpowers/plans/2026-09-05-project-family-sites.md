# Project Family Sites Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a centrally governed, independently deployed, multilingual website for each of the 14 curated `rekurt` open-source products.

**Architecture:** A Go family-site generator in `rekurt.github.io` builds complete sites for projects without a web surface and decorates the four existing sites after their normal builds. A reusable GitHub Pages workflow serves generated sites; existing sites call the same generator from their repository-specific workflows. The portfolio catalog remains the source of truth and all sites expose a validated reciprocal link graph.

**Tech Stack:** Go 1.27, `html/template`, `golang.org/x/net/html`, Astro 7, TypeScript 6, GitHub Actions, GitHub Pages.

**Spec:** `docs/superpowers/specs/2026-09-05-project-family-sites-design.md`

## Global Constraints

- Curated scope is exactly the 14 products in `catalog/projects.yaml`; support repositories and ordinary forks do not receive sites.
- English is the root and `x-default`; complete shared UI is available at `/ru/` and `/zh-cn/`.
- Existing `openkline`, `Mac-Coffee`, `vpn-hub`, and `chislo` application surfaces must remain functional.
- Generated HTML is a Pages artifact and is never committed to a child repository.
- Canonical URLs and external links must use HTTPS.
- Claims must be derived from the catalog or repository content; never invent adoption, performance, compatibility, or security claims.
- A deployment is complete only after its workflow succeeds and the production URL returns the expected canonical metadata.
- All commits use conventional commit messages and the repository's configured Git identity.

---

### Task 1: Extend the catalog contract to three locales and stable accents

**Files:**
- Modify: `internal/catalog/model.go`
- Modify: `internal/catalog/validate.go`
- Modify: `internal/catalog/validate_test.go`
- Modify: `internal/catalog/testdata/valid.yaml`
- Modify: `catalog/projects.yaml`
- Modify: `catalog/projects_test.go`
- Modify: `internal/sync/merge.go`
- Modify: `internal/sync/merge_test.go`

**Interfaces:**
- Produces: `LocalizedText{EN, RU, ZHCN string}` and `ProductConfig.Accent string`.
- Produces: snapshot JSON fields `summary.zhCN` and `accent` for every product.

- [ ] **Step 1: Write failing validation tests**

Add table cases asserting that a missing `summary.zh-cn` and an accent outside `cyan|emerald|violet|amber|coral` are rejected. Extend the valid manifest fixture with:

```yaml
accent: cyan
summary:
  en: Tool summary
  ru: Описание инструмента
  zh-cn: 工具简介
```

- [ ] **Step 2: Run the focused tests and confirm failure**

Run: `go test ./internal/catalog ./internal/sync`

Expected: FAIL because the schema has no Chinese or accent fields and no corresponding validation.

- [ ] **Step 3: Implement the schema and validation**

Use these exact fields:

```go
type LocalizedText struct {
    EN   string `yaml:"en" json:"en"`
    RU   string `yaml:"ru" json:"ru"`
    ZHCN string `yaml:"zh-cn" json:"zhCN"`
}

type ProductConfig struct {
    // existing fields
    Accent string `yaml:"accent" json:"accent"`
}
```

Copy `Accent` from `ProductConfig` to `Product` during snapshot construction and validate it against a closed set.

- [ ] **Step 4: Add factual Chinese summaries and accents for all products**

Assign accents by product domain while keeping adjacent cards visually distinguishable. Translate the existing summary meaning without adding claims.

- [ ] **Step 5: Run schema and sync tests**

Run: `go test ./internal/catalog ./internal/sync && git diff --check`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/catalog internal/sync catalog
git commit -m "feat(catalog): add Chinese summaries and project accents"
```

### Task 2: Build the deterministic family-site domain model

**Files:**
- Create: `internal/projectsite/model.go`
- Create: `internal/projectsite/resolve.go`
- Create: `internal/projectsite/resolve_test.go`
- Create: `internal/projectsite/testdata/repository/README.md`
- Create: `internal/projectsite/testdata/repository/README.ru.md`
- Create: `internal/projectsite/testdata/snapshot.json`

**Interfaces:**
- Consumes: `catalog.Snapshot`, repository root, slug, canonical base URL.
- Produces: `projectsite.Model` and `projectsite.LocalePage` values.

```go
type Options struct {
    Slug         string
    SnapshotPath string
    Repository   string
    Output       string
    BaseURL      string
}

type LocalePage struct {
    Locale, Lang, Path, Canonical, Title, Description string
    Summary, ReadmeHTML string
}

func Resolve(options Options) (Model, error)
```

- [ ] **Step 1: Write failing resolver tests**

Cover product lookup, English default, Russian and Chinese paths, localized README selection, fallback without fabricated localized README, current source SHA, and normalized trailing-slash canonical URL.

- [ ] **Step 2: Add boundary tests**

Reject unknown slugs, non-HTTPS base URLs, base URLs with query/fragment, missing primary repository, and repository/output paths that resolve to the same directory.

- [ ] **Step 3: Run tests and confirm failure**

Run: `go test ./internal/projectsite -run TestResolve -v`

Expected: FAIL because the package does not exist.

- [ ] **Step 4: Implement the minimal resolver**

Read the snapshot with `catalogsync.ReadSnapshot`, resolve the primary repository, map locale-specific README filenames in deterministic priority order, and render them with the existing sanitized Markdown renderer.

- [ ] **Step 5: Run resolver tests**

Run: `go test ./internal/projectsite -run TestResolve -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/projectsite
git commit -m "feat(site-kit): resolve localized project site models"
```

### Task 3: Render complete generated sites

**Files:**
- Create: `internal/projectsite/build.go`
- Create: `internal/projectsite/build_test.go`
- Create: `internal/projectsite/templates/base.html`
- Create: `internal/projectsite/templates/directory.html`
- Create: `internal/projectsite/assets/family.css`
- Create: `internal/projectsite/assets/family.js`
- Create: `internal/projectsite/seo.go`
- Create: `internal/projectsite/seo_test.go`

**Interfaces:**
- Consumes: `projectsite.Model` from Task 2.
- Produces: `func Build(options Options) (BuildManifest, error)`.
- Produces routes `/`, `/ru/`, `/zh-cn/`, locale project directories, `robots.txt`, `sitemap.xml`, `404.html`, and `assets/*`.

- [ ] **Step 1: Write failing golden-structure tests**

Assert that all locale pages contain one `h1`, unique title/description, canonical and reciprocal alternates, source/install/version data, author link, sibling directory link, and `SoftwareSourceCode` JSON-LD.

- [ ] **Step 2: Write failing SEO and filesystem tests**

Assert that sitemap URLs stay under `BaseURL`, every file stays under `Output`, JavaScript is optional, visible text is escaped, README HTML remains sanitized, and a second build is byte-identical except for an explicitly supplied timestamp.

- [ ] **Step 3: Run tests and confirm failure**

Run: `go test ./internal/projectsite -run 'TestBuild|TestSEO' -v`

Expected: FAIL because renderer files do not exist.

- [ ] **Step 4: Implement the templates and shared assets**

Use `html/template` for documents. Render three layout variants selected by `kind`: `terminal` for CLI/skill, `code` for library/SDK, and `systems` for platform/desktop. Use the catalog accent only through validated CSS custom properties.

- [ ] **Step 5: Implement SEO outputs**

Generate JSON-LD with `name`, localized `description`, `codeRepository`, `programmingLanguage`, `license`, `version`, and `dateModified` when present. Generate reciprocal sitemap alternates for all three locales.

- [ ] **Step 6: Run build tests and inspect a fixture artifact**

Run: `go test ./internal/projectsite -run 'TestBuild|TestSEO' -v`

Expected: PASS and no writes outside the test temporary directory.

- [ ] **Step 7: Commit**

```bash
git add internal/projectsite
git commit -m "feat(site-kit): render multilingual project sites"
```

### Task 4: Decorate existing static sites without replacing them

**Files:**
- Create: `internal/projectsite/decorate.go`
- Create: `internal/projectsite/decorate_test.go`
- Create: `internal/projectsite/testdata/existing/index.html`

**Interfaces:**
- Consumes: a completed static output directory and `projectsite.Model`.
- Produces: `func Decorate(options Options) (BuildManifest, error)`.

- [ ] **Step 1: Write failing preservation tests**

Hash the original body subtree and verify that decoration preserves application scripts, forms, canvas elements, relative assets, and existing social images while adding only namespaced family elements and metadata.

- [ ] **Step 2: Write idempotence and collision tests**

Run decoration twice and assert one family bar, one stylesheet, one canonical link, and stable output. Reject missing `index.html`, symlink escapes, malformed HTML, and output outside the workspace.

- [ ] **Step 3: Run tests and confirm failure**

Run: `go test ./internal/projectsite -run TestDecorate -v`

Expected: FAIL because `Decorate` does not exist.

- [ ] **Step 4: Implement DOM-based decoration**

Parse with `golang.org/x/net/html`. Add a `data-rekurt-family` element as the first body child, use `rk-family-*` classes, keep all existing nodes, and write through an atomic temporary file. Generate localized family directories beside the existing application.

- [ ] **Step 5: Run decorator and full package tests**

Run: `go test ./internal/projectsite -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/projectsite
git commit -m "feat(site-kit): decorate existing project sites safely"
```

### Task 5: Add the CLI, artifact validator, and reusable Pages workflow

**Files:**
- Create: `cmd/project-site/main.go`
- Create: `cmd/project-site/main_test.go`
- Create: `.github/workflows/project-pages.yml`
- Modify: `Makefile`
- Modify: `README.md`
- Modify: `README.ru.md`
- Modify: `README.zh-CN.md`

**Interfaces:**
- Produces commands:

```text
project-site build --slug <slug> --snapshot <file> --repo <dir> --out <dir> --base-url <https-url>
project-site decorate --slug <slug> --snapshot <file> --repo <dir> --out <dir> --base-url <https-url>
project-site validate --slug <slug> --snapshot <file> --out <dir> --base-url <https-url>
```

- Produces reusable workflow input `slug` and output `page_url`.

- [ ] **Step 1: Write failing command tests**

Cover help, unknown command, required flags, all three successful modes, non-zero validation failures, and concise stderr without secrets or absolute runner paths.

- [ ] **Step 2: Run tests and confirm failure**

Run: `go test ./cmd/project-site -v`

Expected: FAIL because the command does not exist.

- [ ] **Step 3: Implement the command and Make targets**

Add `site-kit-test` and include it in `make test`. Keep command parsing in a testable `run(args, stdout, stderr)` function.

- [ ] **Step 4: Create the reusable generated-site workflow**

The called workflow must check out the caller as `project`, check out `rekurt/rekurt.github.io@main` as `family`, refresh the catalog using the caller token, build and validate into `_site`, upload the artifact, and deploy with Pages and OIDC permissions.

- [ ] **Step 5: Document onboarding**

Document the three-command local flow and the exact thin caller workflow in English, Russian, and Simplified Chinese.

- [ ] **Step 6: Verify locally**

Run: `make check && make test && make build && git diff --check`

Expected: PASS.

- [ ] **Step 7: Commit and push the central kit**

```bash
git add cmd/project-site internal/projectsite .github/workflows/project-pages.yml Makefile README.md README.ru.md README.zh-CN.md
git commit -m "feat(site-kit): publish reusable project pages pipeline"
git push origin main
```

Wait for the central CI and Pages workflows to succeed before child repositories reference `@main`.

### Task 6: Add Simplified Chinese to the author hub

**Files:**
- Modify: `site/src/lib/catalog.ts`
- Modify: `site/src/lib/routes.ts`
- Modify: `site/src/i18n/copy.ts`
- Modify: `site/src/layouts/BaseLayout.astro`
- Modify: `site/src/components/Header.astro`
- Modify: `site/src/components/Footer.astro`
- Create: `site/src/pages/zh-cn/index.astro`
- Create: `site/src/pages/zh-cn/projects/index.astro`
- Create: `site/src/pages/zh-cn/projects/[slug].astro`
- Create: `site/src/pages/zh-cn/registry/index.astro`
- Create: `site/src/pages/zh-cn/about/index.astro`
- Modify: `site/src/pages/sitemap.xml.ts`
- Modify: `site/tests/catalog.test.ts`
- Modify: `site/tests/i18n.test.ts`
- Modify: `site/tests/pages.test.ts`
- Modify: `site/tests/check-links.test.ts`
- Modify: `site/tests/e2e/navigation.spec.ts`

**Interfaces:**
- Extends `Locale` to `"en" | "ru" | "zh-cn"`.
- Produces reciprocal hub routes and alternates for all 14 product detail pages.

- [ ] **Step 1: Update tests first**

Require Chinese routes in path inventory, localized summaries on cards/details, `zh-CN` HTML language, all three alternates, and a three-language navigation switch.

- [ ] **Step 2: Run frontend tests and confirm failure**

Run: `cd site && npm test`

Expected: FAIL on missing `zh-cn` locale and routes.

- [ ] **Step 3: Implement locale routing and copy**

Add complete Simplified Chinese equivalents for every `CopyKey`. Replace Russian-only path logic with a locale-prefix map so English remains unprefixed.

- [ ] **Step 4: Build and validate the hub**

Run: `cd site && npm test && npm run check && npm run build && npm run check:links`

Expected: PASS with English, Russian, and Chinese page sets.

- [ ] **Step 5: Commit and push**

```bash
git add site
git commit -m "feat(site): add Simplified Chinese portfolio routes"
git push origin main
```

### Task 7: Canary generated site in `git-barber`

**Files in `work/repos/git-barber`:**
- Create: `.github/workflows/pages.yml`
- Modify: `README.md`

**Interfaces:**
- Consumes: `rekurt/rekurt.github.io/.github/workflows/project-pages.yml@main` with `slug: git-barber`.

- [ ] **Step 1: Add the caller workflow**

Trigger on default-branch push, `v*` tags, releases, manual dispatch, and `17 */6 * * *`. Grant only `contents: read`, `pages: write`, and `id-token: write`.

- [ ] **Step 2: Add the verified website link to README**

Link `https://rekurt.github.io/git-barber/` in the existing project links area without changing product documentation.

- [ ] **Step 3: Validate repository changes**

Run: `git -C work/repos/git-barber diff --check && gh workflow view pages.yml --repo rekurt/git-barber --yaml` after push.

- [ ] **Step 4: Commit, push, and enable workflow Pages**

```bash
git -C work/repos/git-barber add .github/workflows/pages.yml README.md
git -C work/repos/git-barber commit -m "feat(site): publish project website"
git -C work/repos/git-barber push origin master
```

Set Pages build type to `workflow`, wait for the deployment, then require HTTP 200 and the correct canonical/title/JSON-LD/sitemap.

- [ ] **Step 5: Stop on canary regression**

Do not roll out other child sites until both the workflow and production artifact pass the central validator.

### Task 8: Canary decoration in `Mac-Coffee`

**Files in `work/repos/Mac-Coffee`:**
- Modify: `.github/workflows/pages.yml`
- Modify: `site/tests/validate-site.mjs`

**Interfaces:**
- Consumes: `project-site decorate --slug mac-coffee` after the existing static-site validation.

- [ ] **Step 1: Extend validation first**

Require the post-build family bar, author hub link, local project directory, canonical URL, and preservation of the existing Open Graph image.

- [ ] **Step 2: Update the Pages workflow**

Copy `site` to `_site`, refresh the family catalog, decorate `_site`, validate the decorated artifact, and deploy `_site`. Add the staggered schedule without removing existing path triggers.

- [ ] **Step 3: Run the local site validator against the decorated fixture**

Run the repository validator plus the central `project-site validate` command.

- [ ] **Step 4: Commit, push, and verify production**

```bash
git -C work/repos/Mac-Coffee add .github/workflows/pages.yml site/tests/validate-site.mjs
git -C work/repos/Mac-Coffee commit -m "feat(site): join the project family navigation"
git -C work/repos/Mac-Coffee push origin main
```

Require the existing site interaction and social metadata to remain intact.

### Task 9: Roll out generated sites to the remaining nine repositories

**Files in each repository:**
- Create: `.github/workflows/pages.yml`
- Modify: primary README only when it has an existing project-links area suitable for the verified website URL.

**Repositories and slugs:**

```text
ymsdk=ymsdk
gost-crypto=gost-crypto
cortex-forge=cortex-forge
dbdiff=dbdiff
depth=depth
gitlab-downloader=gitlab-downloader
go-propisyu=go-propisyu
prt=prt
sprint-velocity=sprint-velocity
```

- [ ] **Step 1: Add thin caller workflows**

Use staggered schedule minutes derived from list order so the ten generated deployments do not start simultaneously. Preserve each default branch and existing CI workflows.

- [ ] **Step 2: Validate every diff and configured identity**

For each checkout run `git diff --check`, inspect `git status --short`, and confirm `git config user.name` and `user.email` before commit.

- [ ] **Step 3: Commit and push each repository**

Use `feat(site): publish project website` and push only its current default branch.

- [ ] **Step 4: Enable Pages and wait for all deployments**

Set `build_type=workflow` through the GitHub Pages API. Treat each site independently: capture workflow URL, conclusion, deployment URL, and production response.

- [ ] **Step 5: Validate the generated fleet**

For each site verify root, `/ru/`, `/zh-cn/`, `/projects/`, `robots.txt`, and `sitemap.xml`. Validate canonical, alternates, JSON-LD, hub link, and absence of mixed-content URLs.

### Task 10: Decorate the remaining existing sites

**Files:**
- Modify: `work/repos/openkline/.github/workflows/pages.yml`
- Remove: `work/repos/openkline/.github/workflows/jekyll-gh-pages.yml`
- Modify: `work/repos/openkline` site validation or add a focused artifact validator
- Modify: `work/repos/vpn-hub/.github/workflows/pages.yml`
- Modify: `work/repos/vpn-hub/site/scripts/test-built-site.mjs`
- Modify: `work/repos/chislo/.github/workflows/pages.yml`
- Add a focused decorated-artifact check to `work/repos/chislo`

**Interfaces:**
- All workflows invoke the same `project-site decorate` and `project-site validate` commands after their current application build.

- [ ] **Step 1: Update tests before workflows**

For each repository add assertions for the family bar and sibling directory while retaining all existing product-specific assertions.

- [ ] **Step 2: Integrate decoration into each artifact assembly**

Keep the current playground, TypeDoc, Astro documentation, WASM, and rustdoc build commands unchanged. Remove only the obsolete competing OpenKline Jekyll Pages workflow because two workflows must not race to deploy different artifacts.

- [ ] **Step 3: Run available local validation**

Run OpenKline package/site tests, vpn-hub content/build tests, and chislo HTML validator. If a platform toolchain is unavailable locally, rely on the unchanged repository CI plus the matching GitHub runner before claiming success.

- [ ] **Step 4: Commit and push each repository**

Use `feat(site): join the project family navigation` on the current default branch.

- [ ] **Step 5: Wait for CI and Pages**

Require both the repository CI and Pages workflow to succeed. Check that each existing interactive root still contains its original primary control or content marker.

### Task 11: Close the reciprocal graph and record production evidence

**Files in the central repository:**
- Modify: `catalog/projects.yaml`
- Modify: `site/src/data/generated/catalog.json`
- Modify: `docs/repository-audit.md`
- Create: `docs/project-sites-verification.md`
- Modify: `README.md`
- Modify: `README.ru.md`
- Modify: `README.zh-CN.md`

**Interfaces:**
- Produces: final central catalog with 14 verified website URLs.
- Produces: evidence table mapping product, repository commit, workflow run, Pages URL, status, locales, and SEO checks.

- [ ] **Step 1: Update GitHub repository homepages only for verified URLs**

Patch each primary repository homepage to its successful Pages URL. Preserve `openkline`'s canonical product domain and existing external package documentation links.

- [ ] **Step 2: Synchronize the central catalog**

Run:

```bash
GITHUB_TOKEN="$(gh auth token)" go run ./cmd/catalog-sync sync
make check
make test
make build
(cd site && npm run check:links)
```

Expected: 14 curated products and 14 verified product websites.

- [ ] **Step 3: Run fleet-level HTTP validation**

Add and run a deterministic checker that follows no arbitrary redirects and verifies only catalog-owned HTTPS URLs. Require every production root and sitemap to return 200 and every site directory to contain the other 13 project URLs plus the author hub.

- [ ] **Step 4: Write the evidence report**

Record exact Git SHAs and GitHub Actions run URLs. Do not infer deployment from a build or commit.

- [ ] **Step 5: Commit, push, and wait for the final portfolio deploy**

```bash
git add catalog site/src/data/generated/catalog.json docs README.md README.ru.md README.zh-CN.md
git commit -m "feat(catalog): link the complete project site family"
git push origin main
```

Wait for central CI, catalog sync, and Pages deployment. Re-run production validation against the post-deploy catalog.

- [ ] **Step 6: Final clean-state audit**

For the central repo and all 14 child checkouts, require a clean tracked worktree, correct tracking branch, and pushed HEAD. Preserve local `work/` clones as untracked task intermediates unless explicitly removed.
