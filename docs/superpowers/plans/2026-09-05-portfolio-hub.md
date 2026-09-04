# Rekurt Portfolio Hub Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Построить, опубликовать и проверить `https://rekurt.github.io` как двуязычный автоматически обновляемый каталог продуктов и полный реестр публичных репозиториев `rekurt`.

**Architecture:** Go CLI получает публичные GitHub-метаданные, объединяет их с кураторским YAML и атомарно создаёт детерминированный snapshot. Astro 7 читает snapshot без сети и генерирует статический EN/RU-сайт; GitHub Actions почасово синхронизирует данные и развёртывает неизменяемый Pages artifact.

**Tech Stack:** Go 1.27, YAML v3.0.1, goldmark v1.8.6, bluemonday v1.0.27, Node.js 24.20 LTS, Astro 7.3.1, TypeScript 6.0.3, Vitest 5.0.0, Playwright 1.63.0, Axe 4.13.0, parse5 8.0.1, GitHub Actions, GitHub Pages.

**Spec:** `docs/superpowers/specs/2026-09-05-portfolio-hub-design.md`

## Global Constraints

- Рабочая ветка: `main`; автор коммитов берётся только из существующего Git config.
- Каждый коммит использует Conventional Commits; пуш разрешён явным поручением пользователя довести проект до результата.
- Канонический production URL первой версии: `https://rekurt.github.io`.
- Интерфейс по умолчанию английский; полная русская локализация находится под `/ru/`.
- Сайт статический: без backend, базы данных, cookies, аналитики и runtime GitHub API.
- Все 49 исходных публичных репозиториев и новый `rekurt.github.io` должны присутствовать в полном реестре после rollout.
- В продуктовом каталоге 14 сущностей; простые форки не представляются как авторские продукты.
- `Mac-Coffee` требует видимой атрибуции `Elliotwu-7/Mac-Coffee`; `tsql` остаётся только в реестре с атрибуцией `fcoury/tsql`.
- HTML README проходит allowlist sanitizer; приватные репозитории, токены и неизвестные URL-схемы запрещены.
- Повторный sync одинаковых входных данных должен давать побайтно одинаковый результат.
- Допустимая задержка обновления метаданных и версий — не более 70 минут.

---

## Карта файлов

```text
.
├── .github/
│   ├── dependabot.yml                 # автоматические dependency PR
│   └── workflows/
│       ├── ci.yml                     # проверки push/PR
│       ├── deploy-pages.yml           # сборка и Pages deployment
│       ├── release-please.yml          # SemVer и CHANGELOG portfolio hub
│       └── sync-catalog.yml           # почасовой sync и commit snapshot
├── catalog/
│   └── projects.yaml                  # кураторский реестр 14 продуктов
├── cmd/catalog-sync/main.go           # CLI sync/check
├── docs/
│   ├── repository-audit.md            # генерируемый полный аудит
│   └── superpowers/                    # спецификация и этот план
├── internal/
│   ├── buildinfo/buildinfo.go          # версия CLI
│   ├── catalog/model.go                # доменные типы snapshot/manifest
│   ├── catalog/manifest.go             # чтение YAML
│   ├── catalog/validate.go             # инварианты и ошибки
│   ├── githubapi/client.go             # HTTP client, retry, pagination
│   ├── githubapi/repositories.go       # public repos и fork parent
│   ├── githubapi/releases.go           # release/tag/branch/readme
│   ├── markdown/render.go              # link rewrite + safe HTML
│   ├── sync/merge.go                   # GitHub + manifest → snapshot
│   ├── sync/report.go                  # Markdown audit
│   └── sync/write.go                   # stable JSON + atomic replace
├── testdata/github/                    # полные HTTP fixtures
├── site/
│   ├── astro.config.mjs
│   ├── package.json
│   ├── playwright.config.ts
│   ├── public/favicon.svg
│   ├── scripts/check-links.mjs
│   ├── src/
│   │   ├── components/                 # Header, ProjectCard, filters, footer
│   │   ├── data/generated/catalog.json # sync snapshot
│   │   ├── i18n/copy.ts                # EN/RU системные строки
│   │   ├── layouts/BaseLayout.astro
│   │   ├── lib/catalog.ts              # типизированные selectors
│   │   ├── pages/                      # EN routes
│   │   ├── pages/ru/                   # RU routes
│   │   └── styles/global.css           # дизайн-токены и layout
│   ├── tests/                          # Vitest и Playwright
│   └── tsconfig.json
├── .gitignore
├── .nvmrc
├── .golangci.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Task 1: Воспроизводимый foundation

**Files:**
- Create: `.gitignore`
- Create: `.nvmrc`
- Create: `go.mod`
- Create: `internal/buildinfo/buildinfo.go`
- Test: `internal/buildinfo/buildinfo_test.go`
- Create: `site/package.json`
- Create: `site/astro.config.mjs`
- Create: `site/tsconfig.json`
- Create: `Makefile`

**Interfaces:**
- Produces: `buildinfo.Version string`, команды `make test`, `make build`, `make check`.

- [x] **Step 1: Зафиксировать падающий тест build metadata**

```go
package buildinfo

