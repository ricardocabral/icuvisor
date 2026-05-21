# Plan Review — Step 4: Docs and verification

Verdict: REVISE

The revised Step 4 plan now covers the concrete docs surfaces, planned `[Unreleased]` changelog entry, exact verification commands, and STATUS evidence requirements from R014. Those parts are ready to execute.

One R014 requirement is still not fully resolved: the Step 4/Step 5 handoff is still a placeholder rather than a policy. The new checkbox says to “Record Step 4 verification evidence and Step 5 handoff policy in STATUS.md,” but it does not state what the policy is. Because Step 4 and Step 5 both claim the same quality gates, the plan should decide this before execution.

Required revision:

1. Replace the handoff placeholder with an explicit policy, for example one of:
   - Step 5 will re-run `go test ./internal/analysis`, `make test`, `make build`, and `make lint` as final confirmation regardless of Step 4 results; or
   - Step 5 may reuse Step 4 command results only if no files changed after those commands, and must re-run any affected gates after changelog/docs/status edits or code changes.

The second option is likely the better fit here because Step 4 intentionally updates docs/CHANGELOG/STATUS while also running verification. Make the rule concrete enough that the Step 5 worker can apply it without reconstructing intent.

After that single clarification, the Step 4 plan should be safe to execute.
