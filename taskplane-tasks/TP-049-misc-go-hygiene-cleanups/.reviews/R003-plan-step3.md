# Plan Review: Step 3 — Move env read into config

Verdict: Approved with implementation clarifications.

The step is correctly scoped: `internal/app` should stop calling `response.DebugMetadataFromEnv()`, and `config.Load` should be the single place that resolves `ICUVISOR_DEBUG_METADATA` into a boolean on `config.Config`. This matches the task acceptance criteria and the repo rule that environment lookups belong in `internal/config`.

Implementation notes to keep the change focused:

- Add a `DebugMetadata bool` field to `config.Config` with `json:"-"`. Do not add a JSON config field for it unless the task is explicitly amended; this step is about the env toggle.
- Move the env-var constant to `internal/config` (for example `config.EnvDebugMetadata`) and update `internal/app/help.go` plus tests that currently refer to `response.EnvDebugMetadata`.
- Remove `response.DebugMetadataFromEnv()` and the `os` import from `internal/response/shaper.go`. Keeping a pure parser is fine, but avoid making `config` depend on `response` if possible; a small `parseDebugMetadata` helper in `internal/config` keeps layering clean.
- Make sure `processEnv()` recognizes `EnvDebugMetadata`; otherwise real process env values will be filtered out before `Load` sees them.
- Preserve the existing quiet parsing semantics: only trimmed, case-insensitive `true` enables debug metadata; empty/invalid values resolve to `false` without an error.
- Be deliberate about precedence/source behavior. The current behavior reads the real process env once before `.env` loading, so naively adding `debugMetadata` to `rawFromEnv` would let `.env` start enabling debug metadata. That may be a reasonable config-consistency improvement, but it is technically user-visible. If the goal is strict no-behavior-change, resolve from the process env map (`opts.Env` / `processEnv()`) only, not from JSON and not from `.env`. If `.env` support is intentionally added, record that decision in `STATUS.md`.
- `Run` does not directly have the loaded config; the natural place to plumb this is in `startServer` after `cfg, err := loader(...)`: set `info.Config = cfg` and `info.DebugMetadata = cfg.DebugMetadata` before calling the starter. Remove the current `ServerInfo{..., DebugMetadata: ...}` initialization from `Run`.
- Update app tests that currently mutate `ICUVISOR_DEBUG_METADATA` around a stubbed `LoadConfig`; with this refactor, stub loaders must return `config.Config{DebugMetadata: true}` to assert propagation. Add/adjust `internal/config` tests to cover env parsing and default false.
- Keep the response-shaper tests for the pure shaping behavior, but move/remove tests for env reading from `internal/response` since env resolution no longer belongs there.
- Update the help golden and the env-var audit test if the constant moves from `response` to `config`.

Recommended verification for this step:

- `grep -rn "os.Getenv\b" internal/ | grep -v "^internal/config/"` returns no matches.
- `grep -rn "DebugMetadataFromEnv" internal/` returns no matches.
- `make build`, `make test`, `make test-race`, and `make lint` pass.

No blockers found.
