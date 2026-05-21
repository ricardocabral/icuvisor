# Plan Review — TP-079 Step 1

**Verdict:** Request changes before implementation

The Step 1 plan is directionally correct, but it is too thin for the risky part of this task: establishing the exact upstream contract that later activity shaping will depend on. I would tighten the plan before coding so Step 2/3 do not end up guessing field names.

## Findings

1. **Activity gear-field evidence needs fixtures, not just gear-list fixtures.**  
   The checklist asks for fixtures for gear-list responses, but the mission also depends on activity list/detail payloads exposing gear IDs. `internal/intervals.Activity` currently has no gear field, and `get_activities` uses an explicit `terseActivityFields` allowlist. Step 1 should require representative fixture(s) for activity payloads containing the discovered gear field(s), and should record whether those fields appear in list responses, details responses, and/or only when requested via `fields=`.

2. **Reuse/extend the existing gear model instead of creating a divergent one.**  
   There is already an `intervals.Gear` type and `GetGear`/`DeleteGear` path in `internal/intervals/delete.go`. The plan says “Add typed intervals client structs”, but should explicitly say to extend/move the existing `Gear` model as needed, not introduce a second gear DTO with different ID/name semantics.

3. **Endpoint probing should cover path and response shape edge cases.**  
   Existing code implies `GET /athlete/{id}/gear/{gearId}` for one item. Step 1 should explicitly verify/model the list endpoint path, the top-level JSON shape (`[]Gear` vs wrapped object), ID type (`number` vs `string`), empty-list behavior, retired/disabled gear fields, and whether names may be absent. These details affect both the typed client and terse response shaping.

4. **Upstream-gap documentation criteria should be explicit.**  
   The plan says to document a gap “if gear IDs are not exposed consistently,” but it should define the cases that trigger that doc: for example, gear IDs available on details but not list rows, unavailable through `fields=`, multiple possible field names, IDs present without names in the list endpoint, or Strava/imported activities omitting gear. If no gap is found, record that in `STATUS.md` discoveries.

## Suggested plan additions

- Add `ListGear(ctx)` (or equivalent) tests using `httptest`/fixtures only; no networked tests.
- Add fixture(s) for activity list/detail gear fields discovered during probing, even if final activity shaping lands in Step 3.
- Record the exact upstream field names and endpoint path in `STATUS.md` so later steps can implement from evidence.
- If any live/manual black-box probing is used, ensure credentials and raw athlete identifiers are not committed or logged.

