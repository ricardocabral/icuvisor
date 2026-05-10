# Plan Review — TP-003 Step 1: SDK spike and plan

Verdict: **changes requested before moving to Step 2**.

## What looks good

- The SDK version choice is sound. I verified `github.com/modelcontextprotocol/go-sdk@v1.3.1` declares `go 1.23.0`, while later stable releases move to Go 1.24/1.25, so pinning v1.3.1 preserves the repository's current `go 1.23` baseline.
- The recorded production/test transport split is correct: `mcp.StdioTransport` uses `os.Stdin`/`os.Stdout`, and `mcp.IOTransport` / `mcp.NewInMemoryTransports` are the right paths for tests that must not require Claude Desktop.
- The SDK license note is accurate for the pinned version: v1.3.1 is MIT.

## Blocking findings

1. **The registry interface is still undefined.**
   Step 1's explicit deliverable is to define the tiny internal registry interface, but `STATUS.md` still has that item unchecked and does not record the proposed shape. Before implementation starts, record the contract that `internal/mcp` will consume and `internal/tools` will implement. At minimum, the plan should say whether the registry exposes a `Register(...) error` callback, returns a slice of tool definitions, and whether SDK types are allowed to leak into `internal/tools`.

2. **The chosen `mcp.AddTool` path needs a panic/error strategy.**
   In v1.3.1, `mcp.AddTool` panics on schema inference/validation errors, and `(*mcp.Server).AddTool` can also panic on invalid/missing schemas. The project has a hard rule of no panics outside `main`, and TP-003 requires startup to return errors instead of panicking. The plan should explicitly require a small adapter/registrar that either prevalidates and/or recovers those SDK panics and converts them to startup errors before Step 2/3 use `AddTool` directly.

3. **License review should include transitive modules before adding the dependency.**
   The task calls out the SDK license, but the repository hard rule is to check the license of every new module. Adding the SDK will pull in transitive modules (`github.com/google/jsonschema-go`, `github.com/segmentio/encoding`, `github.com/golang-jwt/jwt/v5`, `golang.org/x/*`, etc.). Record a license scan/result in `STATUS.md` before or during Step 2 so the no-GPL-dependencies rule is satisfied.

## Suggested plan adjustment

Update `STATUS.md` with a concrete Step 1 decision such as:

- Pin `github.com/modelcontextprotocol/go-sdk` to `v1.3.1` for Go 1.23 compatibility.
- `internal/mcp` owns all SDK adaptation and exposes a constructor/run path that accepts a registry dependency plus an injectable transport for tests.
- `internal/tools` defines only a minimal registry contract; real tool logic remains out of scope for TP-003.
- Tool registration goes through a safe registrar that returns errors for invalid names/schemas and prevents SDK panics from escaping.
- Dependency license review includes the SDK and all newly added transitive modules.

Once those are recorded, the plan is ready to proceed to Step 2.