import "testing"

func TestVersionIsSemver(t *testing.T) {
    if Version != "0.1.0" {
        t.Fatalf("Version = %q, want 0.1.0", Version)
    }
}
```

- [x] **Step 2: Проверить ожидаемое падение**

Run: `go test ./internal/buildinfo`
Expected: FAIL с `undefined: Version`.

- [x] **Step 3: Создать Go module и минимальную реализацию**

```go
// internal/buildinfo/buildinfo.go
package buildinfo

const Version = "0.1.0"
```

`go.mod` должен содержать `module github.com/rekurt/rekurt.github.io` и `go 1.27.0`.

- [x] **Step 4: Создать Astro shell с фиксированными scripts**

```json
{
  "name": "rekurt-portfolio-site",
  "private": true,
  "type": "module",
  "engines": { "node": ">=24.20.0 <25" },
  "scripts": {
    "dev": "astro dev",
    "check": "astro check && tsc --noEmit",
    "test": "vitest run",
    "build": "astro build",
    "test:e2e": "playwright test",
    "check:links": "node scripts/check-links.mjs"
  }
}
```

Установить точные major-compatible зависимости командами:

```bash
cd site
npm install astro@7.3.1
npm install -D @astrojs/check@0.9.10 typescript@6.0.3 vitest@5.0.0 eslint@10.10.0 prettier@3.9.6 @playwright/test@1.63.0 @axe-core/playwright@4.13.0 parse5@8.0.1
```

- [x] **Step 5: Добавить общие команды**

```make
.PHONY: test check build
GO_PACKAGES := $(shell go list ./... | grep -v '/site/')
test:
	go test $(GO_PACKAGES)
	cd site && npm test
check:
	go vet $(GO_PACKAGES)
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './site/*'))"
	cd site && npm run check
build:
	cd site && npm run build
