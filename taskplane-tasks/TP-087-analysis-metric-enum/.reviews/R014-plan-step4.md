# Plan Review — Step 4: Docs and verification

Verdict: REVISE

The Step 4 plan has the right high-level themes, but it is too underspecified for this task's completion criteria and duplicates Step 5 without saying exactly what will be run or recorded.

Required revisions:

1. **Make the docs check concrete.** List the surfaces to inspect and the expected outcome, e.g. `README.md`, `web/content/reference/tools.md`, `web/data/tools.json`, `docs/prd/PRD-icuvisor.md`, and `ROADMAP.md`. Because analyzer tools are not registered yet, the likely outcome is "no generated tool-reference update required"; record that evidence in `STATUS.md` rather than editing generated/catalog docs unnecessarily.
2. **Specify the CHANGELOG entry.** Add a planned `[Unreleased]` entry, preferably under `Added`, describing the reusable closed `analysis_metric` enum/validation helpers and concise unknown-metric hints for planned analyzer tools. This is a task-level must-update.
3. **Expand "full quality gate" into exact commands.** At minimum run and record:
   - `go test ./internal/analysis`
   - `make test`
   - `make build`
   - `make lint`
   If any command fails, the plan must say to fix it or document it in `STATUS.md` as a pre-existing unrelated failure with evidence.
4. **Update `STATUS.md` as part of Step 4.** Record docs surfaces checked, changelog update, verification commands/results, and any discoveries. This is required by the task documentation requirements and prevents Step 5 from having to reconstruct what happened.
5. **Clarify Step 4 vs Step 5 handoff.** Since Step 4 says "Run full quality gate" and Step 5 repeats the same verification, state whether Step 5 will re-run the gates as final confirmation or consume the Step 4 results if no code changes occur afterward. Avoid leaving two steps each claiming the same unchecked gate.

Once those details are added, the plan should be safe to execute. The main implementation already appears scoped to `internal/analysis`, so Step 4 should focus on documenting the new contract, proving no current tool-reference churn is needed, and capturing reproducible verification output.
