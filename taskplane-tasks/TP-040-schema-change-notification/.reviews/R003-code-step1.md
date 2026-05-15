# Review R003 — Code review for Step 1: Catalog hash

Verdict: **Request changes**

## Findings

1. **Formatting is not gofmt-clean, so lint fails.**
   - `internal/mcp/server.go:51`
   - `internal/mcp/catalog_hash_test.go:73`
   - `golangci-lint run ./...` reports both changed Go files as not properly formatted. This violates the repository's gofmt/goimports rule and will fail CI. Running `gofmt -w internal/mcp/server.go internal/mcp/catalog_hash_test.go` fixes the Go formatting.

## Validation performed

- `go test ./internal/mcp` — passed.
- `go test ./...` — passed.
- `golangci-lint run ./...` — failed on the gofmt issues above, plus an existing `gosec` warning in `internal/config/config.go:260` that is outside this Step 1 diff.
- `git diff --check 3c649f966c431ad5d760e265f53a99168f3fa6c6..HEAD` — failed because the two prior review artifacts have trailing blank lines at EOF:
  - `taskplane-tasks/TP-040-schema-change-notification/.reviews/R001-plan-step1.md:48`
  - `taskplane-tasks/TP-040-schema-change-notification/.reviews/R002-plan-step1.md:33`

## Notes

The catalog-hash implementation itself matches the Step 1 plan: it hashes the actually registered tool catalog after safety/toolset filtering, uses sorted framed JSON records, includes tool description/input/output schema, exposes `Server.CatalogHash()`, and has determinism/sensitivity coverage. After formatting cleanup, I do not see a functional blocker in the Step 1 code.
