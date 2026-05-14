# R002 plan review — Step 1: Tier enum and parsing

Verdict: **approved for implementation**

The revised Step 1 plan in `STATUS.md` addresses the gaps called out in R001. It now explicitly keeps toolset policy in `internal/safety`, wires `ICUVISOR_TOOLSET` through config loading, propagates the parsed value into app startup via `ServerInfo`, logs the resolved tier once, and pins the behavior with parsing/config/startup/logging tests.

## What looks good

- **Package boundary is appropriate.** Extending `internal/safety` matches the existing TP-018 registration-gate pattern and avoids a parallel `internal/toolset` policy package.
- **Config/startup propagation is now in scope for Step 1.** Adding `Config.Toolset`, raw/env/.env loading, `Config.String()` rendering, and `ServerInfo` propagation satisfies the “parsed once at startup” acceptance criterion.
- **Testing scope is sufficient.** The planned tests cover the important non-registry behavior for this step: case-insensitive/defaulting parse behavior, defensive string rendering, config precedence including `.env`, propagation to startup, and minimal logging.
- **No premature registry work.** The plan leaves per-tool membership and actual filtering for later steps, which keeps this step focused.

## Implementation notes

- Define a distinct `safety.Toolset` API (`ToolsetCore`, `ToolsetFull`, `ParseToolset`, `EnvToolset`) rather than reusing delete-mode types; delete-mode `full` and toolset `full` have different semantics.
- Keep invalid and empty `ICUVISOR_TOOLSET` values falling back to `core`, per this task prompt.
- The Step 1 startup log should be a single structured line such as `resolved toolset` with only the tier value. Do not add tool names or registered/skipped counts yet; those belong in Step 3 after filtering exists.
- No README/CHANGELOG changes are required in Step 1 unless the step is intentionally shipped independently; the task assigns docs to Step 5.

Proceed with Step 1 implementation.
