# R009 Plan Review — Step 4: Docs and full verification

**Verdict:** REVISE

## Findings

### P1 — Step 4 is premature while Step 3 still has a REVISE review

`STATUS.md:48-53` marks Step 3 complete and `STATUS.md:98-99` records R007/R008 as approved, but the checked-in review files disagree: `R007-plan-step3.md` is `REVISE`, and `R008-code-step3.md:3` is also `REVISE` with unresolved test-coverage/bookkeeping findings. The current Step 4 plan (`STATUS.md:57-62`) does not mention resolving those blockers before moving to docs/full verification.

Revise the Step 4 plan to first close the Step 3 review debt, or move the task back to Step 3. At minimum, explicitly plan to:

- fix the R008 wellness `field_semantics` assertion so missing map keys fail the test;
- add the missing absent-nutrition metadata assertion for wellness;
- add `get_activity_details` zero-calorie and absent-calorie coverage;
- correct the R007/R008 verdicts/execution log in `STATUS.md` so the task history matches the review files;
- rerun the targeted read-tool tests after those fixes.

Without that, the full quality gate in Step 4 can pass while the task still lacks the regression coverage the previous step required.

### P2 — The docs plan does not identify the generated tool-reference path

The plan only says “Update tool reference/examples” (`STATUS.md:60`), but `web/content/reference/tools.md:6-8` is just a generated-catalog wrapper, and the repository’s docs target regenerates `web/data/tools.json` from the registry (`Makefile:90-91`). A plan to hand-edit `web/content/reference/tools.md` would not document the changed return shape in the generated catalog and would be easy to overwrite or drift.

Please expand the Step 4 plan with the exact docs mechanism:

- update the registered tool/output-schema descriptions for the affected reads (`get_activities`, `get_activity_details`/`activityReadOutputSchema`, and `get_wellness_data`) so the generated catalog names the public keys and semantics: activity `calories_burned` means active/exercise calories; wellness nutrition is `calories_intake`, `carbs_g`, `protein_g`, `fat_g`; no `calories_total` or activity macro keys are emitted without upstream evidence;
- run `make docs-tools` and commit the regenerated `web/data/tools.json` if it changes;
- if examples are added outside the generated catalog, name the concrete file(s) in the plan and keep them consistent with the same key names.

### P2 — “Run full quality gate” needs concrete commands and failure handling

`STATUS.md:62` does not specify what the full quality gate is. The task completion criteria require `make test`, `make build`, and `make lint`, while R008 also needs a targeted `go test -count=1 ./internal/tools` after the test fixes. Because Step 5 repeats the verification checklist, Step 4 should still state whether these commands are run now and how their results flow into Step 5.

Add the concrete command list and status-update rule to the plan, e.g.:

```sh
go test -count=1 ./internal/tools
make docs-tools
make test
make build
make lint
```

Then record all outcomes in `STATUS.md`, including any unrelated/pre-existing failures rather than checking the gate complete.

### P3 — Changelog wording should be part of the plan

The changelog update is required and user-visible, but the plan does not constrain the wording. To avoid reintroducing ambiguity, plan the entry under `[Unreleased]` with the exact public contract: activity reads expose `calories_burned` for active/exercise calories, wellness rows expose intake/macros as `calories_intake`, `carbs_g`, `protein_g`, and `fat_g`, and absent upstream nutrition remains omitted. Do not mention unsupported `calories_total` or activity macros as shipped fields.

## Verification

Reviewed:

- `PROMPT.md`
- `STATUS.md`
- `.reviews/R007-plan-step3.md`
- `.reviews/R008-code-step3.md`
- `web/content/reference/tools.md`
- `Makefile`
