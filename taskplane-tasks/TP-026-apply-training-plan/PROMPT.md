# TP-026 — `apply_training_plan`

## Mission

Apply a training plan (a set of workouts mapped to dates) to the athlete's calendar. Bulk-create events from the plan's workout-library entries; surface conflicts; honour partial application when some dates already have events.

Roadmap items (ROADMAP.md v0.3):

- `apply_training_plan`.

PRD anchors: §7.2.C training-plan catalog, §7.4 #18.

Complexity: Blast radius 3 (bulk calendar mutation), Pattern novelty 2, Security 2, Reversibility 2 = 9 → Review Level 2. Size: M.

## Dependencies

- **TP-018** — safety gate
- **TP-019** — workout_doc serializer (events created from library workouts carry structured steps)
- **TP-020** — event write cluster (this task is its bulk variant)
- **TP-023** — workout-library CRUD (plan references library workouts)
- **TP-012** — events read (conflict detection)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C training-plan, §7.4 #18
- `ROADMAP.md` v0.3
- `internal/tools/get_training_plan*.go`, `internal/tools/add_or_update_event*.go`

## File Scope

Expected files:

- `internal/tools/apply_training_plan.go` + `_test.go`
- `CHANGELOG.md`
- `README.md` catalog
- `taskplane-tasks/TP-026-apply-training-plan/STATUS.md`

## Steps

### Step 1: Schema

- [ ] Inputs: `plan_id`, `start_date` (anchor; plan dates are relative), optional `dry_run: bool` (default `true`), `conflict_policy` (`skip_existing` | `replace_existing` — `replace_existing` requires `CanDelete()`)
- [ ] Plan content fetched server-side, never passed in by the LLM
- [ ] Default `dry_run: true` is the safety net even in `safe` mode — the LLM must explicitly set `dry_run: false` to mutate

### Step 2: Dry-run path

- [ ] Returns the proposed event list with conflict markers; no upstream writes
- [ ] Per-day shape: `{date, workout_id, conflicts: [{event_id, reason}]}`

### Step 3: Apply path

- [ ] `dry_run: false` + `conflict_policy: skip_existing` → create events only on conflict-free days; skipped days listed in `_meta.skipped`
- [ ] `dry_run: false` + `conflict_policy: replace_existing` → registered only when `CanDelete()` (this argument is rejected at the validation layer otherwise); deletes the conflicting event then creates the new one
- [ ] Use TP-020's `add_or_update_event` internals; do not duplicate event-build logic

### Step 4: Response shape

- [ ] `_meta.created_count`, `_meta.skipped`, `_meta.replaced` (when applicable), `_meta.delete_mode`

### Step 5: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the test athlete: dry-run, skip-existing apply, replace-existing apply (in `full`)

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint shapes. Do not depend on it.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- `dry_run: true` is the default and produces no upstream writes.
- `replace_existing` is rejected at validation in `safe` mode.
- Bulk apply uses the TP-020 event-create path; no parallel event-build code.
- `_meta` reports created / skipped / replaced counts.

## Do NOT

- Do not accept a `confirm` argument.
- Do not invert the default to `dry_run: false`.
- Do not silently delete on conflict when policy is `skip_existing`.

## Documentation

Must update:

- `STATUS.md`
- `README.md` catalog
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-026`, for example: `TP-026 add apply_training_plan dry-run`.

---

## Amendments

_Add amendments below this line only._
