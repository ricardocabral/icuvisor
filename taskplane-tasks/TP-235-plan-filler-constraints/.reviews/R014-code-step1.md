# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking finding

### Finite constraint arithmetic can still put `Inf` in validation responses

The R013 overflow handling only protects the `Reconciliation` returned by `Reconcile`; it does not protect the remaining-budget arithmetic used to build `CandidateResult` values. `validateCore` computes and serializes `remainingLoad` / `remainingMin` directly at `constraints.go:742-766` (and the analogous time block). `ValidateWeekConstraints` accepts each operand independently as finite, so a valid constraint object can nevertheless overflow the subtraction.

For example, with a zero weekly load target, `CompletedLoad` and `FixedLoad` both `math.MaxFloat64`, and a compatible candidate with `Load: 1`:

1. `ValidateWeekConstraints` returns nil (all supplied numbers are finite and non-negative).
2. `ValidateCandidate` computes `0 - MaxFloat64 - MaxFloat64` as `-Inf` and returns it in both the `zero_remaining_load` warning and `weekly_load_overshoot` violation.
3. `json.Marshal` of that `CandidateResult` fails. `ValidateCandidates` has the same invalid per-candidate values even though its later `Reconcile` call replaces only the batch reconciliation with zero and adds `arithmetic_overflow`.

The same failure is possible for minutes and when the batch's `priorLoad` / `priorMinutes` additions overflow before a later candidate is evaluated. This violates the documented JSON-safe result contract and leaves a future MCP caller with an unmarshalable validation response despite valid finite inputs.

**Required change:** detect overflow for all intermediate budget/daily accumulations before adding them to a `Violation` or `Warning` value. Return a JSON-safe, explicit overflow outcome on both single and batch paths (and do not certify work when the applicable hard budget cannot be computed). Extend constraint preflight validation to reject overflowing completed/fixed target arithmetic as appropriate, but do not rely on that alone: both public validator return types must remain safe when called directly. Add regression tests that marshal single and batch results for completed/fixed overflow, plus overflow in prior batch accumulation, for both load and time.

## Verification

- `go test ./internal/planning` passed.
- `go test -race ./internal/planning` passed.
- `make lint` passed (0 issues).
- `git diff --check 6a30598..HEAD` passed.
- A temporary focused regression test reproduced the finite-input overflow and confirmed that both `CandidateResult` and `BatchResult` fail JSON marshaling; it was removed after verification.
