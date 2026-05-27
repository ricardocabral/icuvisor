# Code Review: Step 2 — Implement opt-in first-tool-call runner

**Verdict: Needs changes**

## Findings

1. **Blocking: lint currently fails on the new runner.**  
   `golangci-lint run ./internal/toolrouting ./scripts/toolroutingeval ./internal/mcp` reports:
   - `internal/toolrouting/runner.go:26` — `gosec G101` on `EnvAnthropicAPIKey`.
   - `internal/toolrouting/runner.go:192` — unnecessary `json.RawMessage(...)` conversion.
   - `internal/toolrouting/runner.go:223` — staticcheck `ST1005` for capitalized error string.  
   These will block the required quality gate. Add a justified `#nosec G101` or rename/refactor the env-var constant, remove the redundant conversion, and lowercase the error text.

2. **Provider calls are not deterministic enough for a smoke eval.**  
   The Anthropic request built at `internal/toolrouting/runner.go:184` does not set `temperature`, and `anthropicRequest` has no field for it. The step-2 plan called for repeatable request bounds; leaving the provider default makes pass/fail routing results potentially flaky. Set temperature to `0` in the request payload and cover it in the request-construction test.

## Verification

- `go test ./internal/toolrouting ./internal/mcp` passes.
- `go run ./scripts/toolroutingeval` validates/skips without provider credentials as expected.
- Targeted `golangci-lint` command above fails with the issues listed.