```

`.gitignore` должен исключать `default.profraw`, `.DS_Store`, `site/node_modules`,
`site/dist`, `site/.astro`, `coverage`, `playwright-report`, `test-results`, но не generated
snapshot.

- [x] **Step 6: Проверить foundation**

Run: `go test ./internal/buildinfo && cd site && npm run check`
Expected: PASS; Astro сообщает 0 errors.

- [x] **Step 7: Commit**

```bash
git add .gitignore .nvmrc go.mod internal/buildinfo site/package.json site/package-lock.json site/astro.config.mjs site/tsconfig.json Makefile docs/superpowers/plans/2026-09-05-portfolio-hub.md
git commit -m "build: initialize portfolio toolchain"
```

### Task 2: Доменная модель и валидация манифеста

**Files:**
- Create: `internal/catalog/model.go`
- Create: `internal/catalog/manifest.go`
- Create: `internal/catalog/validate.go`
- Test: `internal/catalog/manifest_test.go`
- Test: `internal/catalog/validate_test.go`
- Create: `internal/catalog/testdata/valid.yaml`

**Interfaces:**
- Produces: `catalog.LoadManifest(path string) (Manifest, error)`.
- Produces: `catalog.ValidateManifest(Manifest) error`.
- Produces: `Manifest`, `ProductConfig`, `Repository`, `Product`, `Snapshot`, `Link`, `Version`.

- [x] **Step 1: Написать тест загрузки и точных полей**

```go
func TestLoadManifest(t *testing.T) {
    got, err := LoadManifest("testdata/valid.yaml")
    if err != nil { t.Fatal(err) }
    if got.Owner != "rekurt" { t.Fatalf("owner = %q", got.Owner) }
    if got.Products[0].Slug != "mac-coffee" { t.Fatalf("slug = %q", got.Products[0].Slug) }
    if got.Products[0].Upstream != "Elliotwu-7/Mac-Coffee" { t.Fatalf("upstream = %q", got.Products[0].Upstream) }
}
```

- [x] **Step 2: Проверить ожидаемое падение**

Run: `go test ./internal/catalog -run TestLoadManifest`
Expected: FAIL, пакет и функция ещё отсутствуют.

- [x] **Step 3: Определить типы без presentation logic**

```go
type LocalizedText struct { EN string `yaml:"en" json:"en"`; RU string `yaml:"ru" json:"ru"` }
type ProductConfig struct {
    Slug string `yaml:"slug"`
    PrimaryRepo string `yaml:"primary_repo"`
    Repositories []string `yaml:"repositories"`
    Kind string `yaml:"kind"`
    Domain string `yaml:"domain"`
    Featured bool `yaml:"featured"`
    MaintainedFork bool `yaml:"maintained_fork"`
    Upstream string `yaml:"upstream"`
    Summary LocalizedText `yaml:"summary"`
    Install []string `yaml:"install"`
    Website string `yaml:"website"`
    Documentation string `yaml:"documentation"`
}
type Manifest struct { Owner string `yaml:"owner"`; Products []ProductConfig `yaml:"products"` }
```

`Snapshot` должен содержать `SchemaVersion`, `Owner`, `SyncedAt`, `Products`,
`Repositories`; `Repository` — `NameWithOwner`, `Visibility`, `Fork`, `Parent`, `Language`,
`Topics`, `Homepage`, `HasPages`, `DefaultBranch`, `HeadSHA`, `License`, `UpdatedAt`,
`PushedAt`, `Archived`, `Version`, `Readme`; `Version` — `Value`, `Source`, `URL`.

- [x] **Step 4: Реализовать YAML decoder со strict known fields**

Использовать `yaml.Decoder.KnownFields(true)`, отклонять второй YAML document и возвращать
ошибки с путём файла.

- [x] **Step 5: Добавить table-driven validation tests**

Проверить duplicate slug, duplicate primary repo, пустой EN/RU summary, maintained fork без
upstream, неизвестную URL-схему, primary repo вне repositories и slug вне regex
`^[a-z0-9]+(?:-[a-z0-9]+)*$`.

- [x] **Step 6: Реализовать агрегированную детерминированную ошибку**

```go
func ValidateManifest(m Manifest) error
```

Ошибка сортирует все нарушения по `product slug + field` и возвращает их одной строкой через
`errors.Join`.

- [x] **Step 7: Проверить пакет**

Run: `go test ./internal/catalog -count=1`
Expected: PASS.

- [x] **Step 8: Commit**

```bash
git add go.mod go.sum internal/catalog
git commit -m "feat(catalog): add validated project manifest"
```

### Task 3: GitHub API client с pagination, retry и fixtures

**Files:**
- Create: `internal/githubapi/client.go`
- Create: `internal/githubapi/repositories.go`
- Create: `internal/githubapi/releases.go`
- Create: `internal/githubapi/version.go`
- Test: `internal/githubapi/client_test.go`
- Test: `internal/githubapi/repositories_test.go`
- Create: `testdata/github/repos-page-1.json`
- Create: `testdata/github/repos-page-2.json`
- Create: `testdata/github/repo-mac-coffee.json`
- Create: `testdata/github/release-ymsdk.json`
- Create: `testdata/github/tags-dbdiff.json`

**Interfaces:**
- Consumes: `catalog.Repository`, `catalog.Version`.
- Produces: `githubapi.New(baseURL, token string, httpClient *http.Client) *Client`.
- Produces: `(*Client).ListOwnedPublic(ctx context.Context, owner string) ([]catalog.Repository, error)`.
- Produces: `(*Client).Enrich(ctx context.Context, repo catalog.Repository, withReadme bool) (catalog.Repository, error)`.

- [x] **Step 1: Написать failing pagination/retry test**

```go
func TestListOwnedPublicPaginatesAndFilters(t *testing.T) {
    // httptest server returns page 1 with Link: rel=next and page 2 with one private repo.
    got, err := client.ListOwnedPublic(context.Background(), "rekurt")
    if err != nil { t.Fatal(err) }
    if len(got) != 3 { t.Fatalf("len = %d, want 3 public repos", len(got)) }
    if got[0].NameWithOwner != "rekurt/alpha" { t.Fatalf("not sorted: %#v", got) }
}
```

В отдельном тесте сервер возвращает `429`, `Retry-After: 0`, затем `200`; ожидать два запроса.

- [x] **Step 2: Проверить падение**

Run: `go test ./internal/githubapi -run 'TestList|TestRetry'`
Expected: FAIL с отсутствующим `New`/`ListOwnedPublic`.

- [x] **Step 3: Реализовать HTTP boundary**

`Client` задаёт `User-Agent: rekurt-portfolio-sync/0.1.0`, GitHub media/version headers,
Bearer token только если он непустой, timeout из переданного `http.Client`. Retry разрешён
для `429`, `502`, `503`, `504` максимум три попытки с `Retry-After` или exponential backoff.
Тело error response ограничить 8 KiB.

- [x] **Step 4: Реализовать repository pagination**

Запрашивать `/users/{owner}/repos?type=owner&sort=full_name&direction=asc&per_page=100&page=N`,
продолжать по `Link`, жёстко отбрасывать `private`/не-public записи и сортировать
`NameWithOwner` case-insensitively.

- [x] **Step 5: Написать failing enrichment tests**

Проверить parent форка, latest published non-draft release, fallback на SemVer tag, затем на
`Cargo.toml`/`package.json`, default-branch SHA, README payload и сохранение `has_pages`
отдельно от homepage. Отдельно проверить, что `go.mod` не интерпретируется как источник
версии.

- [x] **Step 6: Реализовать enrichment**

Для каждого репозитория использовать REST endpoints `/repos/{full_name}`,
`/releases?per_page=20`, `/tags?per_page=20`, `/branches/{branch}` и `/readme`. Draft и
prerelease не становятся latest version. `v1.2.3` сохраняется как `v1.2.3`; не-SemVer tag
игнорируется. Если release/tag отсутствуют, прочитать `Cargo.toml` или `package.json` из
корня и сохранить version source `manifest`; `go.mod` как version source запрещён.

- [x] **Step 7: Проверить client**

Run: `go test ./internal/githubapi -race -count=1`
Expected: PASS; race detector не находит конфликтов.

- [x] **Step 8: Commit**

```bash
git add internal/githubapi testdata/github
git commit -m "feat(sync): add resilient GitHub client"
```

### Task 4: Безопасный README renderer

**Files:**
- Create: `internal/markdown/render.go`
- Test: `internal/markdown/render_test.go`

**Interfaces:**
- Produces: `markdown.RenderREADME(source []byte, repo, branch, sha string) (catalog.Readme, error)`.

- [x] **Step 1: Написать security и rewrite tests**

```go
func TestRenderREADMERewritesAndSanitizes(t *testing.T) {
    src := []byte("[Guide](docs/guide.md) ![Shot](assets/a.png) <script>alert(1)</script> [x](javascript:alert(1))")
    got, err := RenderREADME(src, "rekurt/demo", "main", "abc123")
    if err != nil { t.Fatal(err) }
    if strings.Contains(got.HTML, "script") || strings.Contains(got.HTML, "javascript:") { t.Fatal(got.HTML) }
    if !strings.Contains(got.HTML, "github.com/rekurt/demo/blob/abc123/docs/guide.md") { t.Fatal(got.HTML) }
    if !strings.Contains(got.HTML, "raw.githubusercontent.com/rekurt/demo/abc123/assets/a.png") { t.Fatal(got.HTML) }
}
```

- [x] **Step 2: Проверить падение**

Run: `go test ./internal/markdown`
Expected: FAIL с отсутствующим `RenderREADME`.

- [x] **Step 3: Реализовать AST rewrite и render**

Использовать goldmark AST walk до render: относительные `Link` направлять на GitHub blob по
SHA, относительные `Image` — на raw.githubusercontent.com по SHA. Абсолютные `https`, `http`
и `mailto` сохранять; fragment сохранять локальным.

- [x] **Step 4: Реализовать allowlist sanitizer**

Использовать bluemonday UGC policy, дополнительно разрешить `class` только на `code`,
запретить inline style, iframe, form, svg, data URI. Добавить `rel="noreferrer noopener"` к
внешним ссылкам. Ограничить source README 256 KiB и HTML 512 KiB.

- [x] **Step 5: Проверить renderer**

Run: `go test ./internal/markdown -count=1`
Expected: PASS для XSS table и relative URL table.

- [x] **Step 6: Commit**

```bash
git add go.mod go.sum internal/markdown
git commit -m "feat(sync): sanitize project documentation"
```

### Task 5: Детерминированный sync, snapshot и audit report

**Files:**
- Create: `internal/sync/merge.go`
- Create: `internal/sync/write.go`
- Create: `internal/sync/report.go`
- Test: `internal/sync/merge_test.go`
- Test: `internal/sync/write_test.go`
- Create: `cmd/catalog-sync/main.go`
- Test: `cmd/catalog-sync/main_test.go`

**Interfaces:**
- Consumes: `catalog.Manifest`, `[]catalog.Repository`, `markdown.RenderREADME`.
- Produces: `sync.Build(manifest catalog.Manifest, repos []catalog.Repository, syncedAt time.Time) (catalog.Snapshot, error)`.
- Produces: `sync.WriteSnapshot(path string, snapshot catalog.Snapshot) (changed bool, err error)`.
- Produces: `sync.RenderAudit(snapshot catalog.Snapshot) []byte`.
- Produces CLI: `catalog-sync sync --manifest <path> --snapshot <path> --audit <path>` и `catalog-sync check ...`.

- [x] **Step 1: Написать failing merge test**

Fixture из трёх repos должен собрать один grouped product, один maintained fork с upstream и
оставить простой fork только в registry. Unknown linked repo и private repo должны дать
ошибку.

- [x] **Step 2: Проверить падение**

Run: `go test ./internal/sync -run TestBuild`
Expected: FAIL с отсутствующим `Build`.

- [x] **Step 3: Реализовать merge rules**

Products сортировать `featured desc`, затем manifest order, затем slug. Registry сортировать
по lowercase full name. Support repositories получают `role: support`; primary —
`role: primary`; неописанные forks — `role: fork`; неописанные originals —
`role: unclassified`.

- [x] **Step 4: Написать deterministic write test**

Дважды передать одинаковый snapshot с переставленными slices; оба файла должны совпасть
побайтно. Затем передать те же данные с более поздним `syncedAt`: файл не
должен измениться, а `changed` должен быть false. Временной файл после ошибки не должен
оставаться рядом с output.

- [x] **Step 5: Реализовать stable JSON и atomic replace**

Перед encode сортировать topics, links и repositories. До сравнения обнулить `SyncedAt` в
новом и существующем snapshot; при равенстве вернуть `changed=false` без записи, сохранив
старый timestamp. При материальном изменении использовать новый `SyncedAt`. Кодировать
`json.Encoder` с двумя пробелами и завершающим newline. Писать через
`os.CreateTemp(filepath.Dir(path), ".catalog-*")`, `Sync`, `Close`, `Rename`, mode `0644`.

- [x] **Step 6: Реализовать Markdown audit**

Таблица содержит repository, role, original/fork, upstream, primary language, version,
website, documentation, last push. Вверху — totals и дата snapshot. URL оформляются
Markdown-ссылками; пустые поля — `—`.

- [x] **Step 7: Реализовать CLI dependency injection**

```go
func run(ctx context.Context, args []string, getenv func(string) string, out, errOut io.Writer) error
```

`sync` допускает сеть и заменяет файлы; `check` строит данные и завершается ошибкой, если
tracked snapshot/audit отличаются. Exit code 2 — invalid arguments, 1 — runtime/validation.

- [x] **Step 8: Проверить весь Go pipeline**

Run: `go list ./... | grep -v '/site/' | xargs go test -race -count=1 && go list ./... | grep -v '/site/' | xargs go vet`
Expected: PASS.

- [x] **Step 9: Commit**

```bash
git add cmd internal/sync
git commit -m "feat(sync): generate deterministic portfolio catalog"
```

### Task 6: Кураторский каталог и живой аудит GitHub

**Files:**
- Create: `catalog/projects.yaml`
- Create: `site/src/data/generated/catalog.json`
- Create: `docs/repository-audit.md`
- Test: `catalog/projects_test.go`

**Interfaces:**
- Consumes: CLI Task 5.
- Produces: 14 validated product records and current full GitHub snapshot.

- [ ] **Step 1: Написать repository-level contract test**

```go
func TestProductionManifestContract(t *testing.T) {
    m, err := catalog.LoadManifest("projects.yaml")
    if err != nil { t.Fatal(err) }
    if err := catalog.ValidateManifest(m); err != nil { t.Fatal(err) }
    if len(m.Products) != 14 { t.Fatalf("products = %d, want 14", len(m.Products)) }
    // Assert Mac-Coffee upstream and that tsql is absent from product primary repos.
}
```

- [ ] **Step 2: Проверить падение**

Run: `go test ./catalog`
Expected: FAIL, production manifest отсутствует.

- [ ] **Step 3: Создать 14 полных product records**

Каждая запись содержит фактические repositories из спецификации, разные EN/RU summary,
корректный domain (`fintech`, `developer-tools`, `infrastructure`, `sdk`, `ai`), kind
(`library`, `cli`, `desktop`, `platform`, `skill`) и рабочие install commands. Featured:
`openkline`, `Mac-Coffee`, `git-barber`, `vpn-hub`, `ymsdk`, `gost-crypto`.

- [ ] **Step 4: Выполнить live sync без вывода токена**

```bash
GITHUB_TOKEN="$(gh auth token)" go run ./cmd/catalog-sync sync \
  --manifest catalog/projects.yaml \
  --snapshot site/src/data/generated/catalog.json \
  --audit docs/repository-audit.md
