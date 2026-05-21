# Plan Review — TP-080 Step 2

Verdict: **REVISE**

The Step 2 plan is directionally aligned with the task: HR should mirror the duration-axis power curve, pace should use distance buckets and preferred pace units, and both tools belong in the full fitness catalog. However, two implementation-critical details are missing from the plan and should be made explicit before coding.

## Blocking plan gaps

1. **Update the shared tool catalog, not only `internal/tools/catalog.go`.**
   - `defaultRegistry.Register` rejects tools that are not in `internal/toolcatalog` via `toolcatalog.IsKnownTool`.
   - The new tools are athlete-scoped read tools, so the plan should explicitly include adding `GetHRCurves` / `GetPaceCurves` constants and entries in `athleteScopedToolNames`, with corresponding `internal/toolcatalog`/registry/coach-ACL test expectations as needed.
   - This also ensures MCP coach-mode schema injection treats them as athlete-scoped tools.

2. **Define the pace default bucket contract.**
   - The plan says `get_pace_curves` uses `distance_meters`, but it does not specify default distance buckets or how they are normalized.
   - To satisfy “same response conventions” and terse-by-default behavior, the plan should name the default `distance_meters` set (or explicitly state that returned upstream buckets are used and why that remains terse), and confirm positive/sorted/dedup normalization plus `missing_buckets` metadata just like power/HR.
   - The schema description should document those defaults in meters.

## Non-blocking clarification

- For `get_pace_curves`, keep the intended point shape explicit: `distance_meters`, upstream/raw `elapsed_seconds`, exactly one preferred pace field (`pace_seconds_per_km` or `pace_seconds_per_mile`), `activity_id`, and `_meta.units`. Also state the profile failure/fallback behavior if athlete profile lookup fails or has unknown unit preferences.

Once these are added to `STATUS.md`, the plan should be ready for implementation.
