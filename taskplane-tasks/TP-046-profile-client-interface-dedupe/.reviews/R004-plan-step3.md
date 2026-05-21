# Review R004 — Step 3 plan

Decision: approved for Step 3 implementation.

The Step 3 plan is sufficient for this refactor. `go build ./...` is the right fast compiler check for the shared interface move and will catch any package-level fake/stub or constructor signature regressions. `make test` and `make test-race` cover the existing tool, resource, MCP, and toolcheck test surfaces without requiring new tests for a pure type-placement change.

Execution notes:

- Run all three commands exactly as listed, and record the outcomes in `STATUS.md` before moving to verification.
- If any command fails, fix the underlying compile/test issue rather than broadening the interface or changing method signatures; the task scope is still a placement-only dedupe.
- Do not skip `make test-race` just because Step 4 also lists it. Step 3 should establish that the implementation is test-clean before the final lint/build/grep sanity checks.
- Keep the Step 4 acceptance checks separate: especially the `grep -rn "type ProfileClient interface" internal/` one-hit verification and `make lint`.