```

Expected до создания remote: 49 registry entries, 14 products, 19 originals, 30 forks.

- [ ] **Step 5: Проверить содержимое snapshot**

Run: `jq '{products:(.products|length),repos:(.repositories|length),private:[.repositories[]|select(.visibility!="public")]}' site/src/data/generated/catalog.json`
Expected: `products: 14`, `repos: 49`, `private: []`.

- [ ] **Step 6: Проверить известные site/doc categories**

Run: `go test ./catalog ./internal/sync -count=1`
Expected: шесть исходных Pages URL имеют link kind `website`; pkg.go.dev/crates.io имеют kind
`documentation`; upstream homepage простых forks не становится author website.

- [ ] **Step 7: Commit**

```bash
git add catalog site/src/data/generated/catalog.json docs/repository-audit.md
git commit -m "feat(catalog): add curated public project inventory"
```

### Task 7: Astro data layer и дизайн-система

**Files:**
- Create: `site/src/lib/catalog.ts`
- Test: `site/tests/catalog.test.ts`
- Create: `site/src/i18n/copy.ts`
- Test: `site/tests/i18n.test.ts`
- Create: `site/src/styles/global.css`
- Create: `site/src/layouts/BaseLayout.astro`
- Create: `site/src/components/Header.astro`
- Create: `site/src/components/Footer.astro`
- Create: `site/src/components/ProjectCard.astro`
- Create: `site/src/components/LinkGroup.astro`
- Create: `site/src/components/VersionBadge.astro`
- Create: `site/public/favicon.svg`

**Interfaces:**
- Consumes: `site/src/data/generated/catalog.json` schema v1.
- Produces: `getCatalog(): Catalog`, `getProducts(locale: Locale): LocalizedProduct[]`,
  `getProduct(slug: string, locale: Locale): LocalizedProduct`, `alternatePath(path, locale)`.
- Produces Astro props: `BaseLayout({title, description, locale, canonicalPath})`.

- [ ] **Step 1: Написать failing selector/i18n tests**

```ts
it('localizes without changing project identity', () => {
  const en = getProduct('vpn-hub', 'en');
  const ru = getProduct('vpn-hub', 'ru');
  expect(en.slug).toBe(ru.slug);
  expect(en.summary).not.toBe(ru.summary);
});

