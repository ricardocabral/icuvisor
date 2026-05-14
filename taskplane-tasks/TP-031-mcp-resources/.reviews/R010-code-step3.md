# Code Review R010 — Step 3: `icuvisor://event-categories`

**Verdict: REVISE**

## Findings

### P2 — Public event write example still advertises `RACE`, which is not in the documented enum

`internal/tools/add_or_update_event.go:210-218` still exposes an `add_or_update_event` schema example with:

```go
"category": "RACE",
```

This step introduces the shared documented category descriptor and updates the category schema descriptions to point at `icuvisor://event-categories`, whose enum contains `RACE_A`, `RACE_B`, and `RACE_C` but not `RACE`. The example is part of the model-facing input schema, so it now contradicts the resource and the descriptor that the same schema points to. A model following the example can send `RACE` for a race event even though the documented upstream enum says the appropriate values are priority-specific race categories.

Please update the public schema example to use a documented value (for example `RACE_B`, matching the “B race” description), or explicitly make it a custom-category example if that is intentional. I would also add a small guard that schema examples using standard categories are drawn from `intervals.EventCategoryValues()` so this does not drift again while preserving the runtime pass-through behavior and existing custom-category tests.

## Notes

- The resource registration, handler contract, golden Markdown, descriptor copy protection, and protocol list/read coverage otherwise look consistent with the Step 3 plan.
- I do not consider the runtime tests that use `RACE` for pass-through behavior a problem; the issue is specifically the public schema example presented as normal guidance after this step documents the upstream enum.

## Verification

- `go test ./...` passes.
