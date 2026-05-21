# Plan Review — Step 1: Inventory + capture regression-gate tests

Decision: **Approved**

The revised Step 1 plan in `STATUS.md` addresses the R001 blockers. It now contains a declaration-by-declaration inventory for every current top-level declaration in `internal/mcp/server.go`, a concrete coach ACL regression gate, the coach/tools import-cycle constraint for Step 3, and notes about existing split-adjacent files to avoid duplicate/stale registrars.

## What I checked

- Compared the inventory against the current top-level declarations in `internal/mcp/server.go`; the listed declarations match the current file, including the less obvious helpers such as `withPanicRecovery`, `newSDKServer`, `firstSentence`, `toolRequirement`, `convertResult`, and `convertContent`.
- Verified the named protocol tests exist in `internal/mcp/protocol_test.go`.
- Verified the supporting advanced-capabilities, coach evaluator/config, and toolcatalog tests exist.
- Verified the recorded import-cycle risk is real: `internal/tools` currently imports `internal/coach`, so Step 3 must not make `internal/coach` import `internal/tools`.

## Notes for the next steps

- The targeted coach gate is good for visibility/ACL work. For mechanical registrar moves, continue using the broader command recorded in the plan (`go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog`) so the non-coach capability/toolset filtering tests in `internal/mcp` also remain covered.
- The status note about `internal/mcp/prompts.go` is important: Step 2 should either leave it alone or intentionally rename it, not create a second prompt registrar implementation.
- Keep the Step 3 seam acyclic and byte-compatible. In particular, avoid exposing `tools.Tool` from `internal/coach` unless the dependency direction is changed elsewhere, which would be a larger refactor than this task calls for.

No plan changes are required before proceeding to Step 2.