it('maps alternate paths symmetrically', () => {
  expect(alternatePath('/projects/vpn-hub/', 'ru')).toBe('/ru/projects/vpn-hub/');
  expect(alternatePath('/ru/projects/vpn-hub/', 'en')).toBe('/projects/vpn-hub/');
});
```

- [ ] **Step 2: Проверить падение**

Run: `cd site && npm test -- catalog.test.ts i18n.test.ts`
Expected: FAIL с отсутствующими modules/functions.

- [ ] **Step 3: Реализовать runtime-free selectors**

JSON импортируется статически. На старте модуль проверяет `schemaVersion === 1`; selectors
возвращают readonly copies и выбрасывают `Unknown product: <slug>` для неверного slug.

- [ ] **Step 4: Определить системный EN/RU copy**

В `copy.ts` тип `CopyKey` выводится из English object; Russian object обязан
`satisfies Record<CopyKey, string>`, чтобы пропущенный перевод ломал typecheck.

- [ ] **Step 5: Реализовать дизайн-токены и layout**

CSS variables: graphite `#080b10`, surface `#10151d`, border `#283241`, text `#f3f7fb`,
muted `#9aa8b8`, cyan `#50d4ff`, emerald `#55e6a5`, warning `#ffca68`. Использовать системный
sans stack и `ui-monospace`; spacing 4/8 px; focus outline 2 px; min touch target 44 px;
breakpoints 640/960/1280 px; reduced-motion media query.

