# Individual project sites — production verification

Verified on 2026-09-06. This report supersedes the generic-generated-page presentation described in [project-sites-verification.md](project-sites-verification.md). Product selection and exclusions remain in [repository-audit.md](repository-audit.md).

## Published experiences

The ten generated landing pages now have individually authored English, Russian and Simplified Chinese stories, product-specific demonstrations, visual themes, benefit sections and adoption actions. They share the author navigation, catalog, metadata and delivery infrastructure, not a generic metadata-first hero.

| Project | Individual presentation | Production | Successful deployment |
| --- | --- | --- | --- |
| git-barber | Green terminal; branch-cleanup workflow and real animated demo. | [site](https://rekurt.github.io/git-barber/) | [34006104383](https://github.com/rekurt/git-barber/actions/runs/34006104383) |
| prt | Blue port-monitor console; actual watch output format and product screenshot. | [site](https://rekurt.github.io/prt/) | [34006104347](https://github.com/rekurt/prt/actions/runs/34006104347) |
| dbdiff | Light editorial layout; schema diff as the primary demonstration. | [site](https://rekurt.github.io/dbdiff/) | [34006104824](https://github.com/rekurt/dbdiff/actions/runs/34006104824) |
| ymsdk | Warm messaging layout; a bot-message API example and integration path. | [site](https://rekurt.github.io/ymsdk/) | [34006104799](https://github.com/rekurt/ymsdk/actions/runs/34006104799) |
| gost-crypto | Cryptographic primitives and signing example; explicit certification/audit limits. | [site](https://rekurt.github.io/gost-crypto/) | [34006104699](https://github.com/rekurt/gost-crypto/actions/runs/34006104699) |
| depth | Market-terminal composition with a full-width repository screenshot. | [site](https://rekurt.github.io/depth/) | [34006104751](https://github.com/rekurt/depth/actions/runs/34006104751) |
| gitlab-downloader | Transfer-plan narrative with explicit Git-only migration limits. | [site](https://rekurt.github.io/gitlab-downloader/) | [34006104586](https://github.com/rekurt/gitlab-downloader/actions/runs/34006104586) |
| go-propisyu | Document-like receipt layout with Russian amount conversion. | [site](https://rekurt.github.io/go-propisyu/) | [34006104355](https://github.com/rekurt/go-propisyu/actions/runs/34006104355) |
| cortex-forge | Experimental lab layout; explicit warning against production or sensitive data. | [site](https://rekurt.github.io/cortex-forge/) | [34006104577](https://github.com/rekurt/cortex-forge/actions/runs/34006104577) |
| sprint-velocity | Weekly report layout; translated example and task-count limitations. | [site](https://rekurt.github.io/sprint-velocity/) | [34006104374](https://github.com/rekurt/sprint-velocity/actions/runs/34006104374) |
| openkline.tech | Existing individual charting site retained; author and project-family navigation added. | [site](https://openkline.tech/) | [34006070958](https://github.com/rekurt/openkline.tech/actions/runs/34006070958) |

Mac Coffee, VPN Hub and chislo retain their existing individual marketing experiences. Their production roots returned HTTP 200 with family navigation during this verification. OpenKline's separate GitHub Pages playground and API documentation are also retained. All 14 products pass the reciprocal family-link check.

CortexForge is presented as an experimental lab, and sprint-velocity as a small automation recipe, not as enterprise-ready flagships. Archived repositories, ordinary forks and auxiliary repositories remain outside the product-marketing set.

## Evidence

- Marketing kit commit: [91f1782](https://github.com/rekurt/rekurt.github.io/commit/91f1782a09f20edbb5bc70635bc0e946dd15abcd).
- Central [CI 34006089934](https://github.com/rekurt/rekurt.github.io/actions/runs/34006089934), [Deploy Pages 34006089887](https://github.com/rekurt/rekurt.github.io/actions/runs/34006089887), and [Sync catalog 34006089925](https://github.com/rekurt/rekurt.github.io/actions/runs/34006089925) completed successfully.
- `make check test build` passed locally: Go vet/tests, Astro/TypeScript checks, 14 frontend unit tests and 55 static pages.
- Each of the ten generated sites passed build and validation against its actual repository checkout.
- `make marketing-check` passed against production: **30 locale pages and 17 stylesheet/image resources**. It checks each product's exact headline, benefits, workflow text, CTA anchor, theme identity, canonical URL and family link. It was also verified to reject the old generic landing output.
- `make site-fleet-check` passed for **all 14 products**, including localized directories and reciprocal links.
- OpenKline's existing suite passed **211 tests**, and its Vite production build passed. The published openkline.tech HTML contains the new author and catalog links.
- The central CI's existing browser suite passed. The new child-page verification is build/HTML/network-based; no new visual browser audit is claimed.

## Maintenance

Profiles are stored in `internal/projectsite/profiles/<slug>.json`; visual treatments live in `internal/projectsite/assets/marketing.css`. To onboard another individual landing, add a catalog record, a complete trilingual profile, its visual treatment, and the thin Pages workflow. Existing sites can be integrated with `decorate`.

Repository pushes, releases and the six-hour child schedules refresh versions and repository documentation while preserving authored marketing copy. The author catalog synchronizes hourly.

The catalog bot's commits do not themselves trigger another push workflow, as documented by [GitHub](https://docs.github.com/en/actions/concepts/security/github_token). The sync workflow now explicitly dispatches publication after successful verification, including no-change runs. Both synchronization and publication are restricted to `main`; the repository-scoped token receives `actions: write` for this dispatch. This avoids requiring a separate personal token.

