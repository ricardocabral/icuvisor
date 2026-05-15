# R003 Code Review — Step 1: Backend selection and contract

## Verdict: approved

No blocking findings.

The new `internal/credstore` package matches the Step 1 contract: it defines the stable service/account constants, a context-aware `Store` interface, the project-local `ErrNotFound` sentinel, and a `NoopStore` that honors cancellation and reports `ErrNotFound` for reads. `STATUS.md` records the backend selection, license review, canonical names, precedence decisions, and error semantics needed for the next implementation step.

## Validation

- `git diff 9e65286..HEAD --name-only`
- `git diff 9e65286..HEAD`
- `go test ./internal/credstore`
- `go test ./...`

## Notes for Step 2

- Keep `ErrNotFound` as the only fall-through signal and map upstream keyring misses/headless Linux startup lookup failures to it with `errors.Is` compatibility.
- Do not let write/delete paths silently succeed when the OS keychain is unavailable; return actionable wrapped errors without including the secret value.
