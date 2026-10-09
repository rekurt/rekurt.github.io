# Contributing

## Change sources

- Edit `catalog/projects.yaml` for curated product metadata.
- For published npm packages, `npm_package` records a registry-confirmed package name and exact version separately from the GitHub release/tag version. Update it after a verified npm release; catalog-sync preserves both sources.
- Edit `cmd/` or `internal/` for synchronization behavior.
- Edit `site/src/` for the generated static experience.
- Never edit `site/src/data/generated/catalog.json` or `docs/repository-audit.md` manually. Regenerate both with `catalog-sync`.

Ordinary forks remain registry-only. A fork may become a product only when it is substantially maintained and its manifest entry contains both `maintained_fork: true` and an explicit `upstream` repository.

## Commit convention

Use Conventional Commits with a focused scope, for example:

```text
feat(catalog): add payment routing project
fix(sync): preserve release provenance
docs: clarify repository onboarding
```

Generated metadata belongs in the same commit as the manifest or synchronizer change that produced it. The hourly automation uses `chore(catalog): sync project metadata`.

## Verification

Install site dependencies once with `cd site && npm ci`, then run:

```bash
make check
make test
make build
cd site
npm run check:links
npm run test:e2e
```

Run `make check-npm` to verify curated and generated npm metadata against the public registry. The check uses unauthenticated read-only requests, verifies exact published identity, valid SHA-512 integrity metadata, official registry tarball URL and install instructions, and prints official package-version links. Missing releases or conflicting metadata fail; a different registry `latest` emits a warning for review without changing pinned versions. CI runs this check, and the existing hourly catalog-sync also runs it before publishing generated updates. No packages, releases, credentials or permissions are created.

A contribution is ready only when Go formatting and vet, Go and Vitest tests, Astro type checking, the production build, internal links, and browser QA all pass.

For per-project layouts and shared browser behavior, also run `make site-kit-test`. It includes localized page rendering and copy/documentation interaction regressions. Check a generated project page at mobile and desktop widths before deploying the family kit.
