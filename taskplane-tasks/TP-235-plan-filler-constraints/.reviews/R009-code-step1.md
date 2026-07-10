# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking finding

### `0` cannot express a zero-session or zero-load/time hard constraint

`WeekConstraints` uses plain numeric fields and assigns `0` the meaning “no cap”: `ValidateCandidates` only enforces `RequestedSessionCount` when it is greater than zero (`internal/planning/constraints.go:525`), and `validateCore` skips weekly time/load checks when each target is zero (`:698`, `:726`). This makes an explicitly requested zero-session week indistinguishable from an omitted session-count constraint, and an explicit zero-hour/zero-load target indistinguishable from no weekly budget.

That violates the task’s required separation of requested sessions (“how many should be placed”) from availability, and permits over-scheduling in the most restrictive case. A focused regression check with `RequestedSessionCount: 0`, an available day, and one positive candidate returns `Valid: true`.

Represent presence separately from a zero value (for example, pointer/optional target and session-count fields), or make zero a hard zero cap and introduce an explicit unlimited/unset representation. Apply the same unambiguous contract to requested count, weekly time, and weekly load; update the design document and boundary table; add tests covering explicit zero versus unset/unlimited.

## Required coverage

`internal/planning` still has no committed test file, despite the change now containing the complete 939-line validator and the STATUS marking Step 2’s implementation checks complete. `go test ./internal/planning` only reports `[no test files]`. Add the requested `constraints_test.go` coverage with this fix, including the zero/unset cases and the prior R008 matching/reconciliation regressions, before relying on this package as the Plan Filler safety boundary.

## Documentation correction

The `Reconciliation` table says `CandidateMinutes` and `CandidateLoad` are the sum of **all** candidates, while the code and surrounding prose exclude invalid-input candidates. Change those table entries to “valid-input candidates” so the reconciliation contract is internally consistent.

## Verification

- `go test ./internal/planning` passes but reports `[no test files]`.
- `go vet ./internal/planning` passes.
- `git diff --check 6a30598..HEAD` passes.
