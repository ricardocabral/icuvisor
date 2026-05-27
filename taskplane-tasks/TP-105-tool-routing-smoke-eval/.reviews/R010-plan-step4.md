# Plan Review: Step 4 — Testing & Verification

**Verdict: Needs revision**

The Step 4 checklist names the required outcomes, but it is still too generic to approve as an execution plan. This step should be deterministic and auditable, especially because the task adds an opt-in networked eval path.

## Required clarifications before execution

1. **Name the targeted commands**
   - Include the specific package tests for the new/changed code, e.g. `go test ./internal/toolrouting ./internal/mcp ./scripts/toolroutingeval` (adjust if the package list differs).
   - Also run `make eval-tool-routing` with provider env unset to verify fixture/catalog validation and skipped live cases remain zero-exit.

2. **Preserve the network-free guarantee**
   - State that normal test gates (`make test`, `make build`, `make lint`) must not require provider credentials or call intervals.icu.
   - If local provider env vars are set, either document that tests are immune to them or run the unset-provider eval with `ICUVISOR_ROUTING_EVAL_PROVIDER`/`ANTHROPIC_API_KEY` explicitly cleared.

3. **Handle the optional provider-backed eval safely**
   - Define the credential check without printing secrets.
   - If credentials are absent, record `skipped: provider credentials unavailable` in `STATUS.md`.
   - If credentials are present, record only a redacted summary: provider/model, pass/fail counts, and failing case IDs; never include API keys or full provider payloads.

4. **Record exact verification results**
   - Update `STATUS.md` with each command, exit status, and a short outcome.
   - If any gate fails, either fix it before Step 4 completes or document why it is pre-existing/unrelated with enough evidence to support that claim.

5. **Formatting check**
   - Add `make fmt-check` or an equivalent `gofmt`/`goimports` check before the final lint/build gates, since repository rules require formatted Go code even though it is not currently listed in the Step 4 checklist.

Once the plan includes those command-level details and the provider/no-provider recording rules, it should be ready to execute.
