# R001 plan review — Step 1: Tier enum and parsing

Verdict: **changes requested before implementation**

The package-boundary choice in `STATUS.md` is sound: extending `internal/safety` matches the existing `ICUVISOR_DELETE_MODE` registration-gate pattern and avoids creating a parallel policy package. The step plan also captures the critical product behavior: `core` default, `full` opt-in, case-insensitive parsing, and no tool-name leakage in startup logs.

However, the current Step 1 plan/checklist is too thin to guarantee the acceptance criterion that `ICUVISOR_TOOLSET` is parsed once at startup and propagated through the process. Please make the following implementation details explicit before coding.

## Required plan adjustments

1. **Define a separate toolset type in `internal/safety`; do not reuse delete-mode types.**
   - Use distinct names such as `type Toolset string`, `ToolsetCore`, `ToolsetFull`, `ParseToolset`, and `EnvToolset = "ICUVISOR_TOOLSET"`.
   - Avoid overloading `safety.Mode`/`ModeFull`; delete mode `full` and toolset `full` have different meanings and must not share APIs accidentally.
   - `Toolset.String()` should defensively return `core` for unknown values, mirroring `Mode.String()`.

2. **Add config-loader plumbing in Step 1, not later.**
   To satisfy “parsed once at startup,” this should flow through `internal/config` the same way `DeleteMode` does today:
   - `Config.Toolset safety.Toolset`
   - `rawConfig.toolset`
   - `rawFromEnv` reads `safety.EnvToolset`
   - `rawConfig.merge` merges it
   - `validate` sets `Toolset: safety.ParseToolset(raw.toolset)`
   - `recognizedEnvKey` includes `safety.EnvToolset` so `.env` works
   - `Config.String()` includes `toolset=%s` and continues not to leak credentials or athlete IDs.

3. **Propagate to app startup and log exactly once.**
   - Add `Toolset safety.Toolset` to `app.ServerInfo` and set it from `cfg.Toolset` in `startServer`.
   - Add a `safety.LogResolvedToolset(logger, toolset)` (or equivalent) near `LogResolvedMode` in `defaultStartServer`.
   - Keep the log structured and minimal, e.g. `logger.Info("resolved toolset", "toolset", "core")`. Do not log tool names. Registration counts belong in Step 3 once membership/filtering exists.

4. **Pin behavior with tests now.**
   Minimum tests for this step should include:
   - table-driven `ParseToolset` cases: empty, `core`, `full`, whitespace, mixed/upper case, unknown → `core`;
   - `Toolset.String()` defensive default for invalid values;
   - config loading from process env and `.env`, including precedence and invalid value fallback to `core`;
   - `Run`/`startServer` propagation of `Config.Toolset` into `ServerInfo`;
   - startup logging contains one resolved-toolset line and no tool names.

## Notes for later steps

- Step 1 should not add per-tool membership or registry filtering yet, but it should leave a clean API that Step 2/3 can consume.
- Keep the “unknown/empty → core” behavior for `ICUVISOR_TOOLSET`, even though delete-mode invalid behavior differs in the PRD text; the task prompt is explicit for this task.
- No README/CHANGELOG update is necessary in Step 1 unless this step is intended to ship independently; the task already assigns docs to Step 5.

Once the plan includes the config/app propagation and test coverage above, Step 1 is ready to implement.
