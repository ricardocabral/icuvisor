BINARY     := icuvisor
CLI_BINARY := icuvisor-cli
PKG        := github.com/ricardocabral/icuvisor
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

GO         ?= go
GOLANGCI   ?= golangci-lint
GORELEASER       ?= goreleaser
GOIMPORTS        ?= goimports
HUGO             ?= hugo
MCPB_CLI_PACKAGE ?= @anthropic-ai/mcpb@2.1.2
HUGO_PORT  ?= 1313

.DEFAULT_GOAL := help

.PHONY: all build install run test test-race docs-guidance-test cover bench lint fmt fmt-check vet tidy \
        download verify generate goimports check clean snapshot release release-preflight \
        validate-registry validate-distribution docs-tools eval-validate eval-tool-routing web-serve web-preview web-build web-clean help

all: build ## Build the binary

build: ## Build the MCP and standalone CLI binaries into ./bin
	@mkdir -p bin
	$(GO) build -trimpath -ldflags='$(LDFLAGS)' -o bin/$(BINARY) ./cmd/$(BINARY)
	$(GO) build -trimpath -ldflags='$(LDFLAGS)' -o bin/$(CLI_BINARY) ./cmd/$(CLI_BINARY)

install: ## Install the MCP and standalone CLI binaries into $GOBIN
	$(GO) install -trimpath -ldflags='$(LDFLAGS)' ./cmd/$(BINARY) ./cmd/$(CLI_BINARY)

run: ## Run the MCP binary
	$(GO) run ./cmd/$(BINARY)

test: ## Run unit tests
	$(GO) test ./...

docs-guidance-test: ## Verify published documentation guidance contracts
	python3 scripts/tests/test_docs_guidance.py
	python3 scripts/tests/test_http_service_docs.py
	python3 scripts/tests/test_build_workouts_guidance.py
	python3 scripts/tests/test_homepage_capability_copy.py

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
	@if gofmt -l . | grep -q .; then echo "gofmt diff:"; gofmt -l .; exit 1; fi
	@if command -v $(GOIMPORTS) >/dev/null 2>&1; then if $(GOIMPORTS) -l -local $(PKG) . | grep -q .; then echo "goimports diff:"; $(GOIMPORTS) -l -local $(PKG) .; exit 1; fi; fi

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

release-preflight: validate-distribution ## Run non-publishing release validation checks
	@rm -f /tmp/icuvisor-goreleaser-check.log; \
	if $(GORELEASER) check > /tmp/icuvisor-goreleaser-check.log 2>&1; then \
		cat /tmp/icuvisor-goreleaser-check.log; \
	elif grep -q "configuration is valid, but uses deprecated properties" /tmp/icuvisor-goreleaser-check.log; then \
		cat /tmp/icuvisor-goreleaser-check.log; \
		echo "warning: local GoReleaser reports a deprecation for the existing Homebrew formula config; workflow-pinned GoReleaser remains authoritative"; \
	else \
		cat /tmp/icuvisor-goreleaser-check.log; \
		rm -f /tmp/icuvisor-goreleaser-check.log; exit 1; \
	fi; \
	rm -f /tmp/icuvisor-goreleaser-check.log
	npx --yes "$(MCPB_CLI_PACKAGE)" validate packaging/mcpb/manifest.json

validate-registry: ## Validate MCP Registry server.json metadata
	python3 scripts/validate_server_json.py server.json

validate-distribution: validate-registry ## Validate registry metadata and README install/deeplink CTAs
	python3 scripts/validate_readme_distribution.py

clean: web-clean ## Remove build artifacts (binary + site)
	rm -rf bin dist coverage.txt coverage.html

# ---- generated docs ----------------------------------------------------------

docs-tools: ## Regenerate website tool catalog and schema data from the tool registry
	$(GO) run ./cmd/gendocs --out web/data/tools.json

eval-validate: ## Validate cookbook eval scenarios against the tool catalog
	python3 scripts/eval/run_eval.py --validate

eval-tool-routing: ## Run opt-in first-tool routing smoke eval (provider skipped unless configured)
	$(GO) run ./scripts/toolroutingeval

# ---- website (web/, Hugo) ----------------------------------------------------

web-serve: ## Run the Hugo dev server for the icuvisor.app site (no search index)
	cd web && $(HUGO) server -D --port $(HUGO_PORT) --bind 127.0.0.1

web-preview: ## Build, index with Pagefind, and serve the site with working search
	cd web && $(HUGO) --minify --gc
	cd web && npx --yes pagefind --site public --serve

web-build: ## Build the Hugo site into web/public
	cd web && $(HUGO) --minify --gc

web-clean: ## Remove Hugo build output
	rm -rf web/public web/resources web/.hugo_build.lock

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
