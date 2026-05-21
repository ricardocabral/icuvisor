# Plan Review — TP-080 Step 5

Verdict: **APPROVE**

The Step 5 verification plan matches the task completion criteria: rerun targeted affected tests, run the full suite, build, lint, and either fix failures or document clearly pre-existing/unrelated failures in `STATUS.md`. This is appropriate even though Step 4 also ran full verification, because Step 5 is the final clean verification gate from the current tree.

## Notes

- Do not treat the Step 4 runs as a substitute unless their exact command outputs are still from the final committed/current state. Step 5 should record fresh results in `STATUS.md`.
- Targeted tests should cover all packages touched by this task, not just the new tool files. At minimum, include tools, intervals client curve endpoint tests, tool catalog/checks, and safety/coach catalog surfaces affected by registration.
- Include the schema/catalog stability check if it is not already covered by the chosen make targets, because this task added public MCP tools and schema snapshots.
- If `make lint` or `make test` fails due to environment/tooling, capture the exact command, short failure summary, and why it is unrelated/pre-existing in `STATUS.md`. Otherwise fix before marking the step complete.
- After verification, run `git status --short` to ensure no generated artifacts or formatting changes were produced by the checks.

## Suggested command sequence

```sh
go test ./internal/tools ./internal/intervals ./internal/toolcatalog ./internal/toolchecks ./internal/safety
go run ./scripts/check_schema_stability.go -baseline-dir internal/tools/schema_snapshot -require-baseline
make test
make build
make lint
git status --short
```

No plan changes are required before executing Step 5.
