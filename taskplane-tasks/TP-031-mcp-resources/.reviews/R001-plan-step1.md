# Plan Review R001 — Step 1: Resource registration plumbing

**Verdict: Needs revision before implementation.**

I read `PROMPT.md`, `STATUS.md`, PRD §7.2.G/KR5/§7.4 #13, and the existing MCP/tool registration code. `STATUS.md` currently only repeats the step checklist; it does not contain an implementation plan to review. Please add a concrete Step 1 plan before coding, because this is the first Resource integration and the SDK details matter.

## Required plan additions

1. **Use the SDK Resource API, not custom JSON-RPC handlers.**
   The project is on `github.com/modelcontextprotocol/go-sdk v1.4.1`; resources should be registered with `(*mcp.Server).AddResource(...)`. The SDK then provides `resources/list`, `resources/read`, capabilities, pagination, and list-change behavior. A plan to manually implement `resources/list` / `resources/read` would be a red flag.

2. **Define the registry boundary explicitly.**
   Mirror the tool pattern with a small internal registry/registrar interface, but avoid coupling domain resources directly to SDK types. A good shape would keep SDK conversion in `internal/mcp` and resource definitions in `internal/resources` (or clearly justify keeping it all in `internal/mcp`). Include validation for:
   - absolute `icuvisor://...` URIs,
   - duplicate URIs,
   - non-empty name/title/description/MIME type,
   - nil handlers/content readers.

3. **Wire the registry through server construction.**
   The plan should state how `mcp.Options` will accept a resource registry and how `internal/app` will eventually pass the default registry alongside the tool registry. Resources must be added before the MCP session initializes so the SDK advertises `resources` capability during `initialize`.

4. **Specify handler error policy.**
   Tool errors are sanitized today; raw Resource handler errors can become protocol errors. The plumbing plan should define how resource read failures are logged and converted to short, safe client-facing errors, while preserving SDK `ResourceNotFoundError` behavior for unknown/unavailable resources.

5. **Pin the content contract for Step 1.**
   State that resource reads return `ReadResourceResult` with one text `ResourceContents` item, populated `URI`, `MIMEType`, and `Text`. Pick default MIME types now (`text/markdown` for long-form docs is likely; `application/json` may be appropriate for `athlete-profile` later) or document that the per-resource steps will set them.

6. **Document static vs dynamic decisions in `STATUS.md`.**
   Step 1 requires this. The plan should record at least:
   - `icuvisor://workout-syntax`: static/derived from `internal/workoutdoc`, no hand-authored duplicate grammar.
   - `icuvisor://event-categories`: static from the same enum/source used by event tools.
   - `icuvisor://custom-item-schemas`: static/derived from the validation/schema source used by custom-item tools.
   - `icuvisor://athlete-profile`: dynamic cached resource with TTL/staleness policy to be finalized in Step 5.

7. **Add protocol-level tests in the plan.**
   Step 1 should include tests using the existing in-memory MCP client helpers to assert:
   - `InitializeResult` advertises resources when a resource registry is configured,
   - `ListResources` returns registered resource metadata,
   - `ReadResource` dispatches and returns the expected text/MIME/URI,
   - duplicate/invalid resource registration returns a `NewServer` error instead of panicking,
   - unknown resource reads produce the SDK not-found protocol error.

8. **Record the canonical SDK docs link in `STATUS.md`.**
   This is explicitly requested by the task context. The plan should include the source consulted for the v1.4.1 Resource API.

## Suggested Step 1 scope

A safe Step 1 is to implement only the plumbing plus a tiny test-only resource registry; do not implement the four real resources yet. Registering placeholder production resources in Step 1 would risk violating later requirements around derivation, golden files, and athlete-profile caching.

Once the plan includes the above details, the implementation should be straightforward and low-risk.
