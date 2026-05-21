# Code Review — TP-007 Step 6

Verdict: **APPROVE**

## Findings

No blocking findings.

The `get_athlete_profile` handler now delegates response shaping through a focused helper that calls the shared `response.Shape` chokepoint with the expected `IncludeFull`, `ServerVersion`, `DebugMetadata`, `QueryType`, and profile-derived `UnitSystem` options. The shaped value is still used for both MCP text content and structured content, preserving the intended single-source response contract.

The added tests cover the main Step 6 integration deltas: explicit `include_full:false` matches the default terse response, `_meta.server_version` is present on structured output, and `_meta.units` reflects the profile-derived imperial unit system.

## Validation

- Ran `go test ./internal/tools ./internal/response` — passed.
- Ran `go test ./...` — passed.

## Non-blocking note

`STATUS.md` records `plan Step 6: APPROVE`, while the added `R001-plan-step6.md` file itself says the plan was not approved yet. This does not affect the code change, but the review/task bookkeeping may be confusing if those files are used as authoritative audit records.
