# R015 plan review — Step 4: `list_athletes` + `select_athlete`

Verdict: **APPROVE**

The revised Step 4 plan addresses the R014 blockers sufficiently to proceed. In particular, it now calls out the important architecture changes that make athlete switching viable: union registration in effective coach mode, active-athlete catalog filtering through one shared visibility helper, session-scoped selection state with process fallback, wiring the selected default into the Step 3 target resolver, config-backed roster responses, coach-mode-only registration, and cache-caveat metadata on `select_athlete`.

## Implementation notes to preserve

- Register athlete-scoped tools in coach mode by the union of tools allowed by at least one roster athlete **after** delete-mode and toolset gates; keep tools denied for every athlete unregistered.
- Filter all catalog views with one authoritative active-athlete helper: protocol `tools/list`, `icuvisor_list_advanced_capabilities`, `select_athlete.allowed_tools`, and any catalog hash/count metadata that claims to describe the visible catalog.
- Keep `tools/call` enforcement authoritative. A hidden/denied tool must still fail at call time for the selected athlete or any per-call `athlete_id` override, using the existing enumeration-safe target error.
- Implement the selection store as concurrency-safe and session-keyed when `req.Session.ID()` is available; initialize missing entries to `coach.default_athlete_id`; document the stdio/process fallback and any cleanup/leak rationale in `STATUS.md`.
- Keep `internal/tools` SDK-agnostic. Pass session/selection access through MCP-side context or handle `select_athlete` on the MCP side, but do not import the go-sdk into `internal/tools`.
- For this iteration, keep `list_athletes` config-backed only (`_meta.source: "config"`) because the authenticated upstream roster probe remains operator-deferred. Return canonical IDs, labels, stable ordering, count, default athlete, and active/session athlete; do not expose credentials or upstream assumptions.
- Compute `_meta.requires_new_conversation` by diffing previous vs. new visible catalog names, not merely by checking whether the athlete ID changed.

## Expected coverage

The Step 4 implementation should include or set up the tests already identified in R014: coach-mode-off absence, dynamic `tools/list` after selection, union-registered tools callable after selecting a broader-ACL athlete, per-call override routing/ACL checks, concurrent sessions under race, config-backed roster response shape, `select_athlete` response metadata, and active-athlete filtering for advanced capabilities.
