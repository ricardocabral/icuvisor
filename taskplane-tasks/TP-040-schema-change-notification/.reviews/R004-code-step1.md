# Review R004 — Code review for Step 1: Catalog hash

Verdict: **Approved**

## Findings

No blocking findings for Step 1.

The implementation hashes the exposed post-filter tool catalog, uses deterministic sorted JSON records, includes the LLM-facing tool description plus input/output schemas, stores the computed hash on `mcp.Server`, and exposes `Server.CatalogHash()`. The tests cover the requested determinism and sensitivity cases, including capability/toolset filtering.

## Validation performed

- `git diff --check 3c649f966c431ad5d760e265f53a99168f3fa6c6..HEAD` — passed.
- `go test ./internal/mcp -run 'TestHashToolCatalog|TestNewServerCatalogHashUsesExposedCatalog' -count=1` — passed.
- `go test ./internal/mcp -run TestProtocolTransportParity -count=1` — passed.
- `go test ./internal/mcp -count=1 -parallel 1` — passed.
- `golangci-lint run ./...` — failed only on the existing `gosec` warning in `internal/config/config.go:260` noted in the prior review, outside this Step 1 diff.

## Notes

- Full parallel package/suite runs intermittently failed in existing Streamable HTTP shutdown tests (`TestProtocolTransportParity` / `TestServeStreamableHTTPInitializesClient`). I could not tie those failures to this Step 1 change: the new catalog-hash tests do not start transports, and the affected tests pass individually and with package parallelism disabled.
