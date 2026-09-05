.PHONY: test check build site-kit-test

GO_PACKAGES := $(shell go list ./... | grep -v '/site/')

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
	go test ./internal/projectsite ./cmd/project-site
