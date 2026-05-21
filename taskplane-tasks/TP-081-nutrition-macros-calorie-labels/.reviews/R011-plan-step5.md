# R011 Plan Review — Step 5: Testing & Verification

**Verdict:** REVISE

## Findings

### P1 — Step 5 cannot proceed as a pure quality-gate run while prior REVISE reviews remain unresolved

The current Step 5 plan is only the generic verification checklist in `STATUS.md:67-74`, but the checked-in review files still contain blocking revision requests: `.reviews/R007-plan-step3.md:3`, `.reviews/R008-code-step3.md:3`, `.reviews/R009-plan-step4.md:3`, and `.reviews/R010-code-step4.md:3` are all `REVISE`. `STATUS.md:98-101` and `STATUS.md:141-144` record those same reviews as approved, so the plan is starting verification from an inconsistent task state.

This is not just bookkeeping. R010 still identifies a public MCP schema issue: `activityReadOutputSchema()` at `internal/tools/get_activity_details.go:293-294` does not document `calories_burned` semantics for activity-detail responses. A Step 5 plan that only runs tests/build/lint will not catch or fix that documentation/schema contract gap.

Revise the Step 5 plan to first close the outstanding review debt before checking verification boxes. At minimum, explicitly plan to:

- resolve the remaining R010 schema finding for `get_activity_details` / `activityReadOutputSchema()`;
- verify whether the R007/R008/R009 findings have actually been fixed, or move the task back to the appropriate earlier step if they have not;
- correct the Reviews table and execution log so they match the checked-in review files or add follow-up approval reviews that supersede the REVISE outcomes;
- only then run the final Step 5 verification commands.

### P2 — The targeted verification command is still unspecified

`STATUS.md:70` says “Targeted tests passing,” but it does not name the package/command. For this task, the targeted command has already been called out by prior reviews as `go test -count=1 ./internal/tools`, because the changed behavior is in the read-tool response shaping and tests. If the schema/doc fix requires regenerated docs, the plan should also include `make docs-tools` before the full gate so generated catalog drift is caught.

Please expand Step 5 with the concrete command list and order, for example:

```sh
go test -count=1 ./internal/tools
make docs-tools
make test
make build
make lint
```

### P2 — The plan needs an explicit STATUS recording rule for command outcomes

R010 already found that quality-gate commands were run but not recorded in `STATUS.md`. The Step 5 plan currently has checkboxes but no rule to log command outcomes or document unrelated/pre-existing failures. Since the prompt requires failures to be fixed or documented in `STATUS.md`, add a plan item to record each command and result in the execution log (or directly under Step 5) before marking the checkbox complete.

## Verification

Reviewed:

- `PROMPT.md`
- `STATUS.md`
- `.reviews/R007-plan-step3.md`
- `.reviews/R008-code-step3.md`
- `.reviews/R009-plan-step4.md`
- `.reviews/R010-code-step4.md`
- `internal/tools/get_activity_details.go`
