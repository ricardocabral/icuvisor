# R009 Plan Review — Step 5: Testing & Verification

**Verdict:** Request changes

## Findings

### 1. Step 5 is being planned while Step 4 still has unresolved request-changes findings

- **Severity:** High
- **Files:** `taskplane-tasks/TP-095-fitness-projection/.reviews/R008-code-step4.md`, `taskplane-tasks/TP-095-fitness-projection/STATUS.md`

R008 is a **Request changes** review with unresolved blockers: the generated website catalog is stale, the full-series response can still omit explicit zero `training_load` values, schema/decoder contract gaps remain, and STATUS history is inaccurate. The Step 5 plan currently only marks the step in progress and lists generic verification gates, but does not acknowledge or include remediation/verification for these known failures.

Please do not treat Step 5 as a pure gate-running step until the R008 findings are either fixed or explicitly carried as blockers with a concrete verification plan. At minimum, Step 5 should include targeted checks for each R008 item before `make test/build/lint` is considered meaningful.

### 2. The planned verification matrix is too generic and misses a known docs-generation gap

- **Severity:** Medium
- **Files:** `web/data/tools.json`, `cmd/gendocs/testdata/tools.golden.json`, `web/content/reference/tools.md`

The plan says “Targeted tests passing” and then `make test`, `make build`, `make lint`. R008 already demonstrated that those gates can pass while `web/data/tools.json` remains stale; I also verified that `web/data/tools.json` still has no `get_fitness_projection` entry.

Add an explicit docs/catalog verification step, for example:

```sh
make docs-tools
git diff --exit-code web/data/tools.json cmd/gendocs/testdata/tools.golden.json web/content/reference/tools.md
```

or the equivalent generator command plus a clean-diff assertion. Without this, the public reference can still omit the new tool even after Step 5 passes.

### 3. Targeted tests should be enumerated for the known API-contract edges

- **Severity:** Medium
- **Files:** `internal/tools/get_fitness_projection.go`, `internal/tools/get_fitness_projection_test.go`

Given the previous reviews, “Targeted tests passing” is not specific enough. The Step 5 plan should name the targeted packages/cases it will run and confirm coverage for:

- `include_full:true` with an explicit zero planned load or `recovery_week_load_pct: 0`, ensuring `training_load: 0` is serialized rather than omitted.
- Omitted horizon behavior, whichever contract is chosen: default `horizon_days` or required exactly-one `horizon_date`/`horizon_days`.
- `recovery_week_cadence: 1`, ensuring schema, decoder, and user-facing error text agree.
- Analyzer `_meta` assumptions and terse/full response shaping after any fixes.

Recommended targeted command set:

```sh
go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs ./internal/safety
```

followed by the full gates already listed in the plan.

### 4. STATUS update work is part of verification for this task

- **Severity:** Low
- **File:** `taskplane-tasks/TP-095-fitness-projection/STATUS.md`

The plan includes documenting pre-existing unrelated failures, but it does not mention correcting the existing review history. STATUS still records rejected reviews as approvals in Notes and leaves the Reviews table empty. Step 5 should explicitly require STATUS to record commands run, outcomes, unresolved findings, and accurate review verdicts before moving to Step 6.

## Suggested revised Step 5 checklist

- [ ] Resolve or explicitly block on all R008 findings.
- [ ] Run targeted tests for projection engine/tool/catalog/docs/safety packages.
- [ ] Regenerate and verify generated docs/catalog artifacts (`make docs-tools` plus clean diff or committed updates).
- [ ] Run full suite: `make test`.
- [ ] Run build: `make build`.
- [ ] Run lint: `make lint`.
- [ ] Update STATUS with accurate review history, command outputs, failures, and whether failures are pre-existing/unrelated.
