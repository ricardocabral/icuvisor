# TP-013 — Workout-library and custom-items reads

## Mission

Ship the read side of the workout-library (templates) and custom-items surfaces. These are smaller, lower-traffic tools, but they round out the catalog and unlock TP-009/TP-010 prompts that want to "list workouts in my Threshold folder" or "show me my custom CTL chart definition."

Roadmap items (ROADMAP.md v0.2):

- `get_workout_library`, `get_workouts_in_folder`
- `get_custom_items`, `get_custom_item_by_id`

PRD anchors: §7.2.C Workout library (read side only), §7.2.C Custom items, §7.2.G MCP Resources (the per-`item_type` `content` schema lands as a Resource in v0.4; inline docs are acceptable at v0.2).

Complexity: Blast radius 1, Pattern novelty 2, Security 1, Reversibility 1 = 5 → Review Level 1. Size: M.

## Dependencies

- **TP-007** — response shaping primitives

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C Workout library + Custom items, §7.2.G Resources
- `ROADMAP.md` v0.2
- Public intervals.icu API docs for workout-library and custom-items endpoints

## File Scope

Expected files:

- `internal/intervals/workout_library.go`
- `internal/intervals/custom_items.go`
- `internal/tools/get_workout_library.go`
- `internal/tools/get_workouts_in_folder.go`
- `internal/tools/get_custom_items.go`
- `internal/tools/get_custom_item_by_id.go`
- `_test.go` for each, with fixtures
- `CHANGELOG.md`
- `taskplane-tasks/TP-013-workout-library-custom-items-reads/STATUS.md`

## Steps

### Step 1: Workout-library reads

- [ ] `get_workout_library`: list folders (and optionally top-level workouts) with terse rows
- [ ] `get_workouts_in_folder`: list workouts inside a given folder ID; terse rows include name, sport, structured-step summary; `include_full` exposes the raw `workout_doc`
- [ ] Round-trip the `workout_doc` shape verbatim on reads; serialization to the upload DSL is a write-side concern (v0.3 / TP-write tasks)

### Step 2: Custom-items reads

- [ ] `get_custom_items`: list custom items (charts / fields / zones), terse rows include `id`, `name`, `item_type`
- [ ] `get_custom_item_by_id`: return the full item including the per-`item_type` `content` payload
- [ ] Long-form `content` schema description for each `item_type` lives inline in the tool description for v0.2; note in `STATUS.md` that this moves to `icuvisor://custom-item-schemas` in v0.4 (TP for that task will land later)

### Step 3: Tests

- [ ] Table-driven tests using `httptest.Server` + fixtures
- [ ] Cover: empty library; nested folders; multiple `item_type` variants of custom items
- [ ] `make test`, `make build`, `make lint` pass

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted. Do not depend on it.
- GPL/copyleft implementation code is off limits.

## Acceptance Criteria

- Four read tools registered.
- `get_workouts_in_folder` returns terse rows by default and raw `workout_doc` only under `include_full`.
- Custom-item `content` shape is preserved verbatim on reads.
- Tests cover Step 3.

## Do NOT

- Do not implement workout-library writes (`create_workout`, `update_workout`, `delete_workout`) — v0.3.
- Do not implement custom-item writes — v0.3.
- Do not serialize `workout_doc` to the upload DSL here (write-only concern, v0.3 / TP-write).

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`
- README catalog (add four tools)

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-013`, for example: `TP-013 add get_workout_library`.

---

## Amendments

_Add amendments below this line only._
