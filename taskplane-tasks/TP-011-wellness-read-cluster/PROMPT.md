# TP-011 — Wellness reads with sleep dual-scale, provenance, staleness, and `_native` sub-fields

## Mission

Ship `get_wellness_data` with the provenance and freshness signals required so an LLM never silently compares incompatible scales or reasons over stale bridged values. This is the highest-precision tool in v0.2; the LLM is poor at recovering from silent scale collisions.

Roadmap items (ROADMAP.md v0.2):

- Distinct sleep fields surfaced separately: manual `sleepQuality` (1–4) and device-imported `sleepScore` (0–100), each with its own in-response scale label.
- Wellness `_meta.provenance` (per bridged field: `source`, `native_scale`, `fetched_at`) + `_meta.stale: true` when the upstream bridge has not refreshed within 24h of the wellness date.
- Raw native sub-fields exposed under `_native.<source>.<field>` (Polar `ans_charge`, `nightly_recharge_status`; Garmin body-battery min/max; Oura raw `sleep_score`).
- `_meta.missing_fields` callout when null-stripping removes a key.
- In-response scale labels on every subjective field (`feel`, `sleepQuality`, `fatigue`, `mood`, etc.).

PRD anchors: §7.2.C Wellness, §7.2.D (response shaping, provenance + freshness block).

Complexity: Blast radius 2, Pattern novelty 3 (provenance + native sub-fields), Security 1, Reversibility 1 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-007** — response shaping primitives (null-strip, scale labels, `_meta` plumbing)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C Wellness (full block, including provenance + freshness), §7.2.D
- `ROADMAP.md` v0.2
- Upstream public discussion references
- `internal/response/` from TP-007
- Public intervals.icu API docs for wellness; black-box probe for bridged fields and raw sub-fields

## File Scope

Expected files:

- `internal/intervals/wellness.go` — typed client method
- `internal/tools/get_wellness_data.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/intervals/testdata/wellness/` — fixtures covering: Polar bridge fresh, Polar bridge stale, Garmin body-battery, Oura raw `sleep_score`, manual-only row, custom-fields row
- `CHANGELOG.md`
- `taskplane-tasks/TP-011-wellness-read-cluster/STATUS.md`

## Steps

### Step 1: Map the wellness payload

- [ ] Catalog every field exposed by the intervals.icu wellness endpoint, distinguishing: athlete-entered scales, device-imported normalized fields, raw native sub-fields, and custom fields
- [ ] Identify each bridged field's provider source (Polar / Garmin / Oura / Whoop / Apple Health / manual / unknown) and native scale
- [ ] Record findings and any uncertainty in `STATUS.md`

### Step 2: Implement typed decoding

- [ ] Decode `sleepQuality` (1–4) and `sleepScore` (0–100) as **distinct fields** — no aliasing, no collapse
- [ ] Decode `sleepSecs` as its own field; do not derive from `sleepScore`
- [ ] Decode raw native sub-fields under a sidecar struct and surface under `_native.<source>.<field>`
- [ ] Preserve custom-field rows (intervals.icu wellness custom fields) — they participate in null-stripping the same as standard fields

### Step 3: Provenance and staleness `_meta` assembly

- [ ] For every bridged field, emit `_meta.provenance.<field> = { source, native_scale, fetched_at }`
- [ ] Where provenance cannot be determined, emit `source: "unknown"` rather than omitting the marker
- [ ] If `now - fetched_at > 24h` relative to the wellness `date`, emit `_meta.stale: true` with a one-line `_meta.stale_reason` (e.g. `"polar bridge refresh requires user to open intervals.icu"`)

### Step 4: In-response scale labels

- [ ] Register every subjective scale in the TP-007 scale-label registry: `feel` 1–5, `sleepQuality` 1–4, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `injury`
- [ ] Verify `_meta.scales` appears in the response for every registered field present in the row

### Step 5: Null-stripping integration

- [ ] Confirm wellness rows pass through the TP-007 null-strip pipeline
- [ ] `_meta.missing_fields: [...]` per row when stripping removed at least one key
- [ ] `include_full: true` opt-out is honoured

### Step 6: Tests

- [ ] Table-driven tests over the `testdata/wellness/` fixtures
- [ ] Assert: distinct sleep fields; provenance per bridged field; `_meta.stale` boundary at exactly 24h; `_native` round-trip for each provider in fixtures; null-strip + `_meta.missing_fields`; scale labels for every subjective field
- [ ] `make test`, `make build`, `make lint` pass

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for wellness endpoint shape. Do not depend on it.
- GPL/copyleft wellness handling is off limits; reimplement from public API behavior.

## Acceptance Criteria

- `get_wellness_data` registered and returning the contract above.
- Distinct `sleepQuality` and `sleepScore` fields with separate scale labels; never aliased.
- `_meta.provenance` present per bridged field (incl. `source: "unknown"` when undetermined).
- `_meta.stale: true` + `stale_reason` when the bridge is >24h behind the wellness date.
- `_native.<source>.<field>` surfaces raw provider sub-fields for at least Polar (`ans_charge`, `nightly_recharge_status`), Garmin (body-battery min/max), and Oura (raw `sleep_score`) where the fixtures expose them.
- Tests cover Step 6 cases.

## Do NOT

- Do not collapse `sleepQuality` and `sleepScore` under a single key.
- Do not silently drop the provenance marker when the source is unknown.
- Do not invent provenance — if a field's source cannot be determined from the upstream payload, emit `source: "unknown"`.
- Do not implement `update_wellness` here (that is v0.3 / TP-write tasks).

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`
- README catalog (add `get_wellness_data`)

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-011`, for example: `TP-011 add wellness provenance meta`.

---

## Amendments

_Add amendments below this line only._
