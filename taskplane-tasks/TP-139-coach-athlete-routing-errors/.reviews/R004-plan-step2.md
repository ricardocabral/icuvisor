# Plan Review — Step 2

Result: **approve**

The Step 2 plan is appropriately scoped for the findings from Step 1. It covers the client-facing MCP routing path, the coach package target-resolution behavior, `select_athlete`, local-mode targeting, ACL/catalog regressions, and the targeted package test set needed for this hardening.

Execution guidance for the worker:

- Prefer stable error values/types in `internal/coach`/routing code and map them once to public `UserError` messages; avoid branching on `err.Error()`.
- Add protocol-level tests that assert denied/invalid routed calls make no upstream intervals.icu request, especially for unauthorized roster targets, selected-athlete ACL denial, and local-mode `athlete_id` rejection.
- Include a real config-load/normalization regression for both `i123` and bare numeric IDs so numeric Strava-linked IDs are preserved and not rewritten with `i`.
- Keep local mode behavior centralized: no `athlete_id` schema injection, reject model-supplied `athlete_id` before tool handlers/upstream calls, but continue using the configured local athlete when no target is supplied.
- For catalog/ACL checks, cover both visibility (`tools/list` hides disallowed tools for the active athlete) and stale-catalog calls (registered but currently disallowed tools return the explicit ACL-denial message).
- Schema/API-key regression should inspect registered tool schemas in coach and local mode for forbidden credential-like parameters, not just rely on unknown-field decoding.

No plan changes are required before implementing Step 2.
