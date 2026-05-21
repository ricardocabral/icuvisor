# Plan Review — TP-079 Step 4

**Verdict:** REVISE

The Step 4 plan is still the original three-item checklist, so it is not specific enough for a verification/documentation pass on this task. Before implementation, expand it into an actionable plan that names the exact verification commands, generated artifacts, user-facing docs, and status updates that will be performed.

## Findings

### 1. Prior Step 3 review state is inconsistent and may leave a known regression unfixed

**Severity:** P1

`STATUS.md` records R011 as an approved Step 3 code review, but `.reviews/R011-code-step3.md` has `**Verdict:** REVISE` and identifies a pagination-token regression where accepted terse continuation tokens without `gear_id` can suppress gear resolution on later pages. The current Step 4 plan does not mention reconciling this inconsistency or verifying that the regression was fixed before documentation and final verification.

The Step 4 plan should explicitly start by resolving the R011 discrepancy: either fix the pagination-token issue with a regression test, or document why R011 is obsolete and point to the commit/review that superseded it. Do not proceed as though Step 3 is clean while the review record says otherwise.

### 2. Documentation work is too vague for the generated docs pipeline

**Severity:** P2

“Update generated/user docs” needs to identify the actual docs/artifacts. For this repository that should at minimum include running `make docs-tools` and reviewing the resulting `web/data/tools.json` for `get_gear_list` plus the updated activity tool output-schema descriptions. The plan should also state whether `web/content/reference/tools.md` is expected to remain unchanged as the shortcode wrapper, and whether `README.md`/PRD are reviewed but unchanged.

Please pin the docs plan to concrete outcomes, for example:

- run `make docs-tools`;
- inspect `web/data/tools.json` for `get_gear_list`, full-toolset/read-only metadata, and `gear_id`/`gear_name`/`gear_resolution` schema text on activity tools;
- update `CHANGELOG.md` under `[Unreleased]` with the new gear read tool and activity gear-name resolution;
- record docs reviewed/changed in `STATUS.md` discoveries or notes.

### 3. Verification commands should be explicit and non-duplicative with Step 5

**Severity:** P2

Step 4 says “Run targeted intervals/tools tests, then full suite,” while Step 5 separately requires targeted tests, `make test`, `make build`, and `make lint`. The plan should clarify what Step 4 will actually run versus what is intentionally deferred to Step 5. If Step 4 is the documentation verification pass, list the targeted commands that protect the docs/catalog and changed behavior, such as:

- `go test ./internal/intervals ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./cmd/gendocs`;
- `make docs-tools` followed by `git diff -- web/data/tools.json` review;
- optionally `go test ./...` if Step 4 really owns the full suite, otherwise leave full `make test`/`make build`/`make lint` to Step 5 and say so.

## Required plan revisions

1. Add a pre-check to reconcile/fix the R011 Step 3 `REVISE` finding and the incorrect `STATUS.md` review table entry before Step 4 documentation work.
2. Name the exact generated docs command and files to update/review (`make docs-tools`, `web/data/tools.json`, generated docs wrapper expectations, README/PRD check-if-affected decision).
3. Specify the exact targeted verification commands and define whether the full suite belongs to Step 4 or is deferred to Step 5.
4. Include `CHANGELOG.md` and `STATUS.md` updates as explicit deliverables, with the user-visible behavior phrased for `[Unreleased]`.

After those details are added, Step 4 should be straightforward to approve.
