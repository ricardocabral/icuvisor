# Code Review: Step 3 — Security posture

**Verdict:** REVISE

I reviewed the diff from `e251058ed4976d135e50d86c4f306fececab879d..HEAD` and ran:

- `go test ./internal/config ./internal/mcp ./internal/app` — passes
- `go test -race ./internal/mcp ./internal/app` — fails with a data race in the HTTP startup warning log test

## Findings

### 1. Race detector still fails in the HTTP security logging path

**File:** `internal/app/app_test.go:269` / `internal/app/app_test.go:288`

Step 3 explicitly added a race-safe HTTP log-redaction requirement, but the current test suite still polls a plain `bytes.Buffer` while `defaultStartServer` writes logs from another goroutine. `go test -race ./internal/mcp ./internal/app` reports concurrent `bytes.Buffer.Write` and `bytes.Buffer.String` in `TestDefaultStartServerWarnsForHTTPNonLoopbackBind`.

This is directly in the Step 3 security posture area because that test is the one covering the non-loopback LAN-bind warning and its redaction behavior. Please switch this test to the synchronized log-buffer pattern already introduced in `internal/mcp/server_test.go` (or an equivalent local helper), and avoid unsynchronized `logs.String()` while the server goroutine can still write.

### 2. The new redaction test does not actually cover the non-loopback warning/config path

**File:** `internal/mcp/server_test.go:422`

`TestServeStreamableHTTPLogsDoNotLeakConfigOrPayload` is useful for the malformed-request/body case, but the `config.Config{APIKey: ..., AthleteID: ...}` passed to `mcp.NewServer` is not used by `ServeStreamableHTTP` logging; `NewServer` only uses `Config` for toolset selection. As a result, this test proves the malformed request body is not logged, but it does not prove that the HTTP startup/non-loopback warning path avoids leaking configured API keys or athlete IDs.

The existing app-level warning test is the right behavioral layer for that coverage, but it is currently racy. Please make that app-level warning test race-safe and keep/assert the forbidden API key and athlete ID there. That will satisfy the Step 3 plan: startup/listen/shutdown plus non-loopback warning without leaking configured secrets.

## Notes

- The README LAN-bind threat model update is clear and matches the task requirement.
- The new config test correctly verifies that `ICUVISOR_TRANSPORT=http` with no bind override produces `DefaultHTTPBindAddress` and that it is loopback.
