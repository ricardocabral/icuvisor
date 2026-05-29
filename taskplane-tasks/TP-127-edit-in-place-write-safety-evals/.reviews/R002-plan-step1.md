# Plan Review — Step 1: Audit write/delete guidance

**Verdict: approved.**

The revised Step 1 plan addresses the R001 blockers: it explicitly includes `create_workout` as the recreate side of the unsafe delete/recreate path, and it adds registration-time delete-gating coverage in `internal/safety/adversarial_test.go` with a decision point for `go test ./internal/safety`.

## Execution notes

- When recording Discoveries, include the rationale for whether `go test ./internal/safety` was run or deferred. If the Step 1 safety contract depends on the adversarial registration matrix, prefer running it.
- While identifying eval gaps, make sure the audit touches the existing eval/adversarial surfaces (`docs/safety/adversarial-prompts.md` and `scripts/eval/scenarios/cookbook_scenarios.json`) even if edits are deferred to Step 2.
- Keep Step 1 read-only except for STATUS.md discovery/log updates; guidance or scenario changes belong in later steps.
