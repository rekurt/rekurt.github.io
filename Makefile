.PHONY: test check build

test:
	go test ./...
	cd site && npm test

check:
	go vet ./...
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './site/*'))"
	cd site && npm run check

build:
	cd site && npm run build
