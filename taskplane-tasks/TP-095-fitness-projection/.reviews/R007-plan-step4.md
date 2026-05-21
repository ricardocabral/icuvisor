# R007 Plan Review — Step 4: Docs and verification

**Verdict:** Needs changes / not yet reviewable

I read `PROMPT.md`, `STATUS.md`, the prior Step 3 review, and ran a focused verification command against the packages currently known to be affected by registration/docs fixtures.

## Blocking issues in the Step 4 plan

### 1. Step 4 must explicitly carry forward unresolved R006 blockers

The current Step 4 plan is only the three checklist bullets from `STATUS.md`:

- update docs/reference with assumptions;
- update `CHANGELOG.md`;
- run full quality gate.

That is not enough because Step 3's code review was a **Request changes**, not an approval. The Step 4 plan should first state how it will resolve or verify the R006 findings before moving to docs-only work:

- update the generated-docs golden/catalog artifacts for the newly registered tool;
- update the safety static catalog matrix/count expectations for `get_fitness_projection`;
- preserve explicit zero `training_load` values in full-series output and add coverage for that case;
- align the public schema, decoder behavior, and user-facing invalid-argument message for horizon defaults/requirements and `recovery_week_cadence` boundaries;
- fix `STATUS.md` review history so rejected reviews are not recorded as approved and blockers/discoveries are visible.

Without this, Step 4 can falsely pass documentation work while CI is still broken.

### 2. The plan needs a concrete documentation inventory

Please spell out the exact docs/artifacts to update. For this task, that should include at least:

- `web/content/reference/tools.md`: add a concise human-readable assumptions note for `get_fitness_projection` near the generated catalog, including that projections are deterministic modeled scenarios, not predictive certainty; the seed comes from a current `get_fitness` row; curve output requires `include_full:true`; and the relevant defaults/bounds once the schema contract is finalized.
- `web/data/tools.json`: regenerate via `make docs-tools` so the public catalog contains `get_fitness_projection`.
- `cmd/gendocs/testdata/tools.golden.json`: update the generated-docs golden if the test expects the new catalog descriptor there.
- `CHANGELOG.md`: add an `[Unreleased]` Added entry for the new full-toolset analyzer-family projection tool.

Also explicitly note whether `README.md` and `docs/prd/PRD-icuvisor.md` were reviewed and why they do or do not need changes.

### 3. Verification plan must include the currently failing packages before the full gate

I ran:

```sh
go test ./cmd/gendocs ./internal/safety ./internal/tools ./internal/toolcatalog
```

It currently fails in `cmd/gendocs` because `testdata/tools.golden.json` is stale/missing `get_fitness_projection`, and in `internal/safety` because the static catalog expectations still omit the new read tool. The Step 4 plan should call out these targeted checks before the broader gate, then run the required full quality gate:

```sh
go test ./cmd/gendocs ./internal/safety ./internal/analysis ./internal/tools ./internal/toolcatalog
make test
make build
make lint
```

If any failure is pre-existing and unrelated, document it in `STATUS.md`; otherwise fix it before marking Step 4 complete.

## Recommendation

Revise the Step 4 plan in `STATUS.md` (or add a short referenced plan note) with the blocker carry-forward, exact docs/generated artifacts, and ordered verification commands above. Do not mark Step 4 complete until the known registration/golden failures and the R006 contract issues are resolved and documented.
