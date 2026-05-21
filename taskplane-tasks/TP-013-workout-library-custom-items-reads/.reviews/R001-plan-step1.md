# R001 plan review — Step 1

Decision: **APPROVE**

The Step 1 plan is appropriately scoped to the workout-library read side. It anchors the implementation to the documented `folders` and `workouts` reads, avoids inventing an unsupported folder-workouts endpoint by filtering locally, keeps workout-library writes and upload-DSL serialization out of scope, and includes registry/docs/change-log wiring plus focused coverage for empty libraries, nesting, top-level workouts, folder filtering, and `include_full` preservation.

No blocking findings.

## Follow-up notes for implementation

- Be careful with API path construction: `DefaultAPIBaseURL` already includes `/api/v1`, so the intervals client methods should follow existing patterns (`doJSON(ctx, ..., "athlete", c.athleteID, "folders")`, etc.) rather than double-prefixing `/api/v1`.
- Define the `get_workout_library` optional top-level-workout behavior explicitly in the input schema, e.g. a boolean such as `include_top_level_workouts`. Top-level workout rows must remain terse by default; do not expose raw workout payloads or `workout_doc` unless an explicit `include_full:true` path is provided and documented.
- Normalize folder and workout IDs defensively. Upstream IDs and `folder_id` may arrive as strings or numbers, so compare using a shared stringification helper and reject an empty requested `folder_id` with a short `UserError`.
- Keep the structured-step summary shallow and non-transforming: count/summarize exposed `workout_doc` fields where available, but do not parse into or serialize from the intervals.icu workout DSL in this read-only step.
- Bound response size for `get_workouts_in_folder` after local filtering. If full pagination is not implemented for this low-traffic surface, include at least a deterministic default/max limit plus `_meta.total_count`, returned count, and truncation metadata so the tool remains terse-by-default.
- Tests should cover both layers: `httptest.Server`/fixture tests for the new intervals client routes and tool-shaping tests proving default responses omit raw `workout_doc` while `include_full:true` preserves the nested shape.
