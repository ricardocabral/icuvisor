# Code Review: Step 3 — Security posture

## Verdict

APPROVE

## Findings

No blocking findings.

## Notes

- `git diff fcff149..HEAD --name-only` and `git diff fcff149..HEAD` are empty, so there are no committed Step 3 code deltas beyond the baseline commit. This review audited the current tree against the Step 3 checklist.
- The default HTTP bind remains `127.0.0.1:8765` in `internal/config`, and config tests assert both the default stdio path and explicit HTTP mode produce a loopback bind.
- HTTP bind validation rejects omitted-host wildcard binds such as `:8765`; non-loopback explicit binds are detected via `HTTPBindAddressIsLoopback` and produce a structured WARN in `internal/app` without API keys or athlete IDs.
- HTTP transport logs in `internal/mcp` are limited to startup/shutdown metadata (`version`, `transport`, listener `address`, path) in the audited paths. Redaction tests cover malformed HTTP traffic and assert no API key or athlete ID appears in logs.
- README documents the loopback default and the LAN-bind threat model: Streamable HTTP has no auth in this task, and anyone who can reach the LAN-bound address can invoke registered tools with the configured intervals.icu credentials.

## Verification

- `go test ./internal/config ./internal/mcp ./internal/app` — passed (cached).
