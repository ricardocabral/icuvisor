# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking finding

### Finite inputs can overflow reconciliation into non-JSON values

`Reconcile` accepts each finite, non-negative candidate and sums its numeric fields without checking the result (`internal/planning/constraints.go:348-357`). `buildReconciliation` similarly adds completed, fixed, and candidate values without guarded arithmetic (`:857-878`). Thus valid inputs can produce `+Inf` in `CandidateLoad`, `CandidateMinutes`, or projected totals. This violates the documented JSON-safe reconciliation contract and makes a future MCP response unmarshalable.

For example, with an otherwise unconstrained day allowing two sessions, two candidates each with `Load: math.MaxFloat64` are both returned valid by `ValidateCandidates`; `Reconciliation.CandidateLoad` becomes `+Inf`, and `json.Marshal(batch)` fails with `json: unsupported value: +Inf`.

Reject or explicitly report arithmetic overflow before it reaches a result (and apply the same treatment to completed/fixed/projected and remaining calculations). Do not silently clamp or drop load/time. Add regression tests that marshal both `Reconcile` and `ValidateCandidates` results at overflow boundaries.

## Verification

- `go test ./internal/planning` passed.
- `go test -race ./internal/planning` passed.
- `make lint`, `make test`, `make build`, and `git diff --check 6a30598..HEAD` passed.
- A focused temporary regression test reproduced the overflow and JSON-marshal failure; it was removed after execution.
