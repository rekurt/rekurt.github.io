# Project family sites design

**Date:** 2026-09-05  
**Status:** approved for autonomous implementation  
**Owner:** `rekurt`

## Goal

Give every significant open-source product in the `rekurt` GitHub account a real, independently deployable website. The sites must form one recognizable family, link to one another, expose accurate project and release metadata, and remain current without copying a frontend codebase into every repository.

The existing portfolio at `https://rekurt.github.io/` remains the canonical author and repository index. Each product site is the canonical public surface for that product.

## Scope

The curated product set is the 14 products already classified in `catalog/projects.yaml`:

| Product | Current public surface | Delivery mode |
|---|---|---|
| `openkline` | interactive playground and TypeDoc | preserve and decorate |
| `Mac-Coffee` | product site | preserve and decorate |
| `vpn-hub` | multilingual documentation site | preserve and decorate |
| `chislo` | WASM playground and rustdoc | preserve and decorate |
| `git-barber` | none | generate full site |
| `ymsdk` | package documentation only | generate full site |
| `gost-crypto` | package documentation only | generate full site |
| `cortex-forge` | none | generate full site |
| `dbdiff` | none | generate full site |
| `depth` | repository README only | generate full site |
| `gitlab-downloader` | none | generate full site |
| `go-propisyu` | package documentation only | generate full site |
| `prt` | package listing only | generate full site |
| `sprint-velocity` | none | generate full site |

Support repositories, ordinary forks, package registries, and distribution taps remain linked from their parent product. They do not receive separate sites.

## Options considered

### Independent site in every repository

This offers maximum product-specific freedom, but duplicates layout, localization, SEO rules, and tests fourteen times. A design change would require fourteen coordinated migrations. Rejected because long-term drift conflicts with the centralization requirement.

### Portfolio-only project pages

This is the simplest operational model and already exists. It does not provide independent project URLs or let a project site become the canonical search result. Rejected because it does not satisfy the standalone-site goal.

### Central family kit with repository-owned deployments

The central repository owns the generator, design tokens, localized interface copy, SEO rules, and link-graph validation. Each project repository owns a small Pages workflow and deploys its own artifact. Existing applications keep their product-specific interface and receive a generated family layer after their normal build. Selected because it combines independent sites, centralized governance, and preservation of working demos.

## Architecture

### Source of truth

`catalog/projects.yaml` is extended with:

- a Simplified Chinese summary for every product;
- a stable accent identifier used by the shared design system;
- optional product-specific SEO terms;
- the public website URL once deployment is verified.

The generated catalog snapshot continues to hold live GitHub metadata: primary language, license, latest release or tag, last push, topics, repositories, and sanitized README content.

### Family-site generator

A Go command in the portfolio repository exposes two modes:

1. `build` creates a complete static product site for a repository with no existing web surface.
2. `decorate` augments an already-built static site without replacing its application or documentation.

Both modes receive a product slug, catalog snapshot, repository checkout, canonical base URL, and output directory. They produce deterministic static assets and fail closed when the slug, repository, canonical URL, or output boundary is invalid.

The generator owns:

- the shared family bar and footer;
- the sibling-project directory;
- English, Russian, and Simplified Chinese UI copy;
- design tokens and responsive styles;
- canonical, `hreflang`, Open Graph, Twitter, robots, sitemap, and JSON-LD metadata;
- the current project version, license, languages, source, documentation, and install commands;
- a build manifest containing the catalog schema, source commit, and generated timestamp.

Generated sites use the English root route, `/ru/`, and `/zh-cn/`. English is the default and `x-default`. Existing sites keep their current routing; the decorator adds canonical metadata to their root and places the localized family directory under `/projects/`, `/ru/projects/`, and `/zh-cn/projects/` so it never collides with the product application.

### Existing-site preservation

The decorator parses built HTML with an HTML parser, not regular-expression replacement. It injects a namespaced family bar and stylesheet and never rewrites application markup, scripts, or content.

Per repository:

- `openkline`: run the current library, playground, and TypeDoc builds, then decorate the assembled `_site` artifact;
- `Mac-Coffee`: validate the existing static product site, copy it to the artifact directory, then decorate it;
- `vpn-hub`: run the existing Astro content checks and build, then decorate `site/dist`;
- `chislo`: run the existing rustdoc and WASM build, then decorate the assembled artifact.

If a product build fails, deployment stops and the last successful Pages version stays live.

### Repository-owned deployment

Every primary repository contains a Pages workflow with:

- `push` on its default branch;
- release publication and version-tag triggers where supported;
- manual dispatch;
- a staggered scheduled rebuild to pick up central catalog and design changes;
- read-only checkout of the project and central family-kit repository;
- least-privilege Pages deployment permissions;
- concurrency control that does not cancel an in-progress production deployment.

The project checkout supplies the newest local README and docs. The central checkout supplies the catalog and generator. Project pushes therefore refresh documentation immediately; central design/catalog changes propagate on the scheduled rebuild without a cross-repository personal token.

The website URL is written to the GitHub repository homepage only after a successful production deployment. The portfolio catalog is synchronized again only after all project URLs are verified.

