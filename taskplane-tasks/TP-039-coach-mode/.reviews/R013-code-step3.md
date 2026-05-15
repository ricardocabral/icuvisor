# R013 code review — Step 3: Tool registry plumbing

Verdict: **APPROVE**

I ran:

- `git diff a527c36..HEAD --name-only`
- `git diff a527c36..HEAD`
- `go test ./...`
- `go test -race ./internal/mcp ./internal/intervals ./internal/tools`

The current Step 3 implementation addresses the prior blocking findings:

- Coach ACL filtering now lives in `internal/mcp.safeRegistrar`, so coach-denied tools are omitted from the SDK catalog, exposed catalog hash inputs, and coach skip counts.
- `icuvisor_list_advanced_capabilities` is filtered from the authoritative MCP-side coach-visible catalog, so coach-denied tools are not discoverable through capability discovery.
- `athlete_id` schema injection/stripping is centralized and gated to effective coach mode, preserving coach-mode-off catalog compatibility.
- Request routing is context-scoped through `intervals.WithTargetAthleteID`, avoiding shared-client mutation.
- `get_activities` continuation tokens are bound to the resolved athlete.
- Direct activity-ID operations now verify ownership against the resolved target athlete before reads/writes/deletes.
- `LinkActivityToEvent` now preflights the target-scoped event before sending the activity update, and the regression test verifies the PUT is not sent after a failed event preflight.
- The resource bypass is handled for the production app path by disabling `icuvisor://athlete-profile` while coach mode is enabled, with the deferral rationale recorded in `STATUS.md`.

I did not find any remaining blocking issues for Step 3.