- [ ] **Step 6: Реализовать базовые компоненты**

Компоненты не содержат hardcoded language strings, используют semantic elements, видимый
focus, `aria-current`, безопасный external link `rel`. `ProjectCard` показывает только
существующие Website/Documentation/Source/Release actions.

- [ ] **Step 7: Проверить unit/type/build shell**

Run: `cd site && npm test && npm run check`
Expected: PASS и 0 Astro errors.

- [ ] **Step 8: Commit**

```bash
git add site/src site/public site/tests
git commit -m "feat(site): add portfolio design system"
```

### Task 8: Двуязычные страницы, каталог и полный реестр

**Files:**
- Create: `site/src/pages/index.astro`
- Create: `site/src/pages/projects/index.astro`
- Create: `site/src/pages/projects/[slug].astro`
- Create: `site/src/pages/registry/index.astro`
- Create: `site/src/pages/about/index.astro`
- Create: `site/src/pages/ru/index.astro`
- Create: `site/src/pages/ru/projects/index.astro`
- Create: `site/src/pages/ru/projects/[slug].astro`
- Create: `site/src/pages/ru/registry/index.astro`
- Create: `site/src/pages/ru/about/index.astro`
- Create: `site/src/pages/404.astro`
- Create: `site/src/pages/robots.txt.ts`
- Create: `site/src/pages/sitemap.xml.ts`
- Create: `site/src/components/FilterBar.astro`
- Create: `site/src/components/RepositoryTable.astro`
- Create: `site/src/components/ReleaseList.astro`
- Test: `site/tests/pages.test.ts`

**Interfaces:**
- Consumes: selectors and components from Task 7.
- Produces: every route in spec, static product paths for 14 slugs × 2 locales.

- [ ] **Step 1: Написать route contract test**

Тест вызывает exported `getStaticPaths` helpers и проверяет 14 EN + 14 RU slugs, canonical
URL, `hreflang=en`, `hreflang=ru`, уникальные titles и отсутствие fork homepage в author
website actions.

- [ ] **Step 2: Проверить падение**

Run: `cd site && npm test -- pages.test.ts`
Expected: FAIL, страницы ещё отсутствуют.

- [ ] **Step 3: Реализовать Home и About**

Hero: `Backend systems, developer tools, and open source built to be operated.`; RU:
`Backend-системы, инструменты разработчика и open source, готовые к эксплуатации.`
Featured grid показывает шесть выбранных продуктов; latest releases сортируются по дате.
About перечисляет Go, Rust, TypeScript, fintech/crypto, безопасные релизы и observability без
непроверяемых работодателей или биографических утверждений.

- [ ] **Step 4: Реализовать Projects и details**

Filter controls используют настоящие `<button>` и query-independent client filtering по
`data-domain`, `data-kind`, `data-language`; без JavaScript все карточки видимы. Detail page
показывает summary, version source, last sync, install commands, связанные repos, upstream
attribution, safe README HTML и четыре раздельных типа ссылок.

- [ ] **Step 5: Реализовать Registry**

Desktop — table, mobile — stacked rows. Обязательные колонки: repository, role, origin,
language, version, website/docs, last push. `original`, `maintained fork`, `fork/mirror`,
`support`, `portfolio hub` имеют разные текстовые labels, а не только цвета.

- [ ] **Step 6: Реализовать SEO и 404**

Каждая страница получает canonical, description, Open Graph, `hreflang`, JSON-LD
`Person`/`SoftwareSourceCode` только из проверенных полей. `robots.txt.ts` возвращает
`User-agent: *`, `Allow: /` и абсолютный sitemap URL. `sitemap.xml.ts` создаёт URL для всех
статических страниц и 28 product detail routes с `xhtml:link` alternate EN/RU; XML escaping
покрывается route contract test. 404 содержит ссылки Home/Projects.

