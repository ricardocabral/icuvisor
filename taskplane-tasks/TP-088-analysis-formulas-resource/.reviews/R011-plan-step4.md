# Plan Review R011 — Step 4: Verify

**Verdict:** Approved

The Step 4 plan is narrow and matches the task’s verification requirements: full-suite/build/lint validation, changelog coverage, and a final source-decision record in `STATUS.md`. It is safe to execute as the verification gate after the resource/docs work already completed in Steps 2–3.

## Execution expectations

- Run the exact project gates named by the prompt: `make test`, `make build`, and `make lint`.
- Add the user-visible entry to `CHANGELOG.md` under `[Unreleased]`, preferably in `### Added`, describing the new `icuvisor://analysis-formulas` MCP resource and stable analyzer formula refs.
- Update `STATUS.md` with:
  - command/result lines for all verification commands;
  - any failure details, clearly marked as fixed or pre-existing/unrelated;
  - final formula-source decisions, including any deviations from the Step 1 draft or confirmation that the implemented markdown/golden content matches it.

## Non-blocking notes

- Because Step 5 repeats the same full-suite/build/lint checks, avoid needless duplicate work by recording Step 4 results clearly enough that Step 5 can either reuse them if no files changed afterward or rerun only after subsequent edits.
- Check `git status --short` before and after verification so only expected task files (`CHANGELOG.md`, `STATUS.md`, and any legitimate fixes from failed gates) remain modified.
