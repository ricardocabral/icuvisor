# Plan Review — TP-084 Step 4

**Verdict:** REQUEST CHANGES

I reviewed `PROMPT.md`, `STATUS.md`, and the prior Step 2/3 reviews. The Step 4 checklist is directionally aligned with the prompt, but the task is not ready to move into verification/documentation because unresolved review blockers are still recorded as approved and Step 3 was marked complete without addressing them.

## Blocking plan gaps

1. **Do not proceed to Step 4 until Step 3 is actually resolved.**  
   `R006-code-step3.md` requested changes, but `STATUS.md` records `R006` as `APPROVE`, marks Step 3 complete, and says no production-code fixes were required. The Step 4 plan must first require reconciling that state: either resolve the Step 2/3 blockers in tests/code/status, or explicitly move the task back to Step 3. Verification over a known-incomplete regression pack is not meaningful.

2. **The plan must account for the still-open regression coverage blockers.**  
   The prior reviews identified two blocking gaps that remain unresolved in the committed history after Step 2: the listed-event/detail-404 `get_event_by_id` mismatch is not locked against recovery semantics, and the numeric/no-`i` Strava sync-chain fixture is not asserted through `get_activities` structured output. Step 4 cannot be just `make test` / `make build` / `make lint`; it needs an explicit precondition that those findings are fixed or a documented, reviewed decision explaining why they no longer apply.

3. **Fix the audit trail before documenting completion.**  
   The Reviews table is still empty while review rows are appended under `## Notes`, and several verdicts are recorded incorrectly as `APPROVE` despite review files saying `REQUEST CHANGES`. Step 4's documentation plan should explicitly include moving those entries into the proper sections and recording the real verdicts before adding fixture provenance/final notes.

4. **Specify verification commands and failure handling.**  
   The checklist says “Run full suite/build/lint,” but the plan should name the exact commands required by the prompt, at minimum `make test`, `make build`, and `make lint`, plus the targeted regression command already used (`go test -count=1 ./internal/tools ./internal/intervals`) after any remaining fixes. It should also state that any failures are fixed or documented as pre-existing unrelated failures in `STATUS.md`.

## Minor recommendations

- Keep `CHANGELOG.md` unchanged if the task only added regression tests and did not alter user-visible behavior; document that decision in `STATUS.md`.
- Record fixture provenance/redaction notes with enough detail to show the fixtures are sanitized and contain no live secrets or athlete-identifying data.
- If any user-visible semantics are changed while resolving the open blockers, update `CHANGELOG.md` under `[Unreleased]` during this step.

## Summary

The Step 4 goals are correct, but the plan is not approvable while Step 2/3 blockers and the incorrect status history remain unresolved. Revise the plan to gate Step 4 on resolving those findings, correct the audit trail, and list the exact verification/documentation actions before proceeding.
