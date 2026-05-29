# Plan Review — Step 2

Verdict: CHANGES REQUESTED

## Findings

### P1 — Hydrate the public-tool registration/catalog plan
Step 1 chose a new public `resolve_calendar_dates` tool, but Step 2 still only says “update catalog/schema snapshots if the public tool surface changes.” Before implementation, the plan should explicitly include all required registration surfaces: constructor + `coreTool` registration in `registryBaseTools`, `internal/toolcatalog` constant and athlete-scoped list, `toolCatalogGroup` placement, schema stability allowlist/snapshot generation, and the committed `internal/tools/schema_snapshot/resolve_calendar_dates.json`.

Without this, it is easy to land a tool that either fails registration (`not present in shared tool catalog`), defaults to the full toolset instead of being available in core planning conversations, or is missing from schema stability coverage.

### P1 — Pin the local-date arithmetic contract
The implementation plan needs to state the deterministic date math rules, not just “using athlete timezone.” Require deriving the default base date from an injected clock in the athlete timezone, parsing any `base_date` as an athlete-local `YYYY-MM-DD`, and applying day offsets with `time.AddDate(0, 0, offset)` in the loaded location rather than UTC or `24h` durations. Also specify validation for bad `base_date`, invalid/missing timezone fallback, negative/non-future offsets if out of scope, duplicate/too many offsets, and any max range.

The tests should include a DST/timezone-boundary case that would catch UTC/duration-based arithmetic.

### P2 — Define the response and schema shape before coding
The plan should pin the input and output contract enough for tests: expected arguments (`base_date`, `offsets` or equivalent), strict JSON decoding / `additionalProperties: false`, and response fields for each anchor (`offset_days`, `date`, `weekday`) plus `_meta` (`timezone`, `base_date`, `base_weekday`, `server_version`, count). This keeps the new deterministic surface stable and makes schema snapshot review meaningful.

## Verification

- Read `PROMPT.md` and `STATUS.md`.
- Reviewed existing `get_today`, registry/catalog, `toolcatalog`, and schema snapshot generation code.
- No tests run; this was a plan review.
