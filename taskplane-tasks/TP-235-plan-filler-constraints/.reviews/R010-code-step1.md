# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. Reconciliation loses the required nil-versus-zero target distinction

`WeekConstraints` now correctly uses pointers so that `nil` means “untracked/unlimited” while `&0` is a hard zero budget. `buildReconciliation` immediately erases that distinction by dereferencing both fields into non-pointer output fields (`internal/planning/constraints.go:820-825`). Consequently, JSON output with `weekly_target_load: 0` can mean either no load budget or an explicit zero budget, despite those inputs having opposite validation behaviour.

It also contradicts the exported `Reconciliation` contract: `RemainingMinutes`/`RemainingLoad` are documented as zero for nil targets (`constraints.go:228-233`), but a nil target with completed or fixed work returns a negative value because the calculation uses an implicit zero target (`:821-822`). For example, nil load target plus `CompletedLoad: 10` yields `RemainingLoad: -10` even though load tracking is disabled.

Preserve target presence in `Reconciliation` (for example with pointer target fields or explicit `*_tracked` fields), and calculate/report remaining values consistently for an untracked dimension. Add a regression test covering nil targets with nonzero completed/fixed totals and JSON output, as well as explicit-zero output.

### 2. The design contract still says the opposite of the implemented zero semantics

The design doc says `RequestedSessionCount` pointer-to-zero means all candidates are excess (`docs/design/plan-filler-constraints.md:65`), but the batch-validation section says a count of zero means “no cap is applied” (`:201`). Likewise, the boundary table says `WeeklyTargetLoad == 0` and `WeeklyTargetMinutes == 0` skip budget checks (`:312-313`), while the model documentation and implementation define non-nil zero as a hard zero that warns and blocks positive work.

This leaves a future caller unable to use the document as the contract for the most restrictive constraints. Update the batch and boundary-condition prose to distinguish `nil` (untracked/no cap) from pointer-to-zero (hard zero), matching `ValidateCandidates` and `validateCore`; include the requested-session case in the boundary table.

## Verification

- `git diff --check 6a30598..HEAD` passed.
- `go test ./internal/planning` passed.
- `go test -race ./internal/planning` passed.
- `go test ./...` passed.
