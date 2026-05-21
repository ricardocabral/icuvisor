# R011 Plan Review — Step 5: Testing & Verification

**Verdict:** Request changes

## Findings

### 1. Step 5 plan has not been updated after the prior Step 5 plan rejection

- **Severity:** Medium
- **Files:** `taskplane-tasks/TP-095-fitness-projection/STATUS.md`, `taskplane-tasks/TP-095-fitness-projection/.reviews/R009-plan-step5.md`

R010 confirms the Step 4 recovery fixed the earlier code/docs blockers, so Step 5 is no longer blocked by unresolved R008 findings. However, the Step 5 plan in `STATUS.md` is still only the generic checklist:

- targeted tests;
- `make test`;
- `make build`;
- `make lint`;
- document unrelated failures.

That does not incorporate the concrete verification matrix requested in R009, and it leaves too much ambiguity about what "targeted tests" means for this new public tool. Before running Step 5, revise the plan/checklist to name the targeted packages and the generated-docs/catalog verification that must stay clean.

Recommended Step 5 command sequence:

```sh
go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs ./internal/safety
make docs-tools
git diff --exit-code web/data/tools.json cmd/gendocs/testdata/tools.golden.json web/content/reference/tools.md
make test
make build
make lint
```

If `make docs-tools` intentionally changes generated artifacts, commit or document those changes before marking the gate complete.

### 2. The plan should explicitly verify the prior contract-risk cases

- **Severity:** Medium
- **Files:** `internal/tools/get_fitness_projection_test.go`, `internal/tools/testdata/analyzer/fitness_projection_recovery_week_full.golden.json`

The Step 4 code review approved the fixes for the known edge cases, but Step 5 should still call them out so the final verification does not regress them while fixing any gate failures. Add explicit targeted-test expectations for:

- `include_full:true` preserving explicit zero `training_load` values;
- default/omitted horizon behavior matching the schema and decoder;
- `recovery_week_cadence: 1` being accepted consistently;
- analyzer `_meta` assumptions and terse/full response shaping.

These can be satisfied by the targeted package command above if the plan states that these cases are included in the targeted test pass.

### 3. STATUS cleanup is part of the Step 5 verification work

- **Severity:** Low
- **File:** `taskplane-tasks/TP-095-fitness-projection/STATUS.md`

`STATUS.md` currently has a dangling execution-log row under `## Notes`:

```md
| 2026-05-20 18:45 | Review R010 | code Step 4: APPROVE |
```

Step 5 already requires documenting command outcomes and unrelated failures. Include a small STATUS cleanup/update item in the plan so the final verification record is well-formed: move that row into the execution log (or remove it), add the R011 review entry, and record the exact Step 5 commands and results.

## Suggested revised Step 5 checklist

- [ ] Run targeted projection/tool/catalog/docs/safety tests: `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs ./internal/safety`.
- [ ] Verify generated docs/catalog artifacts: `make docs-tools` plus a clean diff for `web/data/tools.json`, `cmd/gendocs/testdata/tools.golden.json`, and `web/content/reference/tools.md`.
- [ ] Confirm targeted coverage for zero-load serialization, horizon defaults, cadence boundary, analyzer `_meta`, and terse/full response shaping.
- [ ] Run full test suite: `make test`.
- [ ] Run build: `make build`.
- [ ] Run lint: `make lint`.
- [ ] Fix failures or document pre-existing unrelated failures in `STATUS.md`.
- [ ] Update `STATUS.md` with well-formed review history, execution-log entries, and final Step 5 outcomes.
