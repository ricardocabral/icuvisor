# TP-072 — `update_wellness` missing write fields (`spO2`, `vo2max`, `abdomen`, `respiration`, `menstrualPhase`)

## Mission

Resolves [GitHub issue #8](https://github.com/ricardocabral/icuvisor/issues/8) — *TP-029: update_wellness schema lacks spO2 write field*.

PRD §7.2.C (`docs/prd/PRD-icuvisor.md:248`) explicitly lists the writable wellness field set:

> subjective scales (`feel`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, `injury`), body metrics (`weight`, `bodyFat`, `abdomen`), cardiovascular (`restingHR`, `hrv`, `systolic`, `diastolic`), blood/lab (`bloodGlucose`, `lactate`, `spO2`, `vo2max`), respiration, menstrual phase, and the `locked` flag

The current `update_wellness` tool is missing **five** of those: `spO2`, `vo2max`, `abdomen`, `respiration`, `menstrualPhase`. The TP-029 v0.3 dogfood (W-07) was blocked because `spO2` is not in the registered tool schema, so the LLM correctly refused a partial measurement write rather than silently dropping the field.

This task adds the missing fields end-to-end (request struct, tool schema, client payload, tests) and closes #8.

PRD anchors: §7.2.C wellness write field set (line 248).
CLAUDE.md hard rules: rule 5 (tools terse by default, but schemas must be complete); MCP-server conventions ("Schemas matter").

Complexity: Blast radius 1 (single tool + client), Pattern novelty 1 (mirror existing field pattern), Security 1, Reversibility 1 = 4 → Review Level 1. Size: S.

## Dependencies

- None. Independent of the other v0.2/v0.3 dogfood follow-ups (TP-073 through TP-077).
- Issue #7 (TP-076 — `update_wellness` subjective write refused) is a separate live-API bug. If TP-076 finds that the subjective failure was a field that this task adds (e.g. spO2 was in the same bundle), coordinate the close on #7 there.

## Context to Read First

- [`docs/prd/PRD-icuvisor.md:237-254`](../../docs/prd/PRD-icuvisor.md) — wellness read + write field contract.
- [`docs/dogfood/v0.3-findings.md`](../../docs/dogfood/v0.3-findings.md) lines 25–26 (W-07) and lines 117, 140, 152 — the dogfood evidence that this task closes.
- [`internal/tools/update_wellness.go`](../../internal/tools/update_wellness.go) lines 63–83 (request struct), 105–133 (handler), 325–349 (input schema).
- [`internal/intervals/wellness.go`](../../internal/intervals/wellness.go) lines 20–39 (`WriteWellnessParams`), 73 (read-struct `Wellness.SpO2` — proves field exists on the read side), 139–178 (`UpdateWellness` + `writeWellnessBody`).
- [`internal/tools/update_wellness_test.go`](../../internal/tools/update_wellness_test.go) — existing test pattern for new field assertions.

## File Scope

- `internal/tools/update_wellness.go` — extend request struct, input schema, validation, `fields_updated` echo.
- `internal/intervals/wellness.go` — add fields to `WriteWellnessParams`; extend `writeWellnessBody` with `setSparse` calls.
- `internal/tools/update_wellness_test.go` — assert each new field round-trips into the request body and into the `fields_updated` response.
- `CHANGELOG.md` — `[Unreleased]` under "Fixed" or "Added".
- `STATUS.md` (this dir).

Out of scope:
- Refactoring the rest of `update_wellness` (handler shape, error mapping).
- Adding new validation modes beyond simple non-negative range checks for numeric fields.
- Touching the read-path `get_wellness_data` (it already surfaces these fields).

## Steps

### Step 1: Add fields to client write struct + payload

- [ ] Read [`internal/intervals/wellness.go`](../../internal/intervals/wellness.go) to confirm the JSON field names used by the read struct (`Wellness`): `spO2`, `vo2max`, `abdomen`, `respiration`, `menstrualPhase` (or upstream casing — verify by grepping the read struct tags).
- [ ] Add to `WriteWellnessParams`:
  - `SpO2 *float64`
  - `VO2Max *float64`
  - `Abdomen *float64`
  - `Respiration *float64`
  - `MenstrualPhase *string`
- [ ] Extend `writeWellnessBody` with matching `setSparse(body, "<upstream_key>", params.X)` lines using the exact JSON key the read struct uses.

### Step 2: Expose fields in the tool

- [ ] Extend `updateWellnessRequest` in `internal/tools/update_wellness.go` with the same five fields (JSON tags matching the input-schema names — keep camelCase: `spO2`, `vo2max`, `abdomen`, `respiration`, `menstrualPhase`).
- [ ] Extend `updateWellnessInputSchema` with five new properties. For numeric ones, document the unit in the description (`"spO2: blood oxygen saturation percentage 0-100"`, `"vo2max: ml/kg/min"`, `"abdomen: cm"`, `"respiration: breaths per minute"`). For `menstrualPhase`: document the accepted enum if known, otherwise free-text string.
- [ ] Add validation in the handler:
  - `spO2` ∈ [0, 100]
  - `vo2max` ≥ 0
  - `abdomen` ≥ 0
  - `respiration` ≥ 0
  - `menstrualPhase` non-empty if provided (no enum gate until upstream contract is verified).
- [ ] Wire the new fields into the request → `WriteWellnessParams` mapping in the handler.
- [ ] Include them in `updateWellnessFieldsUpdated` (or whatever helper builds the `fields_updated` echo).

### Step 3: Tests

- [ ] Extend `update_wellness_test.go` with a table-driven case that writes one of each new field and asserts:
  - The outbound HTTP body (use the existing fixture/test-server harness) contains the field under the correct JSON key.
  - The response `fields_updated` includes the field name.
  - Validation rejects out-of-range values (spO2 = 200, negative abdomen, empty menstrualPhase).
- [ ] Add a single combined-fields test that writes all five at once.

### Step 4: Build, lint, manual smoke

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Manually smoke against `.env-dev` test athlete (optional but encouraged): a single `update_wellness` call with `spO2: 97` and a date in the recent past; re-read with `get_wellness_data` and confirm the field is present. Delete or `locked: false` cleanup afterward if you locked anything.

### Step 5: Close the GitHub issue

- [ ] Update `CHANGELOG.md` under `[Unreleased] → Fixed`: `update_wellness now accepts spO2, vo2max, abdomen, respiration, and menstrualPhase per PRD §7.2.C.`
- [ ] Update `STATUS.md`.
- [ ] Commit (Conventional Commits, lowercase, imperative): `fix(wellness): expose spO2/vo2max/abdomen/respiration/menstrualPhase on update_wellness (TP-072, closes #8)`.
- [ ] Push the branch and either open a PR or fold into the combined v0.2/v0.3 dogfood follow-up PR if that pattern is being used. Reference `Closes #8` in the PR body.
- [ ] After merge, verify the issue auto-closed; if it didn't, close manually with `gh issue close 8 --comment "Fixed in <commit-sha> / <PR>"`.

## Acceptance Criteria

- All five fields appear in: request struct, tool input schema, client write params, `writeWellnessBody`, and the response `fields_updated`.
- Out-of-range values are rejected with a clear public error.
- Unit tests cover each field individually and all-together.
- `make build`, `make test`, `make test-race`, `make lint` all pass.
- GitHub issue #8 is closed with a reference to the merging commit/PR.

## Do NOT

- Do not introduce a `confirm: true` argument (forbidden by CLAUDE.md hard rule 5 + MCP conventions). Destructive ops are env-gated, not LLM-gated.
- Do not silently drop unknown fields. The current strict decoder is correct; preserve it.
- Do not refactor `writeWellnessBody`'s `setSparse` helper unless adding a new field type genuinely requires it.
- Do not change the `locked` field semantics — it's already correct.

## Documentation

- `CHANGELOG.md` `[Unreleased]` under "Fixed" (or "Added" if you consider new fields additive — both are defensible; Fixed matches the issue framing).
- `STATUS.md` in this dir.

## Git Commit Convention

Conventional Commits, prefixed with TP-072 in the body. Example title:

```
fix(wellness): expose spO2/vo2max/abdomen/respiration/menstrualPhase on update_wellness

TP-072. Closes #8.
```

---

## Amendments

_Add amendments below this line only._
