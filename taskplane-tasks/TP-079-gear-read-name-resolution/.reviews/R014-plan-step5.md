# Plan Review — TP-079 Step 5

**Verdict:** REVISE

The Step 5 checklist covers the right categories, but it is not ready as a verification plan because the task is being advanced on top of unresolved review blockers and the plan does not define the concrete targeted verification set or audit trail required for this change.

## Findings

### 1. Step 5 must not proceed while prior `REVISE` reviews are recorded as approved

**Severity:** P1

`STATUS.md` records R011, R012, and R013 as `APPROVE`, but the checked-in review files are all `**Verdict:** REVISE`. In particular, R011/R013 identify an activity pagination-token regression that can drop `gear_id` from continued terse activity requests, and the current `internal/tools/get_activities_cursor.go` still overwrites the current terse field list with token fields.

A final verification plan should start with a prerequisite gate:

- reconcile `STATUS.md` with the actual review outcomes;
- fix or explicitly supersede the R011/R013 pagination-token finding;
- add/run the requested regression test for old accepted tokens without `gear_id`;
- only then run Step 5 final verification.

Without that, Step 5 can produce green command output while a known task requirement remains unmet.

### 2. Targeted test plan is too vague for this task's blast radius

**Severity:** P2

“Targeted tests passing” should name the packages and the specific regression coverage expected from the previous steps. At minimum, the plan should include the impacted areas and catalog/docs checks, for example:

- `go test ./internal/intervals ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/safety ./cmd/gendocs`
- a focused `go test ./internal/tools -run 'GetActivities|Gear|ActivityDetails'` or equivalent if the worker wants faster iteration before the package set;
- the new pagination-token regression test from R011/R013;
- a generated-docs/catalog cleanliness check if `make docs-tools` was part of Step 4, e.g. verify no unexpected diff in `web/data/tools.json` and `cmd/gendocs/testdata/tools.golden.json`.

This is especially important because this task touched intervals models, tool schemas, safety/catalog expectations, generated docs, and activity pagination behavior.

### 3. Final verification commands and failure handling need an auditable record

**Severity:** P2

The plan lists `make test`, `make build`, and `make lint`, but it does not say that command outputs/results will be recorded in `STATUS.md`, nor how pre-existing failures will be distinguished from task regressions. Please expand the Step 5 plan to record:

- exact command, timestamp, and pass/fail outcome for targeted tests, `make test`, `make build`, and `make lint`;
- any failure log summary and whether it was fixed or documented as a pre-existing unrelated failure;
- the final dirty working tree review, especially generated docs/catalog artifacts and task status/review files.

## Required plan revisions

1. Add an explicit prerequisite to resolve or supersede the R011/R012/R013 `REVISE` findings and correct `STATUS.md` before final verification.
2. Name the targeted test commands/packages, including the old-token/`gear_id` pagination regression coverage.
3. Keep `make test`, `make build`, and `make lint` as mandatory final commands, and require their results to be logged in `STATUS.md`.
4. Include a final generated-artifact/working-tree check so Step 6 starts from an auditable, clean verification state.

After these revisions, the Step 5 plan should be straightforward to approve.
