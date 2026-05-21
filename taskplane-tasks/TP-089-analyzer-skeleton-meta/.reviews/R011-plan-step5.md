# Plan Review R011 — Step 5: Testing & Verification

Verdict: **REVISE**

I read `PROMPT.md`, `STATUS.md`, the prior review files, and the current analyzer helper/test files. I also ran `go test ./internal/analysis ./internal/tools` and `make lint`. The Step 5 checklist has the right high-level gates, but it is not safe to execute as written because it treats earlier verification as complete even though there are known unresolved review findings and a current lint failure.

## Findings

1. **Step 5 must explicitly fix the known lint failure before rerunning the gate.**
   `make lint` still fails with:
   ```
   internal/tools/analyzer_common.go:35:6: func encodeAnalyzerResponse is unused (unused)
   ```
   The unused helper is still present at `internal/tools/analyzer_common.go:35-41`, and `grep` finds no caller. This is TP-089 code, so it cannot be documented as a pre-existing unrelated failure. The Step 5 plan needs a concrete first item to either remove the unused wrapper or add an intended caller/test that makes it used, then rerun `make lint`.

2. **The plan does not resolve the mandatory zero-value `_meta` coverage gap called out by R008/R009/R010.**
   Current tests still only shape populated demo metadata: `internal/analysis/meta_test.go:8-36` uses non-empty method/tools and `insufficient_sample=true`, while `internal/tools/analyzer_common_test.go:14-50` shapes `n=7`, `missing_days=2`, and a non-empty `source_tools`. There is still no minimal/zero-value marshal or shaped-response test proving the mandatory fields survive with `source_tools: []`, `n: 0`, `missing_days: 0`, `insufficient_sample: false`, and `formula_ref` omitted. Step 5 should not be just a command-running pass; it must include this missing contract test before checking “Targeted tests passing.”

3. **The review audit trail is still inaccurate, and Step 5 does not plan to repair it.**
   `STATUS.md:99-103` records R006/R007/R008/R009/R010 as `APPROVE`, and `STATUS.md:129-133` logs those review outcomes as `APPROVE`. The review files themselves say `REVISE` for R006, R007, R008, R009, and R010. In particular, `.reviews/R010-code-step4.md:3` is `Verdict: **REVISE**`, but `STATUS.md:103` says `APPROVE`. Step 5 must reconcile the table/log to preserve actual historical verdicts and add separate follow-up rows for fixes or later approvals instead of rewriting review outcomes.

4. **The verification command list is incomplete for the known workspace issues.**
   The prompt requires targeted tests, `make test`, `make build`, and `make lint`; R010 also identified whitespace problems in the committed review artifact. `git diff --check 2a15dca..HEAD` still reports trailing whitespace in `.reviews/R009-plan-step4.md` on lines 9, 12, 15, and 18. The Step 5 plan should include `git diff --check` (or the relevant commit-range check for already-committed task changes) after fixing those review-file whitespace issues, and it should record each command and result in `STATUS.md`.

## Suggested revised Step 5 plan

- First clear unresolved review findings from R008/R009/R010:
  - add a minimal/zero-value analyzer meta test, preferably through the shaped response boundary, asserting all mandatory `_meta` keys and `source_tools` as an empty JSON array rather than `null`;
  - remove or use `encodeAnalyzerResponse` so `make lint` has no unused-function failure;
  - remove trailing whitespace from `.reviews/R009-plan-step4.md`.
- Reconcile `STATUS.md` before marking Step 5 complete:
  - update the Reviews table and Execution Log so R006/R007/R008/R009/R010 show their actual `REVISE` verdicts;
  - add new execution-log rows for the follow-up fixes and this R011 plan review, rather than changing old review results to `APPROVE`.
- Run and record the exact verification commands after fixes:
  - `go test ./internal/analysis ./internal/tools`;
  - `make test`;
  - `make build`;
  - `make lint`;
  - `git diff --check` and, if checking committed task history, `git diff --check 2a15dca..HEAD` or the appropriate base range.
- If any command still fails, record the command, exit status, failure summary, and why it is pre-existing/unrelated; otherwise fix it before checking the Step 5 boxes.

Once those items are in the plan, Step 5 will be a reliable final verification pass instead of repeating the incomplete Step 4 verification.
