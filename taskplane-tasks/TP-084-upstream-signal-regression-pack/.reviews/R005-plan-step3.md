# Plan Review — TP-084 Step 3

**Verdict:** REQUEST CHANGES

I reviewed `PROMPT.md`, `STATUS.md`, the Step 2 review, and the current test state. The Step 3 plan in `STATUS.md` is still only the generic checklist (“apply minimal fixes, keep schema-stable, run targeted tests”). That is not specific enough for this step, especially because the prior code review found blocking gaps in the regression pack and those gaps are not reflected in the Step 3 plan.

## Blocking plan gaps

1. **Do not proceed as though Step 2 is clean.**
   `R004-code-step2.md` requested changes, but `STATUS.md` currently records it as `APPROVE` and marks Step 2 complete. Step 3 must explicitly start by resolving or accounting for those review findings, otherwise it may “fix regressions” against an incomplete pack.

2. **Plan must name the exact regressions/failures to fix.**
   The current targeted tests pass (`go test ./internal/tools ./internal/intervals`), which means Step 3 has no concrete failing regression identified. Before production fixes, the plan should say whether it will first add the missing failing coverage from R004, then fix only the failures exposed by that coverage.

3. **`get_event_by_id` behavior needs an explicit decision.**
   The task mission says listed/detail-404 mismatch should remain `unavailable.reason: "upstream_inconsistency"`, while existing code recovers an event from the fallback list when the ID is present. The Step 3 plan should state the intended minimal code change and corresponding test updates, e.g. update the fixture-backed test to feed the listed event into the fake list response, expect structured `upstream_inconsistency`, and adjust/remove the old recovery expectation if it conflicts with the product contract.

4. **Strava list-output coverage should be completed before declaring no code fix needed.**
   R004 noted that numeric/no-`i` sync-chain fixtures are not asserted through the `get_activities` handler output. Step 3 should include adding or correcting that handler-level test before deciding whether `get_activities_row.go` needs no production changes.

5. **Verification commands should be non-cached and targeted.**
   The plan should list commands such as `go test -count=1 ./internal/tools ./internal/intervals`, plus any focused `-run` invocations for the changed tests. Cached success is not enough for a regression-fix step.

## Minor requirements

- Fix the `STATUS.md` audit trail: record R004 with the correct `REQUEST CHANGES` verdict and keep review/execution entries in the intended tables.
- Keep NOTE handling as a no-op unless a strengthened test actually fails; the existing wire-level serialization appears already covered.
- If Step 3 changes user-visible event recovery semantics, plan to update `CHANGELOG.md` in Step 4 as required by the task prompt.

## Summary

The direction of Step 3 is valid, but the plan is too vague and does not account for unresolved Step 2 review findings. Please revise it to list the concrete tests to complete, the expected failing behavior (especially `get_event_by_id`), the minimal files/functions to touch, and the exact targeted verification commands.
