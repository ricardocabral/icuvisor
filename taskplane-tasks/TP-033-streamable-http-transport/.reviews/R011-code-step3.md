# Code Review: Step 3 — Security posture

**Verdict:** APPROVE

I reviewed the diff from `e251058ed4976d135e50d86c4f306fececab879d..HEAD`, read the changed test/docs files plus the relevant app/config/MCP transport implementation, and ran:

- `go test ./internal/config ./internal/mcp ./internal/app` — passes
- `go test -race ./internal/mcp ./internal/app` — passes
- `go test -race ./...` — passes

## Findings

No blocking findings.

## Notes

- The HTTP-mode default bind test now covers `ICUVISOR_TRANSPORT=http` with no bind override and asserts the canonical default remains loopback-only.
- The app-level non-loopback warning test now uses a synchronized log buffer and preserves the secret/athlete-ID leak assertions on the path that actually has configured credentials.
- The MCP HTTP log test covers lifecycle logging plus a malformed request body containing sentinel values without introducing request/header/body access logging.
- The README security note clearly states that LAN binding is opt-in and exposes an unauthenticated MCP server using the configured intervals.icu credentials.
