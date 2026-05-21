# Plan Review: TP-001 Step 4

Verdict: **Revise**

`STATUS.md` marks Step 4 as in progress, but it does not yet contain a concrete implementation plan for the test pass. It only repeats the three checklist bullets from the task prompt. That is not enough to review because the current tree already has some `internal/app` and `internal/config` tests from earlier steps, so Step 4 needs to say which existing coverage will be kept, which gaps remain, and which behaviors will be added or refactored into table-driven tests.

## Required plan changes

1. **Audit existing tests before adding duplicates.** The plan should explicitly note that `internal/config/config_test.go` already covers normalization, precedence/defaults, `.env` absent-only behavior, `ICUVISOR_CONFIG`, validation-error redaction, and `Config.String()` redaction. Step 4 should either reuse/refactor these tests or list the precise missing cases to add.

2. **Name the config precedence cases to be tested.** Add a small matrix covering at least: explicit `Options.Path` wins over `ICUVISOR_CONFIG`, JSON values are loaded, `.env` fills only absent values, process env overrides JSON/`.env`, and defaults apply only after sources are merged. The Step 3 review specifically called out config-path precedence; Step 4 should lock it down with a test.

3. **Keep tests deterministic and secret-safe.** The plan should state that tests will pass explicit `config.Options{Env: ..., DotEnvPath: ...}` or use `t.Setenv` in isolated cases, and must not read the developer's real process env or repository `.env`. Use dummy credential strings and assert errors/string output never contain those strings.

4. **Define the invalid/missing config error assertions.** List the short actionable error cases to cover: missing API key, missing athlete ID, malformed athlete ID, invalid timezone, invalid timeout, invalid API base URL, malformed JSON, and missing/unreadable config file if that surface is intended to be user-facing. Assertions should check stable substrings rather than full wrapped OS errors, and should verify no API key leakage.

5. **Include app-level foundation behavior if it is part of this test pass.** Existing app tests cover `version`, default delegation, `--config=...`, and unknown commands, but the plan should say whether Step 4 will add missing CLI-path cases such as `--config /path`, missing `--config` value, and explicit proof that `version` remains config-free. If not in scope, say so explicitly.

Once the plan is updated with these concrete test cases and boundaries, Step 4 should be straightforward to approve.
