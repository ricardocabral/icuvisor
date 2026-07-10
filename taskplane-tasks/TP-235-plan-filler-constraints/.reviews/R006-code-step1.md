# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. `ValidateCandidate` still returns non-JSON-serializable invalid input

`ValidateCandidates` sanitizes an invalid candidate before embedding it in `CandidateResult` (`internal/planning/constraints.go:413-418`), but the single-candidate public API returns the original candidate unchanged (`:344-347`). Thus `ValidateCandidate(wc, CandidateSession{DurationMinutes: math.NaN()})` correctly produces `invalid_input`, but `json.Marshal` of its result fails because `Candidate.DurationMinutes` is still NaN. This contradicts the R005 fix's stated serialization goal and makes one of the two future-MCP-facing validation APIs unsafe to use.

Sanitize the candidate on this return path too (and adjust the design doc's "echoed verbatim" statement), or return an error that cannot contain non-finite values. Add a regression test that marshals invalid results from both public validation functions.

### 2. Week/day dates are not validated, so a "weekly" constraint can schedule outside its week

`ValidateWeekConstraints` validates numeric fields and duplicate day strings but does not parse `WeekStartDate` or `DayConstraints.Date`, verify that the start is a Monday, or require each available day to fall within that Monday–Sunday week (`internal/planning/constraints.go:263-312`). Consequently, `WeekStartDate: "2026-07-06"` with an available day and candidate date of `"2026-07-20"` passes validation and is reported valid. Malformed date strings likewise become usable availability keys.

This leaves the plan filler able to certify work outside the requested calendar week, contrary to the documented `WeekStartDate` and "days within this week" contract. Validate canonical ISO dates, require the week start to be Monday, and reject available dates outside that week. Add malformed, non-Monday, and out-of-week boundary tests.

### 3. The design contract still states the opposite invalid-candidate accounting behavior from the implementation

The implementation intentionally excludes numeric-invalid candidates from daily-minute accumulation, weekly prior totals, reconciliation totals, and `infeasible_load` (`constraints.go:410-424`, `:466-469`, `:491-510`). However, the design document and exported warning comment still say invalid candidates are included: `docs/design/plan-filler-constraints.md:173,198-200` and `constraints.go:179-181`. The document also says `CandidateResult.Candidate` is echoed verbatim (`docs/...:132`), which is no longer true for batched invalid input, and omits `invalid_input` from the violation-code table.

This is a load-bearing contract for later scheduling code; callers cannot determine whether malformed proposals reduce available budget or suppress infeasibility warnings. Update the design doc and exported comments to match the R005 isolation semantics (and explicitly document the intentional session-counter behavior), document `invalid_input`, and cover the chosen behavior with tests.

## Verification

- `go test ./internal/planning` and `go vet ./internal/planning` pass, but the package has no committed tests.
- A temporary regression test confirmed that single-candidate NaN results fail JSON marshaling and that an out-of-week availability date is accepted.
- `gofmt -d internal/planning/constraints.go` is clean.
- `git diff --check 6a30598..HEAD` fails on trailing whitespace and an extra final blank line in the committed `R001-plan-step1.md` review artifact; clean it before the required formatting/diff check.
