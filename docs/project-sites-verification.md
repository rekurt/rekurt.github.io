# Project sites production verification

Verified on 2026-09-05. The public repository inventory contains 50 repositories; 14 maintained products are published as the curated project family. The complete inventory and the reasons for excluding archived, forked, profile, mirror, and auxiliary repositories are recorded in [repository-audit.md](repository-audit.md).

## Production matrix

| Project | Integration | Source revision | Successful Pages run | Production URL |
| --- | --- | --- | --- | --- |
| openkline | Decorated existing site | `bc686b6` | [33945808455](https://github.com/rekurt/openkline/actions/runs/33945808455) | <https://rekurt.github.io/openkline/> |
| Mac Coffee | Decorated existing site | `5b16ebd` | [33945651862](https://github.com/rekurt/Mac-Coffee/actions/runs/33945651862) | <https://rekurt.github.io/Mac-Coffee/> |
| git-barber | Generated site | `bb59cdf` | [33945062507](https://github.com/rekurt/git-barber/actions/runs/33945062507) | <https://rekurt.github.io/git-barber/> |
| vpn-hub | Decorated existing site | `f4768b2` | [33945809835](https://github.com/rekurt/vpn-hub/actions/runs/33945809835) | <https://rekurt.github.io/vpn-hub/> |
| ymsdk | Generated site | `c246c4e` | [33945483947](https://github.com/rekurt/ymsdk/actions/runs/33945483947) | <https://rekurt.github.io/ymsdk/> |
| gost-crypto | Generated site | `be10e2d` | [33946133342](https://github.com/rekurt/gost-crypto/actions/runs/33946133342) | <https://rekurt.github.io/gost-crypto/> |
| chislo | Decorated existing site | `b1d5b60` | [33945811346](https://github.com/rekurt/chislo/actions/runs/33945811346) | <https://rekurt.github.io/chislo/> |
| cortex-forge | Generated site | `1ef96d5` | [33945486659](https://github.com/rekurt/cortex-forge/actions/runs/33945486659) | <https://rekurt.github.io/cortex-forge/> |
| dbdiff | Generated site | `2d66cdd` | [33945488192](https://github.com/rekurt/dbdiff/actions/runs/33945488192) | <https://rekurt.github.io/dbdiff/> |
| depth | Generated site | `c434c6b` | [33945490153](https://github.com/rekurt/depth/actions/runs/33945490153) | <https://rekurt.github.io/depth/> |
| gitlab-downloader | Generated site | `9c90ca0` | [33946842873](https://github.com/rekurt/gitlab-downloader/actions/runs/33946842873) | <https://rekurt.github.io/gitlab-downloader/> |
| go-propisyu | Generated site | `bb58783` | [33945493779](https://github.com/rekurt/go-propisyu/actions/runs/33945493779) | <https://rekurt.github.io/go-propisyu/> |
| prt | Generated site | `607a7d6` | [33946540177](https://github.com/rekurt/prt/actions/runs/33946540177) | <https://rekurt.github.io/prt/> |
| sprint-velocity | Generated site | `7087582` | [33945497251](https://github.com/rekurt/sprint-velocity/actions/runs/33945497251) | <https://rekurt.github.io/sprint-velocity/> |

Every production endpoint is checked from the central repository with `make site-fleet-check`. The checker validates the build manifest, canonical URL, JSON-LD, shared family navigation, reciprocal links, locale routes, sitemaps, robots policy, HTTP status, redirect behavior, and mixed-content safety. Generated sites expose English, Russian, and Simplified Chinese routes; decorated sites preserve their existing product experience and receive the common project-family layer.

## Delivery model

Project repositories call the reusable workflow from `rekurt/rekurt.github.io`. On every push to the default branch, the workflow checks out the current central site kit, syncs live GitHub metadata, builds or decorates the project site, validates the output, and publishes it to GitHub Pages. Adding another project requires one catalog record and the thin Pages caller workflow.

The central catalog sync runs on changes and on a schedule. It refreshes repository metadata, versions, documentation-derived content, and the public audit without copying project source into the hub repository.

## Quality evidence

The central site passed Go tests, JavaScript unit tests, type checks, Astro production build, 623-link validation, responsive browser tests, accessibility checks, and the live fleet verification. Repository-native test suites were also run for maintained products where available. Notable repairs made during delivery:

- gost-crypto CI was migrated to a golangci-lint version compatible with the current Go toolchain.
- PRT updated `anyhow` and `crossbeam-epoch` to resolve two RustSec advisories; tests, clippy, and Cargo Deny pass.
- gitlab-downloader re-enabled Actions and corrected its Webpack-result assertion to support plural warning output.

Historical Pages failures produced before Pages was enabled are superseded by the successful runs listed above. Existing dependency-audit findings in openkline, depth, and gitlab-downloader are product-maintenance items and are not introduced by the site family.
