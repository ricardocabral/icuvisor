# TP-025 — Destructive deletes cluster (event, activity, custom-item, sport-settings, gear)

## Mission

Land the destructive write tools, all gated by `ICUVISOR_DELETE_MODE`. Each is registered only when `CanDelete()` is true — no per-call `confirm: true` anywhere. `delete_events_by_date_range` is the most dangerous and gets the strictest validation.

Roadmap items (ROADMAP.md v0.3):

- Event delete (`delete_event`, `delete_events_by_date_range`), activity delete, custom-item delete, sport-settings delete, gear delete — all gated by `ICUVISOR_DELETE_MODE`.

PRD anchors: §7.2.C delete catalog, §7.2.D safety model, §7.4.

Complexity: Blast radius 3 (mass deletes), Pattern novelty 1, Security 3, Reversibility 3 = 10 → Review Level 2. Size: M.

## Dependencies

- **TP-018** — safety gate (`CanDelete`)
- **TP-020** — events write (for round-trip / undo guidance)
- **TP-023** — workout-library CRUD (delete sibling pattern)
- **TP-024** — custom-items writes

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C delete catalog, §7.2.D, §7.4
- `ROADMAP.md` v0.3
- `internal/safety/`

## File Scope

Expected files:

- `internal/tools/delete_event.go` + `_test.go`
- `internal/tools/delete_events_by_date_range.go` + `_test.go`
- `internal/tools/delete_activity.go` + `_test.go`
- `internal/tools/delete_custom_item.go` + `_test.go`
- `internal/tools/delete_sport_settings.go` + `_test.go`
- `internal/tools/delete_gear.go` + `_test.go`
- `CHANGELOG.md`
- `README.md` catalog
- `taskplane-tasks/TP-025-destructive-deletes-cluster/STATUS.md`

## Steps

### Step 1: Per-ID deletes

- [ ] `delete_event`, `delete_activity`, `delete_custom_item`, `delete_sport_settings`, `delete_gear`: each takes an opaque ID, returns the deleted ID and a short before-shape echo in `_meta.deleted` so the LLM can confirm
- [ ] Registered only in `full` mode (TP-018 `CanDelete`)
- [ ] No `confirm` argument

### Step 2: `delete_events_by_date_range`

- [ ] Inputs: `start_date`, `end_date` (athlete-TZ; both required; same-day allowed), optional `category` filter
- [ ] Hard validation: range size capped (document the cap in the schema description and in `STATUS.md`); reject open-ended ranges
- [ ] Response includes `_meta.deleted_count` and the ID list
- [ ] Registered only in `full`

### Step 3: Tests

- [ ] Per tool: success in `full`, absent from catalog in `safe` and `none`
- [ ] `delete_events_by_date_range`: range-cap rejection, athlete-TZ correctness on boundary dates
- [ ] Idempotency where upstream supports it (re-delete returns 404 mapped to a typed error, not a 500)

### Step 4: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the test athlete in `full` mode; never in production

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint shapes. Do not depend on it.
- GPL/copyleft implementation code is off limits — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- All six delete tools registered only in `full` mode.
- `delete_events_by_date_range` rejects unbounded or oversize ranges at the validation layer.
- No `confirm` argument anywhere.
- Per-deletion `_meta.deleted` echo lets the LLM report what was destroyed.

## Do NOT

- Do not add a `confirm: true` argument anywhere.
- Do not silently widen a date range; reject and let the LLM split the call.
- Do not run any of these manually against a production athlete account.
- Do not log full deleted-row payloads; one-line summaries only.

## Documentation

Must update:

- `STATUS.md`
- `README.md` catalog
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-025`, for example: `TP-025 add delete_event tool`.

---

## Amendments

_Add amendments below this line only._
