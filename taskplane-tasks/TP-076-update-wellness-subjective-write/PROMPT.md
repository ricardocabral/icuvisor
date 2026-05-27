# TP-076 — `update_wellness` subjective write refused

## Mission

Resolves issue #7 — *TP-029: `update_wellness` subjective write refused during dogfood*.

During the v0.3 dogfood (TP-029, W-06 in [`docs/dogfood/v0.3-findings.md:24`](../../docs/dogfood/v0.3-findings.md)), an `update_wellness` call against the dedicated test athlete attempted to write a synthetic subjective bundle: `feel`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, `locked`. The upstream POST was refused; `get_wellness_data` re-read confirmed none of those fields landed on the row. The same dogfood's W-07 measurement write was BLOCKED preemptively because `spO2` was missing from the schema — that's a separate bug, tracked in [TP-072](../TP-072-update-wellness-missing-write-fields/) (closes #8).

Suspected defects (from triage, ordered by likelihood — confirm via live probe):

1. **Field-name mismatch** — one of the eight fields uses a JSON key that doesn't match the upstream contract. `sleepQuality` is plausibly the offender (camelCase vs snake_case variation across endpoints).
2. **Date format / scoping** — the wellness write may require an exact `start_date_local` date format, or it may require the athlete-ID scoping on the URL path rather than the body.
3. **Validation gate** — upstream may reject an entire payload if one field (e.g. `locked: true`) conflicts with another (e.g. attempting to set subjective fields on a row that's already device-locked).
4. **Less likely but possible:** a 1-based vs 0-based scale mismatch (intervals.icu may reject `feel: 0`).

This task requires **live API probing** against the `.env-dev` test athlete to isolate the rejection cause.

PRD anchors: §7.2.C `update_wellness` writable field set ([line 248](../../docs/prd/PRD-icuvisor.md)); `sleepQuality` 1–4 scale; `locked` semantics.
CLAUDE.md: hard rule 5; MCP-server conventions ("Schemas matter"; "Errors back to the LLM must be short, actionable, and free of internal stack traces").

Complexity: Blast radius 1, Pattern novelty 2 (live-probe loop), Security 1, Reversibility 2 (writes against real account on dedicated test athlete) = 6 → Review Level 2. Size: M.

## Dependencies

- **Soft dependency on [TP-072](../TP-072-update-wellness-missing-write-fields/) (closes #8)** — if TP-072 lands first, your live probe should start from a complete schema. Not a hard blocker; you can probe with the existing subset.
- Independent of TP-073, TP-074, TP-075, TP-077.
- Useful prior art: [TP-021](../TP-021-wellness-write/) shipped the original wellness write path — read its STATUS/PROMPT for the field-name decisions.

## Context to Read First

- [`docs/prd/PRD-icuvisor.md:237-254`](../../docs/prd/PRD-icuvisor.md) — wellness read/write contract; scale labels; `locked` semantics.
- [`docs/dogfood/v0.3-findings.md`](../../docs/dogfood/v0.3-findings.md) line 24 (W-06) and per-tool triage row for `update_wellness`.
- [`internal/tools/update_wellness.go`](../../internal/tools/update_wellness.go):
  - Request struct: lines 63–83.
  - Handler: lines 105–133.
  - Input schema: lines 325–349.
- [`internal/intervals/wellness.go`](../../internal/intervals/wellness.go):
  - `WriteWellnessParams`: lines 20–39.
  - `writeWellnessBody`: lines 155–178.
  - `Wellness` read struct: line 73+ — use this as the source of truth for JSON field names.
- [`internal/tools/update_wellness_test.go`](../../internal/tools/update_wellness_test.go) — existing pattern.
- Existing TP work: [`taskplane-tasks/TP-021-wellness-write/`](../TP-021-wellness-write/).

## File Scope

- `internal/intervals/wellness.go` — fix any field-name mismatch in `writeWellnessBody`.
- `internal/tools/update_wellness.go` — if the fix requires schema-level adjustment (e.g. an enum, or pre-validation against `locked`).
- `internal/tools/update_wellness_test.go` — add a subjective-bundle round-trip test that mirrors what W-06 attempted.
- `internal/intervals/testdata/wellness/` — capture the working request/response from the live probe.
- `CHANGELOG.md` — `[Unreleased]` under "Fixed".
- `STATUS.md` (this dir).

Out of scope:
- Adding the missing measurement fields — that's TP-072.
- Refactoring the `setSparse` helper.
- Changing the read-path tool.

## Steps

### Step 1: Live probe to isolate the rejection

- [ ] Source `.env-dev`.
- [ ] **Out-of-process probe** (curl or scratch Go program — do NOT commit): POST a wellness update directly to intervals.icu for the test athlete on a recent date. Bisect the field set to isolate which field(s) trigger the rejection:
  - Single-field probes: `{ "feel": 3 }`, then `{ "fatigue": 2 }`, etc.
  - Pair probes if singles all pass: `{ "feel": 3, "locked": true }`.
  - Watch the response body for any structured error from intervals.icu — it often names the offending field.
- [ ] Try also varying:
  - `start_date_local: "YYYY-MM-DD"` vs `"YYYY-MM-DDT00:00:00"`.
  - URL path vs body for athlete-ID scoping.
- [ ] Record the exact accepted minimal payload and the exact rejected one. Save sanitized request/response under `internal/intervals/testdata/wellness/subjective_write_request.json` + `subjective_write_response.json`.
- [ ] **Clean up:** reset any probe data on the test athlete (set `locked: false`, clear any unintended subjective values for the probe date). Verify with `get_wellness_data`.

### Step 2: Add a failing test

- [ ] Mirror existing wellness-write tests for the full subjective bundle (`feel`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, `locked`).
- [ ] Use the captured fixture as the expected outbound body.
- [ ] Confirm the test fails on `main`.

### Step 3: Fix the client / tool

- [ ] Apply the minimum diff that makes the probed-good payload come out of `writeWellnessBody`. Most likely a JSON key rename (e.g. `sleepQuality` → `sleep_quality` if upstream snake-cases it) or a date-format change.
- [ ] If the fix involves a tool-level validation gate (e.g. reject `locked: true` combined with conflicting fields), add it to the handler with a clear public error message.

### Step 4: Build + lint + race + live re-validation

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Live re-validation: call `update_wellness` via stdio MCP with the full subjective bundle; confirm via `get_wellness_data` that all eight fields land on the row.

### Step 5: Document amendment

- [ ] If the upstream contract surfaced a field-name or shape inconsistency, capture it in `docs/upstream-gaps/wellness-write-payload.md` so future agents don't re-do the probe.

### Step 6: Close the issue

- [ ] Update `CHANGELOG.md` under `[Unreleased] → Fixed`: `update_wellness subjective field bundle now writes successfully (was refused upstream due to <root cause>).`
- [ ] Update `STATUS.md`.
- [ ] Commit: `fix(wellness): repair update_wellness subjective write (TP-076, closes #7)`.
- [ ] Reference `Closes #7` in the PR body. After merge, verify auto-close; otherwise close issue #7 manually after merge.

## Acceptance Criteria

- A full subjective-bundle `update_wellness` call against the test athlete is accepted upstream, and `get_wellness_data` shows all eight fields on the target row.
- The `locked: true` flag survives a subsequent device-sync simulation (if you can trigger one; otherwise: assert it persists in the immediate re-read).
- A new unit test exercises the subjective-bundle round-trip using a captured fixture.
- `make build`, `make test`, `make test-race`, `make lint` pass.
- Probe-introduced data on the test athlete is reset.
- Issue #7 closed.

## Do NOT

- Do not commit probe scratch files.
- Do not leave probe-introduced wellness rows on the test athlete with `locked: true` (you'll block future device sync).
- Do not relax the strict decoder — its rejection of unknown fields is correct (and is why TP-072 needs to land for the measurement bundle).
- Do not paste raw athlete data, raw HRV values, or date-identifying probe data into the repo without sanitization.

## Documentation

- `CHANGELOG.md` `[Unreleased]` under "Fixed".
- `STATUS.md` in this dir.
- *Optional:* `docs/upstream-gaps/wellness-write-payload.md`.

## Git Commit Convention

Conventional Commits, prefixed with TP-076. Example:

```
fix(wellness): repair update_wellness subjective write

TP-076. Closes #7.
```

---

## Amendments

_Add amendments below this line only._

### Amendment 1 — 2026-05-27
**Issue:** Prior operator/supervisor steering told workers to let TP-076 fail and forbade further live wellness writes, DELETE attempts, endpoint probes, direct GETs, or new probe-row creation, leaving the Step 1 cleanup checkbox blocked.
**Resolution:** The operator explicitly superseded that prior steering for the next batch: complete the remaining live test-athlete cleanup work using the test athlete credentials from `.env-dev`. Keep the original safety constraints: use only the dedicated test athlete, do not print secrets or raw athlete IDs, do not use production credentials, make the minimum live calls required to restore/verify cleanup, and record sanitized evidence in `STATUS.md`.
