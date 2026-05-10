BINARY := icuvisor
PKG    := github.com/ricardocabral/icuvisor
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

GO         ?= go
GOLANGCI   ?= golangci-lint
GORELEASER ?= goreleaser

.PHONY: all build install run test test-race cover bench lint fmt vet tidy clean snapshot release help

all: build ## Build the binary

build: ## Build the binary into ./bin
	@mkdir -p bin
	$(GO) build -trimpath -ldflags='$(LDFLAGS)' -o bin/$(BINARY) ./cmd/$(BINARY)

install: ## Install the binary into $GOBIN
	$(GO) install -trimpath -ldflags='$(LDFLAGS)' ./cmd/$(BINARY)

run: ## Run the binary
	$(GO) run ./cmd/$(BINARY)

test: ## Run unit tests
	$(GO) test ./...

test-race: ## Run tests with the race detector
	$(GO) test -race -count=1 ./...

cover: ## Run tests with coverage report
	$(GO) test -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GO) tool cover -func=coverage.txt | tail -1

bench: ## Run benchmarks
	$(GO) test -bench=. -benchmem -run=^$$ ./...

lint: ## Run golangci-lint
	$(GOLANGCI) run ./...

fmt: ## Format Go code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

tidy: ## Tidy go.mod
	$(GO) mod tidy

snapshot: ## Build a local goreleaser snapshot
	$(GORELEASER) release --snapshot --clean

release: ## Run a goreleaser release (requires tag + creds)
	$(GORELEASER) release --clean

clean: ## Remove build artifacts
	rm -rf bin dist coverage.txt coverage.html

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
