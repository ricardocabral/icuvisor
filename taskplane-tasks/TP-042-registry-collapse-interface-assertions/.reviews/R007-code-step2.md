# Code Review — Step 2: Refactor `Register`

**Verdict: request changes.**

The registry refactor itself follows the intended direction, and `go test ./...` passes. However, the tree does not pass the repository lint gate after this step.

## Findings

### P1 — New/edited files are not gofmt-formatted

`make lint` fails with gofmt errors in:

- `internal/mcp/protocol_registry_test_helpers_test.go:21`
- `internal/toolchecks/schema_stability.go:52`
- `internal/tools/registry_test_helpers_test.go:25`

These are all files touched/added by this step. CI requires gofmt/goimports-clean diffs, so please run formatting on the changed Go files.

### P1 — Legacy registry fake types are now unused and fail lint

After switching registry tests to real no-network `*intervals.Client` instances, the old fake clients are left behind but no longer referenced. `make lint` reports unused symbols for:

- `internal/mcp/protocol_test.go:33` — `advancedProtocolClient` and its method
- `internal/tools/catalog_tiers_test.go:92` — `fullCatalogTierClient` and its methods
- `internal/tools/list_advanced_capabilities_test.go:113` — `staticCatalogPanicClient` and its method

Because `golangci-lint` treats these as errors, the branch cannot pass CI as-is. Remove these stale fakes, or keep only any pieces that are still intentionally referenced by tests.

## Verification

Commands run:

- `git diff 01d92c37b10a322c045476eb08e82d7a573c51e7..HEAD --name-only`
- `git diff 01d92c37b10a322c045476eb08e82d7a573c51e7..HEAD`
- `go test ./...` — passes
- `make lint` — fails with the gofmt and unused-symbol errors listed above
