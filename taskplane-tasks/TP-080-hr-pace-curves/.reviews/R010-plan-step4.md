# Plan Review — TP-080 Step 4

Verdict: **APPROVE**

The Step 4 plan matches the remaining task requirements: regenerate the public tool reference, add a user-visible changelog entry, and run full verification. It is appropriately deferred until after the HR/pace implementation and symmetry tests are complete.

## Notes

- Use the generated-docs path rather than hand-editing the reference page. In this repository `make docs-tools` regenerates `web/data/tools.json`; `web/content/reference/tools.md` is just the Hugo wrapper around that generated data. Verify the generated data now includes both `get_hr_curves` and `get_pace_curves` in the fitness/full catalog.
- Add the CHANGELOG entry under `[Unreleased]` / `### Added`, since this task adds public MCP tools. Keep it user-facing and mention that the new tools are siblings of `get_power_curves` for upstream HR and pace curves.
- After docs/changelog changes, run the required full commands exactly as listed by the task: `make test`, `make build`, and `make lint`. If any failure is environmental or pre-existing, capture the exact command/output summary in `STATUS.md`; otherwise fix it before marking the step complete.
- Update `STATUS.md` with the verification results and any docs decisions before the step-boundary commit. The commit message should include `TP-080` per the task convention.

## Suggested command sequence

```sh
make docs-tools
git diff -- web/data/tools.json CHANGELOG.md
make test
make build
make lint
```

No plan changes are required before implementation.
