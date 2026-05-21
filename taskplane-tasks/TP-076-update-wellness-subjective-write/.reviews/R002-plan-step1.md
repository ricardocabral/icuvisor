# Plan Review: Step 1 live probe

Result: approved.

The revised Step 1 plan addresses the blocking issues from R001. It now starts from the production `UpdateWellness` request shape (`PUT /athlete/{id}/wellness/{YYYY-MM-DD}` with a sparse body), treats alternate method/date/scoping shapes as fallback probes only, adds a secret-safe production-account guard, snapshots the target row before mutation, tests `locked` last, and requires restore/re-read verification.

## Notes for execution

- Keep all scratch scripts and raw probe logs out of the repo; only commit sanitized fixtures.
- When saving fixtures, use the task-specified fixture paths/names unless the implementation later justifies different ones:
  - `internal/intervals/testdata/wellness/subjective_write_request.json`
  - `internal/intervals/testdata/wellness/subjective_write_response.json`
- The restore plan should preserve the pre-probe `locked` value exactly; if the original row was unlocked or absent, ensure the probe does not leave it locked.
- If the pre-probe row has existing values or absent/null subjective fields, verify the API supports restoring that state before running broad bundle probes; otherwise switch to a clearly disposable date/row.

No further plan changes are required before running Step 1.
