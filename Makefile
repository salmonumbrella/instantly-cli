.PHONY: build install clean test lint fmt help

BINARY_NAME := instantly
MODULE := github.com/salmonumbrella/instantly-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "")
LDFLAGS := -ldflags "-s -w \
  -X github.com/salmonumbrella/instantly-cli/internal/buildinfo.Version=$(VERSION) \
  -X github.com/salmonumbrella/instantly-cli/internal/buildinfo.Commit=$(COMMIT) \
  -X github.com/salmonumbrella/instantly-cli/internal/buildinfo.Date=$(DATE)"

## help: Show this help
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed -e 's/^##//'

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

## install: Install the binary to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/$(BINARY_NAME)

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

## test: Run tests
test:
	go test -v -race -cover ./...

## smoke-live: Run a read-only smoke test against the real API (requires INSTANTLY_API_KEY)
smoke-live:
	./scripts/smoke-live.sh

## lint: Run linters
lint:
	golangci-lint run

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## tidy: Tidy dependencies
tidy:
	go mod tidy

## release-snapshot: Build snapshot release (no publish)
release-snapshot:
	goreleaser release --snapshot --clean

## release: Build and publish release
release:
	goreleaser release --clean
