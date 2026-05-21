# R004 Plan Review — Step 4: Documentation

Verdict: APPROVED

No blocking issues with the Step 4 plan. The planned documentation scope matches the prompt: add a short README pointer to `icuvisor --help`, add a `[Unreleased]` changelog entry under `Added`, and avoid duplicating the full CLI/env-var reference outside the help output.

Execution notes:

- Keep the README change to one sentence near Quickstart, e.g. after the initial build/run snippet, so `--help` remains the source of truth for flags and environment variables.
- Add the changelog entry under the existing `## [Unreleased]` / `### Added` section.
- Do not copy the complete flag/env-var list into README; the existing detailed transport/delete/toolset sections should not be expanded as part of this step.
- Before marking the step complete, update `STATUS.md` checkboxes and consider cleaning up the stale `Decisions` line that still says `Library choice: TBD`, since the notes already record the stdlib/hand-rolled parser decision.

The plan is appropriately small and leaves build/test/lint verification to Step 5.
