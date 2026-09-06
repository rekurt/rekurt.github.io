.PHONY: test check build site-kit-test site-fleet-check marketing-check

GO_PACKAGES := $(shell go list ./... | grep -v '/site/' | grep -v '/work/')

test:
	go test $(GO_PACKAGES)
	cd site && npm test

check:
	go vet $(GO_PACKAGES)
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './site/*' -not -path './work/*'))"
	cd site && npm run check

build:
	cd site && npm run build

site-kit-test:
	go test ./internal/projectsite ./internal/sitefleet ./cmd/project-site ./cmd/site-fleet-check

site-fleet-check:
	go run ./cmd/site-fleet-check --snapshot site/src/data/generated/catalog.json

marketing-check:
	node scripts/check-marketing.mjs
