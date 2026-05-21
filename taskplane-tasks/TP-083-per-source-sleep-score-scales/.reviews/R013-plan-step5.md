# Plan Review: Step 5 — Testing & Verification

## Verdict: Request changes

The Step 5 checklist matches the task prompt at a high level, but it is not yet specific enough for a final verification step. Earlier steps introduced focused wellness assertions and generated-doc/golden changes; the plan should name the exact targeted commands and how results will be recorded before execution.

## Findings

1. **Targeted tests are not identified.**
   - “Targeted tests passing” should be expanded to the affected packages/tests from this task, for example:
     - `go test ./internal/tools -run 'TestGetWellnessData(Fixtures|NullStrippingAndIncludeFull)'`
     - `go test ./internal/intervals -run Wellness`
     - `go test ./cmd/gendocs` if the generated tool catalog/golden fixture was changed in Step 4.
   - This protects the provider-native `native_scale` assertions, unknown-source fallback, native extraction coverage, and docs-generation golden behavior without relying only on the full suite.

2. **Step 5 should explicitly repeat the full gates, not rely on Step 4.**
   - Step 4 notes say `make test && make build && make lint` passed after regenerating docs/goldens, but the prompt requires Step 5 to run task-level verification.
   - The plan should state that Step 5 will run fresh, separate commands: `make test`, `make build`, and `make lint`, then record each outcome in `STATUS.md`.

3. **Failure handling needs a concrete standard.**
   - The current “all failures fixed or documented” checkbox is acceptable in spirit, but the plan should require documenting the failing command, concise error summary, and why a remaining failure is demonstrably pre-existing/unrelated.
   - Task-related failures, stale generated docs/goldens, formatting/lint failures, or wellness test regressions should be fixed before marking the step complete.

## Suggested Step 5 acceptance criteria

Before implementation, expand Step 5 in `STATUS.md` with something like:

- Run targeted checks:
  - `go test ./internal/tools -run 'TestGetWellnessData(Fixtures|NullStrippingAndIncludeFull)'`
  - `go test ./internal/intervals -run Wellness`
  - `go test ./cmd/gendocs` if Step 4 touched generated docs/goldens.
- Run full verification gates as fresh commands: `make test`, `make build`, and `make lint`.
- Record exact pass/fail outcomes in `STATUS.md`.
- Fix any task-related failures. Only document failures as pre-existing/unrelated when the command output and scope make that clear.
