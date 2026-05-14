# Plan Review: Step 3 — Security posture

**Verdict:** REVISE

I read `PROMPT.md`, `STATUS.md`, the PRD/roadmap transport anchors, and the current Step 1/2 implementation in `internal/config`, `internal/app`, and `internal/mcp`. The Step 3 entry in `STATUS.md` currently restates the task checklist, but it is not yet a concrete implementation plan for a security-sensitive step.

## Blocking gaps

1. **Specify the default-bind test case precisely.**
   The plan should say where and how the loopback-default invariant will be tested. In particular, cover the HTTP-mode default path: load config with `ICUVISOR_TRANSPORT=http` and no `ICUVISOR_HTTP_BIND`, then assert the resulting bind is `config.DefaultHTTPBindAddress`, parses as a loopback IP, and is not wildcard/non-loopback. Avoid starting the fixed default port in tests; a config-level assertion is less flaky unless a listener seam is added.

2. **Define the log redaction audit/test, not just the goal.**
   “No API keys or athlete IDs in HTTP logs” needs an explicit test plan. Capture HTTP startup/lifecycle logs with a race-safe buffer, use sentinel values such as a fake API key and `i12345`, run the HTTP transport through start/listen/shutdown (and, if practical, one malformed HTTP request), and assert those sentinel values never appear. Also state that no HTTP access-log middleware or request/response/header/body logging will be added, and that SDK logger usage has been checked to avoid logging payloads or headers.

3. **Account for race-safe log capture.**
   Existing HTTP startup tests poll log output while the server goroutine is active. Any Step 3 log-leak test should explicitly use the synchronized buffer pattern already present in the tests, not a plain `bytes.Buffer`, otherwise the required `go test -race ./...` path can fail.

4. **Place the LAN-bind threat-model documentation.**
   The README work should not be left as an abstract checkbox. The plan should name the section to update and the required wording: default Streamable HTTP bind is loopback-only; setting `ICUVISOR_HTTP_BIND`/`--http-bind` to a LAN address exposes an unauthenticated MCP server; anyone who can reach that address can call tools using the configured intervals.icu credentials; opt in deliberately.

## Suggested STATUS.md additions

Add a short Step 3 plan under the checklist, for example:

- Add/extend `internal/config` tests for HTTP transport with no bind override, asserting the canonical default is loopback and not wildcard.
- Add/extend `internal/app` or `internal/mcp` HTTP log tests with a synchronized log buffer and sentinel secret/athlete values, covering startup, non-loopback warning, listener log, and shutdown; assert no API key or raw athlete ID appears.
- Do not introduce request/response logging; verify the SDK Streamable HTTP logger only receives lifecycle/spec/write-warning messages, not headers or JSON-RPC bodies.
- Update README transport/configuration docs with the LAN-bind no-auth threat model and opt-in warning.

Once those details are captured in `STATUS.md`, the Step 3 plan should be ready to implement.