## Experience and design system

The visual thesis is an engineering constellation: dark graphite surfaces, a restrained cyan/emerald system palette, project-specific accent signals, mono status data, and clear technical typography. The shared layer should feel authored and operational rather than like a generic template.

Full generated sites expose the primary useful information in the first viewport:

- exact project purpose;
- current version and status;
- install or start command;
- direct links to source, documentation, and release.

The remainder contains a concise capability view, repository provenance, synchronized README documentation on the English route, and the family directory. Platforms and desktop tools use an architecture-oriented silhouette; libraries and SDKs use an API/code silhouette; CLI tools use a terminal-oriented silhouette. These are layout variants over one token system, not independent themes.

All navigation and controls meet keyboard, focus, contrast, touch-target, reduced-motion, and 200% text enlargement requirements. No discretionary images are needed for the generated technical sites. Existing product screenshots and interactive demos are preserved.

## Localization

English is the canonical default. Russian and Simplified Chinese are complete for shared UI, summaries, calls to action, status labels, and project-family navigation.

Repository documentation is not machine-translated. When a repository has an explicit localized README, the generator uses it. Otherwise it presents the localized project summary and links to the authoritative English README instead of pretending that an untranslated document is localized.

## SEO contract

Every generated locale page must have:

- one unique, product-specific title and description;
- a self-referencing canonical URL;
- reciprocal `hreflang` links for `en`, `ru`, `zh-CN`, and `x-default`;
- `og:type`, URL, locale, title, and description;
- Twitter card title and description;
- `SoftwareSourceCode` JSON-LD, augmented with `SoftwareApplication` only when the product is an end-user application;
- visible version, license, language, and source provenance where the data exists;
- an entry in `sitemap.xml` with locale alternates;
- a crawlable sibling-project directory and a link to the author hub.

Existing roots receive accurate canonical and structured metadata without fabricated locale alternates. Generated family directories carry their own locale alternates.

Descriptions are factual and product-specific. The system must not invent adoption, performance, security, compatibility, or production-readiness claims that are absent from repository evidence.

## Link graph

Each site links to:

- `https://rekurt.github.io/` as the author hub;
- its source repository;
- its authoritative docs and release when present;
- all other curated product sites through the local family directory.

The portfolio links back to every verified product site. Automated validation treats missing canonical sites, self-links in sibling lists, duplicate URLs, and orphaned products as build failures.

## Version and documentation flow

1. A project push or release starts its Pages workflow.
2. The workflow checks out the current project and the latest family kit.
3. The generator reads the local README/docs plus the current central catalog.
4. Live public GitHub metadata is used when available; the catalog snapshot is the deterministic fallback.
5. Tests validate HTML, SEO, locales, links, and output boundaries.
6. GitHub Pages deploys the artifact.
7. The central hourly sync records new versions and commits only material catalog changes.
8. Scheduled child builds propagate central design and catalog changes to all sites.

This yields immediate documentation updates on a project push and bounded eventual consistency for cross-project metadata without requiring an account-wide secret.

## Failure handling

- Unknown or unclassified products fail generation.
- Invalid HTTPS canonical or outbound URLs fail validation.
- Missing localized shared copy fails tests.
- A missing README is allowed and produces a concise metadata-led site.
- GitHub API rate limits fall back to the last committed catalog snapshot and are reported in the build manifest.
- A child site never commits generated HTML to its source branch; Pages artifacts remain disposable.
- Deployment failure never mutates the repository homepage or the central verified-site list.

## Testing and verification

Central tests cover manifest validation, locale completeness, safe README rendering, deterministic output, path traversal resistance, metadata generation, structured data, sitemap alternates, and the complete 14-site link graph.

Each child workflow validates its final artifact before upload. Existing sites additionally run their current product-specific validation.

Completion requires direct evidence for all 14 products:

- a successful default-branch CI run where one exists;
- a successful Pages deployment workflow;
- HTTP 200 for the canonical URL;
- correct title, canonical URL, structured data, robots, and sitemap;
- links from the project site to the hub and sibling directory;
- a reciprocal link from the refreshed portfolio;
- no regression in the four preserved interactive/documentation sites.

## Rollout order

1. Extend the central schema, localization, and catalog.
2. Build and test the family generator against fixtures for generated and decorated sites.
3. Upgrade the portfolio to three locales and validate the global link graph.
4. Deploy one generated-site canary (`git-barber`) and one decorated-site canary (`Mac-Coffee`).
5. Roll out the remaining generated sites.
6. Roll out the remaining decorators.
7. Enable Pages, verify production, update repository homepages, and resync the portfolio.
8. Record final deployment evidence and only then declare the project complete.

## Non-goals

- Hosting ordinary forks or support repositories as separate products.
- Replacing package documentation services such as pkg.go.dev, docs.rs, crates.io, or TypeDoc.
- Adding analytics, cookies, forms, authentication, a CMS, or a runtime backend.
- Translating source documentation that does not already have a maintained localized version.
- Moving DNS or introducing Cloudflare when GitHub Pages satisfies the static deployment requirements.