- [ ] **Step 7: Проверить production build и маршруты**

Run: `cd site && npm test && npm run check && npm run build`
Expected: PASS; `dist` содержит EN/RU index, 28 product detail pages, registry, about, 404,
robots и sitemap.

- [ ] **Step 8: Commit**

```bash
git add site/src site/public
git commit -m "feat(site): publish bilingual project catalog"
```

### Task 9: End-to-end, accessibility и link QA

**Files:**
- Create: `site/playwright.config.ts`
- Create: `site/tests/e2e/navigation.spec.ts`
- Create: `site/tests/e2e/accessibility.spec.ts`
- Create: `site/tests/e2e/responsive.spec.ts`
- Create: `site/scripts/check-links.mjs`
- Test: `site/tests/check-links.test.ts`

**Interfaces:**
- Consumes: production Astro build.
- Produces: `npm run test:e2e`, `npm run check:links`, screenshots as CI artifacts.

- [ ] **Step 1: Написать failing link checker test**

Создать temp dist с `index.html` → `/missing/`; ожидать сообщение
`index.html: broken internal link /missing/`. Fixture с `/projects/` и соответствующим
`projects/index.html` должна пройти.

- [ ] **Step 2: Реализовать deterministic internal link checker**

Парсить HTML через `parse5@8.0.1`, нормализовать trailing slash,
игнорировать `mailto`, fragments и external origins, проверять файлы и directory index.

- [ ] **Step 3: Добавить Playwright configuration**

`webServer.command = "npm run build && npm exec astro preview -- --host 127.0.0.1"`,
`baseURL = http://127.0.0.1:4321`, projects: Chromium desktop 1440×1000 и mobile 390×844.
Retries 1 только в CI; trace сохранять on-first-retry.

- [ ] **Step 4: Добавить navigation и filter tests**

Проверить skip-link, EN↔RU round trip, переход Project → Website/Docs labels, фильтр domain,
страницу maintained fork с upstream attribution и registry row count не меньше 49 до remote.

- [ ] **Step 5: Добавить Axe smoke tests**

Запустить Axe на `/`, `/projects/`, `/projects/mac-coffee/`, `/registry/`, `/ru/`; zero serious
и critical violations. Дополнительно пройти header/filter/cards клавишей Tab.

- [ ] **Step 6: Добавить responsive screenshot tests**

Снимать home, projects, registry в desktop/mobile. Проверять отсутствие horizontal overflow:
`document.documentElement.scrollWidth === document.documentElement.clientWidth`.

- [ ] **Step 7: Запустить полный локальный QA**

Run: `cd site && npx playwright install chromium && npm run build && npm run check:links && npm run test:e2e`
Expected: PASS; screenshots читаемы на обоих viewport.

- [ ] **Step 8: Commit**

```bash
git add site/playwright.config.ts site/tests site/scripts site/package.json site/package-lock.json
git commit -m "test(site): add accessibility and navigation QA"
```

