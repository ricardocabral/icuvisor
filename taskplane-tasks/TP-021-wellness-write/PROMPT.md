# TP-021 — `update_wellness` (full writable field set)

## Mission

Ship the wellness write path. The full writable field set per PRD §7.2.C: subjective scales (`feel`, `fatigue`, `mood`, `sleepQuality`, `motivation`, `soreness`, `stress`), measurements (`weight`, body fat %, blood pressure systolic/diastolic, blood glucose, lactate, resting HR, HRV), `injury` (free-text), and the `locked` flag. Bridged device-imported fields (`sleepScore`, body battery, etc.) remain read-only — those come from the bridge, not the LLM.

Roadmap items (ROADMAP.md v0.3):

- `update_wellness` (full writable field set incl. `injury`, blood pressure, blood glucose, lactate, body fat, `locked`).

PRD anchors: §7.2.C wellness catalog, §7.2.D scale labels + dual sleep fields, §7.4 #11 unit normalization.

Complexity: Blast radius 2, Pattern novelty 2, Security 2 (clinical-adjacent fields), Reversibility 2 = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-018** — safety gate (`CanWrite`)
- **TP-011** — wellness reads (scale registry, dual-sleep distinction, `_native` boundary)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C wellness, §7.2.D
- `ROADMAP.md` v0.3
- `internal/tools/get_wellness_data*.go`
- `internal/response/` scale-label registry

## File Scope

Expected files:

- `internal/tools/update_wellness.go` + `_test.go`
- `internal/intervals/` — typed wellness PUT/PATCH if not present
- `CHANGELOG.md`
- `README.md` catalog
- `taskplane-tasks/TP-021-wellness-write/STATUS.md`

## Steps

### Step 1: Define the writable field set

- [ ] Subjective scales: `feel` (1–5), `fatigue` (1–5), `mood` (1–5), `sleepQuality` (1–4 — manual scale, distinct from device-imported `sleepScore` 0–100), `motivation`, `soreness`, `stress`
- [ ] Measurements: `weight` (athlete-preferred unit), body fat %, systolic / diastolic, blood glucose, lactate, resting HR, HRV
- [ ] Free-text: `injury`
- [ ] Flag: `locked`
- [ ] Read-only here (reject in schema): `sleepScore` and all bridged `_native.*` fields — they originate from the bridge

### Step 2: Schema + scale enforcement

- [ ] Per-field JSON Schema includes range / units / scale label, matching the registry from TP-011
- [ ] Inputs out of range fail at the schema layer, not at the upstream call
- [ ] Unit conversion: `weight` accepts the athlete's preferred unit and converts to upstream's canonical unit at the boundary

### Step 3: Partial-update semantics

- [ ] Inputs are sparse — omitted fields are not touched upstream
- [ ] Test: setting only `feel` does not zero `weight`
- [ ] Test: setting `locked: true` then a follow-up update with another field surfaces the upstream lock behavior in `_meta`

### Step 4: Response shape

- [ ] Echo the updated wellness row through the read shape (TP-011 helpers): scale labels, provenance for any bridged fields the row still carries, `_meta.delete_mode`
- [ ] Strip nulls; surface `_meta.fields_present` / `missing_fields`

### Step 5: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the test athlete: set subjective, set measurement, set `injury` free-text, set `locked: true`

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint shapes. Do not depend on it.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- `update_wellness` registered in `safe` and `full` modes; absent in `none`.
- Schema rejects out-of-scale subjective values and any attempt to write `sleepScore` or `_native.*` keys.
- Partial updates preserve untouched fields.
- Response echoes the updated row through the TP-011 read shape with scales / provenance / `_meta.delete_mode`.

## Do NOT

- Do not accept `sleepScore` (device-imported) as a write target — that is `sleepQuality`'s job.
- Do not log subjective values, injury text, or measurements in a way that's hard to scrub (`slog` with redactable fields only).
- Do not silently coerce a 1–5 value into a 1–4 `sleepQuality` — reject with a typed error.
- Do not commit real wellness data in fixtures.

## Documentation

Must update:

- `STATUS.md`
- `README.md` catalog
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-021`, for example: `TP-021 add update_wellness tool`.

---

## Amendments

_Add amendments below this line only._
