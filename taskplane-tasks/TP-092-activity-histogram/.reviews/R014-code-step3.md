# R014 code review — Step 3: Tool registration and activation hint

Verdict: APPROVE

## Findings

No blocking findings.

The Step 3 changes register `get_activity_histogram` through the shared `registryBaseTools` path as a `full` read tool, add it to the shared `toolcatalog` athlete-scoped/known-tool surface, map it to the activities catalog group, and update the explicit tier expectation. The MCP description leads with the single-activity distribution use case and includes the activation hint to use this tool instead of pulling `get_activity_streams` and binning manually. The generated catalog artifacts now include the new descriptor, resolving the stale-docs issue from R013.

## Verification run

- `go test ./internal/toolcatalog ./internal/tools ./cmd/gendocs` — PASS
- `go test ./internal/mcp` — PASS
