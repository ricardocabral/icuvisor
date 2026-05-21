BINARY := icuvisor
PKG    := github.com/ricardocabral/icuvisor
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

GO         ?= go
GOLANGCI   ?= golangci-lint
GORELEASER ?= goreleaser
GOIMPORTS  ?= goimports
HUGO       ?= hugo
HUGO_PORT  ?= 1313

.DEFAULT_GOAL := help

.PHONY: all build install run test test-race cover bench lint fmt fmt-check vet tidy \
        download verify generate goimports check clean snapshot release \
        docs-tools eval-validate web-serve web-build web-clean help

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

fmt: ## Format Go code (gofmt + goimports)
	$(GO) fmt ./...
	@command -v $(GOIMPORTS) >/dev/null 2>&1 && $(GOIMPORTS) -w -local $(PKG) . || \
		echo "goimports not installed; run 'go install golang.org/x/tools/cmd/goimports@latest'"

fmt-check: ## Fail if files are not gofmt/goimports clean
	@diff=$$(gofmt -l .); \
	if [ -n "$$diff" ]; then echo "gofmt diff:"; echo "$$diff"; exit 1; fi
	@command -v $(GOIMPORTS) >/dev/null 2>&1 && \
		diff=$$($(GOIMPORTS) -l -local $(PKG) .); \
		if [ -n "$$diff" ]; then echo "goimports diff:"; echo "$$diff"; exit 1; fi || true

goimports: ## Run goimports with the project's local import group
	$(GOIMPORTS) -w -local $(PKG) .

vet: ## Run go vet
	$(GO) vet ./...

tidy: ## Tidy go.mod
	$(GO) mod tidy

download: ## Download module dependencies
	$(GO) mod download

verify: ## Verify dependencies have expected content
	$(GO) mod verify

generate: ## Run go generate
	$(GO) generate ./...

check: fmt-check vet lint test-race ## Run formatting, vet, lint, and race tests

snapshot: ## Build a local goreleaser snapshot
	$(GORELEASER) release --snapshot --clean

release: ## Run a goreleaser release (requires tag + creds)
	$(GORELEASER) release --clean

clean: web-clean ## Remove build artifacts (binary + site)
	rm -rf bin dist coverage.txt coverage.html

# ---- generated docs ----------------------------------------------------------

docs-tools: ## Regenerate web/data/tools.json from the tool registry
	$(GO) run ./cmd/gendocs --out web/data/tools.json

eval-validate: ## Validate cookbook eval scenarios against the tool catalog
	python3 scripts/eval/run_eval.py --validate

# ---- website (web/, Hugo) ----------------------------------------------------

web-serve: ## Run the Hugo dev server for the icuvisor.app site
	cd web && $(HUGO) server -D --port $(HUGO_PORT) --bind 127.0.0.1

web-build: ## Build the Hugo site into web/public
	cd web && $(HUGO) --minify --gc

web-clean: ## Remove Hugo build output
	rm -rf web/public web/resources web/.hugo_build.lock

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
