# Code Review R011 — Step 3: `icuvisor://event-categories`

**Verdict: APPROVE**

## Findings

_None._

## Notes

- The R010 issue is addressed: the public `add_or_update_event` example now uses `RACE_B`, and the new schema-example guard checks examples against `intervals.EventCategoryValues()` while still allowing custom-category pass-through.
- The event categories resource is registered in the default registry, returns deterministic `text/markdown`, is golden-file locked, and has protocol coverage for `resources/list`/`resources/read`.
- The shared descriptor is copied on read and is reused by event tool schema descriptions without adding runtime enum validation, preserving the custom/upstream category behavior documented in the step plan.

## Verification

- `go test ./...` passes.
