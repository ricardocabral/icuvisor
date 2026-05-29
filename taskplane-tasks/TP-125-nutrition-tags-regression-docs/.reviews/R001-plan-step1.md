# Plan Review: Step 1 — Audit existing coverage

Verdict: Approved

The Step 1 plan is appropriately scoped for an audit-only step: it checks the three relevant test files, records coverage gaps in `STATUS.md`, and runs the targeted `go test ./internal/tools` suite before moving to regression/doc changes.

Notes for execution:
- Be strict about the task requirement "without requiring `include_full:true`". Existing coverage should be classified separately for terse/default responses versus `include_full` preservation.
- During the audit, record any gaps precisely. In the current tests, fueling fields appear covered in terse/default paths for `get_activities` and `get_activity_details`, while tag coverage appears stronger for `include_full` than for terse/default activity reads and should be verified/recorded accordingly.
- `get_today` already appears to assert tag preservation for completed activities and planned/annotation events in the terse digest; still record this outcome in the audit notes.
- Keep Step 1 to observation and test execution only; add or adjust regression tests/docs in Step 2.

Reviewer check: `go test ./internal/tools` passes in this workspace.
