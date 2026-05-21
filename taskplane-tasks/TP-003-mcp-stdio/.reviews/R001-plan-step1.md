# Plan Review — TP-003 Step 1: SDK spike and plan

Verdict: **approved to proceed to Step 2**.

This review supersedes the earlier Step 1 revision request. The current `STATUS.md` now records the SDK choice, version limitation, license check, registry boundary, panic-to-error strategy, and transitive license scan that were missing before.

## What I verified

- `the MCP Go SDK@v1.3.1` declares `go 1.23.0`, matching the repository's current `go 1.23` baseline. The latest checked `v1.6.0` declares `go 1.25.0`, so pinning v1.3.1 is the right compatibility decision unless this task intentionally raises the project Go version.
- The selected SDK APIs exist in v1.3.1: `mcp.NewServer`, `mcp.AddTool`, `(*mcp.Server).Run`, `mcp.StdioTransport`, `mcp.IOTransport`, `mcp.NewInMemoryTransports`, `mcp.NewClient`, `ClientSession.ListTools`, and `ClientSession.CallTool`.
- The recorded transport limitation is accurate: `mcp.StdioTransport` is tied to process stdio; `IOTransport` or `NewInMemoryTransports` are the appropriate test seams.
- The SDK license for v1.3.1 is MIT, and the recorded transitive dependency licenses are permissive. No GPL/copyleft dependency is planned.
- The revised registry decision is aligned with the project layout: `internal/tools` owns SDK-free tool contracts, and `internal/mcp` owns adaptation to SDK types.
- The revised panic strategy addresses a real SDK hazard: `mcp.AddTool` and `(*mcp.Server).AddTool` can panic for invalid schemas, so routing all registration through a validating/recovering adapter is necessary to satisfy the repository's no-panic-outside-`main` rule.

## Non-blocking guidance for Step 2/3

- Keep stdout protocol-only. Any SDK/app logging used while stdio is running should go to stderr or a discard/test logger, never stdout.
- Make the registry contract concrete when implemented: define the `InputSchema` representation and handler/result shape explicitly enough that `internal/tools` does not need to import the MCP SDK.
- Apply the same panic-to-error discipline around SDK construction options as around tool registration, even if the constructor currently uses only trusted constants.
- Do not add Streamable HTTP/SSE, real intervals.icu calls, or `get_athlete_profile` behavior in this task; keep TP-003 to the stdio server and registry scaffolding.

No blocking issues remain in the Step 1 plan.
