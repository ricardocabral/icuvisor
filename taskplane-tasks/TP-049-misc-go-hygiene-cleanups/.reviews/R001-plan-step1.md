# Plan Review: Step 1 — Recover helper

Verdict: Approved with minor implementation notes.

The Step 1 plan matches the task scope: extract the three near-identical `defer recover()` blocks in `internal/mcp/server.go` into a single helper with a short rationale comment, without changing protocol behavior or widening the cleanup.

Implementation notes to keep the change safe:

- Preserve the current error context strings at the call sites:
  - `constructing MCP server: ...`
  - `registering tool "<name>": ...`
  - `registering resource "<uri>": ...`
- Keep recovery scoped to the SDK boundary calls that currently have recovery (`sdkmcp.NewServer`, `Server.AddTool`, `Server.AddResource`). Do not use this step to add broad runtime recovery around tool/resource handlers or unrelated code paths.
- If the helper signature is `withPanicRecovery(name string, fn func() error) error`, make sure it returns `fn()` errors unchanged/wrapped appropriately and only converts recovered panics to errors.
- In `AddTool` / `AddResource`, keep registration bookkeeping (`registeredTools`, counts) after the helper returns successfully so a recovered SDK panic does not mark the item as registered.
- Consider adding a small same-package unit test for the helper covering: no panic, returned error, and panic conversion. This is not explicitly required by the prompt, but it is cheap and would lock in the helper behavior.
- Run `gofmt` and the requested checks for the step (`make build`, tests, race, lint) before committing.

No blockers found.
