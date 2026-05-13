# TP-022 — `update_sport_settings` (FTP, threshold HR/pace, zones)

## Mission

Land the sport-settings write path: FTP, threshold HR, threshold pace, and zone definitions per sport. Zone-definition *overwrites* (which silently destroy prior zone history if mishandled — forum #35) are gated by `ICUVISOR_DELETE_MODE`. Threshold updates without zone-definition changes are ungated writes.

Roadmap items (ROADMAP.md v0.3):

- `update_sport_settings` (FTP, threshold HR/pace, zones; zone-definition overwrites gated by `ICUVISOR_DELETE_MODE` — forum #35).

PRD anchors: §7.2.C sport-settings catalog, §7.2.D safety model, §7.4 zone-destructive note.

Complexity: Blast radius 2, Pattern novelty 2, Security 2 (zone-history loss is hard to recover), Reversibility 2 (threshold writes auto-recompute history) = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-018** — safety gate (`CanWrite`, `CanDelete`)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C sport-settings, §7.4
- `ROADMAP.md` v0.3
- intervals.icu forum thread #35 on zone-definition overwrites — record canonical link in `STATUS.md`

## File Scope

Expected files:

- `internal/tools/update_sport_settings.go` + `_test.go`
- `internal/intervals/` — typed sport-settings write if not present
- `CHANGELOG.md`
- `README.md` catalog
- `taskplane-tasks/TP-022-sport-settings-write/STATUS.md`

## Steps

### Step 1: Schema design

- [ ] Inputs: `sport` (enum), `effective_date` (athlete-TZ), `ftp`, `threshold_hr`, `threshold_pace` (unit-aware), `zones[]` (optional structured zone definitions)
- [ ] `zones[]` triggers the gated path; everything else is an ungated write
- [ ] Test: omitted `zones` → no zone write

### Step 2: Threshold-only path (ungated)

- [ ] Update FTP / threshold HR / threshold pace, sport-scoped, dated
- [ ] Response echoes the new settings + a `_meta.recompute_pending: true` hint if the upstream recompute is async
- [ ] Tests: per-field updates, unit conversion for pace

### Step 3: Zone-definition path (`CanDelete` gated)

- [ ] If `zones[]` is present, require `CanDelete()` — otherwise reject with a typed error explaining the gate (the tool is still registered in `safe`, but the destructive *argument* is rejected at the schema/validation layer)
- [ ] Document in the schema description that supplying `zones` overwrites prior zone definitions and is therefore gated
- [ ] Tests: `safe` mode rejects `zones`, `full` mode applies them

### Step 4: Response shape

- [ ] Echo new settings; include `_meta.delete_mode` and any unit metadata

### Step 5: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the test athlete: FTP bump in `safe` mode (succeeds), zone overwrite in `safe` mode (rejected), zone overwrite in `full` mode (succeeds)

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint shapes. Do not depend on it.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- Threshold-only updates succeed in `safe` mode.
- Zone-definition overwrites are rejected at the validation layer in `safe` mode and succeed in `full` mode.
- The schema description explicitly warns about zone-overwrite destructiveness.
- No `confirm` argument anywhere.

## Do NOT

- Do not allow a single call to silently overwrite zones when the caller only intended to change FTP — treat the presence of `zones[]` as the destructive signal.
- Do not assume upstream recompute is synchronous; surface a `_meta.recompute_pending` hint.

## Documentation

Must update:

- `STATUS.md`
- `README.md` catalog
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-022`, for example: `TP-022 add update_sport_settings threshold path`.

---

## Amendments

_Add amendments below this line only._
