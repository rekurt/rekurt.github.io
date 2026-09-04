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
