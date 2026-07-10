# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. First-fit slot consumption rejects feasible candidate batches based only on slot declaration order

`ValidateCandidates` consumes the first compatible slot (`constraints.go:488-491`). Slots have no priority or candidate-selected slot identifier, so this greedy choice is part of the validator's implicit allocation policy rather than an input fact. It can reject a schedule that fits all hard constraints:

```go
Slots: [any-sport 60 min, Ride-only 60 min]
Candidates (in order): [Ride 60 min, Run 60 min]
```

The Ride claims the any-sport slot. The Run is then rejected as `sport_not_allowed`, although assigning the Ride to the Ride-only slot and the Run to the any-sport slot places both sessions. This conflicts with the contract's purpose of determining whether a candidate schedule is placeable; candidate order is documented as priority only for `RequestedSessionCount`, not for arbitrary slot order.

Use a deterministic maximum matching/allocation across the day's candidates (with an explicit tie-break), or add a caller-supplied slot assignment / explicitly documented first-fit priority policy. The former is appropriate if this API is meant to validate schedule feasibility. Add the regression case above.

### 2. `Reconcile` bypasses the numeric-safety contract and can return an unmarshalable future MCP response

The documentation says invalid candidate numbers are prevented from propagating into `Reconciliation` fields, and the batch validator correctly excludes them. However, `Reconcile` blindly sums every candidate (`constraints.go:358-361`). `Reconcile(wc, []CandidateSession{{Load: math.NaN()}})` returns a reconciliation with `NaN`, for which `json.Marshal` fails. Negative values similarly manufacture a misleading candidate total.

Either make `Reconcile` exclude invalid candidates exactly as `ValidateCandidates` does, or change it to return an error and require valid inputs. Document the chosen behavior and add JSON-marshalling coverage. The currently documented “does not validate” behavior is not sufficient for a result type intended for future MCP output.

## Contract/documentation correction

The exported `ValidateCandidates` comment still says **all** candidates, including invalid ones, increment daily minutes and weekly prior totals (`constraints.go:389-391`). The implementation and design document instead isolate numeric-invalid candidates from every numeric accumulation. Correct this public comment so callers have one authoritative batch-accounting rule.

## Verification

- `go test ./internal/planning` passes (`[no test files]`).
- `go vet ./internal/planning` passes.
- `git diff --check 6a30598..HEAD` passes.
- Focused temporary checks reproduced both findings and were removed without changing the worktree.
