# Plan Review R001 — Step 1: Help-text design

**Verdict:** Changes requested before implementing Step 2.

The task prompt is a good basis, but `STATUS.md` currently only repeats the Step 1 checklist. It does not yet contain the actual design artifact this step is supposed to produce: a drafted help fixture, the verified environment-variable inventory, the library choice, or the exact exit-code wording. Please complete those design decisions before parser changes begin.

## Required fixes to the Step 1 plan/design

1. **Inventory all real environment variables, not only the shortened list in the prompt.**
   The acceptance criteria require every env var read by `internal/config`, `internal/safety`, and `internal/response`. The fixture should include at least:
   - `INTERVALS_ICU_API_KEY`
   - `INTERVALS_ICU_ATHLETE_ID`
   - `ICUVISOR_CONFIG`
   - `ICUVISOR_ENV_FILE`
   - `ICUVISOR_TIMEZONE`
   - `ICUVISOR_API_BASE_URL`
   - `ICUVISOR_HTTP_TIMEOUT`
   - `ICUVISOR_TRANSPORT`
   - `ICUVISOR_HTTP_BIND`
   - `ICUVISOR_DELETE_MODE`
   - `ICUVISOR_TOOLSET`
   - `ICUVISOR_DEBUG_METADATA`

   Evidence: `internal/config/config.go` defines and reads the config vars; `internal/safety/mode.go` and `internal/safety/toolset.go` define delete/toolset vars; `internal/response/shaper.go` defines `ICUVISOR_DEBUG_METADATA`.

2. **Include every CLI flag that exists today.**
   The design must include `--env-file` alongside `--config`, `--transport`, and `--http-bind`. The current parser supports both `--env-file value` and `--env-file=value`, so omitting it from help would fail discoverability.

3. **Draft and name the golden fixture now.**
   Step 1 says “Draft the `--help` output as a fixture before writing code,” but the status has no fixture path or content. Use a concrete path such as `internal/app/testdata/help.golden` (and likely `version_help.golden` if the version subcommand help is non-trivial), then make Step 3 consume that same artifact.

4. **Make the library choice in Step 1.**
   `STATUS.md` still says `TBD`. Given the current surface, the plan should explicitly choose stdlib/hand-rolled help (or justify a framework with license/dependency rationale). Leaving this undecided makes the next step ambiguous.

5. **Specify exact exit-code semantics and error classification.**
   The fixture should document concrete meanings, preferably:
   - `0` success/help/version
   - `1` runtime/config/server error
   - `2` usage error (unknown flag/command, missing flag value)

   Also note that `cmd/icuvisor/main.go` currently exits `1` for every error, so the implementation will need a stable way to distinguish usage errors from runtime errors.

6. **Design the unknown-flag/missing-value stderr text explicitly.**
   The mission asks for a one-line `Usage:` on unknown-flag errors, and the acceptance criteria require `Run 'icuvisor --help' for usage.`. The plan should pin the exact stderr shape so tests can assert it without drifting.

## Suggested design notes

- Keep top-level help terse and operational: description, usage, commands, flags, env vars, examples, exit codes, docs pointer.
- Avoid documenting API keys as CLI flags; env/config only.
- For `ICUVISOR_HTTP_BIND`, state that loopback is the default and LAN/non-loopback binding is an explicit security choice.
- For `ICUVISOR_DELETE_MODE`, state `safe` default and that `full` enables destructive tools only at registration/startup.
- For `ICUVISOR_TOOLSET`, state `core` default and `full` includes heavier/advanced tools.

Once the fixture and inventory are added to the plan/status, Step 1 should be straightforward to approve.
