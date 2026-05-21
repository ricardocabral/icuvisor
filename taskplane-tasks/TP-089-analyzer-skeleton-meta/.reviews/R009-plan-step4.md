# Plan Review R009 — Step 4: Verify

Verdict: **REVISE**

I read `PROMPT.md`, `STATUS.md`, `taskplane-tasks/CONTEXT.md`, the prior Step 3 reviews, and the current analyzer test files. The Step 4 checklist is heading in the right direction, but the plan is not safe to execute yet because it treats Step 3 as complete even though the recorded review files still contain unresolved `REVISE` findings.

## Findings

1. **Step 4 must first resolve the outstanding Step 3 code-review finding.**  
   `.reviews/R008-code-step3.md:3-8` says `Verdict: **REVISE**` and calls out missing coverage for the mandatory zero-value analyzer `_meta` contract. The current tests still only cover populated demo metadata: `internal/analysis/meta_test.go:8-36` uses non-empty method/tools with `insufficient_sample=true`, and `internal/tools/analyzer_common_test.go:14-28` shapes the populated demo response. There is still no minimal/zero-value marshal or shaped-response test proving `source_tools: []` rather than `null`, `n: 0`, `missing_days: 0`, and `insufficient_sample: false` are emitted. A verify step should not proceed as if all previous code review findings are closed.

2. **The review audit trail is inaccurate and the Step 4 plan does not mention correcting it.**  
   `STATUS.md:99-101` and `STATUS.md:127-129` record R006/R007/R008 as `APPROVE`, but the review files themselves show `REVISE` for all three (`.reviews/R006-code-step2.md:3`, `.reviews/R007-plan-step3.md:3`, `.reviews/R008-code-step3.md:3`). Step 4's “Record conventions in STATUS.md” is too vague to catch this. The plan needs an explicit audit-reconciliation item: preserve the actual review verdicts, then add separate execution-log rows for any follow-up fixes/reviews.

3. **“Run full quality gate” is too underspecified for the verification step.**  
   The prompt's completion criteria require `make test`, `make build`, and `make lint`; Step 3 also had targeted tests. The Step 4 plan should list the exact commands to run and record, for example: `go test ./internal/analysis ./internal/tools`, `make test`, `make build`, `make lint`, and optionally `git diff --check`. If Step 5 is intended to repeat these gates, Step 4 should say so explicitly; otherwise the current Step 4/Step 5 split makes it unclear which step owns failures and status updates.

4. **The CHANGELOG decision needs to be explicit.**  
   `CHANGELOG.md` already records similar analyzer-family scaffolding under `[Unreleased]` (for example the `analysis_metric` helper). The Step 4 plan should either add a concise `[Unreleased]` bullet for the reusable analyzer `_meta`/response skeleton, or record in `STATUS.md` why this internal-only scaffolding is not user-visible. “Update CHANGELOG.md if public scaffolding is visible” leaves the worker to rediscover the policy during verification.

## Suggested revised Step 4 plan

- Reopen/complete the unresolved Step 3 follow-up before verification:
  - add a zero/minimal analyzer meta test that marshals or shapes through the response boundary;
  - assert all mandatory `_meta` keys are present with zero/false/empty values, especially `source_tools` as an empty JSON array;
  - run `go test ./internal/analysis ./internal/tools`.
- Reconcile `STATUS.md` against the actual review files:
  - set R006/R007/R008 verdicts to `REVISE`;
  - add execution-log rows describing the follow-up fixes and any new review outcomes, instead of rewriting history.
- Run and record the quality gate commands: `go test ./internal/analysis ./internal/tools`, `make test`, `make build`, and `make lint` (plus `git diff --check` if desired). Document any unrelated pre-existing failure with the command, failure summary, and why it is unrelated.
- Decide the changelog outcome before marking Step 4 complete: either add an `[Unreleased]` bullet for the analyzer `_meta` scaffolding or record the no-changelog rationale in `STATUS.md`.
- Add a short downstream-conventions note in `STATUS.md` that future analyzer tools must use the shared helper, keep `source_tools` non-nil, default missing data to `skip`, and gate heavy series behind `include_full`.

Once those items are in the Step 4 plan, verification should be straightforward.
