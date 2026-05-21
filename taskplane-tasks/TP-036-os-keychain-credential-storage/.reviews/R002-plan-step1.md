# R002 Plan Review — Step 1: Backend selection and contract

## Verdict: approved to implement

The updated `STATUS.md` addresses the Step 1 blockers from R001. The plan now records a concrete backend choice, dependency/license facts, canonical service/account names, an exact `Store` contract, local `ErrNotFound` semantics, and the key `.env` precedence decision. This is sufficient to proceed with implementation.

## What looks good

- `the selected OS keyring module` is the right default for this task: MIT, small surface area, OS-native backends, no CGO requirement for the target paths, and no encrypted-file fallback that would undercut the project security posture.
- The selected version/release evidence matches `go list -m -json the selected OS keyring module@v0.2.8`.
- The context-aware interface is appropriate for this repo convention:
  ```go
  Get(ctx context.Context, account string) (string, error)
  Set(ctx context.Context, account, secret string) error
  Delete(ctx context.Context, account string) error
  ```
- `ServiceName = "icuvisor"` and `IntervalsAPIKeyAccount = "intervals-icu-api-key"` make the storage key stable and greppable.
- Mapping only project-local `credstore.ErrNotFound` as the fall-through sentinel keeps config precedence testable and avoids leaking the upstream dependency through `internal/config`.
- Treating process env as the only highest-priority explicit env override, with keychain above plaintext `.env`/JSON fallback, aligns with the task's security goal.

## Minor follow-ups before/during Step 2 implementation

1. **Mark the Step 1 checklist items complete once the contract file is created.** The decisions are now in `STATUS.md`, but the checkboxes for the `Store` interface and error semantics should move to done after `internal/credstore` lands.

2. **Be precise about Linux error mapping.** The plan's `Get` degradation should apply to an unavailable D-Bus/session keyring/headless environment. Do not map arbitrary keyring failures, permission denials, malformed responses, or other unexpected errors to `ErrNotFound`; those should remain wrapped actionable errors so real problems are visible.

3. **Preserve or explicitly document the file-source sub-order.** `STATUS.md` correctly says `.env` and JSON are below keychain and report source `file`. When implementing tests, make the sub-order explicit too. The current loader effectively gives JSON config precedence over `.env` for values already present, while process env overrides both. Preserve that unless this task intentionally changes it.

4. **Remember final MVS dependency state.** The license review is fine. Note that the repository already requires `golang.org/x/sys` at a newer version than `go-keyring`'s `v0.27.0`, so the final `go.mod`/`go.sum` may resolve `x/sys` via MVS differently. License remains BSD-style either way.

No further plan changes are required for Step 1.
