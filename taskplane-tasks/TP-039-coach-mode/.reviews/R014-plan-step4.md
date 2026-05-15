# R014 plan review — Step 4: `list_athletes` + `select_athlete`

Verdict: **REVISE**

The Step 4 outline has the right high-level tools, but it is not yet specific enough to implement safely on top of the Step 3 registry design. The main gap is that Step 3 registered a catalog filtered for the default athlete. Once `select_athlete` can switch to an athlete with a different ACL, that default-filtered registration becomes either too narrow (newly allowed tools are protocol-unknown) or too broad (previously allowed tools remain visible/callable unless every path is dynamically filtered). The Step 4 plan must pin down the session-state and catalog-filtering design before coding.

## Required revisions

1. **Change coach-mode registration from “default athlete only” to an authoritative union plus per-session filtering.**
   In coach mode, the SDK must register every tool allowed by delete-mode/toolset and by at least one configured athlete, plus the non-athlete coach tools. Then:
   - `tools/list` must be filtered for the active session athlete before returning to the client.
   - `tools/call` must still enforce the target athlete ACL at call time, including per-call `athlete_id` overrides.
   - tools not allowed for any roster athlete should remain unregistered.

   Without this, selecting an athlete with a broader ACL than the default cannot work, because the SDK will return `unknown tool` for tools that were skipped during startup.

2. **Use a single visibility function for all catalog views.**
   The plan should introduce one authoritative helper that composes delete-mode, toolset, and coach ACL for a given active athlete and is reused by:
   - SDK `tools/list` filtering,
   - `icuvisor_list_advanced_capabilities`,
   - the `select_athlete` response’s `allowed_tools`,
   - any catalog hash/count metadata that claims to describe the visible catalog.

   Do not duplicate default-athlete filtering in `safeRegistrar`, the advanced-capabilities handler, and the new tools independently; that will drift and can leak denied tool names.

3. **Define the session selection store explicitly.**
   The go-sdk request exposes `req.Session`, so the plan should use real session-scoped state where possible, not process-global state by default. Specify:
   - keying by `req.Session.ID()` for Streamable HTTP sessions;
   - a documented process/stdin fallback only when the SDK session ID is empty;
   - a mutex or `sync.Map`-backed store safe for concurrent Streamable HTTP calls;
   - initialization to `config.Coach.DefaultAthleteID`;
   - cleanup strategy if feasible, or an explicit bounded/leak rationale in `STATUS.md`.

4. **Wire selected-athlete state into the existing target resolver.**
   `safeRegistrar.resolveAthleteID` currently defaults to `config.Coach.DefaultAthleteID`. Step 4 must revise it to default to the selected athlete for the current session, while preserving per-call `athlete_id` override semantics. The resolver should continue to normalize once and return the existing enumeration-safe public target error for malformed, unknown, or disallowed targets.

5. **Decide how `select_athlete` gets session context without importing the SDK into `internal/tools`.**
   A normal `tools.Handler` only receives `context.Context` and raw arguments. The plan should state whether the MCP wrapper will put a session key/selection store into context, or whether `select_athlete` is handled as an MCP-side special case. Keep `internal/tools` SDK-agnostic.

6. **Gate and register the new tools only in effective coach mode.**
   `list_athletes` and `select_athlete` must be absent when `ICUVISOR_COACH_MODE` is effectively off. Since `tools.NewRegistryWithOptions` currently does not receive coach config, the plan must specify the dependency injection change (for example, registry options carrying normalized coach config and a selection-store interface) or an MCP-side registration path. Do not rely on the LLM seeing the tools and failing at call time.

7. **Implement `list_athletes` from config only for this task iteration.**
   `STATUS.md` records the authenticated upstream roster probe as blocked/operator-deferred. The Step 4 plan should therefore avoid adding an unvalidated intervals.icu roster call now. Return the normalized config roster with `_meta.source: "config"`, stable ordering, count, default athlete, and active/session athlete. Leave `_meta.source: "upstream"` for a later task after a real coach-key probe.

8. **Specify `select_athlete` response and cache caveat behavior.**
   The response should include previous selection, new selection, the new athlete’s visible tool names, `_meta.scope` (`"session"` or the documented fallback), and `_meta.requires_new_conversation`. Compute `requires_new_conversation` by comparing the previous and new visible catalog names, not merely by checking that the athlete ID changed.

9. **Keep public errors enumeration-safe and credential-free.**
   `select_athlete` should use the same public rejection text for malformed IDs and IDs absent from the roster. Neither tool should accept or return API keys, and logs should not include raw athlete identifiers.

## Required tests for the revised plan

- Coach mode off: `list_athletes` and `select_athlete` are absent from `tools/list` and protocol calls return unknown-tool errors.
- Coach mode on with two athletes whose ACLs differ: initial `tools/list` reflects the default athlete; after selecting the second athlete, a fresh `tools/list` is filtered for the second athlete.
- A tool denied for the default but allowed for another athlete is registered and callable after selecting that athlete (proves union registration), while a tool denied for the selected/overridden target returns the enumeration-safe public target error.
- Per-call `athlete_id` override still routes to and ACL-checks the override athlete even after a different session default was selected.
- Two concurrent sessions can select different athletes without cross-routing or cross-filtering; run the relevant tests under `go test -race`.
- `list_athletes` returns canonical `i12345` IDs, labels, `_meta.source: "config"`, count, default athlete, and active athlete.
- `select_athlete` returns previous/new selections, visible allowed tools, `_meta.scope`, and `requires_new_conversation` true only when the visible catalog changes.
- `icuvisor_list_advanced_capabilities` follows the active session athlete and does not leak full-only tools denied for that athlete.

Once the plan addresses these details, Step 4 should be safe to implement.
