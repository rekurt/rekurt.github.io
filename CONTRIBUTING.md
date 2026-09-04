# Contributing

## Change sources

- Edit `catalog/projects.yaml` for curated product metadata.
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

A contribution is ready only when Go formatting and vet, Go and Vitest tests, Astro type checking, the production build, internal links, and browser QA all pass.
