# Plan Review — TP-079 Step 1

**Verdict:** Approve

The revised Step 1 plan addresses the blocking issues from R001. It now treats upstream gear discovery as evidence-gathering for both sides of the feature: the gear list endpoint/model and the activity list/detail fields that later inline name resolution depends on.

## What is now covered

- The plan explicitly requires activity list/detail gear-field evidence, not only gear-list fixtures.
- It calls out reuse/extension of the existing `intervals.Gear` model instead of creating a divergent DTO.
- It includes endpoint and response-shape edge cases that matter for the typed client: list path, top-level JSON shape, ID type, empty lists, retired fields, and absent names.
- It requires recording exact endpoint/field evidence in `STATUS.md`, or documenting an upstream gap under `docs/upstream-gaps/` if the upstream contract is inconsistent.

## Non-blocking implementation notes

- When adding the list client/model in Step 1, include a small `httptest` or fixture-backed test so the discovered path and top-level shape are executable evidence, not just notes.
- Make the `STATUS.md` discovery precise enough for Step 2/3 to implement without re-probing: endpoint path, query parameters if any, JSON top-level shape, gear ID field name/type, activity field name(s), and whether fields appear in list responses, detail responses, or only with `fields=`.
- Keep any live/manual probing artifacts redacted. Do not commit raw athlete identifiers, API keys, or full personal activity payloads.

With those notes, the plan is sufficiently scoped and safe to implement.
