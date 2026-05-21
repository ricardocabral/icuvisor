# Plan Review R005 — Step 2: `icuvisor://workout-syntax`

**Verdict: APPROVE**

I read `PROMPT.md`, the updated `STATUS.md`, the prior R004 review, and the current resource/MCP/workoutdoc wiring. The Step 2 plan now has enough implementation detail to proceed and addresses the prior blockers: it defines a derived `workoutdoc` syntax descriptor, pins the resource contract, wires a default resource registry into normal server startup, scopes tests, and keeps other resources/tool-description trimming out of this step.

## Approval notes / guardrails

- Keep the resource Markdown generated from `internal/workoutdoc` data. The golden file should lock generated output; it must not become the source of truth.
- Make descriptor examples derive from `workoutdoc.Serialize` wherever practical. If any example text has to be literal, pair it with a serialize assertion so drift is caught.
- The parity test needs a real completeness check, not just snapshot text. At minimum, require every documented step form / target family / unit variant in the `workoutdoc` descriptor to have a representative fixture that serializes successfully and appears in the generated Markdown. If you want future serializer additions to fail tests automatically, the supported-family/unit tables need to live in `workoutdoc` rather than being an unconnected list in `internal/resources`.
- Document current behavior exactly, especially pace handling. `Target{Units: "PACE"}` serializes as plain numeric `... Pace`, while the parser treats generic tokens ending in `Pace` as text pace targets; the resource should not imply a richer semantic round-trip than the code supports.
- The app wiring is in scope for this step. `defaultStartServer` should pass the default resource registry to `mcp.NewServer`, and tests should verify the resource is visible through `resources/list`/`resources/read` in the production-style registry path, not only via an injected test registry.
- Deferring README/tool-description trimming to Step 6 is correct. If `CHANGELOG.md` is deferred too, make sure it is not forgotten before final acceptance.

No further plan changes are required before implementation.