### Task 10: CI, scheduled sync, Pages deployment и документация

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/sync-catalog.yml`
- Create: `.github/workflows/deploy-pages.yml`
- Create: `.github/workflows/release-please.yml`
- Create: `.github/dependabot.yml`
- Create: `release-please-config.json`
- Create: `.release-please-manifest.json`
- Create: `README.md`
- Create: `CONTRIBUTING.md`

**Interfaces:**
- Produces: required CI, hourly sync commit, Pages artifact deployment, dependency PRs.

- [ ] **Step 1: Добавить CI workflow**

Использовать `actions/checkout@v7`, `actions/setup-go@v7` с cache, `actions/setup-node@v7`
с npm cache `site/package-lock.json`. Jobs: Go race/vet/gofmt; site install/check/test/build;
Playwright Chromium; upload test artifacts on failure. Permissions: `contents: read`.

- [ ] **Step 2: Добавить sync workflow**

Triggers: cron `17 * * * *`, `workflow_dispatch`, push paths для `catalog/**`, `cmd/**`,
`internal/**`. Permissions: `contents: write`. Concurrency group `catalog-sync`,
`cancel-in-progress: false`. После `go run ... sync` выполнить tests, commit только при diff:

```bash
git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
git add site/src/data/generated/catalog.json docs/repository-audit.md
git diff --cached --quiet || git commit -m "chore(catalog): sync project metadata"
git push
```

- [ ] **Step 3: Добавить Pages deployment**

Использовать `actions/configure-pages@v6`, `actions/upload-pages-artifact@v5`,
`actions/deploy-pages@v5`. Build job permissions `contents: read`, deploy permissions
`pages: write`, `id-token: write`; environment URL берётся из deployment output. Deploy
только push в `main` и manual dispatch, с concurrency `pages` и `cancel-in-progress: false`.

- [ ] **Step 4: Добавить dependency/release automation**

Dependabot еженедельно обновляет `gomod`, `npm` в `/site`, `github-actions`; лимит 5 PR на
ecosystem. `release-please.yml` использует `googleapis/release-please-action@v5`, permissions
`contents: write`, `pull-requests: write`, запускается на push в `main` и читает локальные
config/manifest. Release Please использует simple release type, package name
`rekurt-portfolio`, начальную версию `0.1.0` и conventional commits.

- [ ] **Step 5: Документировать эксплуатацию**

README содержит architecture diagram в Mermaid, локальные prerequisites, `make check`, live
sync, добавление продукта одной YAML-записью, правило forks, generated files, расписание,
failure recovery и production URL. CONTRIBUTING фиксирует Conventional Commits и запрет
ручного редактирования snapshot/audit.

- [ ] **Step 6: Локально проверить workflow YAML и полный pipeline**

Run: `ruby -e 'require "yaml"; Dir[".github/workflows/*.yml"].each { |f| YAML.load_file(f, aliases: true); puts f }'`
Expected: четыре пути workflow без YAML parse errors.

Run: `make check && make test && make build && cd site && npm run check:links`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add .github release-please-config.json .release-please-manifest.json README.md CONTRIBUTING.md
git commit -m "ci: automate catalog sync and Pages deployment"
```

### Task 11: Удалённый репозиторий, production rollout и финальный аудит

**Files:**
- Modify: `site/src/data/generated/catalog.json`
- Modify: `docs/repository-audit.md`
- Create: `docs/production-verification.md`

**Interfaces:**
- Consumes: весь локально проверенный проект.
- Produces: public `rekurt/rekurt.github.io`, green Actions, live Pages URL и доказательства.

- [ ] **Step 1: Проверить локальный release candidate**

Run: `git status --short --branch && git log --oneline --decorate -12 && make check && make test && make build`
Expected: только игнорируемые artifacts; все проверки PASS; conventional history на `main`.

- [ ] **Step 2: Создать public remote и выполнить первый push**

```bash
gh repo create rekurt/rekurt.github.io --public --source=. --remote=origin --push \
  --description "Unified portfolio and live catalog of rekurt open-source projects"
```

Expected: `origin` указывает на `https://github.com/rekurt/rekurt.github.io.git`; `main`
отслеживает `origin/main`.

- [ ] **Step 3: Включить GitHub Pages workflow mode**

```bash
gh api --method POST repos/rekurt/rekurt.github.io/pages -f build_type=workflow
```

Если endpoint возвращает `409 Already exists`, проверить
`gh api repos/rekurt/rekurt.github.io/pages --jq .build_type`; ожидается `workflow`.

- [ ] **Step 4: Повторить sync после появления самого hub**

```bash
GITHUB_TOKEN="$(gh auth token)" go run ./cmd/catalog-sync sync \
  --manifest catalog/projects.yaml \
  --snapshot site/src/data/generated/catalog.json \
  --audit docs/repository-audit.md
jq '.repositories | length' site/src/data/generated/catalog.json
```

Expected: 50 или больше, `rekurt/rekurt.github.io` имеет роль `portfolio-hub`, private count 0.

- [ ] **Step 5: Зафиксировать self-discovery и push**

```bash
git add site/src/data/generated/catalog.json docs/repository-audit.md
git commit -m "chore(catalog): register portfolio hub"
git push origin main
```

- [ ] **Step 6: Дождаться и проверить Actions**

Run: `gh run list --repo rekurt/rekurt.github.io --limit 10`
Expected: latest CI, sync и deploy runs имеют conclusion `success`. Для активного run:
`gh run watch <run-id> --repo rekurt/rekurt.github.io --exit-status`.

- [ ] **Step 7: Проверить production HTTP и содержание**

```bash
for path in / /projects/ /registry/ /ru/ /projects/mac-coffee/; do
  curl -fsSL -o /dev/null -w "%{http_code} %{url_effective}\n" "https://rekurt.github.io${path}"
done
```

Expected: пять ответов 200. HTML `/registry/` содержит `rekurt/rekurt.github.io`; detail
`mac-coffee` содержит `Elliotwu-7/Mac-Coffee`; `/ru/` имеет `lang="ru"`.

- [ ] **Step 8: Провести requirement-by-requirement audit**

`docs/production-verification.md` фиксирует commit SHA, Actions run URLs, Pages deployment
URL, registry/product counts, шесть исходных project websites, package docs, EN/RU routes,
mobile/desktop QA, security invariants и результат live curl. Каждый пункт спецификации
получает `PASS` и конкретную команду/URL доказательства.

- [ ] **Step 9: Зафиксировать финальный отчёт и push**

```bash
git add docs/production-verification.md
git commit -m "docs: record production verification"
git push origin main
```

- [ ] **Step 10: Проверить финальный чистый state**

Run: `git status --short --branch && git rev-parse HEAD && gh run list --repo rekurt/rekurt.github.io --limit 5`
Expected: `main...origin/main`, нет tracked changes, финальный deploy успешен.

---

## Порядок исполнения

Задачи выполняются строго 1 → 11. После каждого task запускаются указанные тесты и создаётся
отдельный conventional commit. При несовпадении live GitHub API со fixture-контрактом сначала
обновляется тест и документируется причина, затем меняется реализация. Запрещено обходить
ошибки sync публикацией частичного snapshot.
