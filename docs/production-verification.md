# Production verification

Verified on 2026-09-05 (Europe/Moscow) against commit [`0abfaaa`](https://github.com/rekurt/rekurt.github.io/commit/0abfaaa13a36e9472889d6fab2f87b380da3c7f2).

## Production

| Requirement | Result | Evidence |
|---|---|---|
| Public repository | PASS | [`rekurt/rekurt.github.io`](https://github.com/rekurt/rekurt.github.io) |
| Static production site | PASS | [`https://rekurt.github.io/`](https://rekurt.github.io/) returns HTTP 200 |
| Pages workflow mode | PASS | GitHub Pages API: `build_type=workflow`, `status=built`, `https_enforced=true` |
| Curated products | PASS | 14 validated products from `catalog/projects.yaml` |
| Complete public registry | PASS | 50 public repositories: 20 original, 30 forks, 0 private |
| Self-discovery | PASS | `rekurt/rekurt.github.io` is present with role `portfolio-hub` |
| Bilingual routes | PASS | 36 indexed URLs: English default and Russian under `/ru/`; `/ru/` renders `lang="ru"` |
| Maintained-fork attribution | PASS | [`Mac-Coffee`](https://rekurt.github.io/projects/mac-coffee/) names and links `Elliotwu-7/Mac-Coffee` |
| Full fork separation | PASS | No repository with role `fork` has an author `website` link |

The complete per-repository answer to “has a website / has no website” is maintained in `docs/repository-audit.md` and rendered at [`/registry/`](https://rekurt.github.io/registry/). Seven current registry entries have an author website: the six source projects below plus the portfolio hub; the other 43 do not.

## Project websites and package documentation

Direct HTTP verification followed redirects.

| Surface | Result |
|---|---|
| [`chislo`](https://rekurt.github.io/chislo/) | HTTP 200 |
| [`maccoffee-dist`](https://rekurt.github.io/maccoffee-dist/) | HTTP 200 |
| [`openkline`](https://rekurt.github.io/openkline/) | HTTP 200 |
| [`openkline.tech`](https://openkline.tech/) | HTTP 200 |
| [`vpn-hub`](https://rekurt.github.io/vpn-hub/) | HTTP 200 |
| [`Mac-Coffee`](https://rekurt.github.io/Mac-Coffee/) | HTTP 200 |
| [`go-propisyu` API](https://pkg.go.dev/github.com/rekurt/go-propisyu) | HTTP 200 |
| [`gost-crypto` API](https://pkg.go.dev/github.com/rekurt/gost-crypto) | HTTP 200 |
| [`ymsdk` API](https://pkg.go.dev/github.com/rekurt/ymsdk) | HTTP 200 |
| [`prt` crate](https://crates.io/crates/prt) | Link published; automated client receives HTTP 403, so no availability claim is made |

## Automation evidence

All workflows for the verified source revision completed successfully:

- [CI run 33931446636](https://github.com/rekurt/rekurt.github.io/actions/runs/33931446636): Go format/vet/race tests, 14 Vitest tests, Astro type checks, production build, 346 internal references, and 28 Playwright scenarios.
- [Deploy Pages run 33931446765](https://github.com/rekurt/rekurt.github.io/actions/runs/33931446765): static artifact uploaded and deployed.
- [Sync catalog run 33931446529](https://github.com/rekurt/rekurt.github.io/actions/runs/33931446529): live GitHub sync and complete verification succeeded without creating a self-referential follow-up commit.
- [Release Please run 33931446695](https://github.com/rekurt/rekurt.github.io/actions/runs/33931446695): release automation succeeded and maintains [release PR #5](https://github.com/rekurt/rekurt.github.io/pull/5).

The scheduled sync runs at minute 17 of every hour. A new public repository appears in the complete registry automatically. Adding one reviewed record to `catalog/projects.yaml` promotes it into the product catalog. Child repositories require no token, webhook, or workflow changes.

## Route and content checks

The following production paths returned HTTP 200:

```text
https://rekurt.github.io/
https://rekurt.github.io/projects/
https://rekurt.github.io/registry/
https://rekurt.github.io/ru/
https://rekurt.github.io/projects/mac-coffee/
```

The generated sitemap contains 36 localized content URLs with English/Russian alternate links. The build also emits `robots.txt` and `404.html`.

## Security and integrity invariants

| Invariant | Result | Verification |
|---|---|---|
| Public data only | PASS | Snapshot query returned zero records where `visibility != public` |
| Safe upstream attribution | PASS | Simple forks have no author website actions; maintained fork has explicit upstream |
| Sanitized repository docs | PASS | Zero README fragments contain `script`, `iframe`, `form`, `style`, `svg`, or `javascript:` |
| Pinned relative README URLs | PASS | Links and images are rewritten to commit-specific GitHub and raw-content URLs |
| Keyboard-accessible overflow | PASS | Generated README `pre` and `table` regions receive `tabindex=0` |
| No runtime GitHub access | PASS | Site source contains no `fetch`, `XMLHttpRequest`, or GitHub API call outside generated data |
| No credential material | PASS | Repository scan found no GitHub PAT or private-key pattern |
| Deterministic sync | PASS | Snapshot comparison ignores only timestamp and the hub's self-referential commit fields; material project changes remain versioned |
| Accessibility | PASS | Axe found zero serious or critical violations on representative EN/RU, catalog, detail, and registry routes in desktop and mobile Chromium |
| Responsive layout | PASS | Home, projects, and registry had no horizontal document overflow at 1440×1000 and 390×844 |

## Reproduction commands

```bash
make check
make test
make build
cd site
npm run check:links
npm run test:e2e
```

Live catalog verification:

```bash
GITHUB_TOKEN="$(gh auth token)" go run ./cmd/catalog-sync check \
  --manifest catalog/projects.yaml \
  --snapshot site/src/data/generated/catalog.json \
  --audit docs/repository-audit.md
```

Expected result: `catalog is current: 14 products, 50 repositories`.
