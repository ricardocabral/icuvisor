# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking finding

### `ValidateCandidate` bypasses the explicit zero-session constraint

`WeekConstraints.RequestedSessionCount` uses pointer-to-zero as a hard constraint: the public documentation says that, with `RequestedSessionCount == pointer-to-0`, **every** candidate receives `requested_session_count_exceeded` (`docs/design/plan-filler-constraints.md:318`). `ValidateCandidates` implements this correctly, but `ValidateCandidate` passes directly to `validateCore` without applying this constraint (`internal/planning/constraints.go:363-380`). As a result, a compatible single candidate is returned `Valid: true` when the requested session count is zero.

This makes equivalent single- and batch-validation entry points disagree and permits a future caller using `ValidateCandidate` to certify a session in a zero-session week, defeating the hard requested-session constraint.

**Required change:** have `ValidateCandidate` add `ViolationRequestedSessionCountExceeded` when `RequestedSessionCount` is non-nil and zero (or otherwise make the documented zero-count contract explicitly batch-only, which would weaken the public contract). Add a regression test covering the single-candidate zero-count path alongside the existing batch test.

## Verification

- `go test ./internal/planning` passed.
- `go test -race ./internal/planning` passed.
- `make lint` passed.
- `git diff --check 6a30598..HEAD` passed.
- A focused temporary regression test confirmed that `ValidateCandidate` accepts a compatible candidate with `RequestedSessionCount: &0`.
