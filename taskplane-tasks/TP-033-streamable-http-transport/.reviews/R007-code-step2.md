# Code Review: Step 2 — Streamable HTTP transport

**Verdict:** REVISE

I reviewed the diff from `efe31bcb7484d87d8945777d9a39b361fa0d5576..HEAD`, read the changed app/config/MCP files, and ran tests. `go test ./...` passes, but the race-check path required by the task does not.

## Blocking finding

1. **`TestDefaultStartServerWarnsForHTTPNonLoopbackBind` races on its log buffer.**  
   Location: `internal/app/app_test.go:268-292`

   The test starts `defaultStartServer` in a goroutine, wires `slog` to a plain `bytes.Buffer`, and polls `logs.String()` while the server goroutine is still writing startup/registration/listener logs. This is an actual data race under the required race verification:

   ```text
   WARNING: DATA RACE
   Write ... bytes.(*Buffer).Write()
     ... internal/app/app.go:165
     ... internal/app/app_test.go:277
   Previous read ... bytes.(*Buffer).String()
     ... internal/app/app_test.go:288
   --- FAIL: TestDefaultStartServerWarnsForHTTPNonLoopbackBind
       testing.go:1712: race detected during execution of test
   ```

   Repro command:

   ```sh
   go test -race ./internal/mcp ./internal/app ./internal/config
   ```

   This fails Step 2/Step 6’s `go test -race ./...` expectation. The adjacent `TestDefaultStartServerDispatchesHTTPTransport` already uses `safeAppLogBuffer`; use the same synchronized buffer here or otherwise avoid reading the buffer until the logging goroutine has stopped.

## Notes

- The HTTP transport wiring itself follows the approved plan shape: shared `NewServer`, `/mcp` handler via `NewStreamableHTTPHandler`, loopback protections left enabled, injected listener tests, and bounded shutdown.
- I did not find a production blocker in the Streamable HTTP serving path during this review; the requested revision is to make the new/modified tests race-clean.
